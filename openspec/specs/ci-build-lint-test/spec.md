# `build-lint-test` — Workflow Requirements

Workflow implementation: `.github/workflows/provider.yml`

## Purpose

Define the main CI workflow: build, lint (including OpenSpec validation), matrix acceptance tests against Elastic Stack versions, diagnostics, teardown, and optional PR auto-approve.

## Schema

```yaml
on:
  push:
    branches: [main]
  pull_request:
    types: [opened, synchronize, reopened]
  workflow_dispatch: {}

permissions:
  contents: read
```
## Requirements
### Requirement: Workflow identity and triggers (REQ-001–REQ-006)

The workflow name SHALL be `Provider CI`. The workflow SHALL run on `push` to branch `main`. The workflow SHALL run on `pull_request` events of type `opened`, `synchronize`, and `reopened`. The workflow SHALL support manual execution via `workflow_dispatch`.

#### Scenario: Push to main

- GIVEN a `push` to `main`
- WHEN the change-classification job reports `provider_changes=true`
- THEN build, lint, and test jobs MAY run per other requirements

### Requirement: Build and lint jobs (REQ-007–REQ-008, REQ-031)

The `build` job SHALL run on `ubuntu-latest`, set up Go from `go.mod`, set up Node.js (24.x), run `make vendor`, run `make workflow-test`, run `make hook-test`, and run `make build-ci`. The `lint` job SHALL run on `ubuntu-latest`, set up Go from `go.mod`, read the Terraform CLI version from the repository root `.terraform-version` file, set up Terraform without wrapper mode using that pinned version, install Node.js (24.x), run `npm ci`, run `openspec validate --specs` with telemetry disabled, and run `make check-lint`.

#### Scenario: Build job runs workflow and hook tests

- GIVEN the build job runs after Go and Node setup complete
- WHEN the pre-build verification steps execute
- THEN `make workflow-test` SHALL run before `make build-ci`
- AND `make hook-test` SHALL run before `make build-ci`

#### Scenario: Lint validates OpenSpec

- GIVEN the lint job runs after dependencies are installed
- WHEN OpenSpec specs are present under `openspec/specs/`
- THEN `openspec validate --specs` SHALL run successfully before Go/terraform lint checks

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job and the change-classification job. The acceptance test job SHALL run with a non-fail-fast matrix covering configured stack versions and included version-specific overrides. The configured stack versions SHALL NOT include Elastic Stack versions below `8.0.0`. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. The acceptance test job SHALL execute only when the change-classification job reports `provider_changes=true`.

For each matrix entry, the job SHALL free disk space, set up Go and Terraform, run `make vendor`, start the stack via Docker Compose, and wait for Elasticsearch and Kibana readiness. Fleet Server host, agent policy, and package policy setup SHALL be provided by the Docker Compose stack start (`make docker-fleet`) and by the acceptance test PreCheck's default agent download source bootstrap, without any additional per-version-gated Fleet setup step. Forced synthetics installation SHALL run only for configured version subsets. Acceptance tests SHALL run via `make testacc`, with snapshot versions allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.

The stack-start step SHALL have a step-level timeout so that a hung container image pull fails fast instead of consuming the full job timeout.

#### Scenario: Provider change runs stack and tests

- **GIVEN** a matrix version and runner
- **AND** the change-classification job reports `provider_changes=true`
- **WHEN** the test job executes
- **THEN** the stack SHALL be provisioned, readiness waits SHALL pass, and `make testacc` SHALL run with the documented policy for snapshots

#### Scenario: OpenSpec-only change skips matrix acceptance

- **GIVEN** a workflow run whose changed files are all under `openspec/`
- **WHEN** the acceptance test job evaluates its execution conditions
- **THEN** the matrix acceptance `test` job SHALL be skipped

#### Scenario: Compose step timeout prevents hung pull

- **GIVEN** Docker Compose is starting the stack for a matrix entry
- **AND** a container image pull or stack startup hangs
- **WHEN** the configured step timeout is reached
- **THEN** the step SHALL fail and the job SHALL exit early

#### Scenario: Matrix excludes 7.x stack versions

- **WHEN** the acceptance matrix is evaluated
- **THEN** every configured stack version SHALL be `8.0.0` or higher, except snapshot labels that represent later unreleased stack versions

#### Scenario: Fleet bootstrap runs uniformly for every matrix entry

