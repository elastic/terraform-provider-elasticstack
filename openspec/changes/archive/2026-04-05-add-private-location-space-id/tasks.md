## 1. API client (`kbapi`)

- [x] 1.1 Add a `space string` parameter to `KibanaSyntheticsPrivateLocationCreate`, `Get`, and `Delete` function types and implementations, passing it into `basePath` / `basePathWithId` instead of a hard-coded empty string.
- [x] 1.2 Update `libs/go-kibana-rest/kbapi` tests that call private location APIs to pass space (including default-space cases).
- [x] 1.3 Run targeted tests for `kbapi` package and fix any compile errors at other call sites.

## 2. Terraform resource

- [x] 2.1 Add optional `space_id` to `tfModelV0` and `privateLocationSchema()` with `RequiresReplace` and documentation consistent with synthetics monitor `space_id`.
- [x] 2.2 Thread `space_id` into Create, Read, and Delete via the updated `kbapi` functions; map read response to state including `space_id`.
- [x] 2.3 Extend `schema_test.go` (and any model round-trips) for `space_id` default and non-empty values.
- [x] 2.4 Add or extend acceptance test coverage for default vs non-default space per project testing conventions (`dev-docs/high-level/testing.md`).

## 3. Documentation and validation

- [x] 3.1 Regenerate or update resource documentation so `space_id` appears in the provider docs.
- [x] 3.2 Run `make build` and `make check-openspec` (or `openspec validate` for this change) before merge.
