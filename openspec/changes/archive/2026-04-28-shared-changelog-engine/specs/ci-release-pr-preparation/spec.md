## MODIFIED Requirements

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

### Requirement: Release PR can be regenerated manually without pull-request-triggered automation
The release preparation workflow SHALL prepare release branches so that the changelog-generation workflow can be rerun manually in explicit release mode for the same target version, without requiring `pull_request_target` or other pull-request-triggered changelog automation.

#### Scenario: Manual release regeneration targets an existing prep-release branch
- **GIVEN** a release-preparation branch for version `X` already exists
- **WHEN** a maintainer manually dispatches the changelog-generation workflow in release mode for version `X`
- **THEN** the workflow SHALL be able to regenerate the concrete release changelog section for that branch without requiring a new pull-request event
