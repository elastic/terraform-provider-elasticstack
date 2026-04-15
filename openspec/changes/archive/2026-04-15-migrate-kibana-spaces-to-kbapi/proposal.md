## Why

The `elasticstack_kibana_space` resource and `elasticstack_kibana_spaces` data source still call Kibana through the legacy `go-kibana-rest` `KibanaSpaces` client, while the rest of the provider is standardizing on the generated OpenAPI client (`generated/kbapi`) plus `internal/clients/kibanaoapi` helpers. Moving these entities reduces duplicate HTTP stacks, improves type safety, and aligns Spaces with the same client patterns as other Kibana resources.

## What Changes

- Extend `generated/kbapi/transform_schema.go` so regenerated OpenAPI models for space responses decode the same JSON surface as legacy `kbapi.KibanaSpace` (field names and optionality), enabling strongly typed request/response structs without losing fields today mapped into Terraform state.
- Regenerate `generated/kbapi` (transform + `oapi-codegen`) so the kbapi package includes the adjusted space schemas.
- Add `internal/clients/kibanaoapi` helpers for list/get/create/update/delete spaces that wrap the kbapi client, handle HTTP status and errors consistently with other kibanaoapi modules, and return data suitable for the SDK resource and PF data source.
- Migrate `elasticstack_kibana_space` (SDK in `internal/kibana/space.go`) and `elasticstack_kibana_spaces` (PF under `internal/kibana/spaces`) to use those helpers instead of `kibanaClient.KibanaSpaces.*`, while preserving:
  - `solution` minimum version gating (8.16.0) before create/update when `solution` is set,
  - composite `id` handling, import, destroy, post-upsert read refresh, and `image_url` not being populated from read,
  - diagnostics and state values equivalent to current behavior for the same Kibana responses.

## Capabilities

### New Capabilities

_(None — behavior stays under existing capabilities.)_

### Modified Capabilities

- `kibana-space`: Document that Spaces HTTP traffic is satisfied via `generated/kbapi` + `kibanaoapi`, and that typed models MUST remain compatible with the legacy `KibanaSpace` JSON shape for state mapping; existing user-facing requirements (including `solution` gating) remain.
- `kibana-spaces`: Update connection/client wording and list implementation expectations to match kbapi + `kibanaoapi` while keeping list semantics and field mapping unchanged.

## Impact

- `generated/kbapi/transform_schema.go`, regenerated files under `generated/kbapi/`, and the kbapi Makefile-driven codegen pipeline.
- New or updated Go in `internal/clients/kibanaoapi/` (e.g. `spaces.go`).
- `internal/kibana/space.go` and `internal/kibana/spaces/` (read path and client acquisition).
- Tests that assert on client types or mocks may need updates; acceptance tests should remain valid if responses are unchanged.
