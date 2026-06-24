## Context

Kibana exposes Osquery saved queries under `/api/osquery/saved_queries` with full CRUD plus a list endpoint. All four CRUD Go client bindings already exist in `generated/kbapi/kibana.gen.go`:

- `OsqueryCreateSavedQuery` / `OsqueryCreateSavedQueryWithResponse`
- `OsqueryGetSavedQueryDetails` / `...WithResponse`
- `OsqueryUpdateSavedQuery` / `...WithResponse`
- `OsqueryDeleteSavedQuery` / `...WithResponse`

The osquery paths are **not** in the `spaceIdPaths` allow-list in `generated/kbapi/transform_schema.go`, so generated URLs lack `/s/{spaceId}`. Space support is injected at request time via `kibanautil.SpaceAwarePathRequestEditor(spaceID)`, the same mechanism used by `agentbuilder`, `entity_store`, `alerting_rule`, `dashboards`, and `synthetics` — no `transform_schema.go` / `make transform generate` changes required.

The non-trivial implementation areas are:

1. **`ecs_mapping` modelling**: each key maps to `{ field, value: string | []string }` — a three-way exactly-one-of constraint (`field`/`value`/`values`) on a `MapNestedAttribute`.
2. **`interval`/`version` response shapes differ by operation**: Create and GET API responses wrap the entity in `.Data` and type both fields as `json.RawMessage` unions (`int | string`). Update also wraps `.Data`, but types `version` as plain `*string` while `interval` remains a union. The kibanaoapi helper unwraps `.Data` into per-operation entity types (`OsquerySavedQueryCreateEntity`, `OsquerySavedQueryGetEntity`, `OsquerySavedQueryUpdateEntity`); `populateFromAPI` (task 3.4) maps those entities to the model and must branch on entity type.
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

### Decision 2: Identity — composite `id`, `saved_query_id` Required, RequiresReplace

`saved_query_id` is **Required** with `RequiresReplace` and is the API lookup key (`GET/PUT/DELETE /api/osquery/saved_queries/{id}`). State `id` is **Computed** and follows the repo space-aware composite pattern: `<space_id>/<saved_query_id>` (e.g. `"default/list_all_processes"`). `GetID()` returns the composite; `GetResourceID()` returns `saved_query_id` for API calls (entitycore `resolveKibanaResourceIdentity` parses the composite when present).

Kibana does **not** generate an ID when `id` is omitted on create (see task 1.3 discovery evidence).

**Why:** RequiresReplace prevents a silent rename, which would orphan the old query. Optional+Computed was rejected after discovery: no UUID auto-generation in server or UI.

**Discovery evidence (task 1.3):** Kibana `create_saved_query_route.ts`; `create_saved_query.gen.ts` (`id: SavedQueryId.optional()` with no default); `use_saved_query_form.tsx` (user must supply ID). Live Kibana CRUD was not exercised (stack unavailable).

### Decision 3: Space support via `SpaceAwarePathRequestEditor`

