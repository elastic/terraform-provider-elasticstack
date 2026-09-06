# `ci-build-lint-test` — Workflow Requirements

Delta spec for the `test` job in `.github/workflows/provider.yml`.

## ADDED Requirements

### Requirement: compute-packages step gates stack startup

Each matrix test job SHALL include a `compute-packages` step that runs before the fleet image pull, stack startup, and all other expensive steps. The step SHALL set a `has_packages` output (`true` or `false`). All subsequent expensive steps — fleet image pull, stack start, stack readiness wait, API key creation, forced synthetics installation, and the acceptance test run — SHALL be conditioned on `steps.targeted.outputs.has_packages == 'true'`.

The `compute-packages` step SHALL:

- For non-PR events (`github.event_name != 'pull_request'`, including `push`, `workflow_dispatch`, and `merge_group`): set `has_packages=true` and `targeted_pkgs=` (empty string) unconditionally.
- For PR events: fetch the PR base commit (`github.event.pull_request.base.sha`) into a local ref, then invoke `go run ./scripts/targeted-testacc/... --base="<local-ref>" --total-shards=2 --shard-index=${{ matrix.shard }}`. If the base-commit fetch fails, the step SHALL fall back to re-invoking the tool without `--base`, in which case the tool resolves the baseline itself (merge-base of `origin/main` and `HEAD`, or `HEAD~1` when `origin/main` is unavailable), possibly widening the diff rather than skipping tests. If the tool emits at least one package, set `has_packages=true` and `targeted_pkgs=<space-separated list>`. If the tool emits nothing, set `has_packages=false`.

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

#### Scenario: Base-commit fetch fails — tool falls back to self-resolved baseline

- **WHEN** a PR event triggers the workflow
- **AND** the `git fetch` of the PR base commit fails (e.g. pruned ref or shallow-history limitation)
- **THEN** the tool SHALL be re-invoked without `--base`
- **AND** the tool resolves the diff baseline as the merge-base of `origin/main` and `HEAD`, or `HEAD~1` when `origin/main` is unavailable
- **AND** the job SHALL NOT silently skip the acceptance suite

#### Scenario: Tool invocation fails — job fails rather than skipping the suite

- **WHEN** a PR event triggers the workflow
- **AND** the `go run ./scripts/targeted-testacc/...` invocation in the compute-packages step exits non-zero
- **THEN** the step SHALL emit an error annotation and fail
- **AND** `has_packages` SHALL NOT be set to a skipping value
- **AND** the acceptance test suite SHALL NOT be silently skipped with a green job

#### Scenario: Push to main — stack starts and full suite runs

- **WHEN** a push to `main` triggers the workflow
- **THEN** `has_packages=true` is set unconditionally by the compute-packages step
- **AND** stack startup proceeds normally
- **AND** the test step runs `make testacc` (full suite)

#### Scenario: workflow_dispatch — full suite runs

- **WHEN** a `workflow_dispatch` event triggers the workflow
- **THEN** `has_packages=true` is set unconditionally
- **AND** the test step runs `make testacc`

#### Scenario: merge_group — full suite runs

- **WHEN** a `merge_group` event triggers the workflow
- **THEN** `has_packages=true` is set unconditionally by the compute-packages step
- **AND** stack startup proceeds normally
- **AND** the test step runs `make testacc` (full suite)

---

### Requirement: Test step routes between targeted and full suite

The acceptance test step (`make testacc` / `make targeted-testacc`) SHALL be conditioned on `has_packages == 'true'`. The step SHALL pass the `targeted_pkgs` output to the shell via an `env:` block variable (`TARGETED_PKGS`) and expand `"$TARGETED_PKGS"` in the run script, never by interpolating the output directly into shell text. When `targeted_pkgs` is non-empty (PR event with packages), the step SHALL run `make targeted-testacc TARGETED_PKGS="$TARGETED_PKGS"` with no `ACCTEST_TOTAL_SHARDS`/`ACCTEST_SHARD_INDEX` re-sharding, because the tool already applied `--total-shards`/`--shard-index` during package selection and the per-shard package list is final. When `targeted_pkgs` is empty (non-PR event, including `push`, `workflow_dispatch`, and `merge_group`), the step SHALL run `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}` (existing full-suite behaviour, unchanged).

#### Scenario: Non-PR test step is identical to pre-change behaviour

- **WHEN** the workflow runs on a push to `main`
- **THEN** the acceptance test step invocation is `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}`
- **AND** `TARGETED_PKGS` is empty, so the step takes the full-suite branch
- **AND** `make testacc` continues to execute unit-test-only tests inline, because it has no `-run`/`-skip` filter; the dedicated unit-test job is additive on non-PR events, not a replacement
- **AND** on PR events, unit-test-only packages are covered by the dedicated unit-test job, since targeted selection covers only `func TestAcc` packages

