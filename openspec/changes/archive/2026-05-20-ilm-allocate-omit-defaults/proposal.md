## Why

When a user declares an `allocate` block that contains **only** routing-filter attributes (`require`, `include`, or `exclude`) and omits `number_of_replicas` and/or `total_shards_per_node`, the provider silently writes those fields to the Elasticsearch ILM API using their schema-default values (`number_of_replicas = 0`, `total_shards_per_node = -1`). This unintentionally overrides any replica or shard-per-node settings configured in the index template for that phase, contrary to the ILM Allocate action contract, which treats all five parameters as independently optional.

There are two root causes:

1. **Schema defaults** — `number_of_replicas` and `total_shards_per_node` carry `Default: int64default.StaticInt64(0)` / `Default: int64default.StaticInt64(-1)` plus `Computed: true` in `schema_actions.go`. The Plugin Framework fills those defaults in before the expand path runs, so even an omitted attribute arrives at `expandAction` with a concrete non-null value and gets forwarded to the API.
2. **Forced flatten default** — `flatten.go` unconditionally sets `allocateAction["total_shards_per_node"] = int64(-1)` when the API response omits the field, causing perpetual round-trip drift independent of the schema-default issue.

## What Changes

- **`schema_actions.go`**: Remove `Default:` lines from `number_of_replicas` and `total_shards_per_node`; add `PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}` to both; keep `Optional: true, Computed: true`. Update the `number_of_replicas` description to remove the now-inaccurate "Default: `0`" text.
- **`flatten.go`**: Remove the `else` branch that forces `allocateAction["total_shards_per_node"] = int64(-1)` when the field is absent from the API response. The existing `nullValueForType` path already returns `types.Int64Null()` for absent keys when the field is not in the map.
- **`expand.go`**: Remove `def: -1` from the `total_shards_per_node` entry in `ilmActionSettingOptions`; it is dead code (`total_shards_per_node` has no `minVersion` so the version-gate path never uses `def`). `skipEmptyCheck: true` is preserved so explicit `0` and `-1` values still reach the API.
- **`expand_test.go`**: Add a regression test case verifying that an `allocate` block with only a routing filter and no replica/shard fields produces an expand map that omits `number_of_replicas` and `total_shards_per_node`.

No state schema version bump is needed. The `UseStateForUnknown` plan modifier ensures that existing resources that already have `0` / `-1` in state produce **no plan diff** after upgrade. New resources created after the fix will omit these fields from the API payload unless the user explicitly sets them.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`elasticsearch-index-lifecycle`**: `allocate` action schema — `number_of_replicas` and `total_shards_per_node` become genuinely optional (no Default injection); read-back of API responses no longer forces a synthetic `-1` for `total_shards_per_node` when absent.

## Impact

- **Users**: Existing resources with `0` / `-1` already in state are unaffected on upgrade (the `UseStateForUnknown` modifier preserves those values until the user explicitly removes them or recreates the resource). New resources and resources being recreated will no longer have replica/shard settings unexpectedly applied to the index.
- **Code**: Surgical changes to three files in `internal/elasticsearch/index/ilm/` plus one unit test addition. No state migration required.
