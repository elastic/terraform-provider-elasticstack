## 1. Migrate `elasticstack_elasticsearch_index_template_ilm_attachment` to envelope

- [ ] 1.1 Add value-receiver methods `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List` to `internal/elasticsearch/index/templateilmattachment/models.go` `tfModel`. `GetResourceID` SHALL return the derived component template name (`IndexTemplate.ValueString() + "@custom"`).
- [ ] 1.2 Refactor the existing `Read` body in `internal/elasticsearch/index/templateilmattachment/read.go` into a package-level `readILMAttachment(ctx, client, resourceID string, state tfModel) (tfModel, bool, diag.Diagnostics)` callback. Preserve the import derivation logic: when `state.IndexTemplate` is unknown/null, strip `@custom` from `resourceID` and set it. Preserve the `flat_settings=true` Get call and the "not found → false" logic.
- [ ] 1.3 Remove the `Read` method from `internal/elasticsearch/index/templateilmattachment/resource.go`.
- [ ] 1.4 Refactor the existing `Delete` body in `internal/elasticsearch/index/templateilmattachment/delete.go` into a package-level `deleteILMAttachment(ctx, client, resourceID string, state tfModel) diag.Diagnostics` callback. Preserve the read-existing → remove ILM → Put flow and the "already gone → return" short-circuit.
- [ ] 1.5 Remove the `Delete` method from `resource.go`.
- [ ] 1.6 Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[tfModel]` in `templateilmattachment.Resource`. In `newResource()`, call `entitycore.NewElasticsearchResource[tfModel]` with the schema factory, `readILMAttachment`, `deleteILMAttachment`, and `PlaceholderElasticsearchWriteCallbacks` for create/update.
- [ ] 1.7 Strip the `elasticsearch_connection` block from `getSchema()` in `internal/elasticsearch/index/templateilmattachment/schema.go`.
- [ ] 1.8 Preserve `ImportState` passthrough on the concrete `Resource` type.
- [ ] 1.9 Preserve the existing `Create` and `Update` methods on the concrete `Resource` type. Do NOT extract them into envelope callbacks.

## 2. Verification

- [ ] 2.1 `make build` passes.
- [ ] 2.2 `make check-lint` passes.
- [ ] 2.3 `make check-openspec` passes.
- [ ] 2.4 Unit tests in `internal/elasticsearch/index/templateilmattachment/` pass.
- [ ] 2.5 Acceptance tests for `elasticstack_elasticsearch_index_template_ilm_attachment` pass against a running stack.
- [ ] 2.6 Confirm that acceptance tests exercising the ILM attachment alongside `index_template` and `index_lifecycle` still pass (runtime dependency on Phase 2 resources).
