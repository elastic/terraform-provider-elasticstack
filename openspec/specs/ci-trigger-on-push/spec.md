# ci-trigger-on-push Specification

## Purpose
TBD - created by archiving change ci-trigger-on-changelog-push. Update Purpose after archive.
## Requirements
### Requirement: Empty-commit CI trigger after workflow push

When a workflow pushes a commit to a branch using the default `GITHUB_TOKEN`, the push SHALL NOT trigger downstream CI workflows (a GitHub Actions restriction). After each such push, the workflow SHALL push an additional empty commit authenticated with a PAT (`GH_AW_CI_TRIGGER_TOKEN`) to trigger downstream CI and satisfy required status checks on the resulting PR.

The CI trigger push SHALL re-authenticate git by rewriting the remote URL to include the PAT, then push an empty commit with message `chore: trigger CI`. The remote URL SHALL NOT be restored afterward (the Actions runner is ephemeral).

#### Scenario: CI trigger fires after changelog push (release mode)
- **WHEN** the changelog-generation workflow pushes a changelog commit to a release branch using `GITHUB_TOKEN`
- **THEN** the workflow SHALL also push an empty commit with message `chore: trigger CI` re-authenticated with `GH_AW_CI_TRIGGER_TOKEN`
- **AND** the empty-commit push SHALL trigger downstream CI on the release branch

#### Scenario: CI trigger fires after changelog push (unreleased mode)
- **WHEN** the changelog-generation workflow force-pushes a changelog commit to the `generated-changelog` branch using `GITHUB_TOKEN`
- **THEN** the workflow SHALL also force-push an empty commit with message `chore: trigger CI` re-authenticated with `GH_AW_CI_TRIGGER_TOKEN`
- **AND** the empty-commit force-push SHALL trigger downstream CI on the `generated-changelog` branch

#### Scenario: CI trigger fires after release preparation push
- **WHEN** the prep-release workflow pushes a release-preparation commit to a `prep-release-*` branch using `GITHUB_TOKEN`
- **THEN** the workflow SHALL also push an empty commit with message `chore: trigger CI` re-authenticated with `GH_AW_CI_TRIGGER_TOKEN`
- **AND** the empty-commit push SHALL trigger downstream CI on the release branch

### Requirement: Graceful degradation when CI trigger token is absent

If the `GH_AW_CI_TRIGGER_TOKEN` secret is not configured, the workflow SHALL still push the substantive commit successfully and SHALL emit a GitHub Actions warning (`::warning::`) explaining that CI will not be triggered. The workflow SHALL NOT fail when the token is absent.

#### Scenario: Missing token emits warning but does not fail
- **GIVEN** the `GH_AW_CI_TRIGGER_TOKEN` secret is not configured on the repository
- **WHEN** the workflow reaches the CI trigger step
- **THEN** the step SHALL emit a `::warning::` message indicating CI will not be triggered
- **AND** the step SHALL NOT fail the workflow
- **AND** the substantive commit SHALL already have been pushed successfully

#### Scenario: Token is present and CI trigger succeeds
- **GIVEN** the `GH_AW_CI_TRIGGER_TOKEN` secret is configured on the repository
- **WHEN** the workflow reaches the CI trigger step
- **THEN** the step SHALL push the empty commit and the workflow SHALL NOT emit the missing-token warning

