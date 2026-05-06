## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Created semantic-refactor issues are explicitly dispatched to `code-factory`
After safe-output issue creation completes, the workflow SHALL explicitly dispatch the `code-factory` workflow once for each issue created in the current run rather than relying on producer-side `code-factory` labels to trigger implementation intake.

#### Scenario: One created issue dispatches one implementation run
- **WHEN** the workflow creates one semantic refactor issue in a run
- **THEN** it SHALL dispatch exactly one `code-factory` workflow run for that issue

#### Scenario: Three created issues dispatch three implementation runs
- **WHEN** the workflow creates three semantic refactor issues in a run
- **THEN** it SHALL dispatch exactly three independent `code-factory` workflow runs, one per created issue
