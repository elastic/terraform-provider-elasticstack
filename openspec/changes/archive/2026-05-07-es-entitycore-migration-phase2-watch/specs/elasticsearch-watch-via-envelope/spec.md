## ADDED Requirements

### Requirement: Watch resource uses the entitycore envelope
The `elasticstack_elasticsearch_watch` resource SHALL embed `*entitycore.ElasticsearchResource[Data]`. The resource SHALL satisfy `resource.Resource`, `resource.ResourceWithConfigure`, and `resource.ResourceWithImportState`.

#### Scenario: Resource is registered as an envelope resource
- **WHEN** the provider initializes
- **THEN** `elasticstack_elasticsearch_watch` SHALL be constructed via `entitycore.NewElasticsearchResource[Data]` with real create, update, read, and delete callbacks
- **AND** the concrete type SHALL declare `ImportState`

#### Scenario: Schema includes injected connection block
- **WHEN** the schema factory returns a schema without `elasticsearch_connection`
- **THEN** the envelope's `Schema` method SHALL inject the connection block
- **AND** all attributes and defaults SHALL match the pre-migration schema exactly

### Requirement: Model satisfies ElasticsearchResourceModel
The resource's `Data` struct SHALL implement `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List`.

#### Scenario: Model getters return correct fields
- **WHEN** a `Data` value is created with `ID`, `WatchID`, and `ElasticsearchConnection` populated
- **THEN** `GetID()` SHALL return the `ID` field
- **AND** `GetResourceID()` SHALL return the `WatchID` field
- **AND** `GetElasticsearchConnection()` SHALL return the `ElasticsearchConnection` field

### Requirement: Read callback fetches and maps watches
The envelope's read callback SHALL call the Get Watch API, map the response into `Data` via `fromAPIModel`, and return `(model, true, nil)` when the watch exists. It SHALL return `(_, false, nil)` when the watch is not found (404).

#### Scenario: Successful read returns populated model
- **GIVEN** a watch exists in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return a populated `Data` with all fields set from the API response
- **AND** the returned model SHALL preserve prior concrete action secrets at redacted paths

#### Scenario: Missing watch removes from state
- **GIVEN** the watch does not exist in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`
- **AND** the envelope SHALL remove the resource from state

### Requirement: Create and update callbacks use Put Watch API
The `createFunc` and `updateFunc` callbacks SHALL build a `models.PutWatch` from the planned model, call the Put Watch API, compute the composite `id`, and return the model with `ID` set. The envelope SHALL invoke `readFunc` after a successful callback and persist the read result.

#### Scenario: Create puts watch and refreshes state
- **GIVEN** a valid planned `Data` model
- **WHEN** the create callback runs
- **THEN** it SHALL call Put Watch API with the `active` flag as a query parameter
- **AND** the envelope SHALL read back the watch and set state from the read result

#### Scenario: Update puts watch and refreshes state
- **GIVEN** an existing watch and a changed planned model
- **WHEN** the update callback runs
- **THEN** it SHALL call Put Watch API
- **AND** the envelope SHALL read back the watch and set state from the read result

### Requirement: Transform behavior differs between create and update
The create callback SHALL omit `transform` from the Put Watch request body when it is not configured. The update callback SHALL include `transform` with an empty JSON object `{}` when it is not configured, so Elasticsearch clears any existing transform.

#### Scenario: Transform omitted on create
- **GIVEN** `transform` is not configured
- **WHEN** create builds the request body
- **THEN** the `transform` field SHALL be omitted from the Put Watch JSON body

#### Scenario: Transform cleared on update
- **GIVEN** `transform` is not configured
- **WHEN** update builds the request body for an existing watch
- **THEN** the Put Watch JSON body SHALL include `transform` with an empty JSON object

### Requirement: Delete callback removes watches
The envelope's delete callback SHALL call the Delete Watch API with the parsed resource identifier.

#### Scenario: Successful delete
- **GIVEN** a watch exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call Delete Watch API
- **AND** return nil diagnostics on success

### Requirement: Actions redaction is preserved
On read, when the Get Watch API returns the redacted string sentinel at a nested path and the prior known Terraform `actions` JSON value has a concrete value of any JSON type at the same path, the resource SHALL preserve that prior concrete value.

#### Scenario: Redacted action secret preserved from prior state
- **GIVEN** the last-applied `actions` JSON in Terraform state includes a concrete nested secret value at a path
- **WHEN** read runs and the API returns `::es_redacted::` for that path
- **THEN** the final `actions` value in state SHALL keep the prior concrete secret for that path

#### Scenario: Redacted action header preserved when prior is an object
- **GIVEN** the prior `actions` JSON has an object at a nested path
- **WHEN** read runs and the API returns `::es_redacted::` for that path
- **THEN** the final `actions` value in state SHALL keep the prior object at that path

### Requirement: Import behavior is preserved
The concrete resource type SHALL implement `ImportState` as a passthrough on the `id` attribute.

#### Scenario: Import with composite id
- **GIVEN** an import identifier in the format `<cluster_uuid>/<watch_id>`
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state for subsequent read and delete operations
