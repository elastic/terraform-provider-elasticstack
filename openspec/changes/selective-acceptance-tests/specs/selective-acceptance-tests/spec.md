# `selective-acceptance-tests` — Targeted Acceptance Test Selection

## ADDED Requirements

### Requirement: targeted-testacc Go tool exists and is runnable

A Go `main` package SHALL exist at `scripts/targeted-testacc/` within the provider module. It SHALL be invocable via `go run ./scripts/targeted-testacc/...` from the module root without any build step or `go tool` entry. The tool SHALL complete within 10 seconds on the current codebase.

#### Scenario: Tool runs successfully from module root

- **WHEN** `go run ./scripts/targeted-testacc/...` is executed from the module root
- **THEN** the command exits 0 and emits a sorted, newline-separated list of Go import paths to stdout (or nothing if no relevant packages are found)

#### Scenario: Tool runs successfully with all flags

- **WHEN** the tool is invoked with `--total-shards=2 --shard-index=0 --base=origin/main --dry-run`
- **THEN** the command exits 0 and emits human-readable selection rationale (dry-run mode does not emit package paths to stdout)

---

### Requirement: Two-phase package selection

The tool SHALL compute the set of relevant acceptance test packages via two independent phases whose results are unioned and deduplicated. The acceptance test package enumeration SHALL walk the set of repository roots `internal/` and `provider/`: the `provider` package is an acceptance test package because `provider/provider_test.go` and `provider/factory_test.go` drive real live-stack suites via `resource.Test` without any `TestAcc`-prefixed function, with fixtures under `provider/testdata/**`.

**Phase 1 — Go reverse-dependency walk:** For each changed Go package, the tool SHALL walk the reverse import graph (non-test imports only) to find all packages that transitively import the changed package. The import graph SHALL be built over the enumerated roots (`./internal/... ./provider/...`), so reverse-dep edges into `provider/` exist and phase 1 can select the provider package's live-stack suites. Only acceptance test packages SHALL be included. An acceptance test package is a package containing at least one `*_test.go` file that declares a `func TestAcc*` function OR invokes the plugin-testing acceptance harness (`resource.Test`/`resource.ParallelTest`); the harness invocation is part of the definition because some acceptance suites (e.g. `internal/kibana/synthetics`, `provider`) predate the `TestAcc` naming convention and drive `resource.TestCase` purely via `resource.Test` without any `TestAcc`-prefixed function.

**Phase 2 — TF entity name grep:** For each candidate package — every changed Go package plus every package selected by phase 1 — the tool SHALL extract Terraform type name suffixes by scanning the package's non-test `.go` files (files ending in `_test.go` SHALL be excluded, so that entity declarations only in test source are ignored) for calls to the entity-declaring constructors exported by `internal/entitycore`:

- `NewResourceBase(entitycore.Component<X>, "<name>")`
- `NewDataSourceBase(entitycore.Component<X>, "<name>")`
- `NewEphemeralBase(entitycore.Component<X>, "<name>")`
- `NewActionBase(entitycore.Component<X>, "<name>")`
- `NewElasticsearchResource[...]("<name>", ...)`
- `NewElasticsearchDataSource[...](entitycore.Component<X>, "<name>", ...)` (the data source envelope is component-first; the resource/ephemeral/action envelopes are name-first)
- `NewElasticsearchEphemeralResource[...]("<name>", ...)`
- `NewElasticsearchAction[...]("<name>", ...)`
- `NewKibanaResource[...](entitycore.Component<X>, "<name>", ...)`
- `NewKibanaDataSource[...](entitycore.Component<X>, "<name>", ...)`
- `NewKibanaEphemeralResource[...]("<name>", ...)`
- `NewKibanaAction[...]("<name>", ...)`

For the generic constructors, the explicit type-argument list (`[...]`) SHALL be optional, so type-inferred call sites (e.g. `NewElasticsearchResource("synonym_set", opts)` where the type is inferred from the options value) are also extracted. A leading `Component<X>` argument SHALL be optional for the name-first families, so the same pattern covers both envelope styles.

