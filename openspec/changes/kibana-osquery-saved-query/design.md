## Context

Kibana exposes Osquery saved queries under `/api/osquery/saved_queries` with full CRUD plus a list endpoint. All four CRUD Go client bindings already exist in `generated/kbapi/kibana.gen.go`:

- `OsqueryCreateSavedQuery` / `OsqueryCreateSavedQueryWithResponse`
- `OsqueryGetSavedQueryDetails` / `...WithResponse`
- `OsqueryUpdateSavedQuery` / `...WithResponse`
- `OsqueryDeleteSavedQuery` / `...WithResponse`

The osquery paths are **not** in the `spaceIdPaths` allow-list in `generated/kbapi/transform_schema.go`, so generated URLs lack `/s/{spaceId}`. Space support is injected at request time via `kibanautil.SpaceAwarePathRequestEditor(spaceID)`, the same mechanism used by `agentbuilder`, `entity_store`, `alerting_rule`, `dashboards`, and `synthetics` — no `transform_schema.go` / `make transform generate` changes required.

The non-trivial implementation areas are:

1. **`ecs_mapping` modelling**: each key maps to `{ field, value: string | []string }` — a three-way exactly-one-of constraint (`field`/`value`/`values`) on a `MapNestedAttribute`.
2. **`interval`/`version` union types**: the API returns these as `json.RawMessage` unions (`int | string`). Read uses `AsXxx0()/AsXxx1()` accessors; write normalises to `int64` for `interval` and `string` for `version`.
3. **Prebuilt-query protection**: `prebuilt` is server-returned and unknown at plan time — enforced as a runtime diagnostic, not a plan-time validator.

## Goals / Non-Goals

**Goals:**
- Full lifecycle (Create, Read, Update, Delete) of user-managed Osquery saved queries with import support.
- Single-item data source for read-only lookup (prebuilt-safe).
- Space-awareness via composite `<space_id>/<saved_query_id>` import ID.
- Faithful `ecs_mapping` representation covering all three value shapes (`field`, scalar `value`, array `values`).
- `interval` and `version` round-tripped without data loss despite API union types.
- Runtime guard that prevents managing prebuilt queries via the resource, with a clear diagnostic pointing to the data source.

**Non-Goals:**
- Osquery packs (`elasticstack_kibana_osquery_pack`) — separate resource.
- Osquery live queries — ephemeral, not suitable for Terraform.
- List data source (`OsqueryFindSavedQueries`) — follow-up.
- Bulk Saved Objects API import/export.

## Decisions

### Decision 1: Resource pattern — `entitycore.KibanaResource[Model]`

Implement the resource as an `entitycore.KibanaResource[osquerySavedQueryModel]`, matching `maintenance_window`, `slo`, `spaces`, `alerting_rule`, and other recent Kibana resources. The generic wrapper handles `kibana_connection`, timeouts, and configure. A thin `internal/clients/kibanaoapi/osquery_saved_query.go` helper wraps the four kbapi calls.

**Why:** The pattern is standard for all new Kibana resources in this provider. Maintenance Window is the closest existing analogue (medium-complexity, Plugin Framework, kibanaoapi helper). Considered: standalone `resource.Resource` without entitycore — rejected for inconsistency and unnecessary boilerplate.

### Decision 2: Identity — `saved_query_id` Optional+Computed, RequiresReplace

`saved_query_id` is Optional+Computed with `RequiresReplace`. When omitted, the resource falls back to the server-generated ID (expected to be a UUID). `id` in state is Computed and mirrors `saved_query_id`. Import uses `saved_query_id` as the lookup key.

**Why:** Matches the majority of Kibana resources where the user may provide a meaningful ID (e.g., `list_all_processes`) or defer to Kibana. RequiresReplace prevents a silent rename, which would orphan the old query.

**Open verification item**: Confirm during implementation that Kibana generates a UUID when `saved_query_id` is omitted on create; escalate to Required if it doesn't (unlikely but undocumented).

### Decision 3: Space support via `SpaceAwarePathRequestEditor`

