## Context

The `copilot-setup-steps` workflow no longer embeds default Elastic Stack credentials in `jobs.copilot-setup-steps.env`. It also no longer exports shared connection settings through workflow-level `env`. Instead, the workflow relies on the existing repository `.env` file and Makefile defaults for bootstrap and authentication values, while keeping `FLEET_NAME` explicitly set on the Fleet step. This proposal changes a single workflow and its corresponding OpenSpec capability so the documented contract matches that implementation.

## Goals / Non-Goals

**Goals:**

- Remove the job-level credential defaults from `.github/workflows/copilot-setup-steps.yml`.
- Document that the workflow no longer declares top-level or per-step credential env overrides for the setup targets.
- Update the `ci-copilot-setup-steps` OpenSpec capability to document the repository `.env` and Makefile defaults that the workflow now relies on.

**Non-Goals:**

- Changing unrelated workflow improvements added after `6501f04`, such as the Node.js setup step or failure-log collection behavior.
- Redesigning how the Makefile or Docker Compose files interpret credential environment variables.
- Introducing new secret names, new GitHub environments, or a broader CI credential-management convention outside this workflow.

## Decisions

1. **Remove only the job-local default credential block** — Keep later unrelated workflow improvements, but stop declaring credential defaults in `jobs.copilot-setup-steps.env`. **Alternative considered:** reverting the earlier change mechanically, which would also discard later unrelated updates.
2. **Leave bootstrap and authentication values to existing defaults** — The workflow requirements will state that the setup targets rely on the repository `.env` file and Makefile defaults unless the execution environment overrides them. **Alternative considered:** exporting workflow-level variables or adding step-specific credential wiring, which no longer matches the current YAML.
3. **Keep only the Fleet-specific override explicit** — The workflow requirements will document that `FLEET_NAME` remains explicitly set to `fleet` on the Fleet setup step so the Makefile points at the expected Compose service name. **Alternative considered:** relying on the Makefile default `FLEET_NAME`, which would no longer match the current workflow behavior.

## Risks / Trade-offs

- **[Risk] Repository `.env` or Makefile defaults can drift from the workflow’s documented assumptions** -> **Mitigation:** document which targets rely on those defaults and verify the affected targets against the current repo defaults.
- **[Risk] Removing explicit credential wiring makes the workflow contract less obvious when reading the YAML alone** -> **Mitigation:** keep the spec explicit about which values come from repository defaults and which override remains explicit (`FLEET_NAME`).
- **[Risk] Reverting only part of the original change could accidentally remove newer workflow improvements** -> **Mitigation:** implement the revert as a targeted edit against the current file, preserving later setup-node and diagnostics additions.

## Migration Plan

- Confirm the repository `.env` file and Makefile defaults still provide the values expected by the setup targets.
- Confirm the explicit `FLEET_NAME=fleet` step override remains in place for Fleet setup.
- Land the workflow update together with the delta spec so the documented contract matches the implementation immediately.
- If rollback is needed, restore the job-level defaults in the workflow and revert the delta spec to the previous self-contained behavior.

## Open Questions

- None.
