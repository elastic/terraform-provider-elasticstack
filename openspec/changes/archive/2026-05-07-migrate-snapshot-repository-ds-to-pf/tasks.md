## 1. Model and Schema

- [x] 1.1 Define `snapshotRepositoryDataSourceModel` struct embedding `entitycore.ElasticsearchConnectionField` with `tfsdk`-tagged fields: `ID`, `Name`, `Type`, and one list-nested field per repository type (`Fs`, `Url`, `Gcs`, `Azure`, `S3`, `Hdfs`)
- [x] 1.2 Define nested model structs for each repository type (e.g., `fsDataSourceModel`, `s3DataSourceModel`, etc.) with `tfsdk`-tagged computed fields matching the current SDK schema
- [x] 1.3 Create `getDataSourceSchema()` returning `schema.Schema` with Plugin Framework attributes: `id` (computed string), `name` (required string), `type` (computed string), and computed list nested attributes for each repository type
- [x] 1.4 Ensure common settings (chunk_size, compress, max_snapshot_bytes_per_sec, max_restore_bytes_per_sec, readonly) are present in each type block where applicable
- [x] 1.5 Ensure `fs` and `url` blocks include `max_number_of_snapshots`; `s3` block does NOT include `endpoint`

## 2. Read Callback

- [x] 2.1 Implement `readDataSource(ctx, esClient, config)` callback: `func(context.Context, *clients.ElasticsearchScopedClient, snapshotRepositoryDataSourceModel) (snapshotRepositoryDataSourceModel, diag.Diagnostics)`
- [x] 2.2 Resolve `id` via `esClient.ID(ctx, repoName)` and set on model
- [x] 2.3 Call `elasticsearch.GetSnapshotRepository(ctx, esClient, repoName)`; handle errors
- [x] 2.4 If repo is nil with no error: set `id`, add warning diagnostic, return model
- [x] 2.5 Type-switch over the typed API response (`types.Repository` union variants) to determine repository type
- [x] 2.6 Flatten settings into the corresponding nested model using string-to-int/bool conversion logic
- [x] 2.7 Set the matching type block as a single-element list and leave all others empty
- [x] 2.8 If API returns an unsupported type, return error diagnostic

## 3. Envelope Wiring

- [x] 3.1 Replace `DataSourceSnapshotRespository() *schema.Resource` with `NewDataSource() datasource.DataSource` returning `entitycore.NewElasticsearchDataSource[snapshotRepositoryDataSourceModel]`
- [x] 3.2 Remove SDK-based `dataSourceSnapRepoRead` and SDK imports from the data source file
- [x] 3.3 Extract or adapt `flattenRepoSettings` to build PF-compatible nested values (or create a PF-specific flattening helper alongside the existing SDK one)
- [x] 3.4 Remove runtime schema introspection `DataSourceSnapshotRespository().Schema[currentRepo.Type]...` from the data source logic

## 4. Provider Registration

- [x] 4.1 Add `cluster.NewSnapshotRepositoryDataSource` (or equivalent) to `provider/plugin_framework.go` `DataSources` slice
- [x] 4.2 Remove `"elasticstack_elasticsearch_snapshot_repository": cluster.DataSourceSnapshotRespository()` from `provider/provider.go` `DataSourcesMap`

## 5. Testing

- [x] 5.1 Review and update `snapshot_repository_data_source_test.go` to PF patterns if needed
- [x] 5.2 Run `make build` and verify no compile errors
- [x] 5.3 Run targeted acceptance test for snapshot repository data source (`go test ./internal/elasticsearch/cluster/ -run '^TestAccDataSourceSnapRepo' -v`) and verify pass
