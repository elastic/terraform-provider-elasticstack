## Context

The `kibana-spec-impact` GH AW workflow is currently split across three incompatible execution contexts:

- pre-activation runs an inline `actions/github-script` wrapper that shells out to the Go helper, computes gate outputs, and writes the report into the repository workspace;
- repo memory is configured for the workflow, but pre-activation does not have the checkout and initialization needed to rely on that memory path by default;
- the agent job expects the deterministic report to still exist as a repo-root file even though job boundaries only preserve explicit artifacts.

That split makes the workflow fragile. The deterministic report and gate contract are partly in JavaScript and partly in Go, pre-activation cannot reliably consume persisted repo memory, and the agent instructions describe file paths that are not durable between jobs.

## Goals / Non-Goals

**Goals:**

- Make the deterministic pre-activation flow self-contained around repository-local Go helper logic.
- Ensure pre-activation has the repository checkout and repo-memory context it needs before computing Kibana spec impact.
- Preserve the impact report across job boundaries using an explicit workflow artifact downloaded into `/tmp/gh-aw/agent`.
- Align the agent instructions with the actual runtime paths used in the agent job.
- Keep the authored template and compiled workflow artifacts in sync.

**Non-Goals:**

- Changing the matching heuristics for `high_confidence_impacts`, transform hints, or duplicate suppression semantics.
- Changing the issue body policy, issue cap, or repo-memory branch layout beyond what is needed for correct workflow execution.
- Replacing GH AW with a plain GitHub Actions workflow.

## Decisions

### 1. Move pre-activation orchestration into the Go helper

The Go helper under `scripts/kibana-spec-impact/` will own the deterministic pre-activation flow: memory bootstrap if needed, impact report generation, gate derivation, and report-file writing.

Rationale:

- It removes the current split where JavaScript computes gate outputs after parsing Go output.
- It keeps the workflow contract in repository-local code that is easier to test with existing Go tests.
- It reduces template complexity and avoids embedding brittle orchestration logic in `actions/github-script`.

Alternatives considered:

- Keep the current inline script and only change paths. Rejected because it preserves the current duplication and leaves gate logic split across languages.
- Move all computation into shell steps. Rejected because structured JSON and gate outputs are better handled in the existing Go helper surface.

### 2. Initialize checkout and repo memory during pre-activation

`on.steps` will explicitly check out the repository before Go setup and use a custom pre-activation sequence that makes repo memory available before the compute step runs. The workflow will declare the repo-memory `branch-name` explicitly and use a dedicated checkout/init step that targets that same branch for pre-activation setup.

Rationale:

- Pre-activation is where the workflow chooses the baseline and emits the deterministic report, so it must have the same memory context the agent later persists.
- The current agent-only memory initialization is too late for report generation and gating.
- Making the branch explicit avoids relying on GH AW defaults and ensures the pre-activation bootstrap step and the `repo-memory` tool definition stay aligned.

Alternatives considered:

- Continue using fallback temporary memory during pre-activation. Rejected because it breaks continuity with the persisted workflow baseline.
- Recompute or repair memory state in the agent job only. Rejected because the agent should consume precomputed deterministic evidence, not rebuild it.
- Rely on the implicit default repo-memory branch. Rejected because the design now depends on a separate initialization step, so the branch contract should be explicit in frontmatter and in the checkout/init step.

### 3. Use an artifact handoff for the deterministic report

Pre-activation will upload the report as a named GitHub Actions artifact, and the agent job will download it into `/tmp/gh-aw/agent`.

Rationale:

- Artifacts are the durable workflow primitive for passing files between jobs.
- This matches an existing repository pattern already used by the changelog workflow.
- It lets the prompt reference a stable agent-local path rather than an implicit workspace carry-over.

Alternatives considered:

- Keep writing to the repo root. Rejected because job-local workspaces do not survive automatically.
- Encode the full report into job outputs. Rejected because the report is structured file content and can grow beyond what is comfortable to manage as outputs.

### 4. Make agent instructions path-accurate

The prompt will describe the report and issued-file locations under `/tmp/gh-aw/agent`, while memory persistence will continue to target `/tmp/gh-aw/repo-memory/kibana-spec-impact/...`.

Rationale:

- The agent instructions should describe the real filesystem contract of the generated workflow.
- Separating ephemeral agent artifacts from durable repo memory keeps the workflow model clear.

## Risks / Trade-offs

- [Workflow generation drift] -> Regenerate both the source-derived markdown and the compiled lockfile as part of the change and verify them with workflow-specific checks.
- [Helper interface churn] -> Keep the new Go helper surface narrow and reuse existing report and memory code paths rather than rewriting them.
- [Artifact/path mismatch] -> Update both workflow orchestration and prompt text in the same change so the generated workflow and agent instructions stay consistent.

## Migration Plan

1. Extend the Go helper to support the pre-activation gate-and-report contract.
2. Update the workflow source template to add an explicit repo-memory `branch-name`, pre-activation checkout/memory setup for that branch, and artifact upload/download wiring.
3. Rewrite prompt references to the agent artifact paths.
4. Regenerate the committed workflow artifacts.
5. Run focused workflow checks before merge.

Rollback is straightforward: revert the change to restore the previous inline-script and workspace-file behavior if the generated workflow fails unexpectedly.

## Open Questions

- Whether the Go helper should expose the new pre-activation contract as a dedicated subcommand or as an extension of the existing `report` command can be finalized during implementation, as long as the deterministic outputs remain testable and the workflow no longer depends on inline JavaScript gate logic.
