## Why

The changelog engine cluster (~1,500 LOC across 13 JavaScript modules under `.github/scripts/workflows/lib/changelog-*.js` and `.github/scripts/workflows/changelog/`) plus the PR changelog check (~412 LOC across `.github/scripts/workflows/lib/pr-changelog-*.js` and `.github/scripts/workflows/pr-changelog-check/`) carry the most domain logic of any workflow-script cluster in this repo — structured-text parsing, semver-aware rewriting, customer-impact validation, evidence manifests, and PR-body authoring rules. Maintaining two independent parsers in JavaScript (`pr-changelog-parser.js` for PR-body sections and the changelog renderer for the same `## Changelog` contract) is duplication waiting to drift. Migrating both clusters into a single Go binary lets us reuse the rest of the project's primary language, get a strongly-typed shared section parser, and apply the same `go test` infrastructure used by `scripts/auto-approve/`, `scripts/kibana-spec-impact/`, `scripts/ci-deadcode-removal-rotation/`, and `scripts/schema-coverage-rotation/`.

## What Changes

- Introduce a new Go tool at `scripts/changelog/` exposing subcommands: `gather-evidence`, `run-engine`, `manage-unreleased-pr`, `refresh-release-pr`, and `validate-pr-section`.
- Implement a single canonical Go package (e.g. `scripts/changelog/internal/section`) that parses the `## Changelog` contract (customer impact, summary, breaking-changes subsection, fenced-code awareness) and is consumed by both the engine subcommands and the PR-body validator.
- Reuse the existing project Go toolchain — `github.com/google/go-github` (same major as `scripts/auto-approve/`) authenticated via `golang.org/x/oauth2` — for GitHub API access; read context from `$GITHUB_EVENT_PATH` and write results to `$GITHUB_OUTPUT`.
- Update `.github/workflows/changelog-generation.yml`, `.github/workflows/pr-changelog-check.yml`, and any other workflows currently invoking the JS modules to call `go run ./scripts/changelog <subcommand>` after the existing `setup-go` step.
- **BREAKING (internal only)**: Delete `.github/scripts/workflows/lib/changelog-*.js`, `.github/scripts/workflows/lib/pr-changelog-*.js`, `.github/scripts/workflows/changelog/*.js`, and `.github/scripts/workflows/pr-changelog-check/*.js` once Go subcommands reach behavioural parity (verified by ported tests). Port matching `.test.mjs` cases to `_test.go` under `scripts/changelog/`.
- Update `Makefile` so `workflow-test` no longer enumerates the deleted `.test.mjs` files for these clusters; `go test ./scripts/changelog/...` covers them.

## Capabilities

### New Capabilities
- `changelog-tooling`: Go CLI under `scripts/changelog/` that owns the `## Changelog` PR-body parser, the `CHANGELOG.md` rewriter, the release/unreleased engine modes, evidence gathering, and the PR-body validator. Defines the subcommand surface, exit semantics, `$GITHUB_OUTPUT` keys, and shared parser contract reused by every subcommand.

### Modified Capabilities
- `ci-changelog-generation`: the existing requirement "Shared changelog engine reuses existing JavaScript helpers" explicitly forbids a Go implementation under `scripts/`. That requirement must be replaced with an implementation-neutral version that permits (and points at) the new `scripts/changelog/` Go tool, while keeping every observable behaviour (modes, validation gates, GitHub-token usage, merged-PR resolution) intact.
- `ci-pr-changelog-authoring`: the existing requirement that pins validation to the JS function names `parseChangelogSectionFull` / `validateChangelogSectionFull` must be re-phrased in terms of observable validation rules so the requirement is satisfied by either the JS implementation (today) or the Go implementation (after migration).

## Impact

- **Added code**: `scripts/changelog/` Go package (main + internal subpackages for `section` parser, `rewriter`, `engine`, `evidence`, `prcheck`, plus tests).
- **Removed code**: `.github/scripts/workflows/lib/changelog-*.js` (8 files, ~1,300 LOC), `.github/scripts/workflows/lib/pr-changelog-*.js` (2 files, ~330 LOC), `.github/scripts/workflows/changelog/*.js` (4 files, ~210 LOC), `.github/scripts/workflows/pr-changelog-check/*.js` (1 file, ~85 LOC), and their `.test.mjs` siblings.
- **Modified workflows**: `.github/workflows/changelog-generation.yml`, `.github/workflows/pr-changelog-check.yml`, and any agentic `.md` workflows that currently invoke the JS modules (regenerated `.lock.yml` files).
- **Modified Makefile**: `workflow-test` enumeration scope shrinks; `go test ./scripts/changelog/...` is implicitly covered by the existing `go test ./scripts/...` patterns.
- **Dependencies**: Reuses the project's existing Go toolchain. `github.com/google/go-github` is already pulled in by `scripts/auto-approve/` (with `golang.org/x/oauth2`); the new module pins the same major version.
- **No external behaviour change**: PR authors, release operators, and downstream consumers see identical `CHANGELOG.md` output, identical PR-check verdicts, and identical comment formatting. Only the runtime language changes.
- **Cold-start cost**: Each invocation pays one `go run` compilation (~5s on a runner that already executed `setup-go`). Workflows already perform `setup-go`, so the net delta is single-digit seconds per job.
