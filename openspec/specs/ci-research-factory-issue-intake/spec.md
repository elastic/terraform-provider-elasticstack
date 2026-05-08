# ci-research-factory-issue-intake Specification

## Purpose
TBD - created by archiving change research-factory-workflow. Update Purpose after archive.
## Requirements
### Requirement: Workflow source is repository-authored and generated
The repository SHALL define the `research-factory` issue-intake automation as a repository-authored GitHub Agentic Workflow source under `.github/workflows-src/research-factory-issue/` that generates checked-in workflow artifacts under `.github/workflows/`: the compiled markdown `.github/workflows/research-factory-issue.md` and the compiled `.github/workflows/research-factory-issue.lock.yml` from `gh aw compile`. Contributors SHALL NOT hand-edit those generated files; they SHALL be regenerated with repository workflow tooling (`make workflow-generate`). Deterministic GitHub-script logic used for trigger qualification, dispatch input validation, comment-history capture, or context normalization SHALL be factored into repository-local helper code under `.github/workflows-src/lib/` that can be tested independently of the compiled workflow.

#### Scenario: Maintainer updates the workflow source
- **WHEN** maintainers modify the `research-factory` issue-intake workflow
- **THEN** the authored source SHALL live under `.github/workflows-src/research-factory-issue/`
- **AND** the generated `.github/workflows/research-factory-issue.md` and `.github/workflows/research-factory-issue.lock.yml` SHALL be checked in and SHALL match output from the repository generation commands

#### Scenario: Deterministic logic is tested outside the workflow wrapper
- **WHEN** maintainers validate trigger qualification, dispatch input parsing, comment filtering, or context normalization for `research-factory`
- **THEN** the repository SHALL support focused tests for the extracted helper logic without requiring execution of the compiled workflow

### Requirement: Workflow activates the research agent only for qualifying triggers
The workflow SHALL subscribe to GitHub `issues.opened`, `issues.labeled`, and `workflow_dispatch` events. For issue-event intake, eligible triggers SHALL include `issues.labeled` when the newly applied label is exactly `research-factory`, and `issues.opened` when the issue already includes the `research-factory` label at creation time. For dispatch intake, the workflow SHALL accept a typed `issue_number` input (and an optional `source_workflow` input for traceability) and SHALL treat dispatch as eligible when the input identifies one issue in the current repository.

#### Scenario: Label applied after issue creation
- **WHEN** an `issues.labeled` event is received and `github.event.label.name` is `research-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the research agent

#### Scenario: Issue opens with the trigger label already present
- **WHEN** an `issues.opened` event is received and the issue's initial labels include `research-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the research agent

#### Scenario: Non-trigger issue event is ignored
- **WHEN** an `issues` event is received without the `research-factory` label in the qualifying position for that event type
- **THEN** the workflow SHALL NOT activate the research agent for that event

#### Scenario: Internal workflow dispatch requests research
- **WHEN** the workflow is triggered by `workflow_dispatch` with a valid `issue_number` for the current repository
- **THEN** the workflow SHALL treat that dispatch as eligible to activate the research agent subject to its dispatch-mode deterministic gates

#### Scenario: Dispatch input references another repository
- **WHEN** `workflow_dispatch` provides an issue identifier that does not resolve to an issue in the current repository
- **THEN** the workflow SHALL reject or stop that run rather than researching a cross-repository issue

### Requirement: Trigger actor must be trusted for issue-event intake
For issue-event intake, before agent activation the workflow SHALL determine whether the triggering actor is trusted. The actor SHALL be trusted if the sender is `github-actions[bot]`; otherwise the workflow SHALL query repository collaborator permissions and SHALL require effective repository permission `write`, `maintain`, or `admin`. This trust policy SHALL apply only to issue-event intake and SHALL NOT be required for repository-authored `workflow_dispatch` intake.

#### Scenario: GitHub Actions opens a labeled issue
- **WHEN** the sender is `github-actions[bot]` and the event otherwise qualifies for `research-factory` issue intake
- **THEN** the workflow SHALL treat the trigger as trusted without requiring a collaborator-permission lookup

#### Scenario: Maintainer applies the label
- **WHEN** a human actor triggers an otherwise eligible `research-factory` issue event and the repository permission lookup returns `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL treat the trigger as trusted

#### Scenario: Untrusted actor attempts to trigger automation
- **WHEN** a non-bot actor triggers an otherwise eligible `research-factory` issue event and the repository permission lookup does not return `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL skip agent activation

#### Scenario: Internal dispatch bypasses issue-event trust lookup
- **WHEN** the workflow is triggered by repository-authored `workflow_dispatch`
- **THEN** the workflow SHALL NOT require the issue-event actor trust check as a prerequisite for activation

