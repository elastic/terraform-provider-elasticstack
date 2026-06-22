## Context

`elasticstack_fleet_elastic_defend_integration_policy` maps typed Terraform attributes into the
Fleet package policy typed-input config envelope (`config.policy.value`, `config.integration_config`,
etc.). The resource already models mainstream Defend policy settings per OS (`policy.windows`,
`policy.mac`, `policy.linux`) but deliberately deferred advanced settings when the resource was
first introduced.

Elastic Defend advanced settings live in the API as nested objects under `policy.{os}.advanced`
(see Kibana `PolicyConfig` types). Elastic documentation refers to the same settings using
OS-prefixed dot notation (e.g. `linux.advanced.artifacts.global.base_url`). There are 100+ settings,
they evolve frequently, and values are heterogeneous strings (booleans, integers, enums, PEM
certificates, comma-separated lists) all serialized as strings in the policy payload.

Primary user driver: air-gapped deployments that must override artifact download URLs and related
TLS/proxy settings per [Configure offline endpoints and air-gapped environments](https://www.elastic.co/docs/reference/security/defend-advanced-settings).

## Goals / Non-Goals

**Goals:**

- Expose `advanced_settings` as an optional `map(string)` on the Defend integration policy resource.
- Use Elastic's documented dot-notation keys so users can copy names directly from
  [Elastic Defend advanced settings](https://www.elastic.co/docs/reference/security/defend-advanced-settings).
- Translate flat map keys ↔ nested `policy.{os}.advanced` objects in `buildPolicyPayload` /
  `populateModelFromAPI` without `kbapi` changes.
- Preserve existing create bootstrap/finalize flow, private-state handling, and unmanaged-field
  semantics for settings not in Terraform config.
- Add focused unit tests plus one acceptance test for artifact base URL round-trip.

**Non-Goals:**

- Modeling each advanced setting as a typed nested Terraform attribute (too large and volatile).
- A raw `policy_json` or `advanced_settings_json` escape hatch (map is sufficient and plan-friendly).
- Validating setting value semantics, Elastic version gates per setting, or cross-setting
  dependencies (Endpoint enforces these at runtime).
- Expanding `elasticstack_fleet_integration_policy` or generic package-policy resources.

## Decisions

### Use top-level `advanced_settings` map(string) with documented dot-notation keys

Add a resource-level `advanced_settings` attribute rather than nesting under `policy.{os}`.

**Rationale:** Elastic docs and air-gapped guides reference fully qualified names
(`linux.advanced.artifacts.global.base_url`). A single map keeps HCL compact when configuring the
same class of setting across OSes and avoids duplicating a dynamic subtree under three OS blocks.

**Alternative considered:** `policy.linux.advanced` as a nested JSON/object attribute per OS.
Rejected because it splits configuration across three blocks and still requires a dynamic schema for
100+ keys.

### Flatten/unflatten between dot keys and nested API objects

Implement small helpers (e.g. `advancedSettingsToPolicyNested`, `advancedSettingsFromPolicyNested`)
that:

1. Parse key prefix `^(linux|mac|windows)\.advanced\.(.+)$`.
2. Split the remainder on `.` to build nested maps (`artifacts.global.base_url` →
   `{"artifacts":{"global":{"base_url": value}}}`).
3. Merge into `policy[os]["advanced"]` when building the API payload alongside existing typed
   `policy` fields.
4. On read, walk `policy.{os}.advanced` and emit `"{os}.advanced.{path}"` keys.

Leaf values remain strings in both directions.

**Alternative considered:** Store nested structure in Terraform with `DynamicPsuedoType`. Rejected
due to weak plan-time validation and poor documentation ergonomics.

### Unmanaged when null; explicit empty map clears managed keys

Match the `description` unmanaged pattern:

- `advanced_settings` null/absent → omit `advanced` subtrees built from Terraform on write; do not
  read into state.
- `advanced_settings` set → read/write only configured keys; include in finalize/update payloads.

When the user sets `advanced_settings = {}`, send empty `advanced` objects for OSes that previously
had managed advanced keys in state so Terraform can clear settings it used to manage.

**Alternative considered:** Always read advanced settings into computed state. Rejected because it
would force all users to see 100+ keys in plans for policies configured partly in Kibana.

### Merge advanced into existing policy payload builders

Extend `buildPolicyPayload` / `mapPolicyFromAPI` (or adjacent helpers) to merge advanced subtrees
after typed OS blocks are built, rather than a separate top-level config key. Advanced settings in
the Fleet API are part of `config.policy.value.{os}.advanced`, not a sibling of `policy`.

### No schema version bump

`advanced_settings` is a new optional attribute. Existing state without the key is valid (null).
No state upgrader required.

## Risks / Trade-offs

- **[Risk] Incorrect flatten/unflatten for deeply nested keys** → Mitigation: table-driven unit
  tests covering multi-segment paths, multiple OSes, round-trip fidelity, and merge with typed
  `policy` fields.
- **[Risk] Partial management leaves drift between Terraform and Kibana** → Mitigation: document
  that only configured keys are managed; recommend importing/refreshing after adopting advanced
  settings for policies previously edited in UI.
- **[Risk] Empty-map clear semantics may surprise users** → Mitigation: document `advanced_settings
  = {}` behavior in resource docs; acceptance test focuses on set/update, not mass clear.
- **[Risk] New Elastic settings appear without provider updates** → Mitigation: opaque string map
  accepts any valid key matching the prefix pattern; no provider release needed per new setting.

## Migration Plan

1. Implement schema, models, mapping helpers, and request integration.
2. Update generated resource documentation.
3. Ship as additive, non-breaking provider release.
4. Archive change and sync delta into `openspec/specs/fleet-elastic-defend-integration-policy/spec.md`.

Rollback: remove `advanced_settings` from Terraform config; unmanaged server values persist.

## Open Questions

- Should validation reject unknown OS prefixes beyond `linux|mac|windows`? **Proposal:** yes,
  enforce the prefix pattern only; do not maintain an allowlist of setting names.
- Should PEM-heavy settings be marked `Sensitive`? **Proposal:** no for v1; entire map is
  optional and users can use Terraform `sensitive` on variables if needed.

## Sources

- API mapping: `internal/fleet/elastic_defend_integration_policy/{request.go,mapping.go}`
- Kibana policy types: `PolicyConfig.{windows,mac,linux}.advanced`
- Elastic docs: https://www.elastic.co/docs/reference/security/defend-advanced-settings
