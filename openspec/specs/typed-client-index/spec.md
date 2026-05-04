# typed-client-index Specification

## Purpose
TBD - created by archiving change typed-client-index. Update Purpose after archive.
## Requirements
### Requirement: ILM helpers use typed client
`PutIlm`, `GetIlm`, and `DeleteIlm` in `internal/clients/elasticsearch/index.go` SHALL use `GetESTypedClient()` and SHALL call the typed `ILM.PutLifecycle`, `ILM.GetLifecycle`, and `ILM.DeleteLifecycle` APIs. `GetIlm` SHALL return the policy definition and SHALL return `nil` with no error when the policy is not found.

#### Scenario: Create or update ILM policy with typed client
- **WHEN** `PutIlm` is called with a valid policy configuration
- **THEN** it calls the typed `ILM.PutLifecycle` API and returns no error diagnostics on success

#### Scenario: Read existing ILM policy with typed client
- **WHEN** `GetIlm` is called for an existing policy
- **THEN** it returns the policy data and no error diagnostics

#### Scenario: Read missing ILM policy with typed client
- **GIVEN** the requested ILM policy does not exist
- **WHEN** `GetIlm` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete ILM policy with typed client
- **WHEN** `DeleteIlm` is called with a valid policy name
- **THEN** it calls the typed `ILM.DeleteLifecycle` API and returns no error diagnostics on success

### Requirement: Component template helpers use typed client
`PutComponentTemplate`, `GetComponentTemplate`, and `DeleteComponentTemplate` in `internal/clients/elasticsearch/index.go` SHALL use `GetESTypedClient()` and SHALL call the typed `Cluster.PutComponentTemplate`, `Cluster.GetComponentTemplate`, and `Cluster.DeleteComponentTemplate` APIs. `GetComponentTemplate` SHALL return `nil` with no error when the template is not found. `GetComponentTemplate` SHALL NOT request `flat_settings=true` because the typed response type `ComponentTemplateSummary.Settings` is `map[string]IndexSettings`; with flat settings enabled the API returns scalar values (e.g. `"3"`) where the decoder expects a nested object, causing a deserialization error.

#### Scenario: Create or update component template with typed client
- **WHEN** `PutComponentTemplate` is called with a valid template configuration
- **THEN** it calls the typed `Cluster.PutComponentTemplate` API and returns no error diagnostics on success

#### Scenario: Read existing component template with typed client
- **WHEN** `GetComponentTemplate` is called for an existing template
- **THEN** it returns the template data and no error diagnostics

#### Scenario: Read missing component template with typed client
- **GIVEN** the requested component template does not exist
- **WHEN** `GetComponentTemplate` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete component template with typed client
- **WHEN** `DeleteComponentTemplate` is called with a valid template name
- **THEN** it calls the typed `Cluster.DeleteComponentTemplate` API and returns no error diagnostics on success

### Requirement: Index template helpers use typed client
`PutIndexTemplate`, `GetIndexTemplate`, and `DeleteIndexTemplate` in `internal/clients/elasticsearch/index.go` SHALL use `GetESTypedClient()` and SHALL call the typed `Indices.PutIndexTemplate`, `Indices.GetIndexTemplate`, and `Indices.DeleteIndexTemplate` APIs. `GetIndexTemplate` SHALL return `nil` with no error when the template is not found.

#### Scenario: Create or update index template with typed client
- **WHEN** `PutIndexTemplate` is called with a valid template configuration
- **THEN** it calls the typed `Indices.PutIndexTemplate` API and returns no error diagnostics on success

#### Scenario: Read existing index template with typed client
- **WHEN** `GetIndexTemplate` is called for an existing template
- **THEN** it returns the template data and no error diagnostics

#### Scenario: Read missing index template with typed client
- **GIVEN** the requested index template does not exist
- **WHEN** `GetIndexTemplate` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete index template with typed client
- **WHEN** `DeleteIndexTemplate` is called with a valid template name
- **THEN** it calls the typed `Indices.DeleteIndexTemplate` API and returns no error diagnostics on success

### Requirement: Index CRUD helpers use typed client
`PutIndex`, `GetIndex`, `GetIndices`, and `DeleteIndex` in `internal/clients/elasticsearch/index.go` SHALL use `GetESTypedClient()` and SHALL call the typed `Indices.Create`, `Indices.Get`, and `Indices.Delete` APIs. `GetIndex` and `GetIndices` SHALL request `flat_settings=true` and SHALL return `nil` with no error when the index is not found. `PutIndex` SHALL preserve date-math name URI encoding and timeout options.

#### Scenario: Create index with typed client
- **WHEN** `PutIndex` is called with a valid index configuration
- **THEN** it calls the typed `Indices.Create` API and returns the concrete index name and no error diagnostics on success

