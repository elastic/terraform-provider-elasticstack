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
    env:
      ELASTIC_PASSWORD: password
      KIBANA_SYSTEM_USERNAME: kibana_system
      KIBANA_SYSTEM_PASSWORD: password
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

The job SHALL start the Fleet-oriented Docker Compose stack by running `make docker-fleet`. The step SHALL set `ELASTICSEARCH_PASSWORD` and `KIBANA_PASSWORD` from the job’s default superuser password (`ELASTIC_PASSWORD`) so Compose receives the same values used to bootstrap Elasticsearch and Kibana in `docker-compose.yml` (aligned with `.github/workflows/test.yml`).

#### Scenario: Stack containers start

- GIVEN a clean runner with Docker available
- WHEN the stack setup step runs
- THEN `make docker-fleet` SHALL complete successfully before dependency and Kibana steps

### Requirement: Repository setup target (REQ-010)

The job SHALL run `make setup` after the stack is started.

#### Scenario: Project setup completes

- GIVEN the stack step succeeded
- WHEN `make setup` runs
- THEN project setup tasks defined by the Makefile target SHALL complete successfully

### Requirement: Kibana system user password (REQ-011–REQ-012)

The job SHALL run `make set-kibana-password`. The job SHALL define default values on the job-level `env` for `ELASTIC_PASSWORD`, `KIBANA_SYSTEM_USERNAME`, and `KIBANA_SYSTEM_PASSWORD` (matching Makefile defaults and the acceptance-test workflow pattern) so `workflow_dispatch` and path-filtered runs succeed without repository secrets. The step SHALL set `ELASTICSEARCH_PASSWORD` from `ELASTIC_PASSWORD` and pass `KIBANA_SYSTEM_USERNAME` and `KIBANA_SYSTEM_PASSWORD` from the job environment so curl-based password changes authenticate as the Elasticsearch superuser and target the correct Kibana system user. Higher-precedence environment configuration (e.g. Copilot or repository environment variables) MAY override these defaults when injected at job scope.

#### Scenario: Credentials align with Compose

- GIVEN Elasticsearch is listening for bootstrap operations
- WHEN `set-kibana-password` runs
- THEN environment variables SHALL supply values consistent with the running stack’s configured `elastic` and Kibana system user passwords (as consumed by the Makefile and `docker-compose.yml`)

#### Scenario: Self-contained manual run

- GIVEN no repository secrets are required for the default stack password
- WHEN a maintainer runs the workflow via `workflow_dispatch`
- THEN the job-level defaults SHALL be sufficient for stack bootstrap and subsequent Makefile steps

### Requirement: Elasticsearch API key for the agent (REQ-013–REQ-014)

The job SHALL include a step that runs `make create-es-api-key`, parses JSON with `jq` to read the `encoded` API key, and appends `apikey=<value>` to `GITHUB_OUTPUT` for the step. The step SHALL supply `ELASTICSEARCH_PASSWORD` to the environment, derived from the job’s `ELASTIC_PASSWORD` default (or override), for authenticated API key creation.

#### Scenario: API key output is published

- GIVEN Elasticsearch accepts security API calls
- WHEN the API key step succeeds
- THEN the step output named `apikey` SHALL contain the base64-encoded API key material from the `encoded` field

### Requirement: Fleet policy bootstrap (REQ-015–REQ-016)

The job SHALL run `make setup-kibana-fleet` with `ELASTICSEARCH_PASSWORD` set from the job’s `ELASTIC_PASSWORD` default (or override) for authenticated Kibana Fleet API calls. The step SHALL set `FLEET_NAME` to `fleet` so Fleet server host URLs match the Compose service name expected by the Makefile’s `FLEET_ENDPOINT` construction.

#### Scenario: Fleet defaults match Compose service

- GIVEN Kibana is available on localhost
- WHEN Fleet setup runs
- THEN `FLEET_NAME` SHALL be `fleet` for that step

### Requirement: Supply chain for actions (REQ-017)

Third-party actions in the workflow SHALL be pinned by commit SHA.

#### Scenario: Action references

- GIVEN a third-party action is used in the workflow
- WHEN the workflow YAML is inspected
- THEN the action reference SHALL use a commit SHA
