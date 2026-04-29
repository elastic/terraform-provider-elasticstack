## Why

The 7.17 acceptance test matrix job has started stalling during Docker Compose startup because the `elastic/elastic-agent:7.17.13` image (pulled from Docker Hub) hangs indefinitely. The job exceeds its 35-minute timeout and is cancelled, wasting a runner slot and blocking CI feedback. Successful pulls for the same image consistently complete in under 30 seconds.

## What Changes

- Add a **pre-pull step** (before `make docker-fleet`) that pulls the Fleet image with a per-attempt timeout and automatic retry:
  - 90-second timeout per pull attempt
  - Up to 3 retries with 30-second backoff
- Add a **10-minute `timeout-minutes`** to the "Start stack with docker compose" step so that a hung pull fails fast instead of consuming the full job timeout.

## Capabilities

### New Capabilities
_(none — this is a CI reliability change with no Terraform provider behavior changes)_

### Modified Capabilities
_(none — no existing specs are affected)_

## Impact

- `.github/workflows-src/test/workflow.yml.tmpl` — source template for the test workflow
- `.github/workflows/test.yml` — generated workflow file (regenerated via `make workflow-generate`)
