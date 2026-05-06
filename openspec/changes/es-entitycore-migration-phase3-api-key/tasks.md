## 1. Model and schema factory refactoring

- [ ] 1.1 Add `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List` value-receiver methods to `tfModel` in `internal/elasticsearch/security/api_key/models.go`.
- [ ] 1.2 Convert `func (r *Resource) getSchema(version int64) schema.Schema` to a package-level factory `func getSchema(version int64) schema.Schema` that returns the schema without the `elasticsearch_connection` block.
- [ ] 1.3 Update `func (r *Resource) Schema` to delegate to `getSchema(currentSchemaVersion)` (keep the method so `ResourceWithUpgradeState` can reference the schema versions). Or remove the receiver method if no longer needed.
- [ ] 1.4 Refactor `saveClusterVersion` and `clusterVersionOfLastRead` to package-level functions that accept `ctx`, `model tfModel`, and `priv privateData` directly, removing the `*Resource` receiver.
- [ ] 1.5 Update `requiresReplaceIfUpdateNotSupported` to call the package-level `clusterVersionOfLastRead` with `res.Private`.

## 2. Callback extraction

- [ ] 2.1 Extract the body of `func (r *Resource) read(ctx context.Context, model tfModel)` into a package-level function `readAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state tfModel) (tfModel, bool, diag.Diagnostics)` in `internal/elasticsearch/security/api_key/read.go`.
- [ ] 2.2 Update the old `read` receiver method to call `readAPIKey` (or remove it once the override is in place).
- [ ] 2.3 Extract the body of `func (r *Resource) Delete` into a package-level function `deleteAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state tfModel) diag.Diagnostics`.

## 3. Resource struct migration

- [ ] 3.1 In `internal/elasticsearch/security/api_key/resource.go`, change the struct from `type Resource struct { *entitycore.ResourceBase }` to `type Resource struct { *entitycore.ElasticsearchResource[tfModel] }`.
- [ ] 3.2 Update `newResource()` to construct the resource via `entitycore.NewElasticsearchResource[tfModel](entitycore.ComponentElasticsearch, "security_api_key", schemaFactory, readAPIKey, deleteAPIKey, createPlaceholder, updatePlaceholder)` where the placeholders come from `entitycore.PlaceholderElasticsearchWriteCallbacks[tfModel]()`.
- [ ] 3.3 Remove the old `Delete` receiver method; delete behavior now lives in `deleteAPIKey` and the envelope.
- [ ] 3.4 Keep `Create` and `Update` receiver methods on `Resource` so they shadow the envelope's. No logic changes inside them.
- [ ] 3.5 Replace the old `Read` receiver method with an override that duplicates the envelope prelude (state decode, composite-ID parse, client resolution), calls `readAPIKey`, persists state or removes resource, and then calls `saveClusterVersion` when the key was found.
- [ ] 3.6 Keep `UpgradeState` on `Resource` unchanged.

## 4. Verification

- [ ] 4.1 Run `go test -count=1 ./internal/elasticsearch/security/api_key/... -run 'Test(V0ToV1|RoleDescriptors|SetUnknown|Validators)'` (or equivalent focused tests).
- [ ] 4.2 Run `make build`.
- [ ] 4.3 Run `make check-lint`.
- [ ] 4.4 Run `make check-openspec`.
- [ ] 4.5 Run targeted acceptance tests for `security_api_key` if infrastructure is available.
