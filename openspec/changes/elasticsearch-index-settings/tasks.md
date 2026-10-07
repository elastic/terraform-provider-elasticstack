## 1. Shared settings metadata

- [x] 1.1 Add `GetDynamicSettingAttributes() map[string]schema.Attribute` to `internal/elasticsearch/index/settings_keys.go`, one entry per `DynamicSettingsKeys` key, carrying the same type/description/validators currently hand-declared in `internal/elasticsearch/index/index/schema.go` (e.g. `mapping_total_fields_limit`, `number_of_replicas`, `blocks_read_only`, etc.)
- [x] 1.2 Create `internal/elasticsearch/index/dynamicsettings` package with a `Model` struct carrying one `tfsdk`-tagged field per `DynamicSettingsKeys` entry, matching the types used by `GetDynamicSettingAttributes()`
- [x] 1.3 Extend `internal/elasticsearch/index/reflectutil.go`'s `GetFieldValueByTagValue`/`SetFieldValueByTagValue` to recurse into anonymous (`field.Anonymous`) struct fields
- [x] 1.4 Unit tests for the `reflectutil.go` change: flat field lookup (unchanged), embedded field lookup (new), set on both shapes

## 2. Migrate `elasticstack_elasticsearch_index` to the shared definition

- [x] 2.1 Replace the hand-declared dynamic-setting attributes in `internal/elasticsearch/index/index/schema.go` with a merge of `indexparent.GetDynamicSettingAttributes()` into the existing attribute map (static settings stay hand-declared)
- [x] 2.2 Replace the individual dynamic-setting fields in `index/index/models.go`'s `tfModel` with an anonymously-embedded `dynamicsettings.Model`
- [x] 2.3 Verify `toIndexSettings`, `hydrateAllSettingsFromRaw`, and any other reflection-based helper in `index/index` still resolve every `DynamicSettingsKeys` entry correctly through the embedded struct; update `setTFModelField` and `pruneImportHydratedPlanFields` in `index/index/settings_read.go` (both scan direct struct fields) to handle anonymous embedded structs, and keep their regression tests passing
- [x] 2.4 Run existing `index/index` unit tests (`models_test.go`, `settings_read_test.go`, etc.) and confirm no attribute behavior changed; add a regression test if schema attribute equality (type/description/plan modifiers) isn't already covered
- [x] 2.5 Run existing `elasticstack_elasticsearch_index` acceptance tests covering dynamic settings and confirm no plan diff is introduced solely by this refactor

## 3. New `elasticstack_elasticsearch_index_settings` resource

