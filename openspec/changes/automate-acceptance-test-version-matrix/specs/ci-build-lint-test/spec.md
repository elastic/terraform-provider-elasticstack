## MODIFIED Requirements

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job and the change-classification job. The job's `strategy.matrix.version` list SHALL be loaded at run time from the pinned versions artifact `.github/versions/acceptance-test-matrix.json` (via `fromJson()` from a preceding job's output that reads the checked-out file) rather than hardcoded as a literal YAML list. The matrix acceptance test job SHALL run with a non-fail-fast matrix covering the loaded stack versions crossed with a static `shard: [0, 1]` axis. The configured stack versions SHALL NOT include Elastic Stack versions below `8.0.0`. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. The acceptance test job SHALL execute only when the change-classification job reports `provider_changes=true`.

For each matrix entry, the job SHALL free disk space, set up Go and Terraform, run `make vendor`, start the stack via Docker Compose, and wait for Elasticsearch and Kibana readiness. Fleet Server host, agent policy, and package policy setup SHALL be provided by the Docker Compose stack start (`make docker-fleet`) and by the acceptance test PreCheck's default agent download source bootstrap, without any additional per-version-gated Fleet setup step. Forced synthetics installation SHALL run only for configured version subsets, matched by version-range expression per the "Per-version environment rules match version ranges, not exact patches" requirement. Acceptance tests SHALL run via `make testacc`, with snapshot versions allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.

The stack-start step SHALL have a step-level timeout so that a hung container image pull fails fast instead of consuming the full job timeout.

#### Scenario: Provider change runs stack and tests

- **GIVEN** a matrix version and runner
- **AND** the change-classification job reports `provider_changes=true`
- **WHEN** the test job executes
- **THEN** the stack SHALL be provisioned, readiness waits SHALL pass, and `make testacc` SHALL run with the documented policy for snapshots

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
- **THEN** every configured stack version SHALL be `8.0.0` or higher, except snapshot labels that represent later unreleased stack versions

#### Scenario: Fleet bootstrap runs uniformly for every matrix entry

- **GIVEN** any configured matrix version, including one that is not part of any explicit per-version allowlist
- **WHEN** the test job starts the stack via `make docker-fleet`
- **THEN** a default Fleet Server host, a `fleet-server` agent policy, and a `fleet_server` package policy SHALL exist before `make testacc` runs
- **AND** `make testacc`'s `acctest.PreCheck` SHALL ensure a default agent download source exists
- **AND** no separate per-version-gated Fleet setup step SHALL be required for this coverage

#### Scenario: Matrix version list is loaded from the pinned artifact

- **GIVEN** the checked-out commit's `.github/versions/acceptance-test-matrix.json`
- **WHEN** the acceptance test job's matrix is evaluated
- **THEN** `strategy.matrix.version` SHALL equal exactly the JSON array in that file for that commit, independent of what a version-matrix computation would currently produce

### Requirement: Pre-pull fallback fleet image with retry

Before starting the stack via Docker Compose, the workflow SHALL pre-pull the fleet image for matrix entries whose version falls within the Docker-Hub-fallback version range. The pre-pull step SHALL use a timeout per attempt and SHALL retry up to three times with backoff. This step SHALL be skipped for matrix entries outside that range, which use the default `docker.elastic.co` registry.

#### Scenario: Docker Hub fleet image is pre-pulled successfully

- **GIVEN** a matrix entry whose version falls within the Docker-Hub-fallback version range
- **WHEN** the pre-pull step executes
- **THEN** the image SHALL be pulled with a per-attempt timeout
- **AND** failed attempts SHALL be retried up to three times
- **AND** on success, the subsequent `docker compose up` SHALL use the already-pulled image

#### Scenario: Pre-pull is skipped for docker.elastic.co images

- **GIVEN** a matrix entry whose version falls outside the Docker-Hub-fallback version range
- **WHEN** the test job step list is evaluated
- **THEN** the pre-pull step SHALL be skipped
- **AND** the stack-start step SHALL proceed normally

### Requirement: Change classification gate (REQ-032–REQ-033)

The workflow SHALL evaluate whether the `build`, `lint`, `golangci-lint`, and matrix acceptance `test` jobs are required for the current change set via a dedicated change-classification job (`classify`) that runs unconditionally on every trigger. For `pull_request` events, the classifier SHALL set `provider_changes=false` only when every changed file is non-impacting: exactly `CHANGELOG.md`, or any path under `openspec/`, or any path under `.agents/`, or any path under `.github/` other than `.github/workflows/provider.yml` and other than `.github/versions/acceptance-test-matrix.json`. Any change set containing at least one path outside that non-impacting set, or an empty changed-file list, SHALL set `provider_changes=true`. For non-`pull_request` events (including `push` and `workflow_dispatch`), the classifier SHALL skip file inspection entirely and unconditionally set `provider_changes=true`.

