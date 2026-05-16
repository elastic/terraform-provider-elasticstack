# Dead-code Removal Rotation

## Scope

The `ci-deadcode-removal-rotation` workflow is a scheduled GitHub Agentic Workflow
that proposes small, focused dead-code cleanup PRs. It processes **at most one
candidate per run**.

## Candidate Eligibility

- Only functions reported dead by **both** `go tool deadcode ./...` and
  `go tool deadcode -test ./...` are considered.
- This keeps the first iteration to the highest-confidence category: symbols
  that tests do not keep alive.

## Deterministic Selection

1. Run both deadcode scans.
2. Intersect the candidate sets.
3. Filter out candidates whose last attempt is within the cooldown window
   (default **14 days**).
4. Select exactly one candidate using stable lexicographic ordering
   (`package.Symbol`).

## Reference Classification

- `gopls references` is run on the selected symbol to discover all referring
  files.
- Companion test cleanup is **only** allowed when:
  - All references are confined to exactly one local `*_test.go` file.
  - That file's name does **not** match `acc_*test.go`.

## Acceptance-test Backstop

If companion test cleanup is eligible, the agent must inspect the test file for
`resource.Test` or `resource.ParallelTest`. If either is found, the agent must
abort without making changes and record `invalid_candidate_acceptance_test`.

## Verification

Before opening a PR, the agent must run:

- `make build` with a **10-minute timeout**.
- Unit tests (`go test -v`) for all impacted packages.

If either fails, the agent records the failure and stops without creating a PR.

## Cooldown Memory

Every attempt updates the cooldown memory file
(`memory/ci-deadcode-removal-rotation/memory.json`) with:

- Symbol and package.
- Timestamp.
- Deterministic reason code.
- Small structured context (reference file count, test cleanup eligibility,
  impacted packages).

The memory file prevents rapid reselection of the same symbol.

## Reason Codes

| Code | Meaning |
|------|---------|
| `pr_created` | Verified cleanup; PR opened. |
| `build_failed` | `make build` failed. |
| `tests_failed` | Unit tests for impacted packages failed. |
| `verification_timeout` | Build or tests exceeded the timeout. |
| `invalid_candidate_acceptance_test` | Acceptance-test backstop triggered. |
| `invalid_candidate_references` | References changed unexpectedly post-edit. |
| `agent_abort` | Agent could not safely proceed. |
| `no_candidate_available` | No cooldown-eligible candidate existed. |
| `preactivation_blocked` | Deterministic pre-activation rejected the candidate. |

## Outcome Summaries

The `summarize` command produces a compact markdown report of recent attempts
(default last 30 days). It includes:

- Count of attempts per reason code.
- Sticky packages with the most non-PR outcomes.

The summary is included in every cleanup PR body and logged in the workflow
output.

## Manual Review Expectations

- **Do not auto-merge.** All cleanup PRs require human review.
- Maintainers may close incorrect PRs.
- Operational feedback from closed PRs should guide future workflow refinements.
- The workflow is intentionally conservative; false negatives are accepted in
  exchange for safety.
