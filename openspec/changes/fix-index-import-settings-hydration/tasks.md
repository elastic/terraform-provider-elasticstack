## 1. Constant and private state key

- [ ] 1.1 Add `importHydrationPrivateStateKey = "import_hydration"` constant to `constants.go`

## 2. ImportState override

- [ ] 2.1 Override `ImportState` in `resource.go` to call `resource.ImportStatePassthroughID` and then write `[]byte("true")` to `resp.Private` under `importHydrationPrivateStateKey`

## 3. Propagate hydrateAll parameter

- [ ] 3.1 Add `hydrateAll bool` parameter to `readIndex` in `read.go`; pass it through to `populateFromAPI`
- [ ] 3.2 Add `hydrateAll bool` parameter to `populateFromAPI` in `models.go`
- [ ] 3.3 Update the two `readIndex` call sites in `create.go` to pass `false`
- [ ] 3.4 Update the direct `populateFromAPI` call in `create.go` (`adoptExistingIndexOnCreate`) to pass `false`

## 4. Read: flag check and clear

- [ ] 4.1 In `resource.go Read`, read `req.Private.GetKey(ctx, importHydrationPrivateStateKey)` to determine `hydrateAll` and pass it to `readIndex`
- [ ] 4.2 In `resource.go Read`, after `resp.State.Set`, if `hydrateAll` clear the flag: `resp.Private.SetKey(ctx, importHydrationPrivateStateKey, nil)`

## 5. Settings hydration functions

- [ ] 5.1 Create `settings_read.go` with `hydrateAllSettingsFromRaw(ctx context.Context, model *tfModel) diag.Diagnostics`: unmarshal `model.SettingsRaw` into a flat `map[string]json.RawMessage`; for each key in `allSettingsKeys`, look up `"index."+key` in the map; use `convertSettingsKeyToTFFieldKey` to find the struct field; type-convert the raw JSON string value (`strconv.ParseInt` for `types.Int64` fields, `strconv.ParseBool` for `types.Bool` fields, unquote for `types.String` fields); skip keys not present in the map; skip fields that fail conversion
- [ ] 5.2 In `hydrateAllSettingsFromRaw`, handle `"index.analysis"` separately: unmarshal its value as `map[string]json.RawMessage` and populate `model.AnalysisAnalyzer`, `model.AnalysisTokenizer`, `model.AnalysisCharFilter`, `model.AnalysisFilter`, `model.AnalysisNormalizer` from sub-keys `analyzer`, `tokenizer`, `char_filter`, `filter`, `normalizer` respectively (marshal each to `jsontypes.NewNormalizedValue`)
- [ ] 5.3 In `hydrateAllSettingsFromRaw`, handle `"index.query.default_field"` (maps to `QueryDefaultField types.Set`) using the same array-extraction pattern as `extractSortSetting` in `sort_read.go`
- [ ] 5.4 Create `populateOperationalDefaults(model *tfModel)` in `settings_read.go`: set `deletion_protection = true`, `wait_for_active_shards = "1"`, `master_timeout = "30s"`, `timeout = "30s"` when each field is null

## 6. Wire hydration into populateFromAPI

- [ ] 6.1 At the end of `populateFromAPI` in `models.go`, when `hydrateAll` is true, call `hydrateAllSettingsFromRaw(ctx, model)` then `populateOperationalDefaults(model)`

## 7. Build verification

- [ ] 7.1 Run `make build` and confirm no compilation errors

## 8. Acceptance test

- [ ] 8.1 Add a new test config directory `testdata/TestAccResourceIndexImport/` with an index config that sets `number_of_replicas`, `refresh_interval`, and `analysis_analyzer`
- [ ] 8.2 Add `TestAccResourceIndexImport` acceptance test in `acc_test.go` with two steps: (1) create the index, (2) import it with `ImportState: true` and verify no plan drift for `number_of_replicas`, `refresh_interval`, `analysis_analyzer`, `deletion_protection`, `wait_for_active_shards`, `master_timeout`, and `timeout`
- [ ] 8.3 Run the acceptance test against a live Elasticsearch cluster to confirm it passes
