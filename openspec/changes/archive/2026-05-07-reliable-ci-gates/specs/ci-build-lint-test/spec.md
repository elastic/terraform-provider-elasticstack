## MODIFIED Requirements

### Requirement: Workflow identity and triggers (REQ-001–REQ-006)

The workflow name SHALL be `Provider CI`. The workflow file SHALL be `.github/workflows/provider.yml`. The workflow SHALL run on `push` to the `main` branch only. The workflow SHALL run on `pull_request` events of types `opened`, `synchronize`, and `reopened`. The workflow SHALL support manual execution via `workflow_dispatch`. The workflow SHALL NOT use `paths-ignore` at the trigger level; change filtering is handled by the change-classification job.

#### Scenario: Push to main triggers workflow

- **GIVEN** a push to the `main` branch
- **WHEN** the workflow evaluates its trigger
- **THEN** the `classify` job SHALL run and output `provider_changes=true` unconditionally, and all downstream jobs SHALL run

#### Scenario: Pull request opened, synchronised, or reopened triggers workflow

- **GIVEN** a pull request event of type `opened`, `synchronize`, or `reopened`
- **WHEN** the workflow evaluates its trigger
- **THEN** the workflow SHALL run starting with the classify job

#### Scenario: ready_for_review does not trigger workflow

- **GIVEN** a pull request is converted from draft to ready for review
- **WHEN** GitHub evaluates workflow triggers
- **THEN** the `provider.yml` workflow SHALL NOT be triggered
- **AND** the existing check results for the HEAD commit SHALL remain valid

#### Scenario: Push to non-main branch does not trigger workflow

- **GIVEN** a push to any branch other than `main`
- **WHEN** GitHub evaluates workflow triggers
- **THEN** the `provider.yml` workflow SHALL NOT be triggered automatically

### Requirement: Build job (REQ-007)

The `build` job SHALL run on `ubuntu-latest`, depend on the `classify` job, execute only when `classify` outputs `provider_changes=true`, set up Go from `go.mod`, set up Node.js (24.x), run `make vendor`, and run `make build-ci`. The `build` job SHALL NOT run `make workflow-test` or `make hook-test`; those steps belong in `workflows.yml`.

#### Scenario: Build runs for provider changes

- **GIVEN** the classify job reports `provider_changes=true`
- **WHEN** the build job evaluates its condition
- **THEN** `make vendor` and `make build-ci` SHALL run

#### Scenario: Build is skipped for non-provider changes

- **GIVEN** the classify job reports `provider_changes=false`
- **WHEN** the build job evaluates its condition
- **THEN** the build job SHALL be skipped

### Requirement: Lint job (REQ-008, REQ-031)

The `lint` job SHALL run on `ubuntu-latest`, depend on the `classify` job, execute only when `classify` outputs `provider_changes=true`, set up Go from `go.mod`, read the Terraform CLI version from `.terraform-version`, set up Terraform without wrapper mode, install Node.js (24.x) with npm cache, run `npm ci`, and run `make check-lint` (which includes OpenSpec validation).

#### Scenario: Lint runs for provider changes

- **GIVEN** the classify job reports `provider_changes=true`
- **WHEN** the lint job evaluates its condition
- **THEN** `npm ci` and `make check-lint` SHALL run

#### Scenario: Lint is skipped for non-provider changes

- **GIVEN** the classify job reports `provider_changes=false`
- **WHEN** the lint job evaluates its condition
- **THEN** the lint job SHALL be skipped

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job and the `classify` job. The acceptance test job SHALL execute only when `classify` outputs `provider_changes=true`. The job SHALL NOT depend on a `preflight` job. All other requirements regarding the matrix, stack versions, environment variables, Fleet setup, snapshot handling, disk-space cleanup, and teardown remain unchanged.

#### Scenario: Provider change runs stack and tests

- **GIVEN** `classify` reports `provider_changes=true`
- **WHEN** the test job evaluates its execution condition
- **THEN** the stack SHALL be provisioned and `make testacc` SHALL run per the existing matrix and shard policy

#### Scenario: Non-provider change skips matrix acceptance

- **GIVEN** `classify` reports `provider_changes=false`
- **WHEN** the acceptance test job evaluates its condition
- **THEN** the matrix acceptance `test` job SHALL be skipped

### Requirement: Auto-approve job

The `auto-approve` job SHALL depend on the `gate` job and SHALL only run on `pull_request` events. It SHALL run only when the `gate` job succeeds. The `auto-approve` job SHALL execute `go run ./scripts/auto-approve`; approval policy is defined in [`openspec/specs/ci-pr-auto-approve/spec.md`](../../../../../../../openspec/specs/ci-pr-auto-approve/spec.md). The job SHALL NOT include a step to enable auto-merge for the `generated-changelog` branch.

#### Scenario: Auto-approve after gate success

- **GIVEN** a pull request workflow run where `gate` succeeds
- **WHEN** `auto-approve` evaluates its condition
- **THEN** it SHALL invoke `go run ./scripts/auto-approve` with `contents: read` and `pull-requests: write` permissions

#### Scenario: Auto-approve does not run when gate fails

- **GIVEN** a pull request workflow run where `gate` fails
- **WHEN** `auto-approve` evaluates its condition
- **THEN** `auto-approve` SHALL NOT run

### Requirement: Change classification job

The workflow SHALL include a `classify` job that always runs (no upstream dependencies) and outputs `provider_changes`. On `push` to `main` and on `workflow_dispatch`, the job SHALL output `provider_changes=true` unconditionally. On `pull_request` events, the job SHALL compute `provider_changes` from the PR file list: `provider_changes=false` when all changed files match the non-impacting path set; `provider_changes=true` otherwise.

