# snapshot-lifecycle-resource-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase2-snapshots. Update Purpose after archive.
## Requirements
### Requirement: Snapshot lifecycle resource uses the entitycore envelope
The `elasticstack_elasticsearch_snapshot_lifecycle` resource SHALL be implemented on PF and embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: Create callback puts the SLM policy
The `createFunc` callback SHALL build a `models.SlmPolicy` and PUT to `/_slm/policy/{name}`.

#### Scenario: Create SLM policy
- **GIVEN** a valid planned model with schedule, repository, and config
- **WHEN** the create callback runs
- **THEN** it SHALL PUT the SLM policy
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback re-puts the SLM policy
The `updateFunc` callback SHALL re-PUT the full SLM policy.

#### Scenario: Update SLM policy
- **GIVEN** an existing policy with changed retention
- **WHEN** the update callback runs
- **THEN** it SHALL re-PUT the policy

### Requirement: Read callback fetches the SLM policy
The `readFunc` callback SHALL GET `/_slm/policy/{name}`.

#### Scenario: Successful read
- **GIVEN** a policy exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true` with all fields populated

#### Scenario: Missing policy removes from state
- **GIVEN** the policy does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback removes the SLM policy
The `deleteFunc` callback SHALL DELETE `/_slm/policy/{name}`.

#### Scenario: Delete policy
- **GIVEN** a policy exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call the delete API

### Requirement: Import behavior is preserved
The concrete resource SHALL implement `ImportState` as passthrough.

#### Scenario: Import SLM policy
- **GIVEN** a composite import id
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state

