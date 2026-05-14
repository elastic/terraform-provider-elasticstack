## MODIFIED Requirements

### Requirement: Workflow suppresses duplicate linked pull requests
Before agent activation, the workflow SHALL detect whether an open linked `change-factory` pull request already exists for the triggering issue. A pull request SHALL be treated as linked only when it is open, carries the `change-factory` label, uses the deterministic branch name `change-factory/issue-<issue-number>`, and includes the literal phrase `Related to #<issue-number>` in its body. The workflow SHALL use the literal `'related-literal'` duplicate-linkage mode (matching `Related to #N` only, not GitHub closing keywords). When a duplicate is found, the workflow SHALL post exactly one comment on the triggering issue explaining the skip and linking to the existing PR URL, then skip agent activation. The comment SHALL instruct the maintainer to close or convert the PR to a draft before retrying.

#### Scenario: Existing linked PR prevents a duplicate run and posts a comment
- **WHEN** the workflow finds an open pull request that satisfies the linked `change-factory` PR criteria for the triggering issue (open, `change-factory` label, branch `change-factory/issue-<issue-number>`, and body containing `Related to #<issue-number>`)
- **THEN** the workflow SHALL post one comment on the triggering issue referencing the existing PR and instructing the maintainer to close it before retrying
- **AND** the workflow SHALL skip agent activation instead of opening a duplicate pull request

#### Scenario: Unrelated PR does not block issue intake
- **WHEN** an open pull request mentions the issue or has a similar title but does not satisfy the full linked `change-factory` PR criteria (for example, its body uses `Closes #<issue-number>` rather than `Related to #<issue-number>`, or its branch differs)
- **THEN** the workflow SHALL NOT treat that pull request as the canonical linked PR for duplicate suppression

### Requirement: Agent creates exactly one linked OpenSpec proposal pull request
When the deterministic gate passes, the workflow agent SHALL treat the triggering issue title and body as the authoritative source for requested proposal scope, SHALL work on the deterministic branch `change-factory/issue-<issue-number>`, and SHALL create or update exactly one linked pull request labeled `change-factory` and `no-changelog`. The pull request SHALL contain one active OpenSpec change under `openspec/changes/<change-id>/` with the artifacts required for implementation readiness by the active OpenSpec schema. The pull request body SHALL include the literal phrase `Related to #<issue-number>` to provide deterministic linkage for future duplicate-suppression checks. `Related to` is used rather than a GitHub closing keyword because a `change-factory` pull request only delivers an OpenSpec proposal — the underlying request still needs implementation — so merging the proposal MUST NOT auto-close the source issue.

#### Scenario: Eligible issue creates a linked proposal PR
- **WHEN** the workflow runs for a trusted eligible issue event and no open linked `change-factory` pull request already exists
- **THEN** the agent SHALL create an OpenSpec change proposal on branch `change-factory/issue-<issue-number>`
- **AND** it SHALL open one linked pull request carrying the `change-factory` and `no-changelog` labels

#### Scenario: Pull request metadata preserves deterministic linkage
- **WHEN** the agent creates the `change-factory` pull request
- **THEN** the pull request body SHALL contain the literal phrase `Related to #<issue-number>` so later workflow runs can identify it as the canonical PR for the issue
- **AND** the pull request body SHALL NOT include any GitHub closing keyword for that issue (such as `Closes #<issue-number>` or `Fixes #<issue-number>`) so that merging the proposal does not auto-close the source issue

#### Scenario: Proposal artifacts are implementation-ready
- **WHEN** the agent completes a proposal pull request
- **THEN** the pull request SHALL include all OpenSpec artifacts required before implementation can begin according to the repository's active OpenSpec schema