- **GIVEN** any configured matrix version, including one that is not part of any explicit per-version allowlist
- **WHEN** the test job starts the stack via `make docker-fleet`
- **THEN** a default Fleet Server host, a `fleet-server` agent policy, and a `fleet_server` package policy SHALL exist before `make testacc` runs
- **AND** `make testacc`'s `acctest.PreCheck` SHALL ensure a default agent download source exists
- **AND** no separate per-version-gated Fleet setup step SHALL be required for this coverage

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

### Requirement: Snapshot failure PR notice (REQ-015)

On snapshot acceptance failure in `pull_request` events, the workflow SHALL create or update a PR warning comment through `actions/github-script`.

#### Scenario: Snapshot test failure on PR

- GIVEN a snapshot matrix entry fails during a pull request build
- WHEN the failure handling step runs
- THEN a bot comment SHALL be created or updated on the PR with a defined marker

### Requirement: Failure diagnostics and teardown (REQ-016–REQ-017)

The workflow SHALL emit Docker Compose logs when the job fails or acceptance tests fail. The workflow SHALL always tear down the Docker Compose stack via `make docker-clean`, regardless of prior step outcomes.

#### Scenario: Always tear down

- GIVEN any prior step outcome in the test job
- WHEN the job finishes
- THEN `make docker-clean` SHALL run in an `always()` step

### Requirement: Auto-approve job (REQ-018–REQ-021)

The `auto-approve` job SHALL depend on the `gate` job and SHALL only run on `pull_request` events. `auto-approve` SHALL require the `gate` job to succeed before it runs, unconditionally — there is no event-type carve-out. The `auto-approve` job SHALL execute `go run ./scripts/auto-approve`; approval policy and gate behavior are defined in [`openspec/specs/ci-pr-auto-approve/spec.md`](../ci-pr-auto-approve/spec.md). The `auto-approve` job SHALL request `contents: read` and `pull-requests: write` permissions.

#### Scenario: Auto-approve after satisfied validation

- **GIVEN** a pull request workflow and a successful `gate` job result
- **WHEN** auto-approve runs
- **THEN** it SHALL invoke `go run ./scripts/auto-approve` with the specified permissions

#### Scenario: Auto-approve does not run when the gate fails

- **GIVEN** a pull request workflow and a `gate` job result other than `success`
- **WHEN** the workflow evaluates whether to run `auto-approve`
- **THEN** the `auto-approve` job SHALL NOT run

### Requirement: Supply chain for actions (REQ-022)

Third-party actions in the workflow SHALL be pinned by commit SHA.

#### Scenario: Action references

- GIVEN a third-party action is used in the workflow
- WHEN the workflow YAML is inspected
- THEN the action reference SHALL use a commit SHA

### Requirement: Job permissions (REQ-028–REQ-029)

The change-classification job SHALL request the minimum permissions required to inspect pull requests (`contents: read`, `pull-requests: read`). The acceptance test job SHALL request `contents: read`, `issues: write`, and `pull-requests: write` permissions.

#### Scenario: Change-classification permissions

- GIVEN the change-classification job definition
- WHEN permissions are evaluated
- THEN they SHALL match the minimum set for listing PRs

### Requirement: Change classification gate (REQ-032–REQ-033)

The workflow SHALL evaluate whether the `build`, `lint`, `golangci-lint`, and matrix acceptance `test` jobs are required for the current change set via a dedicated change-classification job (`classify`) that runs unconditionally on every trigger. For `pull_request` events, the classifier SHALL set `provider_changes=false` only when every changed file is non-impacting: exactly `CHANGELOG.md`, or any path under `openspec/`, or any path under `.agents/`, or any path under `.github/` other than `.github/workflows/provider.yml` itself. Any change set containing at least one path outside that non-impacting set, or an empty changed-file list, SHALL set `provider_changes=true`. For non-`pull_request` events (including `push` and `workflow_dispatch`), the classifier SHALL skip file inspection entirely and unconditionally set `provider_changes=true`.

When the change-classification job runs, it SHALL expose its result as a workflow output that downstream jobs can consume when deciding whether those jobs are required.

#### Scenario: OpenSpec-only change set

- **GIVEN** a `pull_request` workflow run whose changed files are all under `openspec/`
- **WHEN** the change-classification job evaluates the diff
- **THEN** it SHALL report `provider_changes=false`

#### Scenario: Provider-impacting change set

- **GIVEN** a `pull_request` workflow run whose changed files include at least one path outside the non-impacting set
- **WHEN** the change-classification job evaluates the diff
- **THEN** it SHALL report `provider_changes=true`

#### Scenario: Non-pull_request event always classifies as provider-impacting

- **GIVEN** a non-`pull_request` event triggering the workflow (including `push` or `workflow_dispatch`)
- **WHEN** the change-classification job runs
- **THEN** it SHALL report `provider_changes=true` without inspecting the changed-file list

