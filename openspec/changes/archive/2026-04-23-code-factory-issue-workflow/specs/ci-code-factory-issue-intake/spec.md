# `ci-code-factory-issue-intake` — Issue-intake agentic workflow for `code-factory` issues

Workflow implementation: authored source under `.github/workflows-src/`, compiled to `.github/workflows/`.

## Purpose

Define requirements for a GitHub Agentic Workflow that reacts to qualifying `code-factory` issue events, verifies the triggering actor is trusted, suppresses duplicate linked pull requests, and delegates implementation to an agent that creates exactly one linked pull request per issue.

## ADDED Requirements

### Requirement: Workflow source is repository-authored and generated
The repository SHALL define the `code-factory` issue-intake automation as a repository-authored GitHub agentic workflow source under `.github/workflows-src/` that generates checked-in workflow artifacts under `.github/workflows/`. Deterministic GitHub-script logic used for trigger qualification, trust checks, or duplicate detection SHALL be factored into repository-local helper code that can be unit tested independently of the compiled workflow.

#### Scenario: Maintainer updates the workflow source
- **WHEN** maintainers modify the `code-factory` issue-intake workflow
- **THEN** the authored source SHALL live under `.github/workflows-src/` and generate the checked-in workflow artifacts under `.github/workflows/`

#### Scenario: Deterministic gating logic is tested outside the workflow wrapper
- **WHEN** maintainers validate trigger qualification, trust policy, or duplicate detection
- **THEN** the repository SHALL support focused tests for the extracted helper logic without requiring execution of the compiled workflow

### Requirement: Workflow activates the implementation agent only for qualifying `code-factory` issue events
The workflow MAY subscribe to GitHub `issues.opened` and `issues.labeled` events, but it SHALL activate the implementation agent only for eligible `code-factory` issue triggers. Eligible triggers SHALL include `issues.labeled` when the newly applied label is exactly `code-factory`, and `issues.opened` when the issue already includes the `code-factory` label at creation time.

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
Before agent activation, the workflow SHALL determine whether the triggering actor is trusted. The actor SHALL be trusted if the sender is `github-actions[bot]`; otherwise the workflow SHALL query repository collaborator permissions and SHALL require effective repository permission `write`, `maintain`, or `admin`.

#### Scenario: GitHub Actions opens a labeled issue
- **WHEN** the sender is `github-actions[bot]` and the event otherwise qualifies for `code-factory` issue intake
- **THEN** the workflow SHALL treat the trigger as trusted without requiring a collaborator-permission lookup

#### Scenario: Maintainer applies the label
- **WHEN** a human actor triggers an otherwise eligible `code-factory` issue event and the repository permission lookup returns `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL treat the trigger as trusted

#### Scenario: Untrusted actor attempts to trigger automation
- **WHEN** a non-bot actor triggers an otherwise eligible `code-factory` issue event and the repository permission lookup does not return `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL skip agent activation

### Requirement: Workflow suppresses duplicate linked pull requests
Before agent activation, the workflow SHALL detect whether an open linked `code-factory` pull request already exists for the triggering issue. A pull request SHALL be treated as linked only when it is open, carries the `code-factory` label, uses the deterministic branch name `code-factory/issue-<issue-number>`, and includes an explicit reference to the issue in stable metadata such as `Closes #<issue-number>`.

#### Scenario: Existing linked PR prevents a duplicate run
- **WHEN** the workflow finds an open pull request that satisfies the linked `code-factory` PR criteria for the triggering issue
- **THEN** the workflow SHALL skip agent activation instead of opening a duplicate pull request

#### Scenario: Unrelated PR does not block issue intake
- **WHEN** an open pull request mentions the issue or has a similar title but does not satisfy the full linked `code-factory` PR criteria
- **THEN** the workflow SHALL NOT treat that pull request as the canonical linked PR for duplicate suppression

### Requirement: Agent creates exactly one linked `code-factory` pull request
When the deterministic gate passes, the workflow agent SHALL treat the triggering issue as the source of truth for implementation, SHALL work on the deterministic branch `code-factory/issue-<issue-number>`, and SHALL create or update exactly one linked pull request labeled `code-factory`. The linked pull request SHALL preserve explicit issue linkage in its title or body so future reruns can deterministically identify it.

#### Scenario: Eligible issue creates a linked implementation PR
- **WHEN** the workflow runs for a trusted eligible issue event and no open linked `code-factory` pull request already exists
- **THEN** the agent SHALL implement the issue on branch `code-factory/issue-<issue-number>` and open one linked pull request carrying the `code-factory` label

#### Scenario: Pull request metadata preserves deterministic linkage
- **WHEN** the agent creates the `code-factory` pull request
- **THEN** the pull request SHALL include explicit issue linkage metadata so later workflow runs can identify it as the canonical PR for the issue
