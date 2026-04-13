## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job and the change-classification job. The acceptance test job SHALL run with a non-fail-fast matrix covering configured stack versions and included version-specific overrides. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. The acceptance test job SHALL execute only when the preflight gate outputs `should_run=true` and the change-classification job reports `provider_changes=true`.

For each matrix entry, the job SHALL free disk space, set up Go and Terraform, run `make vendor`, start the stack via Docker Compose, and wait for Elasticsearch and Kibana readiness. Fleet setup and forced synthetics installation SHALL run only for configured version subsets. Acceptance tests SHALL run via `make testacc`, with snapshot versions allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.

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

### Requirement: Auto-approve job (REQ-018–REQ-021)

The `auto-approve` job SHALL depend on the `Test Validation` job and SHALL only run on `pull_request` events. For non-`ready_for_review` events, `auto-approve` SHALL require `Test Validation` to succeed before it runs. For `ready_for_review` events, `auto-approve` SHALL be eligible to run regardless of `Test Validation`'s outcome (because the preflight gate intentionally skips acceptance work, and `Test Validation` succeeds on the preflight-skip path). The `auto-approve` job SHALL execute `go run ./scripts/auto-approve`; approval policy and gate behavior are defined in [`openspec/specs/ci-pr-auto-approve/spec.md`](../ci-pr-auto-approve/spec.md). The `auto-approve` job SHALL request `contents: read` and `pull-requests: write` permissions.

#### Scenario: Auto-approve after satisfied validation

- **GIVEN** a pull request workflow and successful `Test Validation`
- **WHEN** auto-approve runs
- **THEN** it SHALL invoke `go run ./scripts/auto-approve` with the specified permissions

### Requirement: Ready-for-review behavior (REQ-030)

On `ready_for_review` `pull_request` events, the workflow SHALL keep the preflight gate behavior that prevents the `build`, `lint`, change-classification, and matrix acceptance `test` jobs from running. The `Test Validation` job SHALL succeed based on the intentional preflight skip, and `auto-approve` SHALL remain eligible to run.

#### Scenario: Ready for review event

- **GIVEN** a `pull_request` with action `ready_for_review`
- **WHEN** the workflow runs
- **THEN** `build`, `lint`, change-classification, and matrix acceptance `test` SHALL be skipped by the preflight gate
- **AND** `Test Validation` SHALL succeed
- **AND** auto-approve SHALL be eligible to run
