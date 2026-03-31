---
name: openspec-implementation-loop
description: Orchestrates an end-to-end implementation loop for a single OpenSpec change: select a change, delegate implementation to a dedicated subagent, run review and verification subagents, feed findings back for fixes, push to origin, and watch GitHub Actions until the branch is green or blocked. Use when the user wants to implement an approved OpenSpec proposal/change with iterative review and CI feedback.
license: MIT
compatibility: Requires openspec CLI, git, and GitHub CLI.
metadata:
  author: openspec
  version: "1.0"
---

Orchestrate an implementation loop around a single OpenSpec change.

**Input**: Optionally specify a change name. If omitted, you MUST ask the user which change to implement.

**High-level flow**

1. Select the change
2. Load OpenSpec context
3. Start a dedicated implementor subagent
4. Run review subagents in parallel
5. Send findings back to the implementor and repeat until clean
6. Push the branch to `origin`
7. Watch GitHub Actions and loop failures back to the implementor

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

2. **Load OpenSpec status and context**

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

   Read every file listed in `contextFiles`.

   **Handle states**:
   - If `state: "blocked"`: stop and explain what artifact is missing; suggest continuing the change artifacts first
   - If `state: "all_done"`: skip directly to the review stage
   - Otherwise: proceed to implementation

3. **Start the implementor subagent**

   Launch a dedicated write-capable subagent and keep reusing the same subagent for the full loop.

   The implementor prompt should instruct it to:
   - implement only the selected change
   - follow the `openspec-apply-change` skill/process for the change
   - read the OpenSpec context files before editing
   - keep changes minimal and focused
   - create small, focused git commits as coherent pieces of work are completed
   - run targeted validation while implementing
   - never push to remote

   Ask the implementor to report back with:
   - tasks completed
   - commits created
   - tests run
   - blockers or open questions

   **If the implementor is blocked**, stop and surface the blocker to the user.

4. **Determine the review strategy**

   Inspect the change artifacts and implementation to decide whether this is a Terraform entity change.

   Treat it as a Terraform entity change when the change is centered on a Terraform resource or data source, for example:
   - specs reference `Resource implementation:` or `Data source implementation:`
   - changed code is primarily in a resource/data source package

5. **Run review subagents in parallel**

   Launch the following review subagents at the same time after the implementor reports completion of a reviewable chunk:

   a. **Critical code review**
   - Review for coding standards, idiomatic Go/Terraform provider patterns, obvious logic issues, error handling gaps, and risky regressions
   - Return prioritized findings only

   b. **Proposal compliance review**
   - Run the `openspec-verify-change` skill/process for the same change
   - Return only actionable mismatches, missing work, or notable warnings

   c. **Coverage review for Terraform entities**
   - If this is a Terraform entity change, run the `schema-coverage` skill/process
   - Focus on untested or weakly tested high-risk attributes and behaviors

   d. **Coverage review for non-entity changes**
   - If this is not a Terraform entity change, run a thorough test analysis instead
   - Prefer explicit coverage tooling where possible, for example `go test -cover`
   - Identify high-risk code paths that lack direct test coverage

   Ask every reviewer to return:
   - severity
   - concise finding
   - evidence
   - recommended fix

6. **Aggregate findings and decide whether to loop**

   Combine the review outputs into a single actionable list.

   If there are no actionable findings:
   - proceed to push

   If there are actionable findings:
   - resume the same implementor subagent
   - give it the aggregated findings
   - ask it to fix them with minimal diffs and additional small focused commits
   - rerun the relevant reviews

   Repeat until:
   - local review is clear, or
   - the implementor becomes blocked, or
   - the same issue repeats without progress

   If the loop stalls, pause and ask the user how to proceed.

7. **Push the branch**

   After local review is clear:
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

8. **Watch GitHub Actions**

   After pushing:
   - inspect the runs for the current branch using `gh`
   - watch or poll the latest relevant workflow runs until they complete
   - expect the acceptance test suite to take around 15 minutes to complete
   - once the only remaining jobs are long-running acceptance tests, poll less frequently instead of checking aggressively

   If CI succeeds:
   - finish with a concise summary

   If CI fails:
   - collect the failing workflow, job, and relevant log details
   - resume the implementor subagent with those failures
   - ask it to fix the issues and commit the changes
   - push again
   - continue watching CI

   Repeat until:
   - CI is green, or
   - a failure cannot be resolved without user input

9. **Report final outcome**

   Summarize:
   - change name
   - schema
   - implementation/review/CI loop status
   - commits created during the loop
   - tests or coverage checks used
   - final CI state
   - any remaining blockers or risks

**Recommended subagent responsibilities**

- **Implementor**: makes code changes, updates tasks, runs targeted validation, creates small focused commits
- **Critical reviewer**: reviews code quality and logic
- **Spec reviewer**: checks the implementation against the approved OpenSpec change
- **Coverage reviewer**: checks test coverage quality using the appropriate strategy

**Guardrails**

- Operate on one change only
- Always read the OpenSpec context before implementation
- Reuse the same implementor subagent through the whole loop
- Run reviewers in parallel whenever possible
- Prefer actionable findings over style nitpicks
- Feed review and CI failures back into the loop instead of fixing them ad hoc outside the loop
- Keep commit sizes small and purpose-specific
- Stop and ask the user if the process becomes ambiguous or stuck
