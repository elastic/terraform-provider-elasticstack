## MODIFIED Requirements

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job and the change-classification job. The acceptance test job SHALL run with a non-fail-fast matrix covering configured Stack versions and suite assignments. For non-pull-request events, each configured version SHALL retain two general shards; versions whose numeric major/minor is at least 9.4 SHALL additionally have one dedicated dashboard suite entry with its own Stack. Pull-request events SHALL retain the prepared targeted shard plan without an additional dashboard entry. Version-specific environment flags SHALL remain derived from the configured version. The configured stack versions SHALL NOT include Elastic Stack versions below `8.0.0`. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. The acceptance test job SHALL execute only when the change-classification job reports `provider_changes=true`.

Each matrix job SHALL consume the centrally prepared package plan. Disk cleanup, Go and Terraform setup, `make vendor`, Fleet image pull, stack startup via Docker Compose, Elasticsearch and Kibana readiness waits, API key creation, and forced synthetics installation SHALL run only when the prepared plan reports `has_packages=true`; forced synthetics installation SHALL additionally remain limited to configured version subsets. Fleet Server host, agent policy, and package policy setup SHALL be provided by the Docker Compose stack start (`make docker-fleet`) and by the acceptance test PreCheck's default agent download source bootstrap, without any additional per-version-gated Fleet setup step. For PR events with packages, acceptance tests SHALL run via `make targeted-testacc`; for all other events (push, workflow_dispatch, merge_group), acceptance tests SHALL run via `make testacc` with the prepared suite and shard assignment. Snapshot versions are allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.

The stack-start step SHALL have a step-level timeout so that a hung container image pull fails fast instead of consuming the full job timeout.

#### Scenario: Provider change on PR — targeted tests run on relevant shards

- **GIVEN** a PR matrix entry for a version and shard
- **AND** the targeted tool selects packages for this shard
- **WHEN** the test job executes
- **THEN** the prepared plan reports `has_packages=true`
- **AND** the stack SHALL be provisioned and readiness waits SHALL pass
- **AND** `make targeted-testacc` SHALL run with the selected package list

#### Scenario: Provider change on PR — empty shard skips stack

- **GIVEN** a PR matrix entry where the tool selects no packages for this shard
- **WHEN** the test job executes
- **THEN** the prepared plan reports `has_packages=false`
- **AND** fleet pull, stack start, and the test step SHALL all be skipped
- **AND** the job exits 0

#### Scenario: Push to main always runs full suite

- **GIVEN** a push event to the `main` branch
- **WHEN** a matrix test job executes
- **THEN** the prepared plan reports `has_packages=true` unconditionally
- **AND** the prepared full-suite assignment runs, using two general shards and one unsharded dashboard suite for each dashboard-eligible version

#### Scenario: merge_group runs full suite

- **GIVEN** a `merge_group` event (merge queue)
- **WHEN** a matrix test job executes
- **THEN** the prepared plan reports `has_packages=true` unconditionally
- **AND** the prepared full-suite assignment runs, using two general shards and one unsharded dashboard suite for each dashboard-eligible version

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
- **AND** the prepared plan contains packages
- **WHEN** the test job executes
- **THEN** the stack SHALL be provisioned and readiness waits SHALL pass
- **AND** the event-appropriate prepared acceptance execution path SHALL run with the documented policy for snapshots

#### Scenario: Fleet bootstrap runs uniformly for every matrix entry

- **GIVEN** any configured matrix version, including one that is not part of any explicit per-version allowlist
- **WHEN** the test job starts the stack via `make docker-fleet`
- **THEN** a default Fleet Server host, a `fleet-server` agent policy, and a `fleet_server` package policy SHALL exist before `make testacc` runs
- **AND** `make testacc`'s `acctest.PreCheck` SHALL ensure a default agent download source exists
- **AND** no separate per-version-gated Fleet setup step SHALL be required for this coverage

#### Scenario: Matrix version list is loaded from the pinned artifact

- **GIVEN** the checked-out commit's `.github/versions/acceptance-test-matrix.json`
- **WHEN** the acceptance test job's matrix is evaluated
- **THEN** the distinct Stack versions in the prepared matrix SHALL equal exactly the JSON array in that file for that commit, independent of what a version-matrix computation would currently produce

### Requirement: Test step routes between targeted and full suite

For a pull request with packages, the acceptance test step SHALL pass its prepared package list to the shell via an environment variable named `TARGETED_PKGS` and expand `"$TARGETED_PKGS"` in the run script, never by interpolating the list directly into shell text. It SHALL run `make targeted-testacc TARGETED_PKGS="$TARGETED_PKGS"` without further sharding. For push, workflow-dispatch, and merge-group events, the step SHALL run `make testacc` with the suite, total shard count, and shard index prepared for that matrix entry. General assignments SHALL use two shards; the dedicated dashboard assignment SHALL run its complete package selection without further sharding.

#### Scenario: PR test step uses targeted packages

- **GIVEN** a pull-request test job has a prepared nonempty package list
- **WHEN** its acceptance-test step runs
- **THEN** it invokes `make targeted-testacc TARGETED_PKGS="$TARGETED_PKGS"`
- **AND** it does not invoke the selector or apply further sharding

#### Scenario: Non-PR test step is identical to pre-change behaviour

- **GIVEN** the workflow runs for a non-pull-request event
- **WHEN** an acceptance-test step runs
- **THEN** it invokes `make testacc` using its prepared suite, total shard count, and shard index
- **AND** general jobs use two shards while dashboard jobs use one

### Requirement: Acceptance-matrix preparation gates expensive work

