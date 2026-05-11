## 1. Model and Schema

- [x] 1.1 Define `roleDataSourceModel` struct embedding `entitycore.ElasticsearchConnectionField` with `tfsdk`-tagged fields: `ID`, `Name`, `Description`, `Cluster`, `RunAs`, `Global`, `Metadata`, `Applications`, `Indices`, `RemoteIndices`
- [x] 1.2 Define nested model structs: `applicationDataSourceModel`, `indexDataSourceModel`, `remoteIndexDataSourceModel`, `fieldSecurityDataSourceModel`
- [x] 1.3 Map SDK `TypeSet` fields to `types.Set` with `ElementType: types.StringType` or nested object types as appropriate
- [x] 1.4 Map `Global` and `Metadata` to `jsontypes.NormalizedType{}`
- [x] 1.5 Create `getDataSourceSchema()` returning `schema.Schema` with Plugin Framework nested attributes matching the current SDK shape (`schema.SetNestedAttribute` for `applications`, `indices`, `remote_indices`; `schema.ListNestedAttribute` for `field_security`)
- [x] 1.6 Verify enum-like sets (`cluster` privileges) are stored as `types.Set` of `types.String`

## 2. Read Callback

- [x] 2.1 Implement `readDataSource(ctx, esClient, config)` callback: `func(context.Context, *clients.ElasticsearchScopedClient, roleDataSourceModel) (roleDataSourceModel, diag.Diagnostics)`
- [x] 2.2 Resolve `id` via `esClient.ID(ctx, roleName)` and set on model
- [x] 2.3 Call `elasticsearch.GetRole(ctx, esClient, roleName)`; handle errors
- [x] 2.4 If role is nil with no error: set `id` to `types.StringValue("")`, return no diagnostics, and leave other computed values empty/default (matching SDK not-found behavior)
- [x] 2.5 Map scalar fields (`description`, `cluster`, `run_as`) to model
- [x] 2.6 Marshal `global` and `metadata` to JSON strings and set as normalized JSON types
- [x] 2.7 Convert `applications`, `indices`, `remote_indices` API responses to PF nested set values using adapted flatten helpers
- [x] 2.8 Handle `field_security` nested list inside `indices` and `remote_indices`

## 3. Envelope Wiring

- [x] 3.1 Replace `DataSourceRole() *schema.Resource` with `NewDataSource() datasource.DataSource` returning `entitycore.NewElasticsearchDataSource[roleDataSourceModel]`
- [x] 3.2 Remove SDK-based `dataSourceSecurityRoleRead` and SDK imports from the data source file
- [x] 3.3 Replace SDK flatten helpers with PF-native `fromAPIModel` method on `roleDataSourceModel`
- [x] 3.4 Ensure the `MinSupportedDescriptionVersion` variable (if still needed by the resource) remains accessible; do not delete shared security package variables

## 4. Provider Registration

- [x] 4.1 Add `security.NewRoleDataSource` (or equivalent) to `provider/plugin_framework.go` `DataSources` slice
- [x] 4.2 Remove `"elasticstack_elasticsearch_security_role": security.DataSourceRole()` from `provider/provider.go` `DataSourcesMap`

## 5. Testing

- [x] 5.1 Review and update `role_data_source_test.go` to PF patterns if needed (existing ProtoV6ProviderFactories + resource.TestCheck* pattern is correct and compatible)
- [x] 5.2 Run `make build` and verify no compile errors (passes cleanly)
- [ ] 5.3 Run targeted acceptance test for security role data source (`go test ./internal/elasticsearch/security/ -run '^TestAccDataSourceSecurityRole$' -v`) and verify pass (blocked: no Elastic Stack running locally)
