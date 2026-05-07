# `ci-openspec-workflow` — Workflow Requirements

Workflow implementation: `.github/workflows/openspec.yml`

## Purpose

Define the OpenSpec validation workflow: run `make check-openspec` on every PR and post-merge push to `main`, and publish a `gate` result as the sole required check.

## ADDED Requirements

### Requirement: Workflow identity and triggers

The workflow name SHALL be `OpenSpec CI`. The workflow file SHALL be `.github/workflows/openspec.yml`. The workflow SHALL run on `push` to the `main` branch only. The workflow SHALL run on `pull_request` events of types `opened`, `synchronize`, and `reopened`. The workflow SHALL support manual execution via `workflow_dispatch`. The workflow SHALL NOT use `paths-ignore` at the trigger level.

#### Scenario: Pull request triggers OpenSpec validation

- **GIVEN** a pull request event of type `opened`, `synchronize`, or `reopened`
- **WHEN** the workflow evaluates its trigger
- **THEN** the `validate` and `gate` jobs SHALL run

#### Scenario: Push to main triggers OpenSpec validation

- **GIVEN** a push to the `main` branch
- **WHEN** the workflow evaluates its trigger
- **THEN** the `validate` and `gate` jobs SHALL run

### Requirement: Validate job

The `validate` job SHALL run on `ubuntu-latest`, set up Node.js (24.x) with npm cache, run `npm ci`, and run `make check-openspec` with `OPENSPEC_TELEMETRY=0`. The `validate` job SHALL always run (no skip condition); it is fast and provides value on all change sets.

#### Scenario: Validate runs on every trigger

- **GIVEN** any workflow trigger (push to main, pull request, or workflow_dispatch)
- **WHEN** the validate job runs
- **THEN** `make check-openspec` SHALL execute and fail the job if validation fails

#### Scenario: OpenSpec validation failure fails job

- **GIVEN** one or more OpenSpec specs fail validation
- **WHEN** `make check-openspec` runs
- **THEN** the validate job SHALL fail

### Requirement: Gate job

The workflow SHALL include a `gate` job that always runs (`if: always()`) and depends on `validate`. The `gate` job SHALL be the sole required check for this workflow (`openspec / gate`). The `gate` job SHALL succeed when `validate` succeeded. The `gate` job SHALL fail when `validate` failed or was cancelled.

#### Scenario: Validation passes — gate succeeds

- **GIVEN** the validate job completes successfully
- **WHEN** the gate evaluates the outcome
- **THEN** `gate` SHALL succeed

#### Scenario: Validation fails — gate fails

- **GIVEN** the validate job fails
- **WHEN** the gate evaluates the outcome
- **THEN** `gate` SHALL fail

### Requirement: Supply chain for actions

Third-party actions in the workflow SHALL be pinned by commit SHA.

#### Scenario: Action references are SHA-pinned

- **GIVEN** a third-party action is used in the workflow
- **WHEN** the workflow YAML is inspected
- **THEN** the action reference SHALL use a commit SHA
