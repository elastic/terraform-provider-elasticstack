## ADDED Requirements

### Requirement: Generated changelog pull requests can reach auto-approve without full CI
The `Build/Lint/Test` workflow SHALL allow same-repository pull requests from branch `generated-changelog` that modify only `CHANGELOG.md` to reach the `auto-approve` job without requiring the full build, lint, change-classification, or matrix acceptance test path to run.

#### Scenario: Generated changelog PR reaches auto-approve path
- **GIVEN** a same-repository pull request from branch `generated-changelog`
- **AND** the pull request changes only `CHANGELOG.md`
- **WHEN** the workflow evaluates its execution path
- **THEN** the workflow SHALL produce a successful path that leaves `auto-approve` eligible to run
- **AND** it SHALL NOT require the full build, lint, and matrix acceptance test jobs for that PR

### Requirement: Changelog-only bypass remains narrowly scoped
The `Build/Lint/Test` workflow SHALL keep the changelog-only bypass narrowly scoped to the generated changelog automation shape. Other changelog-only pull requests SHALL NOT gain the same bypass unless they satisfy the repository-authored generated-changelog conditions.

#### Scenario: Manual changelog-only PR does not inherit generated-changelog bypass
- **GIVEN** a pull request changes only `CHANGELOG.md`
- **AND** its head branch name is not `generated-changelog`
- **WHEN** the workflow evaluates bypass conditions
- **THEN** it SHALL NOT treat that pull request as the generated-changelog special case
