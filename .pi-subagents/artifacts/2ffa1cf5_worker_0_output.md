 Implemented Task 1 (Delete wait for uninstall).

Changed files:
- `internal/kibana/security_entity_store/helpers.go` — added `waitForUninstall`, `makeUninstallStateChecker`, `uninstallWaitDiagsFromError`
- `internal/kibana/security_entity_store/delete.go` — call `waitForUninstall` after `UninstallSecurityEntityStore` succeeds
- `internal/kibana/security_entity_store/helpers_test.go` — new unit tests for deadline, diagnostic mapping, and checker states
- `openspec/changes/security-entity-store-test-isolation-and-provider-waits/tasks.md` — marked subtasks 1.1–1.3 as done

Commit SHA: `e406ecd5d`

Validation:
- `go build ./internal/kibana/security_entity_store/...` — passed
- `go test ./internal/kibana/security_entity_store/ -run 'Test.*Uninstall|Test.*Wait'` — all passed

Open risks/questions: None.