## MODIFIED Requirements

### Requirement: Shared changelog engine is owned by the `scripts/changelog/` Go tool

The shared changelog engine SHALL be implemented in Go and SHALL ship as the `scripts/changelog/` module documented by the `changelog-tooling` capability. Workflows SHALL invoke the engine by running `go run ./scripts/changelog <subcommand>` from a step that has already executed `actions/checkout` and `actions/setup-go`. The engine SHALL NOT be re-implemented as a parallel JavaScript or shell tool; the previous JavaScript tree under `.github/scripts/workflows/lib/changelog-*.js`, `.github/scripts/workflows/lib/pr-changelog-*.js`, `.github/scripts/workflows/changelog/*.js`, and `.github/scripts/workflows/pr-changelog-check/*.js` SHALL be removed when the Go tool reaches behavioural parity.

The engine SHALL preserve every externally observable behaviour required by this capability — modes, validation gates, GitHub-token usage, merged-PR resolution, `CHANGELOG.md` section format, and link-table maintenance — so callers and downstream consumers see identical output to the prior JavaScript implementation on the same inputs.

#### Scenario: Engine is invoked via `go run ./scripts/changelog`

- **WHEN** a workflow step needs to run the shared changelog engine
- **THEN** it SHALL invoke `go run ./scripts/changelog <subcommand>` (after `actions/checkout` and `actions/setup-go`) rather than calling any `actions/github-script` JavaScript module under `.github/scripts/workflows/`

#### Scenario: No parallel JavaScript engine remains after migration

- **WHEN** the migration to `scripts/changelog/` has landed
- **THEN** the repository SHALL contain no `.github/scripts/workflows/lib/changelog-*.js`, `.github/scripts/workflows/lib/pr-changelog-*.js`, `.github/scripts/workflows/changelog/*.js`, or `.github/scripts/workflows/pr-changelog-check/*.js` files; the `Makefile`'s `workflow-test` target SHALL NOT enumerate any deleted `.test.mjs` siblings for these clusters
