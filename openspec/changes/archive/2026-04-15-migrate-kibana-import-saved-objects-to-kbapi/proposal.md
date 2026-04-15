## Why

`elasticstack_kibana_import_saved_objects` still calls the legacy go-kibana-rest client (`KibanaSavedObject.Import`), which diverges from other Kibana resources that use `generated/kbapi` and blocks use of query parameters that the OpenAPI spec already models (`createNewCopies`, `compatibilityMode`). Moving the import path to the generated client keeps the provider on one HTTP stack and unlocks those options without bespoke REST code.

## What Changes

- Add a `kibanaoapi` helper that builds a `multipart/form-data` body from `file_contents`, sets the Saved Objects Import path with space awareness (`SpaceAwarePathRequestEditor`), and calls `PostSavedObjectsImportWithBodyWithResponse` with `PostSavedObjectsImportParams` (overwrite, create-new-copies, compatibility mode).
- Refactor `internal/kibana/import_saved_objects` create/update to obtain the Kibana OpenAPI client (`GetKibanaOapiClient`) and invoke the helper instead of the legacy SDK import.
- Preserve existing computed attributes (`success`, `success_count`, `errors`, `success_results`) and the same warning/error diagnostics for partial or failed imports when `ignore_import_errors` is unset or false.
- Expose optional `create_new_copies` and `compatibility_mode` schema attributes backed by kbapi fields, with plan-time validation mirroring Kibana rules (`create_new_copies` incompatible with `overwrite` and `compatibility_mode`).

## Capabilities

### New Capabilities

_(none — behavior stays within the existing resource spec.)_

### Modified Capabilities

- `kibana-import-saved-objects`: Switch implementation to kbapi + `kibanaoapi`; require multipart upload semantics for `file_contents`; update client-selection requirements; extend create/update triggers and import parameters for `create_new_copies` and `compatibility_mode`; clarify HTTP error surfacing for the OpenAPI response wrapper.

## Impact

- **Code**: `internal/kibana/import_saved_objects` (especially `create.go`, `schema.go`), new file(s) under `internal/clients/kibanaoapi/`, `generated/kbapi` types already present (`PostSavedObjectsImportParams`, `PostSavedObjectsImportResponse`).
- **Dependencies**: Removes reliance on legacy `KibanaSavedObject.Import` for this resource; no new third-party modules expected (stdlib `mime/multipart`).
- **Docs / tests**: Resource documentation and acceptance tests should mention new optional arguments once implemented; proposal does not require doc edits in the change folder.
