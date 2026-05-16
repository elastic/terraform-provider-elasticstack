---
name: "pr-monitoring-loop"
description: "Monitor GitHub pull requests through a subagent-based loop that watches CI checks, review comments, PR comments, review state, merge conflicts, and branch freshness. Use when a workflow reaches PR monitoring, CI polling, review feedback handling, or asks to keep a PR merge-ready."
license: "MIT"
compatibility: "Requires git, GitHub CLI, and permission to push fixes to the PR branch."
metadata:
  author: openspec
  version: "2.0"
---

Run a reusable PR monitoring loop while keeping the main agent's context small.

**Input**: A PR number or URL, plus any workflow-specific readiness criteria such as required labels, required approving bot, or timeout rules.

The state file is optional. If you do not pass `--state-file`, the script auto-creates one at `.agents/skills/pr-monitoring-loop/scripts/state/.pr-monitor-<pr>.json` (gitignored) on first run and reuses it on every subsequent call for the same PR. You can call the script directly with just a PR number and "new vs old" detection still works across invocations. Override `--state-file <path>` only when you need an isolated state file (for example, parallel watchers on different branches but the same PR number, or tests). Pass `--no-state` to disable persistence entirely (every comment will be reported as new every poll).

`verify-openspec` behavior is opt-in. Do not add the `verify-openspec` label, wait for a verify approval review, or apply any OpenSpec-specific completion rule unless the caller explicitly asks for that behavior.

## High-Level Flow

1. The main agent starts one delegate subagent for the PR. The delegate simply invokes the script; the script auto-creates and reuses `.agents/skills/pr-monitoring-loop/scripts/state/.pr-monitor-<pr>.json` (gitignored) by default. Pass `--state-file <path>` only if you need isolation. Do NOT put state under `.git/` — this repo uses git worktrees, where `.git` is a file pointing at a worktree-specific git dir, which would fragment state across worktrees watching the same PR.
2. The delegate polls the PR with `scripts/check-pr-state.py --state-file <path>`. The state file persists "seen" IDs and `lastPolledAt` so a fresh subagent can still tell new feedback from old.
3. The delegate continues until something is actionable:
   - CI check failure (commit-pinned)
   - new PR review comment, new conversation comment, or new unresolved review thread (judged by `summary.new*` fields, not totals)
   - blocking review state — `summary.reviews.effectiveDecision == "CHANGES_REQUESTED"` (a later APPROVED supersedes an earlier CHANGES_REQUESTED)
   - merge conflict or out-of-date branch
4. If the delegate judges the resolution simple, it may implement the fix, commit it, push it, perform the thread-resolution protocol below for any addressed threads, and continue watching.
5. If the delegate judges the resolution non-simple, ambiguous, risky, or needing product judgment, it reports the actionable item to the main agent.
6. The main agent launches a fresh delegate subagent scoped only to that fix, **passing the same `--state-file` path** so seen IDs persist.
7. After the delegate commits, pushes, replies, and resolves addressed threads, the main agent starts a fresh watch cycle for the new PR head.
8. Repeat until the PR reaches the caller's success criteria or the loop is blocked.

## Main Agent Instructions

When entering PR monitoring:

1. Create the PR first if needed and record its PR number or URL.
2. State persistence works out of the box — the script auto-uses `.agents/skills/pr-monitoring-loop/scripts/state/.pr-monitor-<pr>.json`, so a fresh subagent invoking the script with only a PR number still sees previously seen IDs. Pass `--state-file <path>` explicitly to every subagent only when you need an isolated state file (e.g., parallel watchers, tests). Do NOT put state under `.git/` — see the worktree note above.
3. Launch a write-capable delegate subagent with this skill's watcher prompt.
4. Do not poll GitHub directly in the main agent except to recover from a delegate failure.
5. If the delegate returns an actionable item for delegation:
   - launch a fresh write-capable delegate subagent
   - scope it only to the reported CI failure, review feedback, comment, conflict, or branch freshness issue
   - require minimal commits, a push to the PR branch, and the thread-resolution protocol for any addressed threads
   - restart the watch after the delegate finishes, again passing the same `--state-file`
