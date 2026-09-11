# deadcode-removal-rotation Specification

## Purpose
TBD - created by archiving change deadcode-remover-run-fmt. Update Purpose after archive.

## Requirements

### Requirement: Format code before opening a cleanup PR (REQ-FMT-001)

After verification (build and unit tests) passes and before the `create-pull-request` safe output is called, the agent SHALL run `make fmt` to ensure all modified files conform to the repository's formatting standards.

#### Scenario: Formatting succeeds — PR is opened

- **GIVEN** `make build` and `go test` for the impacted packages have both exited zero
- **WHEN** the agent runs `make fmt`
- **AND** `make fmt` exits zero
- **THEN** the agent SHALL proceed to open the cleanup pull request

#### Scenario: Formatting fails — PR is blocked and attempt is recorded

- **GIVEN** `make build` and `go test` for the impacted packages have both exited zero
- **WHEN** the agent runs `make fmt`
- **AND** `make fmt` exits non-zero
- **THEN** the agent SHALL record the attempt with reason `fmt_failed` in the rotation memory
- **AND** the agent SHALL call `noop` with a concise explanation
- **AND** the agent SHALL NOT open a pull request

### Requirement: Explicit `noop` on every verification-failure stop (REQ-NOOP-001)

When the agent stops the run without opening a pull request because `make build` failed, `make build` timed out, or `go test` failed for the impacted packages, the agent SHALL call the `noop` safe output with a concise reason immediately alongside recording the attempt in the rotation memory, rather than stopping without emitting any safe output.

#### Scenario: `make build` fails — attempt recorded and `noop` called

- **GIVEN** pre-activation selected a dead-code candidate and the agent removed the symbol
- **WHEN** the agent runs `make build`
- **AND** `make build` exits non-zero
- **THEN** the agent SHALL record the attempt with reason `build_failed` in the rotation memory
- **AND** the agent SHALL call `noop` with a concise explanation
- **AND** the agent SHALL NOT open a pull request

#### Scenario: `make build` times out — attempt recorded and `noop` called

- **GIVEN** pre-activation selected a dead-code candidate and the agent removed the symbol
- **WHEN** the agent runs `make build` under its 10-minute timeout
- **AND** the command times out
- **THEN** the agent SHALL record the attempt with reason `verification_timeout` in the rotation memory
- **AND** the agent SHALL call `noop` with a concise explanation
- **AND** the agent SHALL NOT open a pull request

#### Scenario: `go test` fails for an impacted package — attempt recorded and `noop` called

- **GIVEN** `make build` exited zero for the candidate's package
- **WHEN** the agent runs `go test` for the impacted packages
- **AND** any impacted package's tests fail
- **THEN** the agent SHALL record the attempt with reason `tests_failed` in the rotation memory
- **AND** the agent SHALL call `noop` with a concise explanation
- **AND** the agent SHALL NOT open a pull request

### Requirement: Every run terminates with exactly one safe output (REQ-NOOP-002)

Every run where pre-activation finds a dead-code candidate and the dead-code removal rotation agent task executes SHALL end by calling exactly one safe output — either `create-pull-request` or `noop` — and SHALL NOT stop mid-task without calling one of them.

#### Scenario: Run stops for a reason not covered by a specific branch — `noop` is still called

- **GIVEN** the agent reaches a point in the task where it cannot safely proceed and no more specific branch instruction applies
- **WHEN** the agent stops the run
- **THEN** the agent SHALL record the appropriate reason code in the rotation memory
- **AND** the agent SHALL call `noop` with a concise explanation before ending the run
