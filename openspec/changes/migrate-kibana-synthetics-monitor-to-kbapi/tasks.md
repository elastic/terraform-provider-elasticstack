## 1. Discovery and kbapi surface

- [ ] 1.1 Inventory generated operations and types for `PostSyntheticMonitors`, `GetSyntheticMonitor`, `PutSyntheticMonitor`, `DeleteSyntheticMonitor`, including union body handling and JSON response structs.
- [ ] 1.2 Compare legacy `go-kibana-rest/v8/kbapi` monitor JSON (from current `toKibanaAPIRequest` / tests) with a sample marshal of generated request types for http/tcp/icmp/browser to confirm field names and discriminators.

## 2. `kibanaoapi` monitor helpers

- [ ] 2.1 Add `internal/clients/kibanaoapi/synthetics_monitor.go` (name may vary) with `CreateMonitor`, `GetMonitor`, `UpdateMonitor`, `DeleteMonitor` functions using `Client.API` / `ClientWithResponses`, space request editors, and consistent error diagnostics.
- [ ] 2.2 Implement constructors or marshal helpers for `PostSyntheticMonitorsJSONBody` / `PutSyntheticMonitorJSONBody` unions from concrete generated monitor + config structs.
- [ ] 2.3 Add unit tests for marshal helpers and status handling (at minimum 404 vs error for get).

## 3. Replace legacy client in resource

- [ ] 3.1 Update `create.go`, `read.go`, `update.go`, `delete.go` to obtain `*kibanaoapi.Client` via `synthetics.GetKibanaOAPIClientFromScopedClient` (or equivalent) and call new helpers instead of `kibanaClient.KibanaSynthetics.Monitor.*`.
- [ ] 3.2 Refactor `schema.go` (and any satellite files) to remove `github.com/disaster37/go-kibana-rest/v8/kbapi` monitor types from public mapping functions; use `generated/kbapi` types or small internal structs.
- [ ] 3.3 Preserve `enforceVersionConstraints`, composite ID logic, import passthrough, and all plan modifiers / validators unchanged unless a compile error forces an equivalent refactor.

## 4. Tests and verification

- [ ] 4.1 Rewrite `schema_test.go` fixtures to use generated types or test-only wire helpers; keep scenario coverage for all four monitor kinds, SSL branches, alerts, and private locations where present today.
- [ ] 4.2 Run `go test ./internal/kibana/synthetics/monitor/...` and fix regressions.
- [ ] 4.3 Run monitor acceptance tests in `acc_test.go` (HTTP/TCP/ICMP/browser, non-default space, labels) against a stack meeting version assumptions; update skips only if OpenAPI behavior documents a new minimum version.
- [ ] 4.4 Run `make build` and `make check-openspec` (or `make check-lint` if that is the project gate for OpenSpec).

## 5. Cleanup

- [ ] 5.1 Remove unused legacy imports from `internal/kibana/synthetics/monitor`; verify no remaining references to legacy `kbapi` monitor types in that tree.
- [ ] 5.2 If no other package uses legacy synthetics monitor kbapi types, consider a follow-up change to trim `go-kibana-rest` usage (optional, separate task if scope bleeds).
