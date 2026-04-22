# `makefile-workflows` — Makefile Requirements

Implementation: [`Makefile`](../../../Makefile)

## Purpose

Define the **observable behavior** of the repository root **Makefile**: what targets accomplish, which inputs contributors and automation may set, and how that automation relates to CI.

GitHub Actions workflows specify **when** CI invokes `make` (see [`ci-build-lint-test`](../ci-build-lint-test/spec.md) and [`ci-copilot-setup-steps`](../ci-copilot-setup-steps/spec.md)). Those specs remain authoritative for job structure, triggers, and runners. This spec defines **what the Makefile targets guarantee**; it does not restate CI orchestration.

## Schema

### User-configurable inputs

These are the primary **Make variables and conventions** intended for override or pass-through (defaults live in the Makefile). Other variables may exist for internal wiring only and are not part of this contract.

| Input | Role |
| ----- | ---- |
| `VERSION` | Terraform local install path segment for `make install` |
| `USE_TLS` | Select TLS vs non-TLS Docker Compose stack |
| `TEST`, `TESTARGS` | Unit test package scope and extra `go test` arguments |
| `ACCTEST_PARALLELISM`, `RERUN_FAILS`, `RERUN_FAILS_MAX_FAILURES`, `TESTARGS` | Acceptance parallelism, gotestsum rerun policy including the rerun failure cap, and extra test arguments (defaults use `?=`) |
| `ACCTEST_TIMEOUT`, `ACCTEST_COUNT` | Acceptance timeout and test count (defaults in Makefile; override via `make VAR=value` as for other Make variables) |
| `ELASTICSEARCH_USERNAME`, `ELASTICSEARCH_PASSWORD` | Credentials for local stack helpers and `testacc-vs-docker` |
| `KIBANA_SYSTEM_USERNAME`, `KIBANA_SYSTEM_PASSWORD` | Kibana system user password setup against local Elasticsearch |
| `KIBANA_API_KEY_NAME` | Name for API keys created by helper targets |
| `FLEET_NAME`, `FLEET_ENDPOINT` | Fleet service hostname and HTTPS URL used in Fleet bootstrap helpers |
| `STACK_VERSION` | Stack label used for version-specific Compose/Fleet behavior (e.g. agent image source, Kibana config file selection) |
| `GOLANGCIFLAGS` | Extra flags for golangci-lint (e.g. from `make lint` vs `make check-lint`) |
| `GPG_FINGERPRINT_SECRET` | Required in environment for signing-oriented release targets |
| `GITHUB_TOKEN` | Required in environment for publish-oriented release targets |

Environment variables consumed by underlying tools (for example Terraform logging during acceptance tests) follow those tools’ documentation unless the Makefile documents a specific pass-through.

### Targets (summary)

- **Help:** `help` (default goal)
- **Dependencies & build:** `vendor`, `build-ci`, `build`, `clean`, `install`
- **Tests:** `workflow-test`, `hook-test`, `test`, `testacc`, `testacc-vs-docker`, `docker-testacc`, `docker-testacc-with-token`
- **Docker & HTTP helpers:** `docker-elasticsearch`, `docker-kibana`, `docker-fleet`, `docker-clean`, `copy-kibana-ca`, `set-kibana-password`, `setup-synthetics`, `create-es-api-key`, `create-es-bearer-token`, `setup-kibana-fleet`
- **Lint, format, docs, OpenSpec:** `tools`, `golangci-lint-custom`, `golangci-lint`, `lint-perf`, `lint`, `check-lint`, `fmt`, `check-fmt`, `docs-generate`, `workflow-generate`, `check-workflows`, `check-docs`, `setup-openspec`, `check-openspec`, `setup`
- **Release & maintenance:** `prep-release`, `release-snapshot`, `release-no-publish`, `release`, `check-sign-release`, `check-publish-release`, `release-notes`, `renovate-post-upgrade`, `notice`
- **Codegen:** `gen`
## Requirements
### Requirement: Default goal and help (REQ-001–REQ-002)

The default goal when no target is given SHALL be `help`. The `help` target SHALL list documented targets and short descriptions for interactive use.

#### Scenario: Invoking make with no target

