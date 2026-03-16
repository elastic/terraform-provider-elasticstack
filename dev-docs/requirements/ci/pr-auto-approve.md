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

- **[REQ-001] (Trigger)**: The workflow shall run when a `check_suite` for a pull request is completed.
- **[REQ-002] (Trigger)**: The workflow shall also run on `pull_request_target` events (`opened`, `reopened`, `synchronize`, `ready_for_review`) to reevaluate gates when PR state changes.
- **[REQ-003] (Scope)**: The workflow shall only evaluate open, non-draft pull requests in the same repository.
- **[REQ-004] (Auth/Permissions)**: The workflow shall use `GITHUB_TOKEN` and request only `contents: read` and `pull-requests: write`.
- **[REQ-005] (CommitAuthor)**: Every commit in the candidate pull request shall be authored by an allowed Copilot identity (`github-copilot[bot]` or `Copilot`).
- **[REQ-006] (FileAllowlist)**: Every changed file path in the pull request shall match one of: `*_test.go` or `*.tf`.
- **[REQ-007] (DiffThreshold)**: The total pull request line edits (`additions + deletions`) shall be strictly less than `300`.
- **[REQ-008] (ChecksGreen)**: All reported checks for the pull request HEAD commit shall be successful before approval.
- **[REQ-009] (ChecksDefinition)**: A successful checks state means:
  - the commit status rollup state is `success`, and
  - all check runs are completed with conclusions in `{success, neutral, skipped}`.
- **[REQ-010] (ApprovalAction)**: When all gates pass, the workflow shall submit an approving review (`APPROVE`) on the pull request.
- **[REQ-011] (Idempotency)**: If the workflow actor has already submitted an approval review, the workflow shall not create another approval review.
- **[REQ-012] (NoApprovalOnFailure)**: If any gate fails, the workflow shall not approve and shall exit successfully.
- **[REQ-013] (Observability)**: The script shall emit clear, machine-readable gate outcomes and failure reasons.
- **[REQ-014] (Testability)**: Gate evaluation logic shall be unit tested with table-driven tests that cover passing and failing scenarios for each criterion and threshold boundary.
- **[REQ-015] (SupplyChain)**: Third-party actions in the workflow shall be pinned by commit SHA.

