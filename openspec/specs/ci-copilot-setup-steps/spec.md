# `copilot-setup-steps` — Workflow Requirements

Workflow implementation: `.github/workflows/copilot-setup-steps.yml`

## Purpose

Define the GitHub Actions workflow that GitHub Copilot uses as **setup steps** for this repository: provision a local Elastic Stack with Fleet, install development dependencies, configure Kibana system credentials, expose an Elasticsearch API key for the agent environment, and initialize Fleet policies—so Copilot sessions can run against a stack aligned with acceptance-test conventions.

## Schema

```yaml
on:
  workflow_dispatch: {}
  push:
    paths:
      - .github/workflows/copilot-setup-steps.yml
  pull_request:
    paths:
      - .github/workflows/copilot-setup-steps.yml

jobs:
  copilot-setup-steps:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    # Toolchain order (implementation uses SHA-pinned actions per REQ-017):
    # checkout → setup-node (node-version-file: package.json; npm cache via package-lock.json)
    # → setup-go (go.mod) → setup-terraform → stack and Makefile steps
```

## Requirements

### Requirement: Workflow identity and triggers (REQ-001–REQ-003)

The workflow name SHALL be `Copilot Setup Steps`. On `push` and `pull_request`, the workflow SHALL use GitHub Actions path filters such that it runs when **at least one** changed file matches `.github/workflows/copilot-setup-steps.yml` (matching is not limited to exclusive changes to that file). The workflow SHALL support manual execution via `workflow_dispatch`.

#### Scenario: Workflow file change opens a pull request

- GIVEN a pull request that modifies `.github/workflows/copilot-setup-steps.yml`
- WHEN the pull request triggers Actions
- THEN this workflow SHALL run

#### Scenario: Pull request also changes other files

- GIVEN a pull request that modifies `.github/workflows/copilot-setup-steps.yml` and other paths
- WHEN the pull request triggers Actions
- THEN this workflow SHALL still run because the workflow file is among the changed paths

#### Scenario: Manual validation

- GIVEN a maintainer runs the workflow from the Actions tab
- WHEN `workflow_dispatch` is used
- THEN the setup job SHALL execute without requiring a change to the workflow file

### Requirement: Copilot job identity (REQ-004)

The sole job id SHALL be `copilot-setup-steps`. The job SHALL run on `ubuntu-latest`.

#### Scenario: Copilot discovers setup steps

- GIVEN GitHub Copilot evaluates repository setup automation
- WHEN it looks for the designated setup job
- THEN it SHALL find a job named exactly `copilot-setup-steps`

### Requirement: Job permissions (REQ-005)

The job SHALL request `contents: read` and SHALL not request broader permissions in the workflow definition than needed for checkout and read-only repository access during the documented steps.

#### Scenario: Minimal permissions

- GIVEN the job definition in the workflow file
- WHEN permissions are evaluated
- THEN `contents: read` SHALL be the only explicit permission grant

### Requirement: Toolchain and checkout (REQ-006–REQ-008)

The job SHALL check out the repository using `actions/checkout` pinned by commit SHA. The job SHALL install Node.js using `actions/setup-node` pinned by commit SHA with **`node-version-file` set to the repository root `package.json`** and **without** a conflicting `node-version` input that would override the file (per the action’s documented behavior). The resolved version SHALL follow the action’s documented precedence among `package.json` fields (`volta.node`, then `devEngines.runtime` for node, then `engines.node`). The job SHALL enable npm caching and use `package-lock.json` as the cache dependency path. The job SHALL install Go using `actions/setup-go` with `go-version-file: go.mod` and Go module caching enabled. The job SHALL install Terraform using `hashicorp/setup-terraform` with `terraform_wrapper: false`.

#### Scenario: Go version tracks the module

- GIVEN `go.mod` specifies the toolchain
- WHEN setup-go runs
- THEN the Go version SHALL be derived from `go.version-file` / `go.mod` per the action configuration

#### Scenario: Node satisfies the version read from package.json

- GIVEN the repository root `package.json` declares a Node version requirement via the fields `actions/setup-node` reads for `node-version-file` (in precedence order)
- WHEN setup-node runs
- THEN the job SHALL provision a Node.js version that satisfies the semver range (or exact version) resolved from that file so `node` and `npm` meet the repository’s declared requirement for OpenSpec and npm-based Makefile targets

#### Scenario: npm dependencies cached for setup

- GIVEN `package-lock.json` is present in the repository
- WHEN setup-node runs with npm caching configured
- THEN the action SHALL use `package-lock.json` as the cache dependency path for npm

### Requirement: Elastic Stack bootstrap (REQ-009)

The job SHALL start the Fleet-oriented Docker Compose stack by running `make docker-fleet`. The workflow SHALL NOT define workflow-local default credential values in `jobs.copilot-setup-steps.env`. Stack bootstrap SHALL rely on the repository `.env` defaults for Docker Compose bootstrap values such as `ELASTICSEARCH_PASSWORD` and `KIBANA_PASSWORD`, unless the execution environment overrides them before the Make target runs.

