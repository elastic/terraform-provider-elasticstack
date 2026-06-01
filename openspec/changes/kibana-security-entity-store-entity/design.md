## Context

The Kibana Security Entity Store (`/api/security/entity_store/entities/...`) exposes CRUD APIs for individual entity records (user, host, service, generic) stored in the latest Entity Store Elasticsearch index. It also provides a list/search endpoint for querying entities with page-based or cursor-based pagination. There is no `GET /entities/{id}` endpoint — reads must use the list endpoint with an entity-specific filter.

The generated Kibana client (`generated/kbapi/kibana.gen.go`) provides request builders for these operations but returns `[]byte` (raw response bodies) rather than typed response structs. Implementation must parse responses locally.

This change introduces:
- A managed resource (`elasticstack_kibana_security_entity_store_entity`) that owns the full lifecycle of a single entity record.
- A list/search data source (`elasticstack_kibana_security_entity_store_entities`) that exposes the full query capability of the list endpoint, including an optional `entity_id` filter for single-entity lookup.

Both entities are Kibana-backed and follow the Kibana Plugin Framework envelope (`NewKibanaResource[T]`, `NewKibanaDataSource[T]`) and `KibanaResourceModel`/`KibanaDataSourceModel` interface contracts.

## Goals / Non-Goals

**Goals:**
- Provide full CRUD + import for a single entity record with typed attributes covering the complete API body.
- Provide JSON escape-hatch fallbacks (`entity_json`, `host_json`, etc.) for fields whose typed schema would be impractically large or that need non-string value support.
- Enforce `ConflictsWithValidator` between each typed block and its JSON fallback at plan time.
- Use the list endpoint for Read, applying the most deterministic KQL filter available.
- Expose the full list/search parameter surface in the list data source, covering both pagination modes and an optional `entity_id` single-record lookup.
- Version-gate all entities at Elastic Stack ≥ 9.1.0 via `EnforceMinVersion`.

**Non-Goals:**
- Bulk entity upsert (`PutSecurityEntityStoreEntitiesBulk`) — not within scope; the bulk endpoint is for high-volume operations, not single-record management.
- Managing the Entity Store engine or index configuration (separate concern).
- Supporting write-only / sensitive fields in the initial schema: the API does not define credential fields for entity records; if future typed schemas expose credential material, those must be marked `Sensitive`, `WriteOnly`, and tracked with `internal/utils/writeonlyhash`.

## Decisions

### 1. Composite ID: `<space_id>/<entity_id>`

The resource and data source both derive their Terraform state `id` from `<space_id>/<entity_id>`. `entity_id` is user-assigned and stable; it maps to the `entity.id` field in the API body. `space_id` defaults to `"default"`.

`entity_type` is NOT included in the composite ID because an entity's `entity.id` is globally unique within a space across all entity types. Including `entity_type` would be redundant and would break import ergonomics.

*Alternative considered:* `<space_id>/<entity_type>/<entity_id>`. Rejected — the API uses `entity.id` as the primary lookup key in the delete body and is the most discriminating field for the read filter; embedding `entity_type` in the composite ID adds complexity without gain.

### 2. Read via list endpoint with KQL filter

Since `GET /api/security/entity_store/entities/{id}` does not exist, Read uses `GET /api/security/entity_store/entities?filter=entity.id:"<entity_id>"` with `entity_types=[entity_type]` when `entity_type` is known. If no entity is returned, the resource removes itself from state (standard Kibana envelope `found == false` behavior). If multiple results are returned (should not occur given unique IDs), the implementation returns an error diagnostic.

### 3. Typed blocks + JSON fallbacks with conflict enforcement

Each top-level API section (`entity`, `host`, `user`, `service`, `cloud`, `asset`, `orchestrator`, `labels`) has two forms:
- A typed nested block / attribute with first-class attribute modeling.
- A `_json` string attribute accepting a JSON-encoded value for that section.

`ConflictsWith` / `objectvalidator.ConflictsWithAny` enforces mutual exclusion at plan time. Exactly one of the two forms must be set when that section is present.

The `document_json` attribute is computed-only and contains the assembled API document read back from Kibana on the latest read, for troubleshooting and drift inspection. Users cannot set `document_json` as input; it is not a JSON input form.

