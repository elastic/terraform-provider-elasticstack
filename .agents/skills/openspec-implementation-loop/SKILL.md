---
name: "openspec-implementation-loop"
description: "Orchestrates an end-to-end implementation loop for a single OpenSpec change: select a change, ask commit-only vs PR delivery, triage the change to determine an execution strategy (inline, single-implementor, or per-task) based on change size and complexity, implement tasks using the chosen strategy, run review and validation, feed findings back for fixes, push to origin, then either watch GitHub Actions on the branch (commit mode) or create a PR and delegate PR monitoring to the pr-monitoring-loop skill (PR mode). Use when the user wants to implement an approved OpenSpec proposal/change with iterative review and CI feedback."
license: "MIT"
compatibility: "Requires openspec CLI, git, and GitHub CLI."
metadata:
  author: openspec
  version: "3.0"
---

Orchestrate an implementation loop around a single OpenSpec change.

**Input**: Optionally specify a change name. If omitted, you MUST ask the user which change to implement.

**High-level flow**

1. Select the change
2. **Ask delivery mode (commit vs PR) - do this immediately after selecting the change, not after implementation**
3. Load OpenSpec context
4. **Triage: determine execution strategy** - classify as inline, single-implementor, or per-task based on change size and complexity
5. Determine remaining top-level tasks
6. Implement tasks using the chosen strategy
7. Run validation and review at the cadence determined by the strategy
8. Aggregate findings and fix; repeat until clean
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

2. **Choose delivery mode (ask immediately - before loading context or starting the implementor)**

   Use **AskUserQuestion** (or an equivalent explicit user prompt) right after step 1. Do **not** defer this until after implementation or push.

   Offer two options:

   - **Commit-only**: Push your work to `origin` on the current branch. After each push, monitor GitHub Actions for the **branch / commits** you pushed (same behavior as the historical workflow).
   - **Pull request**: After the **initial** push of the implementation loop, **create a PR** (for example with `gh pr create`). Then monitor **PR** workflow runs (checks on the PR), and **actively handle PR reviews** as described in step 11.

   Record the user's choice and refer to it from push onward (steps 9–11).

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
   - Otherwise: proceed to triage and implementation

4. **Triage: determine execution strategy**

   Evaluate the change to choose an execution strategy that balances implementation quality with subagent overhead. The goal is to avoid spawning unnecessary subagents for small changes while preserving full rigor for large ones.

   **Signals to evaluate:**

   | Signal | Source |
   |--------|--------|
   | Top-level task count | Count `## N.` headings in tasks |
   | Total subtask count | Count `- [ ]` / `- [x]` items |
   | File scope | Infer from task descriptions - single package/area vs. cross-cutting |
   | Inter-task coupling | Do later tasks build directly on earlier ones? |
   | Complexity | Do tasks involve non-trivial logic (custom types, plan modifiers, complex CRUD) or are they straightforward (config, docs, Makefile, CI, specs)? |

   **Choose one of three strategies:**

   | Strategy | When to use | Implementation | Review |
   |----------|-------------|----------------|--------|
   | **Inline** | ≤2 top-level tasks AND ≤~10 subtasks AND single area AND straightforward changes | Orchestrator implements directly, no implementor subagent | Run validation commands directly; spawn review subagents only if the change touches non-trivial logic |
   | **Single-implementor** | ≤4 top-level tasks OR ≤~15 subtasks, with coherent scope | One implementor subagent handles all remaining tasks | One round of parallel review after all tasks complete |
   | **Per-task** | >4 top-level tasks, OR >15 subtasks, OR multi-area scope, OR tasks are largely independent across different packages | Fresh implementor per top-level task | Full parallel review after each top-level task |

   These thresholds are guidelines, not rigid rules. Use judgment:
   - A 3-task change where each task is in a different package might warrant **per-task**
   - A 5-task change that is all in one file might be fine as **single-implementor**
   - A 2-task change with complex custom type logic might benefit from **single-implementor** over **inline** for the review coverage

   Announce the chosen strategy and reasoning. The user can override.

5. **Determine the remaining top-level tasks**

   Build an ordered queue of incomplete top-level tasks from the OpenSpec task list.

   Interpret a top-level task as the parent task number such as `1`, `2`, or `3`. Each top-level task includes all of its nested subtasks such as `1.1`, `1.2`, `1.3`.

   For each incomplete top-level task:
   - gather the subtasks that belong to it
   - understand the intended scope from the proposal/design/specs
   - process the top-level tasks sequentially unless the user explicitly asks for a different strategy

