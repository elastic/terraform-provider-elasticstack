# `ci-reproducer-factory-issue-intake` — reproducer-factory agentic workflow issue intake

Workflow implementation: repository-authored source under `.github/workflows-src/reproducer-factory-issue/`, compiled to `.github/workflows/reproducer-factory-issue.lock.yml`.

## Purpose

Define requirements for a GitHub Agentic Workflow that reacts to trusted GitHub issues labeled `reproducer-factory` and attempts to reproduce the described failure condition as a passing acceptance test, producing a sticky comment and optionally a pull request depending on the outcome.

## ADDED Requirements

### Requirement: Workflow source is repository-authored and generated
The repository SHALL define the `reproducer-factory` issue-intake automation as a repository-authored GitHub Agentic Workflow source under `.github/workflows-src/reproducer-factory-issue/` that generates checked-in workflow artifacts under `.github/workflows/`: the compiled `.github/workflows/reproducer-factory-issue.lock.yml` from `gh aw compile`. Contributors SHALL NOT hand-edit those generated files; they SHALL be regenerated with repository workflow tooling (`make workflow-generate`). Deterministic GitHub-script logic used for trigger qualification, dispatch input validation, comment-history capture, or context normalisation SHALL be factored into repository-local helper code under `.github/workflows-src/lib/` that can be tested independently.

#### Scenario: Maintainer updates the workflow source
- **WHEN** maintainers modify the `reproducer-factory` issue-intake workflow
- **THEN** the authored source SHALL live under `.github/workflows-src/reproducer-factory-issue/`
- **AND** the generated `.github/workflows/reproducer-factory-issue.lock.yml` SHALL be checked in and SHALL match output from the repository generation commands

#### Scenario: Deterministic logic is tested outside the workflow wrapper
- **WHEN** maintainers validate trigger qualification or context normalisation for `reproducer-factory`
- **THEN** the repository SHALL support focused tests for the extracted helper logic without requiring execution of the compiled workflow

### Requirement: Workflow activates the reproduction agent only for qualifying triggers
The workflow SHALL subscribe to GitHub `issues.opened`, `issues.labeled`, and `workflow_dispatch` events. For issue-event intake, eligible triggers SHALL include `issues.labeled` when the newly applied label is exactly `reproducer-factory`, and `issues.opened` when the issue already includes the `reproducer-factory` label at creation time. For dispatch intake, the workflow SHALL accept a typed `issue_number` input (and an optional `source_workflow` input for traceability) and SHALL treat dispatch as eligible when the input identifies one issue in the current repository.

#### Scenario: Label applied after issue creation
- **WHEN** an `issues.labeled` event is received and `github.event.label.name` is `reproducer-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the reproduction agent

#### Scenario: Issue opens with the trigger label already present
- **WHEN** an `issues.opened` event is received and the issue's initial labels include `reproducer-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the reproduction agent

#### Scenario: Non-trigger issue event is ignored
- **WHEN** an `issues` event is received without the `reproducer-factory` label in the qualifying position for that event type
- **THEN** the workflow SHALL NOT activate the reproduction agent for that event

#### Scenario: Internal workflow dispatch requests reproduction
- **WHEN** the workflow is triggered by `workflow_dispatch` with a valid `issue_number` for the current repository
- **THEN** the workflow SHALL treat that dispatch as eligible to activate the reproduction agent

### Requirement: Trigger actor must be trusted for issue-event intake
For issue-event intake, before agent activation the workflow SHALL determine whether the triggering actor is trusted. The actor SHALL be trusted if the sender is `github-actions[bot]`; otherwise the workflow SHALL query repository collaborator permissions and SHALL require effective repository permission `write`, `maintain`, or `admin`. This trust policy SHALL apply only to issue-event intake and SHALL NOT be required for repository-authored `workflow_dispatch` intake.

