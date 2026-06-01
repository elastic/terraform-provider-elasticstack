## 1. Specs

- [x] 1.1 Keep delta specs aligned with `proposal.md` and `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-security-entity-store-entity --type change` (or `make check-openspec`) before opening a PR.
- [ ] 1.2 Resolve design open question OQ-1: confirm minimum Kibana version for individual entity CRUD endpoints; update delta specs and `EnforceMinVersion` call if the floor differs from `9.1.0`.
- [ ] 1.3 Resolve design open question OQ-2: verify KQL filter expression `entity.id:"<id>"` works reliably for all valid entity ID formats; document escaping strategy in the resource read implementation.
- [ ] 1.4 Resolve design open question OQ-4: add a plan-time validator that checks consistency between `entity_id` and `entity.id` inside the typed `entity` block (or `entity_json`).
- [ ] 1.5 On completion of implementation, sync or archive this change per project workflow.

## 2. Project structure

- [x] 2.1 Create package `internal/kibana/security_entity_store/` with subdirectories for resource and data source files.
- [x] 2.2 Register the resource and data source in the provider's resource/data source lists (follow the existing registration pattern in `internal/provider/`).

## 3. Shared types and API client helpers

- [x] 3.1 Define local response structs for entity store API responses (create/update, delete, list) in `internal/kibana/security_entity_store/models.go` — do not use the kbapi union types (`PostSecurityEntityStoreEntitiesEntitytypeJSONBody0`, etc.) directly in the provider model.
- [x] 3.2 Implement a `readEntities` helper that calls `GET /api/security/entity_store/entities` with a KQL filter for `entity.id` and optional `entity_types`, returning the first matching record or a not-found signal.
- [x] 3.3 Implement canonical JSON normalization (sorted keys) for computed `document_json`, `response_json`, and all `*_json` attributes to minimize false plan diffs.

## 4. Resource: `elasticstack_kibana_security_entity_store_entity`

- [x] 4.1 Implement `schema.go` with the full attribute set:
  - Identity: `id` (computed, composite), `space_id` (optional, computed, RequiresReplace), `entity_type` (required, RequiresReplace), `entity_id` (required, RequiresReplace).
  - Timestamp: `timestamp` (optional string, maps to `@timestamp`).
  - Typed blocks: `entity`, `host`, `user`, `service`, `cloud`, `asset`, `orchestrator`, `event` (each optional nested block).
  - Labels: `labels` (optional map of string).
  - Tags: `tags` (optional set of string).
  - JSON fallbacks: `entity_json`, `host_json`, `user_json`, `service_json`, `cloud_json`, `asset_json`, `orchestrator_json`, `event_json`, `labels_json` (each optional string, conflicts with typed counterpart).
  - Update control: `force` (optional bool, default false).
  - Computed outputs: `document_json`, `response_json`.
  - `kibana_connection` block (injected by envelope).
- [x] 4.2 Implement `ConflictsWith` validators between each typed block and its `_json` fallback in schema.
- [x] 4.3 Implement plan-time validator for `entity_id` vs `entity.id` consistency (OQ-4 / task 1.4).
- [x] 4.4 Implement `Create` callback:
  - Build POST request body from typed blocks (merged with JSON fallbacks where typed block absent).
  - Call `POST /api/security/entity_store/entities/{entity_type}`.
  - Treat HTTP 200 as success; treat HTTP 409 as a create error with a descriptive diagnostic.
  - After create, invoke Read to populate state (including computed `document_json`) using an authoritative read-after-write pattern.
- [x] 4.5 Implement `Read` callback:
  - Call `readEntities` helper with KQL filter for `entity.id` and `entity_types=[entity_type]`.
  - On not-found, remove resource from state (standard envelope behavior).
  - Map response fields back to typed attributes and set computed outputs `document_json` and `response_json` from normalized response payloads.
