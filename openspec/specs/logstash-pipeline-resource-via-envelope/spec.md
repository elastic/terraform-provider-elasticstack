# logstash-pipeline-resource-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase2-logstash. Update Purpose after archive.
## Requirements
### Requirement: Logstash pipeline resource uses the entitycore envelope
The `elasticstack_elasticsearch_logstash_pipeline` resource SHALL be implemented on PF and embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: Create callback puts the pipeline
The `createFunc` callback SHALL build the pipeline body and settings map, PUT, and compute the composite id.

#### Scenario: Create pipeline with settings
- **GIVEN** a valid planned model with pipeline body and settings
- **WHEN** the create callback runs
- **THEN** it SHALL PUT the pipeline with settings
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback re-puts the pipeline
The `updateFunc` callback SHALL re-PUT the pipeline configuration.

#### Scenario: Update pipeline
- **GIVEN** an existing pipeline with changed `pipeline_batch_size`
- **WHEN** the update callback runs
- **THEN** it SHALL re-PUT the pipeline

### Requirement: Read callback fetches the pipeline
The `readFunc` callback SHALL GET the pipeline and map the flat settings API response to typed model fields.

#### Scenario: Successful read
- **GIVEN** a pipeline exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true` with all fields populated

#### Scenario: Missing pipeline removes from state
- **GIVEN** the pipeline does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback removes the pipeline
The `deleteFunc` callback SHALL DELETE the pipeline.

#### Scenario: Delete pipeline
- **GIVEN** a pipeline exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call the delete API

### Requirement: Import behavior is preserved
The concrete resource SHALL implement `ImportState` as passthrough.

#### Scenario: Import pipeline
- **GIVEN** a composite import id
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state