A unit test SHALL verify that the extractor covers every constructor `internal/entitycore` exports, so future envelope types fail the test instead of silently producing selection gaps. A known accepted extraction gap is the SDKv2 `elasticstack_elasticsearch_ingest_processor_*` data sources: they are not declared via entitycore constructors, but all 40 live in `internal/elasticsearch/ingest`, so phase 1 still selects that package when it changes; only cross-package testdata consumers of those names could be missed.

The tool SHALL construct the full entity name as `elasticstack_<component>_<name>` (e.g. `elasticstack_kibana_space`). It SHALL then recursively grep the enumerated roots (`internal/` and `provider/`) for that string in `*.tf` and `*_test.go` files. The owning package of each matching file SHALL be included in the result set.

#### Scenario: Leaf resource change — direct package selected

- **WHEN** `internal/kibana/slo/resource.go` is the only changed file
- **THEN** `internal/kibana/slo` is selected (phase 1: direct package has TestAcc)
- **AND** packages whose testdata references `elasticstack_kibana_slo` are selected (phase 2)

#### Scenario: Panel sub-package change — parent dashboard package selected via reverse dep

- **WHEN** a file under `internal/kibana/dashboard/panel/lenspie/` is the only changed file
- **THEN** `internal/kibana/dashboard` is selected (phase 1: dashboard imports lenspie)
- **AND** `internal/kibana/dashboard/panel/lenspie` is also selected (direct)
- **AND** cross-package testdata consumers of `elasticstack_kibana_dashboard` (e.g. `internal/kibana/streams`) are selected (phase 2: the entity is extracted from the phase-1-selected `internal/kibana/dashboard` package)

#### Scenario: Shared resource consumed by cross-domain testdata

- **WHEN** `internal/kibana/spaces/resource.go` is the only changed file
- **THEN** packages in `internal/fleet/` whose testdata `.tf` files reference `elasticstack_kibana_space` are selected (phase 2)
- **AND** the `provider` package's live-stack connection suites are selected (phase 2: `provider/testdata/**/main.tf` fixtures reference `elasticstack_kibana_space`)
- **AND** `internal/kibana/spaces` itself is selected (phase 1)

#### Scenario: Non-TestAcc acceptance suite package is selected

- **WHEN** a file under `internal/kibana/synthetics/monitor/` is the only changed file
- **AND** the monitor package's `*_test.go` files drive acceptance suites via `resource.Test` without declaring any `func TestAcc`
- **THEN** `internal/kibana/synthetics/monitor` is selected (phase 1: it invokes the acceptance harness, so it is an acceptance test package)

---

### Requirement: Force-all prefix table

When any changed file path has a prefix matching one of the force-all prefixes, or equals one of the force-all files, the tool SHALL immediately emit all acceptance test packages (equivalent to a "run all" result) without performing phase 1 or phase 2 analysis.

Force-all prefixes: `provider/`, `internal/acctest/`, `internal/clients/`, `internal/entitycore/`, `generated/`, `xpprovider/`, `.github/workflows/`, `scripts/targeted-testacc/`, `internal/kibana/dashboard/dashboardacctest/`, `internal/kibana/dashboard/panelkit/contracttest/`, `internal/providerfwtest/`. The three shared acceptance-test helper packages are force-all because they are imported only from test files, so the phase-1 reverse-dependency walk (non-test imports) cannot see them. The `provider/` prefix is force-all, and because the provider package is itself an enumerated acceptance test package, a `provider/` change selects the full set including the provider package's own live-stack connection suites. The tool's own package (`scripts/targeted-testacc/`) is force-all because it sits outside the enumerated roots and under no other prefix, so a PR touching only the tool would otherwise select zero acceptance packages — the suite it gates would never be exercised by its own changes, which mirrors the shared test-helper reasoning: it avoids false confidence in partial analysis of the tool's behavior. Tool changes are low-frequency, so the CI cost of the resulting full-suite runs is bounded. A guard unit test SHALL fail when a package under `internal/` is imported only from test files and is neither covered by a force-all prefix nor entity-declaring.

Force-all files: `go.mod`, `go.sum`, `Makefile`, `main.go`, `.terraform-version`, `.env.template`, and any `docker-compose*.yml` or `docker-compose*.yaml` file (matched on base name, at any repository path; e.g. `docker-compose.yml`, `docker-compose.tls.yml`). The root `main.go` (the provider server entry point), `.terraform-version` (the pinned Terraform binary used by acceptance tests), and `.env.template` (defaults for every test run) are module-level files that affect every build or test invocation, so a diff touching them cannot be narrowed to a subset of packages.

