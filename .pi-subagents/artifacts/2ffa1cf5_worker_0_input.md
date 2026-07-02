# Task for worker

You are implementing ONE top-level task of an OpenSpec change in the Go Terraform provider repo at /Users/tobio/Projects/terraform-provider-elasticstack/security-entity-store-test-isolation-and-provider-waits.

Change name: security-entity-store-test-isolation-and-provider-waits

FIRST read these context files fully before editing:
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/proposal.md
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/design.md (Decision 1)
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/tasks.md (Task 1)
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/specs/kibana-security-entity-store/spec.md (REQ-WAIT-001)
- internal/asyncutils/state_waiter.go (the WaitForStateTransition primitive you must reuse)
- internal/kibana/security_entity_store/helpers.go (getEntityStoreStatus, entityStoreStatus)
- internal/kibana/security_entity_store/delete.go

SCOPE: Implement ONLY Task 1 (subtasks 1.1, 1.2, 1.3). Do not touch Read, Create, or tests in other packages.

Task 1.1: Add `waitForUninstall(ctx context.Context, client *clients.KibanaScopedClient, spaceID string) diag.Diagnostics` to internal/kibana/security_entity_store/helpers.go. Implement with `asyncutils.WaitForStateTransition(ctx, "security entity store", spaceID, checker, asyncutils.WithPollInterval(5*time.Second))`. The `checker` closure calls `getEntityStoreStatus(ctx, client, spaceID, false)` and returns `true` once `status.Status == kbapi.SecurityEntityAnalyticsAPIStoreStatusNotInstalled`. If getEntityStoreStatus returns error diagnostics inside the checker, treat as a transient error to retry — the design says "transient network errors on the status request SHALL be retried", so the checker should return (false, nil) on a status-read failure rather than aborting the wait (log at debug). Do NOT pass a timeout: the Delete ctx already carries the deadline from the resource timeouts block. Convert the returned error from WaitForStateTransition (ctx deadline exceeded) into a clear error diagnostic via diagutil.FrameworkDiagFromError or a diag.NewErrorDiagnostic with a message describing that uninstall did not complete within the Delete timeout. Return empty diag.Diagnostics on success.

Task 1.2: Call `waitForUninstall` from internal/kibana/security_entity_store/delete.go immediately after `kibanaoapi.UninstallSecurityEntityStore` succeeds (check its returned diags for errors first), before returning. Append its diagnostics to the return.

Task 1.3: Add a unit test in internal/kibana/security_entity_store/ (a new _test.go file, e.g. helpers_test.go, package security_entity_store) covering:
- the deadline path: a cancelled/expired ctx yields an error diagnostic (use context.WithCancel + cancel, or context.WithDeadline in the past)
- the checker/diagnostic mapping logic. Since waitForUninstall needs a live client, structure the test to exercise the StateChecker closure and the diagnostic mapping rather than re-testing asyncutils itself. If waitForUninstall's signature makes it hard to unit test without a client, refactor so the StateChecker construction and the error-to-diagnostic mapping are independently testable (e.g. extract a small helper), but keep the public waitForUninstall API as specified. Prefer testing the transition-to-not_installed logic of the checker and the deadline→error-diagnostic mapping.

CONVENTIONS:
- Follow dev-docs/high-level/coding-standards.md.
- Use the existing Apache license header (copy from delete.go) on any new file.
- Reuse asyncutils; do NOT create a new retryutil package.
- Add `time` import where needed.

After editing:
- Run `go build ./internal/kibana/security_entity_store/...` and `go test ./internal/kibana/security_entity_store/ -run 'Test.*Uninstall|Test.*Wait' 2>&1 | tail -30` (unit tests only, no TF_ACC).
- Do NOT run acceptance tests, do NOT push, do NOT archive the change.
- Update tasks.md: mark subtasks 1.1, 1.2, 1.3 as [x].
- Create a small focused git commit (e.g. "feat(security_entity_store): wait for uninstall completion in Delete").

Report back: files changed, commit SHA, test output summary, any blockers.

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