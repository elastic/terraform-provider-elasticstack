## MODIFIED Requirements

### Requirement: targeted-testacc Go tool exists and is runnable

A Go `main` package SHALL exist at `scripts/targeted-testacc/` within the provider module. It SHALL be invocable via `go run ./scripts/targeted-testacc/...` from the module root without any build step or `go tool` entry. The tool SHALL complete within 10 seconds on the current codebase. The tool SHALL accept a `--verbose` flag (default false) that prints additional selection diagnostics to stderr.

For every successful invocation, stdout SHALL contain one JSON shard plan and no other output. The plan SHALL contain `has_packages`, `selected_packages`, `shards`, and `rationale`. `selected_packages` SHALL be the complete sorted selection, and `shards` SHALL be an ordered array of ordered package arrays that covers every selected package exactly once. `rationale` SHALL identify why the final selection was made and, when individual packages were selected through analysis, the reason for each such package.

#### Scenario: Tool emits a plan from module root

- **GIVEN** the repository can resolve a diff baseline
- **WHEN** `go run ./scripts/targeted-testacc/...` is executed from the module root
- **THEN** the command exits 0 and emits a JSON shard plan to stdout
- **AND** the plan has a sorted `selected_packages` array and an ordered `shards` array
- **AND** the plan includes `rationale`

#### Scenario: Tool runs successfully from module root

- **GIVEN** the repository can resolve a diff baseline
- **WHEN** the tool is invoked from the module root
- **THEN** it exits 0 and emits one JSON shard plan

#### Scenario: Tool runs successfully with all flags

- **GIVEN** `origin/main` is a valid baseline
- **WHEN** the tool is invoked with `--total-shards=2 --base=origin/main --dry-run`
- **THEN** it exits 0 and emits a JSON shard plan with rationale

#### Scenario: Empty selection emits a no-op plan

- **GIVEN** the changed files select no relevant acceptance test packages
- **WHEN** the tool completes selection
- **THEN** the command exits 0 with `has_packages` set to `false`
- **AND** `selected_packages` is empty
- **AND** `shards` is exactly `[[]]`

### Requirement: Shard-aware output

The tool SHALL accept `--total-shards` (default 1) as the maximum shard count and SHALL reject a value less than one. It SHALL NOT accept a shard-index input. After computing the full selected package set:

- If no packages are selected, the tool SHALL emit exactly one empty shard.
- If `|selected_packages| < min_shard_packages` (default 30) and `total_shards > 1`, the tool SHALL emit exactly one shard containing every selected package.
- Otherwise, the tool SHALL emit `min(total_shards, |selected_packages|)` nonempty shards by assigning packages round-robin from the sorted package list.

Every nonempty plan SHALL set `has_packages` to `true`, and every emitted shard in that plan SHALL contain at least one package.

#### Scenario: Small set uses single shard

- **GIVEN** 8 packages are selected
- **WHEN** the tool is invoked with `--total-shards=2`
- **THEN** the plan contains one shard with all 8 packages
- **AND** `has_packages` is `true`

#### Scenario: Large set is split across shards

- **GIVEN** 60 packages are selected
- **WHEN** the tool is invoked with `--total-shards=2`
- **THEN** the plan contains two shards with 30 packages each
- **AND** the first shard contains the even-indexed positions from the sorted package list
- **AND** the second shard contains the odd-indexed positions from the sorted package list

#### Scenario: Small set suppresses shard 1

- **GIVEN** 8 packages are selected with `--total-shards=2`
- **WHEN** the tool creates its plan
- **THEN** the plan contains no second shard assignment

#### Scenario: Out-of-range shard index emits nothing

- **GIVEN** a valid repository diff
- **WHEN** the tool is invoked with a shard-index argument
- **THEN** it exits nonzero
- **AND** it does not emit a successful shard plan

#### Scenario: Requested shard count exceeds package count

- **GIVEN** 30 packages are selected
- **WHEN** the tool is invoked with `--total-shards=31`
- **THEN** the plan contains 30 nonempty shards
- **AND** every selected package appears in exactly one shard

#### Scenario: Invalid shard count fails

- **GIVEN** a valid repository diff
- **WHEN** the tool is invoked with `--total-shards=0`
- **THEN** it exits nonzero
- **AND** it does not emit a successful shard plan

### Requirement: Dry-run mode

When `--dry-run` is passed, the tool SHALL still emit the JSON shard plan to stdout and SHALL exit 0 without running any tests. It SHALL include selection rationale in a machine-readable field of that plan.

#### Scenario: Dry-run shows selection rationale

- **GIVEN** a diff is available for selection
- **WHEN** `--dry-run` is passed
- **THEN** stdout contains a JSON shard plan with the selected packages, shard assignments, and selection rationale
- **AND** stdout does not contain a plaintext package list or human-readable-only report

### Requirement: make targeted-testacc target

A `targeted-testacc` Make target SHALL exist. When `TARGETED_PKGS` is not set, it SHALL invoke the tool with `ACCTEST_TOTAL_SHARDS` as `--total-shards`, read the JSON shard plan, and select the package array at `ACCTEST_SHARD_INDEX`. The `TARGETED_TESTACC_BASE` variable (if set) is exported as `TARGETED_TESTACC_BASE` in the tool's environment, which `ResolveBaseline` treats as an explicit baseline. If the selected package array is empty, the target SHALL print a notice and exit 0 without invoking `gotestsum`. If packages are selected, it SHALL invoke `go tool gotestsum` with the same flags as `make testacc` (format, rerun-fails, package parallelism, test parallelism, count, timeout) and pass the package list via `--packages`. When `TARGETED_PKGS` is set, the target SHALL use its value verbatim as the package list and SHALL NOT invoke the selection tool or apply further sharding.

#### Scenario: No packages selected exits cleanly

- **GIVEN** the selected shard package array is empty
- **WHEN** `make targeted-testacc` runs
- **THEN** a notice is printed to stdout
- **AND** `gotestsum` is not invoked
- **AND** make exits 0

#### Scenario: Packages selected runs gotestsum

- **GIVEN** the selected shard package array contains packages
- **WHEN** `make targeted-testacc` runs
- **THEN** `TF_ACC=1 go tool gotestsum` is invoked with the selected package list

#### Scenario: TARGETED_PKGS bypasses the selection tool

- **GIVEN** `TARGETED_PKGS` is set to `github.com/example/mod/internal/a github.com/example/mod/internal/b`
- **WHEN** `make targeted-testacc` is run
- **THEN** the selection tool is not invoked
- **AND** `gotestsum` runs with exactly those two packages