A changed `.go` file whose directory no longer exists on disk — a diff that deletes (part of) a package — SHALL also trigger a full run: the selection cannot reason about deleted code, so it fails safe in the conservative direction and emits all acceptance test packages.

#### Scenario: Change to shared client triggers full suite

- **WHEN** a file under `internal/clients/` is changed
- **THEN** the tool emits all acceptance test packages
- **AND** no phase 1 or phase 2 analysis is performed

#### Scenario: Change to entitycore triggers full suite

- **WHEN** a file under `internal/entitycore/` is changed
- **THEN** the tool emits all acceptance test packages

#### Scenario: Module-level file change triggers full suite

- **WHEN** `go.mod`, `go.sum`, `Makefile`, `main.go`, `.terraform-version`, `.env.template`, a `docker-compose*.yml` or `docker-compose*.yaml` file, a file under `.github/workflows/`, a file under `scripts/targeted-testacc/`, or a file under a shared acceptance-test helper package (`internal/kibana/dashboard/dashboardacctest/`, `internal/kibana/dashboard/panelkit/contracttest/`, `internal/providerfwtest/`) is changed
- **THEN** the tool emits all acceptance test packages

#### Scenario: Tool-only change triggers full suite

- **WHEN** the only changed files are under `scripts/targeted-testacc/`
- **THEN** the tool emits all acceptance test packages
- **AND** no phase 1 or phase 2 analysis is performed

#### Scenario: Deleted package directory triggers full suite

- **WHEN** a changed `.go` file's directory no longer exists on disk (the diff deletes a package)
- **THEN** the tool emits all acceptance test packages
- **AND** no phase 1 or phase 2 analysis is performed

---

### Requirement: Run-all threshold

If the union of phase 1 and phase 2 results exceeds 70% of the total count of acceptance test packages (packages containing at least one `*_test.go` file that declares a `func TestAcc*` function OR invokes `resource.Test`/`resource.ParallelTest`), the tool SHALL emit all acceptance test packages instead of the computed subset.

#### Scenario: Near-total selection collapses to all

- **WHEN** phase 1 and phase 2 together select more than 70% of all acceptance test packages
- **THEN** the tool emits all acceptance test packages

---

### Requirement: Unresolvable diff defaults to full suite

When the tool cannot compute a resolvable diff (e.g. on `main`, in a shallow clone where the merge-base is unreachable, or when the git diff returns no changed files), it SHALL emit all acceptance test packages.

#### Scenario: No changed files produces full suite

- **WHEN** the git diff is empty or returns no changed files
- **THEN** the tool emits all acceptance test packages

### Requirement: Docs-only diff produces no packages

When no changed file matches the force-all prefix or force-all file table and every changed file is outside Go source (`*.go`) and testdata (`*/testdata/*`) content (for example, files under `docs/` or `openspec/`), the tool SHALL emit nothing (zero packages selected) and exit 0. A diff that matches the force-all table SHALL always emit the full package set, even when it contains no Go source or testdata content.

#### Scenario: Only docs files changed produces no packages

- **WHEN** the only changed files are under `docs/` or `openspec/`
- **AND** no changed file matches the force-all prefix or force-all file table
- **THEN** the tool emits nothing (zero packages selected)
- **AND** the tool exits 0

#### Scenario: Force-all file in a docs-only diff still produces the full suite

- **WHEN** the only changed files are `README.md` and `Makefile`
- **THEN** the tool emits all acceptance test packages (the force-all file table takes precedence over the docs-only rule)

---

### Requirement: Shard-aware output

The tool SHALL accept `--total-shards` (default 1) and `--shard-index` (default 0) flags. After computing the full selected package set:

- If `shard_index >= total_shards`, the tool SHALL emit nothing.
- If `|selected_packages| < min_shard_packages` (default 30) and `total_shards > 1`:
  - `shard_index == 0`: emit all selected packages (single-shard run).
  - `shard_index > 0`: emit nothing.
- Otherwise: apply round-robin — emit packages where `(position % total_shards) == shard_index`, where `position` is the 0-based index in the sorted package list.

