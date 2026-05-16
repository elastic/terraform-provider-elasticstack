# `makefile-workflows` — Makefile Requirements (delta)

Implementation: [`Makefile`](../../../../../Makefile)

## MODIFIED Requirements

### Requirement: Acceptance tests (REQ-023–REQ-024)

The `testacc` target SHALL enable Terraform acceptance testing for the module tree, using gotestsum with rerun-of-fails behavior, a configurable rerun max-failures cap, and tunable in-package and cross-package parallelism, timeout, and count via the acceptance-test variables. It SHALL invoke the repository-wide package scope `./...` by default, and pass verbose Go test output through to the underlying test run. The `testacc-vs-docker` target SHALL run acceptance tests against a local Docker stack on default localhost ports with the configured Elasticsearch credentials.

The `testacc` recipe SHALL support modulo-based package sharding via two independently-overridable Makefile variables: `ACCTEST_TOTAL_SHARDS` (total number of shards) and `ACCTEST_SHARD_INDEX` (zero-based index of the shard to run). When `ACCTEST_TOTAL_SHARDS=1` (the default), `testacc` SHALL run all packages identically to the unsharded behaviour — no packages are excluded. When `ACCTEST_TOTAL_SHARDS > 1`, `testacc` SHALL run only those packages whose zero-based position in the sorted `go list ./...` output satisfies `position % ACCTEST_TOTAL_SHARDS == ACCTEST_SHARD_INDEX`. The sorted package list SHALL be derived from `go list ./...` at recipe invocation time so that any package added to the module is automatically assigned to a shard without requiring a configuration change. The union of all shards (indices 0 through ACCTEST_TOTAL_SHARDS−1) SHALL cover every package in `go list ./...` exactly once.

#### Scenario: Acceptance tests with defaults

- GIVEN `make testacc`
- WHEN the recipe runs
- THEN `TF_ACC` SHALL be set for acceptance mode and tests SHALL run across all packages with the Makefile's timeout and parallelism defaults unless overridden
- AND gotestsum reruns SHALL honor both the configured rerun count and the configured max-failures cap
- AND the underlying `go test` invocation SHALL include both an explicit `-p` value (cross-package parallelism) and an explicit `-parallel` value (in-package `t.Parallel()` cap), each taken from a distinct Makefile variable

#### Scenario: Contributor overrides package parallelism

- GIVEN `make testacc` invoked with the package-parallelism Make variable overridden on the command line
- WHEN the recipe runs
- THEN the underlying `go test` invocation SHALL use the overridden value for `-p` without requiring any change to the recipe itself
- AND the in-package `-parallel` value SHALL remain unchanged unless a separate, dedicated variable is also overridden

#### Scenario: CI runs a specific shard

- GIVEN `make testacc` invoked with `ACCTEST_TOTAL_SHARDS=2` and `ACCTEST_SHARD_INDEX=0`
- WHEN the recipe runs
- THEN only packages at even positions in the sorted `go list ./...` output SHALL be passed to gotestsum
- AND packages at odd positions SHALL not be executed in this invocation

#### Scenario: Shard coverage is complete

- GIVEN `make testacc` run twice with `ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=0` and `ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=1`
- WHEN both runs complete
- THEN the union of packages executed SHALL equal the full output of `go list ./...` with no package appearing in both shards and no package omitted
