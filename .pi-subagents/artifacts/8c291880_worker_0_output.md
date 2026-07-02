Implemented Task 2 (2.1, 2.2, 2.3): added `waitForStarted` in `helpers.go`, wired it into `readEntityStore`, and added unit tests.

Changed files:
- `internal/kibana/security_entity_store/helpers.go`
- `internal/kibana/security_entity_store/helpers_test.go`
- `internal/kibana/security_entity_store/read.go`
- `openspec/changes/security-entity-store-test-isolation-and-provider-waits/tasks.md`

Commit SHA: `d1d246f11` (`feat(security_entity_store): wait for started-state in Read`)

Validation:
- `go build ./internal/kibana/security_entity_store/...` — passed
- `go vet ./internal/kibana/security_entity_store/...` — passed
- `go test ./internal/kibana/security_entity_store/ -run 'Test.*Start|Test.*Wait|Test.*Read' -count=1` — passed
- full package unit tests (`go test ./internal/kibana/security_entity_store/ -count=1`) — passed

Exact kbapi installing constant used: `kbapi.SecurityEntityAnalyticsAPIStoreStatusInstalling`.

No blockers. No staged files remain.