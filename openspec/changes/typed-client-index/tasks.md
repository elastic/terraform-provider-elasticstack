## 1. Typed client migration — ILM

- [ ] 1.1 Replace `PutIlm` raw `esapi` call with typed `ILM.PutLifecycle` using `types.IlmPolicy` or equivalent.
- [ ] 1.2 Replace `GetIlm` raw call with typed `ILM.GetLifecycle`; map 404 → nil.
- [ ] 1.3 Replace `DeleteIlm` raw call with typed `ILM.DeleteLifecycle`.
- [ ] 1.4 Remove `models.Policy` and `models.PolicyDefinition` if no longer used; keep conversion helpers if resources need them.

## 2. Typed client migration — Component templates

- [ ] 2.1 Replace `PutComponentTemplate` raw call with typed `Cluster.PutComponentTemplate`.
- [ ] 2.2 Replace `GetComponentTemplate` raw call with typed `Cluster.GetComponentTemplate` using `FlatSettings(true)`.
- [ ] 2.3 Replace `DeleteComponentTemplate` raw call with typed `Cluster.DeleteComponentTemplate`.
- [ ] 2.4 Remove or narrow `models.ComponentTemplate`, `models.ComponentTemplatesResponse`, `models.ComponentTemplateResponse`.

## 3. Typed client migration — Index templates

- [ ] 3.1 Replace `PutIndexTemplate` raw call with typed `Indices.PutIndexTemplate`.
- [ ] 3.2 Replace `GetIndexTemplate` raw call with typed `Indices.GetIndexTemplate`.
- [ ] 3.3 Replace `DeleteIndexTemplate` raw call with typed `Indices.DeleteIndexTemplate`.
- [ ] 3.4 Remove or narrow `models.IndexTemplate`, `models.IndexTemplatesResponse`, `models.IndexTemplateResponse`.

## 4. Typed client migration — Index CRUD

- [ ] 4.1 Replace `PutIndex` raw call with typed `Indices.Create`; preserve date-math name URI encoding and timeout options.
- [ ] 4.2 Replace `DeleteIndex` raw call with typed `Indices.Delete`.
- [ ] 4.3 Replace `GetIndex` / `GetIndices` raw calls with typed `Indices.Get` using `FlatSettings(true)`; preserve 404 → nil semantics.
- [ ] 4.4 Remove or narrow `models.Index`, `models.PutIndexParams`.

## 5. Typed client migration — Aliases

- [ ] 5.1 Replace `DeleteIndexAlias` raw call with typed `Indices.DeleteAlias`.
- [ ] 5.2 Replace `UpdateIndexAlias` raw call with typed `Indices.PutAlias`.
- [ ] 5.3 Replace `GetAlias` raw call with typed `Indices.GetAlias`.
- [ ] 5.4 Replace `UpdateAliasesAtomic` raw call with typed `Indices.UpdateAliases`; keep the existing `AliasAction` builder.
- [ ] 5.5 Remove or narrow `models.IndexAlias` if fully replaced by `types.Alias`.

## 6. Typed client migration — Settings and mappings

- [ ] 6.1 Replace `UpdateIndexSettings` raw call with typed `Indices.PutSettings`.
- [ ] 6.2 Replace `UpdateIndexMappings` raw call with typed `Indices.PutMapping`.

## 7. Typed client migration — Data streams

- [ ] 7.1 Replace `PutDataStream` raw call with typed `Indices.CreateDataStream`.
- [ ] 7.2 Replace `GetDataStream` raw call with typed `Indices.GetDataStream`.
- [ ] 7.3 Replace `DeleteDataStream` raw call with typed `Indices.DeleteDataStream`.
- [ ] 7.4 Remove or narrow `models.DataStream`.

## 8. Typed client migration — Data stream lifecycle

- [ ] 8.1 Replace `PutDataStreamLifecycle` raw call with typed `Indices.PutDataLifecycle`.
- [ ] 8.2 Replace `GetDataStreamLifecycle` raw call with typed `Indices.GetDataLifecycle`.
- [ ] 8.3 Replace `DeleteDataStreamLifecycle` raw call with typed `Indices.DeleteDataLifecycle`.
- [ ] 8.4 Remove or narrow `models.DataStreamLifecycle`, `models.LifecycleSettings`, `models.Downsampling`.

## 9. Typed client migration — Ingest pipelines

- [ ] 9.1 Replace `PutIngestPipeline` raw call with typed `Ingest.PutPipeline`.
- [ ] 9.2 Replace `GetIngestPipeline` raw call with typed `Ingest.GetPipeline`.
- [ ] 9.3 Replace `DeleteIngestPipeline` raw call with typed `Ingest.DeletePipeline`.
- [ ] 9.4 Remove or narrow `models.IngestPipeline`.

## 10. Resource and test updates

- [ ] 10.1 Update `internal/elasticsearch/index/ilm` resource and acceptance tests to use migrated helpers.
- [ ] 10.2 Update `internal/elasticsearch/index/component_template` resource and tests.
- [ ] 10.3 Update `internal/elasticsearch/index/template` resource and tests.
- [ ] 10.4 Update `internal/elasticsearch/index/index` resource and tests.
- [ ] 10.5 Update `internal/elasticsearch/index/alias` resource and tests.
- [ ] 10.6 Update `internal/elasticsearch/index/data_stream` resource and tests.
- [ ] 10.7 Update `internal/elasticsearch/index/datastreamlifecycle` resource and tests.
- [ ] 10.8 Update `internal/elasticsearch/index/templateilmattachment` resource and tests.
- [ ] 10.9 Update `internal/elasticsearch/ingest/pipeline` resource and tests.

## 11. Verification and cleanup

- [ ] 11.1 Run `make build` and fix all compilation errors.
- [ ] 11.2 Run `make check-lint` and fix any issues.
- [ ] 11.3 Run targeted acceptance tests for at least one resource per group (ILM, component template, index template, index, alias, data stream, data stream lifecycle, ingest pipeline).
- [ ] 11.4 Run the full acceptance test suite for all affected resources if feasible.
- [ ] 11.5 Prune any now-unused imports or models in `internal/models` and `internal/clients/elasticsearch/index.go`.
