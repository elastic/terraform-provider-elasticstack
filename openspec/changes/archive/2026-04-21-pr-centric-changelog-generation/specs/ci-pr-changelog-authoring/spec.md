# `ci-pr-changelog-authoring` — PR-time changelog contract enforcement

Workflow implementation: authored source template under `.github/workflows-src/`, generated GH AW markdown under `.github/workflows/`, and compiled `.lock.yml` via `gh aw compile`.

## Purpose

Define a GitHub Agentic Workflow that runs after pull-request CI completes and ensures each pull request either contains a valid `## Changelog` section in its body or explicitly opts out with the `no-changelog` label.

## ADDED Requirements

### Requirement: Workflow artifacts and compilation
The PR changelog authoring workflow SHALL be authored from a repository template under `.github/workflows-src/` that generates a GitHub Agentic Workflow markdown file under `.github/workflows/` via `scripts/compile-workflow-sources/main.go`. The repository SHALL commit the generated `.md` workflow and the compiled `.lock.yml` produced by `gh aw compile`. Contributors SHALL NOT hand-edit the generated `.md` or `.lock.yml` artifacts.

#### Scenario: Source and compiled artifacts stay paired
- **WHEN** maintainers change the PR changelog authoring workflow behavior
- **THEN** the `.github/workflows-src/` template, generated `.md` workflow, and compiled `.lock.yml` SHALL match the committed compiler output

### Requirement: Trigger on `Build/Lint/Test` workflow completion
The workflow SHALL run from a `workflow_run` trigger for the repository workflow named `Build/Lint/Test` and SHALL continue only when the source workflow run completed for a pull-request event.

#### Scenario: Pull-request CI completion is eligible for gating
- **WHEN** the `Build/Lint/Test` workflow completes for a `pull_request` event
- **THEN** the PR changelog authoring workflow SHALL continue to deterministic pull-request resolution and gating

#### Scenario: Non-pull-request workflow run is skipped
- **WHEN** the `Build/Lint/Test` workflow completes for a non-`pull_request` event such as `push` or `workflow_dispatch`
- **THEN** the PR changelog authoring workflow SHALL NOT invoke changelog validation or authoring for that run

### Requirement: Deterministic pull-request resolution and opt-out gate
Before agent reasoning starts, deterministic repository-authored steps SHALL resolve the pull request associated with the triggering `workflow_run`. The workflow SHALL skip agent authoring when the resolved pull request carries the `no-changelog` label.

#### Scenario: `no-changelog` label suppresses authoring
- **WHEN** deterministic resolution finds the triggering pull request and that pull request carries the `no-changelog` label
- **THEN** the workflow SHALL treat the pull request as explicitly exempt from changelog authoring

#### Scenario: Missing pull request fails gating
- **WHEN** deterministic resolution cannot identify exactly one pull request for the triggering workflow run
- **THEN** the workflow SHALL fail gating without invoking the agent, with the deterministic resolution step exiting non-zero and emitting an error message prefixed with `PR_CHANGELOG_GATING:`

### Requirement: Existing changelog section is validated deterministically
When the pull request body already contains a `## Changelog` section, deterministic repository-authored validation SHALL verify the contract shape before the workflow reports success. The validator SHALL require `Customer impact` to be exactly one of `none`, `fix`, `enhancement`, or `breaking`; these values are case-sensitive and SHALL be matched literally. The validator SHALL require a `Summary` line whenever `Customer impact` is not `none`, and a non-empty optional `### Breaking changes` subsection when that subsection is present.

#### Scenario: Valid changelog section suppresses the agent
- **WHEN** the pull request body already contains a valid `## Changelog` section
- **THEN** the workflow SHALL complete successfully without invoking the agent

#### Scenario: Malformed changelog section fails validation
- **WHEN** the pull request body contains a `## Changelog` section that does not satisfy the deterministic contract
- **THEN** the workflow SHALL fail with a clear validation error instead of overwriting the existing section

### Requirement: Breaking changes subsection may be free-form markdown
Within the `## Changelog` contract, the optional `### Breaking changes` subsection SHALL allow free-form markdown content, including prose, bullet lists, and fenced code blocks. Deterministic validation SHALL treat that subsection as a delimited markdown block rather than as a structured bullet schema.

#### Scenario: Breaking changes block contains fenced code
- **WHEN** the pull request body includes `### Breaking changes` with fenced code blocks or migration prose
- **THEN** the workflow SHALL accept that subsection as valid markdown content when the block is non-empty

### Requirement: Missing changelog sections are drafted from PR metadata
When the resolved pull request lacks a `## Changelog` section and is not exempt via `no-changelog`, the agent SHALL draft the missing section from the pull request title and description and SHALL update the pull request body with that drafted section.

#### Scenario: Missing changelog section is added
- **WHEN** deterministic gating concludes the pull request requires a changelog section and none is present
- **THEN** the workflow SHALL invoke the agent to draft the `## Changelog` section and update the pull request body with the result

### Requirement: Minimal permissions are available to deterministic gating and PR updates
The workflow SHALL request only the minimal permissions needed to resolve the triggering pull request, validate its body, and update that body when necessary. At minimum the workflow SHALL grant `contents: read` and `pull-requests: write`, and those scopes SHALL be available to the deterministic pre-agent steps rather than only to the agent phase.

#### Scenario: Deterministic pre-agent steps can update PR body
- **WHEN** deterministic resolution, validation, or PR-body update steps execute
- **THEN** the workflow permissions SHALL allow those steps to read repository-authored context and update the pull request body without requiring broader repository write scopes

### Requirement: `workflow_run` execution remains metadata-only
Because the workflow runs from `workflow_run`, it SHALL NOT checkout or execute code from the pull-request head branch. Deterministic gating, validation, and agent authoring SHALL operate only on pull-request metadata and repository-authored prompt context.

#### Scenario: PR workflow never checks out untrusted code
- **WHEN** the workflow evaluates or updates the changelog contract for a pull request
- **THEN** it SHALL do so without checking out or executing code from the pull-request branch
