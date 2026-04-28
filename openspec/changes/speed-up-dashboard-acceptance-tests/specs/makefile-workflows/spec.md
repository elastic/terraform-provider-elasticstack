# `makefile-workflows` — Makefile Requirements (delta)

Implementation: [`Makefile`](../../../../../Makefile)

## MODIFIED Requirements

### Requirement: Acceptance tests (REQ-023–REQ-024)

The `testacc` target SHALL enable Terraform acceptance testing for the module tree, using gotestsum with rerun-of-fails behavior, a configurable rerun max-failures cap, and tunable in-package and cross-package parallelism, timeout, and count via the acceptance-test variables. It SHALL invoke the repository-wide package scope `./...` and pass verbose Go test output through to the underlying test run. The `testacc-vs-docker` target SHALL run acceptance tests against a local Docker stack on default localhost ports with the configured Elasticsearch credentials.

The `testacc` recipe SHALL set `go test`'s package-level parallelism (the `-p` flag) explicitly via a Makefile-defined variable rather than relying on the Go default of `GOMAXPROCS`. The variable SHALL be overridable by contributors and CI through the standard `make VAR=value` mechanism, and the in-package `t.Parallel()` cap (the `-parallel` flag) SHALL remain a separate, independently-overridable variable.

#### Scenario: Acceptance tests with defaults

- GIVEN `make testacc`
- WHEN the recipe runs
- THEN `TF_ACC` SHALL be set for acceptance mode and tests SHALL run across `./...` with the Makefile’s timeout and parallelism defaults unless overridden
- AND gotestsum reruns SHALL honor both the configured rerun count and the configured max-failures cap
- AND the underlying `go test` invocation SHALL include both an explicit `-p` value (cross-package parallelism) and an explicit `-parallel` value (in-package `t.Parallel()` cap), each taken from a distinct Makefile variable

#### Scenario: Contributor overrides package parallelism

- GIVEN `make testacc` invoked with the package-parallelism Make variable overridden on the command line
- WHEN the recipe runs
- THEN the underlying `go test` invocation SHALL use the overridden value for `-p` without requiring any change to the recipe itself
- AND the in-package `-parallel` value SHALL remain unchanged unless a separate, dedicated variable is also overridden
