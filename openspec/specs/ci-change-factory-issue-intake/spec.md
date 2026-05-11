# ci-change-factory-issue-intake Specification

## Purpose

Define requirements for a GitHub Agentic Workflow that reacts to trusted GitHub issues labeled `change-factory` and creates exactly one linked OpenSpec change proposal pull request, without implementing provider behavior or provisioning the Elastic Stack.
## Requirements
### Requirement: Workflow source is repository-authored and generated
The repository SHALL define the `change-factory` issue-intake automation as a repository-authored GitHub Agentic Workflow source under `.github/workflows-src/` that generates checked-in workflow artifacts under `.github/workflows/`: the compiled markdown `.github/workflows/change-factory-issue.md` and the compiled `.github/workflows/change-factory-issue.lock.yml` from `gh aw compile`. Contributors SHALL NOT hand-edit those generated files; they SHALL be regenerated with repository workflow tooling (`make workflow-generate`). Deterministic GitHub-script logic used for trigger qualification, trust checks, or duplicate detection SHALL be factored into repository-local helper code that can be tested independently of the compiled workflow.

#### Scenario: Maintainer updates the workflow source
- **WHEN** maintainers modify the `change-factory` issue-intake workflow
- **THEN** the authored source SHALL live under `.github/workflows-src/`
- **AND** the generated `.github/workflows/change-factory-issue.md` and `.github/workflows/change-factory-issue.lock.yml` SHALL be checked in and SHALL match output from the repository generation commands

#### Scenario: Deterministic gating logic is tested outside the workflow wrapper
- **WHEN** maintainers validate trigger qualification, trust policy, or duplicate detection
- **THEN** the repository SHALL support focused tests for the extracted helper logic without requiring execution of the compiled workflow

### Requirement: Workflow activates only for qualifying `change-factory` issue events
The workflow MAY subscribe to GitHub `issues.opened`, `issues.labeled`, and `issue_comment` events, but it SHALL activate the proposal agent only for eligible `change-factory` triggers. Eligible triggers SHALL include `issues.labeled` when the newly applied label is exactly `change-factory`; `issues.opened` when the issue already includes the `change-factory` label at creation time; and `issue_comment` events when the comment is the activation payload of a `slash_command: change-factory` trigger. In gh-aw's event model, `issue_comment` and `pull_request_comment` are distinct event names; only `issue_comment` activations (i.e. comments on issues, not pull requests) are delivered to a workflow that declares `events: [issue_comment]`. The shared `factoryQualifyTriggerEvent` helper SHALL treat `issue_comment` as an automatically eligible event name — returning `event_eligible: true` unconditionally — because gh-aw's routing guarantees that only issue-comment activations reach this handler; actor trust and duplicate checks remain the effective substantive gates.

#### Scenario: Label applied after issue creation
- **WHEN** an `issues.labeled` event is received and `github.event.label.name` is `change-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the proposal agent

#### Scenario: Issue opens with the trigger label already present
- **WHEN** an `issues.opened` event is received and the issue's initial labels include `change-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the proposal agent

#### Scenario: Non-trigger issue event is ignored
- **WHEN** an `issues` event is received without the `change-factory` label in the qualifying position for that event type
- **THEN** the workflow SHALL NOT activate the proposal agent for that event

#### Scenario: Slash command comment is treated as eligible
- **WHEN** an `issue_comment` event is received as the activation payload of a `slash_command: change-factory` trigger
- **THEN** `factoryQualifyTriggerEvent` SHALL return `event_eligible: true` for that event, because gh-aw's `events: [issue_comment]` routing guarantees the payload originates from an issue comment (not a pull request conversation)

#### Scenario: Pull request comments cannot reach the slash command handler
- **WHEN** a `/change-factory` comment is posted on a pull request conversation
- **THEN** gh-aw routes it under `pull_request_comment`, which is not listed in the workflow's `events:`; the workflow SHALL NOT activate, and `factoryQualifyTriggerEvent` is never called for that payload

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
Before agent activation, the workflow SHALL detect whether an open linked `change-factory` pull request already exists for the triggering issue. A pull request SHALL be treated as linked only when it is open, carries the `change-factory` label, uses the deterministic branch name `change-factory/issue-<issue-number>`, and includes explicit issue linkage metadata such as `Closes #<issue-number>`. When a duplicate is found, the workflow SHALL post exactly one comment on the triggering issue explaining the skip and linking to the existing PR URL, then skip agent activation. The comment SHALL instruct the maintainer to close or convert the PR to a draft before retrying.

