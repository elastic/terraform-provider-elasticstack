## MODIFIED Requirements

### Requirement: Pre-activation issue-slot calculation
The `schema-coverage-rotation` workflow SHALL run a repository-local script during pre-activation to count currently open repository issues labeled `schema-coverage` (excluding pull requests) and compute `issue_slots_available` as `max(0, 3 - open_issues)` before agent activation begins.

#### Scenario: Open issue count is below the cap
- **WHEN** the pre-activation script finds fewer than 3 open `schema-coverage` issues
- **THEN** it returns the matching `open_issues` count and a positive `issue_slots_available` value
