## Why

Kibana OpenAPI and generated client changes can land through Renovate or direct updates without any systematic follow-up on which Terraform resources or data sources may need work. Maintainers need an automated way to detect likely impacted entities and open focused issues before implementation planning begins.

## What Changes

- Add a GitHub agentic workflow that reacts to Kibana spec or generated client changes on `main` and on manual or scheduled runs.
- Add deterministic helper tooling that computes a baseline, diffs Kibana client artifacts, and maps changed API symbols to registered Kibana Terraform entities.
- Add repo memory for processed baselines and duplicate suppression so the workflow does not repeatedly open equivalent issues.
- Add agent instructions that turn deterministic impact evidence into one actionable GitHub issue per impacted entity.

## Capabilities

### New Capabilities
- `ci-kibana-spec-impact-issues`: Detect Kibana spec or generated client changes, determine impacted Terraform entities, and open one issue per impacted entity with a summarized impact report.

### Modified Capabilities

## Impact

- GitHub workflow sources and generated workflow outputs under `.github/workflows-src/` and `.github/workflows/`.
- New helper tooling under `scripts/` for baseline tracking, symbol diffing, and entity impact mapping.
- Repo memory under `.github/aw/memory/` for dedupe and processed baseline state.
- Kibana provider packages under `provider/`, `internal/clients/kibanaoapi/`, and `internal/kibana/` as discovery inputs for deterministic impact mapping.
