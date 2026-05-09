# ci-research-factory-comment-format Specification

## Purpose
Define the format of the implementation-research output produced by the `research-factory` workflow as a dedicated sticky comment on the triggering issue, authored by `github-actions[bot]`. The comment is the sole durable output of a research run.

## Requirements

### Requirement: Comment is authored by github-actions[bot] and identified by a marker
The implementation-research output SHALL be a single issue comment authored by `github-actions[bot]`. The comment body SHALL begin with exactly the marker `<!-- gha-research-factory -->` on its own line. The marker serves as a filter key for downstream consumers (e.g., `change-factory`) to locate the research comment among other bot comments on the issue. There SHALL be at most one such research comment per issue. If multiple comments matching both the `github-actions[bot]` author and the `<!-- gha-research-factory -->` marker are found, the producer SHALL update the most recently created matching comment and SHALL NOT create an additional one.

#### Scenario: Research comment is created on a fresh issue
- **WHEN** the `research-factory` workflow runs for an issue with no prior research comment
- **THEN** a new comment SHALL be created by `github-actions[bot]`
- **AND** that comment SHALL begin with `<!-- gha-research-factory -->`
- **AND** the issue body SHALL NOT be modified

#### Scenario: Research comment is updated on re-run
- **WHEN** the `research-factory` workflow re-runs for an issue that already has a research comment
- **THEN** the existing comment SHALL be updated in place
- **AND** the issue SHALL NOT gain an additional research comment

#### Scenario: Multiple matching research comments exist
- **WHEN** the workflow finds more than one comment by `github-actions[bot]` containing the `<!-- gha-research-factory -->` marker
- **THEN** the producer SHALL update the most recently created matching comment
- **AND** the issue SHALL NOT gain an additional research comment

#### Scenario: Downstream consumer locates the research comment
- **WHEN** the `change-factory` workflow needs to read prior research
- **THEN** it SHALL locate the comment by filtering for `github-actions[bot]` author and the `<!-- gha-research-factory -->` marker

### Requirement: Comment opens with a provenance header
Immediately after the marker, the comment SHALL begin with a `## Implementation research` heading followed by a provenance header that records the run timestamp, a link to the workflow run that authored the comment, and an explicit notice that edits inside the comment are read as input on the next run but are not preserved verbatim.

#### Scenario: Maintainer inspects an authored comment
- **WHEN** a maintainer reads a research comment produced by the workflow
- **THEN** the comment SHALL contain a `## Implementation research` heading
- **AND** the comment SHALL include a provenance line stating the timestamp and a link to the workflow run that produced it
- **AND** the comment SHALL include a notice stating that edits inside the comment are not preserved verbatim and that durable feedback should be provided via issue comments or edits to the issue body

### Requirement: Comment contains the mandatory research subsections
The comment SHALL contain the following subsections in order, each introduced by a level-3 heading:

1. `### Problem framing` — one or more paragraphs restating the requested change in concrete terms.
2. `### Approaches considered` — containing two or more level-4 (`#### `) child headings, each describing a distinct candidate approach with its sketch, Terraform shape (when applicable), Elastic API surface (when applicable), and pros / cons.
3. `### Recommendation` — naming exactly one approach from the previous section as the chosen spine, with a brief rationale.
4. `### Open questions` — a (possibly empty) bullet list of questions whose answers would change the recommendation or the proposal scope.
5. `### Out of scope` — a (possibly empty) bullet list of items the recommendation explicitly excludes.
6. `### References` — a list of consulted sources, including elastic-docs URLs and repository paths inspected during research.

#### Scenario: Comment contains the required subsections
- **WHEN** a maintainer or downstream consumer reads a research comment produced by the workflow
- **THEN** the comment SHALL contain the headings `### Problem framing`, `### Approaches considered`, `### Recommendation`, `### Open questions`, `### Out of scope`, and `### References` in that order

#### Scenario: Approaches considered contains at least two approaches
- **WHEN** the comment's `### Approaches considered` section is inspected
- **THEN** it SHALL contain two or more `#### ` child headings, each naming a distinct candidate approach

#### Scenario: Recommendation names exactly one approach
- **WHEN** the comment's `### Recommendation` section is inspected
- **THEN** it SHALL identify exactly one of the approaches enumerated under `### Approaches considered` as the chosen approach
- **AND** it SHALL include a brief rationale for that choice

### Requirement: Comment contains a structured machine-readable JSON metadata block
After the `### References` section, the comment SHALL contain an HTML `<details>` element with a `<summary>` of exactly "🤖 Pipeline metadata". Inside the `<details>` element, the comment SHALL contain exactly one fenced JSON code block (language `json`) containing a machine-readable representation of the research. The JSON object SHALL conform to the following schema:

