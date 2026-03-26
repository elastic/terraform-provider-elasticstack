## Context

`verify-openspec` is currently used as an event trigger for the agentic verification workflow. After a run finishes, the label remains on the pull request, even though it no longer represents an in-flight verification request. Because the workflow can end in several ways (`APPROVE`, `COMMENT`, `noop`, or job failure), cleanup logic needs to be defined once in a terminal phase instead of being duplicated across each branch of the workflow instructions.

## Goals / Non-Goals

**Goals:**
- Make `verify-openspec` a transient trigger label that is removed before the workflow run fully completes.
- Ensure cleanup happens for successful verification, comment-only completion, and early noop exits.
- Capture any permission changes required for label mutation in the workflow contract.

**Non-Goals:**
- Changing how the workflow decides whether to approve, comment, archive, or noop.
- Introducing new labels or changing who is allowed to apply `verify-openspec`.
- Defining retry or requeue behavior beyond allowing maintainers to reapply the label later.

## Decisions

Use a dedicated completion cleanup phase.
The workflow design should treat label removal as a final step that runs after the verification/archive logic reaches a terminal state. This is more reliable than requiring every earlier branch to remove the label itself, and it matches the requirement that cleanup occur regardless of outcome.

Require label-write permission explicitly.
Removing a pull request label is a repository mutation distinct from review submission and branch pushes, so the spec should explicitly require the permission needed to edit labels. That keeps the contract clear when the markdown workflow and compiled lock file are regenerated.

Remove only the triggering `verify-openspec` label.
The cleanup behavior should be narrowly scoped to the workflow's own trigger label. Other labels on the pull request remain outside the workflow's responsibility.

Treat cleanup as part of successful completion of the workflow contract.
If the workflow reaches its final cleanup stage, it should attempt label removal every time. This keeps the label state aligned with the fact that the verification request has been consumed.

## Risks / Trade-offs

- Cleanup after a hard infrastructure failure may still be skipped -> The implementation should prefer a terminal `always()`-style phase, but the spec will define the intended behavior at workflow completion rather than guaranteeing recovery from platform outages.
- Adding label mutation permission broadens workflow authority -> Limit the new permission to what is necessary for removing the single trigger label.
- Removing the label automatically may surprise maintainers who expected it to persist as history -> The proposal keeps the label semantics focused on "run now" instead of "was run before."
