# entitycore-datasource-envelope Specification

## Purpose
TBD - created by archiving change envelope-datasource-generics. Update Purpose after archive.
## Requirements
### Requirement: Envelope constructor produces a valid DataSource
The system SHALL provide a generic constructor `NewKibanaDataSource[T]()` (and `NewElasticsearchDataSource[T]()`) that accepts an options struct and returns a value satisfying `datasource.DataSource`.

#### Scenario: Constructor returns valid data source
- **WHEN** `NewKibanaDataSource[T](component, name, KibanaDataSourceOptions[T]{Schema, Read})` is called
- **THEN** the returned value SHALL satisfy `datasource.DataSource`
- **AND** the returned value SHALL satisfy `datasource.DataSourceWithConfigure`

#### Scenario: Elasticsearch constructor returns valid data source
- **WHEN** `NewElasticsearchDataSource[T](component, name, ElasticsearchDataSourceOptions[T]{Schema, Read})` is called
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
The system SHALL resolve the read identity from the decoded model and invoke the concrete read function with the scoped client, the resolved identity, and the deserialized model, then capture the returned model, `found` boolean, and diagnostics. For Kibana the resolved identity SHALL include both `resourceID` and `spaceID`; for Elasticsearch it SHALL include `resourceID`.

#### Scenario: Elasticsearch read function receives resolved identity and model
- **WHEN** `Read` has successfully deserialized config and resolved the client
- **THEN** the concrete read function SHALL be called with `(context, *ElasticsearchScopedClient, resourceID string, T)`
- **AND** its returned `(T, bool, diag.Diagnostics)` SHALL be used for state setting and not-found handling

#### Scenario: Kibana read function receives resolved identity and model
- **WHEN** `Read` has successfully deserialized config and resolved the client
- **THEN** the concrete read function SHALL be called with `(context, *KibanaScopedClient, resourceID string, spaceID string, T)`
- **AND** its returned `(T, bool, diag.Diagnostics)` SHALL be used for state setting and not-found handling

### Requirement: Envelope owns state persistence
The system SHALL set the Terraform state from the model returned by the concrete read function only when the read function reports `found == true`. Because `T` embeds the connection block field, the returned model naturally preserves the original connection block value.

#### Scenario: State set after successful read
- **WHEN** the concrete read function returns `found == true` without error diagnostics
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
Kibana-backed envelope data source models SHALL embed `entitycore.KibanaConnectionField` (or an equivalent struct field with the `GetKibanaConnection` accessor) and SHALL provide value-receiver methods `GetID() types.String`, `GetResourceID() types.String`, and `GetSpaceID() types.String` so the envelope can decode the `kibana_connection` block and resolve read identity.

#### Scenario: Model embedding satisfies KibanaDataSourceModel
- **WHEN** a data source model embeds `KibanaConnectionField` and declares `GetID`, `GetResourceID`, and `GetSpaceID`
- **THEN** the model SHALL satisfy the `KibanaDataSourceModel` type constraint
- **AND** the envelope SHALL decode the `kibana_connection` block into that field during `Read`

### Requirement: Elasticsearch data source models embed ElasticsearchConnectionField
Elasticsearch-backed envelope data source models SHALL embed `entitycore.ElasticsearchConnectionField` (or an equivalent struct field with the `GetElasticsearchConnection` accessor) and SHALL provide value-receiver methods `GetID() types.String` and `GetResourceID() types.String` so the envelope can decode the `elasticsearch_connection` block and resolve read identity.

#### Scenario: Model embedding satisfies ElasticsearchDataSourceModel
- **WHEN** a data source model embeds `ElasticsearchConnectionField` and declares `GetID` and `GetResourceID`
- **THEN** the model SHALL satisfy the `ElasticsearchDataSourceModel` type constraint
- **AND** the envelope SHALL decode the `elasticsearch_connection` block into that field during `Read`

### Requirement: Envelope resolves read identity centrally
The system SHALL resolve the read identity from the decoded config model before invoking the concrete read function, using the same composite-ID-or-fallback rules as the resource envelope. For Elasticsearch the envelope SHALL resolve a `resourceID`; for Kibana the envelope SHALL resolve a `resourceID` and `spaceID`, honoring the `KibanaUnscopedSpace` opt-out for space validation.