- `schema_version` (string, required): the version of this metadata schema.
- `recommendation` (object, required):
  - `spine` (string, required): a kebab-case identifier for the recommended approach.
  - `confidence` (string, optional, enum `["high","medium","low"]`): the agent's confidence in the recommendation.
  - `approach_index` (number, required): zero-based index of the chosen approach in the `Approaches considered` section.
- `open_questions` (array, optional): each item with:
  - `id` (string, required): a stable question identifier (e.g., `oq-1`, `oq-2`).
  - `text` (string, required): the question text.
  - `blocking` (boolean, required): whether resolving this question is a prerequisite for implementation.
- `affected_capabilities` (array of strings, optional): kebab-case capability identifiers impacted by the recommendation.
- `estimated_scope` (string, enum `["small","medium","large","unknown"]`): rough size of the implementation effort.
- `references` (array, optional): each item with:
  - `type` (string, enum `["elastic-docs","repo-path","issue","pr","external"]`).
  - `url` or `path` (string, required): the reference location.

The agent SHALL ensure that the JSON content is internally consistent with the human-readable subsections above it. The `<details>` element SHALL be closed by default so that human readers do not see the JSON unless they expand it.

#### Scenario: Comment contains valid JSON metadata
- **WHEN** a maintainer or downstream consumer inspects a research comment produced by the workflow
- **THEN** the comment SHALL contain a `<details>` element after `### References`
- **AND** inside that element there SHALL be a fenced JSON block conforming to the schema above
- **AND** the JSON `recommendation.spine` SHALL match the human-readable `### Recommendation` content

#### Scenario: JSON metadata is hidden from human readers by default
- **WHEN** a maintainer reads a research comment on GitHub
- **THEN** the JSON metadata SHALL be collapsed inside a `<details>` element
- **AND** it SHALL NOT be visible without clicking the disclosure triangle

#### Scenario: Downstream consumer parses JSON metadata
- **WHEN** a downstream workflow (e.g., a future classifier or change-factory enhancement) reads a research comment
- **THEN** it SHALL be able to extract the JSON block from the `<details>` element by parsing the comment body
- **AND** it SHALL be able to read `open_questions[].blocking` without regex-parsing headings

### Requirement: Comment is regenerated each run
Each successful research run SHALL replace the entire contents of the existing research comment with a freshly synthesized comment. When no prior comment exists, a new comment SHALL be created. The comment SHALL NOT contain hidden state intended for machine consumption beyond the `<!-- gha-research-factory -->` marker, the human-readable subsections defined above, and the structured JSON metadata block inside the `<details>` element.

#### Scenario: Comment is regenerated on re-run
- **WHEN** the workflow re-runs for an issue whose body already has a prior research comment
- **THEN** the prior comment SHALL be updated with new content
- **AND** the comment SHALL contain only the current run's research

#### Scenario: No hidden machine state outside declared blocks
- **WHEN** a maintainer inspects the raw markdown of a research comment
- **THEN** the only machine-readable content SHALL be:
  - the `<!-- gha-research-factory -->` marker
  - the structured JSON metadata block inside the `<details>` element
- **AND** the rest SHALL be human-readable markdown

### Requirement: Prior comment contents and human-authored comments are read as input on the next run
On any run where a prior research comment exists, the producer SHALL read the prior comment contents and the human-authored comment history as additional inputs alongside the original issue content. The producer SHALL NOT treat the prior comment as authoritative output to preserve verbatim; rather it SHALL integrate the evidence the prior comment carries (recommendation, open questions, references, and any human edits made inside the comment) into the synthesis of the new comment. The issue body and human comments fed to the agent SHALL be sanitised (HTML comments stripped) per the `ci-html-comment-sanitisation` capability.

#### Scenario: Maintainer edits a recommendation inside the comment between runs
- **WHEN** a maintainer (or other actor with comment-edit permission) edits the `### Recommendation` section of a prior research comment (for example, replacing the chosen approach with their own reasoning) and re-applies the `research-factory` label
- **THEN** the next run SHALL read the edit as evidence
- **AND** the next run SHALL produce a freshly synthesized comment that integrates that evidence rather than preserving the edit verbatim

#### Scenario: Maintainer answers an open question via a comment
- **WHEN** a maintainer posts an issue comment that resolves one of the prior comment's open questions and re-applies the `research-factory` label
- **THEN** the next run SHALL read the comment
- **AND** the next run's `### Open questions` section SHALL reflect the resolution (the question is removed or restated based on the new information)

#### Scenario: Input is sanitised before reaching the agent
- **WHEN** the workflow runs for an issue whose body contains an HTML comment
- **THEN** the body text passed to the agent SHALL have that HTML comment stripped
- **AND** the agent SHALL synthesize the new comment from sanitised input
