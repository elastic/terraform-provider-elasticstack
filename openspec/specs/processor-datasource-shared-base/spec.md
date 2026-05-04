# Purpose

Define the shared base infrastructure for ingest processor data sources migrated to the Terraform Plugin Framework.

## Requirements

### Requirement: Generic processor data source base
The system SHALL provide a generic `processorDataSource[T ProcessorModel]` struct that implements `datasource.DataSource` and `datasource.DataSourceWithConfigure`. It SHALL own the `Metadata`, `Read`, and `Configure` methods, eliminating the need for per-processor `Read` implementations.

#### Scenario: Generic Read executes without per-processor code
- GIVEN a processor data source constructed with a concrete type `T` satisfying `ProcessorModel`
- WHEN `Read` is invoked with a Terraform configuration
- THEN the data source SHALL decode the config into `T`, call `T.MarshalBody()`, wrap the result as `{"<name>": body}`, marshal to indented JSON, hash the JSON for `id`, and set state
- AND no processor-specific `Read` function SHALL be required

### Requirement: ProcessorModel interface
The system SHALL define a `ProcessorModel` interface that is satisfied by any struct providing `TypeName() string`, `MarshalBody() (map[string]any, diag.Diagnostics)`, `SetID(string)`, and `SetJSON(string)`.

#### Scenario: Model implements ProcessorModel
- GIVEN a processor model struct implementing all four methods
- WHEN the model is used as the type parameter `T` for `processorDataSource[T]`
- THEN it SHALL compile and satisfy the generic constraint

### Requirement: Common processor fields helper
The system SHALL provide `CommonProcessorModel` (a struct with `tfsdk` tags for `description`, `if`, `ignore_failure`, `on_failure`, `tag`), `CommonProcessorSchemaAttributes()` returning the common schema attributes, and `toCommonProcessorBody()` for use within `MarshalBody()`.

#### Scenario: Common fields merged into processor schema
- GIVEN a processor schema factory that merges its specific attributes with `CommonProcessorSchemaAttributes()`
- WHEN the schema is returned by the data source
- THEN it SHALL include `id`, `json`, `description`, `if`, `ignore_failure`, `on_failure`, and `tag`

### Requirement: No Elasticsearch connection required
Processor data sources SHALL NOT require or expose an `elasticsearch_connection` block. The generic `Configure` method SHALL be a no-op.

#### Scenario: Processor data source read without connection
- GIVEN a processor data source configured without any connection block
- WHEN the data source is read
- THEN it SHALL succeed and produce valid `json` and `id` outputs
