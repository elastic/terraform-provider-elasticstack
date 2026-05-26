## 1. Client layer

- [x] 1.1 Create `internal/clients/elasticsearch/synonyms.go` with the following exported functions:
  - `GetSynonymSet(ctx, apiClient, synonymSetID)` — paginated GET loop (500 per page) returning all rules
  - `PutSynonymSet(ctx, apiClient, synonymSetID, rules)` — PUT request replacing the entire set
  - `DeleteSynonymSet(ctx, apiClient, synonymSetID)` — DELETE with custom HTTP 400 error message

## 2. Models

- [x] 2.1 Create `internal/elasticsearch/synonyms/models.go` with:
  - `SynonymSetData` struct (implements `entitycore.ElasticsearchResourceModel`): `ID`, `SynonymSetID`, `SynonymsSet` fields
  - `SynonymRuleModel` struct: `ID`, `Synonyms` fields
  - `populateFromAPI` method to map `[]types.SynonymRuleRead` into `SynonymSetData`
  - `toAPIRules` method to map `SynonymSetData` into `[]types.SynonymRule` (generating UUIDs for rules with no stored ID)

## 3. Resource

- [x] 3.1 Create `internal/elasticsearch/synonyms/resource.go`:
  - `NewSynonymSetResource() resource.Resource` — wire `entitycore.NewElasticsearchResource[SynonymSetData]`
  - Schema factory: `synonym_set_id` (required, RequiresReplace), `synonyms_set` (required list-nested block with `id` optional+computed and `synonyms` required)
  - Implement `ImportState` via `resource.ImportStatePassthroughID` on `path.Root("id")`
- [x] 3.2 Create `internal/elasticsearch/synonyms/create.go` (upsert handler calling `PutSynonymSet` then setting ID to `<cluster_uuid>/<synonym_set_id>`)
- [x] 3.3 Create `internal/elasticsearch/synonyms/read.go` (calling `GetSynonymSet`, removing from state when 404)
- [x] 3.4 Create `internal/elasticsearch/synonyms/delete.go` (calling `DeleteSynonymSet` with clear HTTP 400 diagnostic)

## 4. Data source

- [ ] 4.1 Create `internal/elasticsearch/synonyms/data_source.go`:
  - `NewSynonymSetDataSource() datasource.DataSource` — wire `entitycore.NewElasticsearchDataSource[SynonymSetData]`
  - Schema: `synonym_set_id` required; all other fields computed
  - Read handler calling `GetSynonymSet`, erroring when not found

## 5. Descriptions

- [ ] 5.1 Create `internal/elasticsearch/synonyms/descriptions/` directory with markdown description files for the resource and data source

## 6. Provider registration

- [ ] 6.1 Register `synonyms.NewSynonymSetResource` in `provider/plugin_framework.go` resources list
- [ ] 6.2 Register `synonyms.NewSynonymSetDataSource` in `provider/plugin_framework.go` data sources list

## 7. Tests

- [ ] 7.1 Create `internal/elasticsearch/synonyms/acc_test.go` with acceptance tests:
  - Basic CRUD: create a synonym set, verify state, update rules, verify state, destroy
  - Rule ordering: verify round-trip preserves rule order
  - Optional rule ID: create with no `id` on a rule, verify provider generates one, re-apply, verify no diff
  - Delete with in-use set: verify clear error diagnostic when set is referenced by an analyzer
  - Import: create resource, import by ID, verify state matches, verify subsequent plan shows no diff
  - Data source: create resource, read via data source, verify all attributes match
- [ ] 7.2 Create `internal/elasticsearch/synonyms/acc_test.go` testdata directories under `testdata/` per test (following enrich pattern)

## 8. Documentation and validation

- [ ] 8.1 Run `make build` and fix any compilation errors
- [ ] 8.2 Run `go vet ./internal/elasticsearch/synonyms/...`
- [ ] 8.3 Run `go test ./internal/elasticsearch/synonyms/...` (unit tests only; TF_ACC=1 for acceptance)
- [ ] 8.4 Run `make generate-docs` and verify docs render correctly
- [ ] 8.5 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-synonym-sets --type change`