#### Scenario: Maintainer applies the label
- **WHEN** a human actor triggers an otherwise eligible `reproducer-factory` issue event and the repository permission lookup returns `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL treat the trigger as trusted

#### Scenario: Untrusted actor attempts to trigger automation
- **WHEN** a non-bot actor triggers an otherwise eligible `reproducer-factory` issue event and the repository permission lookup does not return `write`, `maintain`, or `admin`
- **THEN** the workflow SHALL skip agent activation

#### Scenario: Internal dispatch bypasses issue-event trust lookup
- **WHEN** the workflow is triggered by repository-authored `workflow_dispatch`
- **THEN** the workflow SHALL NOT require the issue-event actor trust check as a prerequisite for activation

### Requirement: Workflow suppresses duplicate linked pull requests
Before running the reproduction agent, the pre-activation job SHALL check for an existing open pull request linked to the triggering issue. A PR SHALL be considered a match when ALL of the following hold: it is open, it carries the `reproducer-factory` label, its head branch is `reproducer-factory/issue-{n}`, and its body contains `Related to #N` (where N is the issue number). If a matching open PR is found, the workflow SHALL skip agent activation and emit `noop` instead.

#### Scenario: Existing linked PR prevents a duplicate issue-event run
- **WHEN** an eligible `reproducer-factory` issue event fires for an issue that already has an open PR carrying the `reproducer-factory` label with `Related to #N` in its body on branch `reproducer-factory/issue-{n}`
- **THEN** the workflow SHALL NOT activate the reproduction agent
- **AND** the workflow SHALL emit `noop`

#### Scenario: Unrelated PR does not block issue intake
- **WHEN** the triggering issue has open PRs that do not carry the `reproducer-factory` label, do not use branch `reproducer-factory/issue-{n}`, or do not include `Related to #N`
- **THEN** the workflow SHALL proceed normally to agent activation

### Requirement: Workflow normalises issue intake context across entry modes
The workflow SHALL normalise issue intake into downstream-consumable outputs that include the resolved issue number, title, body, intake mode, source workflow (if any), gate reason, and the captured comment history so that the agent prompt and downstream steps do not depend directly on `github.event.issue.*`. For dispatch intake, the workflow SHALL fetch the live issue body and title from GitHub rather than trusting text passed through dispatch inputs.

#### Scenario: Issue-event intake exposes normalised outputs
- **WHEN** the workflow is triggered from an eligible issue event
- **THEN** the workflow SHALL publish normalised outputs for the resolved issue number, title, body, intake mode, gate reason, and comment history

#### Scenario: Dispatch intake exposes normalised outputs
- **WHEN** the workflow is triggered from `workflow_dispatch`
- **THEN** the workflow SHALL fetch the live issue from GitHub and publish the same normalised outputs

### Requirement: Workflow sanitises HTML comments from agent input context
Before writing the `issue_body.md` and `issue_comments.md` context files for the agent, the workflow SHALL strip all HTML comments from the issue body and from each human-authored comment using the shared `ci-html-comment-sanitisation` helpers.

#### Scenario: Agent receives clean context
- **WHEN** the `reproducer-factory` workflow runs for an issue whose body contains an injected HTML comment
- **THEN** the `issue_body.md` file written for the agent SHALL NOT contain that comment

### Requirement: Workflow captures human-authored comment history for the agent
Before agent activation, the workflow SHALL capture all comments on the triggering issue in chronological order, filtered to human-authored comments only. Comments authored by `github-actions[bot]` and other known automation bots SHALL be excluded.

#### Scenario: Issue has prior comments
- **WHEN** an eligible `reproducer-factory` event fires on an issue with human comments
- **THEN** the workflow SHALL include the chronological human comment history in the normalised intake context delivered to the agent

#### Scenario: Bot comments are excluded
- **WHEN** capturing comment history for the agent
- **THEN** the captured comments SHALL exclude `github-actions[bot]` and similar automation bots

### Requirement: Workflow removes the factory trigger label in pre-activation when the agent proceeds
The workflow SHALL include a deterministic pre-activation step that removes the `reproducer-factory` label from the triggering issue when the agent gate passes (eligible qualifying issue event, trusted actor, no duplicate PR). The label removal SHALL reuse the shared `remove-trigger-label` library helper. Dispatch-targeted issues SHALL NOT require the label to be present, and the workflow SHALL NOT treat its absence as an error.

