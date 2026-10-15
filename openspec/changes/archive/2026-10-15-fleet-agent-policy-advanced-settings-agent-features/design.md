## Context

`elasticstack_fleet_agent_policy`'s `advanced_settings` nested attribute already exposes 10 of the 13
properties on Kibana's agent-policy `advanced_settings` sub-object (logging, download, `go_max_procs`,
`monitoring_runtime_experimental`); three remain unmapped:
`agent_features_disable_policy_change_acks_enabled`, `agent_features_include_tags_in_events_enabled`,
and `agent_internal`. All three already exist as `interface{}`/`any` fields on the generated client
and on the provider's shadow struct `advancedSettingsAPIValues`
(`internal/fleet/agentpolicy/models_advanced_settings.go:51-65`) — the gap is purely in the
Terraform schema/model layer. `advanced_settings` as a whole is gated at `MinVersionAdvancedSettings`
(8.17.0, `resource.go:43`), enforced via `entitycore.EnforceVersionRequirements` at apply time in
`Create`/`Update` (not at plan time; `ModifyPlan` only checks tamper protection).

## Deviation from the issue's implementation-research comment

The issue carries an automated implementation-research comment (workflow run `38666263036`) whose
`### Recommendation` scoped this change to **only** the two `*_enabled` boolean fields, left
`agent_internal` out of scope as unconfirmed, and flagged three items as **blocking**: the wire type
of the two `*_enabled` fields (hypothesis only, "a bare API round-trip probe is not sufficient
evidence"), `agent_internal`'s shape and intent (no source at all), and the true minimum stack
version for each field (unconfirmed; the research explicitly rejected reusing the shared 8.17.0
`MinVersionAdvancedSettings` floor as unvalidated).

After that research ran, the issue reporter (`@tobio`, the same actor who triggered this workflow)
posted a follow-up comment, "Research follow-up: wire types, versions, and open questions resolved,"
that explicitly resolves every blocking item the research comment raised, with stronger evidence than
the research comment required:

- **Wire types**, sourced from Kibana's `common/settings/agent_policy_settings.tsx` (the `api_field`
  + zod `schema` definitions that drive `server/services/form_settings/form_settings.ts`'s request
  validation), not a bare round-trip probe: the two `*_enabled` fields are
  `z.boolean().default(false)` — live probes against a 9.4.0 stack show the server rejects
  non-boolean values (`"true"` as a string → 400), which is the "evidence that the server enforces a
  given shape" class of proof the research comment asked for. `agent_internal` is
  `type: 'yaml'`, a YAML string the server validates by attempting to parse it, confirmed by 400s on
  non-string and invalid-YAML input and verbatim (non-normalized) echo of valid YAML on read.
- **`agent_internal`'s shape and intent**: it is the Fleet UI's "Advanced internal YAML settings"
  field (explicitly user-settable, with a documented "no stability guarantee" caveat), not an
  internal-only field — resolving the research comment's other blocking question.
- **Minimum versions**, sourced from Kibana PR milestone labels and spot-checked live: `agent_internal`
  9.3.0 (kibana#245819), `agent_features_disable_policy_change_acks_enabled` 9.4.0 (kibana#257275,
  accepted live on 9.4.0), `agent_features_include_tags_in_events_enabled` 9.6.0 (kibana#278475,
  confirmed **rejected** live on 9.4.0 with `Additional properties are not allowed`). The comment
  explicitly confirms the research's suspicion that the shared 8.17.0 `MinVersionAdvancedSettings`
  floor would have been wrong for all three.
- **Naming**: confirms the `agent_` prefix should be stripped (maintainer decision), matching the
  research's own naming proposal.
- **Presence semantics**: Kibana returns `null`, not a server default, when any of the three is
  unset — so this design uses plain `Optional` (not `Optional+Computed`) for all three, which the
  follow-up comment notes keeps "absent" and "null" identical without a planned-unknown conflict.

This is a direct, explicit resolution of the research comment's own named blocking conditions, sourced
from Kibana's source code and live-probe evidence that meets the bar the research comment itself set
("UI-sourced evidence, a behavioural check, or direct maintainer/Kibana-team confirmation" — the zod
schema plus server-side rejection behavior is exactly the "stronger than round-trip" confirmation
requested). Per this workflow's contradiction rule, this design adopts the follow-up comment's
resolved scope and data over the research comment's narrower, explicitly-hedged recommendation:
**all three fields are in scope**, each with its own confirmed minimum version, rather than deferring
`agent_internal` to a separate follow-up.

