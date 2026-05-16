## 1. Scaffold the new componenttemplate package

- [x] 1.1 Create directory `internal/elasticsearch/index/componenttemplate/`.
- [x] 1.2 Add `models.go` defining the Plugin Framework `Data` struct with fields: `ID`, `Name`, `Metadata`, `Template`, `Version`, `ElasticsearchConnection`. Add value-receiver methods `GetID()`, `GetResourceID()` (returning `Name`), and `GetElasticsearchConnection()`.
- [x] 1.3 Add `schema.go` with a `getSchema()` factory returning `schema.Schema` without the `elasticsearch_connection` block. Port all attributes and blocks from the legacy SDK schema (`id`, `name`, `metadata`, `template` with `alias`/`mappings`/`settings`, `version`). Use `jsontypes.Normalized` for JSON strings, `schema.SingleNestedBlock` for `template`, and `schema.SetNestedBlock` for `alias`.

## 2. Port expand and flatten logic

- [x] 2.1 Add `expand.go` with `expandTemplate` and helper functions to build `models.ComponentTemplate` from `Data`. Port logic from `internal/elasticsearch/index/component_template.go` `resourceComponentTemplatePut` and shared `expandTemplate` helper.
- [x] 2.2 Add `flatten.go` with `flattenTemplateData` and helpers to map API responses back to `Data`. Preserve alias routing: pass prior-state routing into flatten so API-omitted fields are not overwritten.
- [x] 2.3 Copy or re-export any shared helpers from `internal/elasticsearch/index/` that are still needed (e.g. alias hashing) into the new package or consume them via import.

## 3. Implement envelope callbacks

- [x] 3.1 Add `read.go` with a package-level `readComponentTemplate(ctx, client, resourceID string, state Data) (Data, bool, diag.Diagnostics)` callback. Use `elasticsearch.GetComponentTemplate`. Return `(_, false, nil)` on 404/missing. Map response via flatten helpers.
- [x] 3.2 Add `delete.go` with a package-level `deleteComponentTemplate(ctx, client, resourceID string, state Data) diag.Diagnostics` callback. Use `elasticsearch.DeleteComponentTemplate`.
- [x] 3.3 Add `create.go` with a package-level `createComponentTemplate(ctx, client, resourceID string, plan Data) (Data, diag.Diagnostics)` callback. Build request body, call `elasticsearch.PutComponentTemplate`, compute composite id via `client.ID(ctx, plan.Name.ValueString())`, set `ID` on returned model.
- [x] 3.4 Add `update.go` with a package-level `updateComponentTemplate` callback identical to create (PUT semantics). Alternatively, pass the same function pointer as both create and update callbacks.

## 4. Wire the resource and provider

- [x] 4.1 Add `resource.go` defining `type Resource struct { *entitycore.ElasticsearchResource[Data] }`. In `newResource()`, call `entitycore.NewElasticsearchResource[Data]` with component `"component_template"`, schema factory, readFunc, deleteFunc, createFunc, updateFunc.
- [x] 4.2 Add `ImportState` passthrough on `id` to the concrete `Resource` type.
- [x] 4.3 Export `NewResource() resource.Resource`.
- [x] 4.4 Register `componenttemplate.NewResource()` in the provider resource registry (e.g. `provider/plugin_framework.go`).

## 5. Remove legacy SDK implementation

- [x] 5.1 Delete `internal/elasticsearch/index/component_template.go`.
- [x] 5.2 Delete `internal/elasticsearch/index/component_template_test.go`.
- [x] 5.3 Audit `internal/elasticsearch/index/template_sdk_shared.go` and other files in `internal/elasticsearch/index/` for dead code now that the SDK resource is gone. Remove only helpers that are no longer referenced.

## 6. Verification

- [ ] 6.1 `make build` passes.
- [ ] 6.2 `make check-lint` passes.
- [ ] 6.3 `make check-openspec` passes.
- [ ] 6.4 Unit tests in `internal/elasticsearch/index/componenttemplate/` pass (`go test ./...` in the new package).
- [ ] 6.5 Acceptance tests for `elasticstack_elasticsearch_component_template` pass against a running stack.
- [ ] 6.6 Confirm generated documentation is unchanged or correctly regenerated for the migrated resource.
