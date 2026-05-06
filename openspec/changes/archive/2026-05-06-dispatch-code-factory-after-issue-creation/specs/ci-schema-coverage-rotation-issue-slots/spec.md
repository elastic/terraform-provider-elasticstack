## ADDED Requirements

### Requirement: Created schema-coverage issues are explicitly dispatched to `code-factory`
After safe-output issue creation completes, the `schema-coverage-rotation` workflow SHALL explicitly dispatch the `code-factory` workflow once for each schema-coverage issue created in the current run rather than relying on a producer-side `code-factory` label to trigger implementation intake.

#### Scenario: One created schema-coverage issue dispatches one implementation run
- **WHEN** the workflow creates one schema-coverage issue in a run
- **THEN** it SHALL dispatch exactly one `code-factory` workflow run for that issue

#### Scenario: Multiple created schema-coverage issues dispatch multiple implementation runs
- **WHEN** the workflow creates multiple schema-coverage issues in a run
- **THEN** it SHALL dispatch exactly one independent `code-factory` workflow run per created issue

### Requirement: Schema-coverage issue labels do not include `code-factory`
The `schema-coverage-rotation` workflow SHALL use schema-coverage-specific issue labels for created issues and SHALL NOT depend on adding `code-factory` to those created issues in order to hand them off for implementation.

#### Scenario: Maintainer inspects schema-coverage issue safe-output configuration
- **WHEN** maintainers inspect the schema-coverage workflow source or generated artifacts
- **THEN** the created issue labels SHALL include schema-coverage-specific labels
- **AND** the created issue labels SHALL NOT include `code-factory`
