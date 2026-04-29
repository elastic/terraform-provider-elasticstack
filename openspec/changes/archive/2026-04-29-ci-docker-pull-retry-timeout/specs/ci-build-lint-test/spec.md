## MODIFIED Requirements

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job and the change-classification job. The acceptance test job SHALL run with a non-fail-fast matrix covering configured stack versions and included version-specific overrides. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. The acceptance test job SHALL execute only when the preflight gate outputs `should_run=true` and the change-classification job reports `provider_changes=true`.

For each matrix entry, the job SHALL free disk space, set up Go and Terraform, run `make vendor`, start the stack via Docker Compose, and wait for Elasticsearch and Kibana readiness. Fleet setup and forced synthetics installation SHALL run only for configured version subsets. Acceptance tests SHALL run via `make testacc`, with snapshot versions allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.

The stack-start step SHALL have a step-level timeout so that a hung container image pull fails fast instead of consuming the full job timeout.

#### Scenario: Provider change runs stack and tests

- **GIVEN** a matrix version and runner
- **AND** the preflight gate allows execution
- **AND** the change-classification job reports `provider_changes=true`
- **WHEN** the test job executes
- **THEN** the stack SHALL be provisioned, readiness waits SHALL pass, and `make testacc` SHALL run with the documented policy for snapshots

#### Scenario: OpenSpec-only change skips matrix acceptance

- **GIVEN** a workflow run whose changed files are all under `openspec/`
- **WHEN** the acceptance test job evaluates its execution conditions
- **THEN** the matrix acceptance `test` job SHALL be skipped

#### Scenario: Compose step timeout prevents hung pull

- **GIVEN** the Docker Compose stack-start step runs
- **AND** the image pull or container startup stalls beyond the step timeout
- **WHEN** the configured step timeout is reached
- **THEN** the step SHALL fail and the job SHALL exit early

## ADDED Requirements

### Requirement: Pre-pull fallback fleet image with retry

Before starting the stack via Docker Compose, the workflow SHALL pre-pull the fleet image for matrix entries that use a Docker Hub fallback image. The pre-pull step SHALL use a timeout per attempt and SHALL retry up to three times with backoff. This step SHALL be skipped for matrix entries that use the default `docker.elastic.co` registry.

#### Scenario: Docker Hub fleet image is pre-pulled successfully

- **GIVEN** a matrix entry with `fleetImage` set to a Docker Hub image
- **WHEN** the pre-pull step executes
- **THEN** the image SHALL be pulled with a per-attempt timeout
- **AND** failed attempts SHALL be retried up to three times
- **AND** on success, the subsequent `docker compose up` SHALL use the already-pulled image

#### Scenario: Pre-pull is skipped for docker.elastic.co images

- **GIVEN** a matrix entry without a `fleetImage` override
- **WHEN** the test job step list is evaluated
- **THEN** the pre-pull step SHALL be skipped
- **AND** the stack-start step SHALL proceed normally
