## ADDED Requirements

### Requirement: Version-matrix selector

The auto-approve script SHALL include a `version-matrix` category selector that matches only same-repository pull requests whose head branch name is exactly `acceptance-test-version-matrix`.

#### Scenario: Version-matrix branch matches the category

- **GIVEN** a same-repository pull request whose head branch name is `acceptance-test-version-matrix`
- **WHEN** category matching runs
- **THEN** the `version-matrix` selector SHALL match

#### Scenario: Other branches do not match the category

- **GIVEN** a pull request whose head branch name is not `acceptance-test-version-matrix`
- **WHEN** category matching runs
- **THEN** the `version-matrix` selector SHALL NOT match

### Requirement: Version-matrix commit authors

Every commit in a `version-matrix` category pull request SHALL be authored by `github-actions[bot]`.

#### Scenario: Foreign commit on version-matrix PR

- **GIVEN** a pull request matched as `version-matrix` but a commit author is not `github-actions[bot]`
- **WHEN** gates run
- **THEN** the pull request SHALL NOT be approved via that category

### Requirement: Version-matrix file allowlist

Every changed file path in a `version-matrix` category pull request SHALL be exactly `.github/versions/acceptance-test-matrix.json`.

#### Scenario: Only the pinned versions artifact is allowed

- **GIVEN** a `version-matrix` pull request changes only `.github/versions/acceptance-test-matrix.json`
- **WHEN** gates run
- **THEN** the file-path gate for that category SHALL pass

#### Scenario: Additional file blocks approval

- **GIVEN** a `version-matrix` pull request changes `.github/versions/acceptance-test-matrix.json` and any other file
- **WHEN** gates run
- **THEN** approval SHALL NOT proceed for that category

### Requirement: Version-matrix approval policy

A pull request that matches the `version-matrix` category SHALL be auto-approved when global approval gates pass, including a successful Provider Gate result for that pull request. The `version-matrix` category SHALL NOT impose a diff-threshold gate. This change SHALL NOT add auto-merge behavior for the `version-matrix` category or any other category; approval remains a review action only.

#### Scenario: Version-matrix PR with passing global gates

- **GIVEN** a `version-matrix` pull request matching the category
- **AND** Provider Gate has succeeded for that pull request
- **WHEN** global gates pass
- **THEN** the script SHALL approve

#### Scenario: Version-matrix PR is not auto-merged

- **GIVEN** a `version-matrix` pull request has been auto-approved
- **WHEN** the auto-approve script completes
- **THEN** it SHALL NOT merge, enable auto-merge on, or otherwise change the mergeable state of the pull request