#### Scenario: Existing linked PR prevents a duplicate run and posts a comment
- **WHEN** the workflow finds an open pull request that satisfies the linked `change-factory` PR criteria for the triggering issue
- **THEN** the workflow SHALL post one comment on the triggering issue referencing the existing PR and instructing the maintainer to close it before retrying
- **AND** the workflow SHALL skip agent activation instead of opening a duplicate pull request

#### Scenario: Unrelated PR does not block issue intake
- **WHEN** an open pull request mentions the issue or has a similar title but does not satisfy the full linked `change-factory` PR criteria
- **THEN** the workflow SHALL NOT treat that pull request as the canonical linked PR for duplicate suppression

### Requirement: Workflow sanitizes HTML comments from agent input context
Before the `change-factory` agent reads the triggering issue body and human-authored comments, the workflow SHALL strip all HTML comments from that content using the shared `ci-html-comment-sanitisation` helpers. The sanitised issue body and comment history SHALL be exposed to the agent prompt. The research-factory sticky comment, when present, SHALL be extracted separately and SHALL NOT be passed through the stripping step because it is bot-authored trusted output.

#### Scenario: Agent receives clean context
- **WHEN** the `change-factory` workflow runs for an issue whose body contains an injected HTML comment
- **THEN** the agent SHALL receive a sanitised body with that comment removed

#### Scenario: Human comments with HTML comments are cleaned
- **WHEN** a human comment on the issue contains an HTML comment
- **THEN** the sanitised comment text delivered to the change-factory agent SHALL have that comment removed

### Requirement: Agent uses the implementation-research comment as the authoritative scope baseline when present
When the triggering issue has a comment authored by `github-actions[bot]` that contains the marker `<!-- gha-research-factory -->` and a heading `## Implementation research`, the `change-factory` agent SHALL treat that entire comment as the authoritative baseline for proposal scope — unless a non-empty `human_direction` is present, in which case `human_direction` SHALL take precedence as the final say on all design decisions. When `human_direction` is empty, the agent SHALL adopt the comment's `### Recommendation` as the proposal spine, carry the comment's `### Open questions` into `design.md`, and treat `### Approaches considered` as already-evaluated context without re-exploring alternatives. If the sanitised issue body or sanitised human comments contain explicit signals that contradict the research recommendation (and no `human_direction` override is present), the agent SHALL prefer the contradicting signal and note the disagreement in the proposal artifacts. When no research comment exists, the agent SHALL use the issue title and body as the authoritative source regardless of `human_direction`.

#### Scenario: Issue has a research comment and no human direction
- **WHEN** a `change-factory` run starts for an issue that has a bot-authored research comment and `human_direction` is empty
- **THEN** the agent SHALL adopt the comment's `### Recommendation` as the chosen approach and use it as the spine of `proposal.md`
- **AND** the agent SHALL copy the comment's `### Open questions` into the resulting `design.md`
- **AND** the agent SHALL NOT re-explore the alternative approaches enumerated in `### Approaches considered`

#### Scenario: Human direction overrides research recommendation
- **WHEN** a `change-factory` run starts for an issue with a research comment and `human_direction` is non-empty
- **THEN** the agent SHALL treat `human_direction` as the final say on design decisions
- **AND** the agent SHALL NOT follow the research comment's `### Recommendation` if it conflicts with `human_direction`

#### Scenario: Issue has no research comment
- **WHEN** a `change-factory` run starts for an issue that does not have a bot-authored research comment
- **THEN** the agent SHALL author the proposal using the issue title and body as the authoritative source
- **AND** any non-empty `human_direction` SHALL still apply as the final say on design decisions

#### Scenario: Issue has a research comment but later comments contradict the recommendation
- **WHEN** a `change-factory` run starts for an issue that has a research comment and whose visible context contradicts the comment's recommendation, and `human_direction` is empty
- **THEN** the agent SHALL prefer the contradicting signal and SHALL note the disagreement in the proposal artifacts