- GIVEN no target is passed to `make`
- WHEN Make runs the default goal
- THEN the `help` target SHALL run

### Requirement: Local tool installation layout (REQ-003–REQ-005)

The Makefile SHALL install repository-local CLI tools under a predictable directory within the repository so contributors do not rely on a global install. When golangci-lint is missing, the `tools` target SHALL install it by piping the **golangci-lint project’s published `install.sh`** to `sh`, targeting the repository-local tools directory and the Makefile’s pinned golangci-lint version.

#### Scenario: Tools target provides linters

- GIVEN the `tools` target runs successfully
- WHEN a contributor runs golangci-lint via the Makefile
- THEN the linter SHALL resolve to the copy provisioned for this repository (installed via that install script when absent)

### Requirement: Provider install layout (REQ-006–REQ-009)

The `install` target SHALL place the built provider binary in the Terraform **local plugin cache** under `registry.terraform.io/elastic/<provider-name>/<VERSION>/<os_arch>`, where `<provider-name>` matches this provider, `<VERSION>` follows `VERSION`, and `<os_arch>` reflects the host Go toolchain’s OS and architecture.

#### Scenario: Local provider install

- GIVEN a successful `make install` with a chosen `VERSION`
- WHEN installation completes
- THEN the binary SHALL exist under the Terraform local plugins tree for that registry address and version

### Requirement: TLS vs non-TLS Compose stack (REQ-010–REQ-011)

When `USE_TLS` is unset or `0`, Docker-related targets SHALL use the repository’s standard Compose definition for non-TLS. When `USE_TLS` is `1`, they SHALL use the TLS Compose definition. The same toggle SHALL apply consistently to all Compose-driven targets.

#### Scenario: TLS stack

- GIVEN `USE_TLS=1`
- WHEN a Docker Compose target runs
- THEN the TLS-oriented Compose file SHALL be used for that invocation

### Requirement: Elasticsearch, Kibana, and Fleet credentials (REQ-012–REQ-016)

The Makefile SHALL supply defaults for Elasticsearch and Kibana credentials and API key naming so local helpers and `testacc-vs-docker` work out of the box; contributors MAY override them via the variables listed in **User-configurable inputs**. Fleet bootstrap helpers SHALL construct the Fleet server URL from `FLEET_NAME` / `FLEET_ENDPOINT` as documented in the Makefile for host URLs exposed to Kibana.

#### Scenario: testacc against local Docker endpoints

- GIVEN `make testacc-vs-docker`
- WHEN the recipe runs
- THEN acceptance tests SHALL receive localhost Elasticsearch and Kibana endpoints and the configured username/password for authentication

### Requirement: Fleet Server image for older stack versions (REQ-017)

When `STACK_VERSION` matches `7.17.%`, `8.0.%`, or `8.1.%`, the Makefile SHALL set the Fleet agent image to **`elastic/elastic-agent` on Docker Hub** so Compose can pull an image that is not published to `docker.elastic.co` for those lines. For other versions, Compose SHALL use the default image source from the Compose files unless overridden elsewhere.

#### Scenario: Older 7.17 / 8.0 / 8.1 line

- GIVEN `STACK_VERSION` matches `7.17.%`, `8.0.%`, or `8.1.%`
- WHEN Compose runs Fleet
- THEN `FLEET_IMAGE` SHALL resolve to Docker Hub’s `elastic/elastic-agent` so pulls can succeed

### Requirement: Vendor dependencies (REQ-018)

The `vendor` target SHALL download Go module dependencies required by the module root.

#### Scenario: Vendor

- GIVEN `make vendor`
- WHEN the command completes successfully
- THEN Go modules SHALL be fetched for the current `go.mod`

### Requirement: Provider build (REQ-019–REQ-021)

The `build-ci` target SHALL produce the provider executable for the current platform. The `build` target SHALL run lint-oriented checks before that build. The `clean` target SHALL remove the built provider binary from the working tree.

#### Scenario: CI-style build without full lint chain

- GIVEN `make build-ci`
- WHEN the build succeeds
- THEN a provider binary SHALL exist at the repository’s conventional output name

#### Scenario: Full local build

