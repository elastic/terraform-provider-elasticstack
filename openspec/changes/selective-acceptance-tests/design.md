## Context

The `provider.yml` CI workflow runs `make testacc` on every PR across a static matrix of 20+ Elastic Stack versions × 2 shards = 40+ parallel jobs, each spinning up a full Elastic Stack and running ~68 acceptance test packages per shard (136 packages across 2 shards). Most PR changes touch one or two resources, and the only acceptance packages that actually need to run are those that test (or depend on) the changed code.

The provider codebase has a consistent, greppable structure: entity declarations use a set of entity-declaring constructors exported by `entitycore` (`NewResourceBase`, `NewDataSourceBase`, `NewEphemeralBase`, `NewActionBase`, `NewElasticsearchResource`, `NewElasticsearchDataSource`, `NewElasticsearchEphemeralResource`, `NewElasticsearchAction`, `NewKibanaResource`, `NewKibanaDataSource`, `NewKibanaEphemeralResource`, `NewKibanaAction`). The tool covers this listed set; a guard unit test asserts that every exported `entitycore` constructor is classified, so future envelope types fail the test rather than silently producing selection gaps. Test fixtures live under `internal/<domain>/<resource>/testdata/` as `.tf` files and reference resource type strings directly. This structure is stable enough to support automated analysis. A known accepted extraction gap: the SDKv2 `elasticstack_elasticsearch_ingest_processor_*` data sources do not use entitycore constructors; all 40 live in `internal/elasticsearch/ingest`, so phase 1 still selects that package when it changes, and only cross-package testdata consumers of those names would be missed.

## Goals / Non-Goals

**Goals:**

- A `make targeted-testacc` target that runs only the acceptance test packages relevant to the current branch diff.
- A `make targeted-testacc-dry-run` target that prints the selection rationale without running tests.
- CI on PRs uses `targeted-testacc`; pushes to `main`, merge queue (`merge_group`), and `workflow_dispatch` always use the full `testacc` suite.
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

Phase 2 (entity grep): catches test suites that *use* changed resources in their testdata but have no Go import relationship — e.g. `internal/fleet/agentpolicy` uses `elasticstack_kibana_space` in its testdata `.tf` files but doesn't import `internal/kibana/spaces`. Phase 1 would miss this. Entity names are extracted from every candidate package — the changed Go packages plus the packages phase 1 selected — via regex on the entity-declaring constructors exported by `entitycore` (see Context for the covered list); full type name is `elasticstack_<component>_<name>`. Including the phase-1 packages closes the sub-package gap: a change in a sub-package imported by an entity-declaring package (e.g. `internal/kibana/dashboard/panel/lenspie`, imported by `internal/kibana/dashboard`) selects the sub-package structurally via phase 1 but would not expose `elasticstack_kibana_dashboard` to phase 2 unless phase-1 packages are also scanned, so cross-package testdata consumers such as `internal/kibana/streams` would be missed. The grep targets both `testdata/**/*.tf` and `*_test.go` files.

**Alternative considered:** Phase 1 alone. Rejected: misses cross-domain testdata consumers (fleet tests using kibana_space, etc.).

**Alternative considered:** Phase 2 alone. Rejected: misses structural dependencies like panel sub-packages that are imported by but don't own a resource name.

### Decision: Force-all prefix table

**Choice:** Certain path prefixes and module-level files unconditionally emit the full package set, bypassing analysis.

Prefixes: `provider/`, `internal/acctest/`, `internal/clients/`, `internal/entitycore/`, `generated/`, `xpprovider/`, `.github/workflows/`, `examples/`, `scripts/targeted-testacc/`, plus shared acceptance-test helper packages `internal/kibana/dashboard/dashboardacctest/`, `internal/kibana/dashboard/panelkit/contracttest/`, and `internal/providerfwtest/`.

The `examples/` tree is force-all for the same reason: it is consumed only from test code — `internal/acctest/examples_plan_test.go` embeds `examples/` and runs `TestAccExamples_planOnly`, a plan-only suite over every embedded example — so an examples-only diff sits outside the enumeration roots and outside the `go list` pattern that builds the import graph, so phase 1 has no reverse-dependency edges into it, would otherwise select zero packages on every shard, and the unit-test job's `-skip '^TestAcc'` excludes that suite too.

The tool's own package (`scripts/targeted-testacc/`) is force-all because it sits outside `internal/` and under no other prefix, so a PR touching only the tool would otherwise select zero acceptance packages — the suite the tool gates would never be exercised by its own changes. This matches the shared test-helper reasoning (avoids false confidence in partial analysis); tool changes are low-frequency, so the CI cost of the resulting full-suite runs is bounded.