### Requirement: Agent prompt documents implementation-research comment awareness
The `change-factory` workflow's authored prompt SHALL include explicit instructions describing the implementation-research comment: it is authored by `github-actions[bot]`, identified by the marker `<!-- gha-research-factory -->`, contains a `## Implementation research` heading, and when present its `### Recommendation` and `### Open questions` are the authoritative inputs for the proposal. The prompt SHALL state that when no such comment exists, the existing title-and-body-authoritative behavior applies unchanged. The prompt SHALL NOT instruct the agent to add, modify, or remove the research comment itself — comment management belongs to the `research-factory` workflow.

#### Scenario: Maintainer inspects the change-factory prompt
- **WHEN** maintainers inspect the authored `change-factory-issue` workflow prompt
- **THEN** the prompt SHALL describe the `github-actions[bot]` research comment with its `<!-- gha-research-factory -->` marker
- **AND** the prompt SHALL state that, when the marker is present, the comment's `### Recommendation` and `### Open questions` are the authoritative inputs for the proposal
- **AND** the prompt SHALL state that, when the marker is absent, the existing title-and-body-authoritative behavior applies unchanged

#### Scenario: Change-factory does not mutate the research comment
- **WHEN** the `change-factory` agent runs against an issue with an implementation-research comment
- **THEN** the agent SHALL NOT emit operations that modify the research comment
- **AND** the agent SHALL NOT add, remove, or rewrite the `<!-- gha-research-factory -->` marker

### Requirement: Agent creates exactly one linked OpenSpec proposal pull request
When the deterministic gate passes, the workflow agent SHALL treat the triggering issue title and body as the authoritative source for requested proposal scope, SHALL work on the deterministic branch `change-factory/issue-<issue-number>`, and SHALL create or update exactly one linked pull request labeled `change-factory` and `no-changelog`. The pull request SHALL contain one active OpenSpec change under `openspec/changes/<change-id>/` with the artifacts required for implementation readiness by the active OpenSpec schema.

#### Scenario: Eligible issue creates a linked proposal PR
- **WHEN** the workflow runs for a trusted eligible issue event and no open linked `change-factory` pull request already exists
- **THEN** the agent SHALL create an OpenSpec change proposal on branch `change-factory/issue-<issue-number>`
- **AND** it SHALL open one linked pull request carrying the `change-factory` and `no-changelog` labels

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

### Requirement: Unclear issues request facts on the source issue without an exploration loop
If the triggering issue lacks enough context for the agent to create a coherent OpenSpec proposal, the workflow SHALL post exactly one `add-comment` on the triggering issue listing the specific facts still needed before emitting `noop`, then emit at most one `noop` with a brief completion note. It SHALL NOT complete that outcome using only `noop` without the required `add-comment`. It SHALL NOT open a back-and-forth comment thread, create a GitHub Discussion, open new issues, or produce speculative proposal artifacts.

#### Scenario: Core scope is unclear
- **WHEN** the issue title and body do not provide enough information to determine the proposed change scope
- **THEN** the agent SHALL use `add-comment` on the triggering issue with a concise list of required facts
- **AND** it SHALL use `noop` with a brief completion note only after that comment

#### Scenario: Issue is clear enough for a proposal
- **WHEN** the issue title and body provide enough context to identify the change scope and capability area
- **THEN** the agent SHALL create the linked OpenSpec proposal pull request without requiring a GitHub comment exploration loop

