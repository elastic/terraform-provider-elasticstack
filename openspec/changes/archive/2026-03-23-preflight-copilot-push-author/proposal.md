## Why

Push-triggered CI should run when the branch has **no** open pull request (so normal pushes keep running), **or** when **every** commit in the push is authored by the Copilot coding agent—so agent-only pushes still get CI even if an open PR already exists. Pushes that **do** have an open PR **and** include any non-Copilot-authored commit are skipped on `push` (dedupe with the PR workflow).

## What Changes

- Update **Requirement: Preflight gate (REQ-023–REQ-027)** in `ci-build-lint-test` so `push` events set `should_run=true` when **either**:
  - no open pull request exists for the pushed branch in the same repository, **or**
  - every commit in the push payload has author email `198982749+Copilot@users.noreply.github.com`.
- For `push` events where **neither** condition holds (open PR exists **and** at least one commit is not Copilot-authored at that email), set `should_run=false`.
- Keep existing behavior for `pull_request` / `workflow_dispatch` and `ready_for_review`, and the rule that `build`, `lint`, and `test` run only when `should_run=true`.
- **Implementation**: Extend the preflight `actions/github-script` step in `.github/workflows/test.yml` to combine the existing open-PR check with an “all commits Copilot email” check using **logical OR**.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `ci-build-lint-test`: Preflight gate rules for `push` (no open PR **or** all commits Copilot at the specified noreply email).

## Impact

- **CI**: `.github/workflows/test.yml` preflight job logic; optional follow-up scenarios in workflow-related tests or docs if any exist.
- **Consumers**: Pushes on branches **without** an open PR always satisfy the first disjunct and run push CI regardless of author. Pushes on branches **with** an open PR run push CI only if every pushed commit is attributed to the Copilot coding agent at the specified noreply email; otherwise the push workflow skips `build`/`lint`/`test`.