6. Stop and ask the user when the watcher or delegate reports that the next decision needs user input.

## Watcher Prompt

Tell the watcher subagent:

```markdown
Monitor PR <pr> using:

  .agents/skills/pr-monitoring-loop/scripts/check-pr-state.py <pr>

(The script auto-creates and reuses a state file at
`.agents/skills/pr-monitoring-loop/scripts/state/.pr-monitor-<pr>.json` so "new since last poll"
detection survives across fresh subagents. Only pass `--state-file <path>` if the main agent
explicitly asked you to use a non-default state file.)

Drive every decision off the script's focused output. In particular:
- New work appears as `comments.newIssueComments`, `comments.newReviewComments`,
  `threads.unresolvedNew`, `threads.unresolvedUpdatedSinceHead`, and
  `reviews.newReviewIds`. Old totals stay under `comments.totalIssueComments` / `comments.totalReviewComments` for reference but MUST NOT
  drive the actionable decision.
- Review state is `reviews.effectiveDecision` (latest review per reviewer; a later APPROVED
  supersedes an earlier CHANGES_REQUESTED).
- verify-openspec state is `verifyOpenspec.requiresOpenspecVerification`. When `true`, apply the `verify-openspec` label. When `false`, do not touch the label.
- CI is `checks` (which prefers commit-pinned data over `gh pr checks` rollup).

Poll until one of these happens:
- the PR satisfies the provided success criteria
- a CI check fails (`failed_checks` in `summary.actionable`)
- there is a new actionable PR comment, review comment, unresolved review thread, or
  CHANGES_REQUESTED review (any of `issue_comments`, `review_comments`, `unresolved_review_threads`,
  `changes_requested` in `summary.actionable`)
- the PR has a merge conflict or stale branch state
- the loop is blocked or has timed out

You may fix and push changes yourself when you judge the fix simple, including mechanical lint,
formatting, generated artifacts, obvious test expectation updates, small typo fixes, or other
low-context changes. After pushing, perform the thread-resolution protocol (see "Thread
resolution" below) for every thread your fix addresses, then continue watching the new PR head.

Return work to the main agent when the fix is non-simple, ambiguous, risky, spans multiple
concerns, requires product/API judgment, repeats after an attempted fix, or needs user input.

Print the entire script output JSON before deciding. In your final result include:
- status: `ready`, `fixed-and-continued`, `delegate`, `blocked`, or `timeout`
- PR URL and head SHA
- actionable item summary
- evidence from `check-pr-state.py` (paste the relevant excerpt, e.g. `actionable`, `checks.failedChecks`, `threadDetails`)
- counts of seen vs new IDs you observed
- fixes you committed and pushed, threads you replied to and resolved (with thread ids)
- recommended delegate scope when status is `delegate`
```

## Cadence enforcement

Prefer `--watch` over hand-rolled sleep loops:

- Active CI window (checks pending, fast iteration): `--watch --interval 60`.
- Acceptance-test-only window (only long-running jobs left): `--watch --interval 300`.
- Always set `--max-duration` so the subagent can't run unbounded; size it to the expected window.

`--watch` exits with code `0` on the first actionable tick (and prints a `{"final": true, "outcome": "actionable", ...}` line), `124` on timeout, and `2` on a transient `gh` failure that survived retries. Each tick is one NDJSON line; the watcher should stream and react to those.

After every push, restart the watch cycle for the new PR head SHA. The state file is updated automatically each tick.

## Deterministic PR state

Use:

```bash
.agents/skills/pr-monitoring-loop/scripts/check-pr-state.py <pr>
```