- GIVEN `make build`
- WHEN prerequisites succeed
- THEN lint-oriented steps SHALL complete before the compile step

### Requirement: Unit tests (REQ-022)

The `test` target SHALL run all repository unit-style test suites. It SHALL run Go unit tests for `TEST` with a bounded wall-clock timeout, fixed `-count`, and repository-chosen parallelism; extra arguments MAY be supplied via `TESTARGS`. It SHALL also run workflow source generation tests and hook JavaScript tests so `make test` provides a single entry point for unit-level verification.

#### Scenario: Go unit tests

- GIVEN `make test`
- WHEN the Go unit-test portion runs
- THEN packages under `TEST` SHALL have been executed under the configured timeout, count, and parallelism constraints

#### Scenario: Aggregate unit-style test coverage

- GIVEN `make test`
- WHEN the target completes successfully
- THEN `workflow-test` SHALL have been executed
- AND hook JavaScript tests SHALL have been executed

### Requirement: Acceptance tests (REQ-023–REQ-024)

The `testacc` target SHALL enable Terraform acceptance testing for the module tree, using gotestsum with rerun-of-fails behavior, a configurable rerun max-failures cap, and tunable parallelism, timeout, and count via the acceptance-test variables. It SHALL invoke the repository-wide package scope `./...` and pass verbose Go test output through to the underlying test run. The `testacc-vs-docker` target SHALL run acceptance tests against a local Docker stack on default localhost ports with the configured Elasticsearch credentials.

#### Scenario: Acceptance tests with defaults

- GIVEN `make testacc`
- WHEN the recipe runs
- THEN `TF_ACC` SHALL be set for acceptance mode and tests SHALL run across `./...` with the Makefile’s timeout and parallelism defaults unless overridden
- AND gotestsum reruns SHALL honor both the configured rerun count and the configured max-failures cap

### Requirement: Docker-wrapped acceptance tests (REQ-025–REQ-026)

The `docker-testacc` target SHALL ensure the Fleet-oriented stack is up, then run acceptance tests inside Compose using the acceptance-test profile. The `docker-testacc-with-token` target SHALL obtain an Elasticsearch bearer token, expose it to the test container, and run the token-oriented acceptance profile.

#### Scenario: Acceptance tests in Compose

- GIVEN `make docker-testacc`
- WHEN the target runs
- THEN Fleet SHALL be started first and Compose SHALL run the acceptance-test service with the appropriate profile

### Requirement: Docker stack services (REQ-027–REQ-029)

The `docker-elasticsearch`, `docker-kibana`, and `docker-fleet` targets SHALL start the corresponding Compose services in the background. For **`STACK_VERSION=9.4.0-SNAPSHOT` only**, `docker-fleet` SHALL set the Kibana config file for Compose to **`kibana-9.4.snapshot.yml`**; for all other values of `STACK_VERSION`, it SHALL use **`kibana.yml`**.

#### Scenario: Fleet with 9.4.0-SNAPSHOT Kibana config

- GIVEN `STACK_VERSION` is exactly `9.4.0-SNAPSHOT`
- WHEN `make docker-fleet` runs
- THEN the environment passed to Compose SHALL select `kibana-9.4.snapshot.yml` for Kibana

#### Scenario: Fleet with default Kibana config

- GIVEN `STACK_VERSION` is unset or not `9.4.0-SNAPSHOT`
- WHEN `make docker-fleet` runs
- THEN the environment passed to Compose SHALL select `kibana.yml` for Kibana

### Requirement: Local stack HTTP helpers — Elasticsearch (REQ-030–REQ-033)

The `set-kibana-password` target SHALL set the configured Kibana system user’s password against Elasticsearch on localhost. The `create-es-api-key` target SHALL create an API key with the configured name. The `create-es-bearer-token` target SHALL obtain a client-credentials OAuth2 token from Elasticsearch. Calls SHALL use the configured Elasticsearch credentials and JSON APIs on localhost.

#### Scenario: Set kibana_system password locally

- GIVEN Elasticsearch is listening on localhost
- WHEN `make set-kibana-password` runs
- THEN the configured system user’s password SHALL be updated via the security API

