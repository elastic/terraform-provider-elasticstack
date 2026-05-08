## ADDED Requirements

### Requirement: ML job state resource uses the entitycore envelope for Schema and Read
The `elasticstack_elasticsearch_ml_job_state` resource SHALL embed `*entitycore.ElasticsearchResource[Data]`. The envelope SHALL own Schema and Read.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: Create override transitions the job state
The concrete resource SHALL override `Create` to transition the ML job to the desired state (opened or closed).

#### Scenario: Open a job
- **GIVEN** a planned model with `state = "opened"`
- **WHEN** create runs
- **THEN** it SHALL open the job
- **AND** wait for the transition to complete

#### Scenario: Close a job
- **GIVEN** a planned model with `state = "closed"`
- **WHEN** create runs
- **THEN** it SHALL close the job with optional force flag
- **AND** wait for the transition to complete

### Requirement: Update override transitions the job state
The concrete resource SHALL override `Update` to transition the job state when the desired state changes.

#### Scenario: Change job state
- **GIVEN** an existing state with `state = "opened"`
- **WHEN** the plan changes to `state = "closed"`
- **THEN** update SHALL close the job
- **AND** wait for completion

### Requirement: Delete override is a no-op
The concrete resource SHALL override `Delete` to simply remove the resource from Terraform state without affecting the actual job.

#### Scenario: Delete job state resource
- **GIVEN** an existing state
- **WHEN** delete runs
- **THEN** the resource SHALL be removed from state
- **AND** the job SHALL remain in its current state

### Requirement: Read callback returns current job state
The `readFunc` callback SHALL GET the job stats and return the current state.

#### Scenario: Read current state
- **GIVEN** a job exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true` with the current state

#### Scenario: Missing job removes from state
- **GIVEN** the job does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Timeouts are preserved
Create and Update overrides SHALL use Terraform framework timeouts.

#### Scenario: Timeout on state transition
- **GIVEN** a state transition that exceeds the configured timeout
- **WHEN** create or update runs
- **THEN** it SHALL return a timeout error
