## 1. Delete path migration

- [ ] 1.1 Update `internal/kibana/synthetics/parameter/delete.go` to obtain the client via `GetKibanaOAPIClientFromScopedClient` (same as create/read/update) and call `DeleteParameterWithResponse` with the resolved parameter id (including composite-id parsing already used in delete).
- [ ] 1.2 Handle delete responses consistently with other `WithResponse` usages in this package (success on expected 2xx; clear diagnostics on failure; no reliance on the legacy `go-kibana-rest` synthetics client for this resource).

## 2. Documentation and spec sync

- [ ] 2.1 Update in-repo references that still imply “legacy client for delete” for this resource (e.g. schema notes in `openspec/specs/kibana-synthetics-parameter/spec.md` Purpose/Schema section, traceability text, and any dev docs that mention the split) when merging this change’s delta into main specs.
- [ ] 2.2 After implementation, run `make check-openspec` (or `make check-lint` as appropriate) and resolve any validation issues for the change artifacts.

## 3. Verification

- [ ] 3.1 Run `make build` to ensure the provider compiles after the client swap.
- [ ] 3.2 Run targeted acceptance tests for the synthetics parameter resource (e.g. `TestSyntheticParameterResource`) against a Kibana that exposes the Synthetics Parameters API, confirming create/read/update/delete, import id handling, read-after-write, and `share_across_spaces` behavior are unchanged.
