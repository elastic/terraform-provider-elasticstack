# `ci-change-factory-issue-intake` — Issue-labeled OpenSpec proposal factory

Workflow implementation: authored source under `.github/workflows-src/change-factory-issue/`, compiled to `.github/workflows/change-factory-issue.md`.

## Purpose

Define requirements for a GitHub Agentic Workflow that reacts to trusted GitHub issues labeled `change-factory` and creates exactly one linked OpenSpec change proposal pull request, without implementing provider behavior or provisioning the Elastic Stack.

## ADDED Requirements

### Requirement: Workflow source is repository-authored and generated
The repository SHALL define the `change-factory` issue-intake automation as a repository-authored GitHub Agentic Workflow source under `.github/workflows-src/` that generates checked-in workflow artifacts under `.github/workflows/`. Deterministic GitHub-script logic used for trigger qualification, trust checks, or duplicate detection SHALL be factored into repository-local helper code that can be tested independently of the compiled workflow.

#### Scenario: Maintainer updates the workflow source
- **WHEN** maintainers modify the `change-factory` issue-intake workflow
- **THEN** the authored source SHALL live under `.github/workflows-src/`
- **AND** the generated workflow artifacts SHALL be checked in under `.github/workflows/`

#### Scenario: Deterministic gating logic is tested outside the workflow wrapper
- **WHEN** maintainers validate trigger qualification, trust policy, or duplicate detection
- **THEN** the repository SHALL support focused tests for the extracted helper logic without requiring execution of the compiled workflow

### Requirement: Workflow activates only for qualifying `change-factory` issue events
The workflow MAY subscribe to GitHub `issues.opened` and `issues.labeled` events, but it SHALL activate the proposal agent only for eligible `change-factory` issue triggers. Eligible triggers SHALL include `issues.labeled` when the newly applied label is exactly `change-factory`, and `issues.opened` when the issue already includes the `change-factory` label at creation time.

#### Scenario: Label applied after issue creation
- **WHEN** an `issues.labeled` event is received and `github.event.label.name` is `change-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the proposal agent

#### Scenario: Issue opens with the trigger label already present
- **WHEN** an `issues.opened` event is received and the issue's initial labels include `change-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the proposal agent

#### Scenario: Non-trigger issue event is ignored
- **WHEN** an `issues` event is received without the `change-factory` label in the qualifying position for that event type
- **THEN** the workflow SHALL NOT activate the proposal agent for that event

### Requirement: Trigger actor must be trusted
Before agent activation, the workflow SHALL determine whether the triggering actor is trusted. The actor SHALL be trusted if the sender is `github-actions[bot]`; otherwise the workflow SHALL query repository collaborator permissions and SHALL require effective repository permission `write`, `maintain`, or `admin`.

#### Scenario: GitHub Actions opens a labeled issue
- **WHEN** the sender is `github-actions[bot]` and the event otherwise qualifies for `change-factory` issue intake
- **THEN** the workflow SHALL treat the trigger as trusted without requiring a collaborator-permission lookup

#### Scenario: Maintainer applies the label
- **WHEN** a human actor triggers an otherwise eligible `change-factory` issue event and the repository permission lookup returns `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL treat the trigger as trusted

#### Scenario: Untrusted actor attempts to trigger automation
- **WHEN** a non-bot actor triggers an otherwise eligible `change-factory` issue event and the repository permission lookup does not return `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL skip agent activation

### Requirement: Workflow suppresses duplicate linked pull requests
Before agent activation, the workflow SHALL detect whether an open linked `change-factory` pull request already exists for the triggering issue. A pull request SHALL be treated as linked only when it is open, carries the `change-factory` label, uses the deterministic branch name `change-factory/issue-<issue-number>`, and includes explicit issue linkage metadata such as `Closes #<issue-number>`.

#### Scenario: Existing linked PR prevents a duplicate run
- **WHEN** the workflow finds an open pull request that satisfies the linked `change-factory` PR criteria for the triggering issue
- **THEN** the workflow SHALL skip agent activation instead of opening a duplicate pull request

#### Scenario: Unrelated PR does not block issue intake
- **WHEN** an open pull request mentions the issue or has a similar title but does not satisfy the full linked `change-factory` PR criteria
- **THEN** the workflow SHALL NOT treat that pull request as the canonical linked PR for duplicate suppression

