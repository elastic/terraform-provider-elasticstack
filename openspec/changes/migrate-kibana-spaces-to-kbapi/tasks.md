## 1. OpenAPI schema and codegen

- [ ] 1.1 Identify current kbapi types and operations for `/api/spaces/space` (list, get by id, create, update, delete) and compare JSON fields to legacy `kbapi.KibanaSpace` in `libs/go-kibana-rest/kbapi/api.kibana_spaces.go`.
- [ ] 1.2 Extend `generated/kbapi/transform_schema.go` so space list and space detail responses generate strongly typed models matching the legacy `KibanaSpace` shape (including `disabledFeatures`, `imageUrl`, `solution`, `_reserved` as applicable).
- [ ] 1.3 Regenerate `generated/kbapi` using the subdirectory workflow (`generated/kbapi` Makefile: transform + `oapi-codegen`) and ensure the repo builds (`make build`).

## 2. kibanaoapi helpers

- [ ] 2.1 Add `internal/clients/kibanaoapi` Spaces helpers (e.g. new `spaces.go`) wrapping the kbapi `WithResponse` methods for list, get, create, update, and delete, including consistent non-success status handling aligned with existing kibanaoapi modules.
- [ ] 2.2 Expose a path for both the SDK resource and PF data source to obtain a `*kibanaoapi.Client` (or equivalent) from the same effective `kibana_connection` resolution used elsewhere, adding a factory/helper only if missing.

## 3. Terraform entities

- [ ] 3.1 Migrate `internal/kibana/space.go` to call the new helpers for create, update, read, and delete; remove direct `KibanaSpaces` usage while preserving composite id handling, import behavior, post-upsert read refresh, optional-field omission on write, read mapping (including not setting `image_url` from API), and diagnostics text where appropriate.
- [ ] 3.2 Preserve `solution` minimum version gating (`>= 8.16.0`) before create/update when `solution` is non-empty, using server version from the same effective connection as the kbapi calls.
- [ ] 3.3 Migrate `internal/kibana/spaces` (Plugin Framework data source `read.go` and any shared wiring) to list spaces via the helpers instead of `GetKibanaClient().KibanaSpaces.List`, preserving `spaces` nested attribute mapping and `id = "spaces"` behavior.

## 4. Verification

- [ ] 4.1 Update or add unit tests for transforms, helpers, or mapping if the repository pattern supports them; otherwise rely on compile-time checks and targeted acceptance tests.
- [ ] 4.2 Run targeted acceptance tests for `elasticstack_kibana_space` and `elasticstack_kibana_spaces` when a stack is available; at minimum run `make build` and relevant `go test` packages.
- [ ] 4.3 Run `make check-openspec` or `./node_modules/.bin/openspec validate --all` after syncing expectations to ensure delta specs validate.
