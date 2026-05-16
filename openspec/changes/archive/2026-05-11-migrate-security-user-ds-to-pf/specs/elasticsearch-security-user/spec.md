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

- **WHEN** the entity-specific read callback is invoked with the scoped client and config
- **THEN** it SHALL call `elasticsearch.GetUser`
- **AND** when the user is found, set `model.ID` to `<cluster_uuid>/<username>`
- **AND** map the API response into the model attributes

## MODIFIED Requirements

### Requirement: Read API (REQ-DS-001)

The data source SHALL read a single Elasticsearch security user by `username` using the Get User API. When the API returns a non-success response, the data source SHALL surface the error to Terraform diagnostics. The read logic SHALL live inside the `readDataSource` callback passed to `entitycore.NewElasticsearchDataSource`.

#### Scenario: User found

- **GIVEN** a user exists in Elasticsearch with the requested `username`
- **WHEN** the data source is read
- **THEN** all computed attributes SHALL be populated from the API response

### Requirement: Identity (REQ-DS-002)

The data source SHALL expose a computed `id` attribute in the format `<cluster_uuid>/<username>`, derived by calling `client.ID(ctx, username)` after resolving the scoped client.

#### Scenario: Computed id set

- **GIVEN** the data source reads an existing user
- **WHEN** read completes
- **THEN** `id` SHALL equal `<cluster_uuid>/<username>`

### Requirement: User not found (REQ-DS-003)

When the Elasticsearch Get users API returns nil without an error, the data source SHALL preserve SDK behavior by setting `id` to an empty string and returning no warning or error diagnostic.

#### Scenario: User does not exist

- **GIVEN** a user does not exist with the requested `username`
- **WHEN** the data source is read
- **THEN** `id` SHALL be set to an empty string
- **AND** no diagnostic SHALL be returned

### Requirement: State mapping (REQ-DS-004–REQ-DS-005)

The data source SHALL map the Get User API response into the following computed attributes: `full_name`, `email`, `roles` (set of strings), `metadata` (as normalized JSON string), and `enabled`. When `email` or `full_name` are absent in the response, they SHALL be set to empty strings rather than null.

#### Scenario: All attributes mapped

- **GIVEN** a successful API response with all user fields present
- **WHEN** read completes
- **THEN** every computed attribute SHALL reflect the corresponding API value

#### Scenario: Null name fields default to empty string

- **GIVEN** a user with no `full_name` or `email`
- **WHEN** read completes
- **THEN** `full_name` and `email` SHALL be `""`

### Requirement: Connection (REQ-DS-006–REQ-DS-007)

The data source SHALL use the provider's configured Elasticsearch client by default. When the `elasticsearch_connection` block is configured, the data source SHALL use that connection. Connection resolution SHALL be owned by the `entitycore.NewElasticsearchDataSource` envelope.

#### Scenario: Data source-scoped connection

- **GIVEN** `elasticsearch_connection` is set
- **WHEN** the data source reads the user
- **THEN** the scoped client SHALL be built from that block
