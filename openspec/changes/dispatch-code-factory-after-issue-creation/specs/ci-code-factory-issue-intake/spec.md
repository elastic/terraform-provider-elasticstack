## MODIFIED Requirements

### Requirement: Workflow activates the implementation agent only for qualifying `code-factory` issue events
The workflow MAY subscribe to GitHub `issues.opened` and `issues.labeled` events. For issue-event intake, eligible triggers SHALL include `issues.labeled` when the newly applied label is exactly `code-factory`, and `issues.opened` when the issue already includes the `code-factory` label at creation time.

#### Scenario: Label applied after issue creation
- **WHEN** an `issues.labeled` event is received and `github.event.label.name` is `code-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the implementation agent

#### Scenario: Issue opens with the trigger label already present
- **WHEN** an `issues.opened` event is received and the issue's initial labels include `code-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the implementation agent

#### Scenario: Non-trigger issue event is ignored
- **WHEN** an `issues` event is received without the `code-factory` label in the qualifying position for that event type
- **THEN** the workflow SHALL NOT activate the implementation agent for that event

### Requirement: Trigger actor must be trusted
Before agent activation, the workflow SHALL determine whether the triggering actor is trusted for issue-event intake. The actor SHALL be trusted if the sender is `github-actions[bot]`; otherwise the workflow SHALL query repository collaborator permissions and SHALL require effective repository permission `write`, `maintain`, or `admin`. This trust policy applies only to issue-event intake and SHALL NOT be required for repository-authored `workflow_dispatch` intake.

#### Scenario: GitHub Actions opens a labeled issue
- **WHEN** the sender is `github-actions[bot]` and the event otherwise qualifies for `code-factory` issue intake
- **THEN** the workflow SHALL treat the trigger as trusted without requiring a collaborator-permission lookup

#### Scenario: Maintainer applies the label
- **WHEN** a human actor triggers an otherwise eligible `code-factory` issue event and the repository permission lookup returns `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL treat the trigger as trusted

#### Scenario: Untrusted actor attempts to trigger automation
- **WHEN** a non-bot actor triggers an otherwise eligible `code-factory` issue event and the repository permission lookup does not return `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL skip agent activation

#### Scenario: Internal dispatch bypasses issue-event trust lookup
- **WHEN** the workflow is triggered by repository-authored `workflow_dispatch`
- **THEN** the workflow SHALL NOT require the issue-event actor trust check as a prerequisite for activation

### Requirement: Workflow suppresses duplicate linked pull requests
Before agent activation, the workflow SHALL detect whether an open linked `code-factory` pull request already exists for the triggering issue, regardless of whether intake came from an issue event or `workflow_dispatch`. A pull request SHALL be treated as linked only when it is open, carries the `code-factory` label, uses the deterministic branch name `code-factory/issue-<issue-number>`, and includes an explicit reference to the issue in stable metadata such as `Closes #<issue-number>`.

#### Scenario: Existing linked PR prevents a duplicate issue-event run
- **WHEN** the workflow finds an open pull request that satisfies the linked `code-factory` PR criteria for the triggering issue
- **THEN** the workflow SHALL skip agent activation instead of opening a duplicate pull request

#### Scenario: Existing linked PR prevents a duplicate dispatch run
- **WHEN** an internally dispatched run targets an issue that already has an open linked `code-factory` pull request
- **THEN** the workflow SHALL skip agent activation instead of opening a duplicate pull request

#### Scenario: Unrelated PR does not block issue intake
- **WHEN** an open pull request mentions the issue or has a similar title but does not satisfy the full linked `code-factory` PR criteria
- **THEN** the workflow SHALL NOT treat that pull request as the canonical linked PR for duplicate suppression

### Requirement: Agent creates exactly one linked `code-factory` pull request
When the deterministic gate passes, the workflow agent SHALL treat the resolved issue as the source of truth for implementation, SHALL work on the deterministic branch `code-factory/issue-<issue-number>`, and SHALL create or update exactly one linked pull request labeled `code-factory`. The linked pull request SHALL preserve explicit issue linkage in its title or body so future reruns can deterministically identify it.

