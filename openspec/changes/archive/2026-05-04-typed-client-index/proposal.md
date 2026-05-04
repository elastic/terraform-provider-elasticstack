## Why

The provider-wide typed-client migration is moving through its phases. The helpers in `internal/clients/elasticsearch/index.go` are the largest single client file (~811 lines) and the last major Elasticsearch surface still using the raw `esapi` client. Migrating this file to the `go-elasticsearch` Typed API (`elasticsearch.TypedClient` via `GetESTypedClient()`) eliminates hand-structured JSON request bodies, manual response decoding, and custom model types that duplicate upstream structs.

## What Changes

Migrate `internal/clients/elasticsearch/index.go` and all callers from raw `esapi` to typed client APIs:

- **ILM** — `PutIlm`, `GetIlm`, `DeleteIlm` → `TypedClient.ILM.PutLifecycle` / `GetLifecycle` / `DeleteLifecycle`
- **Component templates** — `PutComponentTemplate`, `GetComponentTemplate`, `DeleteComponentTemplate` → `TypedClient.Cluster.PutComponentTemplate` / `GetComponentTemplate` / `DeleteComponentTemplate`
- **Index templates** — `PutIndexTemplate`, `GetIndexTemplate`, `DeleteIndexTemplate` → `TypedClient.Indices.PutIndexTemplate` / `GetIndexTemplate` / `DeleteIndexTemplate`
- **Index CRUD** — `PutIndex`, `GetIndex`, `GetIndices`, `DeleteIndex` → `TypedClient.Indices.Create` / `Get` / `Delete`
- **Aliases** — `UpdateIndexAlias`, `DeleteIndexAlias`, `GetAlias`, `UpdateAliasesAtomic` → `TypedClient.Indices.PutAlias` / `DeleteAlias` / `GetAlias` / `UpdateAliases`
- **Settings & mappings** — `UpdateIndexSettings`, `UpdateIndexMappings` → `TypedClient.Indices.PutSettings` / `PutMapping`
- **Data streams** — `PutDataStream`, `GetDataStream`, `DeleteDataStream` → `TypedClient.Indices.CreateDataStream` / `GetDataStream` / `DeleteDataStream`
- **Data stream lifecycle** — `PutDataStreamLifecycle`, `GetDataStreamLifecycle`, `DeleteDataStreamLifecycle` → `TypedClient.Indices.PutDataLifecycle` / `GetDataLifecycle` / `DeleteDataLifecycle`
- **Ingest pipelines** — `PutIngestPipeline`, `GetIngestPipeline`, `DeleteIngestPipeline` → `TypedClient.Ingest.PutPipeline` / `GetPipeline` / `DeletePipeline`

Update all downstream resources, data sources, and tests under:
- `internal/elasticsearch/index/*`
- `internal/elasticsearch/ingest/*`
- `internal/acctest/*` (any index/ingest helpers)

Remove or deprecate custom model types that become redundant when typed API equivalents are available (e.g. `ComponentTemplateResponse`, `IndexTemplateResponse`, `DataStream`, `DataStreamLifecycle`, `IngestPipeline`, `PolicyDefinition`, `AliasAction`, etc.).

No Terraform resource schemas, provider configuration, or user-visible behavior changes.

## Capabilities

### New Capabilities
_(none — this is an internal refactoring with no new user-visible capabilities)_

### Modified Capabilities
_(none — no spec-level requirement changes; all resource schemas, validation, and behavior remain identical)_

## Impact

- **Code**: `internal/clients/elasticsearch/index.go` (~29KB), downstream resource files under `internal/elasticsearch/index/` and `internal/elasticsearch/ingest/`, and acceptance-test helpers.
- **APIs**: No Terraform resource or data source behavior changes.
- **Dependencies**: Relies on existing `go-elasticsearch/v8` `GetESTypedClient()` already exposed via `ElasticsearchScopedClient`.
- **Build / CI**: Compilation and acceptance tests are affected; no new dependencies introduced.
