## Context

`internal/elasticsearch/cluster/settings.go` implements `elasticstack_elasticsearch_cluster_settings` using the Plugin SDK (`schema.Resource`, `schema.ResourceData`). It supports `persistent` and `transient` blocks, each containing a set of `setting` entries with `name`, `value`, and `value_list`. Create, Update, and Delete all call the Elasticsearch Cluster Update Settings API. Read calls the Cluster Get Settings API with `flat_settings=true`. Update is unique because it compares the old and new configuration to null out settings that have been removed.

The entitycore resource envelope (`NewElasticsearchResource[T]`) already owns the standard PF resource prelude for Read, Schema, Configure, and Metadata. No envelope changes are required for this migration.

## Goals / Non-Goals

**Goals:**

- Rewrite the resource from SDK to PF while keeping the exact same schema shape and behavior.
- Migrate to the entitycore envelope so the resource no longer carries its own connection-block schema or Read/Delete prelude.
- Preserve the update logic that nulls out removed settings.
- Preserve the setting value/list type detection on read.
- Preserve import support.

**Non-Goals:**

- Changing the Terraform schema (adding, removing, or renaming attributes).
- Changing acceptance tests.
- Migrating other SDK resources opportunistically.
- Changing the envelope itself.

## Decisions

### D1. Override Create, Update, and Delete; use envelope for Read and Schema

**Choice:** Pass placeholder write callbacks to `NewElasticsearchResource` and define `Create`, `Update`, and `Delete` on the concrete type.

**Rationale:**
- Create and Update both PUT settings, but they share enough logic that we can use a shared helper.
- Update specifically needs the prior Terraform state (`req.State`) to compute which settings were removed and null them out. The envelope's Update callback only receives the planned model, not the prior state.
- Delete needs the prior state to know which keys to unset. The envelope's Delete callback receives the model but the current flow derives the settings to remove from the Terraform state rather than from an API call.

### D2. Keep settings value/list type detection

**Choice:** On read, continue to inspect the flat-settings API response and store each setting as either `value` (string) or `value_list` ([]string) by type-asserting the returned value.

**Rationale:** This is externally-visible behavior. Users rely on the distinction between scalar and list settings round-tripping correctly.

### D3. Schema factory mirrors the SDK shape in PF

**Choice:** Express the existing SDK schema as PF `schema.Schema` using `schema.ListNestedBlock` (max 1) for `persistent` and `transient`, and `schema.SetNestedAttribute` for `setting` (since the SDK used `TypeSet`).

**Rationale:** The existing HCL uses blocks, not attributes, for `persistent`/`transient`, and the `setting` entries are unordered (set semantics). PF `ListNestedBlock` with a `SetNestedAttribute` for `setting` preserves this.

### D4. Model `GetResourceID` returns a constant

**Choice:** Since the resource manages a singleton cluster-scoped entity, `GetResourceID()` returns a fixed string (e.g., `"cluster-settings"`).

**Rationale:** The envelope requires a plan-safe write identity. The cluster settings resource has no natural user-supplied identifier, but the existing code uses the fixed suffix `cluster-settings` for the composite ID.

### D5. Import preserved via passthrough

**Choice:** Implement `ImportState` on the concrete type using `resource.ImportStatePassthroughID` on the `id` attribute.

**Rationale:** The SDK resource already supported import. The PF equivalent is the standard passthrough pattern.

## Risks / Trade-offs

- **Risk:** SDK→PF migration of nested set blocks is subtle. Set equality semantics in PF differ from SDK. **Mitigation:** Use acceptance tests to verify no diff drift on `setting` blocks.
- **Risk:** The update "null out removed settings" logic is easy to break. **Mitigation:** Add focused unit tests for the `updateRemovedSettings` equivalent before refactoring.
- **Risk:** `value_list` round-tripping. The SDK stored lists as `[]interface{}`; PF stores them as `types.List`. **Mitigation:** Explicit conversion helpers and acceptance-test validation.

## Migration Plan

1. Define `tfModel` with `ID`, `ElasticsearchConnection`, `Persistent`, `Transient` fields using PF types.
2. Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()` to the model.
3. Write a `getSchema() schema.Schema` factory that expresses the current block/attribute structure in PF, omitting `elasticsearch_connection`.
4. Write `expandSettings` and `flattenSettings` helpers that operate on PF types instead of `*schema.ResourceData`.
5. Write `readClusterSettings` callback and `deleteClusterSettings` callback.
6. Write `Create`, `Update`, and `Delete` overrides that preserve the current logic.
7. Wire the resource into the provider registrar (replace the SDK resource factory call with the PF resource factory).
8. Run `make build`, `make check-lint`, `make check-openspec`, and acceptance tests.

## Open Questions

None.