### Requirement: Local stack HTTP helpers — Kibana Fleet and synthetics (REQ-034–REQ-035)

The `setup-synthetics` target SHALL ensure the Synthetics Fleet integration package version required by the repository is installed via Kibana on localhost. The `setup-kibana-fleet` target SHALL create the default Fleet server host, fleet-server agent policy, and Fleet Server integration policy expected for local acceptance testing, using the configured Fleet endpoint in the host registration.

#### Scenario: Fleet bootstrap API sequence

- GIVEN `make setup-kibana-fleet` and Kibana reachable on localhost
- WHEN the target runs
- THEN Fleet SHALL be bootstrapped with the default host and policies in the order required for downstream tests

### Requirement: Docker teardown (REQ-036)

The `docker-clean` target SHALL tear down Compose resources associated with the acceptance-test profile and remove their volumes so CI and local runs do not leak state. This SHALL align with how [`ci-build-lint-test`](../ci-build-lint-test/spec.md) invokes teardown.

#### Scenario: CI teardown alignment

- GIVEN workflows invoke `make docker-clean`
- WHEN `docker-clean` runs
- THEN acceptance-test-scoped containers and volumes SHALL be removed

### Requirement: Copy Kibana CA (REQ-037)

The `copy-kibana-ca` target SHALL copy the Kibana TLS CA certificate from the running Kibana container into the workspace as `kibana-ca.pem` for local trust configuration.

#### Scenario: Export Kibana CA to workspace

- GIVEN the `kibana` service is running
- WHEN `make copy-kibana-ca` runs
- THEN `kibana-ca.pem` SHALL exist in the working tree with the CA material from the Kibana container

### Requirement: Documentation, workflow, and code generation (REQ-038–REQ-042)

The `docs-generate` target SHALL regenerate Terraform provider website/markdown documentation using **HashiCorp `terraform-plugin-docs`** (`tfplugindocs`) for provider name `terraform-provider-elasticstack`. `docs-generate` SHALL read the Terraform CLI version from the repository root `.terraform-version` file and SHALL pass that exact version to `tfplugindocs` via `--tf-version`, so documentation generation does not depend on whichever Terraform binary happens to be installed locally. The `workflow-generate` target SHALL regenerate the checked-in GitHub workflow artifacts from the repository-authored workflow sources, and it SHALL run only when explicitly requested. Aggregate targets such as `gen`, `lint`, `check-lint`, and `build` SHALL NOT depend on `workflow-generate`. The `workflow-test` target SHALL run the repository tests that cover workflow source generation. The `hook-test` target SHALL run `node --test .agents/hooks/*.test.mjs`. The `check-workflows` target SHALL verify that generated workflow artifacts are up to date without regenerating them. The `gen` target SHALL run documentation generation and `go generate` for the repository.

#### Scenario: Docs generation

- GIVEN `make docs-generate`
- WHEN it succeeds
- THEN `tfplugindocs` SHALL have regenerated provider docs to match the current schema
- AND the Terraform CLI version used for schema extraction SHALL come from `.terraform-version`

### Requirement: golangci-lint execution (REQ-041–REQ-043)

The `tools` target SHALL provision golangci-lint at the **version pinned in the repository**. The `golangci-lint` target SHALL lint Go code across the repository module using `./...`, while still honoring repository-configured golangci-lint exclusions, with zero tolerance for duplicate identical issues unless `GOLANGCIFLAGS` alters behavior. The `lint` target SHALL enable auto-fix behavior where supported; `check-lint` SHALL not depend on that fix mode for golangci-lint.

#### Scenario: Lint without fix

- GIVEN `make check-lint`
- WHEN golangci-lint runs
- THEN it SHALL report issues without the fix-only mode used by `lint`

#### Scenario: Repository-wide Go lint scope

- GIVEN `make golangci-lint`
- WHEN the target invokes golangci-lint
- THEN it SHALL run against `./...`
- AND Go packages outside `internal/` SHALL be part of the lint scope unless excluded by repository golangci-lint configuration

### Requirement: Lint aggregate targets (REQ-044–REQ-045)
The `lint` target SHALL run setup, golangci-lint (with fix), formatting, and documentation generation, and it SHALL NOT invoke workflow generation. The `check-lint` target SHALL run setup, OpenSpec structural validation, golangci-lint (check mode), workflow generation checks, format check, repository code generation via `gen`, and documentation freshness check.

