## ADDED Requirements

### Requirement: Data stream resource uses the entitycore envelope
The `elasticstack_elasticsearch_data_stream` resource SHALL be implemented on the Terraform Plugin Framework and SHALL embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a Plugin Framework resource
- **WHEN** the provider initializes
- **THEN** `elasticstack_elasticsearch_data_stream` SHALL be registered as a Plugin Framework resource
- **AND** it SHALL be constructed via `entitycore.NewElasticsearchResource[Data]`

#### Scenario: Schema includes injected connection block
- **WHEN** the schema factory returns a schema without `elasticsearch_connection`
- **THEN** the envelope's `Schema` method SHALL inject the connection block

### Requirement: Model satisfies ElasticsearchResourceModel
The resource's `Data` struct SHALL implement `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List` as value-receiver methods.

#### Scenario: Model getters return correct fields
- **WHEN** a `Data` value is created with `ID`, `Name`, and `ElasticsearchConnection` populated
- **THEN** `GetID()` SHALL return the `ID` field
- **AND** `GetResourceID()` SHALL return the `Name` field
- **AND** `GetElasticsearchConnection()` SHALL return the `ElasticsearchConnection` field

### Requirement: Create callback puts the data stream
The `createFunc` callback SHALL PUT `/_data_stream/{name}` and compute the composite `id`.

#### Scenario: Create new data stream
- **GIVEN** a valid planned `Data` with name
- **WHEN** the create callback runs
- **THEN** it SHALL PUT the data stream
- **AND** set `ID` on the returned model
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback behaves as create
Since `name` is ForceNew, the `updateFunc` callback SHALL behave identically to `createFunc`; Terraform will Destroy+Create on name change.

#### Scenario: Update is full replace
- **GIVEN** an existing data stream with unchanged name
- **WHEN** the update callback runs
- **THEN** it SHALL re-PUT the data stream
- **AND** the envelope SHALL read back and persist state

### Requirement: Read callback fetches data stream info
The `readFunc` callback SHALL GET `/_data_stream/{name}` and populate all computed fields.

#### Scenario: Successful read returns populated model
- **GIVEN** a data stream exists in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return a populated `Data` with `timestamp_field`, `indices`, `generation`, etc.

#### Scenario: Missing data stream removes from state
- **GIVEN** the data stream does not exist in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`
- **AND** the envelope SHALL remove the resource from state

### Requirement: Delete callback removes the data stream
The `deleteFunc` callback SHALL DELETE `/_data_stream/{name}`.

#### Scenario: Successful delete
- **GIVEN** a data stream exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call the Delete Data Stream API
- **AND** return nil diagnostics on success

### Requirement: Import behavior is preserved
The concrete resource type SHALL implement `ImportState` as a passthrough on the `id` attribute.

#### Scenario: Import with composite id
- **GIVEN** an import identifier in the format `<cluster_uuid>/<data_stream_name>`
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state for subsequent read and delete operations
