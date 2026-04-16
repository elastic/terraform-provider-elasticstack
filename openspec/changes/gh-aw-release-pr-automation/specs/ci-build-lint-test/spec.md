## ADDED Requirements

### Requirement: Generated changelog pull requests can reach auto-approve without full CI
The `Build/Lint/Test` workflow SHALL allow same-repository pull requests from branch `generated-changelog` that are authored by `github-actions[bot]` and modify only `CHANGELOG.md` to reach the `auto-approve` job without requiring the full build, lint, change-classification, or matrix acceptance test path to run. The skip condition MUST verify all three criteria — branch name, PR author, and file list — in the preflight gate before setting `should_run=false`.

#### Scenario: Generated changelog PR reaches auto-approve path
- **GIVEN** a same-repository pull request from branch `generated-changelog`
- **AND** the PR author is `github-actions[bot]`
- **AND** the pull request changes only `CHANGELOG.md`
- **WHEN** the workflow evaluates its execution path
- **THEN** the workflow SHALL produce a successful path that leaves `auto-approve` eligible to run
- **AND** it SHALL NOT require the full build, lint, and matrix acceptance test jobs for that PR

#### Scenario: Auto-merge is gated on the approval outcome
- **GIVEN** a `generated-changelog` PR
- **WHEN** the `auto-approve` job runs
- **THEN** auto-merge SHALL only be enabled if the auto-approve script determined `ShouldApprove` or `AlreadyApproved` is true (reported via a `GITHUB_OUTPUT` step output)
- **AND** auto-merge SHALL NOT be enabled if the auto-approve gates reject the PR

### Requirement: Changelog-only bypass remains narrowly scoped
The `Build/Lint/Test` workflow SHALL keep the changelog-only bypass narrowly scoped to the generated changelog automation shape. Other changelog-only pull requests SHALL NOT gain the same bypass unless they satisfy all three repository-authored generated-changelog conditions: branch name `generated-changelog`, PR author `github-actions[bot]`, and files limited to `CHANGELOG.md`.

#### Scenario: Manual changelog-only PR does not inherit generated-changelog bypass
- **GIVEN** a pull request changes only `CHANGELOG.md`
- **AND** its head branch name is not `generated-changelog`
- **WHEN** the workflow evaluates bypass conditions
- **THEN** it SHALL NOT treat that pull request as the generated-changelog special case

#### Scenario: Wrong author does not inherit generated-changelog bypass
- **GIVEN** a pull request from branch `generated-changelog` changes only `CHANGELOG.md`
- **AND** the PR author is not `github-actions[bot]`
- **WHEN** the workflow evaluates bypass conditions
- **THEN** it SHALL run full CI rather than skipping to the auto-approve path