### Requirement: Workflow normalizes issue intake context across entry modes
The workflow SHALL normalize issue intake into downstream-consumable outputs that include the resolved issue number, title, body, intake mode, source workflow (if any), gate reason, and the captured comment history so that the research prompt and downstream steps do not depend directly on `github.event.issue.*`. For dispatch intake, the workflow SHALL fetch the live issue body and title from GitHub rather than trusting body or title text passed through dispatch inputs.

#### Scenario: Issue-event intake exposes normalized outputs
- **WHEN** the workflow is triggered from an eligible issue event
- **THEN** the workflow SHALL publish normalized outputs for the resolved issue number, title, body, intake mode, gate reason, and comment history

#### Scenario: Dispatch intake exposes normalized outputs
- **WHEN** the workflow is triggered from `workflow_dispatch`
- **THEN** the workflow SHALL fetch the live issue from GitHub and publish the same normalized outputs for the resolved issue number, title, body, intake mode, gate reason, and comment history

### Requirement: Workflow captures human-authored comment history for the agent
Before agent activation, the workflow SHALL capture all comments on the triggering issue, in chronological order, filtered to human-authored comments only. Comments authored by `github-actions[bot]`, by the workflow's own status-comment author, and by other automation bots known to the repository SHALL be excluded. The captured history SHALL be exposed to the agent prompt alongside the issue body so the agent can read the prior conversation, including any prior research output and any human replies to it. The capture step SHALL be implemented as a shared helper under `.github/workflows-src/lib/` reusable across factory workflows.

#### Scenario: Issue has prior research and human follow-up comments
- **WHEN** an eligible `research-factory` event fires on an issue that already contains a research block in its body and one or more human comments
- **THEN** the workflow SHALL include the chronological human comment history in the normalized intake context delivered to the agent

#### Scenario: Bot comments are excluded
- **WHEN** capturing comment history for the agent
- **THEN** the captured comments SHALL exclude `github-actions[bot]` and similar automation bots
- **AND** SHALL exclude the workflow framework's own status-comment authors

#### Scenario: Issue has no comments
- **WHEN** an eligible event fires on an issue with no comments
- **THEN** the workflow SHALL still publish a (possibly empty) normalized comment-history output without failing

### Requirement: Workflow enforces single-session-per-issue concurrency
The workflow SHALL declare GitHub Actions concurrency keyed by the resolved issue number such that at most one `research-factory` run SHALL execute per issue at any time. Newly arriving triggers for an issue with an in-flight run SHALL be queued rather than canceling the in-flight run; superseded queued runs MAY be collapsed by GitHub's standard concurrency semantics.

#### Scenario: Two triggers fire for the same issue in rapid succession
- **WHEN** a second qualifying trigger arrives for an issue while a `research-factory` run is already in flight for that issue
- **THEN** the second run SHALL be queued and SHALL NOT execute concurrently with the first
- **AND** the in-flight run SHALL NOT be canceled

#### Scenario: Triggers fire for distinct issues
- **WHEN** qualifying triggers arrive for two different issues at the same time
- **THEN** the two runs MAY execute concurrently because they belong to different concurrency groups

### Requirement: Workflow time-boxes the research session and survives partial completion
The workflow SHALL set a job-level `timeout-minutes` of 35 minutes. The agent prompt SHALL communicate a 25-minute self-budget to the agent and SHALL instruct it to reserve the final minutes of the budget for emitting its issue-body update. The prompt SHALL further instruct the agent that, if research time runs short, it SHALL prefer emitting a partial-but-valid research block (with explicit unanswered open questions) over emitting `noop`.

#### Scenario: Maintainer inspects compiled workflow timeout
- **WHEN** maintainers inspect the compiled `research-factory-issue.md` workflow
- **THEN** the agent job SHALL declare `timeout-minutes: 35`

#### Scenario: Agent prompt communicates the self-budget
- **WHEN** maintainers inspect the agent prompt body
- **THEN** the prompt SHALL state the 25-minute research self-budget
- **AND** the prompt SHALL state that the agent SHALL prefer a partial-but-valid research block over `noop` when running short on time

### Requirement: Workflow remains research-only and does not write code
The `research-factory` workflow SHALL NOT implement provider, CI, or documentation behavior, SHALL NOT open pull requests, and SHALL NOT modify repository files. Its only durable output SHALL be a single `update-issue` operation against the triggering issue's body in the format defined by the `ci-implementation-research-block-format` capability. The workflow SHALL NOT enable safe outputs that would permit creating pull requests, creating issues, or posting free-form comments beyond the framework's own `status-comment`.

