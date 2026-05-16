# ci-semantic-refactor-workflow Specification

## Purpose
TBD - created by archiving change semantic-refactor-workflow. Update Purpose after archive.
## Requirements
### Requirement: Workflow artifacts and compilation
The semantic refactor workflow SHALL be authored as a GitHub Agentic Workflow markdown source under `.github/workflows-src/` and SHALL include generated workflow artifacts under `.github/workflows/`, including a compiled `.lock.yml` derived from the authored source. The repository-authored source SHALL identify `https://github.com/github/gh-aw/blob/main/.github/workflows/semantic-function-refactor.md` as its upstream baseline. Contributors SHALL NOT hand-edit the generated workflow artifacts.

#### Scenario: Source and generated artifacts stay paired
- **WHEN** maintainers change semantic refactor workflow behavior
- **THEN** the authored workflow source, generated workflow markdown, and compiled lock artifact SHALL match the repository compiler output

#### Scenario: Upstream workflow source remains referenced
- **WHEN** maintainers review or update the repository-authored semantic refactor workflow source
- **THEN** the workflow source SHALL continue to reference `https://github.com/github/gh-aw/blob/main/.github/workflows/semantic-function-refactor.md` as the upstream workflow source

### Requirement: Scheduled and manual triggering
The semantic refactor workflow SHALL support scheduled daily execution and manual `workflow_dispatch` execution.

#### Scenario: Scheduled run is supported
- **WHEN** the semantic refactor workflow reaches its configured daily schedule
- **THEN** the workflow SHALL start a semantic refactor analysis run subject to the pre-activation gate

#### Scenario: Manual dispatch is supported
- **WHEN** a maintainer triggers the workflow with `workflow_dispatch`
- **THEN** the workflow SHALL start a semantic refactor analysis run subject to the pre-activation gate

### Requirement: LiteLLM-backed engine configuration
The semantic refactor workflow SHALL run the agent through the repository's LiteLLM-backed Claude engine configuration, using model `llm-gateway/claude-sonnet-4-6`, `ANTHROPIC_BASE_URL` set to the Elastic LiteLLM endpoint, `ANTHROPIC_API_KEY` sourced from `CLAUDE_LITELLM_PROXY_API_KEY`, and a `network.allowed` contract that permits access to `elastic.litellm-prod.ai`.

#### Scenario: Authored workflow uses LiteLLM engine settings
- **WHEN** maintainers inspect the authored semantic refactor workflow source
- **THEN** the engine configuration SHALL specify Claude with model `llm-gateway/claude-sonnet-4-6`, the Elastic LiteLLM base URL, a `network.allowed` entry for `elastic.litellm-prod.ai`, and the `CLAUDE_LITELLM_PROXY_API_KEY` secret

#### Scenario: Compiled lock preserves LiteLLM execution settings
- **WHEN** the workflow source is compiled
- **THEN** the compiled lock artifact SHALL preserve the LiteLLM model, base URL, allowlisted host `elastic.litellm-prod.ai`, and secret-backed API key for agent execution

### Requirement: Deterministic semantic-refactor issue-slot gating
Before agent analysis begins, deterministic repository-authored steps SHALL compute available issue slots by counting open GitHub issues with the `semantic-refactor` label and subtracting that count from a workflow-configured issue cap of `3`. The workflow SHALL expose the open-issue count, available slot count, and gate reason through pre-activation outputs, and it SHALL skip the agent job when the available slot count is zero.

#### Scenario: Open semantic-refactor issues leave slots available
- **WHEN** fewer than three open issues carry the `semantic-refactor` label
- **THEN** the workflow SHALL expose a positive `issue_slots_available` value and proceed to agent analysis

#### Scenario: Open semantic-refactor issues reach the cap
- **WHEN** three or more open issues carry the `semantic-refactor` label
- **THEN** the workflow SHALL expose `issue_slots_available` as zero and SHALL skip the agent job

