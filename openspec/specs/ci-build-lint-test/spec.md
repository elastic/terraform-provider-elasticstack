# `build-lint-test` — Workflow Requirements

Workflow implementation: `.github/workflows/test.yml`

## Purpose

Define the main CI workflow: preflight gate, build, lint (including OpenSpec validation), matrix acceptance tests against Elastic Stack versions, diagnostics, teardown, and optional PR auto-approve.

## Schema

```yaml
on:
  push:
    branches: ['**']
    tags-ignore: ['v*']
    paths-ignore: ['README.md', 'CHANGELOG.md']
  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]
    paths-ignore: ['README.md', 'CHANGELOG.md']
  workflow_dispatch: {}

permissions:
  contents: read
```

## Requirements

### Requirement: Workflow identity and triggers (REQ-001–REQ-006)

The workflow name SHALL be `Build/Lint/Test`. The workflow SHALL run on `push` to any branch, excluding tag refs matching `v*` and excluding changes limited to `README.md` and `CHANGELOG.md`. The workflow SHALL run on `pull_request`, excluding changes limited to `README.md` and `CHANGELOG.md`. The workflow SHALL run on `pull_request` events of type `ready_for_review` (in addition to default types `opened`, `synchronize`, `reopened`). The workflow SHALL support manual execution via `workflow_dispatch`.

#### Scenario: Push to feature branch

- GIVEN a push that is not a `v*` tag and not only ignored paths
- WHEN the preflight gate allows execution
- THEN build, lint, and test jobs MAY run per other requirements

### Requirement: Build and lint jobs (REQ-007–REQ-008, REQ-031)

The `build` job SHALL run on `ubuntu-latest`, set up Go from `go.mod`, run `make vendor`, and run `make build-ci`. The `lint` job SHALL run on `ubuntu-latest`, set up Go from `go.mod`, set up Terraform without wrapper mode, install Node.js (24.x), run `npm ci`, run `openspec validate --specs` with telemetry disabled, and run `make check-lint`.

#### Scenario: Lint validates OpenSpec

- GIVEN the lint job runs after dependencies are installed
- WHEN OpenSpec specs are present under `openspec/specs/`
- THEN `openspec validate --specs` SHALL run successfully before Go/terraform lint checks

### Requirement: Acceptance test job structure (REQ-009–REQ-014)

The matrix acceptance test job SHALL depend on successful completion of the `build` job. The acceptance test job SHALL run with a non-fail-fast matrix covering configured stack versions and included version-specific overrides. The acceptance test job SHALL configure required environment variables for Elastic credentials and experimental provider behavior. For each matrix entry, the job SHALL free disk space, set up Go and Terraform, run `make vendor`, start the stack via Docker Compose, and wait for Elasticsearch and Kibana readiness. Fleet setup and forced synthetics installation SHALL run only for configured version subsets. Acceptance tests SHALL run via `make testacc`, with snapshot versions allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.

#### Scenario: Matrix entry runs stack and tests

- GIVEN a matrix version and runner
- WHEN the test job executes
- THEN the stack SHALL be provisioned, readiness waits SHALL pass, and `make testacc` SHALL run with the documented policy for snapshots

### Requirement: Snapshot failure PR notice (REQ-015)

On snapshot acceptance failure in `pull_request` events, the workflow SHALL create or update a PR warning comment through `actions/github-script`.

#### Scenario: Snapshot test failure on PR

- GIVEN a snapshot matrix entry fails during a pull request build
- WHEN the failure handling step runs
- THEN a bot comment SHALL be created or updated on the PR with a defined marker

### Requirement: Failure diagnostics and teardown (REQ-016–REQ-017)

The workflow SHALL emit Docker Compose logs when the job fails or acceptance tests fail. The workflow SHALL always tear down the Docker Compose stack via `make docker-clean`, regardless of prior step outcomes.

#### Scenario: Always tear down

- GIVEN any prior step outcome in the test job
- WHEN the job finishes
- THEN `make docker-clean` SHALL run in an `always()` step

