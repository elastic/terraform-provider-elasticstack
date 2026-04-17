## Context

The changelog-generation workflow currently has two categories of change in flight on this branch:

1. Trigger and activation hardening for release-preparation pull requests, including switching the release-mode trigger contract to `pull_request_target` while continuing to gate execution on `prep-release-*` head branches.
2. Reliability fixes around release evidence generation and transport, including helper-module extraction, supported `github-script` input wiring, and stronger validation around evidence JSON materialization.

The remaining reliability problem is the cross-job transport for the full evidence manifest. In observed workflow runs, the `pre_activation` job completed evidence gathering and set the `evidence_json` job output, but the downstream `agent` job saw an empty `EVIDENCE_JSON` environment variable and failed before agent reasoning started. The current transport contract treats the full manifest as a cross-job string output even though the actual consumer needs a file at `/tmp/gh-aw/agent/evidence.json`.

This design treats the manifest file itself as the durable contract between jobs and folds the already-implemented trigger changes into the same workflow stabilization effort.

## Goals / Non-Goals

**Goals:**
- Make changelog evidence handoff deterministic across jobs by promoting `evidence.json` to an artifact-backed file contract.
- Preserve the existing evidence manifest schema so downstream validation and agent instructions continue to work without semantic churn.
- Keep release-mode activation scoped to same-repository `prep-release-*` release-preparation branches, now triggered through `pull_request_target`.
- Preserve the current pre-activation scalar outputs that are small, stable, and needed for gating or prompt interpolation.
- Keep the authored workflow source, generated workflow artifacts, and workflow tests aligned with the new contract.

**Non-Goals:**
- Redesign the changelog evidence schema, PR classification heuristics, or changelog provenance format.
- Change the scheduled/manual singleton PR behavior for `generated-changelog`.
- Rework the agent prompt beyond what is necessary to reflect the artifact-based evidence handoff.
- Address unrelated GitHub Actions platform warnings such as Node 20 deprecation notices.

## Decisions

### 1. Use an artifact-backed `evidence.json` file as the only cross-job manifest transport

The `pre_activation` job will materialize the manifest to a file and upload it as an artifact. The `agent` job will download that artifact directly to the path the agent and helper commands already expect.

Rationale:
- The consumer contract is already file-based: `/tmp/gh-aw/agent/evidence.json`.
- Cross-job artifacts are a better fit than job outputs for large structured payloads that may be treated as secret-like by the runner.
- The file becomes the source of truth for debugging and reruns.

Alternatives considered:
- Keep `evidence_json` as a job output and continue writing the file in the `agent` job. Rejected because that is the transport currently exhibiting downstream empty-value failures.
- Split the manifest into multiple smaller outputs. Rejected because it complicates the contract and still relies on fragile output propagation.

### 2. Keep only scalar release-context values as job outputs

The workflow will continue to expose lightweight values such as `mode`, `previous_tag`, `compare_range`, `target_version`, `target_branch`, and `has_evidence` as outputs because they are used for gating and prompt interpolation and have not shown the same transport problem.

Rationale:
- These values are small and semantically distinct from the manifest payload.
- They keep job conditions and prompt rendering simple.

Alternatives considered:
- Move all release metadata into the artifact and remove scalar outputs. Rejected because it would make simple gating and prompt interpolation noisier without clear benefit.

### 3. Remove the redundant pre-agent manifest-writing bridge step

Once the `agent` job downloads `evidence.json` directly, the current `Write evidence manifest for agent` bridge step becomes unnecessary and should be removed along with its helper module unless it still provides value as a reusable validator.

Rationale:
- Eliminates an entire failure point.
- Makes the handoff shape match the actual consumption model.

Alternatives considered:
- Retain the bridge step but teach it to read a downloaded artifact file and rewrite it to the same location. Rejected because it adds indirection without improving reliability.

### 4. Keep supported `github-script` input patterns in deterministic steps

Scalar data passed into `github-script` steps should continue to use supported environment variables instead of custom `with:` inputs. This branch already moved `gather-pr-evidence` toward that pattern, and the stabilized workflow should preserve it.

Rationale:
- Avoids unsupported action inputs and associated warning noise.
- Keeps deterministic helper execution compatible with future action upgrades.

Alternatives considered:
- Reintroduce custom `with:` keys for convenience. Rejected because they are not supported by `actions/github-script`.

### 5. Update the trigger contract in the spec and workflow together

The change proposal will explicitly record that release-mode activation now uses `pull_request_target` rather than `pull_request`, while still requiring the `prep-release-*` head-branch contract before changelog work proceeds.

Rationale:
- This branch already changed the authored workflow trigger behavior.
- The canonical spec and delta spec should describe the implemented trust boundary and activation model.

Alternatives considered:
- Leave the spec on `pull_request` and treat `pull_request_target` as an implementation detail. Rejected because event type is part of the workflow’s externally observable execution contract and security posture.

## Risks / Trade-offs

- **Artifact upload/download introduces more workflow plumbing** → Keep the manifest artifact narrowly scoped to the changelog evidence file and test the upload/download contract in fixture coverage.
- **`pull_request_target` has different trust characteristics than `pull_request`** → Preserve the same-repository `prep-release-*` guard and document the trigger contract explicitly in the spec and proposal.
- **Removing the bridge step may require prompt/test rewiring** → Update workflow fixture tests and generated artifacts in the same change so the source of truth stays coherent.
- **Artifacts can go missing if upload or download steps are misconfigured** → Add deterministic upload/download steps before agent reasoning and validate the expected file path before invoking the agent.

## Migration Plan

1. Update the changelog workflow source to materialize and upload `evidence.json` in `pre_activation`.
2. Remove the full-manifest job output and replace the `agent`-job bridge with artifact download directly into `/tmp/gh-aw/agent/evidence.json`.
3. Retain scalar release-context outputs and existing prompt interpolation.
4. Regenerate `.github/workflows/changelog-generation.md` and `.github/workflows/changelog-generation.lock.yml`.
5. Update workflow fixture tests and helper-module tests to match the artifact-backed contract.
6. Validate the change with the existing workflow/unit test suite and an end-to-end workflow run.

Rollback:
- Revert the workflow source, generated artifacts, and helper/test changes to the prior output-based handoff if the artifact flow proves incompatible with GH AW execution.

## Open Questions

- Whether the bridge helper module should be deleted outright or retained as a smaller reusable validator/formatter for local tests.
- Whether the manifest artifact should be consumed via the standard GitHub artifact actions or via a GH AW-specific artifact mechanism if one emerges as a better long-term pattern.