The research comment's structural guidance that is **not** contradicted — the per-attribute mapping
pattern, the independently-adjustable per-field version-requirement mechanism (rather than reusing
the shared floor), and the test extension strategy — is retained unchanged below.

## Goals / Non-Goals

**Goals:**
- Map all three remaining `advanced_settings` keys so the resource's coverage of the API sub-object
  is complete.
- Reuse the resource's existing per-attribute pattern (schema attribute → model field → shadow
  struct → populate/convert functions → attribute-scoped version requirement → capability flag)
  exactly, rather than introducing a new mechanism.
- Gate each field at its own confirmed minimum version, independent of the shared 8.17.0
  `advanced_settings` floor and of each other.
- Reuse the existing `customtypes.NormalizedYamlType`/`NormalizedYamlValue` string type for
  `internal` instead of adding a new YAML-aware type, since one already exists in
  `internal/utils/customtypes` and is unused by any resource today.

**Non-Goals:**
- Changing any of the 10 already-exposed `advanced_settings` keys.
- Adding a passthrough/dynamic escape hatch for future undocumented `advanced_settings` keys (the
  research comment's Approach B) — all three relevant OpenAPI schemas declare
  `additionalProperties: false`, so a passthrough attribute would only ever cover keys Kibana's
  schema already recognizes, buying little over named attributes.
- Regenerating `generated/kbapi` — all three fields are already present as `interface{}` on every
  relevant generated struct.
- Any provider code change in this proposal-only change: this document specifies the implementation
  for a follow-up apply step, per this workflow's scope.

## Decisions

### Decision 1: Three new named attributes inside the existing `advanced_settings` block

Add to `internal/fleet/agentpolicy/schema.go` (alongside `monitoring_runtime_experimental`):

| TF attribute | Type | API key |
|---|---|---|
| `features_disable_policy_change_acks_enabled` | `schema.BoolAttribute`, Optional | `agent_features_disable_policy_change_acks_enabled` |
| `features_include_tags_in_events_enabled` | `schema.BoolAttribute`, Optional | `agent_features_include_tags_in_events_enabled` |
| `internal` | `schema.StringAttribute` with `CustomType: customtypes.NormalizedYamlType{}`, Optional | `agent_internal` |

Add matching fields to `advancedSettingsModel` (`models_advanced_settings.go:33-44`):
`FeaturesDisablePolicyChangeAcksEnabled types.Bool`, `FeaturesIncludeTagsInEventsEnabled types.Bool`,
`Internal customtypes.NormalizedYamlValue`. Extend `hasValues` (lines 108-117) to include all three.
Populate via `typeutils.BoolFromAny(...)` for the two booleans and a string-from-`any` equivalent for
`Internal` in `populateAdvancedSettingsFromAPI` (lines 68-93). Emit via `.ValueBool()` /
`.ValueString()` guarded by `typeutils.IsKnown` in `convertAdvancedSettingsToAPI` (lines 96-157). No
change to `advancedSettingsAPIValues` (`models_advanced_settings.go:51-65`) is needed — the JSON keys
already exist there.

All three are `Optional` only, not `Optional+Computed`: per the follow-up comment, Kibana returns
`null` (not a computed default) when any is unset, so "absent" and "null" stay identical and there is
no planned-unknown value for the provider to resolve.

**Why:** Matches the resource's existing one-named-typed-attribute-per-key pattern exactly (as
`monitoring_runtime_experimental` does); the shadow struct and generated client already carry the
JSON keys, so this is purely additive at the schema/model layer.

### Decision 2: `internal` reuses `customtypes.NormalizedYamlType`, not a new type

`internal/utils/customtypes/normalized_yaml_type.go` already implements a `basetypes.StringTypable`
with semantic equality that ignores insignificant whitespace and key-ordering differences — exactly
what `agent_internal` needs, since Kibana echoes a submitted YAML string back verbatim (no
normalization) and the follow-up comment confirms heredoc-formatting drift must not show as a diff.
This type is defined but not yet used by any resource in `internal/fleet`, so this change is its
first consumer. Using it avoids adding a second YAML-aware string type to the codebase.

**Why:** Avoids duplicating an existing, already-tested type; keeps `internal`'s plan/state behavior
consistent with how the provider already models "YAML text that must compare semantically, not
byte-for-byte."