### Requirement: Auto-approve job (REQ-018–REQ-021)

The `auto-approve` job SHALL depend on successful completion of the `test` (matrix acceptance test) job, except on `ready_for_review` events where it SHALL run without that dependency. The `auto-approve` job SHALL only run on `pull_request` events. The `auto-approve` job SHALL execute `go run ./scripts/auto-approve`; approval policy and gate behavior are defined in [`openspec/specs/ci-pr-auto-approve/spec.md`](../ci-pr-auto-approve/spec.md). The `auto-approve` job SHALL request `contents: read` and `pull-requests: write` permissions.

#### Scenario: Auto-approve after green tests

- GIVEN a pull request workflow and successful test job
- WHEN auto-approve runs
- THEN it SHALL invoke `go run ./scripts/auto-approve` with the specified permissions

### Requirement: Supply chain for actions (REQ-022)

Third-party actions in the workflow SHALL be pinned by commit SHA.

#### Scenario: Action references

- GIVEN a third-party action is used in the workflow
- WHEN the workflow YAML is inspected
- THEN the action reference SHALL use a commit SHA

### Requirement: Preflight gate (REQ-023–REQ-027)

The workflow SHALL evaluate whether to execute CI jobs via a dedicated preflight gate job that emits a `should_run` output.

For `push` events, the preflight gate SHALL set `should_run=true` when either:

* No open pull request exists for the pushed branch in the same repository
* All commits in the push event were authored by an allowed bot user: Copilot coding agent (`198982749+Copilot@users.noreply.github.com`) or GitHub Actions (`41898282+github-actions[bot]@users.noreply.github.com`)

For `push` events where **neither** of the above holds, the preflight gate SHALL set `should_run=false`.

For non-`push` events (`pull_request` and `workflow_dispatch`), the preflight gate SHALL set `should_run=true`, except for `pull_request` events of type `ready_for_review` where it SHALL set `should_run=false`.

The `build`, `lint`, and matrix acceptance `test` jobs SHALL only execute when the preflight gate outputs `should_run=true`.

#### Scenario: Push without open PR

- GIVEN a push to a branch with no open PR in the same repository
- WHEN preflight runs
- THEN `should_run` SHALL be `true`

#### Scenario: Push with open PR and all commits by an allowed bot user

- GIVEN a push to a branch that has an open PR from the same repo
- AND every commit in the push event was authored by Copilot coding agent (`198982749+Copilot@users.noreply.github.com`) or GitHub Actions (`41898282+github-actions[bot]@users.noreply.github.com`)
- WHEN preflight runs
- THEN `should_run` SHALL be `true`

#### Scenario: Push with open PR and a commit not by an allowed bot user

- GIVEN a push to a branch that has an open PR from the same repo
- AND at least one commit in the push event was not authored by Copilot coding agent (`198982749+Copilot@users.noreply.github.com`) or GitHub Actions (`41898282+github-actions[bot]@users.noreply.github.com`)
- WHEN preflight runs
- THEN `should_run` SHALL be `false` and downstream jobs SHALL be skipped

### Requirement: Job permissions (REQ-028–REQ-029)

The preflight gate job SHALL request the minimum permissions required to inspect pull requests (`contents: read`, `pull-requests: read`). The acceptance test job SHALL request `contents: read`, `issues: write`, and `pull-requests: write` permissions.

#### Scenario: Preflight permissions

- GIVEN the preflight job definition
- WHEN permissions are evaluated
- THEN they SHALL match the minimum set for listing PRs

### Requirement: Ready-for-review behavior (REQ-030)

On `ready_for_review` `pull_request` events, only the `auto-approve` job SHALL execute; the `build`, `lint`, and `test` jobs SHALL not run.

#### Scenario: Ready for review event

- GIVEN a `pull_request` with action `ready_for_review`
- WHEN the workflow runs
- THEN only auto-approve SHALL be eligible to run (per gate outputs)
