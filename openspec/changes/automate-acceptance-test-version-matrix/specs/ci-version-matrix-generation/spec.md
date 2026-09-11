## Purpose

Define a scheduled workflow and Go engine that compute the desired Elastic Stack acceptance-test version list (latest GA patch per 8.x/9.x minor, plus the current unreleased-minor SNAPSHOT label) and keep it pinned in git via a single standing, auto-managed pull request — so the Provider CI matrix (`ci-build-lint-test`) always has a reproducible, up-to-date source of versions without hand-edited YAML.

## ADDED Requirements

### Requirement: Desired version list is a pure function of public, complete sources

The system SHALL compute the desired stack-version list from exactly two sources: the latest patch tag per `8.x`/`9.x` minor from `elastic/elasticsearch` git tags matching `^v(8|9)\.\d+\.\d+$`, and the current SNAPSHOT label from `https://snapshots.elastic.co/latest/master.json`. The system SHALL NOT use `artifacts-api.elastic.co`, GCS `releases/current/*`, Elastic Cloud `stack/versions`, or Docker registry tag listing as a source for computing the desired list.

#### Scenario: GA list is the latest patch per minor

- **GIVEN** `elastic/elasticsearch` has multiple patch tags for a given `8.x`/`9.x` minor
- **WHEN** the desired list is computed
- **THEN** it SHALL include exactly the highest-patch tag for that minor and SHALL NOT include any older patch of the same minor

#### Scenario: Tags outside 8.x/9.x are excluded

- **GIVEN** `elastic/elasticsearch` tags include `7.x` or `10.x` (or later) releases
- **WHEN** the desired list is computed
- **THEN** it SHALL NOT include any tag whose major version is not `8` or `9`

#### Scenario: SNAPSHOT label comes from the master snapshot endpoint

- **WHEN** the desired list is computed
- **THEN** it SHALL append exactly one SNAPSHOT-labeled entry, using the version label reported by `https://snapshots.elastic.co/latest/master.json` at computation time

### Requirement: Docker image existence probe falls back instead of failing

For each newly computed GA version not already present in the pinned artifact, the system SHALL verify that a corresponding Docker image manifest resolves before including it in the desired list. When the manifest does not yet resolve for a given minor's newly computed patch, the system SHALL retain that minor's previously pinned version in the desired list for this run instead of failing the run or omitting the minor entirely.

#### Scenario: New patch image already published

- **GIVEN** a newly computed GA patch for a minor and its Docker image manifest resolves
- **WHEN** the desired list is computed
- **THEN** the desired list SHALL include the newly computed patch for that minor

#### Scenario: New patch image not yet published

- **GIVEN** a newly computed GA patch for a minor whose Docker image manifest does not yet resolve
- **WHEN** the desired list is computed
- **THEN** the desired list SHALL include that minor's previously pinned version instead of the newly computed patch
- **AND** the run SHALL NOT fail because of this fallback

### Requirement: Desired list is pinned to a checked-in JSON artifact

The system SHALL persist the desired version list as a JSON array of version strings at `.github/versions/acceptance-test-matrix.json`, sorted ascending by release line with any SNAPSHOT-labeled entry last.

#### Scenario: Pinned artifact reflects the last accepted computation

- **GIVEN** the pinned artifact at a given commit SHA
- **WHEN** that SHA is checked out
- **THEN** the acceptance-test version list for that SHA SHALL be exactly the pinned artifact's contents, independent of what the version-matrix generator would compute if run again at a later time

### Requirement: Scheduled and manual execution

The workflow SHALL run on a daily schedule and SHALL support manual execution via `workflow_dispatch`.

#### Scenario: Daily scheduled run

- **WHEN** the daily schedule fires
- **THEN** the workflow SHALL compute the desired list and compare it against the pinned artifact

#### Scenario: Manual dispatch

- **WHEN** a maintainer dispatches the workflow manually
- **THEN** the workflow SHALL run the same computation and comparison as the scheduled run

### Requirement: No-op when the desired list is unchanged

When the computed desired list is identical to the pinned artifact, the system SHALL NOT create a commit, SHALL NOT push to the standing branch, and SHALL NOT modify any existing open pull request from that branch.

#### Scenario: Unchanged list leaves an open red PR alone

- **GIVEN** an existing open pull request from the standing branch whose checks are currently failing
- **AND** the computed desired list on the next scheduled run is identical to the pinned artifact
- **WHEN** the workflow runs
- **THEN** it SHALL NOT touch that pull request's branch, commits, or body

### Requirement: Changed list updates a single standing pull request

When the computed desired list differs from the pinned artifact, the system SHALL commit the rewritten artifact (authored by `github-actions[bot]`) to a stable branch named `acceptance-test-version-matrix`, SHALL push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` (or the repository's equivalent CI-trigger credential) to trigger downstream CI, and SHALL create a pull request from that branch to the default branch when none is open, or update the existing one when one already exists. The system SHALL apply the `no-changelog` label to that pull request. The pull request body SHALL state that merging it makes the default branch's pinned list equal the full computed list (all-or-nothing).

#### Scenario: First proposal creates the PR

- **GIVEN** no open pull request exists from the standing branch
- **WHEN** the computed desired list differs from the pinned artifact
- **THEN** the workflow SHALL push the updated artifact and the CI re-trigger commit to the standing branch
- **AND** SHALL create a pull request to the default branch with the `no-changelog` label

#### Scenario: Subsequent proposal updates the same PR

- **GIVEN** an open pull request already exists from the standing branch
- **WHEN** the computed desired list differs again from the (now-updated) pinned artifact
- **THEN** the workflow SHALL update that same branch and pull request rather than opening a duplicate

#### Scenario: A red new GA version holds the whole delta

- **GIVEN** the computed desired list bumps multiple minors, and one bumped version is expected to fail acceptance tests
- **WHEN** the pull request is proposed
- **THEN** the pull request SHALL propose the full computed list as a single all-or-nothing change, and SHALL NOT silently exclude the failing version so the other bumps can land separately

### Requirement: Engine is a Go module invoked via `go run`

The version-list computation, image-existence probe, and pinned-artifact diff/rewrite logic SHALL be implemented in the Go module `scripts/version-matrix/`. The workflow SHALL invoke it via `go run ./scripts/version-matrix <subcommand>` from a step that has already run `actions/checkout` and `actions/setup-go`, consistent with how `scripts/changelog` and `scripts/auto-approve` are invoked elsewhere in this repository.

#### Scenario: Workflow step invokes the Go engine

- **WHEN** the workflow needs to compute or compare the desired version list
- **THEN** it SHALL invoke `go run ./scripts/version-matrix <subcommand>` rather than an `actions/github-script` JavaScript module