### Decision 3: Three independent attribute-scoped version requirements, not the shared `advanced_settings` floor

Reusing the shared `MinVersionAdvancedSettings` (8.17.0) for any of the three would under-gate it: the
follow-up comment confirms `agent_internal` requires 9.3.0,
`features_disable_policy_change_acks_enabled` requires 9.4.0, and
`features_include_tags_in_events_enabled` requires 9.6.0 — each a true version gap above 8.17.0, with
the live probe confirming 9.4.0 actively rejects the `include_tags_in_events` key with "Additional
properties are not allowed."

Add, in `internal/fleet/agentpolicy/resource.go` (alongside `MinVersionAdvancedSettings`):

```go
MinVersionAgentInternal                            = version.Must(version.NewVersion("9.3.0"))
MinVersionFeaturesDisablePolicyChangeAcksEnabled    = version.Must(version.NewVersion("9.4.0"))
MinVersionFeaturesIncludeTagsInEventsEnabled        = version.Must(version.NewVersion("9.6.0"))
```

In `version_requirements.go`'s `GetVersionRequirements`, after decoding `settings` from
`model.AdvancedSettings` (lines 111-115), add one `entitycore.NewAttributeVersionRequirement` call
per field guarded by `typeutils.IsKnown(settings.<Field>)`, following the existing
`NewAttributeVersionCheckRequirement` call for `monitoring_runtime_experimental` (lines 117-122) as
the placement precedent — a plain `NewAttributeVersionRequirement` (single minimum version) suffices
here, since none of the three needs `monitoring_runtime_experimental`'s more complex "8.19.x OR
9.1.0+" predicate.

In `capabilities.go`, add three matching `Supports...` flags to `agentPolicyFeatures`, resolved via
`client.EnforceMinVersion` in `resolveAgentPolicyFeatures`, following the existing
`SupportsAdvancedSettings`/`SupportsMonitoringRuntimeExperimental` pattern. `convertAdvancedSettingsToAPI`
only emits each field's JSON key when its capability flag is true, mirroring how
`monitoring_runtime_experimental` is already gated in that function.

**Why:** Matches the existing per-attribute version-gating mechanism exactly; avoids both
under-gating (permitting a setting against a stack that will 400) and over-gating (blocking the two
earlier fields on a floor meant for the latest one).

### Decision 4: Capability flags gate the JSON key, not just the schema attribute

