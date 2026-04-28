# `ci-release-pr-preparation` — Deterministic release PR preparation workflow

Workflow implementation: repository-authored template under `.github/workflows-src/prep-release/workflow.yml.tmpl`, generating checked-in `.github/workflows/prep-release.yml` via `scripts/compile-workflow-sources/main.go`.

"Repository-authored" means the workflow uses standard GitHub Actions tooling - shell scripts, `git`, and the `gh` CLI - rather than agentic (LLM-driven) execution. Using `gh pr create`/`gh pr list` for PR management and configuring a bot git identity (`github-actions[bot]`) for commits are the expected implementation patterns in this context.

## Purpose

Define a deterministic workflow that prepares provider release pull requests by computing the target version, creating or updating a release branch, applying the version bump and regenerating the release changelog section via the shared JavaScript changelog engine, and creating or reusing the release PR.
## Requirements
### Requirement: Workflow trigger and inputs

The release preparation workflow SHALL run only from `workflow_dispatch` and SHALL accept a bump-mode input that supports `patch`, `minor`, and `major`, defaulting to `patch`.

#### Scenario: Workflow dispatch defaults to patch

- **WHEN** a maintainer dispatches the release preparation workflow without overriding the bump mode
- **THEN** the workflow SHALL compute the next release version using the `patch` increment

### Requirement: Deterministic release version and range discovery

Before mutating repository state, deterministic repository-authored steps SHALL identify the previous semver release tag on `main`, compute the target release version from the dispatch bump input, and derive the deterministic release branch and pull-request title from that version.

#### Scenario: Patch release selects the next patch version

- **GIVEN** the latest release tag on `main` is `v0.14.3`
- **WHEN** the workflow is dispatched with bump mode `patch`
- **THEN** the target release version SHALL be `0.14.4`

#### Scenario: Release branch and title derive from the target version

- **GIVEN** the previous release tag is `v0.14.3`
- **WHEN** deterministic pre-activation computes the release target
- **THEN** the workflow SHALL derive a stable branch name such as `prep-release-0.14.4`
- **AND** it SHALL derive a stable pull-request title such as `Prepare 0.14.4 release`

### Requirement: Release preparation changes are limited to deterministic version bump plumbing
The release preparation workflow SHALL apply only the deterministic release-preparation changes owned by that workflow. It SHALL update the top-level provider `VERSION` in `Makefile` to the target version, and it SHALL invoke the shared deterministic changelog engine in release mode to regenerate the concrete release section in `CHANGELOG.md` before opening or reusing the release pull request. The workflow SHALL NOT perform agentic changelog synthesis.

#### Scenario: Release-preparation branch includes version bump and final changelog update
- **WHEN** the workflow prepares a release branch for version `X`
- **THEN** the branch SHALL contain the deterministic version bump changes owned by the workflow
- **AND** it SHALL contain the regenerated concrete changelog section for version `X` before the release PR is created or reused

### Requirement: Release branch and pull request are managed deterministically
After applying its deterministic changes, the workflow SHALL create or update a deterministic release branch named from the target version, SHALL commit the release-preparation changes with a stable release-preparation commit message, SHALL push that branch, and SHALL create or reuse a pull request targeting `main` with a stable release-preparation title.

#### Scenario: Existing release PR is reused on rerun
- **GIVEN** a release pull request for the target version already exists
- **WHEN** the workflow is rerun for the same version
- **THEN** the workflow SHALL update or reuse that release branch and pull request rather than opening a duplicate

#### Scenario: Release preparation uses a single deterministic commit
- **WHEN** the workflow prepares a release branch for version `X`
- **THEN** it SHALL combine the deterministic version bump and release changelog update into a single release-preparation commit before pushing the branch

### Requirement: Release PR carries the no-changelog label

The release preparation workflow SHALL apply the `no-changelog` label to the release PR. This label SHALL be present whether the PR is newly created or already exists (reused on a rerun). The `no-changelog` label is assumed to exist in the repository as a pre-condition.

#### Scenario: New PR is created with no-changelog label

- **WHEN** the workflow creates a new release PR
- **THEN** the PR SHALL have the `no-changelog` label applied at creation time

#### Scenario: Existing PR is labelled on reuse

- **GIVEN** a release PR for the target version already exists
- **WHEN** the workflow reruns and reuses that existing PR
- **THEN** the workflow SHALL apply the `no-changelog` label to the existing PR
- **AND** the label application SHALL be idempotent (re-applying an already-present label SHALL NOT cause an error)

### Requirement: Release PR can be regenerated manually without pull-request-triggered automation
The release preparation workflow SHALL prepare release branches so that the changelog-generation workflow can be rerun manually in explicit release mode for the same target version, without requiring `pull_request_target` or other pull-request-triggered changelog automation.

#### Scenario: Manual release regeneration targets an existing prep-release branch
- **GIVEN** a release-preparation branch for version `X` already exists
- **WHEN** a maintainer manually dispatches the changelog-generation workflow in release mode for version `X`
- **THEN** the workflow SHALL be able to regenerate the concrete release changelog section for that branch without requiring a new pull-request event

