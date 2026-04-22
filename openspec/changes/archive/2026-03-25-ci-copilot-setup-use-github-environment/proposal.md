## Why

The `copilot-setup-steps` workflow no longer needs hard-coded credential defaults at job scope. Instead, it relies on the existing repository `.env` file and Makefile defaults for the bootstrap and authentication values used by the setup targets, while keeping `FLEET_NAME` explicitly set on the Fleet step.

## What Changes

- Remove the `jobs.copilot-setup-steps.env` default credential block from `.github/workflows/copilot-setup-steps.yml`.
- Keep the setup steps free of workflow-level or job-level credential wiring, except for the explicit `FLEET_NAME` override on the Fleet step.
- Update the documented workflow requirements so bootstrap and authentication values are described as coming from the existing repository `.env` file and Makefile defaults unless explicitly overridden.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `ci-copilot-setup-steps`: Change the workflow requirements so bootstrap and authentication values rely on existing repository defaults instead of job-level workflow defaults, with `FLEET_NAME` remaining explicitly set on the Fleet step.

## Impact

- **Workflow**: `.github/workflows/copilot-setup-steps.yml` will stop declaring default credential values in `jobs.copilot-setup-steps.env`.
- **Workflow**: The setup steps will rely on inherited repository defaults, with only `FLEET_NAME` set explicitly for the Fleet setup step.
- **Specs**: `openspec/specs/ci-copilot-setup-steps/spec.md` will be updated through a delta spec to describe the current reliance on repository `.env` and Makefile defaults.
- **Configuration**: The repository `.env` file and Makefile defaults continue to supply bootstrap and authentication defaults unless the execution environment overrides them.
