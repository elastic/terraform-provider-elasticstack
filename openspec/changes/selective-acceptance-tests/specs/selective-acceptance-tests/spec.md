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

The tool SHALL compute the set of relevant acceptance test packages via two independent phases whose results are unioned and deduplicated.

**Phase 1 — Go reverse-dependency walk:** For each changed Go package, the tool SHALL walk the reverse import graph (non-test imports only) to find all packages that transitively import the changed package. Only packages that contain at least one `func TestAcc` function in a `*_test.go` file SHALL be included.

**Phase 2 — TF entity name grep:** For each changed Go package, the tool SHALL extract Terraform type name suffixes by scanning the package's `.go` files for calls matching:
- `NewResourceBase(entitycore.Component<X>, "<name>")`
- `NewElasticsearchResource[...]("<name>", ...)`
- `NewKibanaResource[...](entitycore.Component<X>, "<name>", ...)`
- `NewKibanaDataSource[...](entitycore.Component<X>, "<name>", ...)`

The tool SHALL construct the full entity name as `elasticstack_<component>_<name>` (e.g. `elasticstack_kibana_space`). It SHALL then grep `internal/` recursively for that string in `*.tf` and `*_test.go` files. The owning package of each matching file SHALL be included in the result set.

#### Scenario: Leaf resource change — direct package selected

- **WHEN** `internal/kibana/slo/resource.go` is the only changed file
- **THEN** `internal/kibana/slo` is selected (phase 1: direct package has TestAcc)
- **AND** packages whose testdata references `elasticstack_kibana_slo` are selected (phase 2)

#### Scenario: Panel sub-package change — parent dashboard package selected via reverse dep

- **WHEN** a file under `internal/kibana/dashboard/panel/lensdashboardapp/` is the only changed file
- **THEN** `internal/kibana/dashboard` is selected (phase 1: dashboard imports lensdashboardapp)
- **AND** `internal/kibana/dashboard/panel/lensdashboardapp` is also selected (direct)

#### Scenario: Shared resource consumed by cross-domain testdata

- **WHEN** `internal/kibana/spaces/resource.go` is the only changed file
- **THEN** packages in `internal/fleet/` whose testdata `.tf` files reference `elasticstack_kibana_space` are selected (phase 2)
- **AND** `internal/kibana/spaces` itself is selected (phase 1)

---

### Requirement: Force-all prefix table

When any changed file path has a prefix matching one of the force-all prefixes, the tool SHALL immediately emit all acceptance test packages (equivalent to a "run all" result) without performing phase 1 or phase 2 analysis.

Force-all prefixes: `provider/`, `internal/acctest/`, `internal/clients/`, `internal/entitycore/`, `generated/`.

#### Scenario: Change to shared client triggers full suite

- **WHEN** a file under `internal/clients/` is changed
- **THEN** the tool emits all acceptance test packages
- **AND** no phase 1 or phase 2 analysis is performed

#### Scenario: Change to entitycore triggers full suite

- **WHEN** a file under `internal/entitycore/` is changed
- **THEN** the tool emits all acceptance test packages

---

### Requirement: Run-all threshold

If the union of phase 1 and phase 2 results exceeds 70% of the total count of acceptance test packages (packages containing at least one `func TestAcc` function), the tool SHALL emit all acceptance test packages instead of the computed subset.

#### Scenario: Near-total selection collapses to all

- **WHEN** phase 1 and phase 2 together select more than 70% of all acceptance test packages
- **THEN** the tool emits all acceptance test packages

---

### Requirement: Empty diff defaults to full suite

When the tool cannot compute a non-empty diff (e.g. on `main`, in a shallow clone where merge-base is unreachable, or when only non-Go/non-testdata files changed), it SHALL emit all acceptance test packages.

#### Scenario: No changed files produces full suite

- **WHEN** the git diff is empty or returns no changed files
- **THEN** the tool emits all acceptance test packages

#### Scenario: Only docs files changed produces no packages

- **WHEN** the only changed files are under `docs/` or `openspec/`
- **THEN** the tool emits nothing (zero packages selected)
- **AND** the tool exits 0

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

#### Scenario: Explicit base overrides auto-detection

- **WHEN** `--base=HEAD~5` is passed
- **THEN** the tool diffs `HEAD~5..HEAD` to determine changed files

---

### Requirement: Dry-run mode

When `--dry-run` is passed, the tool SHALL print a human-readable summary to stdout describing: the changed files, which phase produced each selected package, the final package list, and the effective shard assignment. It SHALL NOT emit the bare package list format and SHALL exit 0 without running any tests.

#### Scenario: Dry-run shows selection rationale

- **WHEN** `--dry-run` is passed
- **THEN** stdout contains the list of changed files, selected packages with their selection reason, and shard assignment
- **AND** the output is not a bare newline-separated package list

---

### Requirement: make targeted-testacc target

A `targeted-testacc` Make target SHALL exist. It SHALL invoke the tool, passing `ACCTEST_TOTAL_SHARDS` and `ACCTEST_SHARD_INDEX` as `--total-shards` and `--shard-index`. If the tool emits no packages, the target SHALL print a notice and exit 0 without invoking `gotestsum`. If packages are emitted, it SHALL invoke `go tool gotestsum` with the same flags as `make testacc` (format, rerun-fails, package parallelism, test parallelism, count, timeout) and pass the package list via `--packages`.

#### Scenario: No packages selected exits cleanly

- **WHEN** `make targeted-testacc` is run and the tool emits no packages
- **THEN** a notice is printed to stdout
- **AND** `gotestsum` is not invoked
- **AND** make exits 0

#### Scenario: Packages selected runs gotestsum

- **WHEN** `make targeted-testacc` is run and the tool emits packages
- **THEN** `TF_ACC=1 go tool gotestsum` is invoked with the emitted package list

---

### Requirement: make targeted-testacc-dry-run target

A `targeted-testacc-dry-run` Make target SHALL exist. It SHALL invoke the tool with `--dry-run` and print the selection rationale. It SHALL NOT invoke `gotestsum` or require `TF_ACC`.

#### Scenario: Dry-run target does not require stack

- **WHEN** `make targeted-testacc-dry-run` is run without a running Elastic Stack
- **THEN** the command succeeds and prints the selection plan

---

### Requirement: Tool unit tests

The `scripts/targeted-testacc/` package SHALL include unit tests covering: changed-file-to-package mapping, force-all prefix detection, entity name extraction from source snippets, reverse-dep walk on a synthetic graph, shard-count logic (threshold and round-robin), and the run-all threshold. Tests SHALL not require a live git repository or `go list` invocation.

#### Scenario: Entity name extraction from source

- **WHEN** a source snippet containing `NewResourceBase(entitycore.ComponentKibana, "space")` is parsed
- **THEN** the extracted entity name is `elasticstack_kibana_space`

#### Scenario: Shard suppression below threshold

- **WHEN** shard logic is applied to 5 packages with total_shards=2 and shard_index=1
- **THEN** the result is empty
