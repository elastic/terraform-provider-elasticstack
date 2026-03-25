## Why

The `copilot-setup-steps` workflow currently hard-codes default Elastic Stack credentials at job scope, which duplicates configuration that should live in GitHub repository environment settings. Reverting that behavior removes credential defaults from versioned workflow code and restores the pre-`6501f04` contract where the workflow consumes externally managed environment configuration.

## What Changes

- Remove the `jobs.copilot-setup-steps.env` default credential block from `.github/workflows/copilot-setup-steps.yml`.
- Update the documented workflow requirements so Copilot setup credentials are expected to come from GitHub repository environment settings instead of workflow-defined defaults.
- Revert the credential-wiring behavior introduced in commit `6501f04bc251db7c69f581e3be2cdd20fb041b66` while preserving later unrelated workflow updates.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `ci-copilot-setup-steps`: Change the workflow requirements so Elastic Stack and Kibana credentials are supplied by GitHub-managed environment configuration rather than job-level defaults embedded in the workflow file.

## Impact

- **Workflow**: `.github/workflows/copilot-setup-steps.yml` will stop declaring default credential values in `jobs.copilot-setup-steps.env`.
- **Specs**: `openspec/specs/ci-copilot-setup-steps/spec.md` will be updated through a delta spec to describe externally managed credential inputs instead of self-contained workflow defaults.
- **GitHub configuration**: Repository environment settings become the expected source of the variables needed by the Copilot setup steps.
