# `ci-provider-acceptance-tests` — Provider CI Acceptance Test Job

Delta spec for the `test` job in `.github/workflows/provider.yml`.

## ADDED Requirements

### Requirement: compute-packages step gates stack startup

Each matrix test job SHALL include a `compute-packages` step that runs before the fleet image pull, stack startup, and all other expensive steps. The step SHALL set a `has_packages` output (`true` or `false`). All subsequent expensive steps — fleet image pull, stack start, stack readiness wait, API key creation, fleet setup, forced synthetics installation, and the acceptance test run — SHALL be conditioned on `steps.targeted.outputs.has_packages == 'true'`.

The `compute-packages` step SHALL:
- For non-PR events (`github.event_name != 'pull_request'`): set `has_packages=true` and `targeted_pkgs=` (empty string) unconditionally.
- For PR events: run `git fetch origin main --depth=1`, then invoke `go run ./scripts/targeted-testacc/... --total-shards=2 --shard-index=${{ matrix.shard }}`. If the tool emits at least one package, set `has_packages=true` and `targeted_pkgs=<space-separated list>`. If the tool emits nothing, set `has_packages=false`.

#### Scenario: PR with targeted packages — stack starts and targeted tests run

- **WHEN** a PR event triggers the workflow
- **AND** the tool emits packages for this shard
- **THEN** `has_packages=true` is set
- **AND** all downstream steps including stack startup run normally
- **AND** the test step runs `make targeted-testacc` with `TARGETED_PKGS` set to the tool output

#### Scenario: PR with no packages for this shard — stack is skipped

- **WHEN** a PR event triggers the workflow
- **AND** the tool emits nothing for this shard (e.g. shard 1 of a small targeted run)
- **THEN** `has_packages=false` is set
- **AND** the fleet image pull step is skipped
- **AND** the stack start step is skipped
- **AND** the acceptance test step is skipped
- **AND** the job exits 0

#### Scenario: Push to main — stack starts and full suite runs

- **WHEN** a push to `main` triggers the workflow
- **THEN** `has_packages=true` is set unconditionally by the compute-packages step
- **AND** stack startup proceeds normally
- **AND** the test step runs `make testacc` (full suite)

#### Scenario: workflow_dispatch — full suite runs

- **WHEN** a `workflow_dispatch` event triggers the workflow
- **THEN** `has_packages=true` is set unconditionally
- **AND** the test step runs `make testacc`

---

### Requirement: Test step routes between targeted and full suite

The acceptance test step (`make testacc` / `make targeted-testacc`) SHALL be conditioned on `has_packages == 'true'`. When `targeted_pkgs` is non-empty (PR event with packages), the step SHALL run `make targeted-testacc` passing `ACCTEST_TOTAL_SHARDS=2`, `ACCTEST_SHARD_INDEX=${{ matrix.shard }}`, and `TARGETED_PKGS=${{ steps.targeted.outputs.targeted_pkgs }}`. When `targeted_pkgs` is empty (non-PR event), the step SHALL run `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}` (existing full-suite behaviour, unchanged).

#### Scenario: Non-PR test step is identical to pre-change behaviour

- **WHEN** the workflow runs on a push to `main`
- **THEN** the test step invocation is `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}`
- **AND** no `TARGETED_PKGS` variable is set

#### Scenario: PR test step uses targeted packages

- **WHEN** the workflow runs on a PR and `targeted_pkgs` is non-empty
- **THEN** the test step invocation is `make targeted-testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}`

---

### Requirement: Teardown always runs regardless of shard skip

The stack teardown step (`make docker-clean`) SHALL use `if: always()` and SHALL run even when `has_packages=false`. When the stack was never started, `make docker-clean` SHALL be a no-op and SHALL exit 0.

#### Scenario: Teardown is a no-op when stack was not started

- **WHEN** `has_packages=false` and the stack start step was skipped
- **THEN** `make docker-clean` runs
- **AND** exits 0 without error

---

## MODIFIED Requirements

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job and the change-classification job. The acceptance test job SHALL run with a non-fail-fast matrix covering configured stack versions and included version-specific overrides. The configured stack versions SHALL NOT include Elastic Stack versions below `8.0.0`. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. The acceptance test job SHALL execute only when the preflight gate outputs `should_run=true` and the change-classification job reports `provider_changes=true`.

For each matrix entry, the job SHALL free disk space, set up Go and Terraform, and run `make vendor`. It SHALL then run a `compute-packages` step to determine whether this shard has acceptance test packages to run. Fleet image pull, stack startup via Docker Compose, Elasticsearch and Kibana readiness waits, API key creation, fleet setup, and forced synthetics installation SHALL run only when `compute-packages` outputs `has_packages=true`. For PR events with packages, acceptance tests SHALL run via `make targeted-testacc`; for all other events, acceptance tests SHALL run via `make testacc`. Snapshot versions are allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.

The stack-start step SHALL have a step-level timeout so that a hung container image pull fails fast instead of consuming the full job timeout.

#### Scenario: Provider change on PR — targeted tests run on relevant shards

- **GIVEN** a PR matrix entry for a version and shard
- **AND** the targeted tool selects packages for this shard
- **WHEN** the test job executes
- **THEN** `compute-packages` sets `has_packages=true`
- **AND** the stack SHALL be provisioned and readiness waits SHALL pass
- **AND** `make targeted-testacc` SHALL run with the selected package list

#### Scenario: Provider change on PR — empty shard skips stack

- **GIVEN** a PR matrix entry where the tool selects no packages for this shard
- **WHEN** the test job executes
- **THEN** `compute-packages` sets `has_packages=false`
- **AND** fleet pull, stack start, and the test step SHALL all be skipped
- **AND** the job exits 0

#### Scenario: Push to main always runs full suite

- **GIVEN** a push event to the `main` branch
- **WHEN** a matrix test job executes
- **THEN** `compute-packages` sets `has_packages=true` unconditionally
- **AND** `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=<shard>` runs

#### Scenario: OpenSpec-only change skips matrix acceptance

- **GIVEN** a workflow run whose changed files are all under `openspec/`
- **WHEN** the acceptance test job evaluates its execution conditions
- **THEN** the matrix acceptance `test` job SHALL be skipped (via the change-classification gate, unchanged)

#### Scenario: Compose step timeout prevents hung pull

- **GIVEN** Docker Compose is starting the stack for a matrix entry
- **AND** a container image pull or stack startup hangs
- **WHEN** the configured step timeout is reached
- **THEN** the step SHALL fail and the job SHALL exit early

#### Scenario: Matrix excludes 7.x stack versions

- **GIVEN** the matrix version list
- **THEN** no version below `8.0.0` SHALL appear in the default matrix entries
