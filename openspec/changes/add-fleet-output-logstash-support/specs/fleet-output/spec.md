## ADDED Requirements

### Requirement: Resource supports Logstash output type end-to-end
The `elasticstack_fleet_output` resource SHALL support `type = "logstash"` for create, read, update, and delete operations using the Fleet Outputs API. When `type` is `logstash`, the resource SHALL build a valid Logstash output payload from Terraform plan values and SHALL persist returned API values to Terraform state using the same common field mapping used by other output types.

#### Scenario: Create Logstash output
- **WHEN** a configuration sets `type = "logstash"` with valid common fields (including `name` and `hosts`) and create runs
- **THEN** the provider SHALL call the Fleet create output API with a Logstash output payload and store the returned output in state without diagnostics errors

#### Scenario: Update Logstash output
- **WHEN** an existing `type = "logstash"` resource changes mutable fields and update runs
- **THEN** the provider SHALL call the Fleet update output API for the same output identifier and persist updated values to state

#### Scenario: Read Logstash output state mapping
- **WHEN** read receives a Fleet response for an output with `type = "logstash"`
- **THEN** the provider SHALL map common output fields into state and SHALL NOT raise an unknown output type diagnostic

### Requirement: Acceptance coverage for Logstash output lifecycle
The Fleet output acceptance test suite SHALL include coverage for `type = "logstash"` lifecycle behavior to verify create/read/update/import compatibility and protect against regressions.

#### Scenario: Acceptance test for Logstash lifecycle
- **WHEN** provider acceptance tests are executed for Fleet output resource
- **THEN** at least one test case SHALL provision a `type = "logstash"` output and validate state after create and update

#### Scenario: Acceptance test for Logstash import
- **WHEN** a pre-existing Logstash output is imported into Terraform state
- **THEN** the import/read cycle SHALL populate `output_id` and common fields consistently with managed resources
