# entitycore-datasource-envelope Specification

## Purpose
TBD - created by archiving change envelope-datasource-generics. Update Purpose after archive.
## Requirements
### Requirement: Envelope constructor produces a valid DataSource
The system SHALL provide a generic constructor `NewKibanaDataSource[T]()` (and `NewElasticsearchDataSource[T]()`) that returns a value satisfying `datasource.DataSource`.

#### Scenario: Constructor returns valid data source
- **WHEN** `NewKibanaDataSource[T](component, name, schema, readFunc)` is called
- **THEN** the returned value SHALL satisfy `datasource.DataSource`
- **AND** the returned value SHALL satisfy `datasource.DataSourceWithConfigure`

### Requirement: Envelope injects connection block into schema
The system SHALL inject the scoped connection block (`kibana_connection` or `elasticsearch_connection`) into the schema before exposing it via the `Schema` method.

#### Scenario: Schema includes injected connection block
- **WHEN** a data source is constructed with `NewKibanaDataSource[T]` and a schema that lacks a `kibana_connection` block
- **THEN** calling `Schema` on the resulting data source SHALL return a schema that includes the `kibana_connection` block
- **AND** the concrete schema attributes SHALL remain unchanged

### Requirement: Envelope owns config deserialization
The system SHALL decode the Terraform config into the generic model `T`, which embeds the connection block field (via `KibanaConnectionField` or `ElasticsearchConnectionField`).

#### Scenario: Config decode populates concrete model fields
- **WHEN** `Read` is invoked with a Terraform config containing both `kibana_connection` and concrete attributes
- **THEN** the model `T` SHALL be deserialized via `req.Config.Get(ctx, &model)`, populating both the connection block (via the embedded helper) and the concrete entity attributes

### Requirement: Envelope owns scoped client resolution
The system SHALL resolve the scoped client from the provider factory using the captured connection block value.

#### Scenario: Scoped client resolved from connection block
- **WHEN** `Read` captures a non-empty `kibana_connection` block
- **THEN** the system SHALL call `GetKibanaClient` with that connection value
- **AND** pass the resulting `*KibanaScopedClient` to the concrete read function

#### Scenario: Scoped client resolved from provider defaults
- **WHEN** `Read` captures an empty or null `kibana_connection` block
- **THEN** the system SHALL call `GetKibanaClient` with the captured connection block value
- **AND** when that captured value is null or empty, the factory SHALL return the provider-default scoped client

### Requirement: Envelope delegates entity logic to read function
The system SHALL invoke the concrete read function with the scoped client and deserialized model, then capture the returned model and diagnostics.

#### Scenario: Read function receives client and model
- **WHEN** `Read` has successfully deserialized config and resolved the client
- **THEN** the concrete read function SHALL be called with `(context, *KibanaScopedClient, T)`
- **AND** its returned `(T, diag.Diagnostics)` SHALL be used for state setting

### Requirement: Envelope owns state persistence
The system SHALL set the Terraform state from the model returned by the concrete read function. Because `T` embeds the connection block field, the returned model naturally preserves the original connection block value.

#### Scenario: State set after successful read
- **WHEN** the concrete read function returns a model without error diagnostics
- **THEN** `resp.State.Set` SHALL be called with the returned model `T`
- **AND** the connection block value present in the returned model SHALL reflect the original config value

#### Scenario: State not set on read function error
- **WHEN** the concrete read function returns error diagnostics
- **THEN** `resp.State.Set` SHALL NOT be called
- **AND** the error diagnostics SHALL be appended to `resp.Diagnostics`

### Requirement: Kibana envelope enforces optional model version requirements

The Kibana data source envelope SHALL allow a decoded model to optionally declare pre-read server version requirements. When the model implements the optional version-requirements interface, the envelope SHALL evaluate those requirements after resolving the scoped Kibana client and before invoking the concrete read function.

Version requirements SHALL remain optional. A Kibana data source model that only satisfies the base `KibanaDataSourceModel` contract SHALL continue through the existing read flow without defining no-op version requirements.

