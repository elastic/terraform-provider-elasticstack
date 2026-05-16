## Context

The `elasticstack_elasticsearch_security_role` data source currently lives in `internal/elasticsearch/security/role_data_source.go` as a Terraform Plugin SDK v2 `*schema.Resource`. It has deeply nested computed attributes: `applications` (set of objects), `indices` (set of objects with nested `field_security` list), and `remote_indices` (set of objects with nested `field_security` list). Scalar JSON fields (`global`, `metadata`) are stored as strings. The resource portion is already PF-based; only the data source remains in SDK.

## Goals / Non-Goals

**Goals:**
- Migrate the security role data source to Plugin Framework.
- Wrap with `entitycore.NewElasticsearchDataSource`.
- Maintain 1:1 schema parity, including nested set structures and JSON string attributes.
- Preserve the existing flattening helpers (`flattenApplicationsData`, `flattenIndicesData`, `flattenRemoteIndicesData`).

**Non-Goals:**
- No changes to the security role resource.
- No changes to the security role mapping data source or resource.
- No new attributes.

## Decisions

### Decision: Nested sets become `schema.SetNestedAttribute` with object element type

**Chosen:** PF uses `schema.SetNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{...}}` for `applications`, `indices`, and `remote_indices`.

**Rationale:** Direct equivalent for SDK `TypeSet` with `Elem: &schema.Resource{}`.

### Decision: `field_security` inside indices/remote_indices stays a list

**Chosen:** Inside the nested object, `field_security` is declared as `schema.ListNestedAttribute{Computed: true}`.

**Rationale:** Matches the current SDK schema where `field_security` is `TypeList` with `Elem: &schema.Resource{}`.

### Decision: JSON strings use `jsontypes.NormalizedType{}`

**Chosen:** The `global` and `metadata` attributes use `CustomType: jsontypes.NormalizedType{}`.

**Rationale:** This is the PF convention for JSON string attributes in this provider (used by role mapping data source, enrich policy, etc.).

### Decision: Cluster privilege enums stay as strings

**Chosen:** The `cluster` attribute remains `schema.SetAttribute` with `ElementType: types.StringType`.

**Rationale:** The SDK stores these as plain strings. The PF model will hold `types.Set` of `types.String`. The read callback maps the typed API `ClusterPrivileges` enums to strings.

### Decision: Not-found returns an empty id without diagnostics

**Chosen:** When `elasticsearch.GetRole` returns nil with no error, the callback returns a model with `ID` set to `types.StringValue("")` and returns no warning or error diagnostic.

**Rationale:** This preserves the current SDK behavior, which logs internally and clears the data source ID without surfacing a Terraform diagnostic.

## Risks / Trade-offs

- **Complex nested sets are verbose in PF.** `indices` and `remote_indices` each have five fields plus a nested `field_security` object. The schema definition will be long. Mitigated by following the existing PF resource schema patterns in the same package.

- **Flattening helpers return `[]any`.** The existing flatten functions build `[]any` slices of `map[string]any`. These need to be converted to PF nested set/list values via `tfsdk.ValueFrom` or manual construction with `types.ObjectValue`. A small adapter function bridges the two.

- **Not-found behavior approximation.** SDK data sources clear their ID directly with `d.SetId("")`; PF data sources persist the callback result. The callback must therefore return an empty `id` and avoid warning diagnostics so callers see the same no-error missing-role behavior.

## Migration Plan

No data migration required. Schema shape preserved.

## Open Questions

None.
