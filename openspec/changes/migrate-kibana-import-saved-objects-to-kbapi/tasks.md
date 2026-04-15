## 1. `kibanaoapi` Saved Objects import helper

- [ ] 1.1 Add `internal/clients/kibanaoapi` helper(s) that accept `context.Context`, `*Client`, `spaceID`, `fileContents []byte` or `string`, and `*kbapi.PostSavedObjectsImportParams`, build `multipart/form-data` with a `file` part, and call `PostSavedObjectsImportWithBodyWithResponse` with `SpaceAwarePathRequestEditor(spaceID)`.
- [ ] 1.2 Map `PostSavedObjectsImportResponse` into a result struct suitable for the resource: treat HTTP 2xx with `JSON200` as success payload; for non-2xx or missing `JSON200`, return a clear error including `JSON400` / body snippets when present.
- [ ] 1.3 Add focused tests (table-driven unit tests where possible) for multipart wire format and/or response classification without requiring a full stack.

## 2. Resource wiring and schema

- [ ] 2.1 Update `internal/kibana/import_saved_objects/create.go` (and shared code paths used by update) to resolve `GetKibanaOapiClient` from the scoped/provider client factory and call the new `kibanaoapi` helper instead of `KibanaSavedObject.Import`.
- [ ] 2.2 Preserve existing computed attributes and diagnostic rules (`success`, `success_count`, `errors`, `success_results`, `ignore_import_errors` warning vs error split, error string formatting intent).
- [ ] 2.3 Uncomment and implement `create_new_copies` and `compatibility_mode` in `schema.go` and `modelV0`; wire pointers into `PostSavedObjectsImportParams` only when attributes are true.
- [ ] 2.4 Enable `ResourceWithConfigValidators` (or equivalent) to enforce mutual exclusion between `create_new_copies` and `overwrite`, and between `create_new_copies` and `compatibility_mode`, matching REQ-013.

## 3. Verification and documentation

- [ ] 3.1 Run `make build` and fix any compile or staticcheck issues from the migration.
- [ ] 3.2 Run targeted acceptance test `TestAccResourceImportSavedObjects` against a configured stack; extend tests or fixtures if needed to cover at least one of the new flags without breaking existing cases.
- [ ] 3.3 Update generated resource documentation (`docs/resources/kibana_import_saved_objects.md`) and the canonical spec under `openspec/specs/kibana-import-saved-objects/spec.md` when syncing post-implementation (or as part of apply if your workflow updates docs in the same change).
