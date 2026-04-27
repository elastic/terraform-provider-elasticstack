# ci-code-factory-issue-intake Specification

## Purpose
TBD - created by archiving change code-factory-issue-workflow. Update Purpose after archive.
## Requirements
### Requirement: Workflow source is repository-authored and generated
The repository SHALL define the `code-factory` issue-intake automation as a repository-authored GitHub agentic workflow source under `.github/workflows-src/` that generates checked-in workflow artifacts under `.github/workflows/`. Deterministic GitHub-script logic used for trigger qualification, trust checks, or duplicate detection SHALL be factored into repository-local helper code that can be unit tested independently of the compiled workflow.

#### Scenario: Maintainer updates the workflow source
- **WHEN** maintainers modify the `code-factory` issue-intake workflow
- **THEN** the authored source SHALL live under `.github/workflows-src/` and generate the checked-in workflow artifacts under `.github/workflows/`

#### Scenario: Deterministic gating logic is tested outside the workflow wrapper
- **WHEN** maintainers validate trigger qualification, trust policy, or duplicate detection
- **THEN** the repository SHALL support focused tests for the extracted helper logic without requiring execution of the compiled workflow

### Requirement: Workflow frontmatter allows required agent ecosystems
The `code-factory` issue-intake workflow SHALL declare an authored AWF network policy that allows the default allowlist plus the Node and Go ecosystems, allows `elastic.litellm-prod.ai` for the Claude engine's Anthropic-compatible proxy access, and allows `www.elastic.co` for the Elastic docs MCP server.

#### Scenario: Maintainer inspects workflow frontmatter
- **WHEN** maintainers inspect the authored `code-factory` issue-intake workflow frontmatter
- **THEN** `network.allowed` SHALL include `defaults`
- **AND** `network.allowed` SHALL include `node`
- **AND** `network.allowed` SHALL include `go`
- **AND** `network.allowed` SHALL include `elastic.litellm-prod.ai`
- **AND** `network.allowed` SHALL include `www.elastic.co`

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

### Requirement: Workflow status comments on the issue include the run link
The workflow SHALL set `status-comment: true` in the top-level `on:` configuration (see GitHub Agentic Workflows [status comments](https://github.github.com/gh-aw/reference/triggers/#status-comments-status-comment)) so the activation job posts a status comment on the triggering issue when the run starts and updates it when the run completes, including a link to the workflow run as provided by the framework.

#### Scenario: Status comment enabled
- **WHEN** maintainers inspect the authored `code-factory` issue-intake workflow `on:` frontmatter
- **THEN** it SHALL include `status-comment: true` (or an object form that enables status comments for issues)

#### Scenario: No custom comment step for run linkage
- **WHEN** the workflow is authored for `code-factory` issue intake
- **THEN** the repository SHALL NOT rely on a custom implementation-job step solely to post the workflow run URL to the issue; run visibility SHALL be covered by `status-comment` as above

### Requirement: Workflow removes the factory trigger label in pre-activation when the agent proceeds
The workflow SHALL include a deterministic pre-activation step that removes the `code-factory` label from the triggering issue **using the same mechanism as** OpenSpec verify (label): `actions/github-script@v9` with `x-script-include` to an inline script that delegates to the shared `.github/workflows-src/lib/remove-trigger-label.js` helper (generalized to accept the factory label name and issue number). The step SHALL run only when the workflow would proceed to the implementation agent (eligible qualifying issue event, trusted actor, and no open linked `code-factory` pull request per existing duplicate suppression). The workflow SHALL grant `issues: write` to pre-activation where required for label removal.

#### Scenario: Remove step mirrors verify workflow pattern
- **WHEN** maintainers inspect the authored `code-factory` issue-intake workflow `on.steps`
- **THEN** it SHALL include a remove-label step structurally equivalent to OpenSpec verify (label), including step name `Remove trigger label`, `uses: actions/github-script@v9`, and an `x-script-include` reference for the inline script
- **AND** the included script SHALL reuse the generalized `remove-trigger-label` library (not a forked copy of the GitHub API logic)

#### Scenario: Label removed only when agent gate passes
- **WHEN** pre-activation determines the implementation agent SHALL run for the issue
- **THEN** the remove-label step SHALL run and SHALL attempt to remove `code-factory` from that issue

#### Scenario: Label retained when agent does not run
- **WHEN** pre-activation suppresses the agent (ineligible event, untrusted actor, or duplicate linked PR)
- **THEN** the workflow SHALL NOT remove `code-factory` solely as a side effect of this intake run

### Requirement: Implementation agent has structured access to Elastic documentation
The `code-factory` workflow SHALL configure the Elastic docs MCP server as an HTTP MCP server in the workflow frontmatter so that the implementation agent can query Elastic documentation during issue investigation and implementation. The workflow frontmatter SHALL declare an `mcp-servers.elastic-docs` entry pointing to `https://www.elastic.co/docs/_mcp/`. The agent prompt SHALL instruct the agent to use the docs MCP tools (`search_docs`, `find_related_docs`, `get_document_by_url`) when investigating the API behavior, parameters, or constraints required to implement a `code-factory` issue.

#### Scenario: Agent investigates API behavior before implementing a resource
- **WHEN** a `code-factory` issue involves an Elastic API endpoint or feature whose full parameter set is not evident from the existing codebase
- **THEN** the agent SHALL use the elastic-docs MCP `search_docs` tool to retrieve authoritative API documentation before writing implementation code

#### Scenario: Elastic docs MCP server is unavailable
- **WHEN** the elastic-docs MCP tools return an error or are unreachable during a `code-factory` run
- **THEN** the agent SHALL proceed with implementation using the information available in the issue and the repository codebase
- **AND** it SHALL NOT block the run solely because the docs MCP is unavailable

#### Scenario: Maintainer inspects compiled workflow for docs MCP configuration
- **WHEN** maintainers inspect the compiled `code-factory-issue.md` workflow
- **THEN** the workflow frontmatter SHALL include `mcp-servers.elastic-docs` with `url: https://www.elastic.co/docs/_mcp/`

