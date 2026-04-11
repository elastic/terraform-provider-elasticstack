## Context

The `schema-coverage-rotation` workflow currently spends agent budget on work that can be determined before the prompt is ever delivered: counting open `schema-coverage` issues and calculating how many new issues may be opened in the current run. That logic already has a crisp contract in the markdown workflow and directly controls whether the expensive agent path should run at all.

This change is intentionally narrower than the broader memory-script extraction work. It only moves issue-capacity calculation into deterministic pre-activation behavior and leaves entity discovery, selection, and memory mutation for a separate post-hook change.

## Goals / Non-Goals

**Goals:**
- Compute `open_schema_coverage_issues` and `issue_slots_available` before agent activation using a repository-local script.
- Move the authored workflow source into `.github/workflows-src/schema-coverage-rotation/` so the deterministic path follows the repository's templated workflow pattern.
- Expose scalar outputs that downstream jobs and the agent prompt can consume without rediscovering GitHub issue state.
- Skip the agent job entirely when there are no remaining schema-coverage issue slots.
- Reduce prompt size and ambiguity by removing issue-counting instructions from the agent workflow text.

**Non-Goals:**
- Moving entity-list construction or memory-file updates into pre-activation.
- Changing the `schema-coverage` open-issue cap or issue-label semantics.
- Changing the schema-coverage analysis rubric or issue body content.

## Decisions

Use a repository-local pre-activation helper script.
The workflow should call a checked-in script during `pre_activation` rather than embedding the issue query in prompt prose or inline workflow logic. This keeps the behavior reviewable, testable, and reusable across the markdown and compiled workflow artifacts.

Alternative considered: leave issue-slot calculation in the prompt.
Rejected because the result is deterministic, affects whether the agent should run at all, and costs agent time on runs that may be immediately ineligible.

Author the workflow from `.github/workflows-src/` and keep the inline GitHub-script wrapper thin.
The `schema-coverage-rotation` workflow should follow the `openspec-verify-label` source pattern: a templated workflow source under `.github/workflows-src/schema-coverage-rotation/`, any `actions/github-script` step using an inline wrapper file, and the reusable issue-slot logic extracted into `.github/workflows-src/lib/` for direct unit testing.

Alternative considered: keep all logic inline in `.github/workflows/schema-coverage-rotation.md`.
Rejected because inline workflow code is harder to test, easier to duplicate, and inconsistent with the repository's existing templated workflow pattern.

Publish explicit gate outputs for later workflow logic.
The pre-activation path should emit the open-issue count, slot count, and a short gate status or reason so later job conditions and the prompt can consume the result without rerunning the query.

Alternative considered: only emit `issue_slots_available`.
Rejected because a human-readable reason and the raw count make skipped runs easier to understand and debug.

Skip the agent job when no slots remain.
Once the slot count is known, the workflow should short-circuit before the agent job starts. The agent should not be invoked just to produce a deterministic no-op.

Alternative considered: always run the agent and have it call `noop`.
Rejected because it preserves the most expensive path for a condition the workflow already knows in advance.

Keep the agent prompt focused on analysis-only behavior.
When the agent does run, the prompt should interpolate the precomputed slot values and should not tell the agent to count issues itself.

Alternative considered: keep the counting instructions as defense in depth.
Rejected because duplicated logic creates drift risk between the deterministic workflow path and the natural-language instructions.

## Risks / Trade-offs

- GitHub issue search behavior could drift from the intended label semantics -> Keep the script query narrowly scoped to open issues with the `schema-coverage` label and exclude pull requests.
- The open-issue cap could become duplicated across workflow and script code -> Define the cap once in the scripted path or pass it as an explicit input to avoid mismatches.
- Skipped runs may provide less visibility than an agent-authored `noop` -> Emit a deterministic gate reason in workflow logs and outputs so operators can see why the agent was skipped.
- Moving logic into a script adds one more maintained artifact -> Prefer a small single-purpose helper with stable inputs and outputs.
- Introducing a templated workflow source adds a compile step -> Accept the extra generated/source pairing because it enables unit-tested logic and keeps authored workflow code consistent with existing patterns.

## Migration Plan

1. Add `.github/workflows-src/schema-coverage-rotation/` plus an extracted helper module and unit tests under `.github/workflows-src/lib/`.
2. Implement the issue-slot calculation in the extracted helper and call it from a thin pre-activation wrapper script in the templated workflow source.
3. Update the workflow to publish outputs, gate the agent job on the result, and narrow the prompt so it consumes precomputed slot data instead of querying GitHub for issue counts.
4. Recompile `.github/workflows/schema-coverage-rotation.md` and `.github/workflows/schema-coverage-rotation.lock.yml`, then run the relevant OpenSpec and workflow validation checks.

## Open Questions

- Whether the pre-activation helper should emit only step outputs or also write a short job summary for skipped runs.
