# `ci-workflows-workflow` — Workflow Requirements

Workflow implementation: `.github/workflows/workflows.yml`

## Purpose

Define the workflow-and-hook test CI: classify whether `.github/` files changed, run `make workflow-test` and `make hook-test` when they have, and publish a `gate` result as the sole required check.

## ADDED Requirements

### Requirement: Workflow identity and triggers

The workflow name SHALL be `Workflows CI`. The workflow file SHALL be `.github/workflows/workflows.yml`. The workflow SHALL run on `push` to the `main` branch only. The workflow SHALL run on `pull_request` events of types `opened`, `synchronize`, and `reopened`. The workflow SHALL support manual execution via `workflow_dispatch`. The workflow SHALL NOT use `paths-ignore` at the trigger level.

#### Scenario: Pull request triggers workflows CI

- **GIVEN** a pull request event of type `opened`, `synchronize`, or `reopened`
- **WHEN** the workflow evaluates its trigger
- **THEN** `classify` and `gate` SHALL always run; `test` SHALL run only when `workflow_changes=true`

#### Scenario: Push to main triggers workflows CI

- **GIVEN** a push to the `main` branch
- **WHEN** the workflow evaluates its trigger
- **THEN** all jobs SHALL run with `workflow_changes=true` unconditionally

### Requirement: Change classification job

The workflow SHALL include a `classify` job that always runs and outputs `workflow_changes`. On `push` to `main` and on `workflow_dispatch`, the job SHALL output `workflow_changes=true` unconditionally. On `pull_request` events, the job SHALL set `workflow_changes=true` when any changed file is under `.github/`; otherwise `workflow_changes=false`.

#### Scenario: GitHub directory change sets workflow_changes=true

- **GIVEN** a PR that changes any file under `.github/`
- **WHEN** the classify job evaluates the file list
- **THEN** it SHALL report `workflow_changes=true`

#### Scenario: No GitHub directory change sets workflow_changes=false

- **GIVEN** a PR whose changed files contain no path under `.github/`
- **WHEN** the classify job evaluates the file list
- **THEN** it SHALL report `workflow_changes=false`

#### Scenario: Push to main always runs workflow tests

- **GIVEN** a push event to the `main` branch
- **WHEN** the classify job runs
- **THEN** it SHALL output `workflow_changes=true` unconditionally

### Requirement: Test job

The `test` job SHALL run on `ubuntu-latest`, depend on the `classify` job, execute only when `classify` outputs `workflow_changes=true`, set up Go from `go.mod`, set up Node.js (24.x), run `make vendor`, run `make workflow-test`, and run `make hook-test`.

#### Scenario: Test runs for workflow changes

- **GIVEN** `classify` reports `workflow_changes=true`
- **WHEN** the test job evaluates its condition
- **THEN** `make workflow-test` and `make hook-test` SHALL run

#### Scenario: Test is skipped for non-workflow changes

- **GIVEN** `classify` reports `workflow_changes=false`
- **WHEN** the test job evaluates its condition
- **THEN** the test job SHALL be skipped

### Requirement: Gate job

The workflow SHALL include a `gate` job that always runs (`if: always()`) and depends on `classify` and `test`. The `gate` job SHALL be the sole required check for this workflow (`workflows / gate`). The `gate` job SHALL succeed when any of the following is true:

- `classify` reported `workflow_changes=false` and `test` was skipped
- `test` completed successfully

The `gate` job SHALL fail when any of the following is true:

- `classify` reported `workflow_changes=true` but `test` was skipped (unexpected skip)
- `test` failed or was cancelled

#### Scenario: Workflow tests pass — gate succeeds

- **GIVEN** `classify` reports `workflow_changes=true` and `test` succeeds
- **WHEN** the gate evaluates the outcome
- **THEN** `gate` SHALL succeed

#### Scenario: No workflow changes — gate succeeds

- **GIVEN** `classify` reports `workflow_changes=false` and `test` is skipped
- **WHEN** the gate evaluates the outcome
- **THEN** `gate` SHALL succeed

#### Scenario: Workflow tests fail — gate fails

- **GIVEN** `classify` reports `workflow_changes=true` and `test` fails
- **WHEN** the gate evaluates the outcome
- **THEN** `gate` SHALL fail

### Requirement: Supply chain for actions

Third-party actions in the workflow SHALL be pinned by commit SHA.

#### Scenario: Action references are SHA-pinned

- **GIVEN** a third-party action is used in the workflow
- **WHEN** the workflow YAML is inspected
- **THEN** the action reference SHALL use a commit SHA
