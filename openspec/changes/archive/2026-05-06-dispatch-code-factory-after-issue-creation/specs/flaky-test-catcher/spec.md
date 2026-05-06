## MODIFIED Requirements

### Requirement: Issue creation
The agent SHALL create one GitHub issue per affected resource (up to `issue_slots_available`). Each issue SHALL be labelled `flaky-test`, and SHALL include: broken test list, flaky test list with fail rates, commit analysis result, a sample failure excerpt, and the affected stack versions.

#### Scenario: Issue created for resource with broken and flaky tests
- **WHEN** `elasticstack_elasticsearch_index` has both broken and flaky tests
- **THEN** an issue titled `[flaky-test] elasticstack_elasticsearch_index` is created with sections for Broken Tests, Flaky Tests, Commit Analysis, Sample Failure Output, and Affected Stack Versions, labelled `flaky-test`

#### Scenario: Issue cap enforced
- **WHEN** the agent has already created `issue_slots_available` issues in this run
- **THEN** no further issues are created regardless of remaining affected resources

## ADDED Requirements

### Requirement: Created flaky-test issues are explicitly dispatched to `code-factory`
After safe-output issue creation completes, the workflow SHALL explicitly dispatch the `code-factory` workflow once for each flaky-test issue created in the current run rather than relying on a producer-side `code-factory` label to activate implementation intake.

#### Scenario: One created flaky-test issue dispatches one implementation run
- **WHEN** the workflow creates one flaky-test issue in a run
- **THEN** it SHALL dispatch exactly one `code-factory` workflow run for that issue

#### Scenario: Multiple created flaky-test issues dispatch multiple implementation runs
- **WHEN** the workflow creates multiple flaky-test issues in a run
- **THEN** it SHALL dispatch exactly one independent `code-factory` workflow run per created issue
