## Context

The `Build/Lint/Test` workflow (`test.yml`) is compiled from a single template and handles every CI scenario — feature branch pushes, Copilot bot pushes, draft PRs, `ready_for_review` transitions, changelog-only bot PRs, and post-merge CI on `main` — through a 100-line `preflight` job with six conditional branches. This approach has produced three active bugs:

1. `ready_for_review` fires → `preflight` sets `should_run=false` → `build` and `lint` are skipped → GitHub records a new "skipped" check run for each, overwriting the prior passing run → PR blocked.
2. Push to a PR branch fires both `push` and `pull_request:synchronize` → `push` preflight skips CI but still emits a successful `test-validation` check run → race: if the push check run is recorded after a failing synchronize check run, it overwrites the failure → PR appears mergeable.
3. `classifyChanges` only recognises `openspec/` as a skip-worthy path → changelog-only PRs either run 48+ acceptance-test jobs (regular PRs) or hit the preflight shortcut that skips `build`/`lint` required checks (generated-changelog bot PRs).

## Goals / Non-Goals

**Goals:**
- Required checks are always present and accurate for every PR type, including changelog-only, openspec-only, draft-to-ready, and bot PRs.
- No race conditions between push and pull_request check runs.
- Acceptance tests are skipped when the change set cannot affect provider behaviour.
- A single required check per workflow, not one per intermediate job.
- `auto-approve` continues to work for dependabot and copilot PRs.

**Non-Goals:**
- CI for branches without open PRs on every push (available via `workflow_dispatch`).
- Preserving the generated-changelog auto-merge path (removed; out of scope for this change).
- Changing the acceptance test matrix, stack versions, or test sharding.
- Changing how `make check-lint` or `make check-openspec` work internally.

## Decisions

### Split into three workflows

**Decision**: Replace `test.yml` with `provider.yml`, `openspec.yml`, and `workflows.yml`.

**Rationale**: A single gate job per workflow is the only reliable pattern for required checks when some jobs are conditionally skipped. With a monolithic workflow, the gate must absorb the logic for every domain (Go CI, openspec validation, workflow tests). Splitting by domain makes each workflow's gate simple and self-contained. It also means openspec validation (fast, no stack) no longer queues behind acceptance tests (35 min), and workflow tests no longer share runner budget with the matrix.

**Alternative considered**: Keep a single workflow with a multi-domain gate. Rejected — the gate would still need to reason about `should_run` (push vs. PR), `provider_changes`, and `workflow_changes` simultaneously, which is the same complexity that created the current bugs.

### Remove `push` on PR branches; keep `push: [main]` only

**Decision**: `push` triggers only on `main`. Use `workflow_dispatch` for ad-hoc CI on branches without PRs.

**Rationale**: The race condition (Bug 2) is caused by the same workflow running for both `push` and `pull_request:synchronize` on the same commit. The only way to eliminate it cleanly is to stop the workflow from running on PR-branch pushes. Copilot commits to a PR branch trigger `pull_request:synchronize` — they do not require a `push` trigger. Post-merge validation on `main` remains valuable; keeping `push: [main]` provides that without the race.

**Alternative considered**: Run push only for bot emails (current approach). Rejected — this still creates two competing check runs for bot commits, and the allowlist maintenance burden grows as bots are added.

### Single `gate` job per workflow as the required check

**Decision**: Each workflow has one `gate` job (`if: always()`) that is the only required check. Intermediate jobs (`build`, `lint`, `test`) are never required checks.

**Rationale**: GitHub evaluates required checks by looking at the most recent check run for each required check name per commit. When an intermediate job is skipped (e.g., `build` on a changelog-only PR), GitHub records a "skipped" conclusion, which does not satisfy a required check. A `gate` job that explicitly passes when skipping is justified solves this permanently — the gate always emits a concrete `success` or `failure`.

**Alternative considered**: Configure branch protection to allow skipped checks. Rejected — this would allow any check to be skipped for any reason, including unexpected failures, with no gate to catch it.

### Always run everything on `push: [main]`; classify only for `pull_request`

**Decision**: On `push` to `main` and on `workflow_dispatch`, skip classify and default to `provider_changes=true` / `workflow_changes=true`. Run all jobs unconditionally.

**Rationale**: Post-merge CI on `main` should be comprehensive regardless of what files changed. The purpose of classify is to save CI time on clearly non-impactful PRs — that optimisation is irrelevant after a merge. Defaulting to `true` on push/dispatch makes the gate logic uniform: if classify ran, trust its output; if it didn't (push/dispatch), treat as fully impactful.

### Remove `preflight` entirely

**Decision**: No `preflight` job in any of the three workflows.

**Rationale**: `preflight` existed to (a) suppress duplicate runs when both `push` and `pull_request` fire for the same commit, and (b) short-circuit the generated-changelog bot PR path. With `push` limited to `main` and `ready_for_review` removed from PR types, (a) is no longer needed. With the generated-changelog path removed, (b) is no longer needed. The job has no remaining purpose.

### Extend classify skip-paths beyond `openspec/`

**Decision**: `provider_changes=false` when all changed files match: `CHANGELOG.md`, `openspec/**`, `.agents/**`, or `.github/**` except `.github/workflows/provider.yml`.

**Rationale**: Changes to changelog, agent prompts, or non-CI `.github/` files (issue templates, dependabot config) cannot affect the provider binary or its test behaviour. The current openspec-only rule is too narrow and causes unnecessary acceptance-test runs on changelog and `.agents/` PRs. Changes to `.github/workflows/provider.yml` itself require a full CI run to validate the new workflow against the provider.

## Risks / Trade-offs

**[Risk] Transition window: old required checks still configured while new workflows deploy.**
→ Mitigation: deploy all three new workflow files in a single PR; update branch protection in the same session immediately after the PR merges to `main`. Keep old `test.yml` active (do not delete) until branch protection is updated.

**[Risk] `workflow_dispatch` is a worse DX than automatic push CI for branches without PRs.**
→ Mitigation: acceptable trade-off; developers can trigger manually. Document this in the contributing guide.

**[Risk] The `push: [main]` run doubles CI cost for every PR merge (a push to main fires after each merge).**
→ Accepted trade-off: post-merge CI catches integration issues that PR CI may miss. The matrix is unchanged, so cost is the same as current main-branch push CI.

**[Risk] `workflows.yml` classify may trigger workflow tests on any `.github/**` change, including issue templates.**
→ Mitigation: workflow tests are fast (no stack); triggering them on non-workflow `.github/` changes is a minor over-trigger, not a blocking problem.

## Migration Plan

1. Implement all source template and script changes on a feature branch.
2. Merge to `main` — the three new workflow files are created; `test.yml` remains active for the existing required checks.
3. Immediately after merge, update branch protection:
   - Remove required checks: `Build/Lint/Test / Build`, `Build/Lint/Test / Lint`, `Build/Lint/Test / Test Validation`
   - Add required checks: `provider / gate`, `openspec / gate`, `workflows / gate`
4. Delete `test.yml` and the `test/` workflow source directory in a follow-up cleanup PR.

**Rollback**: Restore the `test.yml` source template from git and revert branch protection changes.

## Open Questions

None — all design decisions were resolved during exploration.