6. **Implement tasks using the chosen strategy**

   **Inline strategy:**

   The orchestrator implements all tasks directly without spawning an implementor subagent:
   - follow the `openspec-apply-change` skill/process for the change
   - sync delta specs when implementation requires spec synchronization, but never archive the change
   - ignore any task that asks for the change to be archived; `verify-openspec` is responsible for archiving when it is satisfied
   - keep changes minimal and focused
   - create small, focused git commits as coherent pieces of work are completed
   - run targeted validation while implementing
   - never push to remote

   **Single-implementor strategy:**

   Launch one write-capable subagent for all remaining tasks:
   - instruct it to implement all remaining top-level tasks in sequence, completing all nested subtasks within each before moving to the next
   - follow the `openspec-apply-change` skill/process for the change
   - sync delta specs when implementation requires spec synchronization, but never archive the change
   - ignore any task that asks for the change to be archived; `verify-openspec` is responsible for archiving when it is satisfied
   - read the OpenSpec context files before editing
   - keep changes minimal and focused
   - create small, focused git commits as coherent pieces of work are completed
   - run targeted validation while implementing
   - never push to remote

   Ask the implementor to report back with:
   - top-level tasks completed
   - nested subtasks completed
   - commits created
   - tests run
   - blockers or open questions

   **Per-task strategy:**

   Launch a dedicated fresh write-capable subagent for the current top-level task only. Do **not** reuse the prior task's implementor for later top-level tasks.

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

   **For all strategies**: if the implementor (or the orchestrator in inline mode) is blocked, stop and surface the blocker to the user.

7. **Run validation and review**

   The validation and review cadence depends on the execution strategy.

   **7a. Determine the review type**

   Inspect the change artifacts and implementation to decide whether this is a Terraform entity change.

   Treat it as a Terraform entity change when the change is centered on a Terraform resource or data source, for example:
   - specs reference `Resource implementation:` or `Data source implementation:`
   - changed code is primarily in a resource/data source package

   **7b. Validation requirements (all strategies)**

   Required baseline commands:
   - run `make lint`
   - run `make build`

   Required acceptance-test validation:
   - determine which acceptance tests are relevant to the code changed
   - prefer targeted acceptance test runs using `go test -v [-run 'filter'] <package>`
   - run them with `TF_ACC=1`
   - include the commonly required environment variables from `dev-docs/high-level/testing.md` when needed: `ELASTICSEARCH_ENDPOINTS`, `ELASTICSEARCH_USERNAME`, `ELASTICSEARCH_PASSWORD`, and `KIBANA_ENDPOINT`
   - before considering starting a new local stack, check whether the Elastic stack is already running using the guidance in `dev-docs/high-level/testing.md`
   - only run `make testacc` when the user explicitly instructs you to run the full acceptance suite

   If the change does not affect code that has relevant acceptance tests, say so explicitly and explain why targeted acceptance coverage is not applicable.

   **7c. Inline strategy: lightweight review**

   The orchestrator runs validation commands (`make lint`, `make build`, relevant acceptance tests) directly rather than spawning a validation subagent.

   For straightforward changes (config, docs, Makefile, CI, spec-only), a self-review by the orchestrator is sufficient. Do not spawn review subagents.

   For changes that touch non-trivial logic (custom types, plan modifiers, complex CRUD, error handling), spawn a minimal set of review subagents:
   - **Critical code review**: review for coding standards, idiomatic patterns, logic issues, error handling gaps
   - **Proposal compliance review**: run the `openspec-verify-change` skill/process
   - Add coverage review only if the change involves a Terraform entity

   **7d. Single-implementor strategy: one review round**

   After all tasks are complete, launch the full parallel review battery **once**:

   a. **Validation runner** - launch a dedicated validation subagent to run the validation requirements from step 7b. Use a subagent so these checks do not consume the orchestrator's working context.

   b. **Critical code review** - review for coding standards, idiomatic Go/Terraform provider patterns, obvious logic issues, error handling gaps, and risky regressions. Return prioritized findings only.

   c. **Proposal compliance review** - run the `openspec-verify-change` skill/process for the same change. Return only actionable mismatches, missing work, or notable warnings.

   d. **Coverage review for Terraform entities** - if this is a Terraform entity change, run the `schema-coverage` skill/process. Focus on untested or weakly tested high-risk attributes and behaviors.

   e. **Coverage review for non-entity changes** - if this is not a Terraform entity change, run a thorough test analysis instead. Prefer explicit coverage tooling where possible, for example `go test -cover`. Identify high-risk code paths that lack direct test coverage.

   Run the validation runner in parallel with the other review subagents, not as a separate serial phase.

   **7e. Per-task strategy: review after each top-level task**

   After each top-level task's implementor reports completion, launch the same full parallel review battery described in 7d (validation runner, critical code review, proposal compliance review, and the appropriate coverage review).

   Run the validation runner in parallel with the other review subagents for the same top-level task.

   **For all review subagents**, ask them to return:
   - severity
   - concise finding
   - evidence
   - recommended fix

