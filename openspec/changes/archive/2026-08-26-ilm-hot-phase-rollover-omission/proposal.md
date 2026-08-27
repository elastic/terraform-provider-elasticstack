## Why

Some index types (for example `lookup` indices) do not support rollover, so an ILM policy for them must have a `hot` phase with no `rollover` action. Practitioners report that `elasticstack_elasticsearch_index_lifecycle` makes this impossible in practice: even when the Terraform configuration omits the `rollover {}` block entirely, the provider behaves as though a rollover were configured, producing a perpetual diff on every plan.

Root cause: the `rollover` block is already `Optional` in the schema (`internal/elasticsearch/index/ilm/schema_actions.go`), and the expand path already omits `rollover` from the API payload when the user has not configured it. The bug is in the flatten/read path (`internal/elasticsearch/index/ilm/flatten.go`), whose `default:` case writes any action Elasticsearch returns — including an empty `"rollover": {}` — directly into state as a non-null object. That is semantically different from the `ObjectNull` value Terraform expects when the user's config omitted the block, so every subsequent plan shows a diff.

## What Changes

- Add an explicit `rollover` case in `flattenPhase` (`internal/elasticsearch/index/ilm/flatten.go`) that treats an empty rollover action (`len(action) == 0`) returned by Elasticsearch as absent, unless the prior state already declared a non-null `rollover` block — mirroring the existing `priorHasDeclaredToggle` pattern used for `readonly`, `freeze`, and `unfollow`.
- Add a `priorHasDeclaredAction` helper (or generalize `priorHasDeclaredToggle`) so the guard can be reused for `rollover`.
- No schema change: `rollover` stays `Optional`, no new attributes, no state version bump.
- Add acceptance test coverage for creating and refreshing a `hot` phase with no `rollover` block, confirming no perpetual diff.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-index-lifecycle`: extend read-state normalization so an empty `rollover: {}` returned by Elasticsearch for a `hot` phase is treated as absent state when the user did not declare `rollover` in configuration or prior state.

## Impact

- `internal/elasticsearch/index/ilm/flatten.go` — add the `rollover` case and the prior-declaration guard helper.
- `internal/elasticsearch/index/ilm/flatten_test.go` (or equivalent unit test file) — add unit coverage for the new flatten case.
- Acceptance tests for `elasticstack_elasticsearch_index_lifecycle` — add a test creating a policy whose `hot` phase has no `rollover` block and asserting no diff after refresh.
- `openspec/specs/elasticsearch-index-lifecycle/spec.md` — add a requirement documenting this normalization.