#### Scenario: Create index with date math name
- **GIVEN** the configured index name is a validated date math expression
- **WHEN** `PutIndex` is called
- **THEN** the name SHALL be URI-encoded before sending in the typed API request path

#### Scenario: Read existing index with typed client
- **WHEN** `GetIndex` is called for an existing index
- **THEN** it returns the index data and no error diagnostics

#### Scenario: Read missing index with typed client
- **GIVEN** the requested index does not exist
- **WHEN** `GetIndex` or `GetIndices` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete index with typed client
- **WHEN** `DeleteIndex` is called with a valid index name
- **THEN** it calls the typed `Indices.Delete` API and returns no error diagnostics on success

### Requirement: Alias helpers use typed client
`UpdateIndexAlias`, `DeleteIndexAlias`, `GetAlias`, and `UpdateAliasesAtomic` in `internal/clients/elasticsearch/index.go` SHALL use `GetESTypedClient()` and SHALL call the typed `Indices.PutAlias`, `Indices.DeleteAlias`, `Indices.GetAlias`, and `Indices.UpdateAliases` APIs. `GetAlias` SHALL return `nil` with no error when the alias is not found. `UpdateAliasesAtomic` SHALL keep the existing `AliasAction` builder and pass the resulting actions body to the typed API.

#### Scenario: Upsert alias with typed client
- **WHEN** `UpdateIndexAlias` is called with a valid alias configuration
- **THEN** it calls the typed `Indices.PutAlias` API and returns no error diagnostics on success

#### Scenario: Delete alias with typed client
- **WHEN** `DeleteIndexAlias` is called with valid index and alias names
- **THEN** it calls the typed `Indices.DeleteAlias` API and returns no error diagnostics on success

#### Scenario: Read alias with typed client
- **WHEN** `GetAlias` is called for an existing alias
- **THEN** it returns the alias data and no error diagnostics

#### Scenario: Read missing alias with typed client
- **GIVEN** the requested alias does not exist
- **WHEN** `GetAlias` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Atomic alias update with typed client
- **WHEN** `UpdateAliasesAtomic` is called with a list of `AliasAction` values
- **THEN** it builds the actions JSON body and calls the typed `Indices.UpdateAliases` API
- **AND** returns no error diagnostics on success

### Requirement: Settings and mappings helpers use typed client
`UpdateIndexSettings` and `UpdateIndexMappings` in `internal/clients/elasticsearch/index.go` SHALL use `GetESTypedClient()` and SHALL call the typed `Indices.PutSettings` and `Indices.PutMapping` APIs.

#### Scenario: Update index settings with typed client
- **WHEN** `UpdateIndexSettings` is called with a settings map
- **THEN** it calls the typed `Indices.PutSettings` API and returns no error diagnostics on success

#### Scenario: Update index mappings with typed client
- **WHEN** `UpdateIndexMappings` is called with a mappings JSON string
- **THEN** it calls the typed `Indices.PutMapping` API and returns no error diagnostics on success

### Requirement: Data stream helpers use typed client
`PutDataStream`, `GetDataStream`, and `DeleteDataStream` in `internal/clients/elasticsearch/index.go` SHALL use `GetESTypedClient()` and SHALL call the typed `Indices.CreateDataStream`, `Indices.GetDataStream`, and `Indices.DeleteDataStream` APIs. `GetDataStream` SHALL return `nil` with no error when the data stream is not found.

#### Scenario: Create data stream with typed client
- **WHEN** `PutDataStream` is called with a valid data stream name
- **THEN** it calls the typed `Indices.CreateDataStream` API and returns no error diagnostics on success

#### Scenario: Read existing data stream with typed client
- **WHEN** `GetDataStream` is called for an existing data stream
- **THEN** it returns the data stream data and no error diagnostics

#### Scenario: Read missing data stream with typed client
- **GIVEN** the requested data stream does not exist
- **WHEN** `GetDataStream` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete data stream with typed client
- **WHEN** `DeleteDataStream` is called with a valid data stream name
- **THEN** it calls the typed `Indices.DeleteDataStream` API and returns no error diagnostics on success

### Requirement: Data stream lifecycle helpers use typed client
`PutDataStreamLifecycle`, `GetDataStreamLifecycle`, and `DeleteDataStreamLifecycle` in `internal/clients/elasticsearch/index.go` SHALL use `GetESTypedClient()` and SHALL call the typed `Indices.PutDataLifecycle`, `Indices.GetDataLifecycle`, and `Indices.DeleteDataLifecycle` APIs. `GetDataStreamLifecycle` SHALL return `nil` with no error when the lifecycle is not found.

