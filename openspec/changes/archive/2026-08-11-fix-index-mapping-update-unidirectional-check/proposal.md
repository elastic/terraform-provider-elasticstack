## Why

Adding a new field to `elasticstack_elasticsearch_index.mappings` on an existing index silently
does nothing: `terraform apply` reports `1 changed` and completes with no error, but the field
never reaches the live cluster (`elastic/protections-cloud#19769`, confirmed on Serverless v0.16.3).

`updateMappings` (`internal/elasticsearch/index/index/update.go:194`) decides whether to call the
Put Mapping API using `planMappings.StringSemanticEquals(ctx, stateMappings)`. That method
(`internal/elasticsearch/index/mappings_value.go:154`) is intentionally **bidirectional**:
`MappingsSemanticallyEqual(vMap, newMap) || MappingsSemanticallyEqual(newMap, vMap)`. The
bidirectional form exists to suppress plan-time drift when the live cluster (or a matching index
template) is a superset of user intent. Reused as an apply-time update-decision gate, the same
`||` also treats "plan is a superset of state" — i.e. the user added a field — as "equal", so the
Put Mapping API is never called and the addition is lost.

`create.go:183` (`adoptExistingIndexOnCreate`, the `use_existing` adoption path) calls the same
`updateMappings` helper, so the bug is present on both the Update and Create-adoption code paths.

## What Changes

- Add a `RequiresMappingsUpdate` method on `MappingsValue` that makes a **unidirectional**
  update-decision: it returns `true` only when the planned mappings contain content not already
  present in the prior state mappings. It returns `false` when state is already a non-drifting
  superset of plan (nothing new to write), which preserves today's tolerance for
  Elasticsearch/template-injected extras.
- Replace `planMappings.StringSemanticEquals(ctx, stateMappings)` with
  `planMappings.RequiresMappingsUpdate(ctx, stateMappings)` at both call sites that decide whether
  to invoke the Put Mapping API: `update.go`'s `updateMappings` and `create.go`'s
  `adoptExistingIndexOnCreate` (via the same shared `updateMappings` helper).
- Leave `StringSemanticEquals` and `MappingsSemanticallyEqual` unchanged. They continue to back
  plan-time semantic equality and the `RequiresReplace` decision, where the existing bidirectional
  drift-suppression behavior is correct and required (see REQ-022–REQ-024).
- Add a unit test for `RequiresMappingsUpdate` covering: plan adds a field (true), plan is a
  strict superset of state (true), state is a superset of plan / template-injected extras only
  (false), plan equals state (false), and null/unknown plan or state.
- Add an acceptance test that manages an existing index, adds a field to `mappings`, applies, and
  verifies the field is present in the live cluster's mapping — the regression test for this issue.
- Extend acceptance coverage for the `use_existing` adoption path (`create.go:182-187`) with a
  config that both adds a field absent from the live mapping and omits a field present only in the
  live mapping, to pin the asymmetry the fix depends on (adopted fields not in config are kept;
  fields added in config are written).

## Capabilities

### Modified Capabilities

- `elasticsearch-index`: Update flow (REQ-015–REQ-018) — the Put Mapping API decision becomes
  unidirectional (state-covers-plan) instead of the bidirectional semantic-equality check. Opt-in
  adoption of existing indices via `use_existing` — the same unidirectional decision applies when
  reconciling mappings during adoption.

## Impact

- `internal/elasticsearch/index/mappings_value.go` — add `RequiresMappingsUpdate`.
- `internal/elasticsearch/index/mappings_value_test.go` — unit tests for the new method.
- `internal/elasticsearch/index/index/update.go` — call `RequiresMappingsUpdate` instead of
  `StringSemanticEquals` in `updateMappings`.
- `internal/elasticsearch/index/index/create.go` — no code change expected (calls the same
  `updateMappings` helper), but adoption behavior changes as a result.
- `internal/elasticsearch/index/index/acc_test.go` — new acceptance test(s) for adding a mapping
  field to an existing/adopted index.
- `openspec/specs/elasticsearch-index/spec.md` — update the Update flow and adoption requirements
  to describe the unidirectional decision.
