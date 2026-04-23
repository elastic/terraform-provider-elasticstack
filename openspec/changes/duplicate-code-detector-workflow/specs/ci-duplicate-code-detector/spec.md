# `ci-duplicate-code-detector` — bounded duplicate-code issue detection workflow

Workflow implementation: repository-authored source under `.github/workflows-src/`, derived from `https://github.com/github/gh-aw/blob/main/.github/workflows/duplicate-code-detector.md`, and compiled to `.github/workflows/`.

## Purpose

Define requirements for a GitHub Agentic Workflow that analyzes recent repository changes for meaningful code duplication and opens bounded, actionable GitHub issues for refactoring follow-up.

## ADDED Requirements

### Requirement: Workflow artifacts and compilation
The duplicate-code detector SHALL be authored as a GitHub Agentic Workflow markdown source under `.github/workflows-src/` and SHALL include generated workflow artifacts under `.github/workflows/`, including a compiled `.lock.yml` derived from the authored source. The repository-authored source SHALL identify `https://github.com/github/gh-aw/blob/main/.github/workflows/duplicate-code-detector.md` as its upstream baseline. Contributors SHALL NOT hand-edit the generated workflow artifacts.

#### Scenario: Source and generated artifacts stay paired
- **WHEN** maintainers change duplicate-code detector behavior
- **THEN** the authored workflow source, generated workflow markdown, and compiled lock artifact SHALL match the repository compiler output

#### Scenario: Upstream workflow source remains referenced
- **WHEN** maintainers review or update the repository-authored duplicate-code detector workflow source
- **THEN** the change artifacts SHALL continue to reference `https://github.com/github/gh-aw/blob/main/.github/workflows/duplicate-code-detector.md` as the upstream workflow source

### Requirement: Scheduled and manual triggering
The duplicate-code detector SHALL support scheduled execution and manual `workflow_dispatch` execution.

#### Scenario: Scheduled run is supported
- **WHEN** the duplicate-code detector reaches its configured schedule
- **THEN** the workflow SHALL start a duplicate-code analysis run

#### Scenario: Manual dispatch is supported
- **WHEN** a maintainer triggers the workflow with `workflow_dispatch`
- **THEN** the workflow SHALL start a duplicate-code analysis run

### Requirement: Deterministic issue-slot gating
Before agent analysis begins, deterministic repository-authored steps SHALL compute available issue slots by counting open GitHub issues with the `duplicate-code` label and subtracting that count from a workflow-configured issue cap. The workflow SHALL expose the open-issue count, available slot count, and gate reason through pre-activation outputs, and it SHALL skip the agent job when the available slot count is zero.

#### Scenario: Open issues leave slots available
- **WHEN** the number of open `duplicate-code` issues is below the configured issue cap
- **THEN** the workflow SHALL expose a positive `issue_slots_available` value and proceed to agent analysis

#### Scenario: Open issues reach the cap
- **WHEN** the number of open `duplicate-code` issues is equal to or greater than the configured issue cap
- **THEN** the workflow SHALL expose `issue_slots_available` as zero and SHALL skip the agent job

### Requirement: Analysis scope is constrained to actionable source-code duplication
The workflow SHALL instruct the agent to analyze recently changed source files first, cross-reference duplication against the broader repository, and skip test files, generated artifacts, workflow files, standard boilerplate, vendored code, and small snippets below the configured minimum significance threshold.

#### Scenario: Changed source files are primary input
- **WHEN** the workflow analyzes a new run
- **THEN** it SHALL direct the agent to prioritize recently changed source files before broader repository cross-reference

#### Scenario: Noisy file classes are excluded
- **WHEN** the workflow encounters tests, generated files, or workflow definitions during duplicate analysis
- **THEN** it SHALL exclude those files from duplicate-code reporting

### Requirement: Significant duplication threshold
The workflow SHALL open issues only for significant duplication findings. A finding SHALL be considered significant only when it exceeds the workflow's documented threshold for duplicated code size or repeated occurrences.

#### Scenario: Significant duplication is eligible for issue creation
- **WHEN** the workflow identifies a duplication pattern above the configured significance threshold
- **THEN** that pattern SHALL be eligible for issue creation subject to the available issue slots

#### Scenario: Minor duplication is not reported
- **WHEN** the workflow identifies only small or low-signal duplication below the configured threshold
- **THEN** it SHALL NOT create a duplicate-code issue for that finding

### Requirement: One issue per duplication pattern
The workflow SHALL create at most one issue per distinct duplication pattern and SHALL NOT create more issues in a run than the computed number of available issue slots.

#### Scenario: Multiple patterns are capped by available slots
- **WHEN** the workflow identifies more significant duplication patterns than the computed number of available issue slots
- **THEN** it SHALL create issues only for the highest-priority patterns up to the available slot count

#### Scenario: Distinct patterns are not bundled
- **WHEN** the workflow creates an issue for a significant duplication finding
- **THEN** that issue SHALL describe exactly one distinct duplication pattern rather than bundling unrelated patterns together

### Requirement: Duplicate-code issue contents are actionable
Each duplicate-code issue created by the workflow SHALL include a concise summary of the duplication pattern, concrete affected locations, a severity or impact assessment, and actionable refactoring guidance sufficient for a follow-up coding agent or maintainer to act on the issue.

#### Scenario: Issue contains evidence and guidance
- **WHEN** the workflow creates a duplicate-code issue
- **THEN** the issue body SHALL include the duplicated locations and actionable refactoring recommendations for that pattern

#### Scenario: Issue titles and labels identify the workflow output
- **WHEN** the workflow creates a duplicate-code issue
- **THEN** the issue SHALL carry the workflow's configured title prefix and the `duplicate-code` label