The workflow SHALL prepare the acceptance-test version matrix and targeted package plan before pull-request acceptance jobs fan out. For pull requests, preparation SHALL preserve the existing first-parent baseline: when `HEAD^1` is available, it SHALL invoke the targeted selector once with `--base=HEAD^1` and `--total-shards=2`. If `HEAD^1` is unavailable, preparation SHALL invoke the selector without `--base`, allowing the selector to use its conservative baseline fallback. If the selector fails, preparation SHALL fail rather than produce a plan that skips tests.

Preparation SHALL expose the pinned versions, version flags, explicit matrix entries with suite and numeric shard assignment, targeted package arrays indexed by shard, and whether the plan contains packages. Matrix jobs SHALL use that result to skip disk cleanup, dependency installation, stack setup, and acceptance execution when the plan has no packages.

For non-pull-request events, preparation SHALL retain general shard indices `[0, 1]` for every configured version and add one dashboard entry only for versions whose numeric major/minor is at least 9.4, including snapshots. Preparation SHALL NOT use targeted selection to alter full-suite execution.

#### Scenario: Pull request planning succeeds

- **GIVEN** a pull request has provider changes
- **WHEN** a pull-request acceptance plan is prepared
- **THEN** the selector is invoked once before matrix jobs start
- **AND** every emitted shard has a corresponding prepared package array

#### Scenario: Base-commit fetch fails

- **GIVEN** a pull request requires targeted selection
- **WHEN** `HEAD^1` is unavailable
- **THEN** preparation invokes the selector without an explicit base
- **AND** the workflow does not silently skip the acceptance suite

#### Scenario: Selector failure fails preparation

- **GIVEN** a pull request requires targeted selection
- **WHEN** the targeted selector exits non-zero during preparation
- **THEN** acceptance-matrix preparation fails
- **AND** no successful no-op plan is emitted

#### Scenario: Non-pull-request preparation preserves full-suite shards

- **GIVEN** the event is a push, workflow-dispatch, or merge-group event
- **WHEN** preparation runs
- **THEN** it exposes general shard indices 0 and 1 for every configured version
- **AND** it additionally exposes one dashboard entry per dashboard-eligible version
- **AND** test jobs use their prepared full-suite assignments

## ADDED Requirements

### Requirement: Dedicated dashboard full-suite partition

On dashboard-eligible versions, the dashboard suite SHALL contain exactly the measured hotspot packages `internal/kibana/dashboard` and `internal/kibana/dashboard/panel/lensxy`. The two general suites SHALL exclude those packages while retaining other dashboard descendants. Their union SHALL cover the original full-suite package set exactly once. Job names SHALL identify the suite so its elapsed duration can be assessed independently.

#### Scenario: Supported full suite covers every package once
- **GIVEN** a full-suite run for Stack 9.4 or later
- **WHEN** the general and dashboard package assignments are combined
- **THEN** every package in the original full-suite selection appears exactly once
- **AND** the two hotspot packages appear only in the dedicated dashboard job
- **AND** other dashboard descendants and similarly named sibling packages remain in a general shard

### Requirement: Dashboard suite version eligibility

Dedicated dashboard jobs SHALL run only for non-PR full-suite versions whose numeric major/minor is at least 9.4. Snapshot suffixes SHALL NOT change numeric eligibility. Older versions SHALL retain their existing two-shard all-package partition without a dedicated dashboard job.

#### Scenario: Older Stack retains full-suite coverage
- **GIVEN** a full-suite run for Stack 9.3 or an earlier supported version
- **WHEN** its jobs are prepared
- **THEN** it has two general shard jobs and no dedicated dashboard job
- **AND** its package selection matches the existing all-package partition

#### Scenario: Snapshot eligibility follows numeric version
- **GIVEN** Stack versions 9.3.9-SNAPSHOT, 9.4.0-SNAPSHOT, and 9.10.0
- **WHEN** dashboard suite eligibility is evaluated
- **THEN** every listed version except 9.3.9-SNAPSHOT is eligible for a dedicated dashboard job

### Requirement: Local full-suite defaults and assignment validation

Local `make testacc` without explicit suite selection SHALL retain its existing all-package behavior. Explicit suite and shard selection SHALL reject unknown suite names, nonpositive shard counts, negative indices, and indices outside the shard count before running tests.

#### Scenario: Default local full suite remains complete
- **GIVEN** no explicit suite is selected
- **WHEN** a developer runs the local full-suite acceptance target
- **THEN** dashboard and other packages remain eligible for the original shard partition

#### Scenario: Invalid assignment fails before test execution
- **GIVEN** suite or shard input is invalid
- **WHEN** package selection is requested
- **THEN** selection fails with a diagnostic
- **AND** no acceptance tests execute

### Requirement: Dedicated dashboard suite preserves existing CI contracts

Dedicated dashboard suite isolation SHALL NOT change targeted PR package selection, no-op plans, test concurrency, snapshot failure policy, diagnostics, teardown, or Provider Gate aggregation. Dedicated jobs SHALL provision their Stack separately from general jobs.

#### Scenario: Targeted PR avoids dashboard duplication
- **GIVEN** a pull-request plan selects dashboard packages or contains no packages
- **WHEN** acceptance jobs are prepared
- **THEN** the existing targeted shard assignments are preserved
- **AND** no extra dashboard job is added
- **AND** an empty plan retains its existing no-op behavior

#### Scenario: Dedicated dashboard job retains lifecycle and gate handling
- **GIVEN** a dedicated dashboard job is scheduled
- **WHEN** it executes
- **THEN** it provisions a Stack separately from the general jobs
- **AND** it uses the same snapshot failure policy, diagnostics, and teardown behavior
- **AND** its result contributes to the existing acceptance job result observed by Provider Gate
