## MODIFIED Requirements

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job, the change-classification job, and acceptance-matrix preparation. The acceptance test job SHALL run with a non-fail-fast matrix covering configured stack versions and included version-specific overrides. The configured stack versions SHALL NOT include Elastic Stack versions below `8.0.0`. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. The acceptance test job SHALL execute only when the change-classification job reports `provider_changes=true`.

For pull-request events, acceptance-matrix preparation SHALL provide the package assignment for every emitted shard before matrix jobs start. A nonempty targeted plan SHALL create only matrix shards that contain packages. A zero-package targeted plan SHALL create exactly one no-op shard per configured stack version; its disk cleanup, dependency installation, stack startup, and acceptance-test execution SHALL be skipped, and its job SHALL succeed. For pull requests with packages, each test job SHALL free disk space, set up Go and Terraform, run `make vendor`, provision the stack, and run `make targeted-testacc` with its prepared package list. Fleet image pull, stack startup via Docker Compose, Elasticsearch and Kibana readiness waits, API key creation, and forced synthetics installation SHALL run only for jobs with packages; forced synthetics installation SHALL additionally remain limited to configured version subsets. Fleet Server host, agent policy, and package policy setup SHALL be provided by the Docker Compose stack start (`make docker-fleet`) and by the acceptance test PreCheck's default agent download source bootstrap, without any additional per-version-gated Fleet setup step.

For push, workflow-dispatch, and merge-group events, the workflow SHALL retain fixed two-shard full-suite execution via `make testacc`. Snapshot versions are allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.

#### Scenario: Provider change on PR — targeted tests run on relevant shards

- **GIVEN** a pull request selects fewer than 30 acceptance test packages
- **WHEN** the acceptance matrix is created
- **THEN** it contains one shard for each configured stack version
- **AND** each test job runs the prepared package list for that shard

#### Scenario: Provider change on PR — empty shard skips stack

- **GIVEN** a pull request has no package assignment for a test job
- **WHEN** the acceptance matrix is created
- **THEN** that job is omitted from a nonempty targeted matrix

#### Scenario: Large pull-request selection uses two populated shards

- **GIVEN** a pull request selects 30 or more acceptance test packages
- **WHEN** the acceptance matrix is created
- **THEN** it contains two populated shards for each configured stack version
- **AND** each test job runs only the prepared package list assigned to its shard

#### Scenario: Pull request with no targeted packages

- **GIVEN** a pull request produces a zero-package selection plan
- **WHEN** the acceptance matrix is created
- **THEN** it contains one no-op shard for each configured stack version
- **AND** each no-op job exits successfully without disk cleanup, dependency setup, stack startup, or acceptance-test execution

#### Scenario: Push to main always runs full suite

- **GIVEN** a push event to the `main` branch
- **WHEN** a matrix test job executes
- **THEN** the fixed two-shard matrix runs `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=<shard>`

#### Scenario: merge_group runs full suite

- **GIVEN** a `merge_group` event
- **WHEN** a matrix test job executes
- **THEN** the fixed two-shard matrix runs `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=<shard>`

#### Scenario: Non-provider change does not run acceptance jobs

- **GIVEN** change classification reports `provider_changes=false`
- **WHEN** the workflow is evaluated
- **THEN** acceptance-matrix preparation and acceptance test jobs are skipped

#### Scenario: OpenSpec-only change skips matrix acceptance

- **GIVEN** an OpenSpec-only change reports `provider_changes=false`
- **WHEN** the workflow is evaluated
- **THEN** acceptance-matrix preparation and acceptance test jobs are skipped

#### Scenario: Compose step timeout prevents hung pull

- **GIVEN** a populated matrix job starts Docker Compose
- **WHEN** stack startup exceeds its configured timeout
- **THEN** the job fails before the full job timeout

#### Scenario: Matrix excludes 7.x stack versions

- **GIVEN** the pinned acceptance version matrix is loaded
- **WHEN** the matrix is evaluated
- **THEN** every configured version is 8.0.0 or later apart from later unreleased snapshot labels

#### Scenario: Provider change runs stack and tests

- **GIVEN** a provider-changing pull request has a nonempty package assignment
- **WHEN** its test job executes
- **THEN** it provisions the stack and runs assigned acceptance packages

#### Scenario: Fleet bootstrap runs uniformly for every matrix entry

- **GIVEN** a populated matrix job starts its stack
- **WHEN** Docker Compose completes startup
- **THEN** the Fleet bootstrap behavior remains available before acceptance tests run

#### Scenario: Matrix version list is loaded from the pinned artifact

- **GIVEN** the checked-out commit contains the pinned version artifact
- **WHEN** preparation loads versions
- **THEN** the matrix version dimension equals that artifact's JSON array

## ADDED Requirements

### Requirement: Acceptance-matrix preparation gates expensive work

The workflow SHALL prepare the acceptance-test version matrix and targeted package plan before pull-request acceptance jobs fan out. For pull requests, preparation SHALL fetch the PR base commit and invoke the targeted selector once with `--base` and `--total-shards=2`. If fetching the base commit fails, preparation SHALL invoke the selector without `--base`, allowing the selector to use its conservative baseline fallback. If the selector fails, preparation SHALL fail rather than produce a plan that skips tests.

Preparation SHALL expose the pinned version matrix and version flags, numeric shard indices, package arrays indexed by shard, and whether the plan contains packages. Matrix jobs SHALL use that result to skip disk cleanup, dependency installation, stack setup, and acceptance execution when the plan has no packages.

