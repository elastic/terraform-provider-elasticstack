## 1. Plugin Framework model and schema

- [ ] 1.1 Define `tfModel` struct in `internal/elasticsearch/cluster/settings.go` with `ID types.String`, `ElasticsearchConnection types.List`, `Persistent types.List` (nested setting objects), and `Transient types.List`.
- [ ] 1.2 Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` value-receiver methods to the model.
- [ ] 1.3 Define nested model types: `settingsBlockModel` (containing `Setting types.Set`) and `settingModel` (containing `Name types.String`, `Value types.String`, `ValueList types.List`).
- [ ] 1.4 Write `getSchema() schema.Schema` factory returning `ListNestedBlock` for `persistent`/`transient` (max 1), each with a `SetNestedAttribute` for `setting` containing `name` (required), `value` (optional), `value_list` (optional list string). No `elasticsearch_connection` block.
- [ ] 1.5 Add schema validators for duplicate names and mutually-exclusive `value`/`value_list`.

## 2. Conversion helpers

- [ ] 2.1 Write `expandSettings(ctx, settingsList) (map[string]any, diag.Diagnostics)` that converts PF nested models into the flat settings map used by the API helpers.
- [ ] 2.2 Write `flattenSettings(category, configuredSettings, apiResponse) []any` that builds the nested block list from the flat API response, deciding `value` vs `value_list` by type assertion.
- [ ] 2.3 Write `getConfiguredSettings(state tfModel) (map[string]any, diag.Diagnostics)` equivalent for the PF model.
- [ ] 2.4 Port `updateRemovedSettings(name, oldSettings, newSettings, targetMap)` to work with expanded maps instead of SDK `GetChange`.

## 3. Read and delete callbacks

- [ ] 3.1 Implement `readClusterSettings(ctx, client, resourceID, state) (tfModel, bool, diag.Diagnostics)` that calls `elasticsearch.GetSettings`, then uses `flattenSettings` for both `persistent` and `transient`.
- [ ] 3.2 Implement `deleteClusterSettings(ctx, client, resourceID, state) diag.Diagnostics` that reads the current state model, builds null maps for all tracked settings, and calls `elasticsearch.PutSettings`.

## 4. Resource struct and overrides

- [ ] 4.1 Define `type clusterSettingsResource struct { *entitycore.ElasticsearchResource[tfModel] }`.
- [ ] 4.2 Implement `newClusterSettingsResource()` that constructs the envelope with the schema factory, read callback, delete callback, and placeholder write callbacks.
- [ ] 4.3 Implement `Create` override: decode plan, validate, expand settings, PUT, read back, set state.
- [ ] 4.4 Implement `Update` override: decode plan and state, expand both, call `updateRemovedSettings` for each category, PUT, read back, set state.
- [ ] 4.5 Implement `Delete` override: decode state, call `deleteClusterSettings`, and return its diagnostics.
- [ ] 4.6 Implement `ImportState` as passthrough on `id`.

## 5. Provider wiring and cleanup

- [ ] 5.1 Replace the SDK `ResourceSettings()` registration in the provider with the new PF `NewClusterSettingsResource()` factory.
- [ ] 5.2 Remove the old SDK resource code once the PF version compiles. Keep any shared helper code (e.g., `elasticsearch.PutSettings`, `elasticsearch.GetSettings`) since the typed-client helpers are already PF-compatible.
- [ ] 5.3 Update any provider-level type assertions or resource lists.

## 6. Verification

- [ ] 6.1 Run `make build`.
- [ ] 6.2 Run `make check-lint`.
- [ ] 6.3 Run `make check-openspec`.
- [ ] 6.4 Run focused unit tests for the conversion helpers.
- [ ] 6.5 Run acceptance tests for `cluster_settings` if infrastructure is available.
