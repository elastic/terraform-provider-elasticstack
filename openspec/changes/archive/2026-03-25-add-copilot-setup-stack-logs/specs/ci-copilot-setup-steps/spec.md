## ADDED Requirements

### Requirement: Failure diagnostics for Elastic Stack setup

When the `copilot-setup-steps` job fails after attempting repository setup or Elastic Stack bootstrap, the workflow SHALL run `docker compose logs --no-color` so the job output includes plain-text diagnostics from the local Elastic Stack services. This diagnostic step SHALL be part of the failure path only and SHALL not run for successful workflow executions.

#### Scenario: Setup failure emits Docker Compose logs

- **WHEN** the `copilot-setup-steps` job fails during or after the steps that bootstrap the Elastic Stack and related setup
- **THEN** the workflow SHALL execute `docker compose logs --no-color` before the job finishes

#### Scenario: Successful setup does not emit diagnostic logs

- **WHEN** the `copilot-setup-steps` job completes successfully
- **THEN** the workflow SHALL not run the Docker Compose log collection step
