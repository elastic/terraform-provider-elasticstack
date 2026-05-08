## ADDED Requirements

### Requirement: Agent uses the implementation-research block as the exclusive scope when present
When the triggering issue's body contains a block conforming to the `ci-implementation-research-block-format` capability — that is, a region delimited by `<!-- implementation-research:start -->` and `<!-- implementation-research:end -->` — the `change-factory` agent SHALL treat that block as the exclusive authoritative source for proposal scope. The agent SHALL adopt the block's `### Recommendation` as the spine of the OpenSpec proposal it authors, SHALL carry the block's `### Open questions` verbatim into the resulting `design.md` (under a section such as `## Open questions`), and SHALL use the block's `### Approaches considered` for context only — the agent SHALL NOT re-explore alternative approaches the block has already evaluated unless the issue body or its comments contradict the recommendation. When no such block is present in the issue body, the agent SHALL retain the existing behavior of treating the issue title and body as the authoritative source.

#### Scenario: Issue body contains a research block
- **WHEN** a `change-factory` run starts for an issue whose body contains a well-formed implementation-research block
- **THEN** the agent SHALL adopt the block's `### Recommendation` as the chosen approach and use it as the spine of `proposal.md`
- **AND** the agent SHALL copy the block's `### Open questions` into the resulting `design.md` (e.g. as a `## Open questions` section)
- **AND** the agent SHALL NOT re-explore the alternative approaches enumerated in `### Approaches considered`, treating them as already-evaluated context

#### Scenario: Issue body contains no research block
- **WHEN** a `change-factory` run starts for an issue whose body does not contain an implementation-research block
- **THEN** the agent SHALL author the proposal using only the issue title and body as the authoritative source, exactly as it does today

#### Scenario: Issue body contains a research block but later comments contradict the recommendation
- **WHEN** a `change-factory` run starts for an issue whose body contains a research block and whose visible context (issue body content outside the block, or any signal the agent has from prior workflow runs) contradicts the block's recommendation
- **THEN** the agent SHALL prefer the contradicting signal and SHALL note the disagreement in the proposal artifacts (for example, in `design.md` under a section explaining the deviation)

### Requirement: Agent prompt documents implementation-research block awareness
The `change-factory` workflow's authored prompt SHALL include explicit instructions describing the implementation-research block: its marker comments, the location of `### Recommendation` and `### Open questions`, the rule that a present block is the exclusive scope source, and the rule that an absent block falls back to today's title-and-body behavior. The prompt SHALL NOT instruct the agent to add, modify, or remove the block itself — block management belongs to the `research-factory` workflow.

#### Scenario: Maintainer inspects the change-factory prompt
- **WHEN** maintainers inspect the authored `change-factory-issue` workflow prompt
- **THEN** the prompt SHALL describe the `<!-- implementation-research:start --> ... <!-- implementation-research:end -->` markers
- **AND** the prompt SHALL state that, when the markers are present, the block's `### Recommendation` and `### Open questions` are the authoritative inputs for the proposal
- **AND** the prompt SHALL state that, when the markers are absent, the existing title-and-body-authoritative behavior applies unchanged

#### Scenario: Change-factory does not mutate the research block
- **WHEN** the `change-factory` agent runs against an issue with an implementation-research block
- **THEN** the agent SHALL NOT emit `update-issue` operations that modify the research block
- **AND** the agent SHALL NOT add, remove, or rewrite the `<!-- implementation-research:* -->` markers