#### Scenario: Model without version requirements reads normally

- **GIVEN** a Kibana envelope data source whose model does not implement the optional version-requirements interface
- **WHEN** `Read` successfully decodes config and resolves the scoped Kibana client
- **THEN** the envelope SHALL invoke the concrete read function without attempting model-specific version enforcement
- **AND** state persistence SHALL follow the existing envelope behavior

#### Scenario: Supported server invokes read function

- **GIVEN** a Kibana envelope data source whose model declares a minimum server version requirement
- **AND** the scoped Kibana client reports that the server satisfies that minimum version
- **WHEN** `Read` evaluates the version requirement
- **THEN** the envelope SHALL invoke the concrete read function
- **AND** the read result SHALL be used for state persistence according to existing envelope behavior

#### Scenario: Unsupported server stops before read function

- **GIVEN** a Kibana envelope data source whose model declares a minimum server version requirement with an error message
- **AND** the scoped Kibana client reports that the server does not satisfy that minimum version
- **WHEN** `Read` evaluates the version requirement
- **THEN** the envelope SHALL add an `Unsupported server version` diagnostic using the model-provided error message
- **AND** the concrete read function SHALL NOT be invoked
- **AND** Terraform state SHALL NOT be set from a read result

#### Scenario: Version requirement diagnostics stop read

- **GIVEN** a Kibana envelope data source whose model implements the optional version-requirements interface
- **AND** collecting or enforcing the requirements returns error diagnostics
- **WHEN** `Read` evaluates version requirements
- **THEN** the envelope SHALL append those diagnostics to the read response
- **AND** the concrete read function SHALL NOT be invoked
- **AND** Terraform state SHALL NOT be set from a read result

### Requirement: Envelope is the sole supported Plugin Framework data source pattern
The system SHALL provide only the generic envelope constructors for Plugin Framework data source wiring. No struct-based alternative SHALL exist within `internal/entitycore`. All existing Plugin Framework data sources SHALL be constructed via `NewKibanaDataSource` or `NewElasticsearchDataSource`.

#### Scenario: All PF data sources use envelope constructor
- **WHEN** the provider enumerates its Plugin Framework data sources
- **THEN** every data source SHALL be constructed via `entitycore.NewKibanaDataSource[T]` or `entitycore.NewElasticsearchDataSource[T]`
- **AND** no data source SHALL embed `*entitycore.DataSourceBase`

#### Scenario: DataSourceBase is not present in the entitycore package
- **WHEN** developers inspect `internal/entitycore` for data source base types
- **THEN** `DataSourceBase` SHALL NOT exist
- **AND** only envelope constructors and connection field embeddable structs SHALL be available

### Requirement: Kibana data source models embed KibanaConnectionField
Kibana-backed envelope data source models SHALL embed `entitycore.KibanaConnectionField` (or an equivalent struct field with the `GetKibanaConnection` accessor) so the envelope can decode the `kibana_connection` block alongside entity attributes.

#### Scenario: Model embedding satisfies KibanaDataSourceModel
- **WHEN** a data source model embeds `KibanaConnectionField`
- **THEN** the model SHALL satisfy the `KibanaDataSourceModel` type constraint
- **AND** the envelope SHALL decode the `kibana_connection` block into that field during `Read`

### Requirement: Elasticsearch data source models embed ElasticsearchConnectionField
Elasticsearch-backed envelope data source models SHALL embed `entitycore.ElasticsearchConnectionField` (or an equivalent struct field with the `GetElasticsearchConnection` accessor) so the envelope can decode the `elasticsearch_connection` block alongside entity attributes.

#### Scenario: Model embedding satisfies ElasticsearchDataSourceModel
- **WHEN** a data source model embeds `ElasticsearchConnectionField`
- **THEN** the model SHALL satisfy the `ElasticsearchDataSourceModel` type constraint
- **AND** the envelope SHALL decode the `elasticsearch_connection` block into that field during `Read`

