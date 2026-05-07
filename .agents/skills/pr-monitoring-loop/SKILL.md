---
name: "pr-monitoring-loop"
description: "Monitor GitHub pull requests through a subagent-based loop that watches CI checks, review comments, PR comments, review state, merge conflicts, and branch freshness. Use when a workflow reaches PR monitoring, CI polling, review feedback handling, or asks to keep a PR merge-ready."
license: "MIT"
compatibility: "Requires git, GitHub CLI, and permission to push fixes to the PR branch."
metadata:
  author: openspec
  version: "1.0"
---

Run a reusable PR monitoring loop while keeping the main agent's context small.

**Input**: A PR number or URL, plus any workflow-specific readiness criteria such as required labels, required approving bot, or timeout rules.

`verify-openspec` behavior is opt-in. Do not add the `verify-openspec` label, wait for a verify approval review, or apply any OpenSpec-specific completion rule unless the caller explicitly asks for that behavior.

## High-Level Flow

1. The main agent starts one worker subagent for the PR.
2. The worker polls the PR with `scripts/check-pr-state.py`.
3. The worker continues until there is something actionable:
   - CI check failure
   - PR review comment or unresolved review thread
   - PR conversation comment
   - blocking review state such as `CHANGES_REQUESTED`
   - merge conflict or out-of-date branch
4. If the worker judges the resolution simple, it may implement the fix, commit it, push it, and continue watching.
5. If the worker judges the resolution non-simple, ambiguous, risky, or needing product judgment, it reports the actionable item to the main agent.
6. The main agent launches a fresh worker subagent scoped only to that fix.
7. After the worker commits and pushes the fix, the main agent starts a fresh watch cycle for the new PR head.
8. Repeat until the PR reaches the caller's success criteria or the loop is blocked.

## Main Agent Instructions

When entering PR monitoring:

1. Create the PR first if needed and record its PR number or URL.
2. Launch a write-capable worker subagent with this skill's watcher prompt.
3. Do not poll GitHub directly in the main agent except to recover from a worker failure.
4. If the worker returns an actionable item for delegation:
   - launch a fresh write-capable worker subagent
   - scope it only to the reported CI failure, review feedback, comment, conflict, or branch freshness issue
   - require minimal commits and a push to the PR branch
   - restart the watch after the worker finishes
5. Stop and ask the user when the watcher or worker reports that the next decision needs user input.

## Watcher Prompt

Tell the watcher subagent:

```markdown
Monitor PR <pr> using `.agents/skills/pr-monitoring-loop/scripts/check-pr-state.py <pr>`.

Poll until one of these happens:
- the PR satisfies the provided success criteria
- a CI check fails
- there is an actionable PR comment, review comment, unresolved review thread, or requested-changes review
- the PR has a merge conflict or stale branch state
- the loop is blocked or has timed out

You may fix and push changes yourself when you judge the fix simple, including mechanical lint, formatting, generated artifacts, obvious test expectation updates, small typo fixes, or other low-context changes. After pushing, continue watching the new PR head.

Return work to the main agent when the fix is non-simple, ambiguous, risky, spans multiple concerns, requires product/API judgment, repeats after an attempted fix, or needs user input.

Return a concise result with:
- status: `ready`, `fixed-and-continued`, `delegate`, `blocked`, or `timeout`
- PR URL and head SHA
- actionable item summary
- evidence from `check-pr-state.py`
- fixes you committed and pushed, if any
- recommended worker scope when status is `delegate`
```

## Polling Cadence

- Poll fast required checks about once per minute while they are pending.
- Poll long-running acceptance tests about every five minutes once they are the only active work.
- Poll comments and reviews on every cadence tick using the same script output as CI checks.
- After any push, restart the watch cycle for the new PR head SHA.

## Deterministic PR State

Use:

```bash
.claude/skills/pr-monitoring-loop/scripts/check-pr-state.py <pr>
```

The script returns one JSON payload with:
- `pr`: PR metadata, head SHA, mergeability, review decision, and labels
- `checks`: status check rollup and `gh pr checks` data when available
- `reviews`: submitted reviews
- `issue_comments`: PR conversation comments
- `review_comments`: inline PR review comments
- `review_threads`: review threads, including unresolved and outdated state
- `summary`: actionable counts and derived state

Base decisions on `summary`, then inspect the detailed arrays for evidence.

## Opt-In Verify-OpenSpec Criteria

Only when the caller explicitly requires `verify-openspec` approval:

1. Add `verify-openspec` only after the current PR head has green required CI and all known actionable comments are addressed.
2. Record when the label was most recently applied.
3. Continue watching until a qualifying `APPROVED` review from the verify automation appears after that label time.
4. Do not treat a green verify workflow check as equivalent to the approval review.
5. Do not treat `summary.pr.reviewDecision == "APPROVED"` as sufficient by itself.
6. Stop with `timeout` if the approval does not arrive within the caller's timeout.

## Guardrails

- Keep watcher context self-contained; return concise summaries to the main agent.
- Prefer fresh worker subagents for delegated fixes.
- Never force-push unless the user explicitly requested it.
- Do not resolve review threads unless the current PR state actually addresses them.
- If a simple watcher fix fails or repeats, delegate it to the main agent.
