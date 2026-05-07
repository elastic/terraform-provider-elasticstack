# data-stream-lifecycle-resource-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase1-simple-pf. Update Purpose after archive.
## Requirements
### Requirement: Data stream lifecycle resource uses the entitycore envelope
The `elasticstack_elasticsearch_data_stream_lifecycle` resource SHALL embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a Plugin Framework resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered as a PF resource

### Requirement: Create callback puts lifecycle settings
The `createFunc` callback SHALL PUT data stream lifecycle settings and compute the composite id.

#### Scenario: Create lifecycle
- **GIVEN** a valid planned model
- **WHEN** the create callback runs
- **THEN** it SHALL PUT lifecycle settings
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback re-puts lifecycle settings
The `updateFunc` callback SHALL behave as create for this resource.

#### Scenario: Update lifecycle
- **GIVEN** an existing lifecycle with changes
- **WHEN** the update callback runs
- **THEN** it SHALL re-PUT lifecycle settings

### Requirement: Read callback fetches lifecycle
The `readFunc` callback SHALL GET the lifecycle and return whether it exists.

#### Scenario: Successful read
- **GIVEN** a lifecycle exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true` with populated fields

#### Scenario: Missing lifecycle removes from state
- **GIVEN** the lifecycle does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback resets lifecycle
The `deleteFunc` callback SHALL reset/remove the lifecycle policy.

#### Scenario: Delete lifecycle
- **GIVEN** a lifecycle exists
- **WHEN** the delete callback runs
- **THEN** it SHALL reset the lifecycle settings