#### Scenario: Lint matches contributor workflow
- GIVEN `make lint`
- WHEN it completes successfully
- THEN formatting, lint with fix, and docs generation SHALL have run after setup

#### Scenario: Check-lint runs workflow generation validation
- **GIVEN** generated workflow sources are out of date with their checked-in templates
- **WHEN** `make check-lint` runs
- **THEN** it SHALL fail before reporting success for repository validation

### Requirement: OpenSpec install and validation (REQ-046–REQ-049)

The `setup-openspec` target SHALL install Node dependencies from the repository’s `package.json` / lockfile so the OpenSpec CLI is available; it SHALL require `npm` to be installed. The `check-openspec` target SHALL validate all canonical specs under `openspec/specs/` and SHALL fail with an actionable message if the CLI is missing. OpenSpec telemetry SHALL be disabled for that validation invocation per project policy.

#### Scenario: OpenSpec validation

- GIVEN `check-openspec` runs after dependencies are installed
- WHEN validation executes
- THEN canonical specs SHALL pass structural validation and telemetry SHALL not be collected for that run

### Requirement: Formatting and format check (REQ-050–REQ-051)

The `fmt` target SHALL format Go sources and Terraform files in the repository. The `check-fmt` target SHALL apply formatting and fail if the working tree remains dirty, indicating uncommitted formatting changes.

#### Scenario: Format check fails on dirty tree

- GIVEN `check-fmt` runs and formatting changes files
- WHEN the cleanliness check runs
- THEN the target SHALL exit with a non-zero status if changes were not committed

### Requirement: Documentation freshness check (REQ-052)

The `check-docs` target SHALL regenerate docs and fail if the `docs/` tree differs from the committed state.

#### Scenario: Docs drift

- GIVEN generated docs differ from committed files under `docs/`
- WHEN `make check-docs` runs
- THEN the target SHALL exit with a non-zero status

### Requirement: Development setup aggregate (REQ-053)

The `setup` target SHALL prepare Go module dependencies, local Go-based tools, and Node/OpenSpec dependencies needed for lint and spec validation.

#### Scenario: One-step dev dependencies

- GIVEN `make setup`
- WHEN it succeeds
- THEN contributors SHALL be able to run lint and OpenSpec checks without ad hoc installs beyond documented prerequisites (Go, Terraform formatter, Node/npm)

### Requirement: NOTICE and Renovate post-upgrade (REQ-054–REQ-056)

The `notice` target SHALL regenerate the `NOTICE` file from module metadata using the repository’s licence-detector template and rule files. The `renovate-post-upgrade` target SHALL refresh modules, regenerate `NOTICE`, and rebuild generated content under `generated/kbapi` via that subdirectory’s Makefile.

#### Scenario: NOTICE regeneration

- GIVEN `make notice`
- WHEN it completes
- THEN `NOTICE` SHALL reflect current dependency licences per repository configuration

### Requirement: Goreleaser release targets (REQ-057–REQ-060)

The `release-snapshot` target SHALL produce a local snapshot release artifact set. The `release-no-publish` target SHALL require signing prerequisites and produce a release without publishing. The `release` target SHALL require signing and publishing prerequisites and produce a full release. Parallelism and skip flags SHALL follow the Makefile’s release recipes.

#### Scenario: Snapshot release

- GIVEN `make release-snapshot`
- WHEN goreleaser completes
- THEN snapshot artifacts SHALL be produced without requiring a publish token

### Requirement: Release precondition checks (REQ-061–REQ-062)

The `check-sign-release` target SHALL fail unless `GPG_FINGERPRINT_SECRET` is set. The `check-publish-release` target SHALL fail unless `GITHUB_TOKEN` is set.

#### Scenario: Missing signing secret

- GIVEN `GPG_FINGERPRINT_SECRET` is unset
- WHEN `make check-sign-release` is evaluated
- THEN Make SHALL report an error and stop

### Requirement: Release notes excerpt (REQ-063)

