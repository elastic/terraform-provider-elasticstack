## Context

Changelog handling in this repo currently spans two parallel JavaScript codebases under `.github/scripts/workflows/`:

1. **Engine cluster** (~1,510 LOC, 13 files): `lib/changelog-engine-factory.js`, `lib/changelog-renderer.js`, `lib/changelog-pr-management.js`, `lib/changelog-rewriter.js`, `lib/changelog-pr-evidence.js`, `lib/changelog-release-context.js`, `lib/changelog-engine-workflow.js`, `lib/changelog-evidence-manifest.js`, `lib/changelog-engine.js`, plus `changelog/{run-changelog-engine,gather-pr-evidence,manage-unreleased-pr,refresh-release-pr}.js`. Consumed by `.github/workflows/changelog-generation.yml`.
2. **PR check cluster** (~412 LOC, 3 files): `lib/pr-changelog-parser.js`, `lib/pr-changelog-check.js`, `pr-changelog-check/check.js`. Consumed by `.github/workflows/pr-changelog-check.yml`.

Both clusters parse the same `## Changelog` contract from PR bodies (customer impact, summary, breaking-changes subsection), but each has its own parser implementation today. The engine cluster additionally handles `CHANGELOG.md` rewriting, semver-aware tag selection (`git tag --list "v[0-9]*..." --sort=-version:refname`), link-table maintenance, evidence manifest gathering, and PR-comment/file management on the `generated-changelog` branch.

Tests live in `.test.mjs` files next to each module and run via `node --test` (invoked from `make workflow-test`). They cover parser corner cases, mode validation, semver fall-back, and rendering invariants.

The rest of the repo's automation scripts (`scripts/auto-approve/`, `scripts/kibana-spec-impact/`, `scripts/ci-deadcode-removal-rotation/`, `scripts/schema-coverage-rotation/`) are written in Go, follow a stable convention (`go-github/v86`, `$GITHUB_EVENT_PATH`, `$GITHUB_OUTPUT`, `os.Exit(1)` on failure), and are exercised by `go test ./scripts/...` in CI. Migrating both changelog clusters to a single Go tool removes parser duplication, harmonises the automation stack on Go, and exploits the same review/test infrastructure.

## Goals / Non-Goals

**Goals:**

- Replace both changelog clusters with a single Go module at `scripts/changelog/` exposing five subcommands: `gather-evidence`, `run-engine`, `manage-unreleased-pr`, `refresh-release-pr`, `validate-pr-section`.
- Provide one canonical `## Changelog` section parser inside `scripts/changelog/internal/section/` consumed by both engine-mode rendering and PR-body validation. Eliminate the today-parallel JS parsers.
- Preserve **exact behavioural parity** for every existing scenario in the `ci-changelog-generation` and `ci-pr-changelog-authoring` specs. The output `CHANGELOG.md` diff, PR-check verdict, and comment payload must be byte-identical (or trivially equivalent) to the JS implementation on the same inputs.
- Port `.test.mjs` test cases to `_test.go` cases that exercise the same inputs and expected outputs.
- Update `.github/workflows/changelog-generation.yml` and `.github/workflows/pr-changelog-check.yml` to invoke `go run ./scripts/changelog <subcommand>` after the existing `setup-go` step; remove the corresponding `actions/github-script` blocks.

**Non-Goals:**

- Changing the `## Changelog` PR-body contract (customer impact values, summary requirement, breaking-changes subsection format). Out of scope; that's owned by `ci-pr-changelog-authoring`.
- Changing the `CHANGELOG.md` section layout, link-table format, or `## [Unreleased]` / `## [x.y.z] - <date>` markers. Out of scope; owned by `ci-changelog-generation`.
- Changing which workflows trigger changelog automation, the `no-changelog` label semantics, or the `generated-changelog` branch convention.
- Migrating any other workflow JS module (factory runners, openspec-verify, classifier, etc.). Those are tracked as separate future changes.
- Producing a pre-built static binary distributed via release artifacts. The tool is `go run` from source like every other `scripts/*` tool in this repo.

## Decisions

### D1. Single Go module with subcommands, not multiple separate tools

The five entry points (`gather-evidence`, `run-engine`, `manage-unreleased-pr`, `refresh-release-pr`, `validate-pr-section`) share the same parser, GitHub client, semver helpers, `$GITHUB_OUTPUT` writer, and error-handling conventions. A single `main` with subcommand routing keeps that shared infrastructure colocated.

