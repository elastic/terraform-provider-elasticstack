## Context

Today `importObjects` in `internal/kibana/import_saved_objects/create.go` uses `apiClient.GetKibanaClient()` and `kibanaClient.KibanaSavedObject.Import([]byte(file_contents), overwrite, spaceID)`. The generated OpenAPI client in `generated/kbapi` already exposes `PostSavedObjectsImportWithBodyWithResponse`, `PostSavedObjectsImportParams` (including `Overwrite`, `CreateNewCopies`, `CompatibilityMode`), and a typed `PostSavedObjectsImportResponse` with `JSON200` containing `Success`, `SuccessCount`, and loosely typed `errors` / `successResults` as slices of maps. Other Kibana features use `internal/clients/kibanaoapi.Client` and `SpaceAwarePathRequestEditor` for space-prefixed paths.

## Goals / Non-Goals

**Goals:**

- Centralize Saved Objects import HTTP details in `kibanaoapi`, calling `client.API.PostSavedObjectsImportWithBodyWithResponse` with the correct `Content-Type` and body.
- Build `multipart/form-data` from the Terraform `file_contents` string (NDJSON export payload) with a file part named `file` (Kibana import API contract).
- Map API JSON into the same Terraform state shape as today (including `mapstructure` or equivalent decoding into lists/objects matching the existing schema).
- Preserve diagnostic behavior: `ignore_import_errors` semantics, warning on partial success, error on total failure, per-error string formatting similar to `importError.String()`.
- Add optional `create_new_copies` and `compatibility_mode` booleans to the resource schema, passed through `PostSavedObjectsImportParams` when set, and validate mutually exclusive combinations at plan time (`ResourceWithConfigValidators`).

**Non-Goals:**

- Changing read/delete no-op behavior, `id` generation, or the write-only resource model.
- Implementing the separate “resolve import errors” API flow.
- Regenerating `kbapi` (spec is already sufficient).

## Decisions

| Decision | Rationale | Alternatives considered |
|----------|-----------|-------------------------|
| New helper file `internal/clients/kibanaoapi/saved_objects_import.go` (name may vary slightly but lives in `kibanaoapi`) | Matches existing pattern (`workflows.go`, `security_lists.go`): thin wrapper around `kbapi`, returns diagnostics-friendly results. | Putting multipart + HTTP directly in the resource package — duplicates auth/space handling and is harder to test. |
| Use `PostSavedObjectsImportWithBodyWithResponse` with a hand-built multipart body | The generated multipart type uses `map[string]interface{}` for `File`, which is awkward for raw NDJSON bytes; `WithBody` accepts any `io.Reader` and content type. | Typed multipart helpers from codegen — not ergonomic for in-memory export strings. |
| Decode `JSON200` maps into the existing `responseModel` / Terraform attribute types | Keeps computed attribute contract stable; `JSON200` fields are already `map[string]interface{}` for nested structures. | Replacing with strongly typed kbapi structs — would require parallel structs and more mapping work for little gain. |
| Surface non-200 responses using `JSON400` / status / body | Aligns with REQ-002; kbapi separates `JSON200` and `JSON400`. | Ignoring 400 bodies — loses actionable Kibana error messages. |
| Config validators for `create_new_copies` vs `overwrite` / `compatibility_mode` | Kibana documents mutual exclusion; fail fast at plan time instead of opaque API errors. | Rely on API-only validation — worse UX and noisy apply failures. |

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Multipart boundary or field name mismatch causes 400 from Kibana | Match Elastic docs: single part `file`; set filename (e.g. `export.ndjson`) on the form file part; integration/acc test covers happy path. |
| `successCount` is `float32` in codegen vs `int64` in Terraform | Convert with explicit truncation/rounding to int64 consistent with current behavior (counts are integers). |
| Large `file_contents` strings memory-double with multipart buffer | Accept for now (same as passing `[]byte` to legacy client); document unchanged limitation. |

## Migration Plan

1. Implement `kibanaoapi` helper and unit-test multipart construction if feasible without a live cluster.
2. Switch resource to OpenAPI client; run `make build` and targeted acceptance test `TestAccResourceImportSavedObjects`.
3. Add new optional attributes + validators; extend acceptance tests for one conflict scenario with `create_new_copies` or `compatibility_mode` if a stable config exists.

Rollback: revert to legacy `KibanaSavedObject.Import` behind the same resource surface (not desired long-term but mechanically simple).

## Open Questions

- Whether acceptance tests should assert a plan-time error for invalid flag combinations, or only unit-test validators (prefer at least one validator unit test to avoid acc-test flakiness).
- Exact Kibana version floor for `compatibilityMode` / `createNewCopies` if any; assume same as current provider-supported Kibana unless QA finds otherwise.