### 4. `entity_id` RequiresReplace

`entity_id` triggers replacement on change because the create and delete endpoints use the entity ID as the primary key: there is no rename/move operation in the API. Similarly, `entity_type` and `space_id` trigger replacement.

### 5. `force` parameter on update only

The `force` optional bool (default `false`) is passed as `?force=true` on `PUT` only. It is not sent on create or delete. This matches the API design where `force` bypasses protected-field guards on update.

### 6. Create returns HTTP 200, not 201

The create endpoint (`POST /api/security/entity_store/entities/{entityType}`) returns HTTP 200 on success. The implementation treats HTTP 200 as success for create, consistent with the API contract. HTTP 409 (entity ID already exists) is treated as a Terraform error during create. There is no automatic import-on-409 behavior; practitioners must use `terraform import` if they need to adopt existing entities.

### 7. Pagination mode exclusivity in the list data source

The list endpoint supports two pagination modes that cannot be combined:
- **Page mode**: `sort_field`, `sort_order`, `page`, `per_page`, `filter_query`.
- **Cursor/search-after mode**: `filter`, `size`, `search_after`, `source`, `fields`.

The list data source validates that parameters from both modes are not used simultaneously. The implementation adds a `RequiresPrerequisite`-style or custom validator that errors at plan time when parameters from both modes are set.

### 8. Response parsing

The kbapi generated client returns `[]byte` for entity store operation responses. Implementation defines local Go structs for each response shape (create, update, delete, list) and uses `encoding/json.Unmarshal` to decode them. This is consistent with the existing pattern for other kbapi endpoints that return raw bytes.

### 9. Version gating at 9.1.0

`EnforceMinVersion("9.1.0")` is set on the resource and data source models. This value is tentative; implementation should verify actual API availability in Kibana release notes and adjust if needed. The test config uses `CheckDestroyWithVersionConstraint` or `SkipIfVersionConstraintNotMet` helpers to skip acceptance tests when the target Elastic Stack version is too old.

## Open Questions

1. **Exact minimum version**: The issue proposes `9.1.0`. Is this confirmed from Kibana release notes for the individual entity CRUD endpoints? Acceptance testing should verify; if the API is available earlier, lower the version gate.
2. **KQL filter for entity.id read**: Does `entity.id:"<id>"` work without tokenization issues for all valid entity ID values (values containing colons, slashes, etc.)? If KQL tokenizes colons, the filter may need escaping or a different expression. Implementation should test edge cases.
3. **Multiple results from list on single-entity read**: Is `entity.id` guaranteed unique within a space+entity_type scope by the API? Or can two different entity types share the same `entity.id` string? If they can share the same ID across types, the read filter must always include `entity_types`.
4. **`entity.id` vs `entity_id` field**: The issue shows the `entity.id` nested within the `entity` block in the request body. When `entity_id` is provided as a top-level resource attribute, must it match `entity.id` inside the `entity` block? If yes, implement a plan-time validator that checks consistency between `entity_id` and the `entity.id` field in the typed `entity` block or the `entity_json` fallback.
5. **`labels_json` non-string values**: The issue notes that `labels_json` supports non-string values if the API accepts them. This needs API-level testing; if only string values are accepted, the typed `labels` map (string → string) is sufficient and the JSON fallback can be removed.

## Risks / Trade-offs

- **No direct GET endpoint**: Every resource read is a filtered list call. If the KQL filter has performance implications on large entity stores, reads may be slow. There is no alternative until Kibana adds a direct GET by ID endpoint.
- **Create returning 200**: Tooling that expects HTTP 201 for creates may be confused. The Kibana envelope handles this by checking response body content rather than status code alone; ensure the implementation does not error on 200 from POST.
- **Escape-hatch JSON drift**: JSON fallback attributes bypass typed plan validation. Drift in deeply nested fields (e.g., `entity.behaviors`, `entity.relationships`) may be invisible to `terraform plan` unless the computed `document_json` attribute is set correctly on every read. Implement a normalizer (canonical JSON sort) to minimize false diffs.
- **Large generated types**: The kbapi generator produces very large union types for the entity body (multiple `JSONBody0`, `JSONBody1` variants for each entity type). Do not use those union types in the provider model; define simpler local structs instead and validate only what the provider supports.
