# ci-research-factory-comment-format Specification

## Purpose
Define the format of the implementation-research output produced by the `research-factory` workflow as a dedicated sticky comment on the triggering issue, authored by `github-actions[bot]`. The comment is the sole durable output of a research run.

## Requirements

### Requirement: Comment is authored by github-actions[bot] and identified by a marker
The implementation-research output SHALL be a single issue comment authored by `github-actions[bot]`. The producer SHALL prepend exactly the marker `<!-- gha-research-factory -->` on its own line to the comment body before creating or updating the comment. The marker serves as a filter key for downstream consumers (e.g., `change-factory`) to locate the research comment among other bot comments on the issue. There SHALL be at most one such research comment per issue. If multiple comments matching both the `github-actions[bot]` author and the `<!-- gha-research-factory -->` marker are found, the producer SHALL update the most recently created matching comment and SHALL NOT create an additional one.

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
6. `### Quality gate` — the human-readable result of the done gate: the gate outcome (`ready-for-change-factory` or `research-needs-human`) with a one-line reason, a table showing pass or fail for each hard-checklist item, the final critic score against the threshold of 85, the number of critique rounds used, and, when the outcome is `research-needs-human`, a short bullet list of the critic's outstanding actionable feedback.
7. `### References` — a list of consulted sources, including elastic-docs URLs and repository paths inspected during research.

The `### Quality gate` section is informational. Downstream consumers SHALL NOT treat its contents as part of the recommended scope.

#### Scenario: Comment contains the required subsections
- **GIVEN** a research comment produced by the workflow
- **WHEN** a maintainer or downstream consumer reads it
- **THEN** the comment SHALL contain the headings `### Problem framing`, `### Approaches considered`, `### Recommendation`, `### Open questions`, `### Out of scope`, `### Quality gate`, and `### References` in that order

#### Scenario: Approaches considered contains at least two approaches
- **GIVEN** a research comment produced by the workflow
- **WHEN** the comment's `### Approaches considered` section is inspected
- **THEN** it SHALL contain two or more `#### ` child headings, each naming a distinct candidate approach

#### Scenario: Recommendation names exactly one approach
- **GIVEN** a research comment produced by the workflow
- **WHEN** the comment's `### Recommendation` section is inspected
- **THEN** it SHALL identify exactly one of the approaches enumerated under `### Approaches considered` as the chosen approach
- **AND** it SHALL include a brief rationale for that choice

#### Scenario: Quality gate explains a research-needs-human outcome
- **GIVEN** a research run whose gate outcome is `research-needs-human` because the `Versioned` checklist item failed
- **WHEN** a maintainer reads the `### Quality gate` section
- **THEN** the section SHALL state the outcome `research-needs-human` and its reason
- **AND** the checklist table SHALL show `Versioned` as failed
- **AND** the section SHALL list the critic's outstanding actionable feedback

#### Scenario: Quality gate for a ready outcome
- **GIVEN** a research run whose gate outcome is `ready-for-change-factory`
- **WHEN** a maintainer reads the `### Quality gate` section
- **THEN** the checklist table SHALL show every item as passed
- **AND** the section SHALL show a critic score of at least 85

### Requirement: Comment contains a structured machine-readable JSON metadata block
After the `### References` section, the comment SHALL contain an HTML `<details>` element with a `<summary>` of exactly "🤖 Pipeline metadata". Inside the `<details>` element, the comment SHALL contain exactly one fenced JSON code block (language `json`) containing a machine-readable representation of the research. The JSON object SHALL conform to the following schema:

- `schema_version` (string, required): the version of this metadata schema. Every comment produced by the workflow SHALL use `"1.1"`. Version `"1.0"` exists only in comments written before this change.
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
- `gate` (object, required in schema version `"1.1"`):
  - `outcome` (string, required, enum `["ready-for-change-factory","research-needs-human"]`): the outcome the agent reports.
  - `checklist` (object, required): boolean values for each of `grounded`, `mapped`, `compatible`, `versioned`, `testable`, and `idiomatic`.
  - `score` (number or null, required): the final critic score, 0 to 100; `null` when no critique round completed.
  - `scores` (array of numbers, required): the critic score for each completed round, in order; empty when no critique round completed.
  - `converged` (boolean, required): whether the convergence rule was satisfied.
  - `rounds` (number, required): the number of completed critique rounds, 0 to 5. It SHALL equal the length of `scores`.
  - `outstanding_feedback` (array of strings, required): the critic's remaining actionable feedback after the final round; empty when the critic has none.
  - `author_model` (string, required): the identifier of the model that authored the research.
  - `critic` (object, required):
    - `model` (string, required): the identifier of the critic model.
    - `status` (string, required, enum `["ok","unavailable","error"]`): whether the critic produced usable critiques.

All fields present in schema version `"1.0"` SHALL keep their meaning, so consumers that ignore `gate` keep working. The agent SHALL ensure that the JSON content is internally consistent with the human-readable subsections above it. The `<details>` element SHALL be closed by default so that human readers do not see the JSON unless they expand it.

#### Scenario: Comment contains valid JSON metadata
- **GIVEN** a research comment produced by the workflow
- **WHEN** a maintainer or downstream consumer inspects it
- **THEN** the comment SHALL contain a `<details>` element after `### References`
- **AND** inside that element there SHALL be a fenced JSON block conforming to the schema above
- **AND** the JSON `recommendation.spine` SHALL match the human-readable `### Recommendation` content
- **AND** the JSON `gate` values SHALL match the human-readable `### Quality gate` content

#### Scenario: JSON metadata is hidden from human readers by default
- **GIVEN** a research comment produced by the workflow
- **WHEN** a maintainer reads it on GitHub
- **THEN** the JSON metadata SHALL be collapsed inside a `<details>` element
- **AND** it SHALL NOT be visible without clicking the disclosure triangle

