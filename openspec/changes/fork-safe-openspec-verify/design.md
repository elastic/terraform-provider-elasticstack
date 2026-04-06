## Context

The current `openspec-verify-label` workflow assumes `label_command` activation, activation-time label removal, and a same-repository execution model. That is a poor fit for pull requests from forks because `label_command` is authored against `pull_request`, while archive and push behavior assumes the workflow can safely mutate the triggering branch after an `APPROVE`.

Moving the workflow to `pull_request_target` changes the trust boundary. Deterministic pre-activation logic runs in the base repository context and can read pull request metadata, labels, and changed files even when the pull request head lives in a fork. The workflow must therefore classify the pull request before agent reasoning begins and expose that classification as prompt inputs rather than letting the agent infer whether privileged operations are allowed.

## Goals / Non-Goals

**Goals:**
- Allow maintainers to trigger `verify-openspec` on fork pull requests from the base repository context.
- Keep trigger matching, cleanup behavior, and archive/push eligibility as deterministic workflow decisions.
- Preserve the existing review model for same-repository pull requests.
- Prevent archive and push behavior for fork pull requests before the agent starts.
- Avoid running the existing repository bootstrap path against fork-controlled content in the trusted `pull_request_target` context.

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

Classify pull requests deterministically into trusted workspace mode vs fork API-only mode.
Pre-activation logic should derive a trust classification from immutable event data, using a same-repository test such as `github.event.pull_request.head.repo.id == github.repository_id`. Same-repository pull requests become `workspace` mode and `archive-push-allowed`; fork pull requests become `api-only` mode and `archive-push-disallowed`. The workflow should also publish human-readable reasons for both outputs so the prompt can explain why archive or push is unavailable.

Alternative considered: use `maintainer_can_modify` to allow archive/push for some fork pull requests.
Rejected because archive/push eligibility should stay conservative and deterministic, and the workflow should not depend on contributor-controlled fork settings or cross-repository write behavior.

Decouple review approval from archive/push eligibility.
The workflow should continue to compute review disposition from the OpenSpec change contents (`approval-eligible` vs `comment-only`) and independently compute whether archive/push is allowed. This lets fork pull requests still receive a substantive review outcome while guaranteeing that archive and push behavior remains disabled by deterministic policy.

Alternative considered: force all fork pull requests to end with `COMMENT`.
Rejected because the user's stated requirement is to disable archive/push for forks, not to reduce the review signal when verification otherwise passes.

Use an API-only verification path for fork pull requests.
Under `pull_request_target`, the existing same-repository bootstrap path (`checkout`, `make setup`, `npx openspec`) should remain limited to trusted workspace mode. Fork pull requests should be reviewed from pull request metadata, changed files, and diffs rather than by checking out fork-controlled content and executing repository bootstrap steps in the trusted workflow context.

Alternative considered: check out the fork pull request branch in `pull_request_target` and run the existing workflow unchanged.
Rejected because it would execute fork-controlled repository content in a trusted context that has review and cleanup authority.

Use a deterministic script step for trigger-label cleanup.
The `label_command` change moved label cleanup into the compiled activation flow rather than the agent safe-output flow, and the effective write scope moved with it: the activation path gained `issues: write` / `pull-requests: write` so it could remove the trigger label before the agent ran, while the later safe-output handling path no longer needed `issues: write`. This change should keep that architectural shape, but as a repository-authored step: after deterministic label verification succeeds, a script step should remove only `verify-openspec` from the triggering pull request and publish a deterministic cleanup result for the prompt if needed. Cleanup should no longer depend on terminal safe outputs or post-agent processing.

Alternative considered: restore `remove-labels` safe output.
Rejected because the user wants label cleanup to be deterministic rather than agent-controlled, and the permission model already showed that deterministic cleanup can be handled in the earlier workflow path.

Alternative considered: add a separate cleanup job.
Rejected because a simple deterministic script step near activation keeps cleanup close to trigger verification without reintroducing agent coupling or extra workflow phases.

## Risks / Trade-offs

- Reduced verification fidelity for fork pull requests -> The API-only path may lack some repository-local CLI validation, so the review body should make the deterministic execution mode clear.
- Larger workflow delta than a trigger-only change -> The change affects trigger handling, cleanup location, deterministic outputs, prompt instructions, bootstrap logic, and permissions; tests and the canonical spec should be updated together.
- Conservative archive/push eligibility may skip cases that could technically work -> This is acceptable because the goal is predictable and safe behavior, not maximum automation surface.
- Compiler output may differ materially when switching back to `pull_request_target` -> Regenerate and inspect the compiled `.md` and `.lock.yml` artifacts rather than reasoning from the source template alone.

## Open Questions

- None for proposal scope; this design intentionally chooses the conservative same-repository-only rule for archive/push eligibility.
