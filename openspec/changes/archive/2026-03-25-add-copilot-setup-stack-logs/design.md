## Context

The `copilot-setup-steps` workflow provisions a local Elastic Stack and supporting credentials so GitHub Copilot sessions can work against a repo-local environment. Unlike `.github/workflows/test.yml`, it does not currently emit Docker Compose logs when setup fails, which makes debugging bootstrap issues harder in manual validation runs and path-triggered checks.

## Goals / Non-Goals

**Goals:**

- Add failure-time Docker Compose log collection to `.github/workflows/copilot-setup-steps.yml`.
- Keep the behavior aligned with the existing diagnostic pattern in `.github/workflows/test.yml` by using `docker compose logs --no-color`.
- Document the requirement in the `ci-copilot-setup-steps` OpenSpec capability.

**Non-Goals:**

- Adding teardown behavior to the Copilot setup workflow.
- Introducing matrix-driven stack version inputs or other `test.yml`-specific environment handling that the Copilot workflow does not use today.
- Changing the successful-path setup sequence for stack bootstrap, dependency installation, API key creation, or Fleet setup.

## Decisions

1. **Use a dedicated failure-path step at the end of the workflow** — Place a `docker compose logs --no-color` step after the existing setup steps with an `if: failure()` condition so diagnostics appear only when the job has failed. **Alternative considered:** inlining logging into each setup step, which would duplicate logic and make failures noisier.
2. **Do not add `STACK_VERSION` or matrix coupling** — The Copilot setup workflow uses the repository’s default Docker Compose configuration rather than the acceptance-test matrix, so the log step should rely on the same default context already used by `make docker-fleet`. **Alternative considered:** copying the `test.yml` environment block verbatim, which would add unused configuration.
3. **Add diagnostics without teardown** — The workflow is a setup contract for Copilot, not a full acceptance-test lifecycle, so this change should only improve failure visibility. **Alternative considered:** pairing logs with `make docker-clean`, which would broaden scope beyond the requested debugging improvement.

## Risks / Trade-offs

- **[Risk] A failure before the stack is fully available may produce limited or empty compose logs** -> **Mitigation:** still run the compose log command because partial container output is often enough to diagnose bootstrap problems.
- **[Risk] Failure output becomes longer** -> **Mitigation:** restrict the log collection step to `if: failure()` and use `--no-color` for readable plain-text logs.
- **[Risk] The Copilot workflow drifts from `test.yml` over time** -> **Mitigation:** document the shared diagnostics behavior in OpenSpec so future workflow changes can keep the intent aligned even if surrounding steps differ.

## Migration Plan

- Land the workflow change and the delta spec together. No user migration is required; failed Copilot setup runs will begin emitting stack logs immediately after the workflow update is active.

## Open Questions

- None.