The `release-notes` target SHALL print the body of the `## [Unreleased]` section of `CHANGELOG.md` through the next version heading, for copy/paste into release processes.

#### Scenario: Unreleased changelog lines

- GIVEN `CHANGELOG.md` contains `## [Unreleased]` and subsequent release sections
- WHEN `make release-notes` runs
- THEN standard output SHALL contain only the Unreleased section body

### Requirement: Consolidated Kibana client codegen (`gen`)

The `gen` target SHALL remain the repository’s single root-Makefile entry point for checked-in code and documentation generation. The root Makefile SHALL NOT define a `generate-clients` target or a separate `generate-slo-client` target.

#### Scenario: Root codegen entry point remains `gen`

- **WHEN** a contributor inspects the root Makefile for repository-wide generation
- **THEN** `gen` SHALL be the available root target for that workflow

#### Scenario: Deprecated generator targets absent

- **WHEN** a contributor inspects the root Makefile for older client-generation entry points
- **THEN** there SHALL be no `generate-clients` phony target
- **AND** there SHALL be no `generate-slo-client` phony target or equivalent recipe that populated `generated/slo`

### Requirement: Custom lint performance measurement target

The Makefile SHALL provide a `lint-perf` target that captures isolated performance data for the repository's custom golangci analyzers without relying on aggregate `make lint` wall time. The target SHALL build or reuse the repository-local custom golangci binary, run `acctestconfigdirlint` against `./...` with fixed single-run concurrency, and write timing plus CPU, memory, and trace artifacts to a repo-local output directory.

#### Scenario: Isolated custom linter profiles

- **GIVEN** a contributor runs `make lint-perf`
- **WHEN** the target invokes the custom golangci binary
- **THEN** `acctestconfigdirlint` SHALL be measured in an isolated run rather than only as part of the full default linter set
- **AND** that run SHALL emit timing/profile artifacts under a repo-local output directory

#### Scenario: Repository-aligned scope and entrypoint

- **GIVEN** `make lint-perf` measures a custom analyzer
- **WHEN** it invokes golangci-lint for that analyzer
- **THEN** it SHALL use the repository's custom golangci binary and the repository-wide package scope `./...`
- **AND** it SHALL keep concurrency fixed so repeated comparisons use a stable execution mode

### Requirement: Custom analyzer benchmark capture

The `lint-perf` target SHALL also run repository-local Go benchmarks for the `acctestconfigdirlint` analyzer package and capture their outputs alongside the isolated golangci-lint measurements. This benchmark capture SHALL use the analyzer package under `analysis/acctestconfigdirlint/...` so future optimizer changes can compare targeted analyzer workloads in addition to the repository-wide isolated run.

#### Scenario: Analyzer benchmark outputs

- **GIVEN** a contributor runs `make lint-perf`
- **WHEN** the measurement target completes successfully
- **THEN** the output directory SHALL contain benchmark output for `analysis/acctestconfigdirlint/...` in addition to the isolated golangci-lint profile artifacts

### Requirement: Release preparation workflow dispatch target
The Makefile SHALL provide a maintainer-facing `prep-release` target that dispatches the release preparation GitHub workflow through `gh workflow run` instead of performing release mutation locally. The target SHALL accept a bump-mode input that supports `patch`, `minor`, and `major`, SHALL default that input to `patch`, and SHALL reject unsupported bump values before invoking `gh`.

#### Scenario: Default bump input dispatches patch release preparation
- **GIVEN** a maintainer runs the release preparation Make target without overriding the bump mode
- **WHEN** the target dispatches the workflow
- **THEN** it SHALL invoke `gh workflow run` using `patch` as the bump input

#### Scenario: Unsupported bump value fails before dispatch
- **GIVEN** a maintainer supplies a bump value other than `patch`, `minor`, or `major`
- **WHEN** the Make target validates its inputs
- **THEN** it SHALL fail before dispatching the workflow

#### Scenario: Make target does not duplicate release logic
- **GIVEN** a maintainer uses the release preparation Make target
- **WHEN** the target runs successfully
- **THEN** it SHALL only dispatch the GitHub workflow rather than editing `Makefile`, editing `CHANGELOG.md`, creating branches, or opening pull requests locally

