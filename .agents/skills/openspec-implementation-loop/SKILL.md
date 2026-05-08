---
name: "openspec-implementation-loop"
description: "Orchestrates an end-to-end implementation loop for a single OpenSpec change: select a change, ask commit-only vs PR delivery, implement one top-level task at a time with a fresh dedicated subagent for each task, run review and verification subagents after each completed top-level task, feed findings back for fixes, push to origin, then either watch GitHub Actions on the branch (commit mode) or create a PR and delegate PR monitoring to the pr-monitoring-loop skill (PR mode). Use when the user wants to implement an approved OpenSpec proposal/change with iterative review and CI feedback."
license: "MIT"
compatibility: "Requires openspec CLI, git, and GitHub CLI."
metadata:
  author: openspec
  version: "2.0"
---

Orchestrate an implementation loop around a single OpenSpec change.

**Input**: Optionally specify a change name. If omitted, you MUST ask the user which change to implement.

**High-level flow**

1. Select the change
2. **Ask delivery mode (commit vs PR) — do this immediately after selecting the change, not after implementation**
3. Load OpenSpec context
4. Determine remaining top-level tasks
5. For each remaining top-level task, start a fresh implementor subagent
6. Run required local validation after each top-level task
7. Run review subagents in parallel after each top-level task
8. Send findings back to that task's implementor and repeat until clean, then move to the next top-level task
9. Push the branch to `origin`
10. **Commit mode**: Watch GitHub Actions on the pushed branch/commit  
    **PR mode**: Create a PR, then use the `pr-monitoring-loop` skill to monitor CI, reviews, comments, mergeability, and `verify-openspec` approval through watcher and delegate subagents
11. Report final outcome

**Steps**

1. **Select exactly one change**

   If a name is provided, use it.

   Otherwise:
   - Run `openspec list --json`
   - Use the **AskUserQuestion tool** to let the user choose a single active change
   - Show the change name, schema if available, and status

   **IMPORTANT**:
   - Do NOT guess
   - Do NOT auto-select
   - Do NOT implement multiple changes in one run

   Always announce: `Using change: <name>`.

2. **Choose delivery mode (ask immediately — before loading context or starting the implementor)**

   Use **AskUserQuestion** (or an equivalent explicit user prompt) right after step 1. Do **not** defer this until after implementation or push.

   Offer two options:

   - **Commit-only**: Push your work to `origin` on the current branch. After each push, monitor GitHub Actions for the **branch / commits** you pushed (same behavior as the historical workflow).
   - **Pull request**: After the **initial** push of the implementation loop, **create a PR** (for example with `gh pr create`). Then monitor **PR** workflow runs (checks on the PR), and **actively handle PR reviews** as described in step 12.

   Record the user’s choice and refer to it from push onward (steps 10–12).

3. **Load OpenSpec status and context**

   Run:
   ```bash
   openspec status --change "<name>" --json
   openspec instructions apply --change "<name>" --json
   ```

   Parse the outputs to determine:
   - `schemaName`
   - current task progress
   - `state`
   - `contextFiles`
   - the ordered list of top-level tasks (for example `1`, `2`, `3`) and which of them are still incomplete

   Read every file listed in `contextFiles`.

   **Handle states**:
   - If `state: "blocked"`: stop and explain what artifact is missing; suggest continuing the change artifacts first
   - If `state: "all_done"`: skip directly to the push / CI stage because all top-level tasks are already complete
   - Otherwise: proceed to the per-top-level-task implementation loop

4. **Determine the remaining top-level tasks**

   Build an ordered queue of incomplete top-level tasks from the OpenSpec task list.

   Interpret a top-level task as the parent task number such as `1`, `2`, or `3`. Each top-level task includes all of its nested subtasks such as `1.1`, `1.2`, `1.3`.

   For each incomplete top-level task:
   - gather the subtasks that belong to it
   - understand the intended scope from the proposal/design/specs
   - process the top-level tasks sequentially unless the user explicitly asks for a different strategy

5. **Start a fresh implementor subagent for the current top-level task**

   Launch a dedicated write-capable subagent for the current top-level task only. Do **not** reuse the prior task's implementor for later top-level tasks.

   The implementor prompt should instruct it to:
   - implement only the selected change
   - focus only on the current top-level task and complete all of its nested subtasks before stopping
   - follow the `openspec-apply-change` skill/process for the change
   - sync delta specs when implementation requires spec synchronization, but never archive the change
   - ignore any task that asks for the change to be archived; `verify-openspec` is responsible for archiving when it is satisfied
   - read the OpenSpec context files before editing
   - keep changes minimal and focused
   - create small, focused git commits as coherent pieces of work are completed
   - run targeted validation while implementing
   - never push to remote

   Ask the implementor to report back with:
   - the top-level task completed
   - nested subtasks completed
   - commits created
   - tests run
   - blockers or open questions

   **If the implementor is blocked**, stop and surface the blocker to the user.