#### Scenario: Label removed only when agent gate passes
- **WHEN** pre-activation determines the reproduction agent SHALL run for the issue
- **THEN** the remove-label step SHALL run and SHALL attempt to remove `reproducer-factory` from that issue

#### Scenario: Dispatch-targeted issue has no `reproducer-factory` label
- **WHEN** the workflow is triggered by `workflow_dispatch` for an issue that does not carry the `reproducer-factory` label
- **THEN** the workflow SHALL continue normally

### Requirement: Workflow enforces single-session-per-issue concurrency
The workflow SHALL declare GitHub Actions concurrency keyed by the resolved issue number such that at most one `reproducer-factory` run SHALL execute per issue at any time. Newly arriving triggers SHALL be queued rather than cancelling the in-flight run.

#### Scenario: Two triggers fire for the same issue in rapid succession
- **WHEN** a second qualifying trigger arrives for an issue while a run is already in flight
- **THEN** the second run SHALL be queued and SHALL NOT execute concurrently with the first

### Requirement: Workflow time-boxes the reproduction session and survives partial completion
The workflow SHALL set a job-level `timeout-minutes` of 65 minutes. The agent prompt SHALL communicate a 55-minute self-budget and SHALL instruct the agent to reserve the final minutes for emitting the reproduction comment. If running short on time, the agent SHALL prefer emitting a partial-but-valid comment over emitting `noop`.

#### Scenario: Maintainer inspects compiled workflow timeout
- **WHEN** maintainers inspect the compiled `reproducer-factory-issue.lock.yml` workflow
- **THEN** the agent job SHALL declare `timeout-minutes: 65`

#### Scenario: Agent prompt communicates the self-budget
- **WHEN** maintainers inspect the agent prompt body
- **THEN** the prompt SHALL state the 55-minute reproduction self-budget
- **AND** the prompt SHALL state that the agent SHALL prefer a partial-but-valid comment over `noop` when running short on time

### Requirement: Workflow provides an Elastic Stack acceptance-test environment
The reproduction agent SHALL have access to a live Elastic Stack for running acceptance tests, using the same environment configuration as the `code-factory` workflow. The agent prompt SHALL document the connection parameters: `ELASTICSEARCH_ENDPOINTS=http://host.docker.internal:9200`, `ELASTICSEARCH_USERNAME=elastic`, `ELASTICSEARCH_PASSWORD=password`, and `KIBANA_ENDPOINT=http://host.docker.internal:5601`. The network allow-list SHALL include `go` to support `go test` downloads.

#### Scenario: Agent runs an acceptance test
- **WHEN** the reproduction agent writes a `TestAccReproduceIssue{N}` test and runs it
- **THEN** the test SHALL be able to reach the Elastic Stack at the documented endpoints

#### Scenario: Maintainer inspects workflow network policy
- **WHEN** maintainers inspect the authored workflow frontmatter
- **THEN** `network.allowed` SHALL include `go`

### Requirement: Agent has structured access to Elastic documentation
The workflow SHALL configure the Elastic docs MCP server as an HTTP MCP server so the agent can query Elastic documentation while investigating the reported bug. The workflow frontmatter SHALL include `www.elastic.co` in `network.allowed` and SHALL declare an `mcp-servers.elastic-docs` entry pointing to `https://www.elastic.co/docs/_mcp/`. If the MCP tools are unavailable the agent SHALL proceed from the issue content alone.

#### Scenario: Agent investigates an unfamiliar Elastic API feature
- **WHEN** the issue references an Elastic API endpoint or feature
- **THEN** the agent SHOULD use `search_docs` or `find_related_docs` to locate relevant documentation before writing the test

#### Scenario: Elastic docs MCP server is unavailable
- **WHEN** the elastic-docs MCP tools return an error or are unreachable
- **THEN** the agent SHALL proceed with investigation using available information
- **AND** the agent SHALL NOT emit `noop` solely because the docs MCP is unavailable

### Requirement: Agent always emits a reproduction comment
Regardless of outcome, the agent SHALL always emit exactly one `update-reproducer-comment` safe-output operation conforming to the `ci-reproducer-factory-comment-format` capability. If the agent cannot make meaningful progress (empty issue, no reproducible scenario), it SHALL emit `noop` with a concise explanation instead.