#### Scenario: Eligible issue-event intake creates a linked implementation PR
- **WHEN** the workflow runs for a trusted eligible issue event and no open linked `code-factory` pull request already exists
- **THEN** the agent SHALL implement the issue on branch `code-factory/issue-<issue-number>` and open one linked pull request carrying the `code-factory` label

#### Scenario: Eligible dispatch intake creates a linked implementation PR
- **WHEN** the workflow runs from `workflow_dispatch` for a valid issue in the current repository and no open linked `code-factory` pull request already exists
- **THEN** the agent SHALL implement the issue on branch `code-factory/issue-<issue-number>` and open one linked pull request carrying the `code-factory` label

#### Scenario: Pull request metadata preserves deterministic linkage
- **WHEN** the agent creates the `code-factory` pull request
- **THEN** the pull request SHALL include explicit issue linkage metadata so later workflow runs can identify it as the canonical PR for the issue

## ADDED Requirements

### Requirement: Workflow activates the implementation agent for valid internal dispatch requests
The workflow SHALL support internal single-issue activation through `workflow_dispatch` when the dispatch provides valid current-repository issue identity and the run passes its dispatch-mode deterministic gates.

#### Scenario: Internal workflow dispatch requests issue intake
- **WHEN** the workflow is triggered by `workflow_dispatch` with a valid issue number for the current repository
- **THEN** the workflow SHALL treat that dispatch as eligible to activate the implementation agent subject to its dispatch-mode deterministic gates

### Requirement: Dispatch intake resolves live issue context from workflow inputs
For `workflow_dispatch` intake, the workflow SHALL accept enough typed input to identify one issue in the current repository and SHALL resolve the live issue title and body from GitHub before activating the implementation agent. The workflow SHALL use the live issue as the canonical source of scope rather than trusting issue body or title text passed through dispatch inputs.

#### Scenario: Dispatch provides issue number and repository
- **WHEN** `workflow_dispatch` provides a valid issue number and current-repository identifier
- **THEN** the workflow SHALL fetch the live issue details from GitHub before prompting the agent

#### Scenario: Dispatch input references another repository
- **WHEN** `workflow_dispatch` provides an issue repository that does not match the current repository
- **THEN** the workflow SHALL reject or stop that run rather than implementing a cross-repository issue

### Requirement: Workflow normalizes issue intake context across entry modes
The workflow SHALL normalize issue intake into downstream-consumable outputs that include the resolved issue number, title, body, intake mode, and gate reason so that the implementation prompt and downstream steps do not depend directly on `github.event.issue.*`.

#### Scenario: Issue-event intake exposes normalized outputs
- **WHEN** the workflow is triggered from an eligible issue event
- **THEN** the workflow SHALL publish normalized outputs for the resolved issue number, title, body, intake mode, and gate reason

#### Scenario: Dispatch intake exposes normalized outputs
- **WHEN** the workflow is triggered from `workflow_dispatch`
- **THEN** the workflow SHALL publish the same normalized outputs for the resolved issue number, title, body, intake mode, and gate reason

### Requirement: Trigger label removal remains issue-event-only
The workflow SHALL keep `code-factory` trigger-label removal scoped to the manual issue-event path. Dispatch-triggered runs SHALL NOT require the target issue to carry the `code-factory` label, and the workflow SHALL NOT treat the absence of that label on dispatch-targeted issues as an error.

#### Scenario: Manual issue-event run removes the trigger label when proceeding
- **WHEN** an eligible trusted issue-event run proceeds past deterministic gates and the issue carries the `code-factory` label
- **THEN** the workflow SHALL remove the `code-factory` label before agent activation as defined by the base intake behavior

#### Scenario: Dispatch-targeted issue has no `code-factory` label
- **WHEN** the workflow is triggered by `workflow_dispatch` for an issue that does not carry the `code-factory` label
- **THEN** the workflow SHALL continue normally and SHALL NOT require trigger-label removal for that run
