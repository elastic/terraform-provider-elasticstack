## Context

The main CI workflow already has two layers of gating: `preflight` decides whether downstream CI should run at all, and `changes` decides whether acceptance coverage is required for the current diff. Today `test-validation` is declared with `if: always()`, so it still runs when `preflight` disables the rest of the workflow, and the shared validation helper converts that path into a passing result.

That behavior is useful for normal downstream-CI runs where `test-validation` should make a final pass/fail decision even if `changes` or `test` were skipped or failed. It is misleading on the pure preflight-skip path, because the check name implies tests passed even though the workflow intentionally never reached the acceptance-validation stage.

## Goals / Non-Goals

**Goals:**

- Skip `Test Validation` entirely when `preflight` outputs `should_run=false`.
- Preserve the existing final-decision behavior for runs where `preflight` allows downstream CI to proceed.
- Keep openspec-only change handling unchanged once downstream CI is allowed: `test` may be skipped and `Test Validation` may still succeed.
- Keep the authored workflow source, generated workflow, and validation helper contract aligned.

**Non-Goals:**

- Change the `preflight` decision logic itself.
- Change the `changes` classifier or what counts as `provider_changes=false`.
- Change matrix contents, Docker setup, snapshot-failure handling, or branch-protection policy outside the repository.

## Decisions

### 1. Gate the `test-validation` job on both `always()` and `preflightShouldRun`

The workflow will keep `always()` semantics for dependency evaluation, but it will only schedule `test-validation` when `needs.preflight.outputs.should_run == 'true'`.

Why:

- This preserves the existing ability to validate `changes` and `test` outcomes even when one of those jobs is skipped or fails during a real CI run.
- It avoids emitting a misleading green validation check when `preflight` intentionally disabled the entire downstream workflow.
- It is the smallest workflow change: the job graph stays the same, but the validation job now reflects the intended lifecycle more accurately.

Alternatives considered:

- Keep the job running and mark the preflight-skip path as neutral inside the script: rejected because the resulting check would still appear as a completed validation job rather than an intentionally skipped one.
- Remove `always()` entirely: rejected because `test-validation` must still run after failed or skipped upstream jobs once `preflight` has allowed downstream CI.

### 2. Remove the helper's preflight-success branch

Once the job itself no longer runs on the preflight-disabled path, the shared validation helper should no longer describe that state as a passing validation outcome.

Why:

- The helper contract should match the reachable workflow states so unit tests are meaningful.
- Leaving a dead "preflight skip equals pass" branch behind would preserve stale behavior in tests and future readers' mental models.

Alternatives considered:

- Keep the dead branch for defensive programming: rejected because the workflow already owns that gate, and retaining unreachable success logic makes the contract harder to reason about.

### 3. Keep `auto-approve` behavior unchanged for `ready_for_review`

`auto-approve` should continue to allow the `ready_for_review` path regardless of `test-validation` result, which now includes the validation job being skipped.

Why:

- The current workflow already treats `ready_for_review` as a special preflight-disabled path.
- Updating only the wording and expectations around that path keeps behavior stable while correcting the meaning of the validation check.

Alternatives considered:

- Make `ready_for_review` wait for `test-validation`: rejected because `preflight` intentionally suppresses the rest of CI on that event.

## Risks / Trade-offs

- **[Risk] A future workflow edit could remove the job-level preflight gate while tests still assume it exists** -> Mitigation: update the helper/unit tests to cover only reachable downstream-CI states and keep the OpenSpec delta explicit about the skip behavior.
- **[Risk] Consumers expecting a green `Test Validation` check on preflight-disabled runs will now see a skipped check instead** -> Mitigation: limit the behavior change to the preflight-disabled path and keep `auto-approve` explicitly documented as independent from validation success on `ready_for_review`.

## Migration Plan

1. Update the authored workflow template so `test-validation` requires `preflight.outputs.should_run == 'true'` in addition to `always()`.
2. Remove or rewrite the helper/test cases that currently encode "preflight skip means validation passed."
3. Regenerate `.github/workflows/test.yml` from the authored source.
4. Run workflow-source verification to confirm the generated workflow and helper tests remain in sync.

## Open Questions

- None.
