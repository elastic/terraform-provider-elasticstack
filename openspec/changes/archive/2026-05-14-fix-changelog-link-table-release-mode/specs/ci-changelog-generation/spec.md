## ADDED Requirements

### Requirement: Release mode updates the Markdown reference link table
In release mode, after rewriting the `## [x.y.z] - <date>` section body, the changelog generator SHALL also update the Markdown reference link table at the bottom of `CHANGELOG.md`. Specifically it SHALL:

1. Replace the `[Unreleased]:` compare URL from `…compare/vOLD…HEAD` to `…compare/vNEW…HEAD`.
2. Insert a new `[x.y.z]: …compare/vOLD…vNEW` entry immediately after the updated `[Unreleased]:` line.

This update SHALL be idempotent: if the `[x.y.z]:` entry already exists the generator SHALL NOT insert a duplicate. If no `[Unreleased]:` line is present in the file the generator SHALL leave the link table unchanged. If no previous semver tag exists (first-ever release) the generator SHALL leave the link table unchanged.

The link table update SHALL NOT be applied in unreleased mode; only the `## [Unreleased]` section body is rewritten in that mode.

#### Scenario: Release mode inserts new version entry and updates Unreleased link
- **GIVEN** `CHANGELOG.md` contains a `[Unreleased]: …/compare/vOLD…HEAD` line in its link table
- **WHEN** the changelog generator runs in release mode for version `NEW` with previous tag `vOLD`
- **THEN** the resulting `CHANGELOG.md` SHALL contain `[Unreleased]: …/compare/vNEW…HEAD`
- **AND** SHALL contain a new line `[NEW]: …/compare/vOLD…vNEW` immediately after it

#### Scenario: Link table update is idempotent on re-run
- **GIVEN** `CHANGELOG.md` already contains both `[Unreleased]: …/compare/vNEW…HEAD` and `[NEW]: …/compare/vOLD…vNEW`
- **WHEN** the changelog generator runs in release mode again for the same version
- **THEN** the resulting `CHANGELOG.md` SHALL contain exactly one `[NEW]:` entry (no duplicate)

#### Scenario: No-op when Unreleased link line is absent
- **GIVEN** `CHANGELOG.md` has no `[Unreleased]:` line in the link table
- **WHEN** the changelog generator runs in release mode
- **THEN** the link table SHALL be left unchanged

#### Scenario: No-op when no previous tag exists
- **GIVEN** no semver release tag exists in the repository
- **WHEN** the changelog generator runs in release mode
- **THEN** the link table SHALL be left unchanged

#### Scenario: Unreleased mode does not touch the link table
- **GIVEN** `CHANGELOG.md` contains a `[Unreleased]: …/compare/vOLD…HEAD` line
- **WHEN** the changelog generator runs in unreleased mode
- **THEN** the `[Unreleased]:` line in the link table SHALL remain unchanged

### Requirement: Changelog-generation PRs are labelled `no-changelog`
The changelog-generation workflow SHALL apply the `no-changelog` label to every PR it creates or manages, so that subsequent changelog runs do not attempt to parse those PRs as feature PRs and fail.

Specifically:
- When `manageUnreleasedPR` creates a new `generated-changelog` PR it SHALL add the `no-changelog` label immediately after creation.
- When `manageUnreleasedPR` updates an existing `generated-changelog` PR it SHALL also ensure the `no-changelog` label is applied (idempotent — the GitHub API `addLabels` call is safe to repeat).
- When `refreshReleasePR` locates and updates a release prep PR it SHALL apply the `no-changelog` label to that PR.

#### Scenario: New generated-changelog PR receives no-changelog label
- **WHEN** `manageUnreleasedPR` creates a new PR from `generated-changelog` to `main`
- **THEN** the `no-changelog` label SHALL be applied to that PR before the function returns

#### Scenario: Existing generated-changelog PR receives no-changelog label on update
- **GIVEN** the `generated-changelog` PR already exists
- **WHEN** `manageUnreleasedPR` updates the PR body on a subsequent run
- **THEN** the `no-changelog` label SHALL be applied (or confirmed present) on that PR

#### Scenario: Release prep PR receives no-changelog label
- **GIVEN** an open PR exists for `prep-release-x.y.z` → `main`
- **WHEN** `refreshReleasePR` locates and refreshes that PR
- **THEN** the `no-changelog` label SHALL be applied to the PR
