## ADDED Requirements

### Requirement: Ingest pipeline resource uses the entitycore envelope
The `elasticstack_elasticsearch_ingest_pipeline` resource SHALL be implemented on the Terraform Plugin Framework and SHALL embed `*entitycore.ElasticsearchResource[Data]`. The resource SHALL satisfy `resource.Resource`, `resource.ResourceWithConfigure`, and `resource.ResourceWithImportState`.

#### Scenario: Resource is registered as a Plugin Framework resource
- **WHEN** the provider initializes
- **THEN** `elasticstack_elasticsearch_ingest_pipeline` SHALL be registered as a Plugin Framework resource
- **AND** it SHALL be constructed via `entitycore.NewElasticsearchResource[Data]`

#### Scenario: Schema includes injected connection block
- **WHEN** the schema factory returns a schema without `elasticsearch_connection`
- **THEN** the envelope's `Schema` method SHALL inject the connection block
- **AND** all other attributes SHALL match the legacy SDK schema shape

### Requirement: Model satisfies ElasticsearchResourceModel
The resource's `Data` struct SHALL implement `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List` as value-receiver methods.

#### Scenario: Model getters return correct fields
- **WHEN** a `Data` value is created with `ID`, `Name`, and `ElasticsearchConnection` populated
- **THEN** `GetID()` SHALL return the `ID` field
- **AND** `GetResourceID()` SHALL return the `Name` field
- **AND** `GetElasticsearchConnection()` SHALL return the `ElasticsearchConnection` field

### Requirement: Create callback puts the ingest pipeline
The `createFunc` callback SHALL decode the planned model, construct a pipeline body from `processors` and `on_failure` JSON strings, PUT it to `/_ingest/pipeline/{name}`, compute the composite `id`, and return the written model.

#### Scenario: Create puts new pipeline
- **GIVEN** a valid planned `Data` with name, description, processors, and on_failure
- **WHEN** the create callback runs
- **THEN** it SHALL PUT the pipeline to Elasticsearch
- **AND** set `ID` on the returned model
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback re-puts the ingest pipeline
The `updateFunc` callback SHALL behave identically to `createFunc` for this resource.

#### Scenario: Update replaces pipeline
- **GIVEN** an existing pipeline with changed processors
- **WHEN** the update callback runs
- **THEN** it SHALL PUT the updated pipeline
- **AND** the envelope SHALL read back and persist state

### Requirement: Read callback fetches pipeline configuration
The `readFunc` callback SHALL GET `/_ingest/pipeline/{name}` and populate the model fields from the response.

#### Scenario: Successful read returns populated model
- **GIVEN** a pipeline exists in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return a populated `Data` with all fields set from the API response

#### Scenario: Missing pipeline removes from state
- **GIVEN** the pipeline does not exist in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`
- **AND** the envelope SHALL remove the resource from state

### Requirement: Delete callback removes the pipeline
The `deleteFunc` callback SHALL DELETE `/_ingest/pipeline/{name}`.

#### Scenario: Successful delete
- **GIVEN** a pipeline exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call the Delete Ingest Pipeline API
- **AND** return nil diagnostics on success

### Requirement: Import behavior is preserved
The concrete resource type SHALL implement `ImportState` as a passthrough on the `id` attribute.

#### Scenario: Import with composite id
- **GIVEN** an import identifier in the format `<cluster_uuid>/<pipeline_name>`
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state for subsequent read and delete operations

### Requirement: JSON fields are validated and normalized
`processors`, `on_failure`, and `metadata` SHALL be validated as JSON during plan validation and normalized.

#### Scenario: Invalid JSON rejected at plan time
- **GIVEN** `processors` contains an invalid JSON string
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation error
