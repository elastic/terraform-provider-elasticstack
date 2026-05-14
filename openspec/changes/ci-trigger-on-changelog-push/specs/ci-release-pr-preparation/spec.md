## MODIFIED Requirements

### Requirement: Release branch and pull request are managed deterministically
After applying its deterministic changes, the workflow SHALL create or update a deterministic release branch named from the target version, SHALL commit the release-preparation changes with a stable release-preparation commit message, SHALL push that branch, and SHALL create or reuse a pull request targeting `main` with a stable release-preparation title. After pushing the release branch, the workflow SHALL push an empty commit re-authenticated with the CI trigger token to trigger downstream CI on the release branch.

#### Scenario: Existing release PR is reused on rerun
- **GIVEN** a release pull request for the target version already exists
- **WHEN** the workflow is rerun for the same version
- **THEN** the workflow SHALL update or reuse that release branch and pull request rather than opening a duplicate
- **AND** it SHALL push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` to trigger CI

#### Scenario: Release preparation uses a single deterministic commit
- **WHEN** the workflow prepares a release branch for version `X`
- **THEN** it SHALL combine the deterministic version bump and release changelog update into a single release-preparation commit before pushing the branch
- **AND** after pushing, it SHALL push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` to trigger CI
