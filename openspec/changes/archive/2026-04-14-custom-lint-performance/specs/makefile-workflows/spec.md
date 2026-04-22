## ADDED Requirements

### Requirement: Custom lint performance measurement target

The Makefile SHALL provide a `lint-perf` target that captures isolated performance data for the repository's custom golangci analyzers without relying on aggregate `make lint` wall time. The target SHALL build or reuse the repository-local custom golangci binary, run `esclienthelper` and `acctestconfigdirlint` individually against `./...` with fixed single-run concurrency, and write timing plus CPU, memory, and trace artifacts to a repo-local output directory for each run.

#### Scenario: Isolated custom linter profiles

- **GIVEN** a contributor runs `make lint-perf`
- **WHEN** the target invokes the custom golangci binary
- **THEN** `esclienthelper` and `acctestconfigdirlint` SHALL be measured in isolated runs rather than only as part of the full default linter set
- **AND** each run SHALL emit timing/profile artifacts under a repo-local output directory

#### Scenario: Repository-aligned scope and entrypoint

- **GIVEN** `make lint-perf` measures a custom analyzer
- **WHEN** it invokes golangci-lint for that analyzer
- **THEN** it SHALL use the repository's custom golangci binary and the repository-wide package scope `./...`
- **AND** it SHALL keep concurrency fixed so repeated comparisons use a stable execution mode

### Requirement: Custom analyzer benchmark capture

The `lint-perf` target SHALL also run repository-local Go benchmarks for the custom analyzer packages and capture their outputs alongside the isolated golangci-lint measurements. This benchmark capture SHALL use the analyzer packages under `analysis/` so future optimizer changes can compare targeted analyzer workloads in addition to full-repository isolated runs.

#### Scenario: Analyzer benchmark outputs

- **GIVEN** a contributor runs `make lint-perf`
- **WHEN** the measurement target completes successfully
- **THEN** the output directory SHALL contain benchmark output for the custom analyzer packages in addition to the isolated golangci-lint profile artifacts
