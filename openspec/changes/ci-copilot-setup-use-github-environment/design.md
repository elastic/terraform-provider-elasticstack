## Context

The `copilot-setup-steps` workflow currently embeds default Elastic Stack credentials in `jobs.copilot-setup-steps.env`, a behavior introduced by commit `6501f04bc251db7c69f581e3be2cdd20fb041b66`. The requested change is to remove those workflow-local defaults and return to the earlier model where the workflow consumes credentials supplied through GitHub repository environment settings. This proposal changes a single workflow and its corresponding OpenSpec capability, but it also changes the contract for manual and path-triggered runs because credentials are no longer self-contained in the YAML.

## Goals / Non-Goals

**Goals:**

- Remove the job-level credential defaults from `.github/workflows/copilot-setup-steps.yml`.
- Restore credential references so setup steps consume externally managed environment values rather than `ELASTIC_PASSWORD` defaults declared in the workflow.
- Update the `ci-copilot-setup-steps` OpenSpec capability to document GitHub-managed configuration as the expected source of Elastic Stack credentials.

**Non-Goals:**

- Changing unrelated workflow improvements added after `6501f04`, such as the Node.js setup step or failure-log collection behavior.
- Redesigning how the Makefile or Docker Compose files interpret credential environment variables.
- Introducing new secret names, new GitHub environments, or a broader CI credential-management convention outside this workflow.

## Decisions

1. **Revert only the credential-defaulting behavior from `6501f04`** — Remove `jobs.copilot-setup-steps.env` and restore step environment wiring to the pre-change variable names (`ELASTICSEARCH_PASSWORD`, `KIBANA_SYSTEM_USERNAME`, `KIBANA_SYSTEM_PASSWORD`) without discarding later unrelated workflow updates. **Alternative considered:** reverting the whole commit mechanically, which would also remove the canonical spec that has since been extended by later changes.
2. **Treat GitHub repository environment settings as the source of truth** — The workflow requirements will state that Elastic Stack credentials are supplied externally by GitHub-managed environment configuration rather than hard-coded workflow defaults, specifically for the variables consumed by the current workflow and Compose bootstrap: `ELASTICSEARCH_PASSWORD`, `KIBANA_PASSWORD`, `KIBANA_SYSTEM_USERNAME`, and `KIBANA_SYSTEM_PASSWORD`. **Alternative considered:** keeping YAML defaults with optional overrides, which preserves configuration duplication in version control.
3. **Document the changed manual-run contract explicitly** — The spec should no longer promise a self-contained `workflow_dispatch` run without repository-managed configuration, because that guarantee depends on the removed defaults. **Alternative considered:** leaving the older self-contained wording in place, which would misdescribe the post-change workflow behavior.

## Risks / Trade-offs

- **[Risk] Manual or validation runs can fail if the expected GitHub environment variables are not configured** -> **Mitigation:** document the external configuration requirement in the delta spec and implementation tasks.
- **[Risk] Required variable names become implicit and drift from the workflow** -> **Mitigation:** keep the workflow variable references unchanged from the pre-`6501f04` behavior and name them explicitly in the updated requirements.
- **[Risk] Reverting only part of the original change could accidentally remove newer workflow improvements** -> **Mitigation:** implement the revert as a targeted edit against the current file, preserving later setup-node and diagnostics additions.

## Migration Plan

- Ensure the required GitHub repository environment settings are present before merging the workflow change.
- Land the workflow update together with the delta spec so the documented contract matches the implementation immediately.
- If rollback is needed, restore the job-level defaults in the workflow and revert the delta spec to the previous self-contained behavior.

## Open Questions

- None.
