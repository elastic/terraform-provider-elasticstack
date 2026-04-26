## Context

Elasticsearch 9.x added a `data_stream_options` field to the index template `template` object. It governs per-template options for data streams, starting with `failure_store.enabled` — a boolean that controls whether the failure store is active for all data streams backed by this template. The Elasticsearch index template API ([docs](https://www.elastic.co/docs/manage-data/data-store/data-streams/failure-store)) already accepts and returns this field, but the provider's `models.Template` struct and schema do not expose it.

The gap forces practitioners to make out-of-band API calls to patch the field, which breaks infrastructure-as-code principles and produces drift between Terraform state and actual cluster configuration.

This change is narrowly scoped: add the new block to the schema, extend the Go model, and wire up serialization and deserialization. No schema migration or state upgrade is needed because the block is additive.

## Goals / Non-Goals

**Goals:**
- Expose `template.data_stream_options.failure_store.enabled` in both the resource and data source for `elasticstack_elasticsearch_index_template`.
- Serialize the field to the Elasticsearch API on create and update.
- Deserialize and store the field from Elasticsearch API responses on read.
- Provide acceptance test coverage for creating, updating, and removing `data_stream_options`.
- Regenerate provider documentation.

**Non-Goals:**
- Expose additional `data_stream_options` sub-fields beyond `failure_store.enabled` (fields not yet documented or stable in the Elasticsearch 9.x API).
- Apply the same block to the component template resource (`elasticstack_elasticsearch_component_template`), which shares the same `Template` model. That resource is out of scope for this change, though the model extension will make it easy to add later.
- Introduce a version guard for `data_stream_options`. The field is sent as provided; if the cluster rejects it, the API error surfaces normally through Terraform diagnostics.

## Decisions

### 1. Block shape: nested `data_stream_options > failure_store > enabled`

The Elasticsearch API shape is:
```json
{
  "template": {
    "data_stream_options": {
      "failure_store": {
        "enabled": true
      }
    }
  }
}
```

The Terraform schema mirrors this directly with two levels of `TypeList/MaxItems:1` nesting:
- `template.data_stream_options` — optional, max one block
- `template.data_stream_options.failure_store` — optional, max one block
- `template.data_stream_options.failure_store.enabled` — optional bool

Using `TypeList/MaxItems:1` is consistent with how `data_stream` and `lifecycle` are modeled in this resource, and with the Plugin SDK pattern used throughout the codebase. No `TypeSet` is needed because neither block has element identity semantics.

### 2. Model extension in `models.Template`

A new `DataStreamOptions` struct and nested `FailureStoreOptions` struct are added to `internal/models/models.go`:

```go
type DataStreamOptions struct {
    FailureStore *FailureStoreOptions `json:"failure_store,omitempty"`
}

type FailureStoreOptions struct {
    Enabled *bool `json:"enabled,omitempty"`
}
```

The `Template` struct gains:
```go
DataStreamOptions *DataStreamOptions `json:"data_stream_options,omitempty"`
```

Using pointer types with `omitempty` ensures the field is omitted when not set, preserving existing behavior for templates that do not use `data_stream_options`.

### 3. Serialization in `expandTemplate`

In `internal/elasticsearch/index/template.go`, `expandTemplate` reads the `data_stream_options` key from the decoded template map and populates the model. The pattern follows how `lifecycle` is expanded in the same function.

### 4. Deserialization in `flattenTemplateData`

`flattenTemplateData` checks whether `template.DataStreamOptions` is non-nil and writes a corresponding list representation into the Terraform state map. This mirrors the `lifecycle` flatten pattern.

### 5. Data source parity

`internal/elasticsearch/index/template_data_source.go` defines an independent schema for the data source. It must receive the same `data_stream_options` block schema addition so that the data source can expose the field when read from the API. The flatten path in `flattenTemplateData` is shared between resource and data source, so only the schema addition and any data-source-specific state assignment needs touching.

### 6. No version guard

The issue reports that `data_stream_options` is available starting in Elasticsearch 9.x. The provider does not currently guard the `lifecycle` block or other newer fields with explicit version checks in all places — errors from the API surface naturally via Terraform diagnostics. Adding a hard version guard is therefore out of scope unless a version check pattern is specifically requested. Practitioners on pre-9.x clusters who configure `data_stream_options` will receive an API-level error.

## Risks / Trade-offs

- **Model shared with component template** — `models.Template` is also used by `ComponentTemplate`. Adding `DataStreamOptions` to the struct will cause the field to be present in the component template JSON marshaling path, but `omitempty` ensures no field is emitted unless set. The component template resource does not expose `data_stream_options` in its schema, so the field will never be populated for component template operations in this change.
- **No version guard** — Practitioners using older clusters may see API errors rather than plan-time diagnostics. This matches the existing behavior for `lifecycle` and is acceptable for now.

## Migration Plan

1. Add `DataStreamOptions` and `FailureStoreOptions` structs to `internal/models/models.go`.
2. Extend `models.Template` with the `DataStreamOptions` field.
3. Add `data_stream_options` block to the schema in `internal/elasticsearch/index/template.go` (resource).
4. Extend `expandTemplate` to read and populate `DataStreamOptions`.
5. Extend `flattenTemplateData` to write `data_stream_options` into state.
6. Add `data_stream_options` block to the schema in `internal/elasticsearch/index/template_data_source.go` (data source).
7. Add or update acceptance tests for the `data_stream_options` block using the test data directory pattern (`internal/elasticsearch/index/testdata/`).
8. Regenerate docs: `docs/resources/elasticsearch_index_template.md` and `docs/data-sources/elasticsearch_index_template.md`.

No state upgrade is required. The block is purely additive.

## Open Questions

- Whether `data_stream_options` should also be gated on Elasticsearch 9.0+, providing a user-friendly plan-time error instead of an apply-time API error for earlier clusters. This can be deferred to a follow-up if the API-level error proves confusing in practice.
- Whether component template should receive the same block in this change or a subsequent one. The current decision is to defer it, since the issue specifically mentions `elasticstack_elasticsearch_index_template`.
