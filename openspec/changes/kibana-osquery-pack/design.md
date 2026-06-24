## Context

Kibana exposes Osquery packs under `/api/osquery/packs` with full CRUD plus a list endpoint. The generated `kbapi` client (at the pinned OAS ref) already exposes four pack CRUD operations in `generated/kbapi/kibana.gen.go`:

- `OsqueryCreatePacks` / `OsqueryCreatePacksWithResponse`
- `OsqueryGetPacksDetails` / `...WithResponse`
- `OsqueryUpdatePacks` / `...WithResponse`
- `OsqueryDeletePacks` / `...WithResponse`

However, the pinned client is **stale**: it is missing scheduling fields (`schedule_type`, `rrule_schedule`, per-query `interval`/`timeout`/`splay`) exposed by the live API. Bumping the OAS ref to a commit that includes these fields (first merged in Kibana PR #270639, commit `9dc7627253d0`, label `v9.5.0`) is **blocked** as of this discovery pass: `transform_schema.go` panics on new Fleet API response shapes when running `make -C generated/kbapi all` against main. **v1 implements packs CRUD with no scheduling attributes** using the pinned client until kbapi regeneration is unblocked (separate transform fix + OAS bump).

The Osquery paths are not in the `spaceIdPaths` allow-list in `generated/kbapi/transform_schema.go`, so generated URLs lack `/s/{spaceId}`. Space support is injected at request time via `kibanautil.SpaceAwarePathRequestEditor(spaceID)`, the same mechanism used by `agentbuilder`, `entity_store`, `alerting_rule`, `dashboards`, and `synthetics` — no `transform_schema.go` changes required.

Non-trivial implementation areas:

1. **kbapi client regeneration (blocked)**: OAS refs at or after `9dc7627253d0` include scheduling fields, but `transform_schema.go` must be updated for new Fleet response shapes before regeneration succeeds. Until then, v1 uses the pinned client with **no scheduling attributes** in schema or API mapping.
2. **`queries` as MapNestedAttribute**: each key is the query name; value holds SQL, platform, ECS mapping, and other per-query options supported by the pinned kbapi types. The inner `id` field is not exposed (map key is canonical).
3. **Scheduling (deferred)**: pack-level `schedule_type`, `interval`, `rrule_schedule`, and per-query scheduling overrides documented in Kibana main OAS (PR #270639, `v9.5.0`) — not in v1.
4. **`shards` format mismatch**: create/update **request** uses `map[string]float32` (`SecurityOsqueryAPIShards`); create **response** OAS declares an array `[{key, value}]` (generated client: `*[]struct{Key, Value *float32}`) while create response **examples** and GET/Update use map form; `GetPacksDetails` returns `*map[string]float32`. Read normalizes to `map(string → number)` in state; create-response array form is normalized on read after follow-up GET or treated as map when empty.
5. **`ecs_mapping` modelling**: same three-way `field`/`value`/`values` constraint as the sibling `osquery_saved_query` resource.
6. **Prebuilt pack protection**: `read_only=true` appears on GET detail responses only (not on Create POST response). Enforced as runtime diagnostic on Read, read-after-write (GET following Create/Update), and Import refresh — not by inspecting the Create POST body.

## Goals / Non-Goals

**Goals (v1):**
- Full lifecycle (Create, Read, Update, Delete) of user-managed Osquery packs with import support.
- Single-item data source for read-only lookup (prebuilt-safe).
- Space-awareness via composite `<space_id>/<pack_id>` import ID (`pack_id` = `saved_object_id` UUID).
- Faithful `ecs_mapping` and `shards` representation using pinned kbapi types.

**Non-Goals (v1):**
- Scheduling fields (`schedule_type`, `interval`, `rrule_schedule`, per-query `interval`/`timeout`) — deferred until kbapi regeneration.
- kbapi regeneration in this change (blocked by `transform_schema.go`).

**Deferred goals (follow-up):**
- Full scheduling model including `interval` mode and `rrule_schedule` mode with provider-side validation.
- kbapi regeneration to include modern scheduling fields.

**Other non-goals:**
- `elasticstack_kibana_osquery_saved_query` resource/data source (separate change: `kibana-osquery-saved-query`).
- Osquery live queries — ephemeral, not suitable for Terraform.
- Plural list data source (`elasticstack_kibana_osquery_packs`) — deferred (list endpoint returns different `ecs_mapping`/`shards` format than the detail endpoint).
- Osquery response actions within security detection rules.

## Decisions

### Decision 1: Resource pattern — `entitycore.KibanaResource[Model]`

Implement the resource as an `entitycore.KibanaResource[osqueryPackModel]`, matching `maintenance_window`, `slo`, `spaces`, `alerting_rule`, and other recent Kibana resources. The generic wrapper handles `kibana_connection`, timeouts, and configure. A thin `internal/clients/kibanaoapi/osquery_pack.go` helper wraps the four kbapi calls.

**Why:** The pattern is standard for all new Kibana resources in this provider. The sibling `osquery_saved_query` is the closest analogue. Considered: standalone `resource.Resource` without entitycore — rejected for inconsistency.

### Decision 2: Identity — `pack_id` Computed from `saved_object_id`

`pack_id` is **Computed-only** (maps to API `saved_object_id`). The Create request body does **not** accept a client-supplied pack ID (`SecurityOsqueryAPICreatePacksRequestBody` has no `pack_id`/`id` field; Kibana `create_pack_route` calls `spaceScopedClient.create()` without an explicit ID, yielding a server-generated UUID). The path parameter `{id}` for GET/PUT/DELETE is this same `saved_object_id`. `id` in state mirrors `pack_id`. Import uses `pack_id` as the lookup key.

**Confirmed (task 1.4)**: Kibana always generates a UUID for `saved_object_id` on Create; user cannot supply an ID. Resource spec requires Computed-only `pack_id`.

### Decision 3: Space support via `SpaceAwarePathRequestEditor`

`space_id` is Optional+Computed, defaults to `"default"`, RequiresReplace. All four kbapi calls in the kibanaoapi helper pass `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as a request editor. No `transform_schema.go` changes needed.

### Decision 4: Composite import ID `<space_id>/<pack_id>`

Import accepts `"<space_id>/<pack_id>"` (e.g., `"default/3c42c847-eb30-4452-80e0-728584042334"` where `pack_id` is the API `saved_object_id`). Matches the import idiom of every other space-aware Kibana resource in the provider.

### Decision 5: `queries` as MapNestedAttribute

Map key is the query name (canonical in Kibana; inner `id` field NOT exposed). Each element is a `SingleNestedAttribute`. See spec for the full field list.

**Why:** Map key is the natural query identifier in Kibana. Exposing the inner `id` would duplicate it and create confusion.

### Decision 6: Scheduling (deferred — post-kbapi bump)

**v1:** No scheduling attributes in resource or data source schema. Pinned `SecurityOsqueryAPICreatePacksRequestBody` and `SecurityOsqueryAPIObjectQueriesItem` omit `schedule_type`, pack-level `interval`, `rrule_schedule`, and per-query `interval`/`timeout`. v1 manages pack metadata, policy assignment, shards, and query definitions (SQL, platform, ECS mapping) only.

**Deferred target state** (post-kbapi bump, Kibana ≥ 9.5.0): pack-level `schedule_type` (`"interval"` | `"rrule"`), exactly-one-of `interval`/`rrule_schedule`, per-query scheduling overrides, and cross-mode ConfigValidators. See Decision 7 for `rrule_schedule` shape; interval Int64 normalization follows the `osquery_saved_query` pattern.

**Why deferred:** OAS bump blocked by `transform_schema.go` incompatibility with Fleet response shape changes in main OAS.

### Decision 7: `rrule_schedule` schema shape (deferred — post-kbapi bump)

```
SingleNestedAttribute (optional at pack/per-query level):
  rrule      StringAttribute           required; validator: must start with FREQ=
  start_date timetypes.RFC3339         required; schedule anchor
  end_date   timetypes.RFC3339         optional; validator: > start_date
  splay      customtypes.DurationType  optional; validator: ≤ 12h (43200s)
  timeout    Int64Attribute            optional; semantics: seconds
```

Both `timetypes.RFC3339` (from `terraform-plugin-framework-timetypes` v0.5.0; used in `security_exception_item`, `ml/calendar_event`) and `customtypes.DurationType` (`internal/utils/customtypes/duration_type.go`) are established in this codebase.

### Decision 8: `shards` — `map(string → number)`, normalize on read

**Wire formats (confirmed task 1.6):**
- Create/Update **request**: `map[string]float32` (`SecurityOsqueryAPIShards`); Kibana io-ts uses `t.record(t.string, toNumberRt)` — **not** an array on write.
- Create **response** (OAS/generated): array `[{key, value}]` (`*[]struct{Key *string; Value *float32}` in pinned client); OAS examples show map — treat array as create-response-only quirk.
- Get (`FindPackResponse`) / Update response: `*map[string]float32`.

Read path uses `GetPacksDetails` map form as canonical for state. Int-vs-Number precision (float32 semantics are integer percent 1–100) deferred to implementation spike.

### Decision 9: Prebuilt pack protection — runtime error diagnostic

When `read_only=true` on a GET detail response, return an error diagnostic:
> `"This Osquery pack is read-only (prebuilt) and cannot be managed by this resource. Use the elasticstack_kibana_osquery_pack data source to read this pack."`

The Create POST response does **not** include `read_only`. The resource guard runs on Read, read-after-write (GET following Create/Update to populate detail fields), and Import refresh — not by inspecting the Create POST body directly.

The data source does NOT error on `read_only=true`.

**Why:** Consistent with `osquery_saved_query` design — no lever to pull for prebuilt packs, so don't pretend to manage them.

### Decision 10: `ecs_mapping` — MapNestedAttribute, three-way exactly-one-of

Identical to the `osquery_saved_query` resource: `field` (Optional string), `value` (Optional string), `values` (Optional set of strings), with `ConfigValidator` enforcing exactly one of the three per element.

### Decision 11: Minimum version — base `8.5.0` + scheduling `9.5.0`

| Capability | Minimum Kibana | Source |
|---|---|---|
| Base packs CRUD (`/api/osquery/packs`) | **8.5.0** | `create_pack_route` present in `v8.5.0` |
| Full scheduling (`schedule_type`, `rrule_schedule`, pack-level `interval`) | **9.5.0** | Kibana PR #270639 (merged 2026-05-28, label `v9.5.0`); `rruleScheduling` experimental flag default `false` until flag-flip PR |

`GetVersionRequirements` (task 3.2): v1 registers **8.5.0** only. Second scheduling-floor entry (`9.5.0`) deferred to post-kbapi-bump follow-up.

### Decision 12: Plural list data source deferred

The list endpoint returns a different `ecs_mapping`/`shards` format than `GetPacksDetails`. Deferring avoids a separate normalization path and keeps this PR focused.

### Decision 13: `platform` per-query — SetAttribute of strings

Per-query optional `SetAttribute` of strings with allowed-values validator (`"linux"`, `"darwin"`, `"windows"`). On write, sorted and joined to comma-separated string. On read, split back to set.

### Decision 14: Response shape — `data` wrapper (confirmed)

**Create (task 1.5):** Response wraps the pack in a `data` field. Type: `SecurityOsqueryAPICreatePacksResponse` with embedded `Data struct { ... } \`json:"data"\``. Unwrap `response.JSON200.Data` on create. Required fields in `data`: `saved_object_id`, `name`. Create `data` omits `read_only`.

**Get / Update detail responses:** `OsqueryGetPacksDetails` and update responses use typed wrappers whose detail payload is at `response.JSON200.Data` (`SecurityOsqueryAPIFindPackResponse.Data`). Unwrap `.Data` before calling `populateFromAPI`.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **kbapi OAS bump blocked by Fleet transform** | Task 1.2 confirmed; v1 has no scheduling attributes; fix `transform_schema.go` + bump to ≥ `9dc7627253d0` in follow-up. |
| **`shards` Int-vs-Number precision (float32 → int64)** | Deferred to implementation spike; use `NumberAttribute` if float32 precision matters. |
| **rrule_schedule validator complexity** | Shallow regex (must start with `FREQ=`); defer deep RFC 5545 validation to API. |
| **Per-query schedule override cross-mode validation** | Provider-side ConfigValidator at plan time; API also returns 400 on mismatch. |
| **Prebuilt packs silently imported** | Runtime guard on Read, read-after-write GET, and Import refresh returns explicit error diagnostic. |

## Open Questions

1. **kbapi regeneration unblock**: Fix `transform_schema.go` for `$ref`-wrapped Fleet responses, then bump to ≥ `9dc7627253d0` and add scheduling scope.
2. **`rruleScheduling` flag**: RRULE scheduling requires experimental flag until Kibana flag-flip PR; document in provider docs / acceptance test skip logic when scheduling lands.
3. **`shards` Int-vs-Number**: Use `Int64Attribute` or `NumberAttribute`? Depends on whether the API round-trips fractional values from other clients.
