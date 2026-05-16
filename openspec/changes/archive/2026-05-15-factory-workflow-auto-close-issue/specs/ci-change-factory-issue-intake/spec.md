## MODIFIED Requirements

### Requirement: Agent creates exactly one linked proposal pull request
When deterministic gates pass, the agent SHALL create exactly one linked `change-factory` pull request on branch `change-factory/issue-{n}` for the triggering issue. The PR SHALL carry the `change-factory` and `no-changelog` labels, SHALL be the only open `change-factory` PR for that issue, and SHALL contain the OpenSpec proposal artifacts produced by the run.

#### Scenario: Eligible issue creates a linked proposal PR
- **WHEN** the workflow runs for a trusted eligible issue event and no open linked `change-factory` pull request already exists
- **THEN** the agent SHALL create an OpenSpec change proposal on branch `change-factory/issue-<issue-number>`
- **AND** it SHALL open one linked pull request carrying the `change-factory` and `no-changelog` labels

#### Scenario: Pull request metadata preserves deterministic linkage
- **WHEN** the agent creates the `change-factory` pull request
- **THEN** the pull request body SHALL contain the literal phrase `Related to #<issue-number>` so later workflow runs can identify it as the canonical PR for the issue
- **AND** the pull request body SHALL NOT include any GitHub closing keyword for that issue (such as `Closes #<issue-number>` or `Fixes #<issue-number>`) so that merging the proposal does not auto-close the source issue

#### Scenario: Safe-output configuration prevents automatic closing references
- **WHEN** maintainers inspect the authored `change-factory` workflow safe-output configuration
- **THEN** `safe-outputs.create-pull-request.auto-close-issue` SHALL be set to `false`
- **AND** generated workflow artifacts derived from that source SHALL preserve the same non-closing PR policy

#### Scenario: Proposal artifacts are implementation-ready
- **WHEN** the agent completes a proposal pull request
- **THEN** the pull request SHALL include all OpenSpec artifacts required before implementation can begin according to the repository's active OpenSpec schema
