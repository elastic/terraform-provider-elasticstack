# `build-lint-test` — Workflow Requirements

Workflow implementation: `.github/workflows/test.yml`

## Purpose

Define the main CI workflow: preflight gate, build, lint (including OpenSpec validation), matrix acceptance tests against Elastic Stack versions, diagnostics, teardown, and optional PR auto-approve.

## Schema

```yaml
on:
  push:
    branches: ['**']
    tags-ignore: ['v*']
    paths-ignore: ['README.md', 'CHANGELOG.md']
  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]
    paths-ignore: ['README.md', 'CHANGELOG.md']
  workflow_dispatch: {}

permissions:
  contents: read
```
## Requirements
### Requirement: Workflow identity and triggers (REQ-001–REQ-006)

The workflow name SHALL be `Build/Lint/Test`. The workflow SHALL run on `push` to any branch, excluding tag refs matching `v*` and excluding changes limited to `README.md` and `CHANGELOG.md`. The workflow SHALL run on `pull_request`, excluding changes limited to `README.md` and `CHANGELOG.md`. The workflow SHALL run on `pull_request` events of type `ready_for_review` (in addition to default types `opened`, `synchronize`, `reopened`). The workflow SHALL support manual execution via `workflow_dispatch`.

#### Scenario: Push to feature branch

- GIVEN a push that is not a `v*` tag and not only ignored paths
- WHEN the preflight gate allows execution
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

The matrix acceptance test job SHALL depend on successful completion of the `build` job and the change-classification job. The acceptance test job SHALL run with a non-fail-fast matrix covering configured stack versions and included version-specific overrides. The configured stack versions SHALL NOT include Elastic Stack versions below `8.0.0`. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. The acceptance test job SHALL execute only when the preflight gate outputs `should_run=true` and the change-classification job reports `provider_changes=true`.

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

- **GIVEN** Docker Compose is starting the stack for a matrix entry
- **AND** a container image pull or stack startup hangs
- **WHEN** the configured step timeout is reached
- **THEN** the step SHALL fail and the job SHALL exit early

#### Scenario: Matrix excludes 7.x stack versions

- **WHEN** the acceptance matrix is evaluated
- **THEN** every configured stack version SHALL be `8.0.0` or higher, except snapshot labels that represent later unreleased stack versions

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

The `auto-approve` job SHALL depend on the `Test Validation` job and SHALL only run on `pull_request` events. For non-`ready_for_review` events, `auto-approve` SHALL require `Test Validation` to succeed before it runs. For `ready_for_review` events, `auto-approve` SHALL be eligible to run regardless of `Test Validation`'s outcome (because the preflight gate intentionally skips acceptance work, and `Test Validation` succeeds on the preflight-skip path). The `auto-approve` job SHALL execute `go run ./scripts/auto-approve`; approval policy and gate behavior are defined in [`openspec/specs/ci-pr-auto-approve/spec.md`](../ci-pr-auto-approve/spec.md). The `auto-approve` job SHALL request `contents: read` and `pull-requests: write` permissions.

#### Scenario: Auto-approve after satisfied validation

- **GIVEN** a pull request workflow and successful `Test Validation`
- **WHEN** auto-approve runs
- **THEN** it SHALL invoke `go run ./scripts/auto-approve` with the specified permissions

### Requirement: Supply chain for actions (REQ-022)

Third-party actions in the workflow SHALL be pinned by commit SHA.

#### Scenario: Action references

- GIVEN a third-party action is used in the workflow
- WHEN the workflow YAML is inspected
- THEN the action reference SHALL use a commit SHA

### Requirement: Preflight gate (REQ-023–REQ-027)

The workflow SHALL evaluate whether to execute CI jobs via a dedicated preflight gate job that emits a `should_run` output.

For `push` events, the preflight gate SHALL set `should_run=true` when either:

* No open pull request exists for the pushed branch in the same repository
* All commits in the push event were authored by an allowed bot user: Copilot coding agent (`198982749+Copilot@users.noreply.github.com`) or GitHub Actions (`41898282+github-actions[bot]@users.noreply.github.com`)

