## Why

Renovate's bump of `generated/kbapi`'s pinned Kibana OpenAPI ref (#5464, `23436cd` → `cd914df`) cannot
merge: regenerating the client against the new spec breaks the build in two independent ways. First,
a new `Cases_attachment_type_filter` query parameter is an inline `oneOf` of a string and a string
array that `oapi-codegen` cannot name, emitting `AsN0()`/`FromN0()` methods that reference undefined
`N0`/`N1` types. Second, Kibana retired the hand-written SLO OpenAPI spec in favor of one generated
from route schemas (elastic/kibana#300516), so every `kbapi.SLOs*` Go type the provider depends on is
replaced by a `kbapi.KibanaHTTPAPIsSLO*`-prefixed type, and — confirmed by regenerating locally against
`cca2b0f` — the replacement is not a pure rename: `TimeWindow` and `KqlWithFilters` change shape, APM
indicator params become `oneOf` union wrappers instead of plain structs, and the existing
`fixSloFloatFormats` transform silently stops applying `format: double` because it still targets the
old `SLOs_*` component paths, which would reintroduce the float32 precision loss originally fixed for
issue #2396. About 19 files and roughly 330 occurrences of `kbapi.SLOs*` under `internal/` are affected.
Until both breakages are fixed, the bump PR stays red and Renovate cannot resume tracking Kibana's spec.

## What Changes

- Pin `generated/kbapi/Makefile`'s `github_ref` to one explicit, immutable target Kibana `main` commit
  SHA (not a moving "latest main") and regenerate (`make -C generated/kbapi all`) against that exact
  ref, so every claim below is verified against one fixed spec rather than assumed.
- Add a `fixCasesAttachmentTypeFilterParameter` transform in `generated/kbapi/transform_schema.go` that
  replaces the `Cases_attachment_type_filter` parameter's inline `oneOf` with refs to the existing
  `Cases_string` / `Cases_string_array` component schemas, after diffing the target spec's inline array
  schema against `Cases_string_array`'s current `maxItems: 100` bound so the fix does not silently
  tighten or loosen the parameter's contract.
- Add a regeneration-time guard, run before the transform step consumes the downloaded spec, that fails
  loudly if a schema-name-dependent transform's source node (the `Cases_attachment_type_filter`
  parameter, and the `SLOs_*`/replacement component paths `fixSloFloatFormats` and
  `fixSloResponseArtifacts` target) is absent from the just-downloaded spec — because `Map.Set` creates
  missing paths rather than erroring, today's transforms can silently no-op against a renamed schema
  with no build failure.
- Update `fixSloFloatFormats` and `fixSloResponseArtifacts` to target the regenerated spec's actual SLO
  component paths (replacing the hardcoded `SLOs_*` paths), so `objective.target`,
  `objective.timesliceTarget`, and the other fields listed in `fixSloFloatFormats` keep resolving to
  Go `float64` rather than silently reverting to `float32`, and the `artifacts` property keeps appearing
  on the SLO response schema.
- Migrate `internal/models/slo.go` and the SLO resource/model/client code under `internal/kibana/slo`
  and `internal/clients/kibanaoapi/slo.go` from `kbapi.SLOs*` to the regenerated `kbapi.KibanaHTTPAPIsSLO*`
  equivalents, including the shape changes confirmed by the local regeneration:
  - Rewrite the `TimeWindow` and `KqlWithFilters` mapping code for their new field/union shape rather
    than treating the rename as type-name-only.
  - Rewrite APM latency and APM availability indicator mapping (`models_apm_latency_indicator.go`,
    `models_apm_availability_indicator.go`) to construct and decode a `oneOf` union wrapper (matching
    the existing `From*`/`As*` pattern already used for `kql_custom_indicator`'s `KqlWithFilters` union)
    instead of assigning params as a plain anonymous struct literal.
  - Preserve the same logical Terraform state and wire JSON for every existing indicator type,
    `group_by` encoding, `settings`, and `artifacts`/dashboards handling — this is a compile-time type
    migration with confirmed shape changes in specific fields, not a behavior change.
- Add generator-level test coverage: a table-driven fixture test for the new
  `fixCasesAttachmentTypeFilterParameter` transform, and fixture tests for `fixSloFloatFormats` /
  `fixSloResponseArtifacts` (currently untested), following the existing pattern in
  `generated/kbapi/transform_schema_test.go`.
- Defer adding support for the newly-documented `sli.synthetics.availability` indicator type to a
  separate follow-up issue: its minimum Kibana version and exact wire shape are not confirmed by any
  source available to this proposal, and the issue itself offers this as an explicit escape hatch
  ("If that is too large, split it into a follow-up and say so here"). The provider's behavior toward
  this indicator type is unchanged by this proposal (still unsupported, same as today).
- Once the bump lands and passes CI, close or supersede Renovate PR #5464 and confirm Renovate resumes
  bumping the pinned ref cleanly — tracked here as a task, not part of the capability requirements below.

## Capabilities

### New Capabilities

- `kbapi-generated-client`: Defines the observable contract for `generated/kbapi`'s OpenAPI-spec-driven
  code generation pipeline — the pinned target ref, the `Cases_attachment_type_filter` transform, and
  the requirement that schema-name-dependent transforms fail the regeneration rather than silently
  no-op when their source node is missing from the downloaded spec.

### Modified Capabilities

- `kibana-slo`: Updates the existing kbapi-type-binding requirements (REQ-032) to reference
  `kbapi.KibanaHTTPAPIsSLO*` types instead of `kbapi.SLOs*`, and adds requirements that the migration
  preserves `float64` precision for objective/threshold fields, preserves the existing JSON wire shape
  for `time_window`, `kql_custom_indicator` filters, and APM indicator params despite the underlying
  Go type shape changes, and leaves `sli.synthetics.availability` unsupported exactly as before.

## Impact

- **Changed code**: `generated/kbapi/Makefile` (pinned `github_ref`), `generated/kbapi/transform_schema.go`
  (`fixCasesAttachmentTypeFilterParameter`, updated `fixSloFloatFormats`/`fixSloResponseArtifacts`, new
  pre-transform guard), the regenerated `generated/kbapi/kibana.gen.go` and `generated/kbapi/oas.yaml`/
  `oas-filtered.yaml` byproducts, `internal/models/slo.go`, `internal/kibana/slo/*.go` (schema, models,
  create/read/update, every per-indicator mapping file and its tests), `internal/clients/kibanaoapi/slo.go`.
- **New tests**: `generated/kbapi/transform_schema_test.go` gains fixture tests for the three transforms
  above; existing SLO unit tests (`models_api_mapping_test.go`, per-indicator `*_test.go` files,
  `internal/clients/kibanaoapi/slo_test.go`) are updated to compile against the renamed types and extended
  where the TimeWindow/KqlWithFilters/APM-union shape changes need new coverage; existing SLO acceptance
  tests (`internal/kibana/slo/acc_test.go`) continue to exercise every indicator type, `group_by` variant,
  and `objective.target` precision end to end against a real Kibana instance.
- **Docs**: none required — this is an internal client/type migration with no schema or behavior change
  visible to practitioners; no `elastic/kibana` or provider documentation changes are in scope.
- **Out of scope**: `sli.synthetics.availability` indicator support (tracked as a follow-up issue), any
  other new SLO indicator type, and Renovate configuration changes beyond closing/superseding #5464.
- **Backward compatibility**: purely internal — no Terraform schema, state shape, or HCL-visible behavior
  changes. The risk is a silent regression (e.g. `objective.target` precision, as in #2396) if the
  renamed schema paths are not correctly re-targeted; this proposal's guard transform and fixture tests
  exist specifically to catch that class of regression during regeneration rather than in production.
