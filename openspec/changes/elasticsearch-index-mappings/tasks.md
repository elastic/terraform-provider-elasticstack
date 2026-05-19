## 1. New resource package scaffold

- [ ] 1.1 Create package `internal/elasticsearch/index/indexmappings/` with files:
  - `resource.go` — wire `entitycore.NewElasticsearchResource[tfModel]` with name `"index_mappings"`; expose `NewIndexMappingsResource() resource.Resource`; implement `ImportState` via `resource.ImportStatePassthroughID`
  - `schema.go` — define `getSchemaFactory`; see task 2 for required attributes
  - `models.go` — define `tfModel` struct with `tfsdk` tags; implement `GetID`, `GetResourceID`, `GetElasticsearchConnection`
  - `create.go` — implement `createIndexMappings` (see task 3)
  - `read.go` — implement `readIndexMappings` (see task 4)
  - `update.go` — implement `updateIndexMappings` (see task 5)
  - `delete.go` — implement `deleteIndexMappings` (see task 6)

## 2. Schema definition (`schema.go`)

- [ ] 2.1 Define the following attributes in `getSchemaFactory`:
  - `id` — computed string, `stringplanmodifier.UseStateForUnknown()`
  - `index` — required string, `stringplanmodifier.RequiresReplace()`, description: "Name of the target Elasticsearch index."
  - `mappings` — required string with `CustomType: index.MappingsType{}`, no plan modifier, description: "JSON mappings object to manage on the index. All top-level keys (`properties`, `dynamic`, `_source`, `dynamic_templates`, `runtime`, etc.) are supported. Only the keys and fields declared here are tracked; dynamic extras added by Elasticsearch are ignored. Destroying this resource does not remove mappings from the index (a no-op)."
- [ ] 2.2 Confirm `elasticsearch_connection` block is injected by `entitycore.ElasticsearchResource` scaffold (not manually added)
- [ ] 2.3 Set the resource schema `Description` to: "Manage a user-declared subset of index mappings on an existing Elasticsearch index. Destroy is a no-op — field mappings are not removed."


## 3. Create operation (`create.go`)

- [ ] 3.1 Implement `createIndexMappings(ctx, client, req)`:
  - Resolve `index` name from `req.Plan.Index`
  - Call `elasticsearch.GetIndex(ctx, client, indexName)` and return an error diagnostic if the index does not exist (nil result)
  - Compute resource ID via `client.ID(ctx, indexName)` and set `plan.ID`
  - Call `elasticsearch.UpdateIndexMappings(ctx, client, indexName, plan.Mappings.ValueString())`
  - Return `WriteResult{Model: plan}` with diagnostics

## 4. Read operation (`read.go`)

- [ ] 4.1 Implement `readIndexMappings(ctx, client, id, state)`:
  - Parse `indexName` from the resource ID (`id.ResourceID`)
  - Call `elasticsearch.GetIndex(ctx, client, indexName)`; if nil (not found), return `ReadResult{Removed: true}`
  - Extract the `Mappings` field from the `IndexState` response
  - Unmarshal the API response mappings JSON into a `map[string]any`
  - Unmarshal the current state `mappings` string into a `map[string]any`
  - If the state `mappings` is empty (e.g. after a passthrough import), store the full API response in `state.Mappings` and return `ReadResult{Model: state}`
  - Otherwise, build a new map containing only the top-level keys that are present in the state config
    - For the `properties` top-level key, recursively intersect: keep only field names present in the state's `properties`; discard dynamically-added fields at every nesting level
    - For all other top-level keys, retain the entire value as-is
  - Re-marshal the intersection map to JSON and store as `state.Mappings`
  - Return `ReadResult{Model: state}` with diagnostics
- [ ] 4.2 Ensure the read result uses `index.MappingsType{}` for semantic equality so unchanged mappings do not produce a diff

## 5. Update operation (`update.go`)

- [ ] 5.1 Implement `updateIndexMappings(ctx, client, req)`:
  - Call `elasticsearch.UpdateIndexMappings(ctx, client, indexName, req.Plan.Mappings.ValueString())`
  - Return `WriteResult{Model: req.Plan}` with diagnostics

## 6. Delete operation (`delete.go`)

- [ ] 6.1 Implement `deleteIndexMappings(ctx, client, id, state)`:
  - Take no API action (Elasticsearch does not support removing field mappings without a reindex)
  - Return nil diagnostics
- [ ] 6.2 Add a comment in `delete.go` explaining the no-op: field mappings in Elasticsearch are append-only; removing them requires a reindex outside the scope of this resource.

## 7. Provider registration

- [ ] 7.1 Add `indexmappings.NewIndexMappingsResource()` to the provider's resource list in `provider.go`

## 8. Acceptance tests

- [ ] 8.1 Create `internal/elasticsearch/index/indexmappings/acc_test.go` with:
  - `TestAccResourceIndexMappings_basic`: create a bare index, declare `properties` with one explicit field, verify no drift on subsequent plan
  - `TestAccResourceIndexMappings_update`: add a second field to an existing mappings resource; verify `terraform plan` shows only the diff and `terraform apply` succeeds
  - `TestAccResourceIndexMappings_drift`: declare one field; simulate dynamic field addition by calling the ES API out-of-band; verify `terraform plan` produces no diff (dynamic field is ignored)
  - `TestAccResourceIndexMappings_allTopLevelKeys`: set `dynamic`, `_source`, `properties`, and `runtime` together; verify round-trip
  - `TestAccResourceIndexMappings_destroyIsNoop`: apply then destroy; verify the index still exists and its mappings are unchanged
  - `TestAccResourceIndexMappings_indexNotFound`: configure with a non-existent index; verify create returns an error diagnostic
- [ ] 8.2 Create corresponding `testdata/TestAccResourceIndexMappings/*/` Terraform fixture files

## 9. Documentation

- [ ] 9.1 Run `make generate` (or equivalent) to regenerate the provider documentation page for `elasticstack_elasticsearch_index_mappings`
- [ ] 9.2 Verify the generated docs page renders the no-op destroy note from the schema description

## 10. Validation

- [ ] 10.1 Run `make build` and confirm the provider compiles without errors
- [ ] 10.2 Run targeted acceptance tests (`TF_ACC=1 go test ./internal/elasticsearch/index/indexmappings/... -run TestAccResourceIndexMappings -v`) against a live Elasticsearch cluster
- [ ] 10.3 Run `make check-lint` to confirm no lint regressions
- [ ] 10.4 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-index-mappings --type change` to confirm spec is valid