For `push` events where **neither** of the above holds, the preflight gate SHALL set `should_run=false`.

For non-`push` events (`pull_request` and `workflow_dispatch`), the preflight gate SHALL set `should_run=true`, except for `pull_request` events of type `ready_for_review` where it SHALL set `should_run=false`.

The `build`, `lint`, and matrix acceptance `test` jobs SHALL only execute when the preflight gate outputs `should_run=true`.

#### Scenario: Push without open PR

- GIVEN a push to a branch with no open PR in the same repository
- WHEN preflight runs
- THEN `should_run` SHALL be `true`

#### Scenario: Push with open PR and all commits by an allowed bot user

- GIVEN a push to a branch that has an open PR from the same repo
- AND every commit in the push event was authored by Copilot coding agent (`198982749+Copilot@users.noreply.github.com`) or GitHub Actions (`41898282+github-actions[bot]@users.noreply.github.com`)
- WHEN preflight runs
- THEN `should_run` SHALL be `true`

#### Scenario: Push with open PR and a commit not by an allowed bot user

- GIVEN a push to a branch that has an open PR from the same repo
- AND at least one commit in the push event was not authored by Copilot coding agent (`198982749+Copilot@users.noreply.github.com`) or GitHub Actions (`41898282+github-actions[bot]@users.noreply.github.com`)
- WHEN preflight runs
- THEN `should_run` SHALL be `false` and downstream jobs SHALL be skipped

### Requirement: Job permissions (REQ-028–REQ-029)

The preflight gate job SHALL request the minimum permissions required to inspect pull requests (`contents: read`, `pull-requests: read`). The acceptance test job SHALL request `contents: read`, `issues: write`, and `pull-requests: write` permissions.

#### Scenario: Preflight permissions

- GIVEN the preflight job definition
- WHEN permissions are evaluated
- THEN they SHALL match the minimum set for listing PRs

### Requirement: Ready-for-review behavior (REQ-030)

On `ready_for_review` `pull_request` events, the workflow SHALL keep the preflight gate behavior that prevents the `build`, `lint`, change-classification, and matrix acceptance `test` jobs from running. The `Test Validation` job SHALL succeed based on the intentional preflight skip, and `auto-approve` SHALL remain eligible to run.

#### Scenario: Ready for review event

- **GIVEN** a `pull_request` with action `ready_for_review`
- **WHEN** the workflow runs
- **THEN** `build`, `lint`, change-classification, and matrix acceptance `test` SHALL be skipped by the preflight gate
- **AND** `Test Validation` SHALL succeed
- **AND** auto-approve SHALL be eligible to run

### Requirement: Change classification gate (REQ-032–REQ-033)

The workflow SHALL evaluate whether matrix acceptance tests are required for the current change set via a dedicated change-classification job, but only when the preflight gate permits downstream CI execution by outputting `should_run=true`. When the preflight gate outputs `should_run=false`, the change-classification job SHALL be intentionally skipped. In the first iteration, when the classifier runs, it SHALL set `provider_changes=false` only when every changed file for the workflow run is under `openspec/`; any change set containing a path outside `openspec/` SHALL set `provider_changes=true`.

When the change-classification job runs, it SHALL expose its result as a workflow output that downstream jobs can consume when deciding whether acceptance coverage is required.

#### Scenario: OpenSpec-only change set

- **GIVEN** a workflow run whose changed files are all under `openspec/`
- **AND** the preflight gate outputs `should_run=true`
- **WHEN** the change-classification job evaluates the diff
- **THEN** it SHALL report `provider_changes=false`

#### Scenario: Provider-impacting change set

- **GIVEN** a workflow run whose changed files include at least one path outside `openspec/`
- **AND** the preflight gate outputs `should_run=true`
- **WHEN** the change-classification job evaluates the diff
- **THEN** it SHALL report `provider_changes=true`

### Requirement: Test validation job (REQ-034–REQ-036)

