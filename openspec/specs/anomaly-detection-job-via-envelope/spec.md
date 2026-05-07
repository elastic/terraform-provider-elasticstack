# anomaly-detection-job-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase4-ml-jobs. Update Purpose after archive.
## Requirements
### Requirement: Anomaly detection job resource uses the entitycore envelope
The `elasticstack_elasticsearch_ml_anomaly_detection_job` resource SHALL embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: Model satisfies ElasticsearchResourceModel
The model SHALL implement `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()`.

#### Scenario: Model getters
- **WHEN** a model is created
- **THEN** getters SHALL return correct field values

### Requirement: Create callback creates the ML job
The `createFunc` callback SHALL build the job config and call the Create Anomaly Detection Job API.

#### Scenario: Create job
- **GIVEN** a valid planned model
- **WHEN** the create callback runs
- **THEN** it SHALL create the job
- **AND** the envelope SHALL read back and persist state

### Requirement: Update override updates the ML job
The concrete resource SHALL override `Update` because updating the job configuration requires comparing the planned model with prior Terraform state.

#### Scenario: Update job
- **GIVEN** an existing job with changed description
- **WHEN** the concrete Update override runs
- **THEN** it SHALL update the job
- **AND** read back and persist refreshed state

### Requirement: Read callback fetches the job
The `readFunc` callback SHALL GET the job config and populate the model.

#### Scenario: Successful read
- **GIVEN** a job exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true`

#### Scenario: Missing job removes from state
- **GIVEN** the job does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback deletes the job
The `deleteFunc` callback SHALL close and delete the job.

#### Scenario: Delete job
- **GIVEN** a job exists
- **WHEN** the delete callback runs
- **THEN** it SHALL close and delete the job

### Requirement: Custom ImportState parses composite id
The concrete resource SHALL implement `ImportState` that parses a composite id and sets `job_id`.

#### Scenario: Import job
- **GIVEN** a composite import id `<cluster_uuid>/<job_id>`
- **WHEN** import runs
- **THEN** the `id` SHALL be stored
- **AND** `job_id` SHALL be set from the resource portion