#### Scenario: Maintainer inspects compiled workflow safe outputs
- **WHEN** maintainers inspect the compiled `research-factory-issue.md` workflow `safe-outputs:` block
- **THEN** it SHALL include `update-issue` with `body:` enabled
- **AND** it SHALL NOT include `create-pull-request`, `push-to-pull-request-branch`, `update-pull-request`, or `create-issue`
- **AND** it SHALL NOT include `add-comment`

#### Scenario: Issue requests provider implementation
- **WHEN** a qualifying issue describes a Terraform resource, data source, or other provider implementation
- **THEN** the agent SHALL produce a research block describing approaches and open questions
- **AND** the agent SHALL NOT modify provider source, generated clients, or documentation

### Requirement: Workflow bootstraps only research-authoring tooling
Before the research agent runs, the workflow SHALL provision tooling needed to author the research block and read the repository: a Git checkout of the default branch with `fetch-depth: 0` for full history, and Node.js via `actions/setup-node` with `node-version-file: package.json` plus `npm ci` so the agent can run repository tooling if needed. The workflow SHALL NOT start the Elastic Stack, create Elasticsearch API keys, set up Fleet, or run Terraform acceptance tests.

#### Scenario: Agent has full repository checkout
- **WHEN** the research agent starts for a qualifying run
- **THEN** the agent's working directory SHALL contain a full Git checkout of the repository default branch with `fetch-depth: 0`

#### Scenario: Elastic Stack services are not provisioned
- **WHEN** the `research-factory` workflow prepares the research-authoring environment
- **THEN** it SHALL NOT run Elastic Stack, Fleet, or Elasticsearch API-key setup steps

### Requirement: Research agent has structured access to Elastic documentation
The `research-factory` workflow SHALL configure the Elastic docs MCP server as an HTTP MCP server in the workflow frontmatter so that the research agent can query Elastic documentation while comparing approaches. The workflow frontmatter SHALL include `www.elastic.co` in `network.allowed` and SHALL declare an `mcp-servers.elastic-docs` entry pointing to `https://www.elastic.co/docs/_mcp/`. The agent prompt SHALL instruct the agent to use the docs MCP tools (`search_docs`, `find_related_docs`, `get_document_by_url`) when investigating the API behavior, parameters, or constraints referenced by a `research-factory` issue.

#### Scenario: Agent investigates an unfamiliar Elastic API feature
- **WHEN** a `research-factory` issue references an Elastic API endpoint or feature the agent has not encountered before
- **THEN** the agent SHALL use the elastic-docs MCP `search_docs` tool to locate relevant Elastic documentation before authoring the research block
- **AND** it SHALL cite the consulted documentation URLs in the block's References section

#### Scenario: Elastic docs MCP server is unavailable
- **WHEN** the elastic-docs MCP tools return an error or are unreachable during a `research-factory` run
- **THEN** the agent SHALL proceed with research using the information available in the issue, the repository codebase, and prior comments
- **AND** it SHALL NOT block the run or emit `noop` solely because the docs MCP is unavailable

#### Scenario: Maintainer inspects compiled workflow for docs MCP configuration
- **WHEN** maintainers inspect the compiled `research-factory-issue.md` workflow
- **THEN** the workflow frontmatter SHALL include `mcp-servers.elastic-docs` with `url: https://www.elastic.co/docs/_mcp/`
- **AND** `network.allowed` SHALL include `www.elastic.co`

### Requirement: Workflow status comments on the issue include the run link
The workflow SHALL set `status-comment: true` in the top-level `on:` configuration so the activation job posts a status comment on the triggering issue when the run starts and updates it when the run completes, including a link to the workflow run as provided by the framework. The repository SHALL NOT add a custom implementation-job step solely to post the workflow run URL to the issue.

#### Scenario: Status comment enabled
- **WHEN** maintainers inspect the authored `research-factory` issue-intake workflow `on:` frontmatter
- **THEN** it SHALL include `status-comment: true` (or an object form that enables status comments for issues)

#### Scenario: No custom comment step for run linkage
- **WHEN** the workflow is authored for `research-factory` issue intake
- **THEN** the repository SHALL NOT rely on a custom implementation-job step solely to post the workflow run URL to the issue; run visibility SHALL be covered by `status-comment` as above

### Requirement: Workflow removes the factory trigger label in pre-activation when the agent proceeds
The workflow SHALL include a deterministic pre-activation step that removes the `research-factory` label from the triggering issue using `actions/github-script@v9` with `x-script-include` to an inline script that delegates to the shared `.github/workflows-src/lib/remove-trigger-label.js` helper (parameterized to accept the factory label name and issue number). The step SHALL run only when the workflow would proceed to the research agent (eligible qualifying issue event, trusted actor — when applicable — and concurrency gate satisfied). The workflow SHALL grant `issues: write` to pre-activation where required for label removal. Dispatch-targeted issues SHALL NOT require the `research-factory` label to be present, and the workflow SHALL NOT treat the absence of that label on dispatch-targeted issues as an error.