8. **Aggregate findings and decide whether to loop**

   Combine the validation results and review outputs into a single actionable list.

   If there are no actionable findings:
   - **Inline / single-implementor**: proceed to push
   - **Per-task**: mark the current top-level task as locally complete; start a fresh implementor for the next incomplete top-level task, or proceed to push if none remain

   If there are actionable findings:
   - **Inline**: the orchestrator fixes the issues directly with minimal diffs and additional small focused commits
   - **Single-implementor**: resume the implementor subagent, give it the aggregated findings, and ask it to fix them with minimal diffs and additional small focused commits
   - **Per-task**: resume the current top-level task's implementor subagent, give it the aggregated findings, and ask it to fix them with minimal diffs and additional small focused commits

   After fixes, rerun validation and relevant reviews before advancing:
   - **Inline / single-implementor**: rerun before proceeding to push
   - **Per-task**: rerun before moving to the next top-level task

   Repeat until:
   - all tasks pass review and the loop reaches push readiness, or
   - the implementor (or orchestrator) becomes blocked, or
   - the same issue repeats without progress

   If the loop stalls, pause and ask the user how to proceed.

9. **Push the branch**

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

10. **Commit-only mode: watch GitHub Actions (branch / commits)**

   If the user chose **commit-only** in step 2:

   After pushing:
   - inspect workflow runs for the current branch and the **commits** you pushed (for example with `gh run list` filtered by branch, or `gh` against the latest commit SHA)
   - watch or poll the latest relevant workflow runs until they complete
   - expect the acceptance test suite to take around 15 minutes to complete
   - once the only remaining jobs are long-running acceptance tests, poll less frequently instead of checking aggressively

   If CI succeeds:
   - finish with a concise summary (or continue to step 12 if you already reported)

   If CI fails:
   - collect the failing workflow, job, and relevant log details
   - launch a fresh write-capable implementor subagent scoped only to resolving the CI failures
   - ask it to fix the issues and commit the changes
   - push again
   - continue watching CI

   Repeat until:
   - CI is green, or
   - a failure cannot be resolved without user input

11. **PR mode: create PR, watch PR checks, poll reviews, address feedback**

    If the user chose **pull request** in step 2:

    **Create the PR after the initial push** (step 9), if it does not already exist:
    - use `gh pr create` (or equivalent) with an appropriate title and body tied to the OpenSpec change
    - record the PR number or URL

    **State file**:
    - the script auto-uses `.agents/skills/pr-monitoring-loop/scripts/state/.pr-monitor-<pr>.json` (gitignored), so subagents can simply invoke `check-pr-state.py <pr>` and `lastPolledAt` / seen IDs persist across watcher restarts without any explicit path management
    - pass `--state-file <path>` only when you need isolation (e.g., parallel watchers on different branches that share a PR number, or tests). When you do override, pass the same path to every subagent in this loop
    - do NOT put the state file under `.git/` - this repo uses git worktrees, where `.git` is a file that points at a worktree-specific git dir, which would fragment state across worktrees watching the same PR

    **Delegate PR monitoring to `pr-monitoring-loop`**:
    - load and follow the `pr-monitoring-loop` skill for the entire PR monitoring phase
    - start a delegate subagent for the PR instead of polling in this main implementation-loop agent
    - instruct the delegate to invoke `.agents/skills/pr-monitoring-loop/scripts/check-pr-state.py <pr>` on every cadence tick (or use `--watch` with the cadence guidance documented in `pr-monitoring-loop`) so CI (commit-pinned), reviews, PR comments, review comments, unresolved threads, merge conflicts, and stale branch state are checked together. Append `--state-file <path>` only if you decided to override the default path above.
    - allow the delegate to fix and push changes it judges simple, then perform the thread-resolution protocol for any addressed threads (reply with addressing commit SHA, then `resolveReviewThread`), and continue watching the new PR head
    - when the delegate returns `delegate`, launch a fresh delegate subagent scoped only to the reported failure or feedback (with the same `--state-file`); after the delegate commits, pushes, replies, and resolves addressed threads, restart the watch cycle for the new head commit with a new delegate

    When monitoring in PR mode, tell the PR watcher to opt in to `verify-openspec` behavior per the `pr-monitoring-loop` skill. The `pr-monitoring-loop` skill defines when to apply the `verify-openspec` label (via `requiresOpenspecVerification`) and when the PR is considered verified (`runState == "approved"`). Do not restate those rules here.

