# ci-implementation-research-block-format Specification

## Purpose
TBD - created by archiving change research-factory-workflow. Update Purpose after archive.
## Requirements
### Requirement: Block is delimited by stable HTML-comment markers
The implementation-research block SHALL be delimited by exactly two HTML comments: an opening marker `<!-- implementation-research:start -->` and a closing marker `<!-- implementation-research:end -->`. The markers SHALL appear at the start of their own lines. Issue bodies SHALL contain at most one such block. Producers (the `research-factory` workflow) and consumers (the `change-factory` workflow) SHALL identify the block by these markers and SHALL NOT rely on framework-generated marker formats such as `gh-aw-begin-*`.

#### Scenario: Body contains a research block
- **WHEN** an issue body contains the markers `<!-- implementation-research:start -->` and `<!-- implementation-research:end -->`
- **THEN** the content between (and including) those markers SHALL constitute the implementation-research block
- **AND** there SHALL be exactly one such pair in the body

#### Scenario: Body contains no research block
- **WHEN** an issue body does not contain the opening marker `<!-- implementation-research:start -->`
- **THEN** the body SHALL be treated as having no implementation-research block

### Requirement: Block opens with a provenance header
The block SHALL begin (immediately after the opening marker) with a `## Implementation research` heading followed by a provenance header that records the run timestamp, a link to the workflow run that authored the block, and an explicit notice that edits inside the block are read as input on the next run but are not preserved verbatim.

#### Scenario: Maintainer inspects an authored block
- **WHEN** a maintainer reads a research block produced by the workflow
- **THEN** the block SHALL contain a `## Implementation research` heading
- **AND** the block SHALL include a provenance line stating the timestamp and a link to the workflow run that produced it
- **AND** the block SHALL include a notice stating that edits inside the block are not preserved verbatim and that durable feedback should be provided via comments or edits outside the block

### Requirement: Block contains the mandatory research subsections
The block SHALL contain the following subsections in order, each introduced by a level-3 heading:

1. `### Problem framing` — one or more paragraphs restating the requested change in concrete terms.
2. `### Approaches considered` — containing two or more level-4 (`#### `) child headings, each describing a distinct candidate approach with its sketch, Terraform shape (when applicable), Elastic API surface (when applicable), and pros / cons.
3. `### Recommendation` — naming exactly one approach from the previous section as the chosen spine, with a brief rationale.
4. `### Open questions` — a (possibly empty) bullet list of questions whose answers would change the recommendation or the proposal scope.
5. `### Out of scope` — a (possibly empty) bullet list of items the recommendation explicitly excludes.
6. `### References` — a list of consulted sources, including elastic-docs URLs and repository paths inspected during research.

#### Scenario: Block contains the required subsections
- **WHEN** a maintainer or downstream consumer reads a research block produced by the workflow
- **THEN** the block SHALL contain the headings `### Problem framing`, `### Approaches considered`, `### Recommendation`, `### Open questions`, `### Out of scope`, and `### References` in that order

#### Scenario: Approaches considered contains at least two approaches
- **WHEN** the block's `### Approaches considered` section is inspected
- **THEN** it SHALL contain two or more `#### ` child headings, each naming a distinct candidate approach

#### Scenario: Recommendation names exactly one approach
- **WHEN** the block's `### Recommendation` section is inspected
- **THEN** it SHALL identify exactly one of the approaches enumerated under `### Approaches considered` as the chosen approach
- **AND** it SHALL include a brief rationale for that choice

### Requirement: Block is regenerated each run; outside content is preserved
The block SHALL be the only mutable region of the issue body managed by the workflow. Each successful research run SHALL replace the entire block contents with a freshly synthesized block. Content of the issue body outside the markers SHALL be preserved byte-for-byte by the producer relative to the pre-run original issue content (the issue body with any prior block removed). The block SHALL NOT contain hidden state intended for machine consumption beyond the markers and the human-readable subsections defined above.

#### Scenario: Original issue content is preserved across re-runs
- **WHEN** the workflow re-runs for an issue whose body contains both an original problem statement (before the markers) and a prior research block
- **THEN** the new body SHALL preserve the original problem statement byte-for-byte
- **AND** the new body SHALL contain exactly one research block, fully regenerated

#### Scenario: Block is the only region the workflow rewrites
- **WHEN** the workflow updates an issue body
- **THEN** the only difference between the prior body and the new body, after both have their research blocks removed, SHALL be empty (modulo trailing whitespace normalization)

### Requirement: Prior block contents and human-authored comments are read as input on the next run
On any run where a prior block exists, the producer SHALL read the prior block contents and the human-authored comment history as additional inputs alongside the original issue content. The producer SHALL NOT treat the prior block as authoritative output to preserve verbatim; rather it SHALL integrate the evidence the prior block carries (recommendation, open questions, references, and any human edits made inside the block) into the synthesis of the new block.

#### Scenario: User edits a recommendation inside the block between runs
- **WHEN** a user edits the `### Recommendation` section of a prior block (for example, replacing the chosen approach with their own reasoning) and re-applies the `research-factory` label
- **THEN** the next run SHALL read the user's edit as evidence
- **AND** the next run SHALL produce a freshly synthesized block that integrates that evidence rather than preserving the user's edit verbatim

#### Scenario: User answers an open question via a comment
- **WHEN** a user posts an issue comment that resolves one of the prior block's open questions and re-applies the `research-factory` label
- **THEN** the next run SHALL read the comment
- **AND** the next run's `### Open questions` section SHALL reflect the resolution (the question is removed or restated based on the new information)