The non-impacting path set is:
- `CHANGELOG.md`
- Any path under `openspec/`
- Any path under `.agents/`
- Any path under `.github/` **except** `.github/workflows/provider.yml`

Any change set containing a path outside this set SHALL produce `provider_changes=true`. An empty file list (e.g., force-push with no commit payload) SHALL default to `provider_changes=true`.

#### Scenario: Changelog-only pull request skips provider CI

- **GIVEN** a PR whose changed files are limited to `CHANGELOG.md`
- **WHEN** the classify job evaluates the file list
- **THEN** it SHALL report `provider_changes=false`

#### Scenario: Agents-only pull request skips provider CI

- **GIVEN** a PR whose changed files are all under `.agents/`
- **WHEN** the classify job evaluates the file list
- **THEN** it SHALL report `provider_changes=false`

#### Scenario: Non-test GitHub config change skips provider CI

- **GIVEN** a PR whose changed files are all under `.github/` and none are `.github/workflows/provider.yml`
- **WHEN** the classify job evaluates the file list
- **THEN** it SHALL report `provider_changes=false`

#### Scenario: Provider workflow file change runs provider CI

- **GIVEN** a PR that changes `.github/workflows/provider.yml`
- **WHEN** the classify job evaluates the file list
- **THEN** it SHALL report `provider_changes=true`

#### Scenario: OpenSpec-only pull request skips provider CI

- **GIVEN** a PR whose changed files are all under `openspec/`
- **WHEN** the classify job evaluates the file list
- **THEN** it SHALL report `provider_changes=false`

#### Scenario: Mixed change runs provider CI

- **GIVEN** a PR that changes a Go source file alongside `CHANGELOG.md`
- **WHEN** the classify job evaluates the file list
- **THEN** it SHALL report `provider_changes=true`

#### Scenario: Push to main always runs provider CI

- **GIVEN** a push event to the `main` branch
- **WHEN** the classify job runs
- **THEN** it SHALL output `provider_changes=true` unconditionally

### Requirement: Gate job

The workflow SHALL include a `gate` job that always runs (`if: always()`) and depends on `classify`, `build`, `lint`, and `test`. The `gate` job SHALL be the sole required check for this workflow (`provider / gate`). The `gate` job SHALL succeed when any of the following is true:

- `classify` reported `provider_changes=false` and `build`, `lint`, and `test` were all skipped
- `build`, `lint`, and `test` all completed successfully

The `gate` job SHALL fail when any of the following is true:

- `classify` reported `provider_changes=true` but any of `build`, `lint`, or `test` was skipped (unexpected skip)
- Any of `build`, `lint`, or `test` failed or was cancelled

#### Scenario: All jobs pass for provider change

- **GIVEN** `classify` reports `provider_changes=true` and `build`, `lint`, `test` all succeed
- **WHEN** the gate evaluates outcomes
- **THEN** `gate` SHALL succeed

#### Scenario: Non-provider change — all jobs skipped

- **GIVEN** `classify` reports `provider_changes=false` and `build`, `lint`, `test` are all skipped
- **WHEN** the gate evaluates outcomes
- **THEN** `gate` SHALL succeed

#### Scenario: Acceptance tests fail

- **GIVEN** `classify` reports `provider_changes=true` and the `test` matrix has at least one failure
- **WHEN** the gate evaluates outcomes
- **THEN** `gate` SHALL fail

#### Scenario: Unexpected skip fails gate

- **GIVEN** `classify` reports `provider_changes=true` but `build` was skipped
- **WHEN** the gate evaluates outcomes
- **THEN** `gate` SHALL fail

## REMOVED Requirements

### Requirement: Preflight gate (REQ-023–REQ-027)

**Reason**: The `preflight` job existed to suppress duplicate CI runs when both `push` and `pull_request` events fired for the same commit, and to short-circuit the generated-changelog bot PR path. With `push` restricted to `main` and `ready_for_review` removed from PR types, the duplicate-run scenario no longer exists. With the generated-changelog path removed, the short-circuit path no longer exists. The job has no remaining purpose.

**Migration**: All `should_run` output references in downstream jobs are replaced by direct `classify` output checks or `if: always()` gate logic.

### Requirement: Ready-for-review behavior (REQ-030)

**Reason**: `ready_for_review` is removed from the workflow's `pull_request` trigger types. When a PR is un-drafted, no new workflow run fires; existing check results on the HEAD commit remain valid. The preflight suppression that caused required checks to enter a "skipped" state is no longer needed.

**Migration**: No downstream change required. Draft PRs accumulate check results via `synchronize` events as commits are pushed.

### Requirement: Test validation job (REQ-034–REQ-036)

**Reason**: Replaced by the `gate` job, which covers `build`, `lint`, and `test` outcomes rather than only `test`. The `gate` job is the required check for this workflow.

**Migration**: Update branch protection required checks from `Build/Lint/Test / Test Validation` to `provider / gate`.

### Requirement: Generated changelog pull requests can reach auto-approve without full CI

**Reason**: The `generated-changelog` special case in `preflight` is removed along with the entire `preflight` job. Changelog-only PRs are now handled uniformly by `classify` (which reports `provider_changes=false` for CHANGELOG-only changes), and the gate succeeds without running any CI jobs.

**Migration**: The `generated-changelog` bot PR no longer gets an explicit fast-path. It goes through `classify`, skips all provider CI, and the gate succeeds. The auto-merge step is removed; the generated-changelog PR merges via the normal auto-approve path for dependabot/copilot if it matches those categories, or requires a manual merge otherwise.

### Requirement: Changelog-only bypass remains narrowly scoped

**Reason**: The narrowly-scoped bypass no longer exists; it is replaced by the general-purpose classify skip-path that applies to all changelog-only PRs regardless of branch name or author.

**Migration**: None required.
