## MODIFIED Requirements

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

### Requirement: Kibana data source models embed KibanaConnectionField and expose identity accessors
Kibana-backed envelope data source models SHALL embed `entitycore.KibanaConnectionField` (or an equivalent struct field with the `GetKibanaConnection` accessor) and SHALL provide value-receiver methods `GetID() types.String`, `GetResourceID() types.String`, and `GetSpaceID() types.String` so the envelope can decode the `kibana_connection` block and resolve read identity.

#### Scenario: Model embedding satisfies KibanaDataSourceModel
- **WHEN** a data source model embeds `KibanaConnectionField` and declares `GetID`, `GetResourceID`, and `GetSpaceID`
- **THEN** the model SHALL satisfy the `KibanaDataSourceModel` type constraint
- **AND** the envelope SHALL decode the `kibana_connection` block into that field during `Read`

### Requirement: Elasticsearch data source models embed ElasticsearchConnectionField and expose identity accessors
Elasticsearch-backed envelope data source models SHALL embed `entitycore.ElasticsearchConnectionField` (or an equivalent struct field with the `GetElasticsearchConnection` accessor) and SHALL provide value-receiver methods `GetID() types.String` and `GetResourceID() types.String` so the envelope can decode the `elasticsearch_connection` block and resolve read identity.

#### Scenario: Model embedding satisfies ElasticsearchDataSourceModel
- **WHEN** a data source model embeds `ElasticsearchConnectionField` and declares `GetID` and `GetResourceID`
- **THEN** the model SHALL satisfy the `ElasticsearchDataSourceModel` type constraint
- **AND** the envelope SHALL decode the `elasticsearch_connection` block into that field during `Read`

## ADDED Requirements

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

### Requirement: Envelope assigns the composite id
The system SHALL compute and assign the composite `id` of the resolved entity from the scoped client and resolved identity, so concrete read functions do not compute or assign `id`.

#### Scenario: Composite id assigned after successful read
- **WHEN** the concrete read function returns `found == true` without error diagnostics
- **THEN** the envelope SHALL assign the model's `id` from the scoped client and resolved identity before setting state

### Requirement: Data source constructors accept an options struct with optional PostRead
The system SHALL accept data source configuration via `ElasticsearchDataSourceOptions[T]` and `KibanaDataSourceOptions[T]` structs carrying `Schema` and `Read`, and an optional `PostRead` hook. When `PostRead` is non-nil it SHALL run after state is set on a found read, mirroring the resource envelope `PostRead` semantics.

#### Scenario: PostRead runs after found read
- **GIVEN** a data source constructed with a non-nil `PostRead`
- **WHEN** the concrete read function returns `found == true` and state is set
- **THEN** the envelope SHALL invoke `PostRead` with the persisted model

#### Scenario: PostRead omitted
- **GIVEN** a data source constructed without a `PostRead`
- **WHEN** the concrete read function returns `found == true` and state is set
- **THEN** the envelope SHALL complete `Read` without invoking any post-read hook
