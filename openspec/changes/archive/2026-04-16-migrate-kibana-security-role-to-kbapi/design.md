## Context

Today `internal/kibana/role.go` maps Terraform state to `kbapi.KibanaRole` from `github.com/disaster37/go-kibana-rest/v8/kbapi` and calls `kibana.KibanaRoleManagement.CreateOrUpdate`, `Get`, and `Delete`. The provider already constructs a `kibanaoapi.Client` with `kbapi.ClientWithResponses` for other Kibana resources. The generated spec defines role payloads such as `PutSecurityRoleNameJSONBody`, which mirror the JSON shape Kibana expects for PUT `/api/security/roles/{name}` (cluster, indices with `field_security` as a map of string to string arrays, `remote_indices`, `run_as`, optional `kibana` entries, `metadata`, `description`).

## Goals / Non-Goals

**Goals:**

- Route all role CRUD for `elasticstack_kibana_security_role` (resource and data source) through `KibanaScopedClient.GetKibanaOapiClient()` and new `kibanaoapi` helpers that call `generated/kbapi` `*WithResponse` methods (or equivalent parsing of `Body` when the generator does not attach typed JSON fields).
- Keep the Terraform schema, version gates, flatten/expand rules, and acceptance test expectations stable unless a true remote/API bug is uncovered.
- Centralize HTTP semantics (status codes, JSON decode, error bodies) in `kibanaoapi` alongside `errors.go` patterns used by connectors and other helpers.

**Non-Goals:**

- Migrating the resource from the Terraform Plugin SDK to Plugin Framework.
- Supporting new Kibana role fields not already modeled in Terraform (for example `allow_restricted_indices`, `remote_cluster`) unless they already exist in schema or are required for parity—default is unchanged schema surface.
- Regenerating or editing `generated/kbapi` sources manually beyond normal project processes.

## Decisions

1. **Helper module location** — Add `internal/clients/kibanaoapi/security_role.go` (name can be `role.go` if consistent with package naming) exposing functions such as `GetSecurityRole`, `PutSecurityRole`, `DeleteSecurityRole` taking `context.Context`, `*Client`, role name, optional `GetSecurityRoleNameParams`, body, and `PutSecurityRoleNameParams`. *Rationale:* Matches `connector.go`, `data_views.go`, keeps `internal/kibana` focused on Terraform mapping.

2. **Typed payloads** — Use `kbapi.PutSecurityRoleNameJSONRequestBody` / the same struct family for read decoding into a local mirror type or the PUT body struct, unmarshalling `GetSecurityRoleNameResponse.Body` when `JSON200` is not generated. *Rationale:* Single shape for read/write mapping; avoids retaining parallel structs from go-kibana-rest.

3. **Create-only semantics** — Map `d.IsNewResource()` to `PutSecurityRoleNameParams.CreateOnly` pointer `true` on create and `false` or nil on update, matching current `CreateOnly` behavior. *Rationale:* Kibana documents `createOnly` as a query parameter on PUT.

4. **Not found on read** — Treat HTTP 404 (and any Kibana-documented empty success the API may return) as “role absent”: for the resource, clear `id`; for the data source, return a clear diagnostic or empty result per existing data source behavior. *Rationale:* Generated responses may not expose `JSON404`; parse `StatusCode()` from `HTTPResponse`.

5. **Version and connection** — Continue using `KibanaScopedClient.ServerVersion(ctx)` (status API via legacy client) for gates until a project-wide decision moves version discovery to kbapi; continue resolving scoped clients via `GetKibanaClientFromSDK` / factory. *Rationale:* Minimizes scope; matches current `role.go`.

6. **Privilege parity** — Before merge, run existing acceptance tests; add a short table-driven unit test that round-trips representative `schema.ResourceData`-like inputs through expand → kbapi struct → flatten (or JSON marshal/unmarshal) for indices, remote_indices, kibana base/feature, and metadata. *Rationale:* Catches ordering or `field_security` map shape regressions without requiring a live cluster for every edge case.

## Risks / Trade-offs

- **[Risk] Subtle JSON shape differences** between go-kibana-rest structs and OpenAPI-generated structs (for example `field_security` map keys, nil slices vs omitted fields) → **Mitigation:** parity tests; compare marshalled JSON for a few fixtures against golden files if needed.
- **[Risk] GET response decoding** if content-type or status handling differs by Stack version → **Mitigation:** reuse existing `reportUnknownError` patterns; log/debug transport already wraps HTTP.
- **[Risk] `replaceDeprecatedPrivileges`** not set on GET → **Mitigation:** document default (omit) unless product requires `true` for stable reads; match current behavior (legacy client did not set it).

## Migration Plan

1. Implement `kibanaoapi` helpers and unit-test decoding with mocked `httptest` handlers if feasible.
2. Switch `internal/kibana/role.go` (and data source) to helpers; run `go test ./internal/kibana/...`.
3. Run targeted acceptance tests for `TestAccResourceKibanaSecurityRole` and `TestAccDataSourceKibanaSecurityRole` against a suitable Stack (per `dev-docs/high-level/testing.md`).
4. Remove unused imports / legacy calls from the role code path; verify `make build` and lint.

## Open Questions

- Whether GET should pass `replaceDeprecatedPrivileges=true` for reads to reduce drift on deprecated feature privileges (needs product input; default keep legacy behavior).
- Exact HTTP codes Kibana returns for “not found” on GET in all supported versions (assume 404).