The script auto-creates a state file at the default path on first run; pass `--state-file <path>` only when you need an isolated state file, or `--no-state` to disable persistence entirely.

When invoked without `--full-payload` (the default) the script returns a focused JSON payload containing only actionable decision data. The raw data arrays (`prChecks`, `commitCheckRuns`, `commitStatuses`, `reviews`, `issue_comments`, `review_comments`, `review_threads`, `issue_events`, `merge_conflicts`) are not included in the default output; use `--full-payload` when you need them.

The focused output contains:
- `pr`: number, url, title, headRefName, headRefOid, mergeable, mergeStateStatus, labels
- `checks`: source, headSha, total, failed, pending, passed, `failedChecks[]` with `{name, url}`, pendingNames
- `comments`: totalIssueComments, totalReviewComments, newIssueCommentIds, newReviewCommentIds, newIssueComments[] with `{id, author, body}`, newReviewComments[] with `{id, author, body}`
- `threads`: unresolved, unresolvedNew, unresolvedUpdatedSinceHead, unresolvedThreadIds, unresolvedNewThreadIds
- `threadDetails`: keyed by thread id, includes path, line, resolved, outdated, comments[] with `{author, body, databaseId}`
- `reviews`: total, newReviewIds, effectiveDecision, latestByReviewer, newReviews[] with `{author, state, id}`
- `verifyOpenspec`: runState, requiresOpenspecVerification
- `merge`: blocked, hasConflicts, conflictFiles, conflictAnalysisAvailable, mergeable, mergeStateStatus
- `actionable`: list of string signals
- `hasActionable`: bool
- `headPushedRecently`: bool

Top-level fields the watcher consumes (there is no separate `summary` dict; the root object is the summary):

- `actionable` (list of strings) and `hasActionable` (bool)
- `checks.{source, total, failed, pending, passed, failedChecks[], pendingNames}` — `failedChecks[]` has `{name, url}` for log lookup; `source` is `commit-pinned` when canonical, `pr-checks` when falling back
- `comments.{totalIssueComments, totalReviewComments, newIssueComments, newReviewComments, newIssueCommentIds, newReviewCommentIds}` — `newIssueComments[]` and `newReviewComments[]` include `{id, author, body}` when there is new content
- `threads.{unresolved, unresolvedNew, unresolvedUpdatedSinceHead, unresolvedThreadIds, unresolvedNewThreadIds}`
- `threadDetails.{<threadId>}.comments[].databaseId` — the REST id needed for `in_reply_to` in the thread-resolution protocol
- `reviews.{total, newReviewIds, latestByReviewer, effectiveDecision, verifyOpenspec}`
- `verifyOpenspec.{runState, requiresOpenspecVerification}` — use `requiresOpenspecVerification` as the single trigger for applying the label
- `merge.{blocked, hasConflicts, conflictFiles, ...}`
- `headPushedRecently` is `true` when the head SHA changed since the last poll recorded in the state file

Run the test suite with:

```bash
python -m pytest .agents/skills/pr-monitoring-loop/scripts/tests
```

## Distinguishing new vs. pending-fix threads

GitHub does not auto-resolve a review thread when you push a fix. The script therefore exposes two complementary numbers:

- `summary.threads.unresolvedNew` — threads not yet recorded in the state file (genuinely new feedback).
- `summary.threads.unresolvedUpdatedSinceHead` — threads where a comment was posted after the recorded `lastPolledAt` (typically a follow-up reply on a thread you already saw).

`summary.actionable` lists `unresolved_review_threads` only when one of these is non-zero. A thread the delegate addressed and resolved (see "Thread resolution") drops out of `unresolved` entirely; a thread the delegate addressed but did NOT resolve will keep firing, which is the bug we are explicitly avoiding.

## Thread resolution

When a delegate pushes a fix that addresses a review thread, it MUST do BOTH of the following, in order:

