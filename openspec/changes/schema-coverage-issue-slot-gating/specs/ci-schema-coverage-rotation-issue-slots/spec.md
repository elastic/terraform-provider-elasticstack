## ADDED Requirements

### Requirement: Pre-activation issue-slot calculation
The `schema-coverage-rotation` workflow SHALL run a repository-local script during pre-activation to count currently open repository issues labeled `schema-coverage` (excluding pull requests) and compute `issue_slots_available` as `max(0, 3 - open_schema_coverage_issues)` before agent activation begins.

#### Scenario: Open issue count is below the cap
- **WHEN** the pre-activation script finds fewer than 3 open `schema-coverage` issues
- **THEN** it returns the matching `open_schema_coverage_issues` count and a positive `issue_slots_available` value

#### Scenario: Open issue count reaches or exceeds the cap
- **WHEN** the pre-activation script finds 3 or more open `schema-coverage` issues
- **THEN** it returns `issue_slots_available = 0`

### Requirement: Workflow source is templated and the gating logic is unit-testable
The `schema-coverage-rotation` workflow SHALL be authored from a templated source under `.github/workflows-src/schema-coverage-rotation/`, and any inline GitHub-script used for pre-activation issue-slot gating SHALL delegate its decision logic to an extracted helper module under `.github/workflows-src/lib/` that can be unit tested independently of workflow compilation.

#### Scenario: Workflow source is compiled from template
- **WHEN** maintainers update the schema-coverage rotation workflow
- **THEN** the authored source lives under `.github/workflows-src/schema-coverage-rotation/` and compiles into the generated workflow artifacts

#### Scenario: Issue-slot logic is tested outside the workflow wrapper
- **WHEN** maintainers validate the deterministic issue-slot calculation
- **THEN** they can unit test the extracted helper module without executing the full workflow

### Requirement: Pre-activation outputs expose gate context
The workflow SHALL publish the pre-activation issue-slot results as downstream-consumable outputs, including the open issue count, the computed slot count, and a short gate status or reason describing whether the agent path remains eligible to run.

#### Scenario: Downstream jobs consume slot outputs
- **WHEN** pre-activation finishes calculating issue-slot capacity
- **THEN** later workflow conditions and prompt interpolation can read the published slot outputs without repeating the GitHub issue query

### Requirement: Agent execution is skipped when no slots remain
The workflow SHALL skip the schema-coverage agent job entirely when `issue_slots_available` is `0`.

#### Scenario: Capacity exhausted
- **WHEN** pre-activation computes `issue_slots_available = 0`
- **THEN** the workflow does not start the agent job for that run

### Requirement: Agent instructions consume precomputed slot state
When the agent job does run, the workflow prompt SHALL provide the precomputed issue-slot context and SHALL NOT instruct the agent to count open `schema-coverage` issues itself.

#### Scenario: Capacity remains available
- **WHEN** pre-activation computes a positive `issue_slots_available`
- **THEN** the prompt handed to the agent references the precomputed slot information rather than telling the agent to query GitHub issue counts
