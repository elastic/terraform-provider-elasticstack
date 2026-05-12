# snapshot-repository-resource-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase2-snapshots. Update Purpose after archive.
## Requirements
### Requirement: Snapshot repository resource uses the entitycore envelope
The `elasticstack_elasticsearch_snapshot_repository` resource SHALL be implemented on PF and embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: ValidateConfig ensures exactly one repository type
The concrete resource SHALL implement `ResourceWithValidateConfig` and reject configurations with zero or multiple repository type blocks.

#### Scenario: Multiple types rejected
- **GIVEN** a configuration with both `fs` and `s3` blocks
- **WHEN** validation runs
- **THEN** the provider SHALL return a validation error

### Requirement: Create callback puts the repository
The `createFunc` callback SHALL determine the repository type, build the API model, and PUT.

#### Scenario: Create fs repository
- **GIVEN** a valid planned model with `fs` block
- **WHEN** the create callback runs
- **THEN** it SHALL PUT the repository
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback re-puts the repository
The `updateFunc` callback SHALL re-PUT the repository.

#### Scenario: Update repository settings
- **GIVEN** an existing repository with changed `readonly`
- **WHEN** the update callback runs
- **THEN** it SHALL re-PUT the repository

### Requirement: Read callback fetches the repository
The `readFunc` callback SHALL GET the repository configuration and identify which type block to populate.

#### Scenario: Successful read
- **GIVEN** a repository exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true` with the correct type block populated

#### Scenario: Missing repository removes from state
- **GIVEN** the repository does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback removes the repository
The `deleteFunc` callback SHALL DELETE the repository.

#### Scenario: Delete repository
- **GIVEN** a repository exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call the delete API

### Requirement: Import behavior is preserved
The concrete resource SHALL implement `ImportState` as passthrough.

#### Scenario: Import repository
- **GIVEN** a composite import id
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state

