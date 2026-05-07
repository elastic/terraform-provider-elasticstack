## Context

The `elasticstack_elasticsearch_snapshot_repository` data source currently lives in `internal/elasticsearch/cluster/snapshot_repository_data_source.go` as a Terraform Plugin SDK v2 `*schema.Resource`. It has a complex schema with six mutually exclusive computed type blocks (`fs`, `url`, `gcs`, `azure`, `s3`, `hdfs`). The read function resolves the client, fetches the repository, type-switches over the result, and uses `d.Set` to populate exactly one type block.

The resource portion of snapshot repository was already migrated to PF elsewhere; the data source is the remaining SDK implementation.

## Goals / Non-Goals

**Goals:**
- Migrate the snapshot repository data source to Plugin Framework.
- Wrap with `entitycore.NewElasticsearchDataSource`.
- Maintain 1:1 schema parity for all six repository type blocks.
- Preserve the existing flattening and type-conversion behavior (string-to-int, string-to-bool).

**Non-Goals:**
- No changes to the snapshot repository resource.
- No changes to the `flattenRepoSettings` helper used by the resource.
- No new repository types.

## Decisions

### Decision: Reuse `flattenRepoSettings` for the data source callback

**Chosen:** The PF read callback will call the existing `flattenRepoSettings` helper (or a PF-compatible variant) to convert API response settings into the nested block model.

**Rationale:** The flattening logic is already well-tested and shared conceptually with the resource. If `flattenRepoSettings` returns SDK-compatible `[]any`, a thin adapter layer converts it to the PF nested list model. Alternatively, a PF-specific flattening function is created alongside it.

### Decision: Each repository type is a computed `schema.ListNestedAttribute`

**Chosen:** The PF schema declares each type as a computed list nested attribute containing computed scalar attributes.

**Rationale:** In the SDK schema these are `TypeList` with `Elem: &schema.Resource{}` and `Computed: true`. PF `ListNestedAttribute` is the direct equivalent.

### Decision: Not-found returns a warning diagnostic

**Chosen:** When the API returns `nil` with no error, the callback sets `id` and returns a warning diagnostic (matching current SDK behavior).

**Rationale:** This is existing behavior — `d.SetId(id.String())` happens before the nil check, and a warning is returned. The envelope will surface the warning and persist the model (with empty type blocks).

## Risks / Trade-offs

- **Flattening logic coupling.** The data source currently does `DataSourceSnapshotRespository().Schema[currentRepo.Type].Elem.(*schema.Resource).Schema` to discover schema keys at runtime. This SDK self-reference is not available in PF. The callback needs an explicit mapping of supported type names to their setting keys. Mitigated by extracting a static map or using the typed API response types.

- **Large schema surface.** Six type blocks with many computed fields make the PF schema verbose. Mitigated by code generation or structured schema building, but for migration a hand-written schema is acceptable.

## Migration Plan

No data migration required. The schema shape is preserved.

## Open Questions

None.
