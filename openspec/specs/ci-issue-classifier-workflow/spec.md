# ci-issue-classifier-workflow Specification

## Purpose
TBD - created by archiving change issue-classifier-workflow. Update Purpose after archive.
## Requirements
### Requirement: Workflow triggers on issue opened, daily schedule, and manual dispatch

The issue classifier workflow SHALL be triggered by three events:
1. `issues: [opened]` — fires immediately when a new issue is created
2. `schedule: daily` — fires once per day as a backlog sweep
3. `workflow_dispatch` with an optional `issue_number` input — for manual testing

#### Scenario: New issue triggers classification
- **WHEN** a new GitHub issue is opened in the repository
- **THEN** the workflow SHALL start and attempt to classify that issue

#### Scenario: Daily schedule triggers backlog sweep
- **WHEN** the daily cron fires
- **THEN** the workflow SHALL query for untriaged backlog issues and classify up to 5

#### Scenario: Manual dispatch with issue number
- **WHEN** `workflow_dispatch` is triggered with a valid `issue_number` input
- **THEN** the workflow SHALL classify that specific issue

#### Scenario: Manual dispatch without issue number
- **WHEN** `workflow_dispatch` is triggered without an `issue_number` input
- **THEN** the workflow SHALL behave identically to the scheduled trigger (process up to 5 backlog issues)

### Requirement: Pre-activation gate skips already-triaged issues

The pre-activation step SHALL check for the `triaged` label before passing any issue to the agent.

#### Scenario: Issue-opened trigger on already-triaged issue
- **WHEN** the `issues: opened` trigger fires for an issue that already has the `triaged` label
- **THEN** the pre-activation step SHALL output a gate reason and the agent job SHALL be skipped

#### Scenario: Scheduled trigger finds no untriaged issues
- **WHEN** the daily schedule fires and all open issues have the `triaged` label
- **THEN** the pre-activation step SHALL set `issue_count` to `0` and the agent job SHALL be skipped

### Requirement: Scheduled path selects up to 5 newest untriaged issues

On the scheduled (and dispatch-without-issue-number) path, the pre-activation step SHALL query open issues lacking the `triaged` label and select up to 5, ordered by creation date descending (newest first).

#### Scenario: Fewer than 5 untriaged issues exist
- **WHEN** there are 3 open untriaged issues
- **THEN** the pre-activation step SHALL select all 3

#### Scenario: More than 5 untriaged issues exist
- **WHEN** there are 12 open untriaged issues
- **THEN** the pre-activation step SHALL select the 5 most recently created

#### Scenario: No untriaged issues exist
- **WHEN** there are no open untriaged issues
- **THEN** the agent job SHALL be skipped via the `noop` safe-output

### Requirement: Agent classifies each issue into exactly one category

The agent SHALL assign each issue to exactly one of the following categories based on its title, body, and any existing labels:

1. **`needs-research`**: The issue is a feature request — a request for a new Terraform resource, data source, or new functionality on an existing entity or the provider itself. The request SHALL be sufficiently specific and well-defined to route to research-factory.

2. **`needs-reproduction`**: The issue is a bug report containing at least one of: a Terraform configuration demonstrating the problem, an error message, or a thorough description of reproduction steps. Suitable to route to reproducer-factory.

3. **`needs-spec`**: The issue already contains sufficient detail to describe the solution accurately — both the problem and the intended fix are clearly articulated. This category SHALL be used rarely and only when the bar is unambiguously met.

4. **`needs-human`**: The issue does not clearly fit any other category, lacks sufficient detail, needs clarification, or requires human judgement to route correctly. This is the catch-all.

#### Scenario: Specific feature request classified as needs-research
- **WHEN** an issue requests a new Terraform resource for a specific Elastic API with clear scope
- **THEN** the agent SHALL classify it as `needs-research`

#### Scenario: Vague feature request classified as needs-human
- **WHEN** an issue requests "better support for X" without specifying what changes are needed
- **THEN** the agent SHALL classify it as `needs-human`

#### Scenario: Bug report with Terraform config classified as needs-reproduction
- **WHEN** an issue describes unexpected behaviour and includes a Terraform configuration block
- **THEN** the agent SHALL classify it as `needs-reproduction`

#### Scenario: Bug report without any reproduction evidence classified as needs-human
- **WHEN** an issue says "resource X is broken" with no error message, config, or steps
- **THEN** the agent SHALL classify it as `needs-human`

#### Scenario: Well-specified implementation issue classified as needs-spec
- **WHEN** an issue contains a clear problem statement and a fully-articulated proposed solution
- **THEN** the agent MAY classify it as `needs-spec` if the bar is unambiguously met

### Requirement: Agent applies triaged label and one needs-* label to every classified issue

After classifying an issue, the agent SHALL emit an `add_labels` safe-output call with both the `triaged` label and the appropriate `needs-*` label for that issue.

#### Scenario: Labels applied on successful classification
- **WHEN** the agent classifies an issue as `needs-reproduction`
- **THEN** the `add_labels` safe-output SHALL apply both `triaged` and `needs-reproduction` to that issue

#### Scenario: Only allowed labels are applied
- **WHEN** the safe-outputs job processes `add_labels` calls
- **THEN** only labels in the allowlist `[triaged, needs-research, needs-reproduction, needs-spec, needs-human]` SHALL be accepted

### Requirement: Agent posts a classification comment on every classified issue

After classifying an issue, the agent SHALL emit an `add_comment` safe-output call with a comment that:
- Begins with the HTML marker `<!-- gha-issue-classifier -->`
- Names the assigned label and explains what it means in plain language
- Describes what happens next (which factory pipeline the issue routes to, if applicable)
- Invites the reporter to add a correction comment if they believe the classification is wrong

#### Scenario: Comment posted after classification
- **WHEN** the agent classifies an issue
- **THEN** the `add_comment` safe-output SHALL post exactly one comment on that issue in the same run

#### Scenario: Re-run hides previous classification comment
- **WHEN** the workflow runs a second time on the same issue (e.g., after label correction)
- **THEN** the previous classification comment SHALL be hidden with reason `outdated` and a new comment SHALL be posted

#### Scenario: Comment marker is present
- **WHEN** a classification comment is posted
- **THEN** the comment body SHALL begin with `<!-- gha-issue-classifier -->`

### Requirement: Agent emits noop when no issues are classified

When the agent completes a run without classifying any issues (all issues already triaged, or no untriaged issues found), it SHALL call the `noop` safe-output with a descriptive reason.

#### Scenario: Noop on empty scheduled run
- **WHEN** the scheduled path runs and finds no untriaged issues
- **THEN** the agent SHALL call `noop` with a reason explaining that no untriaged issues were found

### Requirement: Safe-outputs configuration restricts label scope and comment volume

The workflow safe-outputs configuration SHALL enforce the following bounds:

- `add-labels`: `max: 5`, `target: "*"`, `allowed: [triaged, needs-research, needs-reproduction, needs-spec, needs-human]`, `blocked: ["~*", "*[bot]"]`
- `add-comment`: `max: 5`, `target: "*"`, `hide-older-comments: true`, `allowed-reasons: [outdated]`, `footer: false`

#### Scenario: Agent cannot apply unlisted label
- **WHEN** the agent attempts to apply a label not in the allowlist
- **THEN** the safe-outputs job SHALL reject that label

#### Scenario: Maximum label operations per run
- **WHEN** the scheduled path classifies 5 issues
- **THEN** the workflow SHALL emit at most 5 `add_labels` calls and 5 `add_comment` calls

