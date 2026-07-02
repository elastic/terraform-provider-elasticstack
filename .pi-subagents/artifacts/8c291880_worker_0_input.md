# Task for worker

You are implementing ONE top-level task of an OpenSpec change in the Go Terraform provider repo at /Users/tobio/Projects/terraform-provider-elasticstack/security-entity-store-test-isolation-and-provider-waits.

Change name: security-entity-store-test-isolation-and-provider-waits

Task 1 is already done (waitForUninstall added to helpers.go). Do NOT redo it.

FIRST read these context files fully before editing:
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/design.md (Decision 2)
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/tasks.md (Task 2)
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/specs/kibana-security-entity-store/spec.md (REQ-WAIT-002)
- internal/asyncutils/state_waiter.go
- internal/kibana/security_entity_store/helpers.go (getEntityStoreStatus, entityStoreStatus, waitForUninstall pattern to follow, entityStoreStatusFunc, makeUninstallStateChecker)
- internal/kibana/security_entity_store/read.go
- generated/kbapi — find the SecurityEntityAnalyticsAPIStoreStatus constants (grep for SecurityEntityAnalyticsAPIStoreStatus in generated/kbapi to learn the exact constant names for installing/running/stopped/error/not_installed).

SCOPE: Implement ONLY Task 2 (subtasks 2.1, 2.2, 2.3). Do not touch Delete, Create, or tests in other packages.

Task 2.1: Add `waitForStarted(ctx context.Context, client *clients.KibanaScopedClient, spaceID string) (*entityStoreStatus, []byte, diag.Diagnostics)` to internal/kibana/security_entity_store/helpers.go.
- Perform an initial synchronous getEntityStoreStatus(ctx, client, spaceID, false).
- If the initial read returns error diagnostics, return them.
- If overall status is NOT "installing", return the status, rawBody, and nil diags immediately (no polling).
- If status IS "installing", poll via asyncutils.WaitForStateTransition(ctx, "security entity store", spaceID, checker, asyncutils.WithPollInterval(3*time.Second)) with a StateChecker that re-reads status and returns true once status is no longer "installing" (i.e. running, stopped, error, or not_installed). The checker must capture the latest successfully-read status/rawBody so waitForStarted can return the last-observed data.
- On WaitForStateTransition error (ctx deadline exceeded while still installing): return the LAST-observed status and rawBody together with a WARNING diagnostic (diag.NewWarningDiagnostic) describing that the store is still installing and the read is proceeding with partial data — NOT an error diagnostic. This is the degraded-read path.
- Use the exact kbapi status constants. Check whether an "installing" constant exists (e.g. SecurityEntityAnalyticsAPIStoreStatusInstalling); use it.

Task 2.2: Replace the single getEntityStoreStatus call at the top of readEntityStore in internal/kibana/security_entity_store/read.go with waitForStarted. Preserve the existing not_installed → remove-from-state path (return model, false, nil). Preserve existing flatten logic. Because waitForStarted may return a warning diagnostic (not error), the read must still proceed to flatten/return when only warnings are present — append the warning diags to the returned diagnostics but continue. Only bail early when diags.HasError().

Task 2.3: Add unit tests (extend helpers_test.go or a new file, package security_entity_store) covering:
- installing→running transition path (checker returns false while installing, true when running)
- not_installed early-exit path (initial status not_installed returns immediately without polling)
- deadline-expiry → warning + degraded-read path (expired ctx yields a warning diagnostic and returns last-observed data, no error)
Structure tests to exercise the StateChecker closure and the warning-diagnostic mapping. If needed for testability, extract small helpers (like makeStartedStateChecker and a diags-from-error mapper) mirroring the Task 1 structure, but keep the waitForStarted public API as specified.

CONVENTIONS:
- Follow dev-docs/high-level/coding-standards.md.
- Apache license header on any new file (copy from read.go).
- Reuse asyncutils; no new packages.

After editing:
- Run `go build ./internal/kibana/security_entity_store/...` and `go test ./internal/kibana/security_entity_store/ -run 'Test.*Start|Test.*Wait|Test.*Read' 2>&1 | tail -30` (unit only, no TF_ACC).
- Do NOT run acceptance tests, push, or archive.
- Update tasks.md: mark 2.1, 2.2, 2.3 as [x].
- Create a focused commit (e.g. "feat(security_entity_store): wait for started-state in Read").

Report back: files changed, commit SHA, test output summary, the exact kbapi installing constant name you used, any blockers.

## Acceptance Contract
Acceptance level: reviewed
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Implement the requested change without widening scope
- criterion-2: Return evidence sufficient for an independent acceptance review

Required evidence: changed-files, tests-added, commands-run, validation-output, residual-risks, no-staged-files

Review gate: required by reviewer.

Finish with a fenced JSON block tagged `acceptance-report` in this shape:
Use empty arrays when no items apply; array fields contain strings unless object entries are shown.
```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "specific proof"
    }
  ],
  "changedFiles": [
    "src/file.ts"
  ],
  "testsAddedOrUpdated": [
    "test/file.test.ts"
  ],
  "commandsRun": [
    {
      "command": "command",
      "result": "passed",
      "summary": "short result"
    }
  ],
  "validationOutput": [
    "validation output or concise summary"
  ],
  "residualRisks": [
    "none"
  ],
  "noStagedFiles": true,
  "diffSummary": "short description of the diff",
  "reviewFindings": [
    "blocker: file.ts:12 - issue found, or no blockers"
  ],
  "manualNotes": "anything else the parent should know"
}
```