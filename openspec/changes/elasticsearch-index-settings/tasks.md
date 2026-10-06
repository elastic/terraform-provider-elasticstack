## 1. Shared settings metadata

- [ ] 1.1 Add `getDynamicSettingAttributes() map[string]schema.Attribute` to `internal/elasticsearch/index/settings_keys.go`, one entry per `DynamicSettingsKeys` key, carrying the same type/description/validators currently hand-declared in `internal/elasticsearch/index/index/schema.go` (e.g. `mapping_total_fields_limit`, `number_of_replicas`, `blocks_read_only`, etc.)
- [ ] 1.2 Create `internal/elasticsearch/index/dynamicsettings` package with a `Model` struct carrying one `tfsdk`-tagged field per `DynamicSettingsKeys` entry, matching the types used by `getDynamicSettingAttributes()`
- [ ] 1.3 Extend `internal/elasticsearch/index/reflectutil.go`'s `GetFieldValueByTagValue`/`SetFieldValueByTagValue` to recurse into anonymous (`field.Anonymous`) struct fields
- [ ] 1.4 Unit tests for the `reflectutil.go` change: flat field lookup (unchanged), embedded field lookup (new), set on both shapes

## 2. Migrate `elasticstack_elasticsearch_index` to the shared definition

- [ ] 2.1 Replace the hand-declared dynamic-setting attributes in `internal/elasticsearch/index/index/schema.go` with a merge of `indexparent.getDynamicSettingAttributes()` into the existing attribute map (static settings stay hand-declared)
- [ ] 2.2 Replace the individual dynamic-setting fields in `index/index/models.go`'s `tfModel` with an anonymously-embedded `dynamicsettings.Model`
- [ ] 2.3 Verify `toIndexSettings`, `hydrateAllSettingsFromRaw`, and any other reflection-based helper in `index/index` still resolve every `DynamicSettingsKeys` entry correctly through the embedded struct
- [ ] 2.4 Run existing `index/index` unit tests (`models_test.go`, `settings_read_test.go`, etc.) and confirm no attribute behavior changed; add a regression test if schema attribute equality (type/description/plan modifiers) isn't already covered
- [ ] 2.5 Run existing `elasticstack_elasticsearch_index` acceptance tests covering dynamic settings and confirm no plan diff is introduced solely by this refactor

## 3. New `elasticstack_elasticsearch_index_settings` resource

- [ ] 3.1 Create `internal/elasticsearch/index/indexsettings` package following the `indexmappings` package's structure (`resource.go`, `schema.go`, `models.go`, `create.go`, `read.go`, `update.go`, `delete.go`)
- [ ] 3.2 `models.go`: `tfModel` embeds `entitycore.ResourceTimeoutsField`, `dynamicsettings.Model` (anonymous), plus `ID`, `Index`, `SettingsJSON jsontypes.Normalized`, `ElasticsearchConnection`; implement `GetID`/`GetResourceID`/`GetElasticsearchConnection`
- [ ] 3.3 `schema.go`: merge `indexparent.getDynamicSettingAttributes()` into the schema map alongside `id`, `index` (`RequiresReplace`), `settings_json` (`jsontypes.Normalized`, JSON-object validator), with a validator ensuring at least one dynamic attribute or `settings_json` is declared
- [ ] 3.4 Add a `settings_json` plan-time validator (modeled on `internal/elasticsearch/queryrulesets/validators.go`'s pattern) that parses the JSON object's top-level keys and rejects any literal `StaticSettingsKeys` match; keys outside `AllSettingsKeys` are permitted
- [ ] 3.5 `create.go`: call `elasticsearch.GetIndex`; error if not found; compute `id` via `client.ID(ctx, indexName)`; build the settings payload from declared typed attributes + parsed `settings_json`; call `elasticsearch.UpdateIndexSettings`
- [ ] 3.6 `update.go`: compute the declared-subset union of plan vs. state (typed attributes + `settings_json` keys); diff; for keys present in state but absent from plan, include them as `null` in the update payload (reset semantics, mirroring `index/index/update.go`'s `updateSettings`); skip the API call when the union is unchanged
- [ ] 3.7 `read.go`: call `elasticsearch.GetIndex`; if not found, report not-found; otherwise populate only the typed attributes and `settings_json` keys present in the previously stored state (full API response as the initial subset when state is empty, e.g. post-import)
- [ ] 3.8 `delete.go`: no-op (mirrors `indexmappings/delete.go`)
- [ ] 3.9 `resource.go`: wire `entitycore.NewElasticsearchResource[tfModel]("index_settings", ...)` with `Schema`, `Create`, `Read`, `Update`, `Delete`; implement `ImportState` via `resource.ImportStatePassthroughID`
- [ ] 3.10 Register the new resource (`NewIndexSettingsResource`) with the provider's resource list alongside `elasticstack_elasticsearch_index_mappings`

## 4. Tests

- [ ] 4.1 Unit tests for schema validation: required `index`, `settings_json` static-key rejection, `settings_json` permissive-on-unknown-key, at-least-one-setting-declared
- [ ] 4.2 Unit tests for create: index-not-found error path, success path with correct API payload and computed `id`
- [ ] 4.3 Unit tests for update: changed setting triggers call with correct payload, removed setting sends `null`, no-op when unchanged
- [ ] 4.4 Unit tests for read: declared-subset filtering (unrelated settings ignored), out-of-band drift surfaced for declared settings, not-found removes from state
- [ ] 4.5 Acceptance test: create against a real index, update a declared setting, remove a declared setting and confirm reset, destroy and confirm settings remain on the index, import
- [ ] 4.6 Acceptance test: two resource instances via `for_each` over two concrete index names, confirming independent management
- [ ] 4.7 Run `make build`, `go vet ./...`, `go test ./internal/elasticsearch/index/...` (`TF_ACC=1` for acceptance, against a running Elastic Stack)

## 5. Documentation

- [ ] 5.1 Add an example under `examples/resources/elasticstack_elasticsearch_index_settings/` (`resource.tf`, `import.sh`) mirroring the `index_mappings` example layout
- [ ] 5.2 Regenerate provider docs (`make docs` or repo's documented doc-gen command) so `docs/resources/elasticsearch_index_settings.md` is produced from the schema descriptions
- [ ] 5.3 Note in the resource description and generated docs that `destroy` is a no-op and does not revert or clear settings

## 6. Spec sync

- [ ] 6.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-index-settings --type change` and resolve any reported issues
- [ ] 6.2 After implementation lands, sync delta specs into `openspec/specs/elasticstack-elasticsearch-index-settings/` and `openspec/specs/elasticsearch-index/` (or archive the change) per `openspec-sync-specs` / `openspec-archive-change`
