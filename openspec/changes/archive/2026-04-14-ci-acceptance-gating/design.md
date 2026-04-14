## Context

The `Build/Lint/Test` workflow in `.github/workflows/test.yml` currently exposes acceptance coverage through the `Matrix Acceptance Test` job only. GitHub branch protection can require those matrix checks, but workflow-level path filtering is not a viable escape hatch for non-provider changes because a skipped workflow leaves required checks pending instead of succeeding.

The workflow already has a `preflight` job that decides whether downstream jobs should execute, and `auto-approve` currently keys off `needs.test.result == 'success'`. That means an intentional acceptance-test skip for `openspec/**`-only changes needs a stable replacement status signal, not just a narrower `if:` on the matrix job.

## Goals / Non-Goals

**Goals:**

- Skip matrix acceptance tests when a workflow run changes only `openspec/**` paths in the first iteration.
- Keep the main workflow running so build, lint, and OpenSpec validation behavior remain available for spec-only changes.
- Publish a single stable validation job that can be used as the required acceptance-related branch-protection check.
- Preserve existing auto-approve behavior by keying it off the new validation outcome instead of direct matrix-test success.

**Non-Goals:**

- Generalize the first iteration to all non-provider paths outside `openspec/**`.
- Remove or relax the existing `build` or `lint` jobs.
- Change the acceptance-test matrix contents, snapshot failure policy, or Docker stack setup.
- Manage GitHub branch protection entirely from repository code; the workflow will provide the stable check, but maintainers still need to update the ruleset.

## Decisions

### 1. Add a dedicated change-classification job

The workflow will add a small `changes` job after `preflight` that determines whether provider acceptance tests are required for the current diff. In the first iteration, it will emit `provider_changes=false` only when every changed file is under `openspec/`; any other change set will emit `provider_changes=true`.

Why:

- It keeps the workflow itself active, avoiding the required-check deadlock caused by workflow-level path filters.
- It gives downstream jobs a deterministic, reusable signal instead of duplicating path logic across jobs.
- It preserves room to widen the non-provider path allowlist later without changing the overall workflow shape.

Alternatives considered:

- Expand `paths-ignore` on the workflow trigger: rejected because required checks stay pending when a workflow is skipped by path filters.
- Infer non-provider changes inside the matrix `test` job only: rejected because other jobs, including validation and auto-approve, also need the classification result.

### 2. Keep `build` and `lint` independent of the acceptance classifier

Only the matrix acceptance `test` job will be gated by `provider_changes`. `build` and `lint` will continue to follow the existing `preflight` behavior.

Why:

- `lint` already validates OpenSpec structure and remains useful for `openspec/**`-only changes.
- This keeps the first iteration narrow: save the expensive acceptance matrix without changing the broader CI contract.

Alternatives considered:

- Skip `build` and `lint` for spec-only changes too: rejected because that broadens the scope from acceptance gating into a larger CI redesign.

### 3. Add an always-running `Test Validation` job as the stable required check

The workflow will add a `test-validation` job that runs with `always()` after `preflight`, `changes`, and `test`. It will evaluate workflow outcomes and produce a single final result:

- pass when `preflight` intentionally disables CI execution
- pass when `provider_changes=false` and the matrix test job is skipped
- pass when `provider_changes=true` and the matrix test job succeeds
- fail when `provider_changes=true` and the matrix test job does not succeed

Why:

- GitHub branch protection works reliably with a required job that is present on every run.
- It collapses per-version matrix behavior into one stable check name for required-check configuration.
- It makes the skip policy explicit instead of forcing maintainers to reason about matrix-job absence or skipped states.

Alternatives considered:

- Require the matrix checks directly: rejected because the required-check set is unstable for intentionally skipped acceptance runs.
- Add a second stand-in workflow with inverse path filters: rejected because it recreates the same required-check edge cases and complicates status management across workflows.

### 4. Repoint auto-approve at validation instead of raw test success

`auto-approve` will depend on the new validation outcome, not `needs.test.result == 'success'`.

Why:

- A spec-only pull request should still be able to satisfy the current automation contract when acceptance tests were intentionally not needed.
- Validation is the new normalized signal for whether the test gate was satisfied.

Alternatives considered:

- Treat `needs.test.result == 'skipped'` as success directly in `auto-approve`: rejected because that spreads skip-policy logic into multiple places instead of using the dedicated validation job.

### 5. Treat branch-protection migration as an operational follow-up

The workflow will expose the stable `Test Validation` check, but repository maintainers must still update GitHub branch protection or rulesets to require that check instead of the matrix acceptance checks.

Why:

- Required-check configuration lives outside the repository.
- Splitting the workflow change from the ruleset update makes ownership clear and keeps the spec focused on repo-managed behavior while still documenting the operational step.

Alternatives considered:

- Leave branch protection unchanged: rejected because `openspec/**`-only pull requests would still be blocked by existing required matrix checks.

## Risks / Trade-offs

- **[Risk] The first-iteration classifier is intentionally narrow** -> Mitigation: scope the rule to `openspec/**` only for now, and extend it later through the same `changes` job if more non-provider paths should skip acceptance.
- **[Risk] Validation could hide other CI failures if misconfigured** -> Mitigation: keep `build` and `lint` as separate jobs and recommend retaining them as required checks if they are currently part of the merge gate.
- **[Risk] Ruleset updates may lag behind workflow rollout** -> Mitigation: document the required-check migration as part of the change and verify the new check name from a live run before removing the matrix checks from protection.

## Migration Plan

1. Add the workflow change-classification and validation logic in `.github/workflows/test.yml`.
2. Update the `ci-build-lint-test` OpenSpec delta so the canonical workflow requirements match the new job graph and validation semantics.
3. Trigger a workflow run that publishes the `Build/Lint/Test / Test Validation` check.
4. Update branch protection or rulesets to require `Test Validation` instead of the per-version `Matrix Acceptance Test (...)` checks.
5. Verify both an `openspec/**`-only change and a provider-impacting change against the new required-check behavior.

## Open Questions

- None for the first iteration.