#### Scenario: PR test step uses targeted packages

- **WHEN** the workflow runs on a PR and `targeted_pkgs` is non-empty
- **THEN** the test step invocation is `make targeted-testacc TARGETED_PKGS="$TARGETED_PKGS"`
- **AND** `TARGETED_PKGS` is provided via the step's `env:` block, not interpolated into shell text
- **AND** no `ACCTEST_TOTAL_SHARDS`/`ACCTEST_SHARD_INDEX` re-sharding is performed on the targeted invocation

---

### Requirement: Unit tests run independently of acceptance targeting

The workflow SHALL include a dedicated unit-test job (`go test ./... -skip '^TestAcc'`) that runs on every event where the change-classification job reports `provider_changes=true`. The `-skip '^TestAcc'` filter SHALL exclude acceptance tests, which run in the acceptance matrix with a live stack; several `*ExplicitConnection` acceptance tests assert endpoint environment variables before the framework's `TF_ACC` gate and therefore cannot run in a stackless unit-test job. The unit-test job SHALL NOT be gated on the compute-packages `has_packages` output, because the targeted-testacc selection only covers packages containing acceptance tests (`func TestAcc`), so unit-test-only packages (e.g. `internal/entitycore`, `generated/kbapi`, `internal/clients/*`, `internal/fleet/policyshape`, `internal/asyncutils`, `internal/diagutil`) would otherwise never run on PRs.

#### Scenario: PR with docs-only acceptance shards still runs unit tests

- **WHEN** a PR event triggers the workflow and the provider changes classification passes
- **AND** the compute-packages step sets `has_packages=false` for one or more matrix shards
- **THEN** the unit-test job still runs `go test ./...` and its result gates the PR

#### Scenario: Unit-test job not gated on has_packages

- **WHEN** the unit-test job executes
- **THEN** the job has no `if` condition referencing `steps.targeted.outputs.has_packages`

---

### Requirement: Teardown always runs regardless of shard skip

The stack teardown step (`make docker-clean`) SHALL use `if: always()` and SHALL run even when `has_packages=false`. When the stack was never started, `make docker-clean` SHALL be a no-op and SHALL exit 0.

#### Scenario: Teardown is a no-op when stack was not started

- **WHEN** `has_packages=false` and the stack start step was skipped
- **THEN** `make docker-clean` runs
- **AND** exits 0 without error

---

## MODIFIED Requirements

### Requirement: Workflow identity and triggers (REQ-001–REQ-006)

The workflow name SHALL be `Provider CI`. The workflow SHALL run on `push` to branch `main` and to branches matching `renovate/**`. The workflow SHALL run on `pull_request` events of type `opened`, `synchronize`, and `reopened`. The workflow SHALL support manual execution via `workflow_dispatch`. The workflow SHALL also run on `merge_group` events so that merge-queue runs execute the authoritative full acceptance suite.

#### Scenario: Push to main

- GIVEN a `push` to `main`
- WHEN the change-classification job reports `provider_changes=true`
- THEN build, lint, and test jobs MAY run per other requirements

#### Scenario: Push to a Renovate branch

- GIVEN a `push` to a branch matching `renovate/**`
- WHEN the workflow is dispatched
- THEN the workflow SHALL run so commit check runs exist for branch automerge

#### Scenario: Merge queue event triggers workflow

- GIVEN a `merge_group` event from the pull request merge queue
- WHEN the workflow is triggered
- THEN the workflow SHALL run and execute the full acceptance test suite for the merged result

---

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job and the change-classification job. The acceptance test job SHALL run with a non-fail-fast matrix covering configured stack versions and included version-specific overrides. The configured stack versions SHALL NOT include Elastic Stack versions below `8.0.0`. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. The acceptance test job SHALL execute only when the change-classification job reports `provider_changes=true`.

For each matrix entry, the job SHALL free disk space, set up Go and Terraform, and run `make vendor`. It SHALL then run a `compute-packages` step to determine whether this shard has acceptance test packages to run. Fleet image pull, stack startup via Docker Compose, Elasticsearch and Kibana readiness waits, API key creation, and forced synthetics installation SHALL run only when `compute-packages` outputs `has_packages=true`; forced synthetics installation SHALL additionally remain limited to configured version subsets. Fleet Server host, agent policy, and package policy setup SHALL be provided by the Docker Compose stack start (`make docker-fleet`) and by the acceptance test PreCheck's default agent download source bootstrap, without any additional per-version-gated Fleet setup step. For PR events with packages, acceptance tests SHALL run via `make targeted-testacc`; for all other events (push, workflow_dispatch, merge_group), acceptance tests SHALL run via `make testacc`. Snapshot versions are allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.

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

#### Scenario: merge_group runs full suite

