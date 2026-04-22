## Context

`ci-aw-openspec-verification` currently mixes deterministic workflow concerns with agent reasoning. The markdown instructions tell the agent to re-check the triggering label, fetch pull request files, derive the active change id, enforce modified-only gates, and install OpenSpec tooling before doing the review work. GitHub Agentic Workflows support two step mechanisms that better fit this split:

- `on.steps` run in the `pre_activation` job and can expose step outputs as `needs.pre_activation.outputs.*`.
- Top-level custom `steps:` run before the agent in the agent job and are the right place for workspace-local setup such as `npm ci`.

That lets the workflow compute and publish scalar gate results before the agent starts, skip the expensive agent job when verification is not eligible, and still ensure the agent job has `node_modules` available for `npx openspec` when it does run.

## Goals / Non-Goals

**Goals:**
- Make label verification and change selection deterministic and reproducible before agent reasoning begins.
- Pass the selected change id and gate outcome into job conditions and, when needed, into the agent using workflow outputs instead of asking the agent to rediscover them.
- Install OpenSpec in deterministic job steps so the agent can assume `npx openspec` is ready.
- Preserve the current review, archive, and push behavior once a valid selected change reaches the agent.

**Non-Goals:**
- Replacing the agent with a fully non-agentic workflow.
- Changing the approve/comment decision rules, relevance semantics, or archive-on-approve policy.
- Redesigning label cleanup behavior beyond staying compatible with the separate cleanup change already in flight.

## Decisions

Use `on.steps` for label verification and PR-file change selection.
The workflow should move the current "Trigger and first gate" and "Pull request files and change selection" logic into deterministic pre-activation steps. A GitHub-script step can load the PR files API response, derive candidate change ids, enforce modified-only gating, and emit outputs such as `selected_change`, `selection_status`, and `selection_reason`.

Alternative considered: keep change selection in agent instructions.
Rejected because it duplicates deterministic logic in natural language, increases prompt size, and makes later workflow behavior depend on the agent rediscovering repository state correctly.

Alternative considered: use a separate helper job for selection.
Rejected because `on.steps` already places this logic in `pre_activation` and exposes outputs without adding another workflow job.

Use top-level custom `steps:` for OpenSpec setup.
`npm ci` should run as a deterministic custom step in the agent job, before the prompt is handed to the agent. This keeps setup in the same workspace the agent will use for `npx openspec`.

Alternative considered: run `npm ci` in `on.steps`.
Rejected because `on.steps` run in `pre_activation`, which is a separate job from the agent path. That would not reliably make `node_modules` available in the later agent workspace.

Skip the agent job when deterministic gates fail.
The agent job should run only when `needs.pre_activation.outputs.selection_status` indicates the PR is eligible for verification. When label verification or change selection fails, the workflow should stop before agent execution rather than paying agent-job cost just to emit a `noop`.

Alternative considered: always run the agent and emit a deterministic `noop`.
Rejected because it keeps the most expensive job in the path even when pre-activation already knows the run is ineligible.

Pass the selected change to the agent through pre-activation outputs.
When the agent job does run, the prompt should interpolate `needs.pre_activation.outputs.selected_change` and related outputs directly. This keeps the contract explicit and ensures the agent starts from a known selected change rather than reparsing PR files.

Alternative considered: pass PR file JSON or write a workspace file for the agent.
Rejected because job outputs are already supported and only a small scalar payload is needed. Passing the whole file list would bloat outputs and recreate parsing logic in the agent.

Retain `pull_request` `labeled` plus defensive label checking.
The workflow can continue to use the existing trigger filter while also keeping a deterministic label verification step for defense in depth and clearer outputs.

Alternative considered: switch to `label_command`.
Not chosen for this change because it alters trigger semantics and overlaps with the separate work around label lifecycle and cleanup timing.

## Risks / Trade-offs

- `npm ci` adds work to every qualifying run -> Run change-selection steps before setup and only proceed into the agent job when the workflow is actually eligible to verify.
- Pre-activation outputs have limited practical size -> Emit only scalar data such as selected change id, status, and human-readable failure reason.
- Skipping the agent removes agent-generated `noop` messaging on early exits -> Accept silent early exits for cost savings, or add deterministic GitHub-side messaging later if maintainers need it.
- Workflow logic moves from prose into shell or script steps -> Prefer a single scripted source of truth for selection and cover it in the requirements delta so future edits stay aligned.
- Cleanup must still happen when the agent is skipped -> Keep cleanup in a separate terminal job that tolerates `needs.agent.result == 'skipped'`.

## Migration Plan

- Update the `ci-aw-openspec-verification` delta spec to require deterministic selection outputs and deterministic OpenSpec setup.
- Update `.github/workflows/openspec-verify-label.md` to add `on.steps` for label verification and change selection, gate the agent job on pre-activation outputs, and add top-level custom `steps:` for `npm ci`.
- Regenerate `.github/workflows/openspec-verify-label.lock.yml` with `gh aw compile`.
- Validate the change with OpenSpec checks and a workflow compile/test pass.

## Open Questions

- Whether the change-selection step should emit any extra diagnostic outputs beyond `selected_change`, `selection_status`, and `selection_reason`.