#### Scenario: Elasticsearch identity resolved from model
- **WHEN** `Read` decodes a config model whose identity is expressed via a composite `id` or a resource identifier accessor
- **THEN** the envelope SHALL resolve a non-empty `resourceID` and pass it to the read function
- **AND** when no identity can be resolved the envelope SHALL add an "Invalid resource identifier" error diagnostic and SHALL NOT invoke the read function

#### Scenario: Kibana identity and space resolved from model
- **WHEN** `Read` decodes a config model with a composite `id` of the form `<space>/<resource>` or with `GetResourceID`/`GetSpaceID` values
- **THEN** the envelope SHALL resolve both `resourceID` and `spaceID` and pass them to the read function
- **AND** for a model that opts out of space scoping via `KibanaUnscopedSpace`, an empty `spaceID` SHALL be permitted

### Requirement: Envelope applies a centralized not-found policy
The system SHALL apply a single not-found policy when the concrete read function reports `found == false`: it SHALL append a standardized "not found" error diagnostic that identifies the component, data source name, and resolved identity, and it SHALL NOT set Terraform state.

#### Scenario: Read function reports entity not found
- **WHEN** the concrete read function returns `found == false` without error diagnostics
- **THEN** the envelope SHALL append a standardized not-found error diagnostic
- **AND** `resp.State.Set` SHALL NOT be called

### Requirement: Read function owns id assignment
Mirroring the resource envelope, the concrete read function SHALL compute and assign the model's `id` on the returned model `T`; the envelope SHALL NOT mutate `id`. The model constraint exposes `GetID()` for read-identity resolution only and intentionally provides no identity mutator (`SetID`), so a value-typed generic `T` cannot be assigned by the envelope. Standard entities assign the composite `id` via `client.ID(ctx, resourceID)`; entities whose identity is non-standard assign their own `id` in the read function with no envelope opt-out.

#### Scenario: Read function assigns composite id for a standard entity
- **WHEN** the concrete read function resolves a standard entity for a non-empty `resourceID`
- **THEN** the read function SHALL set the model's `id` via `client.ID(ctx, resourceID)` before returning `found == true`
- **AND** the envelope SHALL persist the returned `id` without modifying it

#### Scenario: Read function assigns a non-standard id
- **WHEN** the concrete read function resolves an entity whose `id` is not `client.ID(ctx, resourceID)` (for example `internal/elasticsearch/cluster/info`, where `id` derives from `cluster_uuid`, or `internal/elasticsearch/index/indices`, where `id` is the target pattern)
- **THEN** the read function SHALL set the model's `id` to the entity-specific value before returning `found == true`
- **AND** no envelope opt-out mechanism SHALL be required

### Requirement: Data source constructors accept an options struct with optional PostRead
The system SHALL accept data source configuration via `ElasticsearchDataSourceOptions[T]` and `KibanaDataSourceOptions[T]` structs carrying `Schema` and `Read`, and an optional `PostRead` hook. The data source `PostRead` signatures SHALL be:

- Elasticsearch: `func(ctx context.Context, client *clients.ElasticsearchScopedClient, model T) diag.Diagnostics`
- Kibana: `func(ctx context.Context, client *clients.KibanaScopedClient, model T) diag.Diagnostics`

These SHALL intentionally omit the resource `PostReadFunc`'s trailing `privateState any` argument, because data sources have no private state (the framework `datasource.ReadResponse` exposes no `Private` field). When `PostRead` is non-nil it SHALL run after state is set on a found read, mirroring the resource envelope `PostRead` ordering.

#### Scenario: PostRead runs after found read
- **GIVEN** a data source constructed with a non-nil `PostRead`
- **WHEN** the concrete read function returns `found == true` and state is set
- **THEN** the envelope SHALL invoke `PostRead` with the scoped client and the persisted model
- **AND** the envelope SHALL NOT pass any private-state argument

#### Scenario: PostRead omitted
- **GIVEN** a data source constructed without a `PostRead`
- **WHEN** the concrete read function returns `found == true` and state is set
- **THEN** the envelope SHALL complete `Read` without invoking any post-read hook

