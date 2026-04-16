## ADDED Requirements

### Requirement: Generated changelog selector
The auto-approve script SHALL include a `generated-changelog` category selector that matches only same-repository pull requests whose head branch name is exactly `generated-changelog`.

#### Scenario: Generated changelog branch matches the category
- **GIVEN** a same-repository pull request whose head branch name is `generated-changelog`
- **WHEN** category matching runs
- **THEN** the `generated-changelog` selector SHALL match

#### Scenario: Other branches do not match the category
- **GIVEN** a pull request whose head branch name is not `generated-changelog`
- **WHEN** category matching runs
- **THEN** the `generated-changelog` selector SHALL NOT match

### Requirement: Generated changelog commit authors
Every commit in a `generated-changelog` category pull request SHALL be authored by `github-actions[bot]`.

#### Scenario: Foreign commit on generated changelog PR
- **GIVEN** a pull request matched as `generated-changelog` but a commit author is not `github-actions[bot]`
- **WHEN** gates run
- **THEN** the pull request SHALL NOT be approved via that category

### Requirement: Generated changelog file allowlist
Every changed file path in a `generated-changelog` category pull request SHALL be exactly `CHANGELOG.md`.

#### Scenario: Only changelog file is allowed
- **GIVEN** a `generated-changelog` pull request changes only `CHANGELOG.md`
- **WHEN** gates run
- **THEN** the file-path gate for that category SHALL pass

#### Scenario: Additional file blocks approval
- **GIVEN** a `generated-changelog` pull request changes `CHANGELOG.md` and any other file
- **WHEN** gates run
- **THEN** approval SHALL NOT proceed for that category
