# Task for worker

You are implementing ONE top-level task of an OpenSpec change in the Go Terraform provider repo at /Users/tobio/Projects/terraform-provider-elasticstack/security-entity-store-test-isolation-and-provider-waits.

Change name: security-entity-store-test-isolation-and-provider-waits. Tasks 1 and 2 are already done.

FIRST read fully before editing:
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/design.md (Decision 3)
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/tasks.md (Task 3)
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/specs/kibana-security-entity-store-entity-link/spec.md (REQ-ESL-RETRY-001)
- internal/asyncutils/state_waiter.go
- internal/kibana/security_entity_store_entity_link/create.go
- internal/kibana/security_entity_store/entity/write.go
- internal/clients/kibanaoapi/entity_store.go (CreateSecurityEntityStoreEntity, UpdateSecurityEntityStoreEntity — note they collapse the HTTP status into diagnostics via diagutil.HandleStatusResponse)
- internal/diagutil/http.go (HandleStatusResponse, ReportUnknownHTTPError)
- An existing asyncutils caller for reference, e.g. internal/fleet/integration/create.go (waitForFleetIntegrationInstalled).

SCOPE: Implement ONLY Task 3 (subtasks 3.1, 3.2, 3.3, 3.4).

REQUIREMENT SUMMARY (REQ-ESL-RETRY-001): On the create call, HTTP 500 → retry on next poll tick; HTTP 2xx → success (done); any other non-2xx (400/403/404/409/etc) → fail fast with error diagnostic. The retry loop is bounded ONLY by the Create ctx deadline (from the resource timeouts block) — do NOT add a separate wall-clock budget. Use asyncutils.WaitForStateTransition with asyncutils.WithPollInterval for cadence. If the deadline is reached with the last attempt still 500, return an error diagnostic describing the HTTP 500 (store may not be fully initialized). The StateChecker performs the create call and maps: 2xx → (true, nil); 500 → (false, nil); other non-2xx → (false, err).

Task 3.1 — entity-link create (internal/kibana/security_entity_store_entity_link/create.go):
This callsite already has direct access to resp.StatusCode() via client.GetKibanaOapiClient().API.PostSecurityEntityStoreResolutionLinkWithResponse. Wrap the POST in asyncutils.WaitForStateTransition. Build a StateChecker closure that:
  - performs the POST
  - on transport err (the `err` return): return (false, err) — fail fast (existing behavior for network error)
  - on resp.StatusCode() == 200 (http.StatusOK): capture success and return (true, nil)
  - on resp.StatusCode() == 500 (http.StatusInternalServerError): return (false, nil) to retry
  - on any other status: build the fail-fast diagnostics via diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body) and return (false, an error) so the wait aborts. To carry the diagnostic detail out through asyncutils (which only propagates `error`), wrap the status+body into a Go error (e.g. fmt.Errorf with the status code and body string), then after WaitForStateTransition returns, map: nil err → success; non-nil err → error diagnostic. Alternatively capture the fatal diag.Diagnostics in a closure variable and return a sentinel error to stop the loop, then surface the captured diagnostics. Prefer whichever is cleaner and preserves the original error/body detail. Also map the ctx-deadline error (the final still-500 case) to a clear error diagnostic. Preserve returning entitycore.KibanaWriteResult{Model: plan} on success.
Choose a sensible poll interval constant (e.g. 5 * time.Second) as a local const.

Task 3.2 — entity create (internal/kibana/security_entity_store/entity/write.go):
The create path calls kibanaoapi.CreateSecurityEntityStoreEntity which returns only diag.Diagnostics and hides the HTTP status. To implement retry-on-500 you must expose the status code. Add a new function in internal/clients/kibanaoapi/entity_store.go, e.g. `CreateSecurityEntityStoreEntityStatus(ctx, client, spaceID, entityType, body io.Reader) (int, []byte, error)` that returns the raw status code, response body, and transport error (do NOT collapse to diagnostics). Refactor the existing CreateSecurityEntityStoreEntity to call the new function and then apply diagutil.HandleStatusResponse (keep the existing public function working for any other callers — check for other callers with grep first). Then in write.go, wrap ONLY the create branch (req.Prior == nil) in asyncutils.WaitForStateTransition with the same 2xx/500/other mapping. Leave the UPDATE branch (req.Prior != nil) unchanged unless the update path also documents 500 retries — the tasks only mention the create (POST) path, so scope to create only. Note write.go re-reads bodyBytes via bytes.NewReader; since the body is a []byte you can create a fresh bytes.NewReader on each retry attempt inside the checker (do NOT reuse a consumed reader across attempts).

Task 3.3: Reuse internal/asyncutils only. Do NOT create a retryutil package.

Task 3.4: Add unit tests for the StateChecker closures in both packages:
  - 500 maps to retry (false, nil)
  - non-500 non-2xx maps to fail-fast (false, err) / captured error diagnostic
  - 2xx maps to done (true, nil)
  - a deadline-expired ctx surfaces an error diagnostic
Structure the checker construction so it is unit-testable without a live Kibana server — extract the checker builder to take a small function that returns (statusCode int, body []byte, err error), so tests can inject fake responses. Keep the public create functions working as specified.

CONVENTIONS: follow dev-docs/high-level/coding-standards.md; Apache license header on new files; add needed imports (net/http, time, fmt, bytes).

After editing:
- Run `go build ./...` (the kibanaoapi refactor touches a shared package, so build everything).
- Run `go test ./internal/kibana/security_entity_store_entity_link/... ./internal/kibana/security_entity_store/entity/... ./internal/clients/kibanaoapi/... 2>&1 | tail -40` (unit only, no TF_ACC).
- Do NOT run acceptance tests, push, or archive.
- Update tasks.md: mark 3.1, 3.2, 3.3, 3.4 as [x].
- Create a focused commit (e.g. "feat(security_entity_store): retry HTTP 500 on entity/entity-link create").

Report back: files changed, commit SHA, test summary, how you surfaced fail-fast diagnostics through asyncutils, other callers of CreateSecurityEntityStoreEntity you checked, any blockers.

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