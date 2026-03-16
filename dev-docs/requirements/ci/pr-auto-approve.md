# `pr-auto-approve` — Script Requirements

Script implementation: `scripts/auto-approve/`

## Requirements

### Shared Script Behavior

- **[REQ-001] (Scope)**: The script shall evaluate only open, non-draft pull requests in the same repository.
- **[REQ-002] (CategoryRouting)**: The script shall evaluate pull requests against named auto-approve categories; a pull request is eligible only when it matches at least one category selector.
- **[REQ-003] (ApprovalAction)**: When a pull request satisfies all applicable category and global gates, the script shall submit an approving review (`APPROVE`).
- **[REQ-004] (Idempotency)**: If the acting account has already submitted an approval review, the script shall not create another approval review.
- **[REQ-005] (NoApprovalOnFailure)**: If no category matches or any gate fails, the script shall not approve and shall exit successfully.
- **[REQ-006] (Observability)**: The script shall emit clear, machine-readable category selection results, gate outcomes, and failure reasons.

### Category: `copilot`

- **[REQ-007] (CopilotSelector)**: The `copilot` category selector shall match pull requests opened by Copilot identities (`github-copilot[bot]` or `Copilot`).
- **[REQ-008] (CopilotCommitAuthor)**: Every commit in a `copilot` category pull request shall be authored by an allowed Copilot identity (`github-copilot[bot]` or `Copilot`).
- **[REQ-009] (CopilotFileAllowlist)**: Every changed file path in a `copilot` category pull request shall match one of: `*_test.go` or `*.tf`.
- **[REQ-010] (CopilotDiffThreshold)**: The total pull request line edits (`additions + deletions`) for a `copilot` category pull request shall be strictly less than `300`.

### Category: `dependabot`

- **[REQ-011] (DependabotSelector)**: The `dependabot` category selector shall match pull requests opened by `dependabot[bot]`.
- **[REQ-012] (DependabotApprovalPolicy)**: A pull request that matches the `dependabot` category shall be auto-approved when global approval gates pass.

### Extensibility and Tests

- **[REQ-013] (CategoryExtensibility)**: Category selectors and category-specific gates shall be structured so that new categories can be added without modifying existing category behavior.
- **[REQ-014] (Testability)**: Category routing and gate evaluation logic shall be unit tested with table-driven tests that cover passing and failing scenarios for each category and threshold boundary, including shared checks-state gates and self-exclusion behavior.