`space_id` is Optional+Computed, defaults to `"default"`, RequiresReplace. All four kbapi calls in the kibanaoapi helper pass `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as a request editor. This injects `/s/{spaceID}` before the first `/api/` segment for non-default spaces, matching the existing pattern in `agentbuilder`, `entity_store`, `alerting_rule`, `dashboards`, `synthetics`. No `transform_schema.go` changes needed.

### Decision 4: Composite import ID `<space_id>/<saved_query_id>`

Import accepts `"<space_id>/<saved_query_id>"` (e.g., `"default/list_all_processes"`). Prefer **`ImportStatePassthroughID`** on `id` (same pattern as `slo`, `security_detection_rule`, `dashboard`) — entitycore `resolveKibanaResourceIdentity` parses the composite on subsequent Read. Read/Create MUST populate `space_id`, `saved_query_id`, and composite `id` in state. If passthrough alone leaves Required `saved_query_id` unset before Read, use a thin custom `ImportState` (as in `alerting_rule`) to seed `space_id` and `saved_query_id` from the composite import string before Read.

### Decision 5: `platform` — set of strings, join to comma-string on write

`platform` is an Optional `schema.SetAttribute` of strings. Allowed values validator restricts to `"linux"`, `"darwin"`, `"windows"`. On write, the set is sorted and joined to a comma-separated string (e.g., `"linux,darwin"`). On read, the comma-string is split back to a set. Sorting ensures deterministic plan output.

### Decision 6: `ecs_mapping` — `MapNestedAttribute` with exactly-one-of-3 value validator

`ecs_mapping` is an Optional `MapNestedAttribute`. Each element is a `SingleNestedAttribute` with three fields:

- `field` (Optional string) — maps to a result column name in the query output
- `value` (Optional string) — static scalar value
- `values` (Optional set of string) — static array value

`ExactlyOneOfNestedAttrsValidator` from `internal/utils/validators` on `ecs_mapping` `MapNestedAttribute.NestedObject.Validators` enforces that exactly one of `field`, `value`, or `values` is set per map element. Maps to the generated `{Field, Value: string|[]string}` shape: `field` → `{field: "..."}`, `value` → `{value: "..."}` (string), `values` → `{value: [...]}` (array).

Partial ECS mapping precedent exists at `internal/kibana/security_detection_rule/models_to_api_type_utils.go:827` (`buildEcsMappingFromModel`) but covers only the `field` case. This resource must handle all three.

**Discovery (task 1.4):** `resourcevalidator.ExactlyOneOf` is resource-level only. `objectvalidator.ExactlyOneOf` is unsuitable on nested objects — it counts the parent object in path resolution (archived change `2026-05-11-expose-lens-chart-presentation-fields`). **Plan:** attach `ExactlyOneOfNestedAttrsValidator` to `MapNestedAttribute.NestedObject.Validators`. Precedent exists on nested/list objects (`internal/kibana/dashboard/panelkit/schema.go` list-item validators) but **not yet directly proven on MapNestedAttribute map values** — task 4.5 validates; **fallback:** custom inline `ValidateObject` only if map nested validation fails during implementation.

### Decision 7: `interval` — `Int64Attribute`

`interval` is Optional `Int64Attribute` (seconds). On read from Create/GET responses, the `json.RawMessage` union is read via `AsXxx0()/AsXxx1()` accessors; if the int arm fails, the string arm is parsed as `int64`. Update response uses the same union type for `interval`. On write, the `int64` is stringified before sending to the API. Nullable (omit from request body when not set).

### Decision 8: `version` — `StringAttribute`

`version` is Optional `StringAttribute`. On read from Create/GET responses, the `json.RawMessage` union is stringified (string accessor; fallback to `fmt.Sprintf` if the int arm). Update response types `version` as plain `*string` — dereference directly without union accessors. On write, the string is sent verbatim. Nullable (omit from request body when not set).

### Decision 9: `snapshot` and `removed` — Optional+Computed, no static default

`snapshot` and `removed` are Optional+Computed booleans with no static default. The API decides on create; on subsequent reads the server value is preserved in state. This avoids Terraform generating spurious diffs if the API defaults change.

### Decision 10: Prebuilt-query protection — runtime error diagnostic

When `prebuilt == true` in the API response (on Read or post-Create Read), the resource returns an error diagnostic:

> `"Prebuilt Osquery saved queries are managed by the osquery_manager integration package and cannot be managed by this resource. Use the elasticstack_kibana_osquery_saved_query data source to read this query."`

This is a runtime guard (not plan-time) because `prebuilt` is server-returned and unknown at plan time. The resource does not expose a `prebuilt` attribute in state. Affects Read and the read-after-write in Create.

### Decision 11: Update is PUT (full replacement of managed fields)

The API uses PUT (full replacement, not PATCH). On Update, the provider sends the **managed field set from plan/state** — `query`, `description`, `platform`, `interval`, `version`, `snapshot`, `removed`, `ecs_mapping` — omitting server-managed fields (`created_at`, `updated_at`, `created_by_profile_uid`, `updated_by_profile_uid`, `saved_object_id`). Optional attributes that are null/unset in plan omit the corresponding JSON keys (same nullable semantics as Create). The Update API response wraps `.Data`; the kibanaoapi helper returns an unwrapped `OsquerySavedQueryUpdateEntity` for model mapping.

### Decision 12: Delete returns empty body — `HandleStatusResponse`

Delete returns an empty body with HTTP 200 on success, and the provider should treat HTTP 404 as idempotent success. The kibanaoapi wrapper uses `HandleStatusResponse(..., http.StatusOK)` with a 404 no-op, matching the `maintenance_window` pattern.

### Decision 13: Response bodies wrap in `data` — unwrap in kibanaoapi; map in `populateFromAPI`

Create, GET, and Update API responses wrap the entity in a `data` field. The kibanaoapi helper (task 2) unwraps `.Data` and returns typed entities: `OsquerySavedQueryCreateEntity`, `OsquerySavedQueryGetEntity`, and `OsquerySavedQueryUpdateEntity`. Create and GET share union semantics for `interval`/`version` but use distinct kbapi generated union types, so they remain separate entity types rather than a single consolidated read type. `populateFromAPI` (task 3.4) consumes those entities and maps to the Terraform model. Response field typing differs by operation — see discovery note 1.1 and Decisions 7–8.

### Decision 14: Server-managed fields not exposed

`created_at`, `updated_at`, `created_by_profile_uid`, `updated_by_profile_uid`, and `saved_object_id` are server-managed and NOT exposed as attributes. Computed `id` stores the composite `<space_id>/<saved_query_id>`.

### Decision 15: Data source — single-item GET-by-id, prebuilt-safe, version-gated

The data source accepts `saved_query_id` (Required), `space_id` (Optional, default `"default"`), and `kibana_connection` (Optional). It calls GET by ID and populates all managed fields. Unlike the resource, it does NOT error on `prebuilt == true`. The data source model (or shared base) implements `GetVersionRequirements` with the same `8.5.0` conservative floor as the resource.

### Decision 16: Minimum version — `8.5.0` (documented/conservative floor)

The resource declares `8.5.0` as the minimum Kibana version via `GetVersionRequirements` (implemented in task 3.2). This is the documented/conservative floor from Kibana API docs and source — not live-validated during discovery.

**Discovery evidence (task 1.2):** Osquery saved-queries CRUD is documented under Kibana v8 API reference (`POST/GET/PUT/DELETE /api/osquery/saved_queries`); Kibana PR [#137162](https://github.com/elastic/kibana/pull/137162) (Osquery API docs) is labeled `v8.5.0`; the Osquery plugin public API version is `2023-10-31` (`API_VERSIONS.public.v1` in Kibana `common/constants.ts`); all four CRUD bindings are present in `generated/kbapi/kibana.gen.go`. Live confirmation deferred to acceptance task 7.9; Kibana stack was unavailable locally.

### Decision 17: Naming

Resource and data source: `elasticstack_kibana_osquery_saved_query`. Go package: `internal/kibana/osquery_saved_query`. Registration in `provider/plugin_framework.go` alongside `maintenance_window` and other Kibana resources.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **`saved_query_id` omitted on create — does Kibana generate a UUID?** | Resolved (task 1.3): no — `saved_query_id` is Required. |
| **`8.5.0` version floor is too low** | Documented/conservative floor from docs/Kibana source (task 1.2); live confirmation in acceptance task 7.9. |
| **`ecs_mapping` exactly-one-of validator inside `MapNestedAttribute`** | Resolved (task 1.4): use `ExactlyOneOfNestedAttrsValidator` on `NestedObject.Validators`; task 7.6 validates behavior. |
| **`interval`/`version` union-type edge cases** | Unit tests in task 3.5 cover union arms and Update plain `*string` version; acceptance round-trip in task 7.1 confirms no data loss. |
| **Prebuilt queries silently imported** | Runtime guard on Read (and post-Create) returns an explicit error diagnostic before touching state. |

## Open Questions

<!-- All task-1 discovery items resolved; see Decisions 2, 6, and 16. -->

## Discovery notes (task group 1)

### 1.1 kbapi CRUD bindings verified

All four methods exist in `generated/kbapi/kibana.gen.go` with signatures matching this design:

| Method | Request body | Response `JSON200` type | `data` wrapper | kibanaoapi entity | `interval` / `version` typing |
|---|---|---|---|---|---|
| `OsqueryCreateSavedQuery` | `SecurityOsqueryAPICreateSavedQueryRequestBody` (`Id *SecurityOsqueryAPISavedQueryId`, optional) | `SecurityOsqueryAPICreateSavedQueryResponse` | yes | `OsquerySavedQueryCreateEntity` | both unions (`AsXxx0()/AsXxx1()`) |
| `OsqueryGetSavedQueryDetails` | path `id` only | `SecurityOsqueryAPIFindSavedQueryDetailResponse` | yes | `OsquerySavedQueryGetEntity` | both unions |
| `OsqueryUpdateSavedQuery` | `SecurityOsqueryAPIUpdateSavedQueryRequestBody` | `SecurityOsqueryAPIUpdateSavedQueryResponse` | yes | `OsquerySavedQueryUpdateEntity` | `interval` union; `version` plain `*string` |
| `OsqueryDeleteSavedQuery` | path `id` only | `SecurityOsqueryAPIDefaultSuccessResponse` (empty map) | n/a | n/a |

Create request uses JSON field `id` (maps to Terraform `saved_query_id`). ECS mapping item type is `SecurityOsqueryAPIECSMappingItem` with `{Field *string, Value *SecurityOsqueryAPIECSMappingItem_Value}` union (`AsSecurityOsqueryAPIECSMappingItemValue0/1` for string|[]string).
