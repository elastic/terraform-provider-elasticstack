## Why

The Copilot setup workflow provisions an Elastic Stack, but when setup fails it does not emit Docker Compose logs to help diagnose the failure. The acceptance-test workflow already collects these logs on failure, so aligning Copilot setup with that behavior improves debugging without changing the successful path.

## What Changes

- Require the `copilot-setup-steps` workflow to collect Docker Compose logs when the job fails after stack startup, using the same `docker compose logs --no-color` pattern already used in `.github/workflows/test.yml`.
- Update the canonical OpenSpec capability for the Copilot setup workflow so failure-time stack log collection is part of the documented CI behavior.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `ci-copilot-setup-steps`: Extend workflow requirements so the job collects Elastic Stack Docker Compose logs when setup fails, making failure diagnostics available for Copilot setup runs that bootstrap the local stack.

## Impact

- **Workflow**: `.github/workflows/copilot-setup-steps.yml` — add a failure-path step to run `docker compose logs --no-color` after stack setup and before teardown/exit.
- **Specs**: `openspec/specs/ci-copilot-setup-steps/spec.md` — updated through a delta spec until the change is synced or archived.
- **CI diagnostics**: Failed Copilot setup runs will expose Elastic Stack container logs, reducing time to identify Elasticsearch, Kibana, Fleet, or bootstrap issues.
