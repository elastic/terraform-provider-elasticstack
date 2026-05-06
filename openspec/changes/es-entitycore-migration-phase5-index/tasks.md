## 1. Model and schema factory

- [ ] 1.1 Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` value-receiver methods to `tfModel` in `internal/elasticsearch/index/index/models.go`.
- [ ] 1.2 Convert `func (r *Resource) Schema` to package-level `func getSchema() schema.Schema` omitting `elasticsearch_connection`.

## 2. Read and delete callbacks

- [ ] 2.1 Refactor `readIndex` in `read.go` to accept `*clients.ElasticsearchScopedClient` and return `(tfModel, bool, diag.Diagnostics)` so it matches the envelope callback signature. The current helper already parses `id` internally; adjust it to use the passed `resourceID` directly.
- [ ] 2.2 Extract `deleteIndex(ctx, client, resourceID, model) diag.Diagnostics` that checks `model.DeletionProtection` and calls `elasticsearch.DeleteIndex`.

## 3. Resource struct migration

- [ ] 3.1 Change `Resource` struct to embed `*entitycore.ElasticsearchResource[tfModel]`.
- [ ] 3.2 Construct with `NewElasticsearchResource` using the schema factory, `readIndex`, `deleteIndex`, and placeholder write callbacks.
- [ ] 3.3 Keep `Create` and `Update` receiver methods unchanged. Remove `Read` and `Delete` receiver methods.
- [ ] 3.4 Keep `ImportState` unchanged.

## 4. Verification

- [ ] 4.1 Run `make build`.
- [ ] 4.2 Run `make check-lint`.
- [ ] 4.3 Run `make check-openspec`.
- [ ] 4.4 Run focused tests for the index package.
- [ ] 4.5 Run acceptance tests for `index` if infrastructure is available.