- [x] 4.6 Implement `Update` callback:
  - Build PUT request body from typed blocks / JSON fallbacks.
  - Pass `?force=true` when `force` attribute is true.
  - Call `PUT /api/security/entity_store/entities/{entity_type}`.
  - After update, invoke Read to populate state (authoritative read-after-write).
- [x] 4.7 Implement `Delete` callback:
  - Call `DELETE /api/security/entity_store/entities/` with JSON body `{"entityId": "<entity_id>"}`.
- [x] 4.8 Implement `ImportState` — split composite `<space_id>/<entity_id>` to populate `space_id` and `entity_id` in state; trigger read.
- [x] 4.9 Add `EnforceMinVersion("9.1.0")` in the resource model's `GetVersionRequirements()` (or equivalent mechanism); adjust version per OQ-1 / task 1.2.

## 5. Data source: `elasticstack_kibana_security_entity_store_entities`

- [x] 6.1 Implement schema with:
  - `space_id` (optional, computed).
  - Single-entity convenience: `entity_id` (optional string; conflicts with `filter` and `filter_query`).
  - Search-after mode: `filter`, `size`, `search_after`, `source`, `fields`.
  - Page mode: `sort_field`, `sort_order` (enum: asc/desc), `page`, `per_page`, `filter_query`.
  - Common: `entity_types` (optional set of string).
  - Computed: `results_json`.
- [x] 6.2 Implement plan-time validation that rejects mixing search-after and page-mode parameters, and rejects `entity_id` combined with `filter` or `filter_query`.
- [x] 6.3 Implement `Read` callback:
  - Build `GetSecurityEntityStoreEntitiesParams` from configured attributes.
  - When `entity_id` is set, generate `filter = entity.id:"<entity_id>"` and do not allow a user-supplied `filter`/`filter_query`.
  - Call `GET /api/security/entity_store/entities`.
  - Serialize full response to normalized JSON and set `results_json`.
- [x] 6.4 Add `EnforceMinVersion("9.1.0")` in the data source model's `GetVersionRequirements()`.

## 7. Testing

- [x] 7.1 Add acceptance test for resource lifecycle of a `host` entity: create with typed `entity` + `host` blocks → plan shows no diff → update `host.ip` → destroy.
- [x] 7.2 Add acceptance tests for `user`, `service`, and `generic` entity types with stable test data.
- [x] 7.3 Add acceptance test for typed `entity` block vs `entity_json` fallback: verify both produce the same API result and that using both together produces a plan-time error.
- [x] 7.4 Add acceptance test for the `force` flag on update (if a protected-field scenario is available in the test environment).
- [x] 7.5 Add acceptance test for `import`: create entity via resource → `terraform import` using composite ID → verify state matches.
- [ ] 7.6 Add acceptance test for single-entity lookup via the list data source `entity_id` filter: create entity via resource → read list data source with `entity_id` → assert `results_json` contains exactly one record with the expected `entity.id`.
- [ ] 7.7 Add acceptance test for the list data source in page mode: verify `results_json` is non-empty when entities exist; verify plan error when page-mode and search-after parameters are combined.
- [ ] 7.8 Add acceptance test for `entity_id` conflict validation: verify plan error when `entity_id` is combined with `filter` or `filter_query`.
- [ ] 7.9 Add unit tests for composite ID construction and parsing (encode/decode of `<space_id>/<entity_id>`).
- [x] 7.10 Add unit tests for canonical JSON normalization to guard against false diffs.
- [x] 7.11 Add unit tests for the pagination-mode exclusivity validator and `entity_id` conflict validator.
- [x] 7.12 Gate all acceptance tests with `SkipIfVersionConstraintNotMet("9.1.0")` (or project-equivalent helper) so they are skipped when the test Elastic Stack is below the minimum version.

## 8. Verify

- [x] 8.1 `make build` passes.
- [ ] 8.2 Acceptance tests pass for all new entities.
- [x] 8.3 `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-security-entity-store-entity --type change` passes.
- [ ] 8.4 Update delta specs if implementation reveals discrepancies from the original spec; then sync or archive the change.
