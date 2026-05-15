## ADDED Requirements

### Requirement: Dead-code rotation runs on a schedule and processes one candidate per run
The `ci-deadcode-removal-rotation` workflow SHALL support scheduled execution and SHALL process at most one dead-code candidate in each run.

#### Scenario: Scheduled run selects at most one candidate
- **WHEN** the workflow starts on its configured schedule
- **THEN** it SHALL compute the current candidate set
- **AND** it SHALL select at most one cooldown-eligible candidate for agent execution

### Requirement: Candidate generation uses only symbols dead with and without tests
The workflow SHALL run both `go tool deadcode ./...` and `go tool deadcode -test ./...` during deterministic pre-activation and SHALL only consider symbols that appear in both result sets as cleanup candidates.

#### Scenario: Candidate survives both deadcode runs
- **WHEN** a symbol appears in the output of `go tool deadcode ./...`
- **AND** the same symbol appears in the output of `go tool deadcode -test ./...`
- **THEN** the workflow MAY consider that symbol for selection subject to cooldown and other deterministic gates

#### Scenario: Symbol is alive in test mode
- **WHEN** a symbol appears in the output of `go tool deadcode ./...`
- **AND** does not appear in the output of `go tool deadcode -test ./...`
- **THEN** the workflow SHALL exclude that symbol from automatic cleanup in the first iteration

### Requirement: Candidate selection is deterministic and cooldown-aware
The workflow SHALL maintain cooldown memory for attempted candidates and SHALL select one candidate per run using deterministic cooldown-aware ordering with a stable tie-breaker.

#### Scenario: Recently attempted candidate is skipped during cooldown
- **WHEN** a candidate has an active cooldown entry in workflow memory
- **THEN** the workflow SHALL exclude that candidate from the current run's eligible selection set

#### Scenario: Eligible candidates are ordered deterministically
- **WHEN** two or more cooldown-eligible candidates are available
- **THEN** the workflow SHALL apply a deterministic ordering with a stable tie-breaker to choose exactly one candidate for the run

### Requirement: Reference classification uses gopls
The workflow SHALL use `gopls references` to collect unique referring file paths for the selected candidate symbol during deterministic pre-activation.

#### Scenario: Single referring file is detected
- **WHEN** `gopls references` for the selected symbol resolves to references in exactly one unique file
- **THEN** the workflow SHALL record that single-file result for downstream test-cleanup eligibility checks

#### Scenario: Multiple referring files are detected
- **WHEN** `gopls references` for the selected symbol resolves to references in multiple unique files
- **THEN** the workflow SHALL record that the candidate is not eligible for single-file companion test cleanup

### Requirement: Deterministic pre-activation excludes acceptance-style companion test cleanup by filename
A selected candidate SHALL only be eligible for companion test cleanup when any local referring test file is not named `acc_*test.go`.

#### Scenario: Acceptance-style test filename blocks companion test cleanup
- **WHEN** the selected candidate's local referring test file matches `acc_*test.go`
- **THEN** the workflow SHALL mark the candidate as ineligible for automatic companion test cleanup

### Requirement: Agent test cleanup is limited to one local non-acceptance test file
The agent SHALL only remove tests referencing the selected symbol when deterministic pre-activation has already established that all test references are confined to exactly one local `*_test.go` file, and the agent's acceptance-style backstop passes.

#### Scenario: Local single-file test cleanup is allowed
- **WHEN** deterministic pre-activation marks the selected candidate as having references confined to exactly one local `*_test.go` file
- **AND** the agent verifies that the file does not contain `resource.Test` or `resource.ParallelTest`
- **THEN** the agent MAY remove tests in that file that reference the selected symbol while removing the symbol itself

#### Scenario: Distributed or cross-package references block automatic test cleanup
- **WHEN** the selected candidate's references span multiple files or packages
- **THEN** the agent SHALL NOT remove companion tests automatically

### Requirement: Agent acceptance-style backstop aborts invalid candidates
Before removing any companion tests, the agent SHALL inspect the eligible local test file and SHALL treat the candidate as invalid for automatic cleanup if the file contains `resource.Test` or `resource.ParallelTest`.

#### Scenario: Acceptance-style test usage aborts cleanup
- **WHEN** the agent finds `resource.Test` or `resource.ParallelTest` in the local referring test file
- **THEN** the agent SHALL stop without attempting the dead-code cleanup
- **AND** the workflow SHALL record cooldown memory for that attempted candidate

### Requirement: Verification requires build and impacted-package unit tests before PR creation
Before opening a pull request, the workflow SHALL run `make build` with a ten-minute timeout and SHALL run unit tests for the impacted package or packages.

#### Scenario: Verification succeeds before PR creation
- **WHEN** `make build` completes successfully within ten minutes
- **AND** unit tests for the impacted package or packages pass
- **THEN** the workflow MAY open a cleanup PR for the selected candidate

#### Scenario: Verification failure stops the run
- **WHEN** `make build` fails, times out, or impacted-package unit tests fail
- **THEN** the workflow SHALL stop without opening a PR
- **AND** it SHALL record cooldown memory for the attempted candidate

### Requirement: Impacted package scope follows the edited symbol and companion test file
The impacted package set SHALL always include the package containing the removed symbol. If the workflow also removes a companion test file from a different package directory, that package SHALL also be included in verification.

#### Scenario: Only the symbol package is impacted
- **WHEN** the workflow removes a dead symbol and does not remove tests in another package directory
- **THEN** it SHALL run unit tests for the symbol's package

#### Scenario: Symbol and companion tests span two package directories
- **WHEN** the workflow removes a dead symbol and also removes eligible companion tests from a different package directory
- **THEN** it SHALL run unit tests for both impacted package directories before opening a PR

### Requirement: Cooldown memory records all attempted candidates regardless of outcome
The workflow SHALL update cooldown memory for every attempted candidate regardless of whether the attempt ends in invalid-candidate detection, verification failure, or PR creation.

#### Scenario: Successful verified attempt still updates cooldown memory
- **WHEN** the workflow successfully verifies a cleanup and opens a PR
- **THEN** it SHALL record cooldown memory for that attempted candidate

#### Scenario: Failed attempt updates cooldown memory
- **WHEN** the workflow stops because the candidate is invalid or verification fails
- **THEN** it SHALL record cooldown memory for that attempted candidate

### Requirement: Human review remains the merge gate
The workflow SHALL open PRs for successful verified cleanups, but SHALL NOT auto-merge them.

#### Scenario: Successful cleanup awaits maintainer decision
- **WHEN** the workflow opens a cleanup PR after successful verification
- **THEN** the PR SHALL remain subject to normal maintainer review and manual merge or closure
