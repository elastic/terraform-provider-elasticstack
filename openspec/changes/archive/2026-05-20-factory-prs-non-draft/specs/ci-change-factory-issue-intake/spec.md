## ADDED Requirements

### Requirement: Proposal pull requests are created as non-draft

The `change-factory` workflow SHALL configure `safe-outputs.create-pull-request.draft: false` so that every proposal pull request is created in ready-for-review state. Draft pull requests cannot receive review until manually converted, which prevents the proposal from being immediately actionable.

#### Scenario: Proposal PR is immediately reviewable
- **WHEN** the `change-factory` agent creates the linked proposal pull request
- **THEN** the pull request SHALL be created as non-draft (ready for review)
- **AND** reviewers SHALL be able to start reviewing the proposal without the maintainer first converting it from draft state

#### Scenario: Maintainer inspects authored workflow safe-output configuration for draft policy
- **WHEN** maintainers inspect the authored `change-factory` issue-intake workflow `safe-outputs` block
- **THEN** `safe-outputs.create-pull-request.draft` SHALL be set to `false`
- **AND** generated workflow artifacts derived from that source SHALL preserve the non-draft policy
