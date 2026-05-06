## Context

The `elasticstack_elasticsearch_info` data source currently lives in `internal/elasticsearch/cluster/cluster_info_data_source.go` as a Terraform Plugin SDK v2 `*schema.Resource`. It defines a computed `version` list block and several computed string attributes. The read function manually resolves the Elasticsearch client via `factory.GetElasticsearchClientFromSDK(d)`, calls `elasticsearch.GetClusterInfo`, and uses `d.Set` / `d.SetId` to persist state.

Four Elasticsearch data sources (`enrich_policy`, `indices`, `index_template`, `security_role_mapping`) have already been migrated to Plugin Framework and wrapped with `entitycore.NewElasticsearchDataSource`. This migration follows the exact same pattern.

## Goals / Non-Goals

**Goals:**
- Migrate `cluster_info` data source to Plugin Framework.
- Wrap the new PF implementation with `entitycore.NewElasticsearchDataSource` to automatically handle `elasticsearch_connection` block injection, config decode, scoped client resolution, and state persistence.
- Maintain 1:1 schema parity with the existing SDK implementation — no breaking attribute or type changes.
- Maintain the same acceptance test coverage.

**Non-Goals:**
- No changes to the resource portion (there is none).
- No changes to the `elasticsearch.GetClusterInfo` client helper.
- No new attributes or behavioral changes beyond the framework migration.

## Decisions

### Decision: Keep `version` as a single-element list block

**Chosen:** The PF schema uses `schema.ListNestedAttribute` with `Computed: true` to preserve the existing `version` block shape.

**Rationale:** Changing the block to a nested object or flattening its fields to top-level attributes would be a breaking schema change. Consumers reference `version.0.build_date`, etc.

### Decision: `id` is set from `cluster_uuid` in the read callback

**Chosen:** The `readDataSource` callback sets `model.ID = types.StringValue(info.ClusterUuid)` directly.

**Rationale:** This matches the current behavior where `d.SetId(info.ClusterUuid)` is called. `esClient.ID()` produces a composite `<cluster_uuid>/<resource_id>` — passing `info.ClusterUuid` as the resource id would yield `<cluster_uuid>/<cluster_uuid>`, which conflicts with the legacy identity. The envelope does not touch `id`; the callback owns it.

### Decision: Build-date type-switch logic moves into the callback unchanged

**Chosen:** The `switch v := info.Version.BuildDate.(type)` logic from the SDK read function is copied into the PF callback.

**Rationale:** This is business logic the envelope cannot own. The typed API already returns a union type for `build_date`, so the conversion is still necessary.

## Risks / Trade-offs

- **PF `types.List` nesting complexity.** The `version` block is a computed list of exactly one element. PF nested attributes require careful use of `types.List` with custom element types. Mitigated by following the existing `index/template` and `enrich` patterns.

- **Test conversion.** SDK acceptance tests use `TestCheckResourceAttr` and similar helpers. PF tests may need `TestCheckResourceAttr` still works for basic attributes, but nested blocks may need `TestMatchResourceAttr` or `testAccCheckAttributeValue`. Mitigated by reviewing existing PF test patterns in the repo.

## Migration Plan

No data migration required. The schema shape is unchanged. Existing Terraform configurations will continue to work after the migration because attribute names, types, and computed/optional flags are preserved.

## Open Questions

None.
