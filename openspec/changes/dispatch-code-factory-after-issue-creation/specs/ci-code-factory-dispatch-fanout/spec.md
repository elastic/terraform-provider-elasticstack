## ADDED Requirements

### Requirement: Producer workflows dispatch `code-factory` once per created issue
Repository-authored issue-producing analysis workflows that hand off implementation to `code-factory` SHALL perform explicit post-safe-outputs fan-out by dispatching the `code-factory` workflow once for each issue created in the current run.

#### Scenario: Multiple created issues are fanned out deterministically
- **WHEN** a producer workflow creates multiple issues in one run
- **THEN** the workflow SHALL dispatch one independent `code-factory` workflow run per created issue
- **AND** each dispatch SHALL target exactly one issue number

#### Scenario: No issues created means no dispatches
- **WHEN** a producer workflow completes without creating any issues
- **THEN** it SHALL NOT dispatch `code-factory`

### Requirement: Fan-out uses the safe-output temporary issue ID map
The producer-side dispatch phase SHALL use the safe-output temporary issue ID map as the source of truth for created issue identities. For each created issue entry, the workflow SHALL read the repository slug and issue number from the map rather than inferring newly created issues by search, title, or label matching.

#### Scenario: Temporary issue ID map contains three created issues
- **WHEN** the producer workflow's safe-output artifacts include three temporary issue ID map entries with repository and issue-number values
- **THEN** the dispatch phase SHALL dispatch `code-factory` exactly three times using those repository and issue-number values

#### Scenario: Temporary issue ID map is missing or malformed
- **WHEN** the dispatch phase cannot read a valid temporary issue ID map
- **THEN** the workflow SHALL fail or stop dispatch fan-out rather than guessing which issues were created

### Requirement: Fan-out dispatch remains deterministic and repository-authored
The fan-out dispatch logic SHALL be implemented in deterministic repository-authored workflow jobs or scripts under `.github/workflows-src/` and SHALL NOT depend on agent-emitted safe outputs to determine which created issues to dispatch.

#### Scenario: Maintainer inspects dispatch handoff implementation
- **WHEN** maintainers review the producer workflow source
- **THEN** the logic that parses created issue identities and dispatches `code-factory` SHALL be repository-authored and deterministic
- **AND** it SHALL run after safe-output issue creation has completed
