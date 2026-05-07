## 1. Model and Schema

- [ ] 1.1 Define `userDataSourceModel` struct embedding `entitycore.ElasticsearchConnectionField` with `tfsdk`-tagged fields: `ID`, `Username`, `FullName`, `Email`, `Roles`, `Metadata`, `Enabled`
- [ ] 1.2 Create `getDataSourceSchema()` returning `schema.Schema` with Plugin Framework attributes: `id` (computed string), `username` (required string), `full_name` (computed string), `email` (computed string), `roles` (computed set of strings), `metadata` (computed normalized JSON string), `enabled` (computed bool)
- [ ] 1.3 Verify `Roles` field uses `types.Set` with `ElementType: types.StringType`
- [ ] 1.4 Verify `Metadata` uses `jsontypes.NormalizedType{}`

## 2. Read Callback

- [ ] 2.1 Implement `readDataSource(ctx, esClient, config)` callback: `func(context.Context, *clients.ElasticsearchScopedClient, userDataSourceModel) (userDataSourceModel, diag.Diagnostics)`
- [ ] 2.2 Resolve `id` via `esClient.ID(ctx, username)` and set on model
- [ ] 2.3 Call `elasticsearch.GetUser(ctx, esClient, username)`; handle errors
- [ ] 2.4 If user is nil with no error: set `id` to `types.StringValue("")`, return no diagnostics, and leave other computed values empty/default
- [ ] 2.5 Map `full_name`, `email`, `roles`, `enabled` from API response to model
- [ ] 2.6 For `email` and `full_name`, when nil in response set `types.StringValue("")` instead of null
- [ ] 2.7 Marshal `user.Metadata` to JSON and set as normalized JSON type

## 3. Envelope Wiring

- [ ] 3.1 Replace `DataSourceUser() *schema.Resource` with `NewDataSource() datasource.DataSource` returning `entitycore.NewElasticsearchDataSource[userDataSourceModel]`
- [ ] 3.2 Remove SDK-based `dataSourceSecurityUserRead` and SDK imports from the data source file
- [ ] 3.3 Remove `schemautil.AddConnectionSchema` usage from the data source file

## 4. Provider Registration

- [ ] 4.1 Add `security.NewUserDataSource` (or equivalent) to `provider/plugin_framework.go` `DataSources` slice
- [ ] 4.2 Remove `"elasticstack_elasticsearch_security_user": security.DataSourceUser()` from `provider/provider.go` `DataSourcesMap`

## 5. Testing

- [ ] 5.1 Review and update `user_data_source_test.go` to PF patterns if needed
- [ ] 5.2 Run `make build` and verify no compile errors
- [ ] 5.3 Run targeted acceptance tests for security user data source (`go test ./internal/elasticsearch/security/ -run '^TestAccDataSourceSecurityUser' -v`) and verify pass
