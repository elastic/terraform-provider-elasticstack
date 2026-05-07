# ml-datafeed-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase4-ml-jobs. Update Purpose after archive.
## Requirements
### Requirement: ML datafeed resource uses the entitycore envelope
The `elasticstack_elasticsearch_ml_datafeed` resource SHALL embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: Create callback creates the datafeed
The `createFunc` callback SHALL build the datafeed config and call the Create Datafeed API.

#### Scenario: Create datafeed
- **GIVEN** a valid planned model
- **WHEN** the create callback runs
- **THEN** it SHALL create the datafeed
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback updates the datafeed
The `updateFunc` callback SHALL update the datafeed configuration.

#### Scenario: Update datafeed
- **GIVEN** an existing datafeed with changed query
- **WHEN** the update callback runs
- **THEN** it SHALL update the datafeed

### Requirement: Read callback fetches the datafeed
The `readFunc` callback SHALL GET the datafeed config and populate the model.

#### Scenario: Successful read
- **GIVEN** a datafeed exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true`

#### Scenario: Missing datafeed removes from state
- **GIVEN** the datafeed does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback deletes the datafeed
The `deleteFunc` callback SHALL stop and delete the datafeed.

#### Scenario: Delete datafeed
- **GIVEN** a datafeed exists
- **WHEN** the delete callback runs
- **THEN** it SHALL stop and delete the datafeed

### Requirement: Custom ImportState parses composite id
The concrete resource SHALL implement `ImportState` that parses a composite id and sets `datafeed_id`.

#### Scenario: Import datafeed
- **GIVEN** a composite import id `<cluster_uuid>/<datafeed_id>`
- **WHEN** import runs
- **THEN** the `id` SHALL be stored
- **AND** `datafeed_id` SHALL be set

