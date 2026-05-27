## 1. Constant and private state key

- [x] 1.1 Add `importHydrationPrivateStateKey = "import_hydration"` constant to `constants.go`

## 2. ImportState override

- [x] 2.1 Override `ImportState` in `resource.go` to call `resource.ImportStatePassthroughID` and then write `[]byte("true")` to `resp.Private` under `importHydrationPrivateStateKey`

## 3. Propagate hydrateAll parameter

- [x] 3.1 Add `hydrateAll bool` parameter to `readIndex` in `read.go`; pass it through to `populateFromAPI`
- [x] 3.2 Add `hydrateAll bool` parameter to `populateFromAPI` in `models.go`
- [x] 3.3 Update the two `readIndex` call sites in `create.go` to pass `false`
- [x] 3.4 Update the direct `populateFromAPI` call in `create.go` (`adoptExistingIndexOnCreate`) to pass `false`

## 4. Read: flag check (do not clear here)

- [x] 4.1 In `resource.go Read`, read `req.Private.GetKey(ctx, importHydrationPrivateStateKey)` to determine `hydrateAll` and pass it to `readIndex`; do NOT clear the flag in `Read` — it is cleared in `ModifyPlan`

## 4b. ModifyPlan: selective clearing of unconfigured Optional settings

- [x] 4b.1 Implement `ModifyPlan` on the index resource in `resource.go` (`resource.ResourceWithModifyPlan` interface): check `req.Private.GetKey(ctx, importHydrationPrivateStateKey)`; if nil, return immediately; otherwise unmarshal both the config and plan models via `req.Config.Get` and `req.Plan.Get`; for each `Optional` (non-`Computed`) settings field populated by `hydrateAllSettingsFromRaw` — all fields from `allSettingsKeys` (e.g. `number_of_replicas`, `refresh_interval`, slowlog thresholds, block settings) plus the analysis fields (`analysis_analyzer`, `analysis_tokenizer`, `analysis_char_filter`, `analysis_filter`, `analysis_normalizer`) — if the corresponding config field is null, set the plan model field to null; write back via `resp.Plan.Set` and clear the flag via `resp.Private.SetKey(ctx, importHydrationPrivateStateKey, nil)`; leave operational defaults (`deletion_protection`, `wait_for_active_shards`, `master_timeout`, `timeout`) untouched (they are `Optional+Computed` and keep their hydrated values)

## 5. Settings hydration functions

- [x] 5.1 Create `settings_read.go` with `hydrateAllSettingsFromRaw(ctx context.Context, model *tfModel) diag.Diagnostics`: unmarshal `model.SettingsRaw` into a flat `map[string]json.RawMessage`; for each key in `allSettingsKeys`, look up `"index."+key` in the map; use `convertSettingsKeyToTFFieldKey` to find the struct field; all flat-settings values are JSON strings (the raw bytes for `number_of_replicas = 1` are `"1"` including the quotes), so type-convert by first calling `json.Unmarshal(rawValue, &s)` to get a Go `string`, then `strconv.ParseInt(s, 10, 64)` for `types.Int64` fields, `strconv.ParseBool(s)` for `types.Bool` fields, or use `s` directly for `types.String` fields; skip keys not present in the map; skip fields that fail conversion
- [x] 5.2 In `hydrateAllSettingsFromRaw`, handle `"index.analysis"` separately: unmarshal its value as `map[string]json.RawMessage` and populate `model.AnalysisAnalyzer`, `model.AnalysisTokenizer`, `model.AnalysisCharFilter`, `model.AnalysisFilter`, `model.AnalysisNormalizer` from sub-keys `analyzer`, `tokenizer`, `char_filter`, `filter`, `normalizer` respectively (marshal each to `jsontypes.NewNormalizedValue`)
- [x] 5.3 In `hydrateAllSettingsFromRaw`, handle `"index.query.default_field"` (maps to `QueryDefaultField types.Set`) using the same array-extraction pattern as `extractSortSetting` in `sort_read.go`
- [x] 5.4 Create `populateOperationalDefaults(model *tfModel)` in `settings_read.go`: set `deletion_protection = true`, `wait_for_active_shards = "1"`, `master_timeout = "30s"`, `timeout = "30s"` when each field is null

## 6. Wire hydration into populateFromAPI

- [x] 6.1 At the end of `populateFromAPI` in `models.go`, when `hydrateAll` is true, call `hydrateAllSettingsFromRaw(ctx, model)` then `populateOperationalDefaults(model)`

## 7. Build verification

- [x] 7.1 Run `make build` and confirm no compilation errors

## 8. Acceptance test

- [x] 8.1 Add a new test config directory `testdata/TestAccResourceIndexImport/` with an index config that sets `number_of_replicas`, `refresh_interval`, and `analysis_analyzer`
- [x] 8.2 Add `TestAccResourceIndexImport` acceptance test in `acc_test.go` with two steps: (1) create the index, (2) import it with `ImportState: true` and verify no plan drift for `number_of_replicas`, `refresh_interval`, `analysis_analyzer`, `deletion_protection`, `wait_for_active_shards`, `master_timeout`, and `timeout`
- [ ] 8.3 Run the acceptance test against a live Elasticsearch cluster to confirm it passes
