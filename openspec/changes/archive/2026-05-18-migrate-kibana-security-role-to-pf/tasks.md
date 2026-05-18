## 1. Package Skeleton and Shared Models

- [x] 1.1 Create `internal/kibana/security_role/models.go` with all nested model types: `ESModel`, `ESIndexModel`, `ESRemoteIndexModel`, `FieldSecurityModel`, `KibanaModel`, `KibanaFeatureModel`; all fields use `types.Set`, `types.String`, or `jsontypes.Normalized` as appropriate
- [x] 1.2 Add `resourceModel` to `models.go` embedding `entitycore.KibanaConnectionField`; implement `GetID() types.String` (returns `Name`), `GetResourceID() types.String` (returns `Name`), `GetSpaceID() types.String` (returns `types.StringValue("")`), `GetVersionRequirements()` returning `>= 8.10.0` when `RemoteIndices` is non-empty and `>= 8.15.0` when `Description` is non-null/non-empty (both when both apply)
- [x] 1.3 Add `dataSourceModel` to `models.go` embedding `entitycore.KibanaConnectionField`; no `GetVersionRequirements()` needed

## 2. Flatten Helpers

- [x] 2.1 Create `internal/kibana/security_role/flatten.go` with `flattenESModel()` mapping `kbapi` response fields to `ESModel` (omit empty `cluster`/`run_as` slices)
- [x] 2.2 Add `flattenESIndexModel()` mapping index privilege response fields to `ESIndexModel` (omit `query` when empty string, populate `FieldSecurity` when present)
- [x] 2.3 Add `flattenESRemoteIndexModel()` mapping remote index fields to `ESRemoteIndexModel`
- [x] 2.4 Add `flattenKibanaModel()` mapping `kibana` privilege entries to `KibanaModel` (including `base` and `feature` entries)
- [x] 2.5 Add `expandResourceModel()` / `expandESModel()` / `expandKibanaModel()` helpers that build `kbapi` request body types from plan model (omit null-optional fields, handle `base`/`feature` XOR, serialize `metadata` JSON)

## 3. Read Helpers

- [x] 3.1 Create `internal/kibana/security_role/read.go` with `fetchRole(ctx, client *clients.KibanaScopedClient, name string) (*kbapi.SecurityRole, bool, diag.Diagnostics)` calling `kibanaoapi.GetSecurityRole()`; return `found=false` on 404/not-found
- [x] 3.2 Add `readRoleResource` callback with entitycore resource signature: calls `fetchRole()`, maps response into `resourceModel` using flatten helpers, returns `found=false` to trigger state removal when role is absent
- [x] 3.3 Add `readRoleDataSource` callback with entitycore data source signature: calls `fetchRole()` using `config.Name`, maps response into `dataSourceModel`; when the role is absent, clears computed attributes and returns without an error diagnostic (preserves pre-migration SDK / REQ-012 behavior so `terraform plan` shows nulls rather than failing)

## 4. Schema

- [x] 4.1 Create `internal/kibana/security_role/schema.go` with `getResourceSchema()` returning a `resource.Schema` replicating the SDK schema exactly:
  - `name`: Required, `stringplanmodifier.RequiresReplace()`
  - `id`: Computed, `UseStateForUnknown`
  - `description`: Optional
  - `metadata`: Optional+Computed, `jsontypes.NormalizedType`
  - `elasticsearch`: Required, `SetNestedBlock` (max 1), containing `cluster`, `run_as`, `indices`, `remote_indices` with nested field_security
  - `kibana`: Optional, `SetNestedBlock`, containing `spaces`, `base`, `feature` (with nested `name`/`privileges`)
- [x] 4.2 Add `getDataSourceSchema()` returning a `datasource.Schema` with all attributes set to Computed (except `name` which is Required)
- [x] 4.3 Add a `kibana` block `ObjectValidator` enforcing `base` XOR `feature` mutual exclusivity and requiring at least one; attach to resource schema as a `ConfigValidator` or block-level validator

## 5. CRUD Callbacks

- [x] 5.1 Create `internal/kibana/security_role/create.go`: call `expandResourceModel()`, call `kibanaoapi.PutSecurityRole()` with `createOnly=true`, call `fetchRole()` to refresh state, set `ID = Name`
- [x] 5.2 Create `internal/kibana/security_role/update.go`: call `expandResourceModel()`, call `kibanaoapi.PutSecurityRole()` with `createOnly=false`, call `fetchRole()` to refresh state
- [x] 5.3 Create `internal/kibana/security_role/delete.go`: call `kibanaoapi.DeleteSecurityRole()` with role name from `resourceID`

## 6. Resource and Data Source Entry Points

- [x] 6.1 Create `internal/kibana/security_role/resource.go` with `type Resource struct { *entitycore.KibanaResource[resourceModel] }` and `NewResource() resource.Resource` using `entitycore.NewKibanaResource[resourceModel]`
- [x] 6.2 Add `ImportState` method to `Resource` using `resource.ImportStatePassthroughID`
- [x] 6.3 Create `internal/kibana/security_role/data_source.go` with `NewDataSource() datasource.DataSource` using `entitycore.NewKibanaDataSource[dataSourceModel]`

## 7. Provider Wiring

- [x] 7.1 Register `security_role.NewResource` in `provider/plugin_framework.go` `resources()` function
- [x] 7.2 Register `security_role.NewDataSource` in `provider/plugin_framework.go` `dataSources()` function
- [x] 7.3 Remove `kibana.ResourceRole()` from `provider/provider.go` `ResourcesMap`
- [x] 7.4 Remove `kibana.DataSourceRole()` from `provider/provider.go` `DataSourcesMap`

## 8. Tests

- [x] 8.1 Create `internal/kibana/security_role/acc_test.go` (package `security_role_test`); move all test functions from `internal/kibana/role_test.go` and `internal/kibana/role_data_source_test.go`; update imports and package name
- [x] 8.2 Add unit tests in `internal/kibana/security_role/flatten_test.go` asserting stable round-trip for: one `indices` entry with `field_security`, one `remote_indices` entry, one `kibana` feature block, one `kibana` base block
- [x] 8.3 Create `security_role/testdata/TestAccKibanaSecurityRoleResourceFromSDK/create/main.tf` with a config creating a role with `elasticsearch` and `kibana` blocks
- [x] 8.4 Add `TestAccKibanaSecurityRoleResourceFromSDK` acceptance test: step 1 uses `ExternalProviders` with `VersionConstraint: "0.15.1"`, step 2 uses `ProtoV6ProviderFactories`

## 9. Cleanup

- [x] 9.1 Delete `internal/kibana/role.go`
- [x] 9.2 Delete `internal/kibana/role_data_source.go`
- [x] 9.3 Delete `internal/kibana/role_test.go`
- [x] 9.4 Delete `internal/kibana/role_data_source_test.go`
- [x] 9.5 Update `openspec/specs/kibana-security-role/spec.md` implementation path references

## 10. Verification

- [x] 10.1 `make build` passes
- [x] 10.2 Unit tests pass: `go test ./internal/kibana/security_role/... -v -count=1`
- [ ] 10.3 Acceptance tests pass: `go test ./internal/kibana/security_role/... -v -count=1 -run TestAcc` (validate in CI — no local Elastic stack in this worktree)