Because Fleet's `PUT` replaces `advanced_settings` wholesale when the object is present
(confirmed by the follow-up comment's probe: sending only `acks=false` dropped `agent_internal`), a
field whose capability flag is false must be omitted from the request body even if a stale value is
present elsewhere in state, consistent with how `monitoring_runtime_experimental` is already handled
in `convertAdvancedSettingsToAPI`. This is an existing pattern in that function, applied unchanged to
the three new fields.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| The two `*_enabled` fields' boolean type ultimately rests on Kibana's zod schema source, not a published OpenAPI type (the spec itself only says `nullable: true`) | The follow-up comment's evidence (zod `z.boolean()`, server-side 400 on non-boolean input) is the strongest confirmation available short of an Elastic maintainer sign-off on the PR; treat as resolved for this change. |
| `agent_internal` has no stability guarantee per Kibana's own UI copy, and PR kibana#286410 already hid `disable_policy_change_acks` from the Fleet UI for being confusing | Document `internal`'s "no stability guarantee" caveat directly in the attribute's schema description (see tasks.md docs task); this does not block mapping the field, since the API itself is unchanged. |
| 9.3.0/9.6.0 minimums are sourced from Kibana PR milestone labels, not live-verified beyond the 9.4.0 stack the follow-up comment's probes ran against | Tasks.md includes a pre-merge verification task to confirm both boundaries against real stacks before the acceptance tests are considered complete, consistent with the follow-up comment's own caveat ("Not live-verified... worth a CI acceptance run against those versions via the `SkipFunc` pattern"). |

## Open questions

Carried over verbatim from the issue's implementation-research comment, to the extent not already
resolved by the follow-up comment above:

- Are there behavioural caveats for the two `*_enabled` fields (e.g. does disabling policy-change
  acks affect agent health/status tracking in Fleet UI)? Docs are silent. Non-blocking for
  implementation; the follow-up comment adds that `disable_policy_change_acks` is hidden in the
  Fleet UI since kibana#286410 for being confusing to customers, and that `include_tags_in_events`
  additionally depends on future Fleet Server/Elastic Agent changes to have effect — worth a doc
  note once fully characterized, not blocking.
- The 9.3.0 (`agent_internal`) and 9.6.0 (`features_include_tags_in_events_enabled`) minimums rely on
  Kibana PR milestone labels rather than a second live stack probe (only 9.4.0 was available to the
  follow-up comment's author). Non-blocking for authoring this change; tracked as a pre-merge
  verification task in tasks.md.

## Pre-implementation verification record (tasks 1.1-1.3)

Only one live stack was reachable during implementation: Elasticsearch/Kibana **9.5.5** on
`localhost:9200`/`localhost:5601` (the worktree's env-var ports were not listening, so the default
ports were used). No 9.3.0 or 9.6.0 stack was started. Findings against 9.5.5 via direct Fleet API
probes (`POST /api/fleet/agent_policies`):

- `advanced_settings.agent_internal` (`"a: 1"`) and
  `agent_features_disable_policy_change_acks_enabled` (`true`) are accepted and echoed back.
  This is consistent with, but does not prove, the 9.3.0 / 9.4.0 minimums.
- `agent_features_include_tags_in_events_enabled` is rejected with
  `Additional properties are not allowed`. This is consistent with the 9.6.0 minimum (the
  rejection reproduces on 9.4.0 per the issue and now also on 9.5.5, which is below 9.6.0).

Not verified live: the exact 9.3.0 lower boundary for `agent_internal`, acceptance of
`features_include_tags_in_events_enabled` on 9.6.0, and rejection of `agent_internal` below 9.3.0.
No deviation from the documented 9.3.0 / 9.4.0 / 9.6.0 minimums was found; implementation
proceeds with those minimums. Acceptance tests use per-field `SkipFunc`s so CI stacks at other
versions exercise the remaining boundaries.

## Implementation notes and deviations (acceptance tests)

- Because the three fields have different minimums (9.3.0 / 9.4.0 / 9.6.0), the acceptance tests add
  new per-field steps (`agent_internal_create`, `agent_acks_create`, `agent_features_create`,
  `agent_internal_reformatted`, `agent_features_update`, `agent_features_removed`) to
  `TestAccResourceAgentPolicyWithAdvancedSettings`, each with its own `SkipFunc`, instead of editing
  the existing `create_with_logging`/`update_settings` configs (which must keep running on 8.17.0+).
  Negative gating lives in `TestAccResourceAgentPolicyAgentFeaturesVersionGate`.
- Semantic equality: Terraform requires the planned value of an `Optional` (non-`Computed`)
  attribute to equal its configured value, so the framework cannot suppress the plan diff when only
  the YAML formatting in configuration changes. The result is a single in-place update that leaves no
  persistent diff; semantic equality still applies on refresh/apply (e.g. if Kibana normalised the
  YAML). The delta spec scenario was reworded accordingly, and acceptance step 10 asserts a clean
  apply with an empty post-apply plan rather than an empty pre-apply plan.

## Validation record (group 9)

Acceptance tests (`TestAccResourceAgentPolicy*`) were run against the 9.5.5 stack only. On it, the
`internal` (9.3.0) and `features_disable_policy_change_acks_enabled` (9.4.0) positive steps and the
`features_include_tags_in_events_enabled` negative version-gate step executed; the 9.6.0 positive
steps and the lower-version negative steps for the other two fields were skipped by their `SkipFunc`s
and remain to be exercised by CI stacks at those versions. `make check-lint`, `make check-openspec`
and the `internal/fleet/agentpolicy` unit tests pass. Self-review against the delta spec (task 9.5)
found the implementation consistent with every scenario; the only deviation is the semantic-equality
wording noted above.
- `convertAdvancedSettingsToAPI` returns nil when no emittable field remains (for example only
  capability-unsupported fields are set), so `advanced_settings` is omitted from the request instead of
  being sent as `{}`, which Fleet's wholesale PUT would treat as clearing every key. This is unreachable
  on stacks where version-requirement diagnostics already stop the apply; it is a defensive guard covered
  by a unit test.
- `internal` population deviates from `typeutils.StringFromAny`: Kibana can return structured YAML as a
  JSON object, so non-string values are JSON-marshalled (JSON is valid YAML) rather than read as null,
  which would cause a permanent diff. A marshal failure surfaces as a warning diagnostic rather than
  being silently dropped.