`space_id` is Optional+Computed, defaults to `"default"`, RequiresReplace. All four kbapi calls in the kibanaoapi helper pass `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as a request editor. This injects `/s/{spaceID}` before the first `/api/` segment for non-default spaces, matching the existing pattern in `agentbuilder`, `entity_store`, `alerting_rule`, `dashboards`, `synthetics`. No `transform_schema.go` changes needed.

### Decision 4: Composite import ID `<space_id>/<saved_query_id>`

Import accepts `"<space_id>/<saved_query_id>"` (e.g., `"default/list_all_processes"`). When `space_id` is `"default"`, the default space is used. This matches the import idiom of every other space-aware Kibana resource in the provider.

### Decision 5: `platform` — set of strings, join to comma-string on write

`platform` is an Optional `schema.SetAttribute` of strings. Allowed values validator restricts to `"linux"`, `"darwin"`, `"windows"`. On write, the set is sorted and joined to a comma-separated string (e.g., `"linux,darwin"`). On read, the comma-string is split back to a set. Sorting ensures deterministic plan output.

### Decision 6: `ecs_mapping` — `MapNestedAttribute` with exactly-one-of-3 value validator

`ecs_mapping` is an Optional `MapNestedAttribute`. Each element is a `SingleNestedAttribute` with three fields:

- `field` (Optional string) — maps to a result column name in the query output
- `value` (Optional string) — static scalar value
- `values` (Optional set of string) — static array value

A `ConfigValidator` enforces that exactly one of `field`, `value`, or `values` is set per element. Maps to the generated `{Field, Value: string|[]string}` shape: `field` → `{field: "..."}`, `value` → `{value: "..."}` (string), `values` → `{value: [...]}` (array).

Partial ECS mapping precedent exists at `internal/kibana/security_detection_rule/models_to_api_type_utils.go:827` (`buildEcsMappingFromModel`) but covers only the `field` case. This resource must handle all three.

**Open verification item**: Confirm during implementation that `plugin-framework-validators` exactly-one-of works inside `MapNestedAttribute` values for the three-way constraint.

### Decision 7: `interval` — `Int64Attribute`

`interval` is Optional `Int64Attribute` (seconds). On read, the `json.RawMessage` union is read via `AsInt` accessor; if that fails, the string arm is parsed as `int64`. On write, the `int64` is stringified before sending to the API (the API wire type accepts both `int` and `string`; sending as string is always valid). Nullable (omit from request body when not set).

### Decision 8: `version` — `StringAttribute`

`version` is Optional `StringAttribute`. On read, the `json.RawMessage` union is stringified (using the string accessor; fallback to `fmt.Sprintf` if the int arm). On write, the string is sent verbatim. Nullable (omit from request body when not set).

### Decision 9: `snapshot` and `removed` — Optional+Computed, no static default

`snapshot` and `removed` are Optional+Computed booleans with no static default. The API decides on create; on subsequent reads the server value is preserved in state. This avoids Terraform generating spurious diffs if the API defaults change.

### Decision 10: Prebuilt-query protection — runtime error diagnostic

When `prebuilt == true` in the API response (on Read or post-Create Read), the resource returns an error diagnostic:

> `"Prebuilt Osquery saved queries are managed by the osquery_manager integration package and cannot be managed by this resource. Use the elasticstack_kibana_osquery_saved_query data source to read this query."`

This is a runtime guard (not plan-time) because `prebuilt` is server-returned and unknown at plan time. The resource does not expose a `prebuilt` attribute in state. Affects Read and the read-after-write in Create.

### Decision 11: Update is PUT (full body)

The API uses PUT (full replacement, not PATCH). On Update, all managed fields are sent; server-managed fields (`created_at`, `updated_at`, `created_by_profile_uid`, `updated_by_profile_uid`, `saved_object_id`) are omitted from the request body. No drift risk.

### Decision 12: Delete returns empty body — `HandleStatusResponse`

Delete returns an empty body with HTTP 200 on success, and the provider should treat HTTP 404 as idempotent success. The kibanaoapi wrapper uses `HandleStatusResponse(..., http.StatusOK)` with a 404 no-op, matching the `maintenance_window` pattern.

### Decision 13: Create response wraps in `data` — unwrap in `populateFromAPI`

The Create response wraps the entity in a `data` field. The `populateFromAPI` function must unwrap `data` before mapping to the model.

### Decision 14: Server-managed fields not exposed

`created_at`, `updated_at`, `created_by_profile_uid`, `updated_by_profile_uid`, and `saved_object_id` are server-managed and NOT exposed as attributes. Only the internal Computed `id` is stored in state (mirrors `saved_query_id`).

### Decision 15: Data source — single-item GET-by-id, prebuilt-safe

The data source accepts `saved_query_id` (Required), `space_id` (Optional, default `"default"`), and `kibana_connection` (Optional). It calls GET by ID and populates all managed fields. Unlike the resource, it does NOT error on `prebuilt == true` — the data source is the intentional path for referencing prebuilt queries.

### Decision 16: Minimum version — `8.5.0` (verify during implementation)

The resource declares `8.5.0` as the minimum Kibana version via `GetVersionRequirements`. This must be verified against actual Kibana CHANGELOG / API availability during implementation; it may need to be raised.

### Decision 17: Naming

Resource and data source: `elasticstack_kibana_osquery_saved_query`. Go package: `internal/kibana/osquery_saved_query`. Registration in `provider/plugin_framework.go` alongside `maintenance_window` and other Kibana resources.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **`saved_query_id` omitted on create — does Kibana generate a UUID?** | Verification task in tasks.md; escalate to Required if not. |
| **`8.5.0` version floor is too low** | Verification task in tasks.md; the acceptance test exercises real CRUD and will fail if the floor is wrong. |
| **`ecs_mapping` exactly-one-of validator inside `MapNestedAttribute`** | Verification task in tasks.md; test confirms validator fires on invalid and accepts all three valid forms. |
| **`interval`/`version` union-type edge cases** | Both union arms (`AsXxx0`/`AsXxx1`) are exercised in unit tests; round-trip acceptance test confirms no data loss. |
| **Prebuilt queries silently imported** | Runtime guard on Read (and post-Create) returns an explicit error diagnostic before touching state. |

## Open Questions

1. **Exact minimum Kibana version**: Confirm the CRUD endpoints exist at `8.5.0`; raise if testing fails.
2. **Server-generated ID on create**: Confirm Kibana generates a UUID when `saved_query_id` is omitted; escalate to Required if it doesn't.
3. **`plugin-framework-validators` exactly-one-of inside `MapNestedAttribute`**: Confirm this works; if not, implement an inline `ValidateObject` function on the nested attribute.
