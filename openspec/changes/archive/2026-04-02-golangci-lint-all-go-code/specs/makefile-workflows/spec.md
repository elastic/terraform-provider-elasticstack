## MODIFIED Requirements

### Requirement: golangci-lint execution (REQ-041–REQ-043)

The `tools` target SHALL provision golangci-lint at the **version pinned in the repository**. The `golangci-lint` target SHALL lint Go code across the repository module using `./...`, while still honoring repository-configured golangci-lint exclusions, with zero tolerance for duplicate identical issues unless `GOLANGCIFLAGS` alters behavior. The `lint` target SHALL enable auto-fix behavior where supported; `check-lint` SHALL not depend on that fix mode for golangci-lint.

#### Scenario: Lint without fix

- GIVEN `make check-lint`
- WHEN golangci-lint runs
- THEN it SHALL report issues without the fix-only mode used by `lint`

#### Scenario: Repository-wide Go lint scope

- GIVEN `make golangci-lint`
- WHEN the target invokes golangci-lint
- THEN it SHALL run against `./...`
- AND Go packages outside `internal/` SHALL be part of the lint scope unless excluded by repository golangci-lint configuration
