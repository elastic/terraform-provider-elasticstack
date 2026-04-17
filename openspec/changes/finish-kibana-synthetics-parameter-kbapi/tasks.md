## 1. Delete path migration

- [x] 1.1 Update `internal/kibana/synthetics/parameter/delete.go` to obtain the client via `GetKibanaOAPIClientFromScopedClient` (same as create/read/update) and delete via `DELETE /api/synthetics/params` with `{"ids": [...]}` body through the kbapi HTTP transport (including composite-id parsing already used in delete). Note: `DeleteParameterWithResponse` (`DELETE /api/synthetics/params/{id}`) was not used because it only works on Kibana >= 8.17.0 and returns 404 on 8.12.x–8.16.x; the bulk-style endpoint works on all supported versions from 8.12.0.
- [x] 1.2 Handle delete responses consistently with other `WithResponse` usages in this package (success on expected 2xx; clear diagnostics on failure; no reliance on the legacy `go-kibana-rest` synthetics client for this resource).

## 2. Documentation and spec sync

- [x] 2.1 Update in-repo references that still imply "legacy client for delete" for this resource (e.g. schema notes in `openspec/specs/kibana-synthetics-parameter/spec.md` Purpose/Schema section, traceability text, and any dev docs that mention the split) when merging this change's delta into main specs.
- [x] 2.2 After implementation, run `make check-openspec` (or `make check-lint` as appropriate) and resolve any validation issues for the change artifacts.

## 3. Verification

- [x] 3.1 Run `make build` to ensure the provider compiles after the client swap.
- [x] 3.2 Run targeted acceptance tests for the synthetics parameter resource (e.g. `TestSyntheticParameterResource`) against a Kibana that exposes the Synthetics Parameters API, confirming create/read/update/delete, import id handling, read-after-write, and `share_across_spaces` behavior are unchanged.
