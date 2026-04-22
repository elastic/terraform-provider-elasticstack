# `ci-release-pr-preparation` — Deterministic release PR preparation workflow

Workflow implementation: repository-authored GitHub Actions workflow under `.github/workflows/`.

"Repository-authored" means the workflow uses standard GitHub Actions tooling - shell scripts, `git`, and the `gh` CLI - rather than agentic (LLM-driven) execution. Using `gh pr create`/`gh pr list` for PR management and configuring a bot git identity (`github-actions[bot]`) for commits are the expected implementation patterns in this context.

## Purpose
Define a deterministic workflow that prepares provider release pull requests by computing the target version, creating or updating a release branch, applying the simple version bump changes, and creating or reusing the release PR. Changelog generation is handled separately by `ci-changelog-generation`.
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
The release preparation workflow SHALL apply only the deterministic release-preparation changes owned by that workflow. It SHALL update the top-level provider `VERSION` in `Makefile` to the target version, and it SHALL NOT perform agentic changelog synthesis itself.

#### Scenario: Release-preparation branch starts with version bump changes
- **WHEN** the workflow prepares a release branch for version `X`
- **THEN** the branch SHALL contain the deterministic version bump changes owned by the workflow
- **AND** changelog content generation SHALL remain delegated to `ci-changelog-generation`

### Requirement: Release branch and pull request are managed deterministically
After applying its deterministic changes, the workflow SHALL create or update a deterministic release branch named from the target version, SHALL commit the release-preparation changes with a stable release-preparation commit message, SHALL push that branch, and SHALL create or reuse a pull request targeting `main` with a stable release-preparation title.

#### Scenario: Release branch name is derived from the target version
- **WHEN** the target release version is `0.14.4`
- **THEN** the workflow SHALL use a deterministic branch name derived from that version, such as `prep-release-0.14.4`

#### Scenario: Existing release PR is reused on rerun
- **GIVEN** a release pull request for the target version already exists
- **WHEN** the workflow is rerun for the same version
- **THEN** the workflow SHALL update or reuse that release branch and pull request rather than opening a duplicate

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

### Requirement: Release PR triggers changelog regeneration for the release section
The release preparation workflow SHALL create the release branch and pull request in a way that allows `ci-changelog-generation` to recognize the branch as a release-preparation PR and regenerate the concrete `## [x.y.z] - <date>` changelog section on that branch.

#### Scenario: Release PR branch matches changelog generator contract
- **WHEN** the release preparation workflow creates a release branch
- **THEN** that branch name SHALL match the naming pattern expected by the changelog generator's pull-request mode

