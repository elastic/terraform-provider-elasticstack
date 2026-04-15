## Context

Today `elasticstack_kibana_synthetics_private_location` uses `synthetics.GetKibanaClientFromScopedClient` and `kibanaClient.KibanaSynthetics.PrivateLocation.{Create,Get,Delete}` from `go-kibana-rest`, with Terraform models converting to `kbapi.PrivateLocationConfig` / `kbapi.PrivateLocation` (`internal/kibana/synthetics/privatelocation/schema.go`). Other Synthetics surfaces already use `GetKibanaOAPIClientFromScopedClient` and `internal/clients/kibanaoapi` helpers. The generated Kibana client already exposes `PostPrivateLocationWithResponse`, `GetPrivateLocationWithResponse`, and `DeletePrivateLocationWithResponse` plus `PostPrivateLocationJSONBody` and `SyntheticsGetPrivateLocation` in `generated/kbapi/kibana.gen.go`.

## Goals / Non-Goals

**Goals:**

- Implement private location CRUD through `kibanaoapi` + `generated/kbapi`, including correct space-prefixed paths for non-default spaces (same effective URLs as today).
- Remove `go-kibana-rest` kbapi private location types from this resource’s model mapping path.
- Keep `space_id` 9.4.0-SNAPSHOT gating, composite import ids, `ImportStatePassthroughID`, all `RequiresReplace` plan modifiers, Update-as-error behavior, and `kibana_connection` behavior unchanged from a practitioner perspective.
- Prove mapping fidelity for `SyntheticsGetPrivateLocation` (including JSON fields not modeled as first-class struct fields but preserved via `AdditionalProperties`, such as `tags` if the OpenAPI model omits them).

**Non-Goals:**

- Regenerating or editing the upstream Kibana OpenAPI spec in this change.
- Supporting in-place updates or new Terraform attributes.
- Migrating other `go-kibana-rest` Synthetics code paths beyond this resource.

## Decisions

1. **`kibanaoapi` helper module** — Add functions such as `CreatePrivateLocation`, `GetPrivateLocation`, and `DeletePrivateLocation` (exact names follow repo conventions) that accept `context.Context`, `*kibanaoapi.Client`, effective `spaceID`, location id / body, and return typed results plus `diag.Diagnostics`. Centralize the `kbapi.RequestEditorFn` that rewrites `req.URL.Path` to prefix `/s/<space_id>` when the effective space is not default, matching patterns used elsewhere for space-scoped Kibana routes.

2. **Typed read model vs loose create response** — `GetPrivateLocation` parses JSON200 into `kbapi.SyntheticsGetPrivateLocation`. `PostPrivateLocation` currently exposes `JSON200` as `*map[string]interface{}` in generated code. **Decision:** After a successful POST, normalize the created object by unmarshaling the raw `Body` (or re-encoding the map) into `kbapi.SyntheticsGetPrivateLocation` so create and read share one mapping function to Terraform state. If unmarshaling fails, return a clear diagnostic (do not silently drop fields).

3. **404 handling** — Use HTTP status from `*WithResponse` results (`StatusCode() == 404` on success path with nil JSON, or documented generator behavior) rather than `go-kibana-rest`’s `kbapi.APIError`. Align with other PF resources that use generated clients.

4. **Geo numeric types** — OpenAPI uses `float32` in nested geo structs where Terraform uses `float64`. **Decision:** Convert explicitly at boundaries; document acceptable precision limits (float32) so behavior matches API, not introduce new validation beyond existing schema.

5. **Tags on read** — If `tags` appear only under `AdditionalProperties` on `SyntheticsGetPrivateLocation`, extract them in the mapper with a small, tested helper rather than fork-generating the OpenAPI model.

## Risks / Trade-offs

- **[Risk] Generated POST typing is weak (`map[string]interface{}`)** → **Mitigation:** Always normalize through `SyntheticsGetPrivateLocation` + shared mapper; add unit tests for JSON fixtures.

- **[Risk] OpenAPI model incomplete vs real Kibana JSON** → **Mitigation:** Rely on custom `UnmarshalJSON` / `AdditionalProperties` on `SyntheticsGetPrivateLocation`; acceptance tests plus focused unit tests on tags/geo.

- **[Risk] Space path regression** → **Mitigation:** Reuse the same `RequestEditorFn` approach as other space-aware `kbapi` calls; keep existing acceptance tests for default and non-default space.

## Migration Plan

1. Land `kibanaoapi` helpers and unit-tested mappers.
2. Switch `privatelocation` create/read/delete to the OpenAPI client; remove legacy imports from that package.
3. Run `make build`, unit tests for the package, and targeted acceptance tests (`TestSyntheticPrivateLocationResource` and related) on a stack that satisfies test version gates.

## Open Questions

- None blocking proposal: confirm during implementation whether create responses are always JSON-compatible with `SyntheticsGetPrivateLocation` on supported stack versions; if not, extend the normalizer with a narrow fallback for `id`/`label` only and fail closed for inconsistent payloads.