6. **Run required local validation for the current top-level task**

   Launch a dedicated validation subagent for the current top-level task and have that subagent run the local validation expected by this repository.

   Use the validation subagent so these checks do **not** consume the orchestrator's working context. Keep the validation work encapsulated in that subagent and ask it to return a concise validation summary.

   This validation subagent should run in parallel with the other local review subagents for the same top-level task, not as a separate serial phase beforehand.

   Required baseline commands:
   - run `make lint`
   - run `make build`

   Required acceptance-test validation:
   - determine which acceptance tests are relevant to the code changed by the current top-level task
   - prefer targeted acceptance test runs using `go test -v [-run 'filter'] <package>`
   - run them with `TF_ACC=1`
   - include the commonly required environment variables from `dev-docs/high-level/testing.md` when needed: `ELASTICSEARCH_ENDPOINTS`, `ELASTICSEARCH_USERNAME`, `ELASTICSEARCH_PASSWORD`, and `KIBANA_ENDPOINT`
   - before considering starting a new local stack, check whether the Elastic stack is already running using the guidance in `dev-docs/high-level/testing.md`
   - only run `make testacc` when the user explicitly instructs you to run the full acceptance suite

   If the current top-level task does not affect code that has relevant acceptance tests, say so explicitly and explain why targeted acceptance coverage is not applicable.

   Ask the validation subagent to report back with:
   - commands run
   - acceptance tests selected
   - pass/fail status
   - relevant logs or failure summaries
   - an explicit explanation when acceptance tests were not applicable

7. **Determine the review strategy**

   Inspect the change artifacts and implementation to decide whether this is a Terraform entity change.

   Treat it as a Terraform entity change when the change is centered on a Terraform resource or data source, for example:
   - specs reference `Resource implementation:` or `Data source implementation:`
   - changed code is primarily in a resource/data source package

8. **Run review subagents in parallel**

   Launch the following at the same time immediately after the current top-level task's implementor reports that the full top-level task is complete:

   a. **Validation runner**
   - Run the step 6 local validation work in its dedicated subagent
   - Return the validation summary described in step 6

   b. **Critical code review**
   - Review for coding standards, idiomatic Go/Terraform provider patterns, obvious logic issues, error handling gaps, and risky regressions
   - Return prioritized findings only

   c. **Proposal compliance review**
   - Run the `openspec-verify-change` skill/process for the same change
   - Return only actionable mismatches, missing work, or notable warnings

   d. **Coverage review for Terraform entities**
   - If this is a Terraform entity change, run the `schema-coverage` skill/process
   - Focus on untested or weakly tested high-risk attributes and behaviors

   e. **Coverage review for non-entity changes**
   - If this is not a Terraform entity change, run a thorough test analysis instead
   - Prefer explicit coverage tooling where possible, for example `go test -cover`
   - Identify high-risk code paths that lack direct test coverage

   Ask every validation/review subagent to return:
   - severity
   - concise finding
   - evidence
   - recommended fix

9. **Aggregate findings for the current top-level task and decide whether to loop**

   Combine the validation results and review outputs into a single actionable list.

   If there are no actionable findings:
   - mark the current top-level task as locally complete
   - start a fresh implementor for the next incomplete top-level task, if any
   - if no top-level tasks remain, proceed to push

   If there are actionable findings:
   - resume the current top-level task's implementor subagent
   - give it the aggregated findings
   - ask it to fix them with minimal diffs and additional small focused commits
   - rerun the required local validation and relevant reviews for that same top-level task before moving on

   Repeat until:
   - the current top-level task passes local review and either advances to the next top-level task or the loop reaches push readiness, or
   - the current implementor becomes blocked, or
   - the same issue repeats without progress

   If the loop stalls, pause and ask the user how to proceed.

10. **Push the branch**

   After every incomplete top-level task has been implemented and passed local review:
   - verify the branch state is ready to push
   - push the current branch to `origin`
   - use upstream tracking if needed

   Example:
   ```bash
   git push -u origin HEAD
   ```

   **Guardrails**:
   - Never force-push unless the user explicitly asks
   - Do not push before local review passes

11. **Commit-only mode: watch GitHub Actions (branch / commits)**

   If the user chose **commit-only** in step 2:

   After pushing:
   - inspect workflow runs for the current branch and the **commits** you pushed (for example with `gh run list` filtered by branch, or `gh` against the latest commit SHA)
   - watch or poll the latest relevant workflow runs until they complete
   - expect the acceptance test suite to take around 15 minutes to complete
   - once the only remaining jobs are long-running acceptance tests, poll less frequently instead of checking aggressively

   If CI succeeds:
   - finish with a concise summary (or continue to step 13 if you already reported)

   If CI fails:
   - collect the failing workflow, job, and relevant log details
   - launch a fresh write-capable implementor subagent scoped only to resolving the CI failures
   - ask it to fix the issues and commit the changes
   - push again
   - continue watching CI

   Repeat until:
   - CI is green, or
   - a failure cannot be resolved without user input

