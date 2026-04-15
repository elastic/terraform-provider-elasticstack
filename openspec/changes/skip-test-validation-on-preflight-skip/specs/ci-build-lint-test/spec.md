## MODIFIED Requirements

### Requirement: Test validation job (REQ-034–REQ-036)

The workflow SHALL publish a `Test Validation` job that evaluates the change-classification output and the matrix acceptance job result whenever the preflight gate permits downstream CI execution. When the preflight gate outputs `should_run=false`, the `Test Validation` job SHALL be intentionally skipped.

When the preflight gate allows downstream execution, the `Test Validation` job SHALL succeed when any of the following is true:

* The change-classification job reports `provider_changes=false` and the matrix acceptance `test` job is intentionally skipped
* The change-classification job reports `provider_changes=false` and the matrix acceptance `test` job completes successfully
* The change-classification job reports `provider_changes=true` and the matrix acceptance `test` job completes successfully

When the preflight gate allows downstream execution, the `Test Validation` job SHALL fail if either of the following is true:

* The change-classification job reports `provider_changes=true` and the matrix acceptance `test` job does not complete successfully
* The change-classification job reports `provider_changes=false` and the matrix acceptance `test` job still runs but does not complete successfully

The validation job SHALL provide a stable required-check target for workflow runs where the preflight gate allows downstream execution, so GitHub branch protection or rulesets can require one normalized acceptance signal instead of per-version matrix checks.

#### Scenario: OpenSpec-only pull request

- **GIVEN** a pull request whose changed files are all under `openspec/`
- **AND** the preflight gate outputs `should_run=true`
- **WHEN** the workflow reaches `Test Validation`
- **THEN** the matrix acceptance `test` job SHALL be treated as intentionally skipped
- **AND** `Test Validation` SHALL succeed

#### Scenario: Preflight-disabled workflow run

- **GIVEN** a workflow run where the preflight gate outputs `should_run=false`
- **WHEN** downstream jobs evaluate their execution conditions
- **THEN** the `Test Validation` job SHALL be skipped

#### Scenario: Provider change with failing acceptance coverage

- **GIVEN** a workflow run with `provider_changes=true`
- **AND** the preflight gate outputs `should_run=true`
- **AND** the matrix acceptance `test` job does not complete successfully
- **WHEN** `Test Validation` evaluates the workflow state
- **THEN** `Test Validation` SHALL fail

### Requirement: Auto-approve job (REQ-018–REQ-021)

The `auto-approve` job SHALL depend on the `Test Validation` job and SHALL only run on `pull_request` events. For non-`ready_for_review` events, `auto-approve` SHALL require `Test Validation` to succeed before it runs. For `ready_for_review` events, `auto-approve` SHALL be eligible to run regardless of `Test Validation`'s result, because the preflight gate intentionally skips the downstream CI path on that event. The `auto-approve` job SHALL execute `go run ./scripts/auto-approve`; approval policy and gate behavior are defined in [`openspec/specs/ci-pr-auto-approve/spec.md`](../ci-pr-auto-approve/spec.md). The `auto-approve` job SHALL request `contents: read` and `pull-requests: write` permissions.

#### Scenario: Auto-approve after satisfied validation

- **GIVEN** a pull request workflow and successful `Test Validation`
- **WHEN** auto-approve runs
- **THEN** it SHALL invoke `go run ./scripts/auto-approve` with the specified permissions

#### Scenario: Ready-for-review bypasses validation result

- **GIVEN** a `pull_request` workflow with action `ready_for_review`
- **WHEN** `auto-approve` evaluates its execution conditions
- **THEN** it SHALL remain eligible to run regardless of whether `Test Validation` succeeded or was skipped

### Requirement: Ready-for-review behavior (REQ-030)

On `ready_for_review` `pull_request` events, the workflow SHALL keep the preflight gate behavior that prevents the `build`, `lint`, change-classification, matrix acceptance `test`, and `Test Validation` jobs from running. `auto-approve` SHALL remain eligible to run on that path.

#### Scenario: Ready for review event

- **GIVEN** a `pull_request` with action `ready_for_review`
- **WHEN** the workflow runs
- **THEN** `build`, `lint`, change-classification, matrix acceptance `test`, and `Test Validation` SHALL be skipped by the preflight gate
- **AND** auto-approve SHALL be eligible to run
