## 1. Data Source Implementation

- [ ] 1.1 Create `internal/kibana/connectors/data_source.go` with `connectorDataSourceModel` struct embedding `entitycore.KibanaConnectionField` and implementing `GetKibanaConnection()`
- [ ] 1.2 Add schema factory `getDataSourceSchema()` returning a `datasource.Schema` with attributes: `id` (Computed), `space_id` (Optional+Computed, default "default"), `name` (Required), `connector_type_id` (Optional), `connector_id` (Computed), `config` (Computed, `jsontypes.NormalizedType`), `is_deprecated` (Computed bool), `is_missing_secrets` (Computed bool), `is_preconfigured` (Computed bool)
- [ ] 1.3 Implement `readConnectorDataSource` callback: call `kibanaoapi.SearchConnectors()`, error on zero matches, error on >1 match, set `id` via `clients.CompositeID{ClusterID: spaceID, ResourceID: connectorID}`, populate all computed fields
- [ ] 1.4 Export `NewDataSource() datasource.DataSource` using `entitycore.NewKibanaDataSource[connectorDataSourceModel](entitycore.ComponentKibana, "action_connector", getDataSourceSchema, readConnectorDataSource)`

## 2. Provider Wiring

- [ ] 2.1 Register `connectors.NewDataSource` in `provider/plugin_framework.go` `dataSources()` function
- [ ] 2.2 Remove `kibana.DataSourceConnector()` from `provider/provider.go` `DataSourcesMap`

## 3. Tests

- [ ] 3.1 Move all test functions from `internal/kibana/connector_data_source_test.go` into `internal/kibana/connectors/acc_test.go` (update imports, package name to `connectors_test`)
- [ ] 3.2 Create `connectors/testdata/TestAccConnectorsDataSourceFromSDK/create/main.tf` with a config that creates a connector resource and reads it with the data source
- [ ] 3.3 Add `TestAccConnectorsDataSourceFromSDK` acceptance test using `ExternalProviders` with `VersionConstraint: "0.15.1"` for the first step, then `ProtoV6ProviderFactories` for the second step

## 4. Cleanup

- [ ] 4.1 Delete `internal/kibana/connector_data_source.go`
- [ ] 4.2 Delete `internal/kibana/connector_data_source_test.go`
- [ ] 4.3 Update `openspec/specs/kibana-action-connector/spec.md` implementation path reference from `internal/kibana/connector_data_source.go` to `internal/kibana/connectors/data_source.go`

## 5. Verification

- [ ] 5.1 `make build` passes
- [ ] 5.2 `go test ./internal/kibana/connectors/... -v -count=1 -run TestAcc` passes