#### Scenario: Small set uses single shard

- **WHEN** 8 packages are selected, `--total-shards=2 --shard-index=0`
- **THEN** all 8 packages are emitted

#### Scenario: Small set suppresses shard 1

- **WHEN** 8 packages are selected, `--total-shards=2 --shard-index=1`
- **THEN** nothing is emitted

#### Scenario: Large set is split across shards

- **WHEN** 60 packages are selected, `--total-shards=2 --shard-index=0`
- **THEN** 30 packages are emitted (even-indexed positions)

#### Scenario: Out-of-range shard index emits nothing

- **WHEN** `--total-shards=1 --shard-index=1`
- **THEN** nothing is emitted

---

### Requirement: Git diff baseline

The tool SHALL resolve the diff baseline in order:

1. `--base` flag value (if provided).
2. `TARGETED_TESTACC_BASE` environment variable (if set).
3. `git merge-base origin/main HEAD` (if the command succeeds).
4. `HEAD~1` (fallback).

The diff SHALL be computed with the two-dot form (`git diff --name-only <base>..HEAD`) when the baseline was supplied explicitly via the `--base` flag or `TARGETED_TESTACC_BASE` (such a baseline is expected to be a commit already contained in `HEAD`, e.g. the PR base, so the two-dot diff is the correct PR diff and resolves even in a shallow checkout where the merge base is unreachable). For auto-detected baselines the diff SHALL use the three-dot form (`git diff --name-only <base>...HEAD`), i.e. comparing against the merge base of `<base>` and `HEAD`, so that a moving base (e.g. `origin/main` in CI) does not over-select changes that have already been merged.

When the git diff fails for any reason, the tool SHALL fall back to the full acceptance test package set and print a warning to stderr describing the failure and the fallback.

#### Scenario: Explicit base overrides auto-detection

- **WHEN** `--base=HEAD~5` is passed
- **THEN** the tool diffs `<base>..HEAD` (two-dot diff) to determine changed files

#### Scenario: Auto-detected base uses merge-base semantics

- **WHEN** no `--base` flag or `TARGETED_TESTACC_BASE` variable is set
- **AND** `git merge-base origin/main HEAD` resolves
- **THEN** the tool diffs the merge base of `origin/main` and `HEAD` (three-dot diff) to determine changed files

#### Scenario: Diff failure falls back visibly to full suite

- **WHEN** the git diff command fails (e.g. an unresolvable baseline in a shallow checkout)
- **THEN** the tool prints a warning to stderr noting the fallback
- **AND** the tool emits all acceptance test packages

---

### Requirement: Dry-run mode

When `--dry-run` is passed, the tool SHALL print a human-readable summary to stdout describing: the changed files, which phase produced each selected package, the final package list, and the effective shard assignment. It SHALL NOT emit the bare package list format and SHALL exit 0 without running any tests.

#### Scenario: Dry-run shows selection rationale

- **WHEN** `--dry-run` is passed
- **THEN** stdout contains the list of changed files, selected packages with their selection reason, and shard assignment
- **AND** the output is not a bare newline-separated package list

---

### Requirement: make targeted-testacc target

A `targeted-testacc` Make target SHALL exist. It SHALL invoke the tool, passing `ACCTEST_TOTAL_SHARDS` and `ACCTEST_SHARD_INDEX` as `--total-shards` and `--shard-index`. If the tool emits no packages, the target SHALL print a notice and exit 0 without invoking `gotestsum`. If packages are emitted, it SHALL invoke `go tool gotestsum` with the same flags as `make testacc` (format, rerun-fails, package parallelism, test parallelism, count, timeout) and pass the package list via `--packages`. When `TARGETED_PKGS` is set, the target SHALL use its value verbatim as the package list and SHALL NOT invoke the selection tool (no re-sharding; the package list is used as-is).

#### Scenario: No packages selected exits cleanly

- **WHEN** `make targeted-testacc` is run and the tool emits no packages
- **THEN** a notice is printed to stdout
- **AND** `gotestsum` is not invoked
- **AND** make exits 0

#### Scenario: Packages selected runs gotestsum

- **WHEN** `make targeted-testacc` is run and the tool emits packages
- **THEN** `TF_ACC=1 go tool gotestsum` is invoked with the emitted package list

