## ADDED Requirements

### Requirement: Workflow Go runtime sync maintenance targets
The repository SHALL provide Makefile targets that help contributors keep `.github/workflows/openspec-verify-label.md` `runtimes.go.version` aligned with the Go version declared in `go.mod`. At a minimum, the repository SHALL provide one target that checks for drift and fails when the versions do not match, and one target that synchronizes `runtimes.go.version` from `go.mod` and regenerates `.github/workflows/openspec-verify-label.lock.yml` via `gh aw compile`.

#### Scenario: Drift check fails on mismatch
- **GIVEN** `.github/workflows/openspec-verify-label.md` declares a `runtimes.go.version` value that differs from the Go version in `go.mod`
- **WHEN** the dedicated Makefile drift-check target runs
- **THEN** it SHALL exit with a non-zero status and explain that the workflow Go runtime is out of sync

#### Scenario: Sync target updates workflow and lock file
- **GIVEN** `.github/workflows/openspec-verify-label.md` `runtimes.go.version` does not match `go.mod`
- **WHEN** the dedicated Makefile sync target runs
- **THEN** it SHALL rewrite the workflow frontmatter to the `go.mod` version and regenerate `.github/workflows/openspec-verify-label.lock.yml`

## MODIFIED Requirements

### Requirement: Lint aggregate targets (REQ-044–REQ-045)
The `lint` target SHALL run setup, golangci-lint (with fix), formatting, and documentation generation. The `check-lint` target SHALL run setup, OpenSpec structural validation, the workflow Go runtime drift-check target, golangci-lint (check mode), format check, and documentation freshness check.

#### Scenario: Lint matches contributor workflow
- GIVEN `make lint`
- WHEN it completes successfully
- THEN formatting, lint with fix, and docs generation SHALL have run after setup

#### Scenario: Check-lint fails when workflow Go runtime drifts
- **GIVEN** `.github/workflows/openspec-verify-label.md` `runtimes.go.version` differs from `go.mod`
- **WHEN** `make check-lint` runs
- **THEN** it SHALL fail before reporting success for repository validation
