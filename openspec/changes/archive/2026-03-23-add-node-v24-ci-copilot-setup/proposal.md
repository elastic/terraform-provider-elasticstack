## Why

The repository’s OpenSpec CLI and `package.json` `engines` define the required Node range, but the Copilot setup workflow does not install Node. Without a matching runtime, Copilot cannot reliably run OpenSpec (`openspec`, `npm ci`) during sessions aligned with local and CI lint expectations.

## What Changes

- Require the `copilot-setup-steps` workflow to install Node using `actions/setup-node` (pinned by commit SHA) with **`node-version-file: package.json`**, so the version is resolved from `package.json` (`volta.node`, `devEngines.runtime` for node, or `engines.node`, per the action’s documented precedence) instead of duplicating a major in YAML. Keep npm caching aligned to `package-lock.json`.
- Update the canonical OpenSpec capability spec for this workflow so toolchain requirements explicitly include Node for OpenSpec-driven workflows, described in terms of that file-driven resolution rather than a hardcoded workflow string.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `ci-copilot-setup-steps`: Extend toolchain requirements (currently checkout, Go, Terraform) to include Node setup via `node-version-file: package.json` so the agent environment matches repository `engines` (and related fields) and can run OpenSpec and npm-based tooling without duplicating the version in the workflow.

## Impact

- **Workflow**: `.github/workflows/copilot-setup-steps.yml` — add a `setup-node` step (order: after checkout, alongside or before Go/Terraform as specified in design/tasks).
- **Specs**: `openspec/specs/ci-copilot-setup-steps/spec.md` — updated via delta spec until synced/archived.
- **Copilot**: Setup steps expose `node` / `npm` satisfying the version spec read from `package.json`, consistent with `make setup` / OpenSpec usage.