#### Scenario: Bug is reproduced
- **WHEN** the agent successfully writes and runs a passing `TestAccReproduceIssue{N}` test
- **THEN** the agent SHALL emit `update-reproducer-comment` with an outcome-A comment body
- **AND** the agent SHALL also emit `create-pull-request`

#### Scenario: Bug cannot be reproduced
- **WHEN** the agent cannot write a test that demonstrates the reported failure condition
- **THEN** the agent SHALL emit `update-reproducer-comment` with an outcome-B comment body
- **AND** the agent SHALL NOT emit `create-pull-request`

#### Scenario: Bug appears already fixed
- **WHEN** the agent writes a test and runs it without `ExpectError` and the test passes cleanly
- **THEN** the agent SHALL emit `update-reproducer-comment` with an outcome-C comment body
- **AND** the agent SHALL NOT emit `create-pull-request`

### Requirement: Agent creates exactly one linked pull request when reproduction succeeds
When the reproduction test passes, the agent SHALL create exactly one pull request on branch `reproducer-factory/issue-{n}` labeled `reproducer-factory`. The PR body SHALL include `Related to #N` to establish deterministic linkage for duplicate-PR suppression on future runs. The PR SHALL contain only the reproduction test file and no other changes. `Related to` is used rather than a closing keyword because the reproduction does not resolve the issue — it confirms it.

#### Scenario: Eligible issue-event intake creates a reproduction PR
- **WHEN** the reproduction test passes for an eligible issue event
- **THEN** the agent SHALL emit `create-pull-request` with branch `reproducer-factory/issue-{n}` and body containing `Related to #N`

#### Scenario: No PR is created when reproduction fails
- **WHEN** the agent reaches outcome B (cannot reproduce) or outcome C (appears fixed)
- **THEN** the agent SHALL NOT emit `create-pull-request`

### Requirement: Agent places the reproduction test in the correct file location
The agent SHALL place the `TestAccReproduceIssue{N}` test function in `internal/acctest/reproductions/issue_{n}_acc_test.go` by default. If the issue clearly identifies a single Terraform resource — by TF type name (e.g. `elasticstack_kibana_alerting_rule`) or by an unambiguous human description (e.g. "alerting rule resource", "kibana dashboard", "SLO resource") — the agent SHALL instead place the test in the resource's own package (e.g. `internal/kibana/alertingrule/issue_{n}_acc_test.go`). When the bug involves multiple resources, is provider-level, or cannot be clearly attributed to one resource, the agent SHALL use the fallback path.

#### Scenario: Issue names a specific TF resource type
- **WHEN** the issue body or title contains a string matching a known `elasticstack_*` resource type
- **THEN** the agent SHALL place the test in that resource's package directory

#### Scenario: Issue uses a human description of a resource
- **WHEN** the issue body or title uses a human description that unambiguously maps to one resource (e.g. "the dashboard resource", "kibana SLOs", "alerting rules")
- **THEN** the agent SHALL place the test in that resource's package directory

#### Scenario: Issue is ambiguous or provider-level
- **WHEN** the issue does not clearly identify a single resource, or the bug is provider-level or spans multiple resources
- **THEN** the agent SHALL place the test in `internal/acctest/reproductions/`

### Requirement: Workflow status comments on the issue include the run link
The workflow SHALL set `status-comment: true` in the top-level `on:` configuration so the activation job posts a status comment on the triggering issue when the run starts and updates it when the run completes.

#### Scenario: Status comment enabled
- **WHEN** maintainers inspect the authored workflow `on:` frontmatter
- **THEN** it SHALL include `status-comment: true`

### Requirement: Workflow frontmatter allows required agent ecosystems
The workflow SHALL declare a network policy that allows the default allowlist plus the Node and Go ecosystems, `elastic.litellm-prod.ai` for the Claude engine proxy, and `www.elastic.co` for the Elastic docs MCP server.

#### Scenario: Maintainer inspects workflow frontmatter
- **WHEN** maintainers inspect the authored workflow frontmatter
- **THEN** `network.allowed` SHALL include `defaults`, `node`, `go`, `elastic.litellm-prod.ai`, and `www.elastic.co`
