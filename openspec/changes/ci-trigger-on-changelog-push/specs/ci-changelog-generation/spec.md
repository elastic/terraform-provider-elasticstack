## MODIFIED Requirements

### Requirement: Scheduled/manual mode updates the singleton generated changelog PR
In scheduled or manually dispatched unreleased mode, after deterministic validation succeeds, the workflow SHALL rewrite only the `## [Unreleased]` section of `CHANGELOG.md` and SHALL push the result to the singleton branch named `generated-changelog`. The workflow SHALL use repository-authored GitHub Actions logic to look up an existing pull request from `generated-changelog` to `main`, create that pull request when none exists, and update the existing PR body when one already exists. After pushing the changelog commit to `generated-changelog`, the workflow SHALL push an empty commit re-authenticated with the CI trigger token to trigger downstream CI.

#### Scenario: Generated changelog branch PR is reused
- **GIVEN** the singleton branch `generated-changelog` already has an open pull request to `main`
- **WHEN** the scheduled/manual unreleased changelog generator runs again
- **THEN** it SHALL update that same branch and refresh the existing pull request instead of opening a duplicate
- **AND** it SHALL push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` to trigger CI

#### Scenario: Missing generated changelog PR is created
- **GIVEN** the singleton branch `generated-changelog` has no open pull request to `main`
- **WHEN** the scheduled/manual unreleased changelog generator produces an updated `CHANGELOG.md`
- **THEN** the workflow SHALL create the pull request after pushing the branch update
- **AND** it SHALL push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` to trigger CI

### Requirement: Explicit release mode updates the targeted release section and removes Unreleased
In explicit release mode, after deterministic validation succeeds, repository-authored helper logic SHALL update only the concrete `## [x.y.z] - <date>` section for the checked out release branch and SHALL push that change only to the targeted release branch. After pushing the changelog commit to the release branch, the workflow SHALL push an empty commit re-authenticated with the CI trigger token to trigger downstream CI. Manual release-mode execution MAY refresh release PR metadata when the corresponding pull request is known, but release-mode changelog generation SHALL NOT depend on `pull_request_target` event metadata or automatic pull-request triggers.

In release mode, when the rewriter mutates `CHANGELOG.md` to emit the new `## [x.y.z] - <date>` section, it SHALL also remove any existing `## [Unreleased]` section (header and body) from the file. This SHALL hold both on the first run against a release-preparation branch (when no `## [x.y.z]` heading exists yet) and on any re-run (when the `## [x.y.z]` heading is already present alongside a stale `## [Unreleased]` section). Release-mode mutation SHALL NOT preserve, duplicate, or insert content alongside the Unreleased section; the resulting `CHANGELOG.md` SHALL contain exactly one block representing the work shipped in the release, headed by `## [x.y.z] - <date>`.

#### Scenario: Release mode updates only the targeted branch
- **WHEN** the changelog generator runs in explicit release mode for a release-preparation branch
- **THEN** it SHALL push changelog updates only to that targeted release branch
- **AND** it SHALL push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` to trigger CI

#### Scenario: Release mode does not regenerate Unreleased on the release branch
- **WHEN** the changelog generator runs in explicit release mode
- **THEN** it SHALL regenerate the concrete release section needed for that branch and SHALL NOT preserve or regenerate any `## [Unreleased]` section, without treating the branch as the singleton `Unreleased` maintenance branch

#### Scenario: Release mode replaces the Unreleased section with the new versioned section
- **GIVEN** `CHANGELOG.md` on a `prep-release-x.y.z` branch contains a `## [Unreleased]` section with body content and no `## [x.y.z]` heading
- **WHEN** the changelog generator runs in explicit release mode for that branch
- **THEN** the resulting `CHANGELOG.md` SHALL contain a single `## [x.y.z] - <date>` section in place of the previous `## [Unreleased]` section, with no `## [Unreleased]` heading remaining in the file

#### Scenario: Release mode re-run collapses lingering Unreleased section
- **GIVEN** `CHANGELOG.md` on a `prep-release-x.y.z` branch already contains a `## [x.y.z] - <date>` section and also contains a `## [Unreleased]` section
- **WHEN** the changelog generator runs in explicit release mode again for that branch
- **THEN** the resulting `CHANGELOG.md` SHALL contain a single regenerated `## [x.y.z] - <date>` section and SHALL NOT contain any `## [Unreleased]` heading

#### Scenario: Release mode with no prior Unreleased section prepends the new section
- **GIVEN** `CHANGELOG.md` contains no `## [Unreleased]` heading and no `## [x.y.z]` heading
- **WHEN** the changelog generator runs in explicit release mode for version `x.y.z`
- **THEN** the resulting `CHANGELOG.md` SHALL begin with the new `## [x.y.z] - <date>` section followed by the prior changelog content
