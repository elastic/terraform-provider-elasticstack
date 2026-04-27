## Context

The Elasticsearch index template API `template` object accepts a `data_stream_options` field starting in Elasticsearch 9.x. That field currently supports one sub-field:

```json
{
  "data_stream_options": {
    "failure_store": {
      "enabled": true
    }
  }
}
```

The failure store is a secondary set of indices within a data stream that captures documents that fail ingest (mapping conflicts, pipeline exceptions). Enabling it requires a data-stream-enabled index template. The `data_stream_options.failure_store.enabled` boolean is the canonical API control surface for this feature.

The provider's current `models.Template` struct (`internal/models/models.go`) has no `DataStreamOptions` field. The `expandTemplate` function (`internal/elasticsearch/index/template.go`) and `flattenTemplateData` function do not handle `data_stream_options`. The Terraform schema for the `template` block has no `data_stream_options` block.

## Goals / Non-Goals

**Goals:**
- Add `data_stream_options.failure_store.enabled` to the `template` block in the Terraform schema.
- Wire the new field through create/update (expand) and read (flatten) paths.
- Add a version gate: return an error diagnostic when `data_stream_options` is configured against Elasticsearch < 9.0.0.
- Update the data source schema and flatten path for symmetry (read-only, no expand needed).
- Update the OpenSpec delta spec to formally document the requirement.

**Non-Goals:**
- Support future `data_stream_options` fields beyond `failure_store.enabled` in this change (schema is extensible, so future fields can be added incrementally).
- Migrate the resource to Plugin Framework — this change stays within the existing Plugin SDK implementation.
- Change `data_stream` (top-level) block semantics.

## Decisions

### 1. Nested block shape

Use a nested block structure rather than a JSON string:

```hcl
template {
  data_stream_options {
    failure_store {
      enabled = true
    }
  }
}
```

This matches the provider's existing pattern for structured sub-objects (e.g., `lifecycle`, `alias`). It allows plan-time validation and meaningful diffs. The Elasticsearch API shape maps cleanly to this two-level nesting.

`data_stream_options` is `TypeList, MaxItems: 1` (same as `lifecycle`).
`failure_store` is `TypeList, MaxItems: 1` inside it.
`enabled` is `TypeBool`, required when the `failure_store` block is present.

### 2. Version gate

`data_stream_options` is documented as an Elasticsearch 9.x feature. The resource already fetches server version for the `ignore_missing_component_templates` gate. This change will add an analogous check:

```go
var MinSupportedDataStreamOptionsVersion = version.Must(version.NewVersion("9.0.0"))
```

When `data_stream_options` is configured and the server reports a version below `9.0.0`, the provider returns an error diagnostic before calling the Put API.

### 3. Model layer

Add a new `DataStreamOptions` struct to `internal/models/models.go`:

```go
type DataStreamOptions struct {
    FailureStore *FailureStoreOptions `json:"failure_store,omitempty"`
}

type FailureStoreOptions struct {
    Enabled bool `json:"enabled"`
}
```

Add `DataStreamOptions *DataStreamOptions` to `models.Template` with JSON tag `data_stream_options,omitempty`.

### 4. Expand path (`expandTemplate`)

Check for `data_stream_options` in the flattened config map, and populate `templ.DataStreamOptions` when present.

### 5. Flatten path (`flattenTemplateData`)

When `template.DataStreamOptions != nil`, write the structured list back to the `tmpl` map.

### 6. Data source symmetry

`template_data_source.go` exposes a read-only view of the `template` block. Add the same schema definition (without ForceNew or Required constraints that don't apply to data sources) and flatten it in the data source read path, so users can inspect `data_stream_options` when reading existing templates.

## Risks / Trade-offs

- **Version gating may be imprecise if the feature appears in a late 8.x patch** — Mitigation: the issue explicitly states 9.x; use 9.0.0 as the gate and update if evidence of earlier availability surfaces.
- **`enabled = false` must still be sent when the block is configured** — Mitigation: use a required bool attribute (not optional with default) inside `failure_store`, so that the presence of the block always results in an explicit API payload.

## Open Questions

- Whether `data_stream_options` is also valid inside component templates (currently out of scope for this change).
- Whether the Elasticsearch API for `GET _index_template` always returns `data_stream_options` when set, or only when non-default — implementation should verify via acceptance test.
