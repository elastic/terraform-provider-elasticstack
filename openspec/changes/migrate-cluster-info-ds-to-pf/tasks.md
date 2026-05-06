## 1. Model and Schema

- [ ] 1.1 Define `clusterInfoDataSourceModel` struct embedding `entitycore.ElasticsearchConnectionField` with `tfsdk`-tagged fields: `ID`, `ClusterName`, `ClusterUuid`, `Name`, `Tagline`, and `Version` (as `types.List` of a nested object type)
- [ ] 1.2 Define `versionDataSourceModel` struct with `tfsdk`-tagged fields: `BuildDate`, `BuildFlavor`, `BuildHash`, `BuildSnapshot`, `BuildType`, `LuceneVersion`, `MinimumIndexCompatibilityVersion`, `MinimumWireCompatibilityVersion`, `Number`
- [ ] 1.3 Create `getDataSourceSchema()` returning `schema.Schema` with Plugin Framework attributes matching the existing SDK schema (all computed, `name` not applicable since this data source has no inputs besides connection)
- [ ] 1.4 Verify `getDataSourceSchema()` compiles and satisfies `datasource.Schema`

## 2. Read Callback

- [ ] 2.1 Implement `readDataSource(ctx, esClient, config)` callback signature: `func(context.Context, *clients.ElasticsearchScopedClient, clusterInfoDataSourceModel) (clusterInfoDataSourceModel, diag.Diagnostics)`
- [ ] 2.2 Inside callback, call `elasticsearch.GetClusterInfo(ctx, esClient)` and handle errors
- [ ] 2.3 Set the returned model's `ID = types.StringValue(info.ClusterUuid)` directly
- [ ] 2.4 Map `cluster_name`, `cluster_uuid`, `name`, `tagline` from API response to model fields
- [ ] 2.5 Build `versionDataSourceModel` with build-date type-switch logic (string/int64/fallback) and wrap in `types.ListValue` with exactly one element
- [ ] 2.6 Return populated model and diagnostics

## 3. Envelope Wiring

- [ ] 3.1 Replace `DataSourceClusterInfo() *schema.Resource` with `NewDataSource() datasource.DataSource` returning `entitycore.NewElasticsearchDataSource[clusterInfoDataSourceModel]`
- [ ] 3.2 Remove all SDK imports (`github.com/hashicorp/terraform-plugin-sdk/v2/diag`, `helper/schema`) from the data source file
- [ ] 3.3 Remove `schemautil.AddConnectionSchema` usage from the data source file
- [ ] 3.4 Remove SDK-based `dataSourceClusterInfoRead` function

## 4. Provider Registration

- [ ] 4.1 Add `cluster.NewDataSource` (or equivalent constructor name) to the `DataSources` slice in `provider/plugin_framework.go`
- [ ] 4.2 Remove `"elasticstack_elasticsearch_info": cluster.DataSourceClusterInfo()` from `provider/provider.go` `DataSourcesMap`
- [ ] 4.3 Verify `cluster` package import is still needed in `provider/provider.go` (for resources); adjust if necessary

## 5. Testing

- [ ] 5.1 Review existing `cluster_info_data_source_test.go` and update to PF testing patterns if needed (check if SDK-based tests still work via mux)
- [ ] 5.2 Run `make build` and verify no compile errors
- [ ] 5.3 Run targeted acceptance test for cluster info data source (`go test ./internal/elasticsearch/cluster/ -run TestAcc.*ClusterInfo -v`) and verify pass
