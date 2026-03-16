# `pr-auto-approve` — Workflow and Gate Requirements

Workflow implementation: `.github/workflows/auto-approve-pr.yml`
Script implementation: `scripts/auto-approve/`

## Schema

```yaml
on:
  check_suite:
    types: [completed]
  pull_request_target:
    types: [opened, reopened, synchronize, ready_for_review]

permissions:
  contents: read
  pull-requests: write
```

## Requirements

### Shared Workflow Requirements

- **[REQ-001] (Trigger)**: The workflow shall run when a `check_suite` for a pull request is completed.
- **[REQ-002] (Trigger)**: The workflow shall also run on `pull_request_target` events (`opened`, `reopened`, `synchronize`, `ready_for_review`) to reevaluate gates when PR state changes.
- **[REQ-003] (Scope)**: The workflow shall only evaluate open, non-draft pull requests in the same repository.
- **[REQ-004] (Auth/Permissions)**: The workflow shall use `GITHUB_TOKEN` and request only `contents: read` and `pull-requests: write`.
- **[REQ-005] (CategoryRouting)**: The workflow shall evaluate pull requests against named auto-approve categories; a pull request is eligible when it satisfies the selector for at least one category.
- **[REQ-006] (ApprovalAction)**: When a pull request satisfies all requirements for its matched category, the workflow shall submit an approving review (`APPROVE`).
- **[REQ-007] (Idempotency)**: If the workflow actor has already submitted an approval review, the workflow shall not create another approval review.
- **[REQ-008] (NoApprovalOnFailure)**: If no category matches or any matched category gate fails, the workflow shall not approve and shall exit successfully.
- **[REQ-009] (Observability)**: The script shall emit clear, machine-readable category selection results, gate outcomes, and failure reasons.
- **[REQ-010] (SupplyChain)**: Third-party actions in the workflow shall be pinned by commit SHA.

### Category: `copilot`

- **[REQ-011] (CopilotSelector)**: The `copilot` category selector shall match pull requests opened by Copilot identities (`github-copilot[bot]` or `Copilot`).
- **[REQ-012] (CopilotCommitAuthor)**: Every commit in a `copilot` category pull request shall be authored by an allowed Copilot identity (`github-copilot[bot]` or `Copilot`).
- **[REQ-013] (CopilotFileAllowlist)**: Every changed file path in a `copilot` category pull request shall match one of: `*_test.go` or `*.tf`.
- **[REQ-014] (CopilotDiffThreshold)**: The total pull request line edits (`additions + deletions`) for a `copilot` category pull request shall be strictly less than `300`.

### Category: `dependabot`

- **[REQ-015] (DependabotSelector)**: The `dependabot` category selector shall match pull requests opened by `dependabot[bot]`.
- **[REQ-016] (DependabotApprovalPolicy)**: A pull request that matches the `dependabot` category shall be auto-approved when shared workflow gates pass.

### Global Approval Gates

- **[REQ-017] (ChecksGreen)**: All reported checks for the pull request HEAD commit shall be successful before approving any auto-approve category pull request.
- **[REQ-018] (ChecksDefinition)**: A successful checks state means:
  - the commit status rollup state is `success`, and
  - all check runs are completed with conclusions in `{success, neutral, skipped}`.
- **[REQ-019] (ChecksSelfExclusion)**: Checks evaluation shall exclude the currently executing `PR Auto Approve` workflow run from check-run completeness/conclusion gating to avoid self-blocking while the auto-approve job is still in progress.

### Extensibility and Tests

- **[REQ-020] (CategoryExtensibility)**: Category selectors and gates shall be structured so that new auto-approve categories can be added without modifying existing category behavior.
- **[REQ-021] (Testability)**: Category routing and gate evaluation logic shall be unit tested with table-driven tests that cover passing and failing scenarios for each category and threshold boundary, including shared checks-state gates and self-exclusion behavior.

