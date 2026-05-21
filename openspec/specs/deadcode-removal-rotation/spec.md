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

