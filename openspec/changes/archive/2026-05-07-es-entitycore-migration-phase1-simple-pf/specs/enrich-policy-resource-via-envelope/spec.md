## ADDED Requirements

### Requirement: Enrich policy resource uses the entitycore envelope
The `elasticstack_elasticsearch_enrich_policy` resource SHALL embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a Plugin Framework resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered as a PF resource

### Requirement: Create callback puts and optionally executes the policy
The `createFunc` callback SHALL PUT the policy and call Execute API when `execute = true`.

#### Scenario: Create and execute policy
- **GIVEN** a valid planned model with `execute = true`
- **WHEN** the create callback runs
- **THEN** it SHALL PUT the policy
- **AND** call the Execute API
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback re-puts the policy
The `updateFunc` callback SHALL re-PUT the policy (most fields are ForceNew).

#### Scenario: Update policy
- **GIVEN** an existing policy with changes
- **WHEN** the update callback runs
- **THEN** it SHALL re-PUT the policy
- **AND** optionally execute if configured

### Requirement: Read callback fetches the policy
The `readFunc` callback SHALL GET the policy.

#### Scenario: Successful read
- **GIVEN** a policy exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true` with populated fields

#### Scenario: Missing policy removes from state
- **GIVEN** the policy does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback removes the policy
The `deleteFunc` callback SHALL DELETE the policy.

#### Scenario: Delete policy
- **GIVEN** a policy exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call the delete policy API

### Requirement: Import sets execute to true
The concrete resource's `ImportState` SHALL passthrough the id and set `execute = true` in state.

#### Scenario: Import enrich policy
- **GIVEN** a composite import id
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state
- **AND** `execute` SHALL be set to `true`