### Requirement: Workflow status comments on the issue include the run link
The workflow SHALL set `status-comment: true` in the top-level `on:` configuration (see GitHub Agentic Workflows [status comments](https://github.github.com/gh-aw/reference/triggers/#status-comments-status-comment)) so the activation job posts a status comment on the triggering issue when the run starts and updates it when the run completes, including a link to the workflow run as provided by the framework.

#### Scenario: Status comment enabled
- **WHEN** maintainers inspect the authored `change-factory` issue-intake workflow `on:` frontmatter
- **THEN** it SHALL include `status-comment: true` (or an object form that enables status comments for issues)

#### Scenario: No custom comment step for run linkage
- **WHEN** the workflow is authored for `change-factory` issue intake
- **THEN** the repository SHALL NOT rely on a custom implementation-job step solely to post the workflow run URL to the issue; run visibility SHALL be covered by `status-comment` as above

### Requirement: Workflow removes the factory trigger label in pre-activation when the agent proceeds
The workflow SHALL include a deterministic pre-activation step that removes the `change-factory` label from the triggering issue **using the same mechanism as** OpenSpec verify (label): `actions/github-script@v9` with `x-script-include` to an inline script that delegates to the shared `.github/workflows-src/lib/remove-trigger-label.js` helper (generalized to accept the factory label name and issue number). The step SHALL run only when the workflow would proceed to the proposal agent (eligible qualifying issue event, trusted actor, and no open linked `change-factory` pull request per existing duplicate suppression). The workflow SHALL grant `issues: write` to pre-activation where required for label removal.

#### Scenario: Remove step mirrors verify workflow pattern
- **WHEN** maintainers inspect the authored `change-factory` issue-intake workflow `on.steps`
- **THEN** it SHALL include a remove-label step structurally equivalent to OpenSpec verify (label), with the `Remove trigger label` step using `actions/github-script@v9` and `x-script-include` to include the shared trigger-label removal script/helper path
- **AND** the included script SHALL reuse the generalized `remove-trigger-label` library (not a forked copy of the GitHub API logic)

#### Scenario: Label removed only when agent gate passes
- **WHEN** pre-activation determines the proposal agent SHALL run for the issue
- **THEN** the remove-label step SHALL run and SHALL attempt to remove `change-factory` from that issue

#### Scenario: Label retained when agent does not run
- **WHEN** pre-activation suppresses the agent (ineligible event, untrusted actor, or duplicate linked PR)
- **THEN** the workflow SHALL NOT remove `change-factory` solely as a side effect of this intake run

### Requirement: Proposal agent has structured access to Elastic documentation
The `change-factory` workflow SHALL configure the Elastic docs MCP server as an HTTP MCP server in the workflow frontmatter so that the proposal agent can query Elastic documentation during issue investigation. The workflow frontmatter SHALL include `www.elastic.co` in `network.allowed` and SHALL declare an `mcp-servers.elastic-docs` entry pointing to `https://www.elastic.co/docs/_mcp/`. The agent prompt SHALL instruct the agent to use the docs MCP tools (`search_docs`, `find_related_docs`, `get_document_by_url`) when investigating the API behavior, parameters, or constraints referenced by a `change-factory` issue.

#### Scenario: Agent investigates an unfamiliar Elastic API feature
- **WHEN** a `change-factory` issue references an Elastic API endpoint or feature the agent has not encountered before
- **THEN** the agent SHALL use the elastic-docs MCP `search_docs` tool to locate relevant Elastic documentation before authoring the OpenSpec proposal
- **AND** it SHALL use findings from the documentation to populate accurate API parameter names, types, and behavior in the delta specs

#### Scenario: Elastic docs MCP server is unavailable
- **WHEN** the elastic-docs MCP tools return an error or are unreachable during a `change-factory` run
- **THEN** the agent SHALL proceed with proposal authoring using the information available in the issue and the repository codebase
- **AND** it SHALL NOT block the run or emit `noop` solely because the docs MCP is unavailable

#### Scenario: Maintainer inspects compiled workflow for docs MCP configuration
- **WHEN** maintainers inspect the compiled `change-factory-issue.md` workflow
- **THEN** the workflow frontmatter SHALL include `mcp-servers.elastic-docs` with `url: https://www.elastic.co/docs/_mcp/`
- **AND** `network.allowed` SHALL include `www.elastic.co`

### Requirement: Agent prompt documents human direction as a design override
The `change-factory` workflow's authored prompt SHALL include a `## Human direction` section that is presented when `human_direction` is non-empty. The section SHALL state that the human direction is the final say on all design decisions for this proposal, that it overrides the research comment's `### Recommendation` and any other design inferences, and that the agent SHALL apply it without second-guessing.

#### Scenario: Maintainer inspects the change-factory prompt for human direction handling
- **WHEN** maintainers inspect the authored `change-factory-issue` workflow prompt
- **THEN** the prompt SHALL include a section for `human_direction` that describes it as the final say on design decisions when non-empty
- **AND** the prompt SHALL explicitly state that it overrides the research comment's `### Recommendation`

#### Scenario: Empty human direction does not change agent behaviour
- **WHEN** `human_direction` is empty (label trigger or bare slash command)
- **THEN** the prompt section SHALL have no effect on agent behaviour and the existing research-recommendation handling SHALL apply unchanged

