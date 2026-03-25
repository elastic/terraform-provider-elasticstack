## MODIFIED Requirements

### Requirement: Elastic Stack bootstrap (REQ-009)

The job SHALL start the Fleet-oriented Docker Compose stack by running `make docker-fleet`. The workflow SHALL rely on GitHub repository environment settings to provide the bootstrap credential variables used by Docker Compose, including `ELASTICSEARCH_PASSWORD` and `KIBANA_PASSWORD`, and SHALL NOT define workflow-local default credential values in `jobs.copilot-setup-steps.env`.

#### Scenario: Stack containers start

- **GIVEN** GitHub-managed environment configuration provides the bootstrap credentials expected by Docker Compose
- **WHEN** the stack setup step runs
- **THEN** `make docker-fleet` SHALL complete successfully before dependency and Kibana steps

#### Scenario: Workflow does not embed bootstrap defaults

- **WHEN** the workflow YAML is inspected
- **THEN** the `copilot-setup-steps` job SHALL NOT declare bootstrap credential defaults in `jobs.copilot-setup-steps.env`

### Requirement: Kibana system user password (REQ-011–REQ-012)

The job SHALL run `make set-kibana-password`. The step SHALL pass `ELASTICSEARCH_PASSWORD`, `KIBANA_SYSTEM_USERNAME`, and `KIBANA_SYSTEM_PASSWORD` from GitHub repository environment settings so curl-based password changes authenticate as the Elasticsearch superuser and target the configured Kibana system user. The workflow SHALL NOT declare default values for these variables in the job definition.

#### Scenario: Credentials align with configured environment

- **GIVEN** Elasticsearch is listening for bootstrap operations
- **AND** GitHub-managed environment configuration provides `ELASTICSEARCH_PASSWORD`, `KIBANA_SYSTEM_USERNAME`, and `KIBANA_SYSTEM_PASSWORD`
- **WHEN** `set-kibana-password` runs
- **THEN** environment variables SHALL supply values consistent with the running stack’s configured `elastic` and Kibana system user passwords

#### Scenario: Manual run uses GitHub-managed configuration

- **GIVEN** a maintainer runs the workflow via `workflow_dispatch`
- **AND** the required GitHub-managed environment configuration is available to the workflow
- **WHEN** the setup job executes
- **THEN** the job SHALL use those provided values instead of workflow-defined defaults

### Requirement: Elasticsearch API key for the agent (REQ-013–REQ-014)

The job SHALL include a step that runs `make create-es-api-key`, parses JSON with `jq` to read the `encoded` API key, and appends `apikey=<value>` to `GITHUB_OUTPUT` for the step. The step SHALL supply `ELASTICSEARCH_PASSWORD` from GitHub repository environment settings for authenticated API key creation.

#### Scenario: API key output is published

- **GIVEN** Elasticsearch accepts security API calls
- **AND** GitHub-managed environment configuration provides `ELASTICSEARCH_PASSWORD`
- **WHEN** the API key step succeeds
- **THEN** the step output named `apikey` SHALL contain the base64-encoded API key material from the `encoded` field

### Requirement: Fleet policy bootstrap (REQ-015–REQ-016)

The job SHALL run `make setup-kibana-fleet` with `ELASTICSEARCH_PASSWORD` supplied from GitHub repository environment settings for authenticated Kibana Fleet API calls. The step SHALL set `FLEET_NAME` to `fleet` so Fleet server host URLs match the Compose service name expected by the Makefile’s `FLEET_ENDPOINT` construction.

#### Scenario: Fleet defaults match Compose service

- **GIVEN** Kibana is available on localhost
- **AND** GitHub-managed environment configuration provides `ELASTICSEARCH_PASSWORD`
- **WHEN** Fleet setup runs
- **THEN** `FLEET_NAME` SHALL be `fleet` for that step