- [x] 3.1 Create `internal/elasticsearch/index/indexsettings` package following the `indexmappings` package's structure (`resource.go`, `schema.go`, `models.go`, `create.go`, `read.go`, `update.go`, `delete.go`)
- [x] 3.2 `models.go`: `tfModel` embeds `entitycore.ResourceTimeoutsField`, `dynamicsettings.Model` (anonymous), plus `ID`, `Index`, `SettingsJSON jsontypes.Normalized`, `ElasticsearchConnection`; implement `GetID`/`GetResourceID`/`GetElasticsearchConnection`
- [x] 3.3 `schema.go`: merge `indexparent.GetDynamicSettingAttributes()` into the schema map alongside `id`, `index` (`RequiresReplace`), `settings_json` (`jsontypes.Normalized`, JSON-object validator); an index-only configuration (no typed attribute, no `settings_json`) is valid — no at-least-one-setting validator (only the `settings_json = {}` empty-object literal is rejected)
- [x] 3.4 Add a `settings_json` plan-time validator (modeled on `internal/elasticsearch/queryrulesets/validators.go`'s pattern) that parses the JSON object's top-level keys and rejects `{}`, nested object values, explicit `null` values and any literal `StaticSettingsKeys` match; keys outside `AllSettingsKeys` are permitted. Also reject any `settings_json` key that overlaps a configured typed attribute, after canonicalizing key spellings (with/without `index.` prefix)
- [x] 3.5 `create.go`: call `elasticsearch.GetIndex`; error if not found; compute `id` via `client.ID(ctx, indexName)`; build the settings payload from declared typed attributes + parsed `settings_json`; call `elasticsearch.UpdateIndexSettings`
- [x] 3.6 `update.go`: compute the declared-subset union of plan vs. state (typed attributes + `settings_json` keys); diff; for keys present in state but absent from plan, include them as `null` in the update payload (reset semantics, mirroring `index/index/update.go`'s `updateSettings`); skip the API call when the union is unchanged
- [x] 3.6a Share one flat-key normalizer (strip optional `index.` prefix) between the overlap validator, update diffing and read
- [x] 3.7 `read.go`: call `elasticsearch.GetIndex`; if not found, report not-found; otherwise populate only the typed attributes and `settings_json` keys present in the previously stored state, reconciling API string values to the declared JSON scalar type (reuse the `settings_read.go` conversion logic) ; hydrate the known dynamic typed attributes (leaving `settings_json` unset) only when the import private-state flag is set, then clear it — never infer import from empty tracked state; when the API no longer reports a tracked setting, set that attribute to null / drop that `settings_json` key so drift shows
- [x] 3.8 `delete.go`: no-op (mirrors `indexmappings/delete.go`)
- [x] 3.9 `resource.go`: wire `entitycore.NewElasticsearchResource[tfModel]("index_settings", ...)` with `Schema`, `Create`, `Read`, `Update`, `Delete`; implement a custom `ImportState` that parses the composite ID, sets `id` and `index`, and sets the import private-state key (instead of `resource.ImportStatePassthroughID`); verify the envelope exposes private state to `Read` and allows clearing it, extending `entitycore` minimally if not
- [x] 3.10 Register the new resource (`NewIndexSettingsResource`) with the provider's resource list alongside `elasticstack_elasticsearch_index_mappings`

## 4. Tests

- [x] 4.1 Unit tests for schema validation: required `index`, `settings_json` static-key rejection, `settings_json` permissive-on-unknown-key, nested-object and explicit-null rejection, typed/`settings_json` overlap rejection (including a nested-overlap case), (including `index.`-prefixed spelling); index-only configuration must be valid (no empty-declaration error)
- [x] 4.2 Unit tests for create: index-not-found error path, success path with correct API payload and computed `id`
- [x] 4.3 Unit tests for update: changed setting triggers call with correct payload, removed setting sends `null`, no-op when unchanged; non-success API responses surface error diagnostics and leave state unchanged (REQ-007)
- [x] 4.4 Unit tests for read: declared-subset filtering (unrelated settings ignored), out-of-band drift surfaced for declared settings, not-found removes from state
- [x] 4.4a Unit tests for the import flag: set by `ImportState`, hydration only when set, cleared after the first read, not set by create/update; and read with no remaining tracked settings does not hydrate
- [x] 4.4b Unit tests for read: numeric/boolean `settings_json` values round-trip from string API values with no drift
- [x] 4.5 Acceptance test: create against a real index, update a declared setting, remove a declared setting and confirm reset, destroy and confirm settings remain on the index, import
- [x] 4.5a Acceptance test: import, then narrow the config and apply successfully (only dynamic keys sent as `null`); also repeated refresh after an external reset of the last tracked setting adopts nothing
- [x] 4.6 Acceptance test: two resource instances via `for_each` over two concrete index names, confirming independent management
- [x] 4.7 Run `make build`, `go vet ./...`, `go test ./internal/elasticsearch/index/...` (`TF_ACC=1` for acceptance, against a running Elastic Stack)

## 5. Documentation

- [x] 5.1 Add an example under `examples/resources/elasticstack_elasticsearch_index_settings/` (`resource.tf`, `import.sh`) mirroring the `index_mappings` example layout
- [x] 5.2 Regenerate provider docs (`make docs` or repo's documented doc-gen command) so `docs/resources/elasticsearch_index_settings.md` is produced from the schema descriptions
- [x] 5.3 Note in the resource description and generated docs that `destroy` is a no-op and does not revert or clear settings

## 6. Spec sync

- [x] 6.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-index-settings --type change` and resolve any reported issues
- [x] 6.2 After implementation lands, sync delta specs into `openspec/specs/elasticstack-elasticsearch-index-settings/` and `openspec/specs/elasticsearch-index/` (or archive the change) per `openspec-sync-specs` / `openspec-archive-change`

## 7. Approved refinement — arrays in settings_json, empty/reset behavior, unknown-value validation

- [x] 7.1 Support arrays in the `settings_json` escape hatch: values may be arrays of string/number/boolean elements, including empty arrays; reject at plan time array elements that are objects or `null` (consistent with the flat non-null contract); overlap and shape checks defer safely when a `settings_json` value is unknown at plan time (no spurious empty-declaration errors)
- [x] 7.2 Index-only configuration is valid and empty/reset behavior: remove the at-least-one-setting validator (keep the `settings_json = {}` empty-object rejection); index-only create verifies the target index exists and records the resource in state with the computed `id` without issuing a settings PUT; removing the last/all declared settings sends `null` for the previously tracked keys only; repeated refresh of an empty declared subset stays empty and never adopts unrelated keys
- [x] 7.3 Read array reconciliation: reconcile API values arriving as real JSON arrays or as JSON-encoded array strings against the element types declared in state, preserving element order and drift, never silently dropping a valid tracked array, and never changing string elements to numbers (or vice versa) accidentally; typed `query_default_field` decodes JSON-encoded string arrays following the existing `stringSliceFromAny` precedent in the `index` resource
- [x] 7.4 Restore long attribute descriptions as `//go:embed` Markdown files in the shared package (`GetDynamicSettingAttributes()`), preserving exact content and trailing newline so descriptions remain byte-for-byte identical to the pre-refactor hand-declared form
- [x] 7.5 Update the stale not-yet-implemented Create callback comment in `internal/elasticsearch/index/indexsettings/acc_validation_test.go` now that Create is implemented
- [x] 7.6 Unit and acceptance tests for the delta-spec scenarios this refinement adds: array validation (scalar arrays permitted incl. empty, null/object elements rejected), array read reconciliation, typed `query_default_field` encoded-array read, index-only create (no settings PUT), removing the last/all declared settings (null for previously tracked keys only), empty repeated refresh adopting nothing, and unknown-value overlap deferral; then `make build`, `go vet ./...`, and `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-index-settings --type change`
