## Context

`internal/elasticsearch/index/index` manages Elasticsearch indices with a large, complex schema (static settings, dynamic settings, aliases, mappings, analysis settings, deletion protection, and `use_existing` adoption). It currently embeds `*entitycore.ResourceBase` and implements standard `Schema`, `Create`, `Read`, `Update`, `Delete`, and `ImportState`.

The resource has three mechanics that make its Create and Update non-standard:

1. **`use_existing` adoption** — On create, if `use_existing` is true and the configured name is a static (non-date-math) index name, the resource checks if the index already exists. If so, it adopts the existing index by reconciling aliases, settings, and mappings instead of calling the Create Index API.
2. **Date-math names** — Index names can be Elasticsearch date-math expressions. The resource URI-encodes them for create and stores the resolved concrete name separately from the configured expression.
3. **Update derives identity from state** — Update reads the concrete index name from the current state (via the composite `id`), not from the plan, because the configured `name` may be a date-math expression while the actual index has a resolved concrete name.

## Goals / Non-Goals

**Goals:**

- Migrate to the entitycore envelope while preserving all existing behavior.
- Keep `use_existing`, date-math, and concrete-name logic intact.
- Keep deletion protection intact.
- Keep mappings plan modifiers and semantic equality intact.

**Non-Goals:**

- Changing the schema or acceptance tests.
- Changing the adoption or reconciliation logic.

## Decisions

### D1. Override Create and Update; envelope handles Read, Delete, Schema

**Choice:** Pass placeholder write callbacks for create and update, and define `Create` and `Update` on `Resource`.

**Rationale:** Both Create and Update need plan/state handling that exceeds the envelope callback contract. Create needs `use_existing` and adoption logic. Update needs the prior state's concrete index identity.

### D2. Read callback delegates to existing readIndex helper

**Choice:** The `readIndex` helper already has the shape needed for a callback. Refactor it to accept `*clients.ElasticsearchScopedClient` as a parameter and return `(tfModel, bool, diag.Diagnostics)`.

**Rationale:** `readIndex` is well-factored and can be called both as the envelope's read callback and from the Create override's read-after-adoption logic.

### D3. Delete callback checks deletion_protection

**Choice:** Move the deletion-protection check from the receiver method into a package-level `deleteIndex` callback.

**Rationale:** Deletion protection is a straightforward pre-flight check on the model. It fits cleanly in a callback and lets the envelope own the Delete prelude.

### D4. Schema factory strips connection block

**Choice:** The schema factory returns `schema.Schema` without `elasticsearch_connection`.

**Rationale:** Standard envelope convention. The schema is large (~40 dynamic settings, aliases, mappings, etc.) but the change is mechanical.

### D5. GetResourceID returns configured name

**Choice:** `tfModel.GetResourceID()` returns the configured `Name` attribute.

**Rationale:** For create, the write identity is the configured index name (which may be a date-math expression). The envelope passes this to the create callback, which then decides whether to adopt or create.

## Risks / Trade-offs

- **Risk:** The Create override for adoption is complex (~80 lines). **Mitigation:** Keep the adoption logic in its existing helper (`adoptExistingIndexOnCreate`) and call it from the override.
- **Risk:** Update's reliance on state-derived concrete name could be accidentally broken if the envelope's Update prelude changes. **Mitigation:** The override decodes `req.State` directly, so it is self-contained.
- **Trade-off:** Read and Delete become callbacks but Create/Update remain overrides. This is a mixed pattern, but it correctly reflects which operations are standard and which are bespoke.

## Migration Plan

1. Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` to `tfModel`.
2. Convert `Schema` receiver to a package-level factory omitting `elasticsearch_connection`.
3. Refactor `readIndex` to the callback signature.
4. Extract `deleteIndex` callback that checks `deletion_protection` and calls `elasticsearch.DeleteIndex`.
5. Replace `*ResourceBase` with `*entitycore.ElasticsearchResource[tfModel]` in the struct.
6. Keep `Create` and `Update` receiver methods. Remove `Read` and `Delete` receiver methods.
7. Keep `ImportState` unchanged.
8. Run `make build`, `make check-lint`, `make check-openspec`, and acceptance tests.

## Open Questions

None.
