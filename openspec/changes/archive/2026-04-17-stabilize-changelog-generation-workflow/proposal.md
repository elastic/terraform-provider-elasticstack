## Why

The changelog-generation workflow has accumulated several reliability and trigger-contract changes on this branch, but its current evidence handoff still depends on passing a large serialized manifest through cross-job outputs. In observed runs, the pre-activation job successfully built the evidence manifest while the downstream agent job received an empty `EVIDENCE_JSON`, so the workflow needs a more durable cross-job transport and a spec update that reflects the trigger and wiring changes already made.

## What Changes

- Change changelog release-mode activation from `pull_request` to `pull_request_target` while preserving the existing `prep-release-*` head-branch guard for release-preparation PRs.
- Replace the cross-job `evidence_json` handoff with an artifact-backed `evidence.json` file produced in pre-activation and consumed directly by the agent job.
- Keep lightweight pre-activation outputs only for scalar release context and gating metadata; stop treating the full evidence manifest as a job output contract.
- Align the authored workflow, generated workflow artifacts, and workflow fixture tests with the new trigger mode and artifact-based evidence transport.
- Preserve the existing deterministic evidence gathering, manifest shape, PR classification, and proof-carrying changelog generation behavior while hardening the transport between jobs.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ci-changelog-generation`: update the workflow trigger contract to use `pull_request_target` for release-preparation PRs and change the evidence-manifest handoff from cross-job JSON outputs to an artifact-backed file contract.

## Impact

- Affected workflow sources and generated artifacts under `.github/workflows-src/changelog-generation/` and `.github/workflows/`.
- Affected helper modules and tests for changelog evidence gathering and manifest materialization.
- Changes the cross-job transport contract for changelog evidence from job outputs/env interpolation to artifact upload/download.
