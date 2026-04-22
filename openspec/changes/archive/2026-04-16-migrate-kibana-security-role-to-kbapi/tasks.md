## 1. kibanaoapi role helpers

- [x] 1.1 Add `internal/clients/kibanaoapi/security_role.go` with `GetSecurityRole`, `PutSecurityRole`, and `DeleteSecurityRole` (names adjustable) wrapping `kbapi.ClientWithResponses`, decoding GET `Body` into `PutSecurityRoleNameJSONBody` (or shared struct), and handling 404 on read as not found.
- [x] 1.2 Reuse `reportUnknownError` / existing error helpers for non-2xx responses; add focused tests with `httptest` for 200, 404, and error body paths where practical.

## 2. Terraform mapping migration

- [x] 2.1 Replace `kbapi.KibanaRole` / `KibanaRoleManagement` usage in `internal/kibana/role.go` with `kibanaoapi` helpers and `kbapi.PutSecurityRoleNameJSONRequestBody` (and related types); obtain clients via `GetKibanaOapiClient()` from `KibanaScopedClient`.
- [x] 2.2 Refactor expand functions to emit `kbapi` OpenAPI structs (including `field_security` as `*map[string][]string` per generated schema), preserving version gates for `remote_indices` and `description`.
- [x] 2.3 Refactor flatten functions to read from the decoded OpenAPI response shape while preserving omission rules for empty `cluster` / `run_as` and existing set/list flattening behavior.
- [x] 2.4 Update the data source implementation in `internal/kibana` to use the same helper read path; align not-found behavior with REQ-012–REQ-014.

## 3. Parity, tests, and cleanup

- [x] 3.1 Add unit tests that round-trip representative privilege configurations through expand and flatten (and/or JSON) to lock privilege parity.
- [x] 3.2 Run `make build` and `go test ./internal/kibana/...`; run acceptance tests for `TestAccResourceKibanaSecurityRole` and `TestAccDataSourceKibanaSecurityRole` when a Stack is available.
- [x] 3.3 Remove unused go-kibana-rest imports from the role code path; verify no regressions in dependent testdata (for example Fleet agent policy restricted user).

## 4. OpenSpec sync

- [x] 4.1 After implementation, run `make check-openspec` (or `openspec validate`) and archive or sync per project workflow so `openspec/specs/kibana-security-role/spec.md` absorbs the delta.
