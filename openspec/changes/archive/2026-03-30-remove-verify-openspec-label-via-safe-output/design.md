## Context

The `openspec-verify-label` workflow currently removes `verify-openspec` with a dedicated `completion_cleanup` job that runs `actions/github-script` against a custom inline script. GitHub Agentic Workflows already provide a built-in `remove-labels` safe output for issue and pull request label mutation, so the workflow can express the same behavior declaratively in frontmatter and in the agent prompt instead of carrying a bespoke cleanup job.

This workflow already depends on safe outputs for review submission and branch updates. Folding label cleanup into that same mechanism keeps post-run mutations in one contract, but it also means label removal is requested by the agent rather than guaranteed by a terminal non-agent job.

## Goals / Non-Goals

**Goals:**
- Replace the dedicated label-removal job and script with the built-in `remove-labels` safe output.
- Keep label cleanup scoped to the triggering pull request and the single `verify-openspec` label.
- Update the agent instructions so label removal is part of the workflow's declared safe-output behavior.

**Non-Goals:**
- Changing review, archive, push, or change-selection semantics.
- Broadening label mutation beyond `verify-openspec`.
- Preserving the old terminal-job cleanup behavior for runs that never reach agent completion.

## Decisions

Declare `remove-labels` in workflow safe outputs.
The workflow should add a `remove-labels` safe-output block with `target: triggering`, `allowed: [verify-openspec]`, and a low `max` suitable for a single cleanup action. This matches the built-in capability described in the GitHub Agentic Workflows safe-outputs reference and constrains the agent to removing only the workflow's trigger label.

Alternative considered: keep the existing `completion_cleanup` job.
Rejected because it duplicates safe-output behavior with extra workflow YAML, a custom script include, and a separate permission-bearing job.

Move cleanup instructions into the agent prompt.
The markdown prompt should explicitly tell the agent to emit the `remove-labels` safe output for `verify-openspec` as part of its terminal handling for the triggering pull request. That keeps the cleanup contract visible alongside the existing review and push safe outputs.

Alternative considered: rely on the safe-output declaration without prompt guidance.
Rejected because the agent needs explicit instructions to actually request the label-removal operation.

Remove the bespoke cleanup implementation.
The source workflow template and regenerated lock file should no longer define `completion_cleanup` or reference the inline label-removal script. After the change, label mutation should flow only through safe outputs.

Alternative considered: keep the cleanup job as fallback for skipped or failed agent runs.
Rejected for this change because the requested simplification is specifically to remove the extra job and use `remove-labels` instead.

## Risks / Trade-offs

- Agent-skipped or agent-failed runs will no longer have a separate terminal cleanup path -> Accept this narrower contract for now and document that cleanup occurs through agent-emitted safe outputs rather than an unconditional final job.
- Safe-output misuse could remove the wrong label -> Constrain the workflow with `allowed: [verify-openspec]` and instruct the agent to remove only that label.
- Workflow behavior moves from a script to prompt plus frontmatter -> Update the spec delta so future edits treat safe-output cleanup as part of the canonical workflow contract.
