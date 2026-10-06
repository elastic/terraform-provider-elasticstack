## Why

The Kibana Dashboard API's root `query` object has a single string field, `expression` (`kbn-as-code-query` in the pinned OpenAPI spec: `additionalProperties: false`, `required: [expression, language]`). The `elasticstack_kibana_dashboard` resource instead exposes `text` and `json`, both of which map onto that same `Expression` field on write (`dashboardQueryToAPI`, `internal/kibana/dashboard/models.go`), while read sniffs the returned string for a leading `{` to decide which of `text` / `json` to populate. `json` therefore has no real API counterpart: Kibana always treats the value as a KQL/Lucene query string, never as structured JSON.

This is reported in issue #5081: a dashboard exported from Kibana as JSON has `query.expression`, but the equivalent HCL needs `query.text`, so a practitioner (or a conversion tool) has to translate the field name. Every other query block in this resource — Lens chart queries (`lenscommon/filter_simple.go`), discover-session queries, and all the typed `vis` panel `query` blocks — already uses `expression`. The root `query` block is the one place that doesn't, and it is also the only place where a lossy heuristic (JSON-object sniffing) decides what the practitioner meant.

## What Changes

- Root `query` becomes `{ language (required), expression (required) }`. `text` and `json` are removed, along with their `ExactlyOneOf` validators and the "exactly one of `query.text` or `query.json`" diagnostic.
- Write mapping (`dashboardQueryToAPI`) becomes a straight copy of `Expression` and `Language`; the JSON-object write branch is deleted.
- Read mapping (`models.go`) sets `query.expression` directly from the API `Expression` string; the `{`-prefix JSON-sniffing heuristic is deleted.
- **Breaking configuration change.** Existing `.tf` files using `query.text` or `query.json` must be updated to `query.expression`. This is intentional: the dashboard resource is Kibana "Technical Preview" (9.4+).
- **State migrates automatically.** The resource schema version moves from 1 to 2. A new `migrateV1ToV2` upgrader copies `query.text` (or, if `text` is absent, `query.json`) into `query.expression` unchanged, and removes `text`/`json` from the stored map. The existing v0 → v1 upgrader (`migrateV0ToV1`, REQ-040) is composed with the new transform so v0 state upgrades directly to v2 in one step. No manual `terraform state` surgery is required.
- `openspec/specs/kibana-dashboard/spec.md` is updated: the REQ-036 query-union scenarios are replaced with an `expression` scenario (including empty-string expression, the exact case from the issue), REQ-007's query-mapping sentence is updated, and a new requirement describes the v1 → v2 (and composed v0 → v2) upgrade.
- Tests, examples, docs, and the CHANGELOG are updated to match (see `tasks.md`).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: root `query` attribute shape changes from `{ language, text?, json? }` to `{ language, expression }`; schema version 1 → 2 with an automatic state upgrade.

## Impact

- `internal/kibana/dashboard/schema.go` (query attribute definitions, schema `Version`).
- `internal/kibana/dashboard/models.go` (`DashboardQueryModel`, `dashboardQueryToAPI`, read-side query mapping).
- `internal/kibana/dashboard/state_upgrade.go` (new `migrateV1ToV2`, updated `UpgradeState` map).
- `openspec/specs/kibana-dashboard/spec.md` (REQ-036, REQ-007, new state-upgrade requirement).
- Unit tests: `models_dashboard_root_test.go`, `models_optional_root_blocks_test.go`, `create_test.go`, `pinned_panels_mapping_test.go`, `state_upgrade_test.go`.
- Acceptance tests: `TestAccResourceDashboardRootQueryJSON` deleted; `TestAccResourceDashboardQueryTransition` rewritten; new upgrade-path acceptance test; ~20 `testdata/**/main.tf` files with `text =` renamed to `expression =`.
- `docs/resources/kibana_dashboard.md` (regenerated) and `examples/` using `query.text` or the JSON-query example.
- `CHANGELOG` breaking-change entry referencing issue #5081.

## Out of scope

- Other `query` shapes in the dashboard resource (Lens / discover-session / typed `vis` panel blocks already use `expression` and are unaffected).
- Any change to the generated Kibana client; the upstream OpenAPI schema already matches the target shape (`expression`/`language`), so no `generated/kbapi` regeneration is required.
- A future Kibana release reintroducing an object-valued query would be handled as a new attribute, not by overloading `expression` again.
