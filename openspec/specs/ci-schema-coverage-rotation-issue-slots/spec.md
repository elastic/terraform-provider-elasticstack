# ci-schema-coverage-rotation-issue-slots Specification

## Purpose
TBD - created by archiving change schema-coverage-issue-slot-gating. Update Purpose after archive.
## Requirements
### Requirement: Pre-activation issue-slot calculation
The `schema-coverage-rotation` workflow SHALL run a repository-local script during pre-activation to count currently open repository issues labeled `schema-coverage` (excluding pull requests) and compute `issue_slots_available` as `max(0, 3 - open_issues)` before agent activation begins.

#### Scenario: Open issue count is below the cap
- **WHEN** the pre-activation script finds fewer than 3 open `schema-coverage` issues
- **THEN** it returns the matching `open_issues` count and a positive `issue_slots_available` value

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

