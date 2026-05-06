## 1. Migrate `elasticstack_elasticsearch_index_template`

- [ ] 1.1 Add value-receiver methods `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List` to `internal/elasticsearch/index/template/models.go` `Model`.
- [ ] 1.2 Refactor `readIndexTemplate` in `internal/elasticsearch/index/template/read.go` to accept `(ctx, client, resourceID string, prior Model) (Model, bool, diag.Diagnostics)`. Inside the callback: fetch API data, apply `applyTemplateAliasReconciliationFromReference(ctx, &out, &prior)`, apply `canonicalizeTemplateAliasSetInModel(ctx, &out)`, copy `out.ID = prior.ID` and `out.ElasticsearchConnection = prior.ElasticsearchConnection`, then return.
- [ ] 1.3 Extract `deleteIndexTemplate` callback from the existing `Delete` method body in `internal/elasticsearch/index/template/delete.go`. Signature: `(ctx, client, resourceID string, state Model) diag.Diagnostics`. Body calls `elasticsearch.DeleteIndexTemplate`.
- [ ] 1.4 Remove the `Read` and `Delete` methods from `internal/elasticsearch/index/template/resource.go`.
- [ ] 1.5 Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Model]` in `template.Resource`. In `newResource()`, call `entitycore.NewElasticsearchResource[Model]` with the schema factory, `readIndexTemplate`, `deleteIndexTemplate`, and `PlaceholderElasticsearchWriteCallbacks` for create/update.
- [ ] 1.6 Strip the `elasticsearch_connection` block from `resourceSchema()` in `internal/elasticsearch/index/template/schema.go`. The envelope injects it.
- [ ] 1.7 Preserve `ImportState`, `UpgradeState`, `ModifyPlan`, and `ValidateConfig` as methods on the concrete `Resource` type.
- [ ] 1.8 Preserve the existing `Create` and `Update` methods on the concrete `Resource` type. Do NOT extract them into envelope callbacks.
- [ ] 1.9 Update interface assertions in `resource.go` to include `resource.ResourceWithModifyPlan` and `resource.ResourceWithValidateConfig`.
- [ ] 1.10 Run `go test ./internal/elasticsearch/index/template/...` and the index template acceptance tests.

## 2. Migrate `elasticstack_elasticsearch_index_lifecycle`

- [ ] 2.1 Add value-receiver methods `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List` to `internal/elasticsearch/index/ilm/models.go` `tfModel`.
- [ ] 2.2 Refactor the existing `Read` body in `internal/elasticsearch/index/ilm/read.go` into a package-level `readILM(ctx, client, resourceID string, prior tfModel) (tfModel, bool, diag.Diagnostics)` callback. Copy `ID` and `ElasticsearchConnection` from `prior` to the returned model.
- [ ] 2.3 Extract `deleteILM` callback from the existing `Delete` method body in `internal/elasticsearch/index/ilm/delete.go`. Signature: `(ctx, client, resourceID string, state tfModel) diag.Diagnostics`. Body calls `elasticsearch.DeleteIlm`.
- [ ] 2.4 Extract `createILM` callback from the existing `Create` method body in `internal/elasticsearch/index/ilm/create.go`. Signature: `(ctx, client, resourceID string, plan tfModel) (tfModel, diag.Diagnostics)`. Body: expand policy, version-gate, PUT, compute id, return model with `ID` set.
- [ ] 2.5 Extract `updateILM` callback from the existing `Update` method body in `internal/elasticsearch/index/ilm/update.go`. Signature identical to create. Body differs only in using plan name (same as create for ILM PUT semantics).
- [ ] 2.6 Remove `Read`, `Delete`, `Create`, and `Update` methods from `internal/elasticsearch/index/ilm/resource.go`.
- [ ] 2.7 Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[tfModel]` in `ilm.Resource`. In `newResource()`, call `entitycore.NewElasticsearchResource[tfModel]` with the schema factory, `readILM`, `deleteILM`, `createILM`, and `updateILM`.
- [ ] 2.8 Strip the `elasticsearch_connection` block from the schema factory in `internal/elasticsearch/index/ilm/schema.go`.
- [ ] 2.9 Preserve `ImportState` and `UpgradeState` as methods on the concrete `Resource` type.
- [ ] 2.10 Run `go test ./internal/elasticsearch/index/ilm/...` and the ILM acceptance tests.

## 3. Verification

- [ ] 3.1 `make build` passes.
- [ ] 3.2 `make check-lint` passes.
- [ ] 3.3 `make check-openspec` passes.
- [ ] 3.4 Acceptance test sweep against the running stack: `index_template` and `index_lifecycle`.
- [ ] 3.5 Generated docs unchanged: confirm `terraform-docs` / `tfplugindocs` produces no diff for the two resources.
