## ADDED Requirements

### Requirement: Release PR carries the no-changelog label
The release preparation workflow SHALL apply the `no-changelog` label to the release PR. This label SHALL be present whether the PR is newly created or already exists (reused on a rerun). The `no-changelog` label is assumed to exist in the repository as a pre-condition.

#### Scenario: New PR is created with no-changelog label
- **WHEN** the workflow creates a new release PR
- **THEN** the PR SHALL have the `no-changelog` label applied at creation time

#### Scenario: Existing PR is labelled on reuse
- **GIVEN** a release PR for the target version already exists
- **WHEN** the workflow reruns and reuses that existing PR
- **THEN** the workflow SHALL apply the `no-changelog` label to the existing PR
- **AND** the label application SHALL be idempotent (re-applying an already-present label SHALL NOT cause an error)