#### Scenario: Remove step mirrors verify workflow pattern
- **WHEN** maintainers inspect the authored `research-factory` issue-intake workflow `on.steps`
- **THEN** it SHALL include a remove-label step structurally equivalent to the change-factory and code-factory remove-label steps, with the `Remove trigger label` step using `actions/github-script@v9` and `x-script-include` to include the shared trigger-label removal script/helper path
- **AND** the included script SHALL reuse the shared `remove-trigger-label` library (not a forked copy of the GitHub API logic)

#### Scenario: Label removed only when agent gate passes
- **WHEN** pre-activation determines the research agent SHALL run for the issue
- **THEN** the remove-label step SHALL run and SHALL attempt to remove `research-factory` from that issue

#### Scenario: Dispatch-targeted issue has no `research-factory` label
- **WHEN** the workflow is triggered by `workflow_dispatch` for an issue that does not carry the `research-factory` label
- **THEN** the workflow SHALL continue normally and SHALL NOT require trigger-label removal for that run

### Requirement: Workflow frontmatter allows required agent ecosystems
The `research-factory` workflow SHALL declare an authored AWF network policy that allows the default allowlist plus the Node ecosystem, allows `elastic.litellm-prod.ai` for the Claude engine's Anthropic-compatible proxy access, and allows `www.elastic.co` for the Elastic docs MCP server.

#### Scenario: Maintainer inspects workflow frontmatter
- **WHEN** maintainers inspect the authored `research-factory` issue-intake workflow frontmatter
- **THEN** `network.allowed` SHALL include `defaults`
- **AND** `network.allowed` SHALL include `node`
- **AND** `network.allowed` SHALL include `elastic.litellm-prod.ai`
- **AND** `network.allowed` SHALL include `www.elastic.co`

### Requirement: Agent updates the issue body with a single research block per run
When the deterministic gate passes and the agent completes its research, the agent SHALL emit a single `update_issue` safe-output operation with `operation: replace` whose `body` payload contains the original issue content (everything outside the implementation-research markers preserved byte-for-byte) followed by exactly one current `<!-- implementation-research:start --> ... <!-- implementation-research:end -->` block conforming to the `ci-implementation-research-block-format` capability. The agent SHALL NOT emit additional `update_issue` operations, SHALL NOT append a second block, and SHALL NOT emit `add-comment`. If the agent cannot make any meaningful research progress at all (for example because the issue body is empty and no comments exist), it MAY emit `noop` instead.

#### Scenario: Agent produces a research block on a fresh issue
- **WHEN** the workflow runs for an eligible issue with no prior research block in its body
- **THEN** the agent SHALL emit one `update_issue` operation with `operation: replace`
- **AND** the new body SHALL preserve the original issue content unchanged before the markers
- **AND** the new body SHALL contain exactly one `<!-- implementation-research:start --> ... <!-- implementation-research:end -->` block

#### Scenario: Agent regenerates an existing research block
- **WHEN** the workflow runs for an eligible issue whose body already contains a research block
- **THEN** the agent SHALL emit one `update_issue` operation with `operation: replace`
- **AND** the new body SHALL contain exactly one research block (the prior block SHALL NOT remain alongside the new one)
- **AND** the new body SHALL preserve all content outside the markers byte-for-byte relative to the pre-block original issue content

#### Scenario: Agent times out before reaching a confident recommendation
- **WHEN** the agent's self-budget expires before research is complete
- **THEN** the agent SHALL emit a partial-but-valid research block with explicit unanswered open questions
- **AND** the agent SHALL NOT emit `noop` solely because research is partial

### Requirement: Workflow does not promote the issue to a downstream factory
The `research-factory` workflow SHALL NOT apply the `change-factory`, `code-factory`, or any other factory trigger label as part of its research run. Promotion of an issue from the research stage to a downstream stage SHALL be performed by a human maintainer or by a separate (future) classifier workflow, not by `research-factory` itself.

#### Scenario: Research run completes successfully
- **WHEN** the research agent finishes a successful run
- **THEN** the workflow SHALL NOT add the `change-factory`, `code-factory`, or any other factory trigger label to the issue

#### Scenario: Maintainer inspects compiled workflow safe outputs
- **WHEN** maintainers inspect the compiled `research-factory-issue.md` workflow
- **THEN** its `safe-outputs:` block SHALL NOT enable `add-labels`

