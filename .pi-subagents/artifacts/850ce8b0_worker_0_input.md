# Task for worker

Implement ONE task of an OpenSpec change in the Go Terraform provider at /Users/tobio/Projects/terraform-provider-elasticstack/security-entity-store-test-isolation-and-provider-waits. Change: security-entity-store-test-isolation-and-provider-waits. Tasks 1-3 already done.

IMPORTANT — avoid context overflow: do NOT read files under generated/kbapi (they are huge). Do NOT read node_modules. Only read the specific files named below.

Read these files (only these):
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/design.md (Decision 4)
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/tasks.md (Task 4)
- openspec/changes/security-entity-store-test-isolation-and-provider-waits/specs/kibana-security-entity-store/spec.md (REQ-TEST-ISOLATION-001)
- internal/acctest/security_helpers.go (shows the non-_test.go helper pattern, license header, NewAcceptanceTesting* usage)
- internal/asyncutils/state_waiter.go
- internal/kibana/security_entity_store/acc_test.go
- internal/kibana/security_entity_store/entity/acc_test.go
- internal/kibana/security_entity_store_entity_link/acc_test.go
- internal/kibana/security_entity_store_resolution_group/acc_test.go

SCOPE: Task 4 only (4.1, 4.2, 4.3, 4.4). This is TEST-ONLY. Do NOT touch provider (non-test) code, do NOT touch Tasks 5/6.

Key facts you can rely on WITHOUT reading kbapi:
- `clients.NewAcceptanceTestingKibanaScopedClient() (*clients.KibanaScopedClient, error)` builds a Kibana client (see internal/clients/kibana_scoped_client.go:162).
- A `*clients.KibanaScopedClient` exposes `.GetKibanaOapiClient()` returning a `*kibanaoapi.Client`.
- `kibanaoapi.UninstallSecurityEntityStore(ctx, client, spaceID, body kbapi.PostSecurityEntityStoreUninstallJSONRequestBody) diag.Diagnostics` uninstalls. Pass an empty body `kbapi.PostSecurityEntityStoreUninstallJSONRequestBody{}` to uninstall all types.
- `kibanaoapi.GetSecurityEntityStoreStatus(ctx, client, spaceID, includeComponents bool) (*kbapi.GetSecurityEntityStoreStatusResponse, diag.Diagnostics)` reads status. The response has `.Body []byte`. To detect not_installed WITHOUT importing heavy kbapi types, json-unmarshal resp.Body into a small local struct `struct{ Status string `json:"status"` }` and compare `== "not_installed"`.
- `asyncutils.WaitForStateTransition(ctx, resourceType, resourceID string, checker asyncutils.StateChecker, opts ...asyncutils.Option) error` with `asyncutils.WithPollInterval(5*time.Second)`.

Task 4.1: Add a shared function to internal/acctest/security_helpers.go (NOT a _test.go file, so it is reusable across packages):
```
func CleanupEntityStore(t *testing.T, spaceID string)
```
(Exported, since it will be called from other packages' _test.go files. The tasks.md calls it cleanupEntityStore but it MUST be exported to cross package boundaries — name it CleanupEntityStore.)
Behavior:
- t.Helper(); SkipIfNotAcceptanceTest(t) at the top (mirror CreateESAccessToken).
- Build client via clients.NewAcceptanceTestingKibanaScopedClient(); on error t.Logf and return (cleanup should not fail the test on client construction issues — but you may t.Fatalf if you judge that better; prefer t.Logf + return so cleanup is best-effort).
- Create a `ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute); defer cancel()`.
- Call UninstallSecurityEntityStore with empty body. If it returns error diagnostics, t.Logf a warning but CONTINUE to the wait (idempotency: uninstalling an already-uninstalled store may error; do not fail).
- Wait via asyncutils.WaitForStateTransition with WithPollInterval(5*time.Second) and a StateChecker that calls GetSecurityEntityStoreStatus, json-unmarshals Body into the small status struct, and returns true once status == "not_installed". On status-read error inside the checker, t.Logf debug and return (false, nil) to retry (transient).
- t.Log progress messages so CI logs show what cleanup is doing (e.g. "cleaning up entity store in space default", "entity store reached not_installed").
- On WaitForStateTransition returning an error (deadline), t.Logf a warning (do NOT t.Fatal — cleanup is best-effort).

Task 4.4 (idempotency): ensure calling when already not_installed is a no-op success — the checker returning true immediately on not_installed, and tolerating uninstall errors, achieves this.

Task 4.2: At the very top (first statement) of EVERY acceptance test function in internal/kibana/security_entity_store/acc_test.go, register `t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })`. Add for ALL Test funcs in that file (there are 9: _basic, _singleType, _updateLogExtraction, _import, _shrinkGuardFails, _shrinkWithFlag, _startedFalse, _historySnapshot, and TestAccDataSourceKibanaSecurityEntityStoreStatus_basic). Check the acctest package is imported (it almost certainly already is — verify the import path github.com/elastic/terraform-provider-elasticstack/internal/acctest).

Task 4.3: Same registration at the top of every Test func in:
- internal/kibana/security_entity_store/entity/acc_test.go (all TestAccResourceKibanaSecurityEntityStoreEntity_* funcs)
- internal/kibana/security_entity_store_entity_link/acc_test.go (TestAccResourceSecurityEntityStoreEntityLink, TestAccResourceSecurityEntityStoreEntityLink_SingleElement, and any other Test funcs present)
- internal/kibana/security_entity_store_resolution_group/acc_test.go (TestAccDataSourceSecurityEntityStoreResolutionGroup and any other Test funcs)
Ensure each of these files imports the acctest package (add the import if missing).

Place the t.Cleanup registration AFTER any existing `acctest.SkipIfNotAcceptanceTest(t)` / resource.Test preamble is fine, but ideally as one of the first statements so cleanup is guaranteed even if the test body panics. If a test already calls SkipIfNotAcceptanceTest first, put the t.Cleanup immediately after it.

CONVENTIONS: follow dev-docs/high-level/coding-standards.md; Apache license header already present in files you edit; do not reformat unrelated code.

After editing:
- Run `go build ./...` and `go vet ./internal/acctest/... ./internal/kibana/security_entity_store/... ./internal/kibana/security_entity_store_entity_link/... ./internal/kibana/security_entity_store_resolution_group/... 2>&1 | tail -30`. Fix compile errors. Do NOT run acceptance tests (TF_ACC) — no stack needed for compile/vet.
- Run gofmt -l on the files you changed; fix any formatting.
- Update tasks.md: mark 4.1, 4.2, 4.3, 4.4 as [x].
- Commit: "test(security_entity_store): add shared entity store cleanup helper and register in acc tests".
- Do NOT push or archive.

Report back: files changed, the exact exported helper name, commit SHA, build/vet result, list of test funcs you added t.Cleanup to per file, any blockers.

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