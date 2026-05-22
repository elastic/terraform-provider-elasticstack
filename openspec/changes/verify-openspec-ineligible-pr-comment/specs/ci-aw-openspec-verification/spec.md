## ADDED Requirements

### Requirement: Comment on ineligible PR (REQ-017)

When the `verify-openspec` label is applied to a pull request that does not meet the eligibility criteria for verification, the workflow SHALL post a pull request comment that includes the specific ineligibility reason and remediation guidance explaining what must be changed for the PR to become eligible. The comment SHALL be posted by a deterministic step in the `pre_activation` inject-steps block and SHALL NOT require agent execution.

The comment step SHALL only run when BOTH of the following are true:

1. The deterministic `verify_label` step confirmed the triggering label is `verify-openspec` (`label_verified == 'true'`).
2. The deterministic `classify_and_select` step determined the pull request is ineligible (`selection_status == 'ineligible'`).

When either condition is false the step SHALL be skipped. In particular, the step SHALL NOT run when `label_verified == 'false'` (a different label triggered the workflow, no user action required).

The comment body SHALL include:

- A clear heading indicating that the `verify-openspec` verification was skipped.
- The verbatim `selection_reason` string from the `classify_and_select` step output.
- A "How to fix" section explaining the eligibility requirements: exactly one active OpenSpec change directory under `openspec/changes/<id>/`, with all changed files having status `added` or `modified`, and no unsupported statuses such as `renamed` or `deleted`.
- A link to the OpenSpec authoring guide or equivalent documentation.

#### Scenario: Ineligible PR receives a comment

- **GIVEN** the `verify-openspec` label is applied to a pull request with no files under `openspec/changes/` (non-archive)
- **WHEN** the deterministic `classify_and_select` step outputs `selection_status == 'ineligible'`
- **THEN** the `comment_ineligible` step SHALL post a PR comment containing the `selection_reason` and "How to fix" remediation guidance

#### Scenario: Ineligible PR comment includes all ineligibility scenarios

- **GIVEN** the `classify_and_select` step sets `selection_reason` to any of the known ineligibility messages (e.g., "No files under openspec/changes/ (non-archive) found in this PR", "Multiple active change ids: ...", "Unsupported file status under openspec/changes/: ...")
- **WHEN** the `comment_ineligible` step runs
- **THEN** the comment body SHALL contain the verbatim `selection_reason` string and SHALL include the "How to fix" section

#### Scenario: Wrong-label trigger does not produce an ineligible comment

- **GIVEN** a label other than `verify-openspec` is applied to a pull request
- **WHEN** `label_verified` is `false`
- **THEN** the `comment_ineligible` step SHALL be skipped and no comment SHALL be posted

#### Scenario: Eligible PR does not receive an ineligible comment

- **GIVEN** the `verify-openspec` label is applied and `classify_and_select` outputs `selection_status == 'eligible'`
- **WHEN** downstream steps run
- **THEN** the `comment_ineligible` step SHALL be skipped

#### Scenario: Duplicate comments are acceptable

- **GIVEN** the `verify-openspec` label is applied repeatedly to the same ineligible PR
- **WHEN** the `comment_ineligible` step runs each time
- **THEN** a new comment SHALL be posted for each activation (deduplication is not required)