### Requirement: Agent creates exactly one linked OpenSpec proposal pull request
When the deterministic gate passes, the workflow agent SHALL treat the triggering issue title and body as the authoritative source for requested proposal scope, SHALL work on the deterministic branch `change-factory/issue-<issue-number>`, and SHALL create or update exactly one linked pull request labeled `change-factory`. The pull request SHALL contain one active OpenSpec change under `openspec/changes/<change-id>/` with the artifacts required for implementation readiness by the active OpenSpec schema.

#### Scenario: Eligible issue creates a linked proposal PR
- **WHEN** the workflow runs for a trusted eligible issue event and no open linked `change-factory` pull request already exists
- **THEN** the agent SHALL create an OpenSpec change proposal on branch `change-factory/issue-<issue-number>`
- **AND** it SHALL open one linked pull request carrying the `change-factory` label

#### Scenario: Pull request metadata preserves deterministic linkage
- **WHEN** the agent creates the `change-factory` pull request
- **THEN** the pull request SHALL include explicit issue linkage metadata so later workflow runs can identify it as the canonical PR for the issue

#### Scenario: Proposal artifacts are implementation-ready
- **WHEN** the agent completes a proposal pull request
- **THEN** the pull request SHALL include all OpenSpec artifacts required before implementation can begin according to the repository's active OpenSpec schema

### Requirement: Workflow remains proposal-only
The `change-factory` workflow SHALL NOT implement the requested provider, CI, or documentation behavior as part of the proposal-generation run. It SHALL limit repository changes to OpenSpec change artifacts and any workflow metadata required by the proposal workflow itself.

#### Scenario: Issue asks for provider behavior
- **WHEN** a qualifying issue describes a new Terraform resource, data source, or provider behavior change
- **THEN** the agent SHALL create OpenSpec proposal artifacts for that work
- **AND** it SHALL NOT implement provider code in the same pull request

#### Scenario: Proposal requires assumptions
- **WHEN** the issue context is sufficient to propose a change but leaves secondary details unresolved
- **THEN** the agent SHALL capture assumptions or open questions in the OpenSpec artifacts rather than implementing speculative behavior

### Requirement: Workflow bootstraps only proposal-authoring tooling
Before the proposal agent runs OpenSpec commands, the workflow SHALL provision the repository tooling needed to author and validate OpenSpec artifacts. At minimum, it SHALL set up Node using `actions/setup-node` with `node-version-file: package.json` and SHALL install repository npm dependencies so the local OpenSpec CLI is available. The workflow SHALL NOT start the Elastic Stack, create Elasticsearch API keys, set up Fleet, or run Terraform acceptance tests.

#### Scenario: OpenSpec CLI is available before agent reasoning
- **WHEN** the agent starts for a qualifying `change-factory` run
- **THEN** deterministic setup SHALL have made the repository-pinned OpenSpec CLI available in the workspace

#### Scenario: Elastic Stack services are not provisioned
- **WHEN** the `change-factory` workflow prepares the proposal-authoring environment
- **THEN** it SHALL NOT run Elastic Stack, Fleet, or Elasticsearch API-key setup steps

#### Scenario: Acceptance tests are out of scope
- **WHEN** the proposal agent validates its work
- **THEN** it SHALL validate OpenSpec structure rather than running Terraform acceptance tests

### Requirement: Unclear issues result in noop rather than an exploration loop
If the triggering issue lacks enough context for the agent to create a coherent OpenSpec proposal, the workflow SHALL emit a single no-op result with a concise clarification reason instead of opening an exploratory comment thread, creating a GitHub Discussion, or producing speculative proposal artifacts.

#### Scenario: Core scope is unclear
- **WHEN** the issue title and body do not provide enough information to determine the proposed change scope
- **THEN** the agent SHALL use `noop` with a concise explanation of what clarification is needed

#### Scenario: Issue is clear enough for a proposal
- **WHEN** the issue title and body provide enough context to identify the change scope and capability area
- **THEN** the agent SHALL create the linked OpenSpec proposal pull request without requiring a GitHub comment exploration loop
