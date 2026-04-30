## ADDED Requirements

### Requirement: Pre-activation repo-memory checkout path

The workflow's pre-activation repo-memory initialization step SHALL use a workspace-relative checkout path that is accepted by `actions/checkout`. The checkout path SHALL be co-located within the workflow workspace so that the deterministic helper can read the memory file from that path without requiring an absolute path outside the workspace.

#### Scenario: Workspace-relative repo-memory checkout succeeds
- **GIVEN** the workflow pre-activation job checks out the repo-memory branch for Kibana spec-impact baseline state
- **WHEN** the checkout step runs
- **THEN** it SHALL complete successfully using a workspace-relative path
- **AND** the subsequent deterministic helper SHALL be able to read the memory file from that path

## MODIFIED Requirements

(none)

## REMOVED Requirements

(none)