#### Scenario: Create or update data stream lifecycle with typed client
- **WHEN** `PutDataStreamLifecycle` is called with valid lifecycle settings
- **THEN** it calls the typed `Indices.PutDataLifecycle` API and returns no error diagnostics on success

#### Scenario: Read existing data stream lifecycle with typed client
- **WHEN** `GetDataStreamLifecycle` is called for an existing lifecycle
- **THEN** it returns the lifecycle data and no error diagnostics

#### Scenario: Read missing data stream lifecycle with typed client
- **GIVEN** the requested data stream lifecycle does not exist
- **WHEN** `GetDataStreamLifecycle` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete data stream lifecycle with typed client
- **WHEN** `DeleteDataStreamLifecycle` is called with a valid data stream name
- **THEN** it calls the typed `Indices.DeleteDataLifecycle` API and returns no error diagnostics on success

### Requirement: Ingest pipeline helpers use typed client
`PutIngestPipeline`, `GetIngestPipeline`, and `DeleteIngestPipeline` in `internal/clients/elasticsearch/index.go` SHALL use `GetESTypedClient()` and SHALL call the typed `Ingest.PutPipeline`, `Ingest.GetPipeline`, and `Ingest.DeletePipeline` APIs. `GetIngestPipeline` SHALL return `nil` with no error when the pipeline is not found.

#### Scenario: Create or update ingest pipeline with typed client
- **WHEN** `PutIngestPipeline` is called with a valid pipeline configuration
- **THEN** it calls the typed `Ingest.PutPipeline` API and returns no error diagnostics on success

#### Scenario: Read existing ingest pipeline with typed client
- **WHEN** `GetIngestPipeline` is called for an existing pipeline
- **THEN** it returns the pipeline data and no error diagnostics

#### Scenario: Read missing ingest pipeline with typed client
- **GIVEN** the requested ingest pipeline does not exist
- **WHEN** `GetIngestPipeline` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete ingest pipeline with typed client
- **WHEN** `DeleteIngestPipeline` is called with a valid pipeline name
- **THEN** it calls the typed `Ingest.DeletePipeline` API and returns no error diagnostics on success

### Requirement: Index resources consume typed helpers
All index-related resource packages under `internal/elasticsearch/index/` and `internal/elasticsearch/ingest/` SHALL compile successfully after `index.go` helper signatures and internals are updated to use the typed client.

#### Scenario: ILM resource compiles after migration
- **WHEN** compiling `internal/elasticsearch/index/ilm`
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Component template resource compiles after migration
- **WHEN** compiling `internal/elasticsearch/index/component_template`
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Index template resource compiles after migration
- **WHEN** compiling `internal/elasticsearch/index/template`
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Index resource compiles after migration
- **WHEN** compiling `internal/elasticsearch/index/index`
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Alias resource compiles after migration
- **WHEN** compiling `internal/elasticsearch/index/alias`
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Data stream resource compiles after migration
- **WHEN** compiling `internal/elasticsearch/index/data_stream`
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Data stream lifecycle resource compiles after migration
- **WHEN** compiling `internal/elasticsearch/index/datastreamlifecycle`
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Index template ILM attachment resource compiles after migration
- **WHEN** compiling `internal/elasticsearch/index/templateilmattachment`
- **THEN** it succeeds with no errors referencing the updated helper signatures

#### Scenario: Ingest pipeline resource compiles after migration
- **WHEN** compiling `internal/elasticsearch/ingest/pipeline`
- **THEN** it succeeds with no errors referencing the updated helper signatures

### Requirement: Redundant custom index model structs removed
The custom model types in `internal/models/models.go` that are fully replaced by typed API equivalents SHALL be removed once all callers have been migrated. This includes at minimum `ComponentTemplateResponse`, `ComponentTemplatesResponse`, `IndexTemplateResponse`, `IndexTemplatesResponse`, `DataStream`, `DataStreamLifecycle`, `IngestPipeline`, `PolicyDefinition`, `Policy`, `IndexAlias`, and related wrapper structs.

#### Scenario: Custom index types have no remaining references
- **GIVEN** all callers have been updated to use typed client types
- **WHEN** searching the codebase for the removed custom model types
- **THEN** no references exist in compiling source files

### Requirement: Project builds successfully after index migration
`make build` SHALL complete without errors after all index files are migrated and redundant models are removed.

#### Scenario: Clean build after index migration
- **WHEN** running `make build`
- **THEN** the command exits with status 0 and produces no compilation errors

### Requirement: Lint checks pass after index migration
`make check-lint` SHALL complete without new lint failures introduced by the typed-client migration.

#### Scenario: Lint passes after index migration
- **WHEN** running `make check-lint`
- **THEN** the command exits with status 0

