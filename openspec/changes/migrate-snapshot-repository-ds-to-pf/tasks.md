## 1. Model and Schema

- [ ] 1.1 Define `snapshotRepositoryDataSourceModel` struct embedding `entitycore.ElasticsearchConnectionField` with `tfsdk`-tagged fields: `ID`, `Name`, `Type`, and one list-nested field per repository type (`Fs`, `Url`, `Gcs`, `Azure`, `S3`, `Hdfs`)
- [ ] 1.2 Define nested model structs for each repository type (e.g., `fsDataSourceModel`, `s3DataSourceModel`, etc.) with `tfsdk`-tagged computed fields matching the current SDK schema
- [ ] 1.3 Create `getDataSourceSchema()` returning `schema.Schema` with Plugin Framework attributes: `id` (computed string), `name` (required string), `type` (computed string), and computed list nested attributes for each repository type
- [ ] 1.4 Ensure common settings (chunk_size, compress, max_snapshot_bytes_per_sec, max_restore_bytes_per_sec, readonly) are present in each type block where applicable
- [ ] 1.5 Ensure `fs` and `url` blocks include `max_number_of_snapshots`; `s3` block does NOT include `endpoint`

## 2. Read Callback

- [ ] 2.1 Implement `readDataSource(ctx, esClient, config)` callback: `func(context.Context, *clients.ElasticsearchScopedClient, snapshotRepositoryDataSourceModel) (snapshotRepositoryDataSourceModel, diag.Diagnostics)`
- [ ] 2.2 Resolve `id` via `esClient.ID(ctx, repoName)` and set on model
- [ ] 2.3 Call `elasticsearch.GetSnapshotRepository(ctx, esClient, repoName)`; handle errors
- [ ] 2.4 If repo is nil with no error: set `id`, add warning diagnostic, return model
- [ ] 2.5 Type-switch over the typed API response (`types.Repository` union variants) to determine repository type
- [ ] 2.6 Flatten settings into the corresponding nested model using string-to-int/bool conversion logic
- [ ] 2.7 Set the matching type block as a single-element list and leave all others empty
- [ ] 2.8 If API returns an unsupported type, return error diagnostic

## 3. Envelope Wiring

- [ ] 3.1 Replace `DataSourceSnapshotRespository() *schema.Resource` with `NewDataSource() datasource.DataSource` returning `entitycore.NewElasticsearchDataSource[snapshotRepositoryDataSourceModel]`
- [ ] 3.2 Remove SDK-based `dataSourceSnapRepoRead` and SDK imports from the data source file
- [ ] 3.3 Extract or adapt `flattenRepoSettings` to build PF-compatible nested values (or create a PF-specific flattening helper alongside the existing SDK one)
- [ ] 3.4 Remove runtime schema introspection `DataSourceSnapshotRespository().Schema[currentRepo.Type]...` from the data source logic

## 4. Provider Registration

- [ ] 4.1 Add `cluster.NewSnapshotRepositoryDataSource` (or equivalent) to `provider/plugin_framework.go` `DataSources` slice
- [ ] 4.2 Remove `"elasticstack_elasticsearch_snapshot_repository": cluster.DataSourceSnapshotRespository()` from `provider/provider.go` `DataSourcesMap`

## 5. Testing

- [ ] 5.1 Review and update `snapshot_repository_data_source_test.go` to PF patterns if needed
- [ ] 5.2 Run `make build` and verify no compile errors
- [ ] 5.3 Run targeted acceptance test for snapshot repository data source (`go test ./internal/elasticsearch/cluster/ -run '^TestAccDataSourceSnapRepo' -v`) and verify pass
