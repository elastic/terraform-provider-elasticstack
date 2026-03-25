## MODIFIED Requirements

### Requirement: Elastic Stack bootstrap (REQ-009)

The job SHALL start the Fleet-oriented Docker Compose stack by running `make docker-fleet`. The workflow SHALL NOT define workflow-local default credential values in `jobs.copilot-setup-steps.env`. Stack bootstrap SHALL rely on the repository `.env` defaults for Docker Compose bootstrap values such as `ELASTICSEARCH_PASSWORD` and `KIBANA_PASSWORD`, unless the execution environment overrides them before the Make target runs.

#### Scenario: Stack containers start

- **GIVEN** the repository `.env` file provides the Docker Compose bootstrap defaults
- **WHEN** the stack setup step runs
- **THEN** `make docker-fleet` SHALL complete successfully before dependency and Kibana steps

#### Scenario: Workflow leaves bootstrap defaults external

- **WHEN** the workflow YAML is inspected
- **THEN** the workflow SHALL not declare workflow-level bootstrap credential defaults

#### Scenario: Workflow does not embed bootstrap defaults

- **WHEN** the workflow YAML is inspected
- **THEN** the `copilot-setup-steps` job SHALL NOT declare bootstrap credential defaults in `jobs.copilot-setup-steps.env`

### Requirement: Kibana system user password (REQ-011–REQ-012)

The job SHALL run `make set-kibana-password` without step-specific credential overrides. The step SHALL rely on the existing Makefile defaults for `ELASTICSEARCH_USERNAME`, `KIBANA_SYSTEM_USERNAME`, and `KIBANA_SYSTEM_PASSWORD` unless the execution context overrides them. The workflow SHALL NOT declare default values for these variables in the job definition.

#### Scenario: Credentials align with configured environment

- **GIVEN** Elasticsearch is listening for bootstrap operations
- **AND** the Makefile defaults remain `ELASTICSEARCH_USERNAME=elastic`, `KIBANA_SYSTEM_USERNAME=kibana_system`, and `KIBANA_SYSTEM_PASSWORD=password` unless explicitly overridden
- **WHEN** `set-kibana-password` runs
- **THEN** environment variables SHALL supply values consistent with the running stack’s configured `elastic` and Kibana system user passwords

#### Scenario: Manual run uses repository defaults unless overridden

- **GIVEN** a maintainer runs the workflow via `workflow_dispatch`
- **AND** the repository defaults remain available for the Elasticsearch and Kibana system user values unless the execution environment overrides them
- **WHEN** the setup job executes
- **THEN** the job SHALL use those provided values instead of workflow-defined defaults

### Requirement: Elasticsearch API key for the agent (REQ-013–REQ-014)

The job SHALL include a step that runs `make create-es-api-key`, parses JSON with `jq` to read the `encoded` API key, and appends `apikey=<value>` to `GITHUB_OUTPUT` for the step. The step SHALL rely on the current Makefile authentication defaults unless the execution environment overrides them.

#### Scenario: API key output is published

- **GIVEN** Elasticsearch accepts security API calls
- **AND** the Makefile authentication defaults or execution-environment overrides provide the Elasticsearch connection settings used by the Makefile
- **WHEN** the API key step succeeds
- **THEN** the step output named `apikey` SHALL contain the base64-encoded API key material from the `encoded` field

### Requirement: Fleet policy bootstrap (REQ-015–REQ-016)

The job SHALL run `make setup-kibana-fleet` while relying on the current Makefile authentication defaults unless the execution environment overrides them. The step SHALL explicitly set `FLEET_NAME` to `fleet` so Fleet server host URLs match the Compose service name expected by the Makefile’s `FLEET_ENDPOINT` construction.

#### Scenario: Fleet defaults match Compose service

- **GIVEN** Kibana is available on localhost
- **AND** the Makefile authentication defaults or execution-environment overrides provide the Elasticsearch connection settings used by the Makefile
- **WHEN** Fleet setup runs
- **THEN** `FLEET_NAME` SHALL be `fleet` for that step
