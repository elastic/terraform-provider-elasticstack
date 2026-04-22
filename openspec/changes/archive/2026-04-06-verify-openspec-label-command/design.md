## Context

The current `openspec-verify-label` workflow is written as a filtered `pull_request` `labeled` trigger, then performs an explicit pre-activation label check and later asks the agent to remove `verify-openspec` through `remove-labels`. That was necessary when the workflow modeled the label as normal PR state, but it now duplicates behavior provided by GitHub Agentic Workflows `label_command`.

This change needs to preserve the rest of the workflow contract: PR-only activation, deterministic active-change selection, review submission, and archive/push behavior. The change should narrow scope to trigger mechanics and cleanup semantics without altering verification behavior.

## Goals / Non-Goals

**Goals:**
- Use `label_command` as the canonical trigger for `verify-openspec`.
- Keep activation limited to pull requests only.
- Remove workflow configuration and prompt instructions that only exist to clean up the trigger label manually.
- Preserve change-selection, review, archive, and push behavior once the workflow is activated.

**Non-Goals:**
- Changing how the workflow selects an active OpenSpec change from PR files.
- Changing approval vs comment-only disposition rules.
- Changing archive, commit, or push behavior after review submission.
- Expanding the trigger to issues or discussions.

## Decisions

Use `label_command` with the `verify-openspec` label and `events: [pull_request]` in the authored workflow source.
This makes the workflow explicitly model `verify-openspec` as a one-shot command on pull requests instead of a generic `labeled` event plus custom filtering. It keeps PR-only behavior while letting the compiler inject the label-removal flow automatically. The compiled `.lock.yml` may still normalize that declaration into lower-level `pull_request` / `labeled` wiring and compiler-managed activation steps; that output is acceptable as long as the source of truth remains `label_command`.

Alternative considered: keep `pull_request.types: [labeled]` with `names: [verify-openspec]`.
Rejected because it preserves the old custom trigger path and manual cleanup contract instead of using the built-in command semantics.

Remove the dedicated repository-authored label-verification gate from pre-activation outputs.
Because `label_command` already matches the target label, the workflow should no longer need a custom `verify_label` step, corresponding outputs, or an agent-job condition that depends on them. The remaining deterministic gate should be active-change selection eligibility. If the compiled lockfile still shows lower-level labeled-event wiring, the repository should not reintroduce a custom `verify_label` step or mirrored job condition solely to restate what `label_command` already guarantees.

Alternative considered: keep the explicit label verification step as defense in depth or to make the compiled lockfile visibly filter labels.
Rejected because it duplicates the trigger contract, adds prompt and output complexity, and increases maintenance burden without changing the intended activation surface.

Remove `remove-labels` from safe outputs and delete end-of-run cleanup instructions from the prompt.
Under `label_command`, label removal is managed by workflow activation behavior rather than by terminal agent outputs. The prompt should therefore stop instructing the agent to emit label cleanup, and the workflow should no longer request label-removal authority for the agent.

Alternative considered: retain `remove-labels` as a fallback cleanup path.
Rejected because it would overlap with `label_command` and keep unnecessary permission and prompt surface area.

Preserve the rest of the workflow body and generated outputs unless they depend specifically on label verification or manual cleanup.
The active-change selection step, review instructions, review safe outputs, archive behavior, and push behavior should remain intact so the change is low risk and easy to verify.

Alternative considered: refactor broader workflow structure while touching the trigger.
Rejected because it would increase risk and make it harder to validate that only trigger semantics changed.

## Risks / Trade-offs

- `label_command` compilation may inject slightly different generated workflow structure than the current hand-authored trigger model -> Regenerate the compiled workflow and review both the markdown source and `.lock.yml` for unexpected permission or cleanup changes, while treating compiler-expanded labeled-event wiring as acceptable.
- Removing the old label-verification step could hide assumptions in prompt text or job conditions -> Update the prompt and pre-activation outputs together so no references to `label_verified` remain.
- Automatic label removal now happens as part of trigger handling rather than agent terminal outputs -> Capture that contract explicitly in the OpenSpec delta so future changes do not reintroduce manual cleanup.