#### Scenario: TARGETED_PKGS bypasses the selection tool

- **WHEN** `make targeted-testacc TARGETED_PKGS="github.com/example/mod/internal/a github.com/example/mod/internal/b"` is run
- **THEN** the selection tool is not invoked
- **AND** `gotestsum` runs with exactly those two packages (no re-sharding)

---

### Requirement: make targeted-testacc-dry-run target

A `targeted-testacc-dry-run` Make target SHALL exist. It SHALL invoke the tool with `--dry-run` and print the selection rationale. It SHALL NOT invoke `gotestsum` or require `TF_ACC`.

#### Scenario: Dry-run target does not require stack

- **WHEN** `make targeted-testacc-dry-run` is run without a running Elastic Stack
- **THEN** the command succeeds and prints the selection plan

---

### Requirement: Tool unit tests

The `scripts/targeted-testacc/` package SHALL include unit tests covering: changed-file-to-package mapping, force-all prefix detection, entity name extraction from source snippets, reverse-dep walk on a synthetic graph, shard-count logic (threshold and round-robin), and the run-all threshold. The pure-logic tests for changed-file mapping, force-all detection, entity extraction, reverse-dep walking, shard logic, and the run-all threshold SHALL not require a live git repository or `go list` invocation. The repository guard tests (e.g. test-only import classification) are exempt from this constraint and MAY invoke `go list` and `go env`.

#### Scenario: Entity name extraction from source

- **WHEN** a source snippet containing `NewResourceBase(entitycore.ComponentKibana, "space")` is parsed
- **THEN** the extracted entity name is `elasticstack_kibana_space`

#### Scenario: Constructor coverage guard

- **WHEN** the exported `New*` constructors of `internal/entitycore` are enumerated
- **THEN** every one is classified either as an entity constructor covered by the extractor or as a non-entity constructor
- **AND** a newly added entitycore constructor that is unclassified fails the unit test

#### Scenario: Shard suppression below threshold

- **WHEN** shard logic is applied to 5 packages with total_shards=2 and shard_index=1
- **THEN** the result is empty

---

### Requirement: Acceptance package enumeration guard

The tool SHALL enumerate acceptance test packages by walking the set of repository roots `internal/` and `provider/`, and every acceptance suite in the repository — not only those under `internal/` — SHALL be reachable through the force-all, run-all, and unresolvable-diff fallback paths, because all three emit exactly the enumerated set. The provider package's live-stack connection suites (`TestElasticsearchAPIKeyConnection`, `TestFleetConfiguration`, `TestKibanaConfiguration`, etc., in `provider/provider_test.go` and `provider/factory_test.go`) are exactly the non-`TestAcc` live-stack case this requirement exists for.

A guard test SHALL independently enumerate every package in the repository — walking from the repository root and excluding `vendor/`, `.git/`, `testdata/` fixture trees (e.g. `analysis/acctestconfigdirlint/testdata/src/...` contains analyzer fixtures that use `resource.Test`), and the tool's own package — that declares a `func TestAcc*` function or invokes `resource.Test`/`resource.ParallelTest` in a `*_test.go` file, using detection that does not share the enumeration's implementation, and SHALL fail when any such package is absent from the tool's acceptance test package enumeration. Because the Run-all threshold, Unresolvable-diff fallback, and Force-all paths all emit exactly the enumerated acceptance test packages, a package missing from the enumeration is unreachable through every selection path; this guard catches acceptance suites that do not follow the `TestAcc` naming convention before they silently stop running. The guard test MAY invoke `go list` and `go env` (it is exempt from the no-git / no-go-list constraint on the pure-logic unit tests).

#### Scenario: Non-TestAcc acceptance package caught by the guard

- **WHEN** a package in the repository invokes `resource.Test` in a `*_test.go` file without declaring any `func TestAcc`
- **AND** that package is not returned by the tool's acceptance test package enumeration
- **THEN** the guard test fails

#### Scenario: Provider live-stack suites are reachable

- **WHEN** the enumerated roots are `internal/` and `provider/`
- **THEN** the provider package appears in the acceptance test package enumeration
- **AND** the force-all, run-all, and unresolvable-diff fallback paths all emit the provider package
