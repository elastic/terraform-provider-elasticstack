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
- **THEN** it SHALL call `elasticsearch.GetRole`
- **AND** when the role is found, set `model.ID` to `<cluster_uuid>/<role_name>`
- **AND** map the API response into the model's nested attributes

## MODIFIED Requirements

### Requirement: Data source read API (DS-REQ-001)

The data source SHALL read a single Elasticsearch security role by `name` using the Get Role API. The read logic SHALL live inside the `readDataSource` callback passed to `entitycore.NewElasticsearchDataSource`.

#### Scenario: Role found

- **GIVEN** a role exists in Elasticsearch with the requested `name`
- **WHEN** the data source is read
- **THEN** all computed attributes SHALL be populated from the API response

### Requirement: Data source identity (DS-REQ-003)

The data source SHALL expose a computed `id` attribute in the format `<cluster_uuid>/<role_name>`, derived by calling `client.ID(ctx, roleName)` after resolving the scoped client.

#### Scenario: Computed id set

- **GIVEN** the data source reads an existing role
- **WHEN** read completes
- **THEN** `id` SHALL equal `<cluster_uuid>/<role_name>`

### Requirement: Data source not found behavior (DS-REQ-004)

When a role is not found, the data source SHALL preserve SDK behavior by setting `id` to an empty string and returning no warning or error diagnostic.

#### Scenario: Role not found

- **GIVEN** a role does not exist with the requested `name`
- **WHEN** the data source is read
- **THEN** `id` SHALL be set to an empty string
- **AND** no diagnostic SHALL be returned

### Requirement: Data source connection (DS-REQ-005–DS-REQ-006)

The data source SHALL use the provider's configured Elasticsearch client by default. When the `elasticsearch_connection` block is configured, the data source SHALL use that connection. Connection resolution SHALL be owned by the `entitycore.NewElasticsearchDataSource` envelope.

#### Scenario: Data source-scoped connection

- **GIVEN** `elasticsearch_connection` is set
- **WHEN** the data source reads the role
- **THEN** the scoped client SHALL be built from that block

### Requirement: Data source attribute mapping (DS-REQ-007)

The data source SHALL map the Get Role API response into the following computed attributes: `description`, `cluster`, `run_as`, `global` (as normalized JSON string), `metadata` (as normalized JSON string), `applications` (set of objects), `indices` (set of objects with nested `field_security` list), and `remote_indices` (set of objects with nested `field_security` list). `cluster` privileges SHALL be mapped as strings.

#### Scenario: All attributes mapped

- **GIVEN** a successful API response with all role fields present
- **WHEN** read completes
- **THEN** every computed attribute SHALL reflect the corresponding API value
