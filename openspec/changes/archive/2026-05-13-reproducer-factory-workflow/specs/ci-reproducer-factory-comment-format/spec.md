# `ci-reproducer-factory-comment-format` — sticky comment format for reproducer-factory outcomes

## Purpose

Define the format of the reproduction-output comment produced by the `reproducer-factory` workflow as a sticky comment on the triggering issue, authored by `github-actions[bot]`. The comment is always emitted regardless of outcome and covers three variants: bug reproduced, cannot reproduce, and appears already fixed.

## ADDED Requirements

### Requirement: Comment is authored by github-actions[bot] and identified by a marker
The reproduction-output SHALL be a single issue comment authored by `github-actions[bot]`. The producer SHALL prepend exactly the marker `<!-- gha-reproducer-factory -->` on its own line to the comment body before creating or updating the comment. There SHALL be at most one such comment per issue. If multiple comments matching both the `github-actions[bot]` author and the `<!-- gha-reproducer-factory -->` marker are found, the producer SHALL update the most recently created matching comment and SHALL NOT create an additional one.

#### Scenario: Reproduction comment is created on a fresh issue
- **WHEN** the `reproducer-factory` workflow runs for an issue with no prior reproduction comment
- **THEN** a new comment SHALL be created by `github-actions[bot]`
- **AND** that comment SHALL begin with `<!-- gha-reproducer-factory -->`
- **AND** the issue body SHALL NOT be modified

#### Scenario: Reproduction comment is updated on re-run
- **WHEN** the `reproducer-factory` workflow re-runs for an issue that already has a reproduction comment
- **THEN** the existing comment SHALL be updated in place
- **AND** the issue SHALL NOT gain an additional reproduction comment

#### Scenario: Multiple matching comments exist
- **WHEN** the workflow finds more than one comment by `github-actions[bot]` containing the `<!-- gha-reproducer-factory -->` marker
- **THEN** the producer SHALL update the most recently created matching comment
- **AND** the issue SHALL NOT gain an additional comment

### Requirement: Comment opens with a provenance header
Immediately after the marker, the comment SHALL begin with a `## Bug reproduction` heading followed by a provenance header that records the run timestamp, a link to the workflow run that authored the comment, and a notice that the comment is replaced on each run.

#### Scenario: Maintainer inspects an authored comment
- **WHEN** a maintainer reads a reproduction comment
- **THEN** the comment SHALL contain a `## Bug reproduction` heading
- **AND** the comment SHALL include a provenance line stating the timestamp and a link to the workflow run that produced it

### Requirement: Comment body varies by outcome
The comment body after the provenance header SHALL differ by outcome. The `### Summary` section SHALL always be present. Outcome-specific sections SHALL follow.

#### Scenario: Outcome A comment contains root cause and PR link
- **WHEN** the reproduction agent confirms the bug with a passing test
- **THEN** the comment SHALL contain `### Summary`, `### Root cause`, and `### Reproduction test` sections
- **AND** `### Reproduction test` SHALL include the test function name, file path, and a link to the PR

#### Scenario: Outcome B comment contains investigation avenues
- **WHEN** the reproduction agent cannot reproduce the failure condition
- **THEN** the comment SHALL contain `### Summary` and `### Investigation avenues` sections
- **AND** `### Investigation avenues` SHALL contain exactly 3 numbered items
- **AND** each item SHALL reference a specific file path or code symbol in the repository

#### Scenario: Outcome C comment contains evidence of a fix
- **WHEN** the reproduction agent determines the bug appears to be fixed in the current provider
- **THEN** the comment SHALL contain `### Summary` and `### Evidence` sections
- **AND** `### Evidence` SHALL include the test config used, the test output (confirming no error), and any relevant recent git commits or changes to the affected code area

### Requirement: Outcome B investigation avenues are specific and codebase-anchored
Each investigation avenue in an outcome-B comment SHALL be specific and actionable. An avenue SHALL NOT consist solely of vague guidance (e.g. "check the API docs"). Each avenue SHALL reference at least one concrete location: a repository file path, a Go symbol, a git commit, or a named API endpoint.

#### Scenario: Agent cannot write a credible test configuration
- **WHEN** the agent cannot determine how to configure the Terraform resource to trigger the described failure
- **THEN** one or more avenues SHALL name the specific schema attribute(s) or provider code path(s) that are candidates for the root cause

#### Scenario: Agent suspects a version dependency
- **WHEN** the issue mentions an Elastic Stack version and the agent suspects the failure may be version-gated
- **THEN** an avenue SHALL identify the specific version-check or API-version code in the provider

### Requirement: Outcome C evidence documents the test configuration used
When the outcome is "appears fixed", the comment SHALL document the exact test configuration the agent attempted so a maintainer can assess whether the configuration correctly mirrors the reported scenario. The test code SHALL be included in the comment body (inline or as a fenced code block) or the branch where it was attempted SHALL be referenced.

#### Scenario: Maintainer reviews appears-fixed comment
- **WHEN** a maintainer reads an outcome-C comment
- **THEN** the comment SHALL include the Terraform configuration used in the attempted test
- **AND** the comment SHALL include the test output confirming that no error matching the issue description was produced

### Requirement: Comment contains a structured machine-readable JSON metadata block
After the outcome-specific sections and `### References`, the comment SHALL contain an HTML `<details>` element with `<summary>🤖 Pipeline metadata</summary>`. Inside, the comment SHALL contain exactly one fenced JSON code block (language `json`) conforming to:

- `schema_version` (string, required): e.g. `"1.0"`
- `outcome` (string, required, enum `["reproduced", "cannot-reproduce", "appears-fixed"]`)
- `test_name` (string, optional): the `TestAccReproduceIssue{N}` function name, present for outcome A and C
- `test_file` (string, optional): the relative file path of the test, present for outcome A
- `pr_url` (string, optional): the PR URL, present for outcome A
- `references` (array, optional): each with `type` (enum `["elastic-docs","repo-path","issue","pr","external"]`) and `url` or `path`

The `<details>` element SHALL be closed by default.

#### Scenario: Comment contains valid JSON metadata
- **WHEN** a maintainer or downstream consumer inspects a reproduction comment
- **THEN** the comment SHALL contain a `<details>` element after `### References`
- **AND** inside it there SHALL be a fenced JSON block with `schema_version` and `outcome` fields

#### Scenario: JSON metadata is hidden from human readers by default
- **WHEN** a maintainer reads a reproduction comment on GitHub
- **THEN** the JSON metadata SHALL be collapsed inside a `<details>` element

#### Scenario: Outcome A metadata includes PR and test details
- **WHEN** reproduction succeeds
- **THEN** the JSON SHALL include `"outcome": "reproduced"`, `test_name`, `test_file`, and `pr_url`

#### Scenario: Outcome B metadata records cannot-reproduce
- **WHEN** reproduction cannot be confirmed
- **THEN** the JSON SHALL include `"outcome": "cannot-reproduce"`

#### Scenario: Outcome C metadata records appears-fixed
- **WHEN** the bug appears to be fixed
- **THEN** the JSON SHALL include `"outcome": "appears-fixed"` and `test_name`

### Requirement: Comment includes a References section
The comment SHALL end (before the `<details>` block) with a `### References` section listing the sources consulted during investigation, including repository file paths inspected and elastic-docs URLs queried.

#### Scenario: References section is present
- **WHEN** a maintainer reads a reproduction comment
- **THEN** the comment SHALL contain a `### References` section before the `<details>` block
- **AND** it SHALL list at least one source consulted during investigation