12. **PR mode: create PR, watch PR checks, poll reviews, address feedback**

    If the user chose **pull request** in step 2:

    **Create the PR after the initial push** (step 10), if it does not already exist:
    - use `gh pr create` (or equivalent) with an appropriate title and body tied to the OpenSpec change
    - record the PR number or URL

    **State file**:
    - the script auto-uses `.agents/skills/pr-monitoring-loop/scripts/state/.pr-monitor-<pr>.json` (gitignored), so subagents can simply invoke `check-pr-state.py <pr>` and `lastPolledAt` / seen IDs persist across watcher restarts without any explicit path management
    - pass `--state-file <path>` only when you need isolation (e.g., parallel watchers on different branches that share a PR number, or tests). When you do override, pass the same path to every subagent in this loop
    - do NOT put the state file under `.git/` — this repo uses git worktrees, where `.git` is a file that points at a worktree-specific git dir, which would fragment state across worktrees watching the same PR

    **Delegate PR monitoring to `pr-monitoring-loop`**:
    - load and follow the `pr-monitoring-loop` skill for the entire PR monitoring phase
    - start a delegate subagent for the PR instead of polling in this main implementation-loop agent
    - instruct the delegate to invoke `.agents/skills/pr-monitoring-loop/scripts/check-pr-state.py <pr>` on every cadence tick (or use `--watch` with the cadence guidance documented in `pr-monitoring-loop`) so CI (commit-pinned), reviews, PR comments, review comments, unresolved threads, merge conflicts, and stale branch state are checked together. Append `--state-file <path>` only if you decided to override the default path above.
    - allow the delegate to fix and push changes it judges simple, then perform the thread-resolution protocol for any addressed threads (reply with addressing commit SHA, then `resolveReviewThread`), and continue watching the new PR head
    - when the delegate returns `delegate`, launch a fresh delegate subagent scoped only to the reported failure or feedback (with the same `--state-file`); after the delegate commits, pushes, replies, and resolves addressed threads, restart the watch cycle for the new head commit with a new delegate

    When monitoring in PR mode, tell the PR watcher to opt in to `verify-openspec` behavior per the `pr-monitoring-loop` skill. The `pr-monitoring-loop` skill defines when to apply the `verify-openspec` label (via `requiresOpenspecVerification`) and when the PR is considered verified (`runState == "approved"`). Do not restate those rules here.

13. **Report final outcome**

    Summarize:
    - change name
    - schema
    - delivery mode (commit-only vs PR)
    - implementation/review/CI loop status
    - top-level tasks completed in the loop
    - commits created during the loop
    - local validation run during the loop (`make lint`, `make build`, and relevant acceptance tests or an explicit explanation when acceptance tests were not applicable)
    - tests or coverage checks used
    - final CI state (and PR link if PR mode)
    - PR review handling summary if PR mode (including polling cadence, whether the loop restarted after pushes, and whether `verify-openspec` approved the PR or timed out)
    - any remaining blockers or risks

**Recommended subagent responsibilities**

- **Implementor**: one fresh implementor per top-level task; each implementor makes code changes for its assigned task, updates tasks, runs targeted validation, and creates small focused commits
- **Validation runner**: a fresh subagent for each top-level task that runs `make lint`, `make build`, and relevant acceptance tests, then reports a concise validation summary
- **Critical reviewer**: reviews code quality and logic
- **Spec reviewer**: checks the implementation against the approved OpenSpec change
- **Coverage reviewer**: checks test coverage quality using the appropriate strategy
- **PR watcher**: a fresh subagent using `pr-monitoring-loop` to poll PR state, directly fix simple actionable issues, and return non-simple work to the main agent for delegation to a fresh subagent

**Guardrails**

- Operate on one change only
- Ask **commit vs PR** at the **start** (step 2), not when implementation is finished
- Always read the OpenSpec context before implementation
- Create a fresh implementor subagent for each top-level task; do not reuse one implementor across the whole change
- Do not advance to the next top-level task until the current task has passed local review
- Never archive a change in this workflow; only sync delta specs when needed
- Ignore tasks that request archiving the change proposal; `verify-openspec` will archive the change when it is happy
- Run local validation in a dedicated subagent so the orchestrator does not spend its own context on lint/build/test execution
- Run the validation subagent in parallel with the other local review subagents for the same top-level task
- Local review for each top-level task must include `make lint`, `make build`, and relevant acceptance tests run according to `dev-docs/high-level/testing.md`
- Run reviewers in parallel whenever possible
- Prefer actionable findings over style nitpicks
- Feed local review and commit-mode CI failures back into the loop instead of fixing them ad hoc outside the loop
- In PR mode, use `pr-monitoring-loop`; do not spend main-agent context on repeated PR polling
- Keep commit sizes small and purpose-specific
- Stop and ask the user if the process becomes ambiguous or stuck
