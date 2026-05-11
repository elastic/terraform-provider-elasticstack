# ci-change-factory-slash-command Specification

## Purpose
TBD - created by archiving change change-factory-slash-command. Update Purpose after archive.
## Requirements
### Requirement: Workflow responds to `/change-factory` slash commands on issues
The `change-factory-issue` workflow SHALL add a `slash_command:` trigger with `name: change-factory` and `events: [issue_comment]`. In gh-aw's event model, `issue_comment` and `pull_request_comment` are distinct event names; declaring `events: [issue_comment]` confines the slash command to issue comments and excludes pull request conversation comments without any additional payload guard. This trigger SHALL coexist with the existing `issues: [labeled]` trigger on the same workflow.

#### Scenario: Maintainer posts slash command on an issue
- **WHEN** a trusted maintainer posts a comment beginning with `/change-factory` on a GitHub issue
- **THEN** the `change-factory-issue` workflow SHALL activate for that issue

#### Scenario: Slash command on a pull request comment is ignored
- **WHEN** a `/change-factory` comment is posted on a pull request conversation
- **THEN** the workflow SHALL NOT activate, because gh-aw routes pull request conversation comments under `pull_request_comment`, which is not listed in `events:`

#### Scenario: Label trigger continues to work alongside slash command
- **WHEN** a maintainer applies the `change-factory` label to an issue
- **THEN** the workflow SHALL activate exactly as before, independently of whether a slash command was ever used on that issue

### Requirement: Text following the slash command is captured as human direction
The workflow SHALL include a deterministic `capture_command_text` pre-activation step that, when the triggering event is `issue_comment`, extracts the content of the triggering comment body after the `/change-factory` command token (stripping leading and trailing whitespace) and exposes it as a `human_direction` pre-activation output. When the triggering event is not `issue_comment`, the step SHALL output an empty string for `human_direction`.

#### Scenario: Maintainer posts a slash command with direction text
- **WHEN** a comment `/change-factory use approach B — the SDK migration approach` triggers the workflow
- **THEN** the `human_direction` pre-activation output SHALL be `use approach B — the SDK migration approach`

#### Scenario: Slash command with no text after the command name
- **WHEN** a comment contains only `/change-factory` with no additional text
- **THEN** the `human_direction` output SHALL be an empty string and the workflow SHALL proceed as a standard `change-factory` run

#### Scenario: Label path produces no human direction
- **WHEN** the workflow activates via the `issues: [labeled]` path
- **THEN** the `human_direction` output SHALL be an empty string

### Requirement: Non-empty human direction overrides the research recommendation in the agent prompt
When `human_direction` is non-empty, the `change-factory` agent prompt SHALL present it under a clearly labelled section as the final say on all design decisions for the proposal, explicitly stating that it overrides the research comment's `### Recommendation` and any other design inferences. The agent SHALL apply the direction without second-guessing it.

#### Scenario: Human direction selects a non-recommended approach
- **WHEN** the research comment recommends Approach A but the human posts `/change-factory use approach B instead`
- **THEN** the proposal agent SHALL plan the OpenSpec change around Approach B
- **AND** the agent SHALL NOT follow the research comment's `### Recommendation` for Approach A

#### Scenario: Human direction adjusts scope beyond approach selection
- **WHEN** the human posts `/change-factory skip the caching layer for now, focus only on the read path`
- **THEN** the proposal agent SHALL incorporate that scope constraint as an authoritative design directive
- **AND** the agent SHALL NOT reinstate the caching layer based on the research comment

#### Scenario: Empty human direction falls back to research recommendation
- **WHEN** `human_direction` is empty (label-triggered run or bare slash command)
- **THEN** the agent SHALL follow the research comment's `### Recommendation` as normal, or use the issue title and body if no research comment is present

### Requirement: Duplicate-PR gate posts an explanatory comment before skipping
When the `change-factory` workflow determines that an open linked `change-factory` pull request already exists for the triggering issue (duplicate gate fires), the workflow SHALL post exactly one comment on the triggering issue explaining the skip and linking to the existing PR, before the activation job exits. The comment SHALL instruct the maintainer to close or convert the existing PR to a draft before retrying. This behaviour SHALL apply regardless of whether the run was triggered by label or slash command.

#### Scenario: Slash command fires but a PR is already open
- **WHEN** a maintainer posts `/change-factory use approach B` and a linked `change-factory` PR is already open for that issue
- **THEN** the workflow SHALL post a comment on the issue referencing the existing PR URL and instructing the maintainer to close it before retrying
- **AND** the workflow SHALL NOT activate the proposal agent

#### Scenario: Label trigger fires but a PR is already open
- **WHEN** the `change-factory` label is applied to an issue that already has a linked open `change-factory` PR
- **THEN** the workflow SHALL post the same explanatory comment on the issue
- **AND** the workflow SHALL NOT activate the proposal agent

#### Scenario: No existing PR — duplicate gate does not fire
- **WHEN** the workflow runs and no open linked `change-factory` PR exists for the issue
- **THEN** the workflow SHALL NOT post the duplicate-blocked comment
- **AND** it SHALL proceed to agent activation normally

### Requirement: Slash command triggers use the same trust policy as label triggers
The actor trust check SHALL apply to slash command triggers using the same collaborator-permission policy as label triggers: `github-actions[bot]` is trusted unconditionally; human actors require effective repository permission `write`, `maintain`, or `admin`.

#### Scenario: Trusted maintainer uses slash command
- **WHEN** a collaborator with `write` or higher permission posts `/change-factory`
- **THEN** the actor SHALL be trusted and the workflow SHALL proceed past the trust gate

#### Scenario: Untrusted actor posts the slash command
- **WHEN** a user without sufficient repository permission posts `/change-factory` on an issue
- **THEN** the workflow SHALL not activate the proposal agent (trust gate fails)