- **GIVEN** a `merge_group` event (merge queue)
- **WHEN** a matrix test job executes
- **THEN** `compute-packages` sets `has_packages=true` unconditionally
- **AND** `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=<shard>` runs

#### Scenario: OpenSpec-only change skips matrix acceptance

- **GIVEN** a workflow run whose changed files are all under `openspec/`
- **WHEN** the acceptance test job evaluates its execution conditions
- **THEN** the matrix acceptance `test` job SHALL be skipped

#### Scenario: Compose step timeout prevents hung pull

- **GIVEN** Docker Compose is starting the stack for a matrix entry
- **AND** a container image pull or stack startup hangs
- **WHEN** the configured step timeout is reached
- **THEN** the step SHALL fail and the job SHALL exit early

#### Scenario: Matrix excludes 7.x stack versions

- **WHEN** the acceptance matrix is evaluated
- **THEN** every configured stack version SHALL be 8.0.0 or higher, except snapshot labels that represent later unreleased stack versions

#### Scenario: Provider change runs stack and tests

- **GIVEN** a matrix version and runner
- **AND** the change-classification job reports `provider_changes=true`
- **WHEN** the test job executes
- **THEN** the stack SHALL be provisioned, readiness waits SHALL pass, and `make testacc` SHALL run with the documented policy for snapshots

#### Scenario: Fleet bootstrap runs uniformly for every matrix entry

- **GIVEN** any configured matrix version, including one that is not part of any explicit per-version allowlist
- **WHEN** the test job starts the stack via `make docker-fleet`
- **THEN** a default Fleet Server host, a `fleet-server` agent policy, and a `fleet_server` package policy SHALL exist before `make testacc` runs
- **AND** `make testacc`'s `acctest.PreCheck` SHALL ensure a default agent download source exists
- **AND** no separate per-version-gated Fleet setup step SHALL be required for this coverage

---

### Requirement: Provider gate job (REQ-034–REQ-036)

The workflow SHALL publish a `gate` job ("Provider Gate") that always reports a final required-check result for the workflow run, evaluating the change-classification result together with the `build`, `lint`, `golangci-lint`, matrix acceptance `test`, and `unit-test` job results.

The `gate` job SHALL succeed when either of the following is true:

* The change-classification job reports `provider_changes=false` and `build`, `lint`, `golangci-lint`, the matrix acceptance `test` job, and the `unit-test` job are all intentionally skipped
* `build`, `lint`, `golangci-lint`, the matrix acceptance `test` job, and the `unit-test` job all complete successfully (regardless of the classify result)

The `gate` job SHALL fail when any of the following is true:

* Any of `build`, `lint`, `golangci-lint`, the matrix acceptance `test` job, or the `unit-test` job reports `failure` or `cancelled`
* The change-classification job reports `provider_changes=true` and at least one of `build`, `lint`, `golangci-lint`, the matrix acceptance `test` job, or the `unit-test` job reports an unexpected `skipped` result
* Any other job-result combination, including an unrecognised classify result (not `true`/`false`) or an unrecognised job result value (not one of `success`, `skipped`, `failure`, `cancelled`)

The `gate` job SHALL provide a stable required-check target that can be used by GitHub branch protection or rulesets instead of the per-version matrix acceptance checks or the individual `build`/`lint`/`golangci-lint`/`unit-test` checks.

#### Scenario: OpenSpec-only pull request

- **GIVEN** a pull request whose changed files are all under `openspec/`
- **WHEN** the workflow reaches the `gate` job
- **THEN** `build`, `lint`, `golangci-lint`, the matrix acceptance `test` job, and the `unit-test` job SHALL be treated as intentionally skipped
- **AND** the `gate` job SHALL succeed

#### Scenario: Provider change with failing acceptance coverage

- **GIVEN** a workflow run with `provider_changes=true`
- **AND** the matrix acceptance `test` job does not complete successfully
- **WHEN** the `gate` job evaluates the workflow state
- **THEN** the `gate` job SHALL fail

#### Scenario: Provider change with an unexpected skip

- **GIVEN** a workflow run with `provider_changes=true`
- **AND** at least one of `build`, `lint`, `golangci-lint`, the matrix acceptance `test` job, or the `unit-test` job reports `skipped`
- **WHEN** the `gate` job evaluates the workflow state
- **THEN** the `gate` job SHALL fail

#### Scenario: Unit-test job failure gates the PR

- **GIVEN** a workflow run with `provider_changes=true`
- **AND** the `unit-test` job reports `failure` while `build`, `lint`, `golangci-lint`, and the matrix acceptance `test` job report `success`
- **WHEN** the `gate` job evaluates the workflow state
- **THEN** the `gate` job SHALL fail
- **AND** auto-approval SHALL NOT fire
