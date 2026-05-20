## 1. Schema fix (`schema_actions.go`)

- [ ] 1.1 In `blockAllocate()`, remove `Default: int64default.StaticInt64(0)` from `number_of_replicas`.
- [ ] 1.2 Add `PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}` to `number_of_replicas`.
- [ ] 1.3 Update the `number_of_replicas` `Description` field: remove the sentence fragment `"Default: \`0\`"` so it accurately reflects the new optional-only semantics.
- [ ] 1.4 In `blockAllocate()`, remove `Default: int64default.StaticInt64(-1)` from `total_shards_per_node`.
- [ ] 1.5 Add `PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}` to `total_shards_per_node`.
- [ ] 1.6 Remove the now-unused `int64default` import from `schema_actions.go` if no other attribute in that file uses it (check `blockDeleteAction`, `blockFreeze`, `blockMigrate`, etc. — they use `booldefault`, not `int64default`).

## 2. Flatten fix (`flatten.go`)

- [ ] 2.1 In `flattenPhase`, case `"allocate"`, remove the `else` branch:
  ```go
  } else {
      allocateAction["total_shards_per_node"] = int64(-1)
  }
  ```
  When the API response omits `total_shards_per_node`, the key should simply not be added to `allocateAction`. The downstream `phaseDataToObjectAttrs` / `nullValueForType` path already returns `types.Int64Null()` for absent keys.

## 3. Expand cleanup (`expand.go`)

- [ ] 3.1 In `ilmActionSettingOptions`, change the `total_shards_per_node` entry from `{skipEmptyCheck: true, def: -1}` to `{skipEmptyCheck: true}`. The `def: -1` field was dead code (no `minVersion` means the version-gate path never reads `def` for this setting). Keep `skipEmptyCheck: true` so explicit `0` and `-1` values still reach the API.

## 4. Unit test (`expand_test.go`)

- [ ] 4.1 Add a test case to `TestExpandAction` (or a companion test function) that exercises the allocate-filter-only path:
  - Input: `allocate` action with only `require` set (e.g. `{"zone": "zone-1"}`); `number_of_replicas` and `total_shards_per_node` absent from the action map (i.e. null/not present).
  - Expected output: expanded map contains `require` but does **not** contain `number_of_replicas` or `total_shards_per_node`.
  - This confirms that removing `skipEmptyCheck` is NOT correct for the routing filter fields, and that the expand path correctly skips null/absent int64 values when `skipEmptyCheck` is false (the default).

## 5. Delta spec

- [ ] 5.1 Keep `openspec/changes/ilm-allocate-omit-defaults/specs/elasticsearch-index-lifecycle/spec.md` aligned with the implementation changes (REQ-023 or next available number — verify against the current main spec).
