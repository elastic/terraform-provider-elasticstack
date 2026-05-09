# ci-change-factory-issue-intake Specification

## ADDED Requirements

### Requirement: Workflow sanitizes HTML comments from agent input context
Before the `change-factory` agent reads the triggering issue body and human-authored comments, the workflow SHALL strip all HTML comments from that content using the shared `ci-html-comment-sanitisation` helpers. The sanitised issue body and comment history SHALL be exposed to the agent prompt. The research-factory sticky comment, when present, SHALL be extracted separately and SHALL NOT be passed through the stripping step because it is bot-authored trusted output.

#### Scenario: Agent receives clean context
- **WHEN** the `change-factory` workflow runs for an issue whose body contains an injected HTML comment
- **THEN** the agent SHALL receive a sanitised body with that comment removed

#### Scenario: Human comments with HTML comments are cleaned
- **WHEN** a human comment on the issue contains an HTML comment
- **THEN** the sanitised comment text delivered to the change-factory agent SHALL have that comment removed

## MODIFIED Requirements

### Requirement: Agent uses the implementation-research comment as the authoritative scope baseline when present
When the triggering issue has a comment authored by `github-actions[bot]` that contains the marker `<!-- gha-research-factory -->` and a heading `## Implementation research`, the `change-factory` agent SHALL treat that entire comment as the authoritative baseline for proposal scope. The agent SHALL adopt the comment's `### Recommendation` as the spine of the OpenSpec proposal it authors, SHALL carry the comment's `### Open questions` verbatim into the resulting `design.md` (under a section such as `## Open questions`), and SHALL use the comment's `### Approaches considered` for context only — the agent SHALL NOT re-explore alternative approaches the research has already evaluated. If the sanitised issue body or sanitised human comments contain explicit signals that contradict the research comment's recommendation, the agent SHALL prefer the contradicting signal and SHALL note the disagreement in the proposal artifacts. When no such comment exists on the issue, the agent SHALL retain the existing behavior of treating the issue title and body as the authoritative source.

#### Scenario: Issue has a research comment
- **WHEN** a `change-factory` run starts for an issue that has a bot-authored comment containing `<!-- gha-research-factory -->` and `## Implementation research`
- **THEN** the agent SHALL adopt the comment's `### Recommendation` as the chosen approach and use it as the spine of `proposal.md`
- **AND** the agent SHALL copy the comment's `### Open questions` into the resulting `design.md` (e.g. as a `## Open questions` section)
- **AND** the agent SHALL NOT re-explore the alternative approaches enumerated in `### Approaches considered`, treating them as already-evaluated context

#### Scenario: Issue has no research comment
- **WHEN** a `change-factory` run starts for an issue that does not have a bot-authored research comment
- **THEN** the agent SHALL author the proposal using only the issue title and body as the authoritative source, exactly as it does today

#### Scenario: Issue has a research comment but later comments contradict the recommendation
- **WHEN** a `change-factory` run starts for an issue that has a research comment and whose visible context (sanitised issue body or sanitised human comments) contradicts the comment's recommendation
- **THEN** the agent SHALL prefer the contradicting signal and SHALL note the disagreement in the proposal artifacts (for example, in `design.md` under a section explaining the deviation)

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