#### Scenario: Stack containers start

- GIVEN the repository `.env` file provides the Docker Compose bootstrap defaults
- WHEN the stack setup step runs
- THEN `make docker-fleet` SHALL complete successfully before dependency and Kibana steps

#### Scenario: Workflow leaves bootstrap defaults external

- WHEN the workflow YAML is inspected
- THEN the workflow SHALL not declare workflow-level bootstrap credential defaults

#### Scenario: Workflow does not embed bootstrap defaults

- WHEN the workflow YAML is inspected
- THEN the `copilot-setup-steps` job SHALL NOT declare bootstrap credential defaults in `jobs.copilot-setup-steps.env`

### Requirement: Repository setup target (REQ-010)

The job SHALL run `make setup` after the stack is started.

#### Scenario: Project setup completes

- GIVEN the stack step succeeded
- WHEN `make setup` runs
- THEN project setup tasks defined by the Makefile target SHALL complete successfully

### Requirement: Kibana system user password (REQ-011–REQ-012)

The job SHALL run `make set-kibana-password` without step-specific credential overrides. The step SHALL rely on the existing Makefile defaults for `ELASTICSEARCH_USERNAME`, `KIBANA_SYSTEM_USERNAME`, and `KIBANA_SYSTEM_PASSWORD` unless the execution context overrides them. The workflow SHALL NOT declare default values for these variables in the job definition.

#### Scenario: Credentials align with configured environment

- GIVEN Elasticsearch is listening for bootstrap operations
- AND the Makefile defaults remain `ELASTICSEARCH_USERNAME=elastic`, `KIBANA_SYSTEM_USERNAME=kibana_system`, and `KIBANA_SYSTEM_PASSWORD=password` unless explicitly overridden
- WHEN `set-kibana-password` runs
- THEN environment variables SHALL supply values consistent with the running stack’s configured `elastic` and Kibana system user passwords

#### Scenario: Manual run uses externally provided configuration

- GIVEN a maintainer runs the workflow via `workflow_dispatch`
- AND the repository defaults remain available for the Elasticsearch and Kibana system user values unless the execution environment overrides them
- WHEN the setup job executes
- THEN the job SHALL use those provided values instead of workflow-defined defaults

### Requirement: Elasticsearch API key for the agent (REQ-013–REQ-014)

The job SHALL include a step that runs `make create-es-api-key`, parses JSON with `jq` to read the `encoded` API key, and appends `apikey=<value>` to `GITHUB_OUTPUT` for the step. The step SHALL rely on the current Makefile authentication defaults unless the execution environment overrides them.

#### Scenario: API key output is published

- GIVEN Elasticsearch accepts security API calls
- AND the Makefile authentication defaults or execution-environment overrides provide the Elasticsearch connection settings used by the Makefile
- WHEN the API key step succeeds
- THEN the step output named `apikey` SHALL contain the base64-encoded API key material from the `encoded` field

### Requirement: Fleet policy bootstrap (REQ-015–REQ-016)

The job SHALL run `make setup-kibana-fleet` while relying on the current Makefile authentication defaults unless the execution environment overrides them. The step SHALL set `FLEET_NAME` to `fleet` so Fleet server host URLs match the Compose service name expected by the Makefile’s `FLEET_ENDPOINT` construction.

#### Scenario: Fleet defaults match Compose service

- GIVEN Kibana is available on localhost
- AND the Makefile authentication defaults or execution-environment overrides provide the Elasticsearch connection settings used by the Makefile
- WHEN Fleet setup runs
- THEN `FLEET_NAME` SHALL be `fleet` for that step

### Requirement: Supply chain for actions (REQ-017)

Third-party actions in the workflow SHALL be pinned by commit SHA.

#### Scenario: Action references

- GIVEN a third-party action is used in the workflow
- WHEN the workflow YAML is inspected
- THEN the action reference SHALL use a commit SHA

### Requirement: Failure diagnostics for Elastic Stack setup (REQ-018)

When the `copilot-setup-steps` job fails after attempting repository setup or Elastic Stack bootstrap, the workflow SHALL run `docker compose logs --no-color` so the job output includes plain-text diagnostics from the local Elastic Stack services. This diagnostic step SHALL be part of the failure path only and SHALL not run for successful workflow executions.

#### Scenario: Setup failure emits Docker Compose logs

- WHEN the `copilot-setup-steps` job fails during or after the steps that bootstrap the Elastic Stack and related setup
- THEN the workflow SHALL execute `docker compose logs --no-color` before the job finishes

#### Scenario: Successful setup does not emit diagnostic logs

- WHEN the `copilot-setup-steps` job completes successfully
- THEN the workflow SHALL not run the Docker Compose log collection step
