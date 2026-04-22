## Context

The current `openspec-verify-label` workflow assumes `label_command` activation and same-repository archive/push behavior. Moving the workflow to `pull_request_target` changes the activation model and requires deterministic pre-activation logic for trigger-label verification, cleanup, and archive/push eligibility.

The current implementation does not preserve a separate `workspace` versus `api-only` verification split. Instead, it keeps a single verification/bootstrap flow and uses deterministic pre-activation outputs only to control review disposition and whether archive/push is allowed.

## Goals / Non-Goals

**Goals:**
- Allow maintainers to trigger `verify-openspec` on fork pull requests from the base repository context.
- Keep trigger matching, cleanup behavior, and archive/push eligibility as deterministic workflow decisions.
- Preserve the existing review and bootstrap model used by the current workflow.
- Prevent archive and push behavior for fork pull requests before the agent starts.
- Align the change artifacts with the current implementation, including explicit `on.permissions` for deterministic pre-activation steps.

**Non-Goals:**
- Enabling archive or push behavior for fork pull requests.
- Changing active-change selection rules for `openspec/changes/<id>/`.
- Changing net-new change proposals from comment-only to approval-eligible.
- Reworking unrelated OpenSpec review rules or the broader AWF repository policy.

## Decisions

Use `pull_request_target` with explicit deterministic label verification.
The authored workflow should move back to `pull_request_target` with `types: [labeled]` and restore the dedicated `verify_label` step. This preserves repository-authored control over which label activates the expensive path and avoids relying on `label_command`, whose authored trigger contract is tied to `pull_request`.

Alternative considered: keep `label_command` and try to rely on compiler behavior for fork support.
Rejected because the authored workflow contract would still describe `pull_request`-scoped activation and would not make the trusted base-context execution model explicit.

Classify pull requests deterministically for archive/push eligibility.
Pre-activation logic should derive archive/push eligibility from immutable event data, using a same-repository test such as `github.event.pull_request.head.repo.id == github.repository_id`. Same-repository pull requests become `archive-push-allowed`; fork pull requests become `archive-push-disallowed`. The workflow should also publish human-readable reasons for these outputs so the prompt can explain why archive or push is unavailable.

Alternative considered: use `maintainer_can_modify` to allow archive/push for some fork pull requests.
Rejected because archive/push eligibility should stay conservative and deterministic, and the workflow should not depend on contributor-controlled fork settings or cross-repository write behavior.

Decouple review approval from archive/push eligibility.
The workflow should continue to compute review disposition from the OpenSpec change contents (`approval-eligible` vs `comment-only`) and independently compute whether archive/push is allowed. This lets fork pull requests still receive a substantive review outcome while guaranteeing that archive and push behavior remains disabled by deterministic policy.

Alternative considered: force all fork pull requests to end with `COMMENT`.
Rejected because the user's stated requirement is to disable archive/push for forks, not to reduce the review signal when verification otherwise passes.

Keep the current shared verification/bootstrap path.
Under `pull_request_target`, the current implementation keeps one verification/bootstrap path rather than introducing a separate `api-only` mode. The change artifacts should therefore describe deterministic archive/push gating and review behavior without requiring mode-specific prompt branches or pre-activation outputs that the workflow does not produce.

Alternative considered: require a separate `api-only` verification path for fork pull requests.
Rejected for this change scope because the implementation being documented does not include that split.

Use a deterministic script step for trigger-label cleanup.
The `label_command` change moved label cleanup into the compiled activation flow rather than the agent safe-output flow, and the effective write scope moved with it: the activation path gained `issues: write` / `pull-requests: write` so it could remove the trigger label before the agent ran, while the later safe-output handling path no longer needed `issues: write`. This change should keep that architectural shape, but as a repository-authored step: after deterministic label verification succeeds, a script step should remove only `verify-openspec` from the triggering pull request and publish a deterministic cleanup result for the prompt if needed. Cleanup should no longer depend on terminal safe outputs or post-agent processing.

Declare deterministic pre-activation permissions in frontmatter.
The authored workflow should declare `on.permissions` with the explicit scopes needed by deterministic pre-activation steps, including `issues: write` for label cleanup, `pull-requests: write` for pull-request operations in the deterministic path, and `contents: read` for repository access. The change artifacts should describe that explicit frontmatter contract rather than assuming the compiler alone will synthesize the required write scopes.

Alternative considered: restore `remove-labels` safe output.
Rejected because the user wants label cleanup to be deterministic rather than agent-controlled, and the permission model already showed that deterministic cleanup can be handled in the earlier workflow path.

Alternative considered: add a separate cleanup job.
Rejected because a simple deterministic script step near activation keeps cleanup close to trigger verification without reintroducing agent coupling or extra workflow phases.

## Risks / Trade-offs

- The workflow still executes a shared bootstrap path even for fork-triggered runs under `pull_request_target` -> this change records the current implementation rather than redesigning that behavior.
- Larger workflow delta than a trigger-only change -> The change affects trigger handling, cleanup location, deterministic outputs, prompt instructions, permissions, and generated artifacts; tests and the canonical spec should be updated together.
- Conservative archive/push eligibility may skip cases that could technically work -> This is acceptable because the goal is predictable and safe behavior, not maximum automation surface.
- Compiler output may differ materially when switching back to `pull_request_target` -> Regenerate and inspect the compiled `.md` and `.lock.yml` artifacts rather than reasoning from the source template alone.

## Open Questions

- None for proposal scope; this design intentionally chooses the conservative same-repository-only rule for archive/push eligibility.
