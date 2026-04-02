## MODIFIED Requirements

### Requirement: Copilot diff threshold (REQ-009)

The total pull request line edits (`additions + deletions`) for a `copilot` category pull request SHALL be strictly less than `1000`.

#### Scenario: Large Copilot PR

- GIVEN additions plus deletions are 1000 or more
- WHEN the Copilot category is considered
- THEN the threshold gate SHALL fail