### Requirement: Provider gate job (REQ-034–REQ-036)

The workflow SHALL publish a `gate` job ("Provider Gate") that always reports a final required-check result for the workflow run, evaluating the change-classification result together with the `build`, `lint`, `golangci-lint`, and matrix acceptance `test` job results.

The `gate` job SHALL succeed when either of the following is true:

* The change-classification job reports `provider_changes=false` and `build`, `lint`, `golangci-lint`, and the matrix acceptance `test` job are all intentionally skipped
* `build`, `lint`, `golangci-lint`, and the matrix acceptance `test` job all complete successfully (regardless of the classify result)

The `gate` job SHALL fail when any of the following is true:

* Any of `build`, `lint`, `golangci-lint`, or the matrix acceptance `test` job reports `failure` or `cancelled`
* The change-classification job reports `provider_changes=true` and at least one of `build`, `lint`, `golangci-lint`, or the matrix acceptance `test` job reports an unexpected `skipped` result
* Any other job-result combination, including an unrecognised classify result (not `true`/`false`) or an unrecognised job result value (not one of `success`, `skipped`, `failure`, `cancelled`)

The `gate` job SHALL provide a stable required-check target that can be used by GitHub branch protection or rulesets instead of the per-version matrix acceptance checks or the individual `build`/`lint`/`golangci-lint` checks.

#### Scenario: OpenSpec-only pull request

- **GIVEN** a pull request whose changed files are all under `openspec/`
- **WHEN** the workflow reaches the `gate` job
- **THEN** `build`, `lint`, `golangci-lint`, and the matrix acceptance `test` job SHALL be treated as intentionally skipped
- **AND** the `gate` job SHALL succeed

#### Scenario: Provider change with failing acceptance coverage

- **GIVEN** a workflow run with `provider_changes=true`
- **AND** the matrix acceptance `test` job does not complete successfully
- **WHEN** the `gate` job evaluates the workflow state
- **THEN** the `gate` job SHALL fail

#### Scenario: Provider change with an unexpected skip

- **GIVEN** a workflow run with `provider_changes=true`
- **AND** at least one of `build`, `lint`, `golangci-lint`, or the matrix acceptance `test` job reports `skipped`
- **WHEN** the `gate` job evaluates the workflow state
- **THEN** the `gate` job SHALL fail

### Requirement: Snapshot-to-GA version promotion

When the Elastic Stack release tracked by the acceptance matrix's snapshot-labeled entry
(`<version>-SNAPSHOT`) reaches general availability, the workflow SHALL replace that matrix entry
with the released version string rather than adding a separate, additional matrix entry for the same
stack line. The promoted entry SHALL be added to every per-version step condition (such as Fleet
setup) that had matched the snapshot label only via the `-SNAPSHOT` suffix, so the promoted version
does not lose step coverage it received while labeled as a snapshot. The promoted entry SHALL no
longer match `endsWith(matrix.version, '-SNAPSHOT')` and SHALL therefore be treated as blocking
(`continue-on-error: false`) like every other non-snapshot matrix entry, and SHALL NOT trigger the
snapshot-failure PR warning comment.

#### Scenario: Snapshot entry is promoted to its GA release

- **GIVEN** the acceptance matrix contains a snapshot-labeled entry `X.Y.0-SNAPSHOT` tracking an
  in-development stack line
- **AND** that stack line reaches general availability as `X.Y.0`
- **WHEN** the matrix is updated for the release
- **THEN** the `X.Y.0-SNAPSHOT` entry SHALL be rewritten to `X.Y.0`
- **AND** no additional matrix entry SHALL be added for the same `X.Y` stack line

#### Scenario: Promoted entry keeps per-version step coverage

- **GIVEN** a per-version step condition that previously matched a snapshot entry only via
  `endsWith(matrix.version, '-SNAPSHOT')` (for example, Fleet setup)
- **WHEN** that snapshot entry is promoted to its GA version string
- **THEN** the promoted version string SHALL be added explicitly to that step's condition
- **AND** the step SHALL continue to run for the promoted version exactly as it did while the entry
  was labeled as a snapshot

#### Scenario: Promoted entry becomes blocking

- **GIVEN** a matrix entry that was promoted from a snapshot label to its GA version string
- **WHEN** the acceptance test step (`make testacc`) fails for that entry
- **THEN** `continue-on-error` SHALL NOT apply to that failure
- **AND** the snapshot-failure PR warning comment step SHALL NOT fire for that entry