Files: `go.mod`, `go.sum`, `Makefile`, `main.go`, `.terraform-version`, `.env.template`, and `docker-compose*.yml` / `docker-compose*.yaml` matched on base name at any repository path (e.g. `docker-compose.yml`, `docker-compose.tls.yml`) — these affect every build or the test harness itself (the root `main.go` is the provider server entry point; `.terraform-version` pins the Terraform binary used by acceptance tests; `.env.template` supplies defaults for every test run), so a diff touching them cannot be narrowed to a subset of packages.

**Trade-off — non-test imports only:** Phase 1 builds its reverse-dep graph from `go list` non-test imports over `internal/` and `provider/`, so packages imported *only from test files* are invisible to the reverse walk (e.g. a change to `internal/kibana/dashboard/dashboardacctest` would otherwise not select the 33 dashboard acceptance test consumers). The shared helper packages above are force-all for exactly this reason; a guard unit test scans the test imports of both `internal/` and `provider/` (provider/ itself is force-all, so a test-only-imported package under it cannot silently skip, and a package under internal/ may be test-imported only from provider/'s test files — `internal/acctest`, which has no non-test importer anywhere in the module, is test-imported from provider/) including `TestImports`/`XTestImports` and fails when a package is imported only from test files but is neither force-all nor entity-declaring.

**Rationale:** These packages have test-only import paths (provider, acctest) or fan out to nearly the entire acceptance suite (transitive importers of 136 total acc-test packages: internal/clients — 130/136 (96%), internal/entitycore — 130/136 (96%), generated/kbapi — 131/136 (96%)) — all above the 70% threshold of 136 total acc-test packages. Running analysis on them is pointless; the result will always be "run all". Hard-coding them avoids false confidence in partial analysis and keeps the tool fast.

A 70% threshold (`run-all-threshold`) also acts as a safety net for any shared package not in the table: if phase 1+2 selects more than 70% of the 136 acc-test packages, the tool emits all packages.

### Decision: Shard-count determined by the tool, not CI matrix

**Choice:** Tool takes `--total-shards` and `--shard-index` (defaulting to 1 and 0). Internally, if `|selected_packages| < min_shard_packages` (default 30), shard index 0 emits all packages and all other shard indices emit nothing. Otherwise, round-robin applies.

**Rationale:** The CI matrix is static (`shard: [0, 1]`). The only way to avoid spinning up unnecessary stack instances for shard 1 on small targeted runs is to skip the expensive steps when the shard has no packages. A `compute-packages` step before stack startup gates all downstream steps via `steps.targeted.outputs.has_packages == 'true'`. The tool handles the "single effective shard" logic internally; CI passes `--total-shards=2 --shard-index=${{ matrix.shard }}` unconditionally.

The threshold of 30 was chosen from the consumer count distribution: nearly all single-resource changes produce fewer than 30 packages (most produce 1–8). The two highest-consumption resources (`elasticstack_kibana_space` at 39 packages and `elasticstack_kibana_dashboard` at 37) both exceed the threshold and split across both shards, which is appropriate — they are foundational fixtures used across the suite.

**Alternative considered:** Dynamic matrix via `fromJSON` from a pre-flight job. Rejected: would require restructuring `include:` entries for version-specific runner assignments, and adds a new job that must complete before any test job starts.

### Decision: Tool location and invocation

**Choice:** `scripts/targeted-testacc/` (package main, within the module). Invoked via `go run ./scripts/targeted-testacc/...`.

**Rationale:** Consistent with `scripts/schema-coverage-rotation/` and `scripts/auto-approve/`. No new `go tool` entry needed. `golang.org/x/tools v0.45.0` is already in `go.mod` and available for dep-graph construction.

### Decision: Git diff baseline

**Choice:** Auto-detect: try `git merge-base origin/main HEAD`, fall back to `HEAD~1` if no remote or shallow clone prevents merge-base resolution. Override via `--base` flag or `TARGETED_TESTACC_BASE` env var.

**Rationale:** Merge-base is the correct semantic for "what this branch changed". The fallback to `HEAD~1` handles shallow clones and detached HEAD states gracefully. In CI, a `git fetch origin main --depth=1` step before the tool ensures merge-base works without a full history fetch.

Empty or unresolvable diff (on main, shallow clone, or `git diff` failure) → tool emits all acc-test packages (conservative default). A diff that resolves but contains only non-code files emits zero packages (see Non-Goals).

### Decision: Dedicated stackless unit-test job

**Choice:** Add a dedicated `unit-test` job to `provider.yml` that runs `go test ./... -skip '^TestAcc'` on every event where the change-classification job reports `provider_changes=true`. It is gated on the classification result and deliberately **not** on `steps.targeted.outputs.has_packages`, so unit tests still run when targeting selects zero acceptance packages. Its result is wired into the `gate` job via `PROVIDER_GATE_UNIT_TEST_RESULT` and `gate-provider.js`.

**Rationale:** The targeted-testacc selection only covers packages containing acceptance tests (`func TestAcc` or `resource.Test`/`resource.ParallelTest`), so unit-test-only packages (`internal/entitycore`, `generated/kbapi`, `internal/clients/*`, `internal/fleet/policyshape`, `internal/asyncutils`, `internal/diagutil`) would otherwise never run on PRs. Before this change they rode along inside the blocking acceptance `test` job (`make testacc` has no `-run`/`-skip` filter and shards over `go list ./...`); after it they run only in the dedicated job, so the job must gate the PR. `-skip '^TestAcc'` is sufficient isolation in a stackless job because `resource.Test` self-skips without `TF_ACC` and `versionutils.SkipIfUnsupported` returns early, so acceptance suites are skipped rather than failing; several `*ExplicitConnection` tests assert endpoint env vars before the framework gate, which is why a plain `go test ./...` cannot run without a stack. This introduces a new required check surfaced through Provider Gate, which is the user-visible CI change this decision records.

### Decision: CI event routing

**Choice:** The `compute-packages` step in each test matrix job checks `github.event_name`. For non-PR events (push to main, `workflow_dispatch`, `merge_group`), it sets `has_packages=true` and `targeted_pkgs=` (empty) unconditionally. The test step then runs `make testacc` (full suite) instead of `make targeted-testacc`. For PR events (opened, synchronize, reopened), the step runs the tool and sets `has_packages` and `targeted_pkgs` from its output.

**Rationale:** This keeps the event-type logic in a single step. All downstream `if:` conditions use only `steps.targeted.outputs.has_packages == 'true'` — they don't need to re-check event type. The test step distinguishes targeted vs full by whether `targeted_pkgs` is non-empty. Merge queue runs are treated as non-PR events because they validate the final merged state, which requires the full suite as the authoritative gate.

## Risks / Trade-offs

**[Risk] False negatives — a relevant test package is not selected.**
→ The force-all prefix table and 70% threshold are conservative safety nets. The two-phase approach covers both Go structural dependencies and testdata consumers. Phase 2 extracts entity names from the union of the changed packages and the phase-1 selection, which closes the sub-package gap: a change in a sub-package imported by an entity-declaring package (e.g. `internal/kibana/dashboard/panel/lenspie` → `internal/kibana/dashboard` → cross-package consumers like `internal/kibana/streams`) now also selects its testdata consumers. The most likely remaining gap is a test suite that uses a resource in an inline Go string (not a testdata `.tf` file) without the standard `resource "elasticstack_..."` prefix — these are rare and fall back to being caught by phase 1 if there's any import relationship. The SDKv2 `elasticstack_elasticsearch_ingest_processor_*` data sources remain an accepted gap (see Context).

**[Risk] Tool performance degrades as the codebase grows.**
→ `go list ./internal/... ./provider/...` takes ~0.5s today; grep across testdata takes ~0.2s. Both scale linearly with codebase size. Target: tool completes in under 5s. The `go list` result is not cached between shard jobs in CI (each job runs the tool independently), but at 3–5s per invocation this is acceptable.

**[Risk] Shard 1 jobs waste CI minutes on setup+teardown when they have no packages.**
→ Accepted. Checkout + setup-go + get-dependencies + compute-packages takes ~2–3 minutes before the tool outputs `has_packages=false`. All expensive steps (fleet image pull ~3min, stack start ~5min, stack wait ~3min) are skipped. Net waste per unnecessary shard-1 job: ~3 min. With 25 matrix entries × 1 unnecessary shard-1 job each = 75 min of CI machine time, all running in parallel so elapsed impact is ~3 min. This is acceptable vs. the alternative of restructuring the matrix.

**[Risk] Entity name regex fails to extract names from non-standard source patterns.**
→ entitycore exports multiple entity-declaring constructors and the tool covers the listed set (see Context). The unit tests validate the regexes against synthetic source snippets and against real repository call sites, so a regex that stops matching the tree fails loudly; constructor drift is instead caught by the guard test that asserts every exported `entitycore` constructor is classified, and by the guard test for test-only-imported packages. The SDKv2 `elasticstack_elasticsearch_ingest_processor_*` data sources are a known accepted gap (see Context). New resources that deviate from the covered patterns would need the regex updated — this is a known, low-frequency maintenance burden.

**[Note] Makefile shellcheck suppression scope.** The Makefile carries `# shellcheck disable=SC1073,SC1065,SC1064,SC1072` directives placed directly above each `define`/`$(eval $(call ...))` block that uses make function syntax shellcheck cannot parse; they are scoped to those blocks only (not file-wide).

## Migration Plan

1. Implement and test the Go tool locally (`scripts/targeted-testacc/`).
2. Add `make targeted-testacc` and `make targeted-testacc-dry-run` targets.
3. Update `provider.yml`: add `compute-packages` step, gate expensive steps, route test step by event type.
4. Validate on a real PR branch with `make targeted-testacc-dry-run` before merging.
5. No rollback complexity — `make testacc` remains unchanged and always available.

## Open Questions

None — resolved during exploration.
