## ADDED Requirements

### Requirement: ML datafeed state resource uses the entitycore envelope for Schema and Read
The `elasticstack_elasticsearch_ml_datafeed_state` resource SHALL embed `*entitycore.ElasticsearchResource[Data]`. The envelope SHALL own Schema and Read.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: Create override transitions the datafeed state
The concrete resource SHALL override `Create` to start or stop the datafeed.

#### Scenario: Start a datafeed
- **GIVEN** a planned model with `state = "started"`
- **WHEN** create runs
- **THEN** it SHALL start the datafeed

#### Scenario: Stop a datafeed
- **GIVEN** a planned model with `state = "stopped"`
- **WHEN** create runs
- **THEN** it SHALL stop the datafeed

### Requirement: Update override transitions the datafeed state
The concrete resource SHALL override `Update` to transition the datafeed when the state changes.

#### Scenario: Change datafeed state
- **GIVEN** an existing state with `state = "started"`
- **WHEN** the plan changes to `state = "stopped"`
- **THEN** update SHALL stop the datafeed

### Requirement: Delete override stops the datafeed if running
The concrete resource SHALL override `Delete` to stop the datafeed if it is started before removing from state.

#### Scenario: Delete while started
- **GIVEN** an existing state with `state = "started"`
- **WHEN** delete runs
- **THEN** it SHALL stop the datafeed
- **AND** remove the resource from state

#### Scenario: Delete while stopped
- **GIVEN** an existing state with `state = "stopped"`
- **WHEN** delete runs
- **THEN** it SHALL remove the resource from state without calling stop

### Requirement: Read callback returns current datafeed state
The `readFunc` callback SHALL GET datafeed stats and return the current state.

#### Scenario: Read current state
- **GIVEN** a datafeed exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true` with the current state

#### Scenario: Missing datafeed removes from state
- **GIVEN** the datafeed does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Timeouts are preserved
Create and Update overrides SHALL use Terraform framework timeouts.

#### Scenario: Timeout on state transition
- **GIVEN** a state transition that exceeds the configured timeout
- **WHEN** create or update runs
- **THEN** it SHALL return a timeout error
