## Context

Today, `internal/kibana/space.go` and `internal/kibana/spaces/read.go` obtain a `*kbapi.Client` from the provider factory and call `KibanaSpaces.List|Get|Create|Update|Delete` from `github.com/disaster37/go-kibana-rest/v8/kbapi`. That client uses hand-written REST helpers with loosely typed JSON unmarshalling into `kbapi.KibanaSpace`. The repository already ships a large generated OpenAPI client under `generated/kbapi` plus thin wrappers in `internal/clients/kibanaoapi` (for example `data_views_spaces.go`, `alerting_rule.go`) that call `client.API.*WithResponse`, normalize status codes, and return structured errors for Terraform diagnostics.

The Kibana OpenAPI schema for Spaces may currently expose list/get responses as `additionalProperties`/generic maps or incomplete structs, which is why `transform_schema.go` must be extended before regeneration so space entities decode into Go structs that cover the same fields as legacy `KibanaSpace`.

## Goals / Non-Goals

**Goals:**

- Route all Spaces CRUD and list operations for `elasticstack_kibana_space` and `elasticstack_kibana_spaces` through `generated/kbapi` using `kibanaoapi` helpers consistent with other migrated entities.
- After regeneration, use strongly typed kbapi models for space payloads whose JSON tags align with legacy `KibanaSpace` (`id`, `name`, `description`, `disabledFeatures`, `initials`, `color`, `imageUrl`, `solution`, `_reserved` when present) so Terraform attribute mapping stays equivalent.
- Preserve `solution` minimum version checks (8.16.0) on create/update when `solution` is non-empty, using the same server version resolution path as today (`client.ServerVersion` on the effective connection wrapper), independent of whether HTTP is issued via resty or oapi-codegen transport.
- Preserve composite `id` parsing, import passthrough, not-found read behavior, post-upsert read refresh, omitting optional unset fields on write, and omitting `image_url` from read mapping.

**Non-Goals:**

- Changing Terraform schema, attribute names, or validation rules for either entity.
- Migrating the SDK space resource to Plugin Framework (out of scope).
- Removing or altering `libs/go-kibana-rest` for unrelated consumers.

## Decisions

1. **Transform + regenerate instead of ad-hoc structs** — Prefer fixing the OpenAPI-derived schema in `transform_schema.go` (similar to `fixGetSpacesParams` and other transformers) so `oapi-codegen` emits dedicated types for space list/item responses. **Rationale:** Keeps a single source of truth and avoids parallel hand-maintained DTOs. **Alternative considered:** private structs with `json.Unmarshal` into `map[string]any` — rejected for weaker compile-time guarantees and drift risk.

2. **`kibanaoapi` facade for Spaces** — Add a focused module (e.g. `spaces.go`) exposing functions or methods on `*kibanaoapi.Client` that wrap the specific `kbapi` operations (GET `/api/spaces/space`, GET `/api/spaces/space/{id}`, POST, PUT, DELETE) and map HTTP failures to Terraform-friendly errors, mirroring patterns in `alerting_rule.go` / `data_views.go`. **Rationale:** Keeps resource/data source code thin and centralizes status handling. **Alternative considered:** calling `client.API` directly from `internal/kibana` — rejected to avoid duplicating HTTP semantics.

3. **SDK resource keeps `GetKibanaClientFromSDK` until a broader factory migration** — Only replace the `KibanaSpaces` call sites with code that builds or reuses `*kibanaoapi.Client` from the same effective `kibana_connection` / provider defaults as other PF entities. **Rationale:** Minimizes churn outside Spaces; align with how other resources obtain the oapi client from the factory if a helper already exists, or add a narrow factory method documented in tasks.

4. **PF data source** — Replace `apiClient.GetKibanaClient()` + `KibanaSpaces.List` with the same `kibanaoapi` list helper used by the design, mapping typed results into the existing `spaces` model. **Rationale:** One implementation of list semantics and error handling.

## Risks / Trade-offs

- **[Risk] OpenAPI schema mismatch** — Kibana’s published OAS may not describe every field the legacy client returned. **Mitigation:** Compare generated structs against `libs/go-kibana-rest/kbapi/api.kibana_spaces.go` struct; extend transforms or use merge-patch overlays in `transform_schema.go` until fields match; add unit tests with fixture JSON from both default and custom spaces.

- **[Risk] Subtle state diffs** — Differences in nil vs empty slice or omitted JSON fields could cause plan noise. **Mitigation:** Map through explicit normalization functions shared by resource read and data source; mirror existing `Set` / `ListValueFrom` behavior.

- **[Risk] `solution` gating regression** — If the kbapi path skips version lookup, older clusters could error differently. **Mitigation:** Keep the explicit version guard in `resourceSpaceUpsert` before any mutating kbapi call; add a scenario-style unit test if feasible.

## Migration Plan

1. Land schema transforms and regenerate `generated/kbapi` in the developer environment (`make` targets under `generated/kbapi/` per repo docs).
2. Implement `kibanaoapi` spaces helpers and unit tests with mocked `*http.Client` or golden responses if the repo pattern supports it.
3. Switch `internal/kibana/space.go` and `internal/kibana/spaces` to the helpers; run `make build` and targeted acceptance tests for space + spaces data source.
4. No user migration steps — state format unchanged.

## Open Questions

- The exact factory helper used elsewhere to obtain `*kibanaoapi.Client` from `kibana_connection` for SDK resources should be confirmed during implementation (search `kibanaoapi.New` / factory patterns); if none exists for SDK, add a small bridge without widening scope beyond Spaces.
