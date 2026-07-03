## Context

The `provider.yml` CI workflow runs `make testacc` on every PR across a static matrix of 20+ Elastic Stack versions × 2 shards = 40+ parallel jobs, each spinning up a full Elastic Stack and running ~50 acceptance test packages. Most PR changes touch one or two resources, and the only acceptance packages that actually need to run are those that test (or depend on) the changed code.

The provider codebase has a consistent, greppable structure: every resource and data source declares its Terraform type name via `entitycore.NewResourceBase`, `entitycore.NewElasticsearchResource`, `entitycore.NewKibanaResource`, or `entitycore.NewKibanaDataSource`. Test fixtures live under `internal/<domain>/<resource>/testdata/` as `.tf` files and reference resource type strings directly. This structure is stable enough to support automated analysis.

## Goals / Non-Goals

**Goals:**
- A `make targeted-testacc` target that runs only the acceptance test packages relevant to the current branch diff.
- A `make targeted-testacc-dry-run` target that prints the selection rationale without running tests.
- CI on PRs uses `targeted-testacc`; pushes to `main` and `workflow_dispatch` always use the full `testacc` suite.
- Shard count is dynamic: small targeted sets run on a single shard; larger sets split across the static `shard: [0, 1]` CI matrix.
- The static CI matrix (versions, runner assignments, include entries) is preserved unchanged.
- Stack startup is skipped entirely for matrix jobs where a shard has no packages to test.

**Non-Goals:**
- Changing how `make testacc` works.
- Dynamically altering the number of CI matrix shards via `fromJSON` — the static `shard: [0, 1]` matrix stays as-is.
- Covering non-Go file changes (docs, openspec) with acceptance tests — these correctly result in 0 selected packages.
- Modifying any existing test code or resource implementation.

## Decisions

### Decision: Two-phase package selection

**Choice:** Union of (1) Go reverse-dependency walk and (2) TF entity name grep.

**Rationale:** The two phases are complementary and non-overlapping.

Phase 1 (reverse-dep walk): catches packages that *import* changed code — e.g. changing `internal/kibana/dashboard/panel/lensdashboardapp` triggers `internal/kibana/dashboard` because `dashboard` imports `lensdashboardapp`. Uses `go list -f '{{.ImportPath}} {{join .Imports " "}}'` to build a forward-dep map, then inverts it. Non-test imports only (avoids test-only dep cascades).

Phase 2 (entity grep): catches test suites that *use* changed resources in their testdata but have no Go import relationship — e.g. `internal/fleet/agentpolicy` uses `elasticstack_kibana_space` in its testdata `.tf` files but doesn't import `internal/kibana/spaces`. Phase 1 would miss this. Entity names are extracted from the changed package's source via regex on `NewResourceBase`/`NewElasticsearchResource`/`NewKibanaResource`/`NewKibanaDataSource` calls; full type name is `elasticstack_<component>_<name>`. The grep targets both `testdata/**/*.tf` and `*_test.go` files.

**Alternative considered:** Phase 1 alone. Rejected: misses cross-domain testdata consumers (fleet tests using kibana_space, etc.).

**Alternative considered:** Phase 2 alone. Rejected: misses structural dependencies like panel sub-packages that are imported by but don't own a resource name.

### Decision: Force-all prefix table

**Choice:** Certain path prefixes unconditionally emit the full package set, bypassing analysis.

Prefixes: `provider/`, `internal/acctest/`, `internal/clients/`, `internal/entitycore/`, `generated/`.

**Rationale:** These packages have test-only import paths (provider, acctest) or fan out to 70+ importers (clients: 74, entitycore: 77, generated/kbapi: 69) — all above the 70% threshold of 101 total acc-test packages. Running analysis on them is pointless; the result will always be "run all". Hard-coding them avoids false confidence in partial analysis and keeps the tool fast.

A 70% threshold (`run-all-threshold`) also acts as a safety net for any shared package not in the table: if phase 1+2 selects more than ~71 packages, the tool emits all packages.

### Decision: Shard-count determined by the tool, not CI matrix

**Choice:** Tool takes `--total-shards` and `--shard-index` (defaulting to 1 and 0). Internally, if `|selected_packages| < min_shard_packages` (default 30), shard index 0 emits all packages and all other shard indices emit nothing. Otherwise, round-robin applies.

**Rationale:** The CI matrix is static (`shard: [0, 1]`). The only way to avoid spinning up unnecessary stack instances for shard 1 on small targeted runs is to skip the expensive steps when the shard has no packages. A `compute-packages` step before stack startup gates all downstream steps via `steps.targeted.outputs.has_packages == 'true'`. The tool handles the "single effective shard" logic internally; CI passes `--total-shards=2 --shard-index=${{ matrix.shard }}` unconditionally.