The workflow SHALL publish a `Test Validation` job that always reports a final acceptance-gate result for the workflow run. The validation job SHALL evaluate the preflight output, the change-classification output, and the matrix acceptance job result.

The `Test Validation` job SHALL succeed when any of the following is true:

* The preflight gate intentionally disables downstream CI execution
* The change-classification job reports `provider_changes=false` and the matrix acceptance `test` job is intentionally skipped
* The matrix acceptance `test` job completes successfully

When the preflight gate allows downstream execution, the `Test Validation` job SHALL fail if either of the following is true:

* The change-classification job reports `provider_changes=true` and the matrix acceptance `test` job does not complete successfully
* The change-classification job reports `provider_changes=false` and the matrix acceptance `test` job still runs but does not complete successfully

The validation job SHALL provide a stable required-check target that can be used by GitHub branch protection or rulesets instead of the per-version matrix acceptance checks.

#### Scenario: OpenSpec-only pull request

- **GIVEN** a pull request whose changed files are all under `openspec/`
- **WHEN** the workflow reaches `Test Validation`
- **THEN** the matrix acceptance `test` job SHALL be treated as intentionally skipped
- **AND** `Test Validation` SHALL succeed

#### Scenario: Provider change with failing acceptance coverage

- **GIVEN** a workflow run with `provider_changes=true`
- **AND** the matrix acceptance `test` job does not complete successfully
- **WHEN** `Test Validation` evaluates the workflow state
- **THEN** `Test Validation` SHALL fail

### Requirement: Generated changelog pull requests can reach auto-approve without full CI
The `Build/Lint/Test` workflow SHALL allow same-repository pull requests from branch `generated-changelog` that are authored by `github-actions[bot]` and modify only `CHANGELOG.md` to reach the `auto-approve` job without requiring the full build, lint, change-classification, or matrix acceptance test path to run. The skip condition MUST verify all three criteria — branch name, PR author, and file list — in the preflight gate before setting `should_run=false`.

#### Scenario: Generated changelog PR reaches auto-approve path
- **GIVEN** a same-repository pull request from branch `generated-changelog`
- **AND** the PR author is `github-actions[bot]`
- **AND** the pull request changes only `CHANGELOG.md`
- **WHEN** the workflow evaluates its execution path
- **THEN** the workflow SHALL produce a successful path that leaves `auto-approve` eligible to run
- **AND** it SHALL NOT require the full build, lint, and matrix acceptance test jobs for that PR

#### Scenario: Auto-merge is gated on the approval outcome
- **GIVEN** a `generated-changelog` PR
- **WHEN** the `auto-approve` job runs
- **THEN** auto-merge SHALL only be enabled if the auto-approve script determined `ShouldApprove` or `AlreadyApproved` is true (reported via a `GITHUB_OUTPUT` step output)
- **AND** auto-merge SHALL NOT be enabled if the auto-approve gates reject the PR

### Requirement: Changelog-only bypass remains narrowly scoped
The `Build/Lint/Test` workflow SHALL keep the changelog-only bypass narrowly scoped to the generated changelog automation shape. Other changelog-only pull requests SHALL NOT gain the same bypass unless they satisfy all three repository-authored generated-changelog conditions: branch name `generated-changelog`, PR author `github-actions[bot]`, and files limited to `CHANGELOG.md`.

#### Scenario: Manual changelog-only PR does not inherit generated-changelog bypass
- **GIVEN** a pull request changes only `CHANGELOG.md`
- **AND** its head branch name is not `generated-changelog`
- **WHEN** the workflow evaluates bypass conditions
- **THEN** it SHALL NOT treat that pull request as the generated-changelog special case

#### Scenario: Wrong author does not inherit generated-changelog bypass
- **GIVEN** a pull request from branch `generated-changelog` changes only `CHANGELOG.md`
- **AND** the PR author is not `github-actions[bot]`
- **WHEN** the workflow evaluates bypass conditions
- **THEN** it SHALL run full CI rather than skipping to the auto-approve path

