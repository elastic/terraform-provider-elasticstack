## ADDED Requirements

### Requirement: Inference endpoint resource uses the entitycore envelope
The `elasticstack_elasticsearch_inference_endpoint` resource SHALL embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a Plugin Framework resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered as a PF resource

### Requirement: Create callback puts the inference endpoint
The `createFunc` callback SHALL PUT the inference endpoint configuration and compute the composite id.

#### Scenario: Create endpoint
- **GIVEN** a valid planned model
- **WHEN** the create callback runs
- **THEN** it SHALL PUT the endpoint configuration
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback re-puts the endpoint
The `updateFunc` callback SHALL re-PUT the endpoint configuration.

#### Scenario: Update endpoint
- **GIVEN** an existing endpoint with changes
- **WHEN** the update callback runs
- **THEN** it SHALL re-PUT the configuration

### Requirement: Read callback fetches the endpoint
The `readFunc` callback SHALL GET the endpoint configuration.

#### Scenario: Successful read
- **GIVEN** an endpoint exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true`

#### Scenario: Missing endpoint removes from state
- **GIVEN** the endpoint does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback removes the endpoint
The `deleteFunc` callback SHALL DELETE the endpoint.

#### Scenario: Delete endpoint
- **GIVEN** an endpoint exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call the delete API
