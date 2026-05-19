# Changelog workflow tool (`scripts/changelog`)

Operator-focused entry points for changelog generation and PR-body validation. Behaviour and migration rationale are documented in OpenSpec change `openspec/changes/migrate-changelog-engine-to-go/` (`proposal.md`, `design.md`, and **design § D5** — per-subcommand entry contract).

Run from the repository root after `actions/checkout`; workflows use `actions/setup-go` plus:

```bash
go run ./scripts/changelog <subcommand> [flags]
```

## Subcommands (what each replaces)

| Subcommand | Replaces (historical JS) |
|------------|--------------------------|
| `gather-evidence` | `changelog/gather-pr-evidence.js`, `lib/changelog-pr-evidence.js` |
| `run-engine` | `changelog/run-changelog-engine.js`, `lib/changelog-engine*.js`, `lib/changelog-renderer.js`, `lib/changelog-rewriter.js`, `lib/changelog-release-context.js` |
| `manage-unreleased-pr` | `changelog/manage-unreleased-pr.js` plus unreleased-branch logic in `lib/changelog-pr-management.js` |
| `refresh-release-pr` | `changelog/refresh-release-pr.js` plus release-branch logic in `lib/changelog-pr-management.js` |
| `validate-pr-section` | `pr-changelog-check/check.js`, `lib/pr-changelog-check.js`, `lib/pr-changelog-parser.js` |

Shared parsing and rendering live under `internal/section/`; `run-engine` and `validate-pr-section` both consume that parser for merged PR bodies and PR-description checks respectively.

## Environment variables

Workflows rely on Actions-injected vars where applicable.

| Variable | Role |
|----------|------|
| `GITHUB_TOKEN` | Authenticated REST access (`go-github`). Required for API-backed subcommands. |
| `GITHUB_REPOSITORY` | `owner/repo` slug. |
| `GITHUB_EVENT_PATH` | Path to webhook JSON (`pull_request`, `workflow_dispatch`, etc.). Used to resolve PR number and event context where flags are omitted. |
| `GITHUB_OUTPUT` | File for step outputs (`key=value` and heredoc-framed multi-line values). |
| `MODE` | Engine / evidence mode: `unreleased` or `release` (see flags for overrides). |
| `TARGET_VERSION` | Release semver `X.Y.Z` without leading `v`. |
| `COMPARE_RANGE` | Git revision range driving merged-PR lookups (e.g. `tag..HEAD`). |
| `CHANGELOG_PATH` | Path to `CHANGELOG.md` (default `CHANGELOG.md`). |
| `TARGET_BRANCH` | Optional branch override for engine outputs after release runs. |
| `PREVIOUS_TAG` | Explicit previous release tag baseline for evidence (when wired by the workflow). |

Additional `INPUT_*` fallbacks mirror legacy `actions/core` wiring for dispatched workflows.

## Invocation examples (mirroring CI)

### Changelog generator job

After checkout, fetch depth, tags, and `setup-go`:

```bash
CHANGELOG_PATH=CHANGELOG.md go run ./scripts/changelog run-engine
```

Release follow-up steps (subset):

```bash
COMPARE_RANGE="$COMPARE_RANGE_FROM_ENGINE_OUTPUT" TARGET_VERSION="$TARGET_VERSION" \
  go run ./scripts/changelog refresh-release-pr
```

Unreleased branch after push:

```bash
COMPARE_RANGE="$COMPARE_RANGE_FROM_ENGINE_OUTPUT" \
  go run ./scripts/changelog manage-unreleased-pr
```

### PR changelog check

Minimal workflow step (`pull_request_target` provides `GITHUB_EVENT_PATH` automatically):

```bash
GITHUB_TOKEN=${{ secrets.GITHUB_TOKEN }} \
  go run ./scripts/changelog validate-pr-section
```

### Evidence manifest

When a job needs the standalone manifest (inputs depend on upstream steps):

```bash
go run ./scripts/changelog gather-evidence \
  --mode unreleased \
  --compare-range 'v1.2.3..HEAD'
```

Use `go run ./scripts/changelog <subcommand> -h` for per-command flags.

## Tests

```bash
go test ./scripts/changelog/... -count=1
```

Also exercised via `make workflow-test` alongside other workflow-related Go tests.