#### Scenario: Agent receives deterministic slot context
- **WHEN** the workflow proceeds to agent analysis
- **THEN** the prompt SHALL tell the agent the open `semantic-refactor` issue count, available issue slots, and gate reason from pre-activation outputs

### Requirement: Existing semantic-refactor issues remain open
The semantic refactor workflow SHALL NOT close existing open `semantic-refactor` issues as part of routine scheduled or manual analysis. Existing open issues SHALL count against the issue-slot cap until maintainers close or relabel them.

#### Scenario: Existing issues are not closed before analysis
- **WHEN** a semantic refactor workflow run starts and open `semantic-refactor` issues already exist
- **THEN** the workflow SHALL count those issues for capacity and SHALL NOT close them merely because a new analysis run is starting

### Requirement: Analysis scope is constrained to Go semantic refactor opportunities
The workflow SHALL instruct the agent to analyze non-test Go source files, with primary focus on repository implementation code, and identify high-value semantic refactor opportunities such as misplaced functions, duplicate or near-duplicate functions, scattered helpers, or cohesive function clusters that should be extracted or moved. The workflow SHALL exclude test files, generated artifacts, workflow files, vendored dependencies, and non-Go files from issue findings.

#### Scenario: Non-test Go files are analyzed
- **WHEN** the workflow performs semantic refactor analysis
- **THEN** it SHALL direct the agent to inspect non-test `.go` source files and collect relevant function and method organization evidence

#### Scenario: Noisy file classes are excluded
- **WHEN** the workflow encounters tests, generated files, workflow definitions, vendored dependencies, or non-Go files during analysis
- **THEN** it SHALL exclude those files from semantic refactor issue findings

### Requirement: One issue per semantic refactor opportunity
The workflow SHALL create at most one issue per distinct semantic refactor opportunity and SHALL NOT create more issues in a run than the computed number of available issue slots. The workflow SHALL create no more than three semantic refactor issues in a run even when no matching issues are already open.

#### Scenario: Multiple opportunities are capped by available slots
- **WHEN** the workflow identifies more actionable semantic refactor opportunities than the computed number of available issue slots
- **THEN** it SHALL create issues only for the highest-priority opportunities up to the available slot count

#### Scenario: Distinct opportunities are not bundled
- **WHEN** the workflow creates an issue for an actionable semantic refactor opportunity
- **THEN** that issue SHALL describe exactly one distinct opportunity or tightly related refactor cluster rather than bundling unrelated findings together

### Requirement: Semantic-refactor issue contents are actionable
Each semantic refactor issue created by the workflow SHALL include a concise summary, concrete affected locations, the observed organization or duplication problem, impact assessment, and actionable refactoring guidance sufficient for a follow-up coding agent or maintainer to act on the issue.

#### Scenario: Issue contains evidence and guidance
- **WHEN** the workflow creates a semantic refactor issue
- **THEN** the issue body SHALL include affected file paths or symbols, evidence for the finding, impact, and recommended refactoring steps

#### Scenario: Issue titles and labels identify the workflow output
- **WHEN** the workflow creates a semantic refactor issue
- **THEN** the issue SHALL carry the configured title prefix `[semantic-refactor] ` and the label `semantic-refactor`
- **AND** the issue SHALL NOT include the `code-factory` label

### Requirement: Created semantic-refactor issues are explicitly dispatched to `code-factory`
After safe-output issue creation completes, the workflow SHALL explicitly dispatch the `code-factory` workflow once for each issue created in the current run rather than relying on producer-side `code-factory` labels to trigger implementation intake.

#### Scenario: One created issue dispatches one implementation run
- **WHEN** the workflow creates one semantic refactor issue in a run
- **THEN** it SHALL dispatch exactly one `code-factory` workflow run for that issue

#### Scenario: Three created issues dispatch three implementation runs
- **WHEN** the workflow creates three semantic refactor issues in a run
- **THEN** it SHALL dispatch exactly three independent `code-factory` workflow runs, one per created issue