The threshold of 30 was chosen from the consumer count distribution: nearly all single-resource changes produce fewer than 30 packages (most produce 1–8). The two highest-consumption resources (`elasticstack_kibana_space` at 32 packages and `elasticstack_kibana_dashboard` at 28) straddle the threshold, which is appropriate — space is a foundational fixture used across the entire suite.

**Alternative considered:** Dynamic matrix via `fromJSON` from a pre-flight job. Rejected: would require restructuring `include:` entries for version-specific runner assignments, and adds a new job that must complete before any test job starts.

### Decision: Tool location and invocation

**Choice:** `scripts/targeted-testacc/` (package main, within the module). Invoked via `go run ./scripts/targeted-testacc/...`.

**Rationale:** Consistent with `scripts/schema-coverage-rotation/` and `scripts/auto-approve/`. No new `go tool` entry needed. `golang.org/x/tools v0.45.0` is already in `go.mod` and available for dep-graph construction.

### Decision: Git diff baseline

**Choice:** Auto-detect: try `git merge-base origin/main HEAD`, fall back to `HEAD~1` if no remote or shallow clone prevents merge-base resolution. Override via `--base` flag or `TARGETED_TESTACC_BASE` env var.

**Rationale:** Merge-base is the correct semantic for "what this branch changed". The fallback to `HEAD~1` handles shallow clones and detached HEAD states gracefully. In CI, a `git fetch origin main --depth=1` step before the tool ensures merge-base works without a full history fetch.

Empty diff (on main, or when only non-code files changed) → tool emits all acc-test packages (conservative default).

### Decision: CI event routing

**Choice:** The `compute-packages` step in each test matrix job checks `github.event_name`. For non-PR events (push to main, workflow_dispatch), it sets `has_packages=true` and `targeted_pkgs=` (empty) unconditionally. The test step then runs `make testacc` (full suite) instead of `make targeted-testacc`. For PR events, the step runs the tool and sets `has_packages` and `targeted_pkgs` from its output.

**Rationale:** This keeps the event-type logic in a single step. All downstream `if:` conditions use only `steps.targeted.outputs.has_packages == 'true'` — they don't need to re-check event type. The test step distinguishes targeted vs full by whether `targeted_pkgs` is non-empty.

## Risks / Trade-offs

**[Risk] False negatives — a relevant test package is not selected.**
→ The force-all prefix table and 70% threshold are conservative safety nets. The two-phase approach covers both Go structural dependencies and testdata consumers. The most likely gap is a test suite that uses a resource in an inline Go string (not a testdata `.tf` file) without the standard `resource "elasticstack_..."` prefix — these are rare and fall back to being caught by phase 1 if there's any import relationship.

**[Risk] Tool performance degrades as the codebase grows.**
→ `go list ./internal/...` takes ~0.5s today; grep across testdata takes ~0.2s. Both scale linearly with codebase size. Target: tool completes in under 5s. The `go list` result is not cached between shard jobs in CI (each job runs the tool independently), but at 3–5s per invocation this is acceptable.

**[Risk] Shard 1 jobs waste CI minutes on setup+teardown when they have no packages.**
→ Accepted. Checkout + setup-go + get-dependencies + compute-packages takes ~2–3 minutes before the tool outputs `has_packages=false`. All expensive steps (fleet image pull ~3min, stack start ~5min, stack wait ~3min) are skipped. Net waste per unnecessary shard-1 job: ~3 min. With 25 matrix entries × 1 unnecessary shard-1 job each = 75 min of CI machine time, all running in parallel so elapsed impact is ~3 min. This is acceptable vs. the alternative of restructuring the matrix.

**[Risk] Entity name regex fails to extract names from non-standard source patterns.**
→ All resources in the codebase use one of four consistent patterns. The regex is validated against the full codebase as part of the tool's unit tests. New resources that deviate from these patterns would need the regex updated — this is a known, low-frequency maintenance burden.

## Migration Plan

1. Implement and test the Go tool locally (`scripts/targeted-testacc/`).
2. Add `make targeted-testacc` and `make targeted-testacc-dry-run` targets.
3. Update `provider.yml`: add `compute-packages` step, gate expensive steps, route test step by event type.
4. Validate on a real PR branch with `make targeted-testacc-dry-run` before merging.
5. No rollback complexity — `make testacc` remains unchanged and always available.

## Open Questions

None — resolved during exploration.
