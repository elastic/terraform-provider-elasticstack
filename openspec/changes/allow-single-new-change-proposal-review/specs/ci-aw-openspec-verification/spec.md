## MODIFIED Requirements

### Requirement: Discover active change id from PR files (REQ-005)
The workflow SHALL use a deterministic pre-activation step to load the pull request changed files list, including each file entry's status (`added`, `modified`, `removed`, `renamed`, and so on). It SHALL consider only paths matching `openspec/changes/<id>/...` where `<id>` is a single path segment and `archive` is not the first segment (that is, exclude `openspec/changes/archive/**`). For each such path, it SHALL record the status of that file entry and SHALL publish pre-activation outputs that include the gate result, the selected active change id when selection succeeds, and a deterministic review disposition that distinguishes approval-eligible modified-only changes from comment-only net-new change proposals.

#### Scenario: Derive change id from path
- **GIVEN** a modified file `openspec/changes/my-feature/tasks.md`
- **WHEN** the deterministic selection step parses paths
- **THEN** the active change id SHALL be recognized as `my-feature`

#### Scenario: Selected change and review disposition are exposed to the agent
- **GIVEN** exactly one active change satisfies the gating rules
- **WHEN** the deterministic selection step completes
- **THEN** the workflow SHALL expose that change id and the deterministic review disposition as pre-activation outputs for the later agent job

### Requirement: Noop when change selection rules fail (REQ-006)
The workflow SHALL not submit a pull request review and SHALL not archive when the deterministic gating result indicates any of the following:

1. More than one distinct `<id>` appears among paths under `openspec/changes/<id>/` (non-archive).
2. Zero distinct `<id>` appears among paths under `openspec/changes/<id>/` (non-archive).
3. Any file under the selected `openspec/changes/<id>/` tree has a status other than `added` or `modified` among the set the workflow cares about (for example, `removed` or `renamed`).

When exactly one active change id is present and every relevant file status is `modified`, the workflow SHALL select that change and mark the run approval-eligible. When exactly one active change id is present and one or more relevant file statuses are `added` while the remaining relevant statuses, if any, are `modified`, the workflow SHALL select that change and mark the run comment-only rather than treating it as ineligible.

#### Scenario: Two active changes are present
- **GIVEN** the pull request modifies `openspec/changes/foo/proposal.md` and `openspec/changes/bar/tasks.md`
- **WHEN** deterministic gating runs
- **THEN** the workflow SHALL skip verification and SHALL not submit a review

#### Scenario: Unsupported status under active change
- **GIVEN** the pull request renames `openspec/changes/foo/tasks.md`
- **WHEN** deterministic gating runs
- **THEN** the workflow SHALL skip verification

#### Scenario: Single change updated by modifications only
- **GIVEN** all non-archive `openspec/changes/` paths in the pull request refer to a single `<id>` and every such entry has status `modified`
- **WHEN** deterministic gating runs
- **THEN** the workflow SHALL select that `<id>`, continue to verification, and mark the run approval-eligible

#### Scenario: Single net-new change proposal is comment-only eligible
- **GIVEN** all non-archive `openspec/changes/` paths in the pull request refer to a single `<id>` and at least one such entry has status `added` while all remaining relevant entries, if any, have status `modified`
- **WHEN** deterministic gating runs
- **THEN** the workflow SHALL select that `<id>`, continue to verification, and mark the run comment-only

### Requirement: Verification using active OpenSpec tooling (REQ-007)
For the selected change id and deterministic review disposition published by the pre-activation step, the agent SHALL use standard OpenSpec commands and `.agents/skills/openspec-verify-change/SKILL.md`, including where applicable `npx openspec status --change "<id>" --json` and `npx openspec instructions apply --change "<id>" --json`, and SHALL perform verification with context rooted at `openspec/changes/<id>/`. The prompt SHALL consume the selected change id, review disposition, and disposition reason from workflow outputs rather than requiring the agent to rediscover PR files or approval eligibility before verification. The verification report SHALL include Summary, Issues by priority (CRITICAL, WARNING, SUGGESTION), and Final assessment per the skill.

#### Scenario: CLI resolves the change
- **GIVEN** `openspec/changes/<id>/` exists on the pull request branch and deterministic setup has completed
- **WHEN** `npx openspec status --change "<id>"` runs in the workflow environment
- **THEN** it SHALL succeed for a well-formed active change

#### Scenario: Agent receives deterministic comment-only guidance
- **GIVEN** deterministic gating selected a single active change with one or more added files
- **WHEN** the agent prompt is rendered
- **THEN** it SHALL receive the selected change id and a comment-only disposition from workflow outputs without re-inspecting PR file statuses

### Requirement: Pull request review body (REQ-010)
The review body SHALL summarize verification (Issues by priority) and SHALL include Out-of-scope / unassociated changes with the same expectations as the prior design (list `unassociated`, summarize `uncertain`, note accepted `relevant`). When deterministic pre-activation outputs mark the run comment-only because the selected active change includes added files, the review body SHALL explicitly explain that limitation and SHALL state that the pull request is limited to a `COMMENT` review because it implements a net-new spec change, even if the normal approval criteria are otherwise satisfied.

#### Scenario: Body states unassociated outcome
- **GIVEN** relevance review completes
- **WHEN** the review is submitted
- **THEN** the body SHALL state whether any `unassociated` files were found

#### Scenario: Body explains net-new comment-only limitation
- **GIVEN** the selected active change includes added files and verification finds zero CRITICAL issues and zero `unassociated` files
- **WHEN** the review body is generated
- **THEN** it SHALL explain that the PR met the normal approval criteria but is limited to `COMMENT` because it introduces a net-new spec change

### Requirement: Review event APPROVE vs COMMENT (REQ-012)
The agent SHALL submit a pull request review with `APPROVE` if and only if the deterministic review disposition is approval-eligible and there are zero CRITICAL issues and zero `unassociated` files; otherwise `COMMENT`. The agent SHALL submit `COMMENT` whenever the deterministic review disposition is comment-only, including cases where verification finds zero CRITICAL issues and zero `unassociated` files. It SHALL not use `REQUEST_CHANGES`. WARNING and SUGGESTION alone SHALL NOT block `APPROVE` for approval-eligible runs.

#### Scenario: Approve when gates pass for a modified change
- **GIVEN** deterministic gating marked the selected active change approval-eligible and verification found zero CRITICAL issues and zero `unassociated`
- **WHEN** the review is submitted
- **THEN** `event` SHALL be `APPROVE`

#### Scenario: Net-new change proposal remains comment-only
- **GIVEN** deterministic gating marked the selected active change comment-only because it includes added files
- **AND** verification found zero CRITICAL issues and zero `unassociated`
- **WHEN** the review is submitted
- **THEN** `event` SHALL be `COMMENT`

### Requirement: Archive after APPROVE only (REQ-013)
Only when the agent submits a pull request review with `APPROVE` for an approval-eligible run, the workflow SHALL archive the selected change `<id>` using repository-standard automation (for example, `openspec archive <id>` and/or steps aligned with `openspec-archive-change`), updating `openspec/changes/archive/` and canonical specs per project policy.

#### Scenario: Comment review does not archive
- **GIVEN** the review event is `COMMENT`
- **WHEN** the run completes
- **THEN** the workflow SHALL NOT move the change to `archive/` or mutate canonical specs for this id via this workflow

#### Scenario: Net-new change proposal does not archive
- **GIVEN** deterministic gating marked the selected active change comment-only because it includes added files
- **WHEN** the run completes after submitting its review
- **THEN** the workflow SHALL NOT archive that change through this workflow run