- **Why over separate `scripts/changelog-engine/` + `scripts/pr-changelog-check/`?** The parser shared between them is the strongest argument for the migration. Two tools would force the parser into a third package and re-impose a cross-tool import boundary that adds nothing.
- **Why over a generic `scripts/workflow-runner` with all script logic?** Bloats the scope; the changelog cluster is self-contained and deserves its own module.
- **Subcommand routing**: use the standard library `flag.NewFlagSet` per subcommand, dispatched from `main.go` by `os.Args[1]`. Avoids adding `cobra`/`urfave/cli`. Aligns with `scripts/kibana-spec-impact/main.go` style.

### D2. Reuse `go-github/v86` and the existing GitHub client convention

`scripts/auto-approve/main.go` already establishes the pattern: read `GITHUB_TOKEN`, build an `oauth2`-wrapped HTTP client, instantiate `*github.Client`. The changelog tool follows the same pattern verbatim — pinned to the same `go-github` major version to share a single transitive dependency.

- **Why over `cli/go-gh`?** The repo doesn't use it elsewhere; introducing a second GitHub library is unjustified churn.
- **Why over shelling out to `gh`?** PR listing, file comparisons, paginated tag walks, and comment management are noticeably cleaner with the typed client. The existing scripts agree.

### D3. Shared section parser lives in `scripts/changelog/internal/section/`

The parser is the keystone of the migration. Both the engine (which assembles `CHANGELOG.md` sections from merged-PR bodies) and the PR check (which validates the PR body at PR open/sync/edit time) call the same `section.Parse([]byte) (Section, error)` API.

```go
package section

type CustomerImpact int
const (
    ImpactNone CustomerImpact = iota
    ImpactFix
    ImpactEnhancement
    ImpactBreaking
)

type Section struct {
    CustomerImpact   CustomerImpact
    Summary          string  // required when CustomerImpact != ImpactNone
    BreakingChanges  string  // optional subsection content
    Raw              string  // verbatim extracted text for downstream rendering
}

func Parse(body []byte) (Section, error)         // PR body → Section
func Render(sec Section, opts RenderOpts) string // Section → CHANGELOG.md entry
```