12. **Report final outcome**

    Summarize:
    - change name
    - schema
    - delivery mode (commit-only vs PR)
    - execution strategy used (inline, single-implementor, or per-task) and reasoning
    - implementation/review/CI loop status
    - top-level tasks completed in the loop
    - commits created during the loop
    - local validation run during the loop (`make lint`, `make build`, and relevant acceptance tests or an explicit explanation when acceptance tests were not applicable)
    - tests or coverage checks used
    - final CI state (and PR link if PR mode)
    - PR review handling summary if PR mode (including polling cadence, whether the loop restarted after pushes, and whether `verify-openspec` approved the PR or timed out)
    - any remaining blockers or risks

**Recommended subagent responsibilities**

Subagent usage scales with the chosen strategy:

- **Implementor** (single-implementor and per-task strategies): makes code changes, updates tasks, runs targeted validation, and creates small focused commits. Per-task uses one fresh implementor per top-level task; single-implementor uses one for the entire change.
- **Validation runner** (single-implementor and per-task strategies): a subagent that runs `make lint`, `make build`, and relevant acceptance tests, then reports a concise validation summary. Use a subagent so validation does not consume the orchestrator's context.
- **Critical reviewer**: reviews code quality and logic
- **Spec reviewer**: checks the implementation against the approved OpenSpec change
- **Coverage reviewer**: checks test coverage quality using the appropriate strategy
- **PR watcher**: a fresh subagent using `pr-monitoring-loop` to poll PR state, directly fix simple actionable issues, and return non-simple work to the main agent for delegation to a fresh subagent

For the **inline** strategy, the orchestrator fills the implementor and validation runner roles directly. Review subagents are spawned only when the change touches non-trivial logic.

**Guardrails**

- Operate on one change only
- Ask **commit vs PR** at the **start** (step 2), not when implementation is finished
- Always triage the change and announce the execution strategy before implementation
- The user can override the chosen strategy at the triage step
- Always read the OpenSpec context before implementation
- **Per-task strategy**: create a fresh implementor subagent for each top-level task; do not reuse one implementor across top-level tasks. Do not advance to the next top-level task until the current task has passed local review.
- **Single-implementor strategy**: use one implementor for all tasks; run one review round after all tasks complete
- **Inline strategy**: the orchestrator implements directly; spawn review subagents only for non-trivial logic changes
- Never archive a change in this workflow; only sync delta specs when needed
- Ignore tasks that request archiving the change proposal; `verify-openspec` will archive the change when it is happy
- For single-implementor and per-task strategies, run local validation in a dedicated subagent so the orchestrator does not spend its own context on lint/build/test execution
- Run the validation subagent in parallel with the other review subagents, not as a separate serial phase
- All strategies must include `make lint`, `make build`, and relevant acceptance tests run according to `dev-docs/high-level/testing.md` (or an explicit explanation when acceptance tests are not applicable)
- Run reviewers in parallel whenever possible
- Prefer actionable findings over style nitpicks
- Feed local review and commit-mode CI failures back into the loop instead of fixing them ad hoc outside the loop
- In PR mode, use `pr-monitoring-loop`; do not spend main-agent context on repeated PR polling
- Keep commit sizes small and purpose-specific
- Stop and ask the user if the process becomes ambiguous or stuck