For non-pull-request events, preparation SHALL expose the fixed shard indices `[0, 1]` and SHALL not use targeted selection to alter full-suite execution.

#### Scenario: Pull request planning succeeds

- **GIVEN** a pull request has provider changes
- **WHEN** a pull-request acceptance plan is prepared
- **THEN** the selector is invoked once before matrix jobs start
- **AND** every emitted shard has a corresponding prepared package array

#### Scenario: Base-commit fetch fails

- **GIVEN** a pull request requires targeted selection
- **WHEN** fetching the PR base commit fails
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
- **THEN** it exposes exactly shard indices 0 and 1
- **AND** test jobs use the existing full-suite execution path

## REMOVED Requirements

### Requirement: compute-packages step gates stack startup

**Reason**: Package selection is prepared once before the pull-request matrix
fans out, so per-job package computation is redundant and creates avoidable
empty jobs.

**Migration**: Use acceptance-matrix preparation outputs to construct populated
pull-request shards and gate no-op jobs.

## MODIFIED Requirements

### Requirement: Test step routes between targeted and full suite

For a pull request with packages, the acceptance test step SHALL pass its prepared package list to the shell via an environment variable named `TARGETED_PKGS` and expand `"$TARGETED_PKGS"` in the run script, never by interpolating the list directly into shell text. It SHALL run `make targeted-testacc TARGETED_PKGS="$TARGETED_PKGS"` without further sharding. For push, workflow-dispatch, and merge-group events, the step SHALL run `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}`.

#### Scenario: PR test step uses targeted packages

- **GIVEN** a pull-request test job has a prepared nonempty package list
- **WHEN** its acceptance-test step runs
- **THEN** it invokes `make targeted-testacc TARGETED_PKGS="$TARGETED_PKGS"`
- **AND** it does not invoke the selector or apply further sharding

#### Scenario: Non-PR test step is identical to pre-change behaviour

- **GIVEN** the workflow runs for a non-pull-request event
- **WHEN** an acceptance-test step runs
- **THEN** it invokes `make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}`

### Requirement: Unit tests run independently of acceptance targeting

The workflow SHALL include a dedicated unit-test job (`go test ./... -skip '^TestAcc'`) that runs on every event where the change-classification job reports `provider_changes=true`. The `-skip '^TestAcc'` filter SHALL exclude acceptance tests, which run in the acceptance matrix with a live stack; several `*ExplicitConnection` acceptance tests assert endpoint environment variables before the framework's `TF_ACC` gate and therefore cannot run in a stackless unit-test job. The unit-test job SHALL NOT be gated on the centrally prepared acceptance-package plan, because targeted selection only covers packages containing acceptance tests (`func TestAcc`), so unit-test-only packages would otherwise never run on PRs.

#### Scenario: PR with docs-only acceptance shards still runs unit tests

- **GIVEN** a pull request produces a zero-package acceptance plan and change classification passes
- **WHEN** the unit-test job runs
- **THEN** the unit-test job still runs `go test ./... -skip '^TestAcc'`
- **AND** its result gates the PR

#### Scenario: Unit-test job not gated on has_packages

- **GIVEN** change classification reports provider changes
- **WHEN** the unit-test job executes
- **THEN** its condition does not depend on the prepared acceptance-package plan

### Requirement: Pre-pull fallback fleet image with retry

Before starting the stack via Docker Compose, the workflow SHALL pre-pull the fleet image for matrix entries that use a Docker Hub fallback image. The pre-pull step SHALL use a timeout per attempt and SHALL retry up to three times with backoff. It SHALL be skipped for entries that use the default `docker.elastic.co` registry and for a no-op pull-request acceptance job.

#### Scenario: Docker Hub fleet image is pre-pulled successfully

- **GIVEN** a matrix entry has a Docker Hub fleet image and a nonempty package assignment
- **WHEN** the pre-pull step executes
- **THEN** the image is pulled with a per-attempt timeout
- **AND** failed attempts are retried up to three times

#### Scenario: Pre-pull is skipped for a no-op job

- **GIVEN** a pull-request matrix job has an empty package assignment
- **WHEN** the job step list is evaluated
- **THEN** the pre-pull step is skipped
- **AND** the stack-start and acceptance-test steps are skipped

#### Scenario: Pre-pull is skipped for docker.elastic.co images

- **GIVEN** a populated matrix entry uses the default Elastic registry
- **WHEN** its steps are evaluated
- **THEN** the pre-pull step is skipped
- **AND** stack startup proceeds normally

#### Scenario: Pre-pull is skipped when the shard has no packages

- **GIVEN** a pull-request job has an empty package assignment
- **WHEN** its steps are evaluated
- **THEN** pre-pull, stack startup, and acceptance testing are skipped

### Requirement: Failure diagnostics and teardown (REQ-016–REQ-017)

The workflow SHALL emit Docker Compose logs when a job that starts a stack fails or its acceptance tests fail. The workflow SHALL tear down a stack that it started via `make docker-clean`, regardless of prior step outcomes. A no-op pull-request job that never checks out the repository or starts a stack SHALL skip diagnostics and teardown.

#### Scenario: Always tear down

- **GIVEN** a test job has started a Docker Compose stack
- **WHEN** the job finishes after any prior step outcome
- **THEN** `make docker-clean` runs in an `always()` step

#### Scenario: Teardown is a no-op when stack was not started

- **GIVEN** a pull-request job has an empty package assignment
- **WHEN** the job finishes
- **THEN** Docker Compose diagnostics and teardown are skipped