1. Post a reply on the thread citing the addressing commit SHA and a one-line summary of what changed. Use the REST `databaseId` of the FIRST comment in the thread as `in_reply_to`:

   ```bash
   gh api repos/<owner>/<repo>/pulls/<pr>/comments \
     -f body="Addressed in <sha>: <one-line summary>" \
     -F in_reply_to=<root_comment.databaseId>
   ```

   In the focused output, thread data is at `threadDetails.<thread-id>.comments[]`. Each comment has:
   - `databaseId` — the REST id you need for `in_reply_to`
   - The GraphQL thread id is the key of the `threadDetails` dict (e.g. `threadDetails["MIDAC..."]`) which you need for the resolve mutation.

2. Resolve the thread via GraphQL, using the thread's GraphQL node id (`review_threads[].id`):

   ```bash
   gh api graphql \
     -f query='mutation($id:ID!){resolveReviewThread(input:{threadId:$id}){thread{isResolved}}}' \
     -f id=<thread.id>
   ```

A bare resolve without the reply is forbidden — the reply is what makes the resolution auditable for the human reviewer and distinguishable from "silently closed". Without resolution, `unresolvedNew` drift makes the loop indistinguishable from a fresh request and the watcher will keep delegating the same item.

Do NOT resolve threads the delegate did not address — for example, a thread the human is actively discussing or a thread whose feedback was deliberately not applied. When in doubt, leave unresolved and delegate to the main agent.

## Opt-In Verify-OpenSpec Criteria

Only when the caller explicitly requires `verify-openspec` approval:

1. Read `verifyOpenspec.runState` and `verifyOpenspec.requiresOpenspecVerification`. The runState values are:
   - `none` — no label and no review for this PR
   - `pending-pickup` — `verify-openspec` label is currently applied but the workflow has not started
   - `in-progress` — workflow has picked up the label and removed it, but no review has arrived yet
   - `approved` — the verify-openspec workflow submitted APPROVED. Approvals are permanent; they do not go stale.
   - `changes-requested` — the verify-openspec workflow submitted CHANGES_REQUESTED; fix needed first.

   The verify-openspec workflow runs as `github-actions[bot]` (the standard GITHUB_TOKEN identity), not as a dedicated `verify-openspec[bot]` user. The script identifies its reviews by the body containing `OpenSpec verify` or `Verification Report` so other workflows that also post as `github-actions[bot]` are not confused with it.
2. **Label state clarification** — the `verify-openspec` workflow REMOVES its own label as soon as it picks up the PR. Therefore label absence on `pr.labels` is NOT a signal that verify "was never requested". Always read `verifyOpenspec.runState`, never `pr.labels`, when deciding whether to re-trigger.
3. **Apply the label** when `requiresOpenspecVerification` is `true`. This boolean is computed by the script and encodes every guardrail: no label if already approved, already pending, checks failing/pending, or any actionable item exists.
4. **End successfully** only when `verifyOpenspec.runState == "approved"` AND `checks.failed == 0`. Do not treat `pr.reviewDecision == "APPROVED"` or a green verify workflow check as equivalent.
5. **Stop with `timeout`** if `runState` does not transition to `"approved"` within the caller's `--max-duration`.

## Resilience

The script retries transient `gh` failures once with backoff, then exits with code `2` and prints a JSON body containing `"transient": true`. The watcher MUST treat exit code `2` as "retry on next tick", NOT as `blocked`. Only escalate to `blocked` when transient failures persist across multiple ticks.

## Guardrails

- Keep watcher context self-contained; return concise summaries to the main agent.
- Prefer fresh delegate subagents for delegated fixes; always pass the same `--state-file`.
- Never force-push unless the user explicitly requested it.
- Do not resolve review threads unless the current PR state actually addresses them (and follow the two-step thread-resolution protocol above when you do).
- Never re-apply the `verify-openspec` label when `requiresOpenspecVerification` is `false`.
- If a simple watcher fix fails or repeats, delegate it to the main agent.