When the change-classification job runs, it SHALL expose its result as a workflow output that downstream jobs can consume when deciding whether those jobs are required.

#### Scenario: OpenSpec-only change set

- **GIVEN** a `pull_request` workflow run whose changed files are all under `openspec/`
- **WHEN** the change-classification job evaluates the diff
- **THEN** it SHALL report `provider_changes=false`

#### Scenario: Provider-impacting change set

- **GIVEN** a `pull_request` workflow run whose changed files include at least one path outside the non-impacting set
- **WHEN** the change-classification job evaluates the diff
- **THEN** it SHALL report `provider_changes=true`

#### Scenario: Non-pull_request event always classifies as provider-impacting

- **GIVEN** a non-`pull_request` event triggering the workflow (including `push` or `workflow_dispatch`)
- **WHEN** the change-classification job runs
- **THEN** it SHALL report `provider_changes=true` without inspecting the changed-file list

#### Scenario: Pinned versions artifact is provider-impacting

- **GIVEN** a `pull_request` workflow run whose only changed file is `.github/versions/acceptance-test-matrix.json`
- **WHEN** the change-classification job evaluates the diff
- **THEN** it SHALL report `provider_changes=true`

### Requirement: Snapshot-to-GA version promotion

When the Elastic Stack release tracked by the acceptance matrix's snapshot-labeled entry
(`<version>-SNAPSHOT`) reaches general availability, the pinned versions artifact SHALL be rewritten
to replace that entry with the released version string rather than adding a separate, additional
entry for the same stack line. This rewrite SHALL be performed by the version-matrix generator
(`ci-version-matrix-generation` capability) as part of its normal desired-list computation, not by a
human hand-editing the workflow YAML. Because per-version step conditions in `provider.yml` match by
version-range expression rather than by exact version string (see "Per-version environment rules
match version ranges, not exact patches"), a promoted entry SHALL continue to receive the same step
coverage it received while labeled as a snapshot without requiring any edit to those conditions. The
promoted entry SHALL no longer match `endsWith(matrix.version, '-SNAPSHOT')` and SHALL therefore be
treated as blocking (`continue-on-error: false`) like every other non-snapshot matrix entry, and
SHALL NOT trigger the snapshot-failure PR warning comment.

#### Scenario: Snapshot entry is promoted to its GA release

- **GIVEN** the pinned versions artifact contains a snapshot-labeled entry `X.Y.0-SNAPSHOT` tracking an
  in-development stack line
- **AND** that stack line reaches general availability as `X.Y.0`
- **WHEN** the version-matrix generator next computes the desired list
- **THEN** the `X.Y.0-SNAPSHOT` entry SHALL be rewritten to `X.Y.0` in the pinned artifact
- **AND** no additional entry SHALL be added for the same `X.Y` stack line

#### Scenario: Promoted entry keeps per-version step coverage

- **GIVEN** a per-version-range step condition (for example, forced synthetics install) that matches a
  snapshot entry's minor via a range expression rather than `endsWith(matrix.version, '-SNAPSHOT')`
- **WHEN** that snapshot entry is promoted to its GA version string
- **THEN** the promoted version string SHALL continue to satisfy that step's range condition without
  any change to the workflow YAML

#### Scenario: Promoted entry becomes blocking

- **GIVEN** a matrix entry that was promoted from a snapshot label to its GA version string
- **WHEN** the acceptance test step (`make testacc`) fails for that entry
- **THEN** `continue-on-error` SHALL NOT apply to that failure
- **AND** the snapshot-failure PR warning comment step SHALL NOT fire for that entry

## ADDED Requirements

### Requirement: Per-version environment rules match version ranges, not exact patches

Per-version environment rules in the acceptance test job — Docker-Hub-fallback fleet image selection, `ubuntu-22.04` runner selection, and forced synthetics install — SHALL be expressed as version-range expressions evaluated against `matrix.version` (for example, `startsWith(matrix.version, '8.0.')`), rather than as a `strategy.matrix.include` list or an `if:` condition keyed to exact patch strings. These rules SHALL NOT be part of the pinned versions artifact.

#### Scenario: Range rule survives an automated patch bump

- **GIVEN** a per-version-range rule matches every patch of a given minor (for example, all `8.14.x` patches trigger forced synthetics install)
- **WHEN** the pinned versions artifact is updated to a newer patch of that same minor
- **THEN** the rule SHALL continue to match the new patch without any edit to the workflow YAML

#### Scenario: No `include:` list is used for per-version overrides

- **WHEN** the acceptance test job's `strategy` block is inspected
- **THEN** it SHALL NOT contain a `matrix.include` list keyed to exact version strings for runner or fleet-image selection
