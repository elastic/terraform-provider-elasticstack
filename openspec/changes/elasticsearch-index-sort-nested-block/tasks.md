## 1. Schema

- [ ] 1.1 Add `sort` as a `ListNestedAttribute` in `internal/elasticsearch/index/index/schema.go` with nested `field` (required), `order` (optional, `asc`/`desc`), `missing` (optional, `_last`/`_first`), and `mode` (optional, `min`/`max`) attributes; attach `sortMigrationPlanModifier{}` and `ConflictsWith(path.MatchRoot("sort_field"), path.MatchRoot("sort_order"))` validators.
- [ ] 1.2 Add `DeprecationMessage` and `ConflictsWith(path.MatchRoot("sort"))` validator to `sort_field` in `schema.go`; replace the existing `setplanmodifier.RequiresReplace()` with `legacySortFieldPlanModifier{}`.
- [ ] 1.3 Add `DeprecationMessage` and `ConflictsWith(path.MatchRoot("sort"))` validator to `sort_order` in `schema.go`; replace the existing `listplanmodifier.RequiresReplace()` with `legacySortOrderPlanModifier{}`.

## 2. Plan modifiers

- [ ] 2.1 Create `internal/elasticsearch/index/index/sort_plan_modifiers.go` (or similar) implementing `sortMigrationPlanModifier` (planmodifier.List): suppresses replace when sort is null in state, private state has ordered sort config, plan fields+orders match private state (with `asc` default), and plan `missing`/`mode` are semantically equivalent to existing settings (treat explicit defaults as equivalent to absent settings).
- [ ] 2.2 Implement `legacySortFieldPlanModifier` (planmodifier.Set): reads `sort` from plan via `req.Plan.GetAttribute`; suppresses replace when `sort` is non-null in plan (new attribute handles the replace decision).
- [ ] 2.3 Implement `legacySortOrderPlanModifier` (planmodifier.List): same logic as `legacySortFieldPlanModifier`.

## 3. Model

- [ ] 3.1 Add `sortEntryModel` struct to `models.go` with `Field`, `Order`, `Missing`, `Mode` as `types.String`.
- [ ] 3.2 Add `Sort types.List` field to `tfModel` in `models.go`.
- [ ] 3.3 Add `"sort.missing"` and `"sort.mode"` to `staticSettingsKeys` in `models.go`.
- [ ] 3.4 Update `toIndexSettings()` in `models.go`: add a pre-processing step — when `Sort` is non-null, expand to `sort.field`, `sort.order`, `sort.missing`, `sort.mode` parallel arrays in the settings map (omitting null-valued per-entry fields); skip those ES keys in the reflection loop to avoid double-processing.

## 4. Read / flatten

- [ ] 4.1 Update `read.go` (or `resource.go`) to store ordered sort config in private state under key `"sort_config"` (JSON-marshaled `sortPrivateState{Fields, Orders}`) on every successful read.
- [ ] 4.2 When state has non-null `sort`: populate `Sort` from ES response `sort.field`/`sort.order`/`sort.missing`/`sort.mode` arrays (by index).
- [ ] 4.3 When state has null `sort`: populate `SortField` and `SortOrder` from ES response arrays as before (legacy path).

## 5. `use_existing.go`

- [ ] 5.1 Add `case "sort.missing"` in `compareStaticPlanAndES()` using `stringSliceOrderedFromAny` (mirroring the `"sort.order"` case).
- [ ] 5.2 Add `case "sort.mode"` in `compareStaticPlanAndES()` using `stringSliceOrderedFromAny`.
- [ ] 5.3 Update adoption comparison logic so `sort.field` is validated as order-sensitive when the plan is produced from the new `sort` `ListNestedAttribute`, while preserving the existing unordered comparison behavior for the legacy `sort_field` `SetAttribute`.
- [ ] 5.4 Add `case "sort.missing"` and `case "sort.mode"` in `configuredDisplayFromPlanValue()`.

## 6. Tests

- [ ] 6.1 Add unit tests for `sortMigrationPlanModifier`: verify replace is suppressed when migrating with equivalent field+order, verify explicit default `missing`/`mode` values are treated as equivalent to absent settings, verify replace is required for non-equivalent `missing`/`mode`, verify replace defaults to required when private state is absent.
- [ ] 6.2 Add unit tests for `legacySortFieldPlanModifier` and `legacySortOrderPlanModifier`: verify replace is suppressed when `sort` is present in plan.
- [ ] 6.3 Update `acc_test.go` to add acceptance test coverage for:
  - Creating an index with the new `sort` attribute (field, order, missing, mode).
  - Creating an index with the deprecated `sort_field`/`sort_order` attributes (regression).
  - Migrating from legacy `sort_field`/`sort_order` to `sort` without replace when settings are equivalent.
  - Confirming replace is required when `missing` or `mode` are added to the new `sort` block after initial creation.

## 7. Documentation and OpenSpec

- [ ] 7.1 Regenerate `docs/resources/elasticsearch_index.md` after schema changes (run `make generate` or equivalent doc target).
- [ ] 7.2 Add migration guide note to release notes / CHANGELOG: first apply after upgrade may force one replace if private state is absent; `terraform refresh` before apply avoids this; explicit default `missing`/`mode` values are equivalent to absent settings; non-equivalent `missing`/`mode` values require replace.
- [ ] 7.3 Sync delta spec from `openspec/changes/elasticsearch-index-sort-nested-block/specs/elasticsearch-index/spec.md` into `openspec/specs/elasticsearch-index/spec.md` and archive this change.