- The parser preserves the fenced-code handling from `pr-changelog-parser.js` (lines 36-50: open fences for `` ``` `` and `~~~`, closing on matching tokens, terminating the section on the next `## ` heading outside a fence).
- A single set of golden-file tests under `scripts/changelog/internal/section/testdata/` covers every PR-body shape currently in `.github/scripts/workflows/lib/pr-changelog-parser.test.mjs`.

### D4. `internal/` packages partition concerns

```
scripts/changelog/
  main.go                          subcommand router
  internal/
    section/                       PR-body parser + CHANGELOG renderer (the shared core)
    rewriter/                      CHANGELOG.md section/link-table rewriter
    engine/                        unreleased + release mode orchestration
    evidence/                      gather-evidence manifest + format
    prcheck/                       PR-body validation pipeline (uses section.Parse)
    githubx/                       go-github client construction, $GITHUB_OUTPUT helpers, env parsing
    semver/                        TAG_LIST_CMD wrapper, previous-tag selection, compare-range
```

- `internal/` enforces a tight import surface; nothing outside `scripts/changelog/` can depend on these packages.
- `githubx` is the only package allowed to read env vars or call out to git; everything else takes typed inputs. This makes the unit-test seam clean.

### D5. Per-subcommand entry contract

Each subcommand:

1. Parses subcommand-specific flags via `flag.NewFlagSet`.
2. Loads context from `$GITHUB_EVENT_PATH` (PR number, head SHA, etc.) using `internal/githubx/event.go`.
3. Executes its pipeline.
4. Writes results to `$GITHUB_OUTPUT` via `internal/githubx/output.go` (handles EOF heredoc framing, mirroring the JS `core.setOutput` semantics).
5. Returns non-zero on failure with `fmt.Fprintf(os.Stderr, ...)`.

This mirrors `scripts/auto-approve/main.go` and `scripts/kibana-spec-impact/main.go` so reviewers see one shape across the repo.

### D6. Workflow wiring

For each step that currently uses `actions/github-script` with `require('${{ github.workspace }}/.github/scripts/workflows/.../X.js')`, replace with:

```yaml
- name: <step name>
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    GITHUB_REPOSITORY: ${{ github.repository }}
    # ... step-specific inputs as env vars
  run: go run ./scripts/changelog <subcommand> [flags]
```

- The pre-existing `actions/setup-go` step in `changelog-generation.yml` and `pr-changelog-check.yml` already provides the toolchain; no new setup is added.
- All workflows that invoke the JS modules currently checkout the repo already (per `remove-workflow-template-engine`), so `./scripts/changelog` resolves.
- Outputs that the YAML downstream references (e.g. `steps.X.outputs.changelog_status`) are preserved by writing the same `<name>=<value>` pairs to `$GITHUB_OUTPUT`.

### D7. Test parity, not test rewrite

Every `.test.mjs` case in the doomed clusters maps to an exactly-equivalent Go test:

- Parser corner cases → `section_test.go` golden files.
- Engine mode validation → `engine_test.go` table tests.
- Renderer invariants → `section_test.go` (round-trip Parse→Render).
- Rewriter idempotence → `rewriter_test.go`.
- PR-check verdicts → `prcheck_test.go` table tests.

Tests run under `go test ./scripts/changelog/...` and are picked up by the existing `Workflow tests` job (already invokes `go test ./scripts/kibana-spec-impact/...`); the Makefile target adds the new path.

### D8. Cut-over strategy: parallel then atomic switch

Implementation order avoids a long-lived dual-stack:

1. Land the Go module + tests with no workflow wiring change (it's a no-op build).
2. Switch the workflows over in one commit (both `changelog-generation.yml` and `pr-changelog-check.yml`, plus any `.lock.yml` regen).
3. Delete the JS clusters and their tests in the same commit. Removes any temptation to maintain both.

This keeps the PR reviewable (Go addition is one diff, workflow switch + JS removal is another), and the working tree is never in a half-migrated state.

## Risks / Trade-offs

- **Cold-start cost on every step** → Each `go run ./scripts/changelog X` triggers a Go build. On the changelog-generation workflow this happens 3–5 times per run. Mitigation: build once per job into a temp binary and reuse, OR rely on Go's build cache (which is warm after the first invocation in a job since `setup-go` enables `cache: true`). Expected real-world cost: 2–4s per invocation after cache warm-up. Acceptable.
- **Two-parser drift during migration** → For the brief window where both implementations exist (D8 step 1), a contributor could edit only the JS side. Mitigation: D8's atomic switch lands JS deletion in the same commit as workflow switch; nobody touches JS in between.
- **Different error messages** → `core.setFailed` strings from JS and `fmt.Fprintf(os.Stderr, ...)` strings from Go won't be byte-identical. Mitigation: the user-visible verdict strings (PR comments, `gate_reason`, etc.) are pinned by tests; internal logs may differ but aren't part of any spec.
- **`go-github` major-version drift** → The repo currently uses `go-github/v86`. Mitigation: pin to the same version explicitly in `go.mod`; do not introduce a fresh major.
- **Workflow YAML lock-file churn** → Agentic workflows compile from `.md` to `.lock.yml`. Mitigation: run `make workflow-generate` in the same commit that updates the YAML, exactly as `remove-workflow-template-engine` did.
- **Reviewer load** → ~1,900 LOC removed, ~1,200 LOC added. Mitigation: split commits along D8's boundaries (Go land → workflow switch + JS removal → cleanup). Tests provide the parity guarantee.

## Migration Plan

1. **Add Go module skeleton** with `main.go`, `internal/section/` parser, and ported parser tests. Build green, no workflow changes.
2. **Add `engine/`, `rewriter/`, `evidence/`, `prcheck/`, `semver/`, `githubx/` packages** with ported tests. All `go test ./scripts/changelog/...` green.
3. **Switch workflows + delete JS** in one commit. Regenerate `.lock.yml`. Update `Makefile` to drop the JS test enumeration for these clusters.
4. **CI verification**: trigger the `changelog-generation` workflow against a controlled merge and the `pr-changelog-check` workflow against a PR with each customer-impact variant; assert byte-identical comment/diff output to a recent JS-produced reference run.
5. **No rollback flag needed**: the migration is reversible by revert. The JS clusters live in git history and would restore cleanly.

## Open Questions

- Should the tool live as `scripts/changelog/` (kebab → just `changelog`) or `scripts/changelog-tooling/` for symmetry with `ci-deadcode-removal-rotation/` and `schema-coverage-rotation/`? Recommendation: `scripts/changelog/`; the others have hyphenated names because their domain is multi-word, ours is a single noun.
- Do we keep `internal/githubx/` as a private helper, or hoist into a shared `scripts/internal/githubx/` package once a second Go script needs it? Recommendation: defer; YAGNI until a second consumer appears.
- Should `validate-pr-section` write its full verdict to a single `result_json` output (current JS shape) or to flat key/value outputs (`status`, `reason`, `summary_present`, …)? Recommendation: keep `result_json` for byte-identical compatibility with the existing PR-check workflow consumers; revisit later if a flat shape is preferred.
