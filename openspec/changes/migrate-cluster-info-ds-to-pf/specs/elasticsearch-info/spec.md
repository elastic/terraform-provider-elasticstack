## ADDED Requirements

### Requirement: Data source uses Plugin Framework and entitycore envelope

The data source SHALL be implemented as a Plugin Framework `datasource.DataSource` constructed via `entitycore.NewElasticsearchDataSource`. The concrete model SHALL embed `entitycore.ElasticsearchConnectionField` and SHALL satisfy `entitycore.ElasticsearchDataSourceModel`. The envelope SHALL own config decode, scoped client resolution, and state persistence.

#### Scenario: Envelope handles connection and decode

- **WHEN** the data source is evaluated
- **THEN** `entitycore.NewElasticsearchDataSource` SHALL decode the configuration into the concrete model
- **AND** resolve the scoped Elasticsearch client from the model's `elasticsearch_connection` block
- **AND** invoke the entity-specific read callback
- **AND** persist the returned model to state

#### Scenario: Read callback owns API call and id assignment

- **WHEN** the entity-specific read callback is invoked
- **THEN** it SHALL call `elasticsearch.GetClusterInfo` via the scoped client
- **AND** set `model.ID` to the cluster UUID
- **AND** map all response fields into the model

## MODIFIED Requirements

### Requirement: Read API (REQ-001)

The data source SHALL use the Elasticsearch cluster info API (root `GET /`) to retrieve cluster metadata. When the API returns an error, the data source SHALL surface the error to Terraform diagnostics. The read logic SHALL live inside the `readDataSource` callback passed to `entitycore.NewElasticsearchDataSource`.

#### Scenario: API failure

- **GIVEN** the Elasticsearch cluster info API returns an error
- **WHEN** read runs
- **THEN** the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-002)

The data source SHALL set `id` to the cluster's UUID (`cluster_uuid`) returned by the API. The `id` SHALL be set by the read callback inside the model's `ID` field.

#### Scenario: Computed id equals cluster_uuid

- **GIVEN** a successful API response
- **WHEN** read completes
- **THEN** `id` SHALL equal the `cluster_uuid` value from the response

### Requirement: State mapping — top-level fields (REQ-003)

The data source SHALL populate `cluster_uuid`, `cluster_name`, `name`, and `tagline` directly from the API response. If mapping any attribute fails, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: Top-level attributes set

- **GIVEN** a successful API response
- **WHEN** read completes
- **THEN** `cluster_uuid`, `cluster_name`, `name`, and `tagline` SHALL reflect the response values

### Requirement: State mapping — version block (REQ-004)

The data source SHALL populate the `version` list attribute as a single-element list containing all version sub-fields: `build_date` (formatted as a string), `build_flavor`, `build_hash`, `build_snapshot`, `build_type`, `lucene_version`, `minimum_index_compatibility_version`, `minimum_wire_compatibility_version`, and `number`. If mapping the `version` attribute fails, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: Version block populated

- **GIVEN** a successful API response
- **WHEN** read completes
- **THEN** the `version` block SHALL contain exactly one element with all sub-fields populated from the response

### Requirement: Connection (REQ-005–REQ-006)

The data source SHALL use the provider's configured Elasticsearch client by default. When the `elasticsearch_connection` block is configured, the data source SHALL use that connection to construct an Elasticsearch client for its API call. Connection resolution SHALL be performed by the `entitycore.NewElasticsearchDataSource` envelope.

#### Scenario: Data source-scoped connection

- **GIVEN** `elasticsearch_connection` is set
- **WHEN** the API call runs
- **THEN** the client SHALL be built from that block
