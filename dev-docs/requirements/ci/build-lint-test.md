# `build-lint-test` — Workflow Requirements

Workflow implementation: `.github/workflows/test.yml`

## Schema

```yaml
on:
  push:
    branches: ['**']
    tags-ignore: ['v*']
    paths-ignore: ['README.md', 'CHANGELOG.md']
  pull_request:
    paths-ignore: ['README.md', 'CHANGELOG.md']
  workflow_dispatch: {}

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

permissions:
  contents: read
```

## Requirements

- **[REQ-001] (WorkflowName)**: The workflow name shall be `Build/Lint/Test`.
- **[REQ-002] (PushTrigger)**: The workflow shall run on `push` to any branch, excluding tag refs matching `v*` and excluding changes limited to `README.md` and `CHANGELOG.md`.
- **[REQ-003] (PullRequestTrigger)**: The workflow shall run on `pull_request`, excluding changes limited to `README.md` and `CHANGELOG.md`.
- **[REQ-006] (ManualTrigger)**: The workflow shall support manual execution via `workflow_dispatch`.
- **[REQ-007] (BuildJob)**: The `build` job shall run on `ubuntu-latest`, set up Go from `go.mod`, run `make vendor`, and run `make build-ci`.
- **[REQ-008] (LintJob)**: The `lint` job shall run on `ubuntu-latest`, set up Go from `go.mod`, set up Terraform without wrapper mode, and run `make check-lint`.
- **[REQ-009] (AcceptanceDependency)**: The matrix acceptance test job shall depend on successful completion of the `build` job.
- **[REQ-010] (AcceptanceMatrix)**: The acceptance test job shall run with a non-fail-fast matrix covering configured stack versions and included version-specific overrides.
- **[REQ-011] (AcceptanceEnvironment)**: The acceptance test job shall configure required environment variables for Elastic credentials and experimental provider behavior.
- **[REQ-012] (AcceptanceProvisioning)**: For each matrix entry, the job shall free disk space, set up Go and Terraform, run `make vendor`, start the stack via Docker Compose, and wait for Elasticsearch and Kibana readiness.
- **[REQ-013] (FleetSetupConditions)**: Fleet setup and forced synthetics installation shall run only for configured version subsets.
- **[REQ-014] (AcceptanceExecutionPolicy)**: Acceptance tests shall run via `make testacc`, with snapshot versions allowed to fail (`continue-on-error`) while non-snapshot versions remain blocking.
- **[REQ-015] (SnapshotFailurePRNotice)**: On snapshot acceptance failure in `pull_request` events, the workflow shall create or update a PR warning comment through `actions/github-script`.
- **[REQ-016] (FailureDiagnostics)**: The workflow shall emit Docker Compose logs when the job fails or acceptance tests fail.
- **[REQ-017] (TeardownGuarantee)**: The workflow shall always tear down the Docker Compose stack via `make docker-clean`, regardless of prior step outcomes.
- **[REQ-018] (AutoApproveDependency)**: The `auto-approve` job shall depend on successful completion of the `test` (matrix acceptance test) job.
- **[REQ-019] (AutoApproveScope)**: The `auto-approve` job shall run on both `pull_request` and `push` events.
- **[REQ-020] (AutoApproveDelegation)**: The `auto-approve` job shall execute `go run ./scripts/auto-approve`; approval policy and gate behavior are defined in `dev-docs/requirements/ci/pr-auto-approve.md`.
- **[REQ-021] (AutoApprovePermissions)**: The `auto-approve` job shall request `contents: read` and `pull-requests: write` permissions.
- **[REQ-022] (SupplyChain)**: Third-party actions in the workflow shall be pinned by commit SHA.
- **[REQ-029] (ConcurrencyCancellation)**: The workflow shall define top-level GitHub Actions `concurrency` controls with `group: ${{ github.workflow }}-${{ github.ref }}` and `cancel-in-progress: true` so duplicate runs for the same workflow and ref cancel older in-progress runs.