#### Scenario: Downstream consumer parses JSON metadata
- **GIVEN** a research comment produced by the workflow
- **WHEN** a downstream workflow (e.g., a future classifier or change-factory enhancement) reads it
- **THEN** it SHALL be able to extract the JSON block from the `<details>` element by parsing the comment body
- **AND** it SHALL be able to read `open_questions[].blocking` and `gate.outcome` without regex-parsing headings

#### Scenario: Consumer written for schema 1.0 reads a 1.1 comment
- **GIVEN** a consumer that only understands schema version `"1.0"` fields
- **WHEN** it parses a comment whose metadata has `schema_version` `"1.1"`
- **THEN** every `"1.0"` field SHALL be present with unchanged meaning

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

### Requirement: Research is judged against a two-part done gate
Research SHALL be judged against a hard checklist and a convergence rule.

The hard checklist SHALL consist of these items, each of which passes or fails:

- **Grounded**: every claimed capability cites a verifiable source: an Elastic Stack API specification node, Elastic documentation, or existing client or provider code.
- **Mapped**: each new capability has a proposed Terraform schema mapping: attribute name, type, whether it is required, optional, or computed, and its nesting.
- **Compatible**: the change is assessed as additive or breaking; if breaking, a migration or state-upgrade note is present.
- **Versioned**: the minimum Elastic Stack version is stated and a version-gating strategy is named.
- **Testable**: an acceptance-test outline is present, covering configuration, assertions, unset or empty values, and update cases.
- **Idiomatic**: the design references existing repository patterns rather than inventing new ones.

With a score threshold of 85, research SHALL be considered **converged** when the final critic score is at least 85 and either the critic scores at least 85 on two consecutive rounds, or the critic reports no actionable feedback. A plateau below 85 SHALL NOT count as converged.

#### Scenario: High score with no feedback converges in one round
- **GIVEN** a single critique round with score 90 and no actionable feedback
- **WHEN** convergence is evaluated
- **THEN** the research SHALL be converged

#### Scenario: Plateau below the threshold does not converge
- **GIVEN** a critique round with score 70 and no actionable feedback
- **WHEN** convergence is evaluated
- **THEN** the research SHALL NOT be converged

#### Scenario: One high round after a low round does not converge
- **GIVEN** critique rounds scored 80 then 88, with actionable feedback remaining after the second round
- **WHEN** convergence is evaluated after the second round
- **THEN** the research SHALL NOT be converged

### Requirement: Gate-outcome label is derived from the published metadata
The gate-outcome label applied to the issue SHALL be derived deterministically from the comment's published JSON metadata, independently of the outcome the agent reports. The label SHALL be `ready-for-change-factory` only when all of the following hold in the metadata:

- every `gate.checklist` value is `true`;
- the convergence rule holds when recomputed from the metadata: the last value in `gate.scores` is at least 85, and either the last two values in `gate.scores` are both at least 85 or `gate.outstanding_feedback` is empty;
- `gate.rounds` is between 1 and 5 and equals the length of `gate.scores`;
- `gate.critic.status` is `ok`;
- no `open_questions` item has `blocking: true`.

The agent-reported `gate.converged` flag SHALL NOT be relied on; convergence SHALL be recomputed as above. In every other case, including missing, unparseable, or schema-invalid metadata, the label SHALL be `research-needs-human`. When the agent-reported `gate.outcome` disagrees with the derived label, the derived label SHALL win, the disagreement SHALL be recorded in the workflow run's step summary, and the published comment SHALL be corrected so that its `### Quality gate` outcome and its `gate.outcome` match the applied label, with a visible note that the reported outcome was overridden. The comment and the label SHALL never disagree.

#### Scenario: Consistent ready metadata
- **GIVEN** metadata with all checklist items `true`, scores `[86, 88]`, 2 rounds, critic status `ok`, and no blocking open questions
- **WHEN** the label is derived
- **THEN** the issue SHALL be labelled `ready-for-change-factory`

#### Scenario: Agent claims ready despite a failed checklist item
- **GIVEN** metadata with `gate.outcome` `ready-for-change-factory` and `gate.checklist.versioned` `false`
- **WHEN** the label is derived
- **THEN** the issue SHALL be labelled `research-needs-human`
- **AND** the run's step summary SHALL record that the agent-reported outcome was overridden
- **AND** the published comment's `### Quality gate` section SHALL show the outcome `research-needs-human` with a note that the reported outcome was overridden by the gate rule

#### Scenario: Blocking open question prevents ready
- **GIVEN** metadata that otherwise satisfies every ready condition
- **AND** one `open_questions` item with `blocking: true`
- **WHEN** the label is derived
- **THEN** the issue SHALL be labelled `research-needs-human`

#### Scenario: Metadata is missing or malformed
- **GIVEN** a research comment whose JSON metadata block is absent or cannot be parsed
- **WHEN** the label is derived
- **THEN** the issue SHALL be labelled `research-needs-human`

#### Scenario: Agent claims convergence the scores do not support
- **GIVEN** metadata with `gate.converged` `true`, scores `[80, 88]`, and non-empty `outstanding_feedback`
- **AND** every other ready condition satisfied
- **WHEN** the label is derived
- **THEN** the issue SHALL be labelled `research-needs-human`

#### Scenario: No critique round completed
- **GIVEN** metadata with `gate.critic.status` `unavailable`, `gate.score` `null`, empty `gate.scores`, and `gate.rounds` `0`
- **WHEN** the metadata is validated and the label is derived
- **THEN** the metadata SHALL be schema-valid
- **AND** the issue SHALL be labelled `research-needs-human`
