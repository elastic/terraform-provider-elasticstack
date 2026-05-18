## Why

Both `elasticstack_kibana_security_role` entities — resource and data source — remain on `terraform-plugin-sdk/v2`. They are the last SDK-based entities in the Kibana security domain. The resource (`role.go`, 766 LOC) and data source (`role_data_source.go`, 258 LOC) share schema structure and read logic via direct function delegation, making them natural to migrate together into a single Package Framework package.

## What Changes

- Create new package `internal/kibana/security_role/` containing:
  - `resource.go` — `entitycore.NewKibanaResource[resourceModel]`
  - `data_source.go` — `entitycore.NewKibanaDataSource[dataSourceModel]`
  - `models.go` — `resourceModel`, `dataSourceModel`, and all shared nested model types
  - `schema.go` — resource and data source schema factories plus shared attribute/block helpers
  - `read.go` — `fetchRole()` shared helper plus the two read callbacks
  - `create.go`, `update.go`, `delete.go` — resource lifecycle callbacks
- Wire `security_role.NewResource` and `security_role.NewDataSource` in `provider/plugin_framework.go`
- Remove `kibana.ResourceRole()` and `kibana.DataSourceRole()` from `provider/provider.go`
- Delete `internal/kibana/role.go` and `internal/kibana/role_data_source.go`
- Move acceptance tests from `internal/kibana/role_test.go` and `internal/kibana/role_data_source_test.go` into `internal/kibana/security_role/acc_test.go`
- Add SDK upgrade test (`TestAccKibanaSecurityRoleResourceFromSDK`) with `VersionConstraint: "0.15.1"`
- Update `openspec/specs/kibana-security-role/spec.md` implementation path references

## Capabilities

### New Capabilities

None. Both the resource and data source schemas and behaviors are unchanged.

### Modified Capabilities

- `kibana-security-role`: Implementation paths change from `internal/kibana` to `internal/kibana/security_role/`. No schema or behavioral requirement changes.

## Impact

- `internal/kibana/security_role/` — new package (all new files)
- `internal/kibana/role.go` — deleted
- `internal/kibana/role_data_source.go` — deleted
- `provider/provider.go` — two entries removed (resource + data source)
- `provider/plugin_framework.go` — two entries added
- `openspec/specs/kibana-security-role/spec.md` — implementation path updated
