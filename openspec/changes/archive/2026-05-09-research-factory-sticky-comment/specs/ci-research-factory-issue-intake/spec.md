# ci-research-factory-issue-intake Specification

## ADDED Requirements

### Requirement: Workflow sanitizes HTML comments from agent input context
Before writing the `issue_body.md` and `issue_comments.md` context files for the agent, the workflow SHALL strip all HTML comments from the issue body and from each human-authored comment using the shared `ci-html-comment-sanitisation` helpers. Bot-authored comments SHALL already be excluded from `issue_comments.md` by the existing filter.

#### Scenario: Agent receives clean context
- **WHEN** the `research-factory` workflow runs for an issue whose body contains an injected `<!-- fake-marker -->` comment
- **THEN** the `issue_body.md` file written for the agent SHALL NOT contain that comment
- **AND** the agent SHALL therefore be unable to read or act on the injected marker

#### Scenario: Human comments with HTML comments are cleaned
- **WHEN** a human comment on the issue contains an HTML comment
- **THEN** the sanitised comment text delivered to the agent SHALL have that comment removed

### Requirement: Workflow fetches prior research comment as agent input
On any run where a prior research comment exists on the issue, the workflow SHALL fetch that comment (identified by `github-actions[bot]` author and the `<!-- gha-research-factory -->` marker) and provide its full body to the agent as a separate context file or prompt section. The prior research comment SHALL NOT be passed through `stripHtmlComments` because it is trusted bot-authored output. The agent SHALL read the prior comment alongside the sanitised issue body and sanitised human comment history.

#### Scenario: Prior research comment is provided to agent verbatim
- **WHEN** the workflow re-runs for an issue that already has a research comment by `github-actions[bot]`
- **THEN** the workflow SHALL fetch that comment and provide it to the agent without HTML-comment stripping
- **AND** the agent SHALL receive the intact `<!-- gha-research-factory -->` marker and all prior research content

## MODIFIED Requirements

### Requirement: Agent emits a single research comment via custom safe-output script
When the deterministic gate passes and the agent completes its research, the agent SHALL emit a single `update_research_comment` safe-output operation whose `body` payload contains the research content conforming to the `ci-research-factory-comment-format` capability. The workflow SHALL define a custom `safe-outputs.scripts` entry named `update-research-comment` that creates or updates an issue comment authored by `github-actions[bot]`. If an existing comment by `github-actions[bot]` containing the marker `<!-- gha-research-factory -->` is found on the issue, the script SHALL update that comment; otherwise it SHALL create a new comment. The agent SHALL NOT emit `update_issue`, `add-comment`, or any other safe-output operation as part of its research output.

#### Scenario: Agent produces research on a fresh issue
- **WHEN** the workflow runs for an eligible issue with no prior research comment
- **THEN** the agent SHALL emit one `update_research_comment` operation
- **AND** the custom script SHALL create a new comment on the issue
- **AND** that comment SHALL contain `<!-- gha-research-factory -->` as its first line

#### Scenario: Agent regenerates an existing research comment
- **WHEN** the workflow runs for an eligible issue that already has a research comment by `github-actions[bot]`
- **THEN** the agent SHALL emit one `update_research_comment` operation
- **AND** the custom script SHALL update the existing comment in place
- **AND** the issue SHALL NOT gain an additional research comment

#### Scenario: Agent times out before reaching a confident recommendation
- **WHEN** the agent's self-budget expires before research is complete
- **THEN** the agent SHALL emit a partial-but-valid research comment with explicit unanswered open questions
- **AND** the agent SHALL NOT emit `noop` solely because research is partial

### Requirement: Workflow time-boxes the research session and survives partial completion
The workflow SHALL set a job-level `timeout-minutes` of 35 minutes. The agent prompt SHALL communicate a 25-minute self-budget to the agent and SHALL instruct it to reserve the final minutes of the budget for emitting its research comment. The prompt SHALL further instruct the agent that, if research time runs short, it SHALL prefer emitting a partial-but-valid research comment (with explicit unanswered open questions) over emitting `noop`.

#### Scenario: Maintainer inspects compiled workflow timeout
- **WHEN** maintainers inspect the compiled `research-factory-issue.md` workflow
- **THEN** the agent job SHALL declare `timeout-minutes: 35`

#### Scenario: Agent prompt communicates the self-budget
- **WHEN** maintainers inspect the agent prompt body
- **THEN** the prompt SHALL state the 25-minute research self-budget
- **AND** the prompt SHALL state that the agent SHALL prefer a partial-but-valid research comment over `noop` when running short on time

### Requirement: Workflow remains research-only and does not write code
The `research-factory` workflow SHALL NOT implement provider, CI, or documentation behavior, SHALL NOT open pull requests, and SHALL NOT modify repository files. Its only durable output SHALL be a single `update_research_comment` safe-output operation executed by the custom `update-research-comment` script, producing a comment conforming to the `ci-research-factory-comment-format` capability. The workflow SHALL NOT enable safe outputs that would permit creating pull requests, creating issues, or posting free-form comments beyond the framework's own `status-comment`.

#### Scenario: Maintainer inspects compiled workflow safe outputs
- **WHEN** maintainers inspect the compiled `research-factory-issue.md` workflow `safe-outputs:` block
- **THEN** it SHALL include a `scripts` entry named `update-research-comment`
- **AND** it SHALL NOT include `update-issue`, `create-pull-request`, `push-to-pull-request-branch`, `update-pull-request`, or `create-issue`
- **AND** it SHALL NOT include `add-comment`

#### Scenario: Issue requests provider implementation
- **WHEN** a qualifying issue describes a Terraform resource, data source, or other provider implementation
- **THEN** the agent SHALL produce a research comment describing approaches and open questions
- **AND** the agent SHALL NOT modify provider source, generated clients, or documentation
