# Design: Migrate Elasticsearch ingest pipeline to entitycore envelope

## Overview

Move `elasticstack_elasticsearch_ingest_pipeline` from Plugin SDK to Plugin Framework with the entitycore `NewElasticsearchResource` envelope. The resource is a straightforward PUT-based resource: PUT creates, PUT updates, GET reads, DELETE deletes.

## Current State

- File: `internal/elasticsearch/ingest/pipeline.go`
- SDK resource with `CreateContext`, `UpdateContext`, `ReadContext`, `DeleteContext`
- Schema: `name` (ForceNew), `description`, `on_failure` (list of JSON strings), `processors` (list of JSON strings), `metadata` (JSON string)
- Registered in `provider/provider.go` (SDK provider)

## Target State

- Re-use the existing `internal/elasticsearch/ingest/` package for the PF resource files.
- PF resource embedding `*entitycore.ElasticsearchResource[Data]`
- Envelope owns Schema, Read, Update, Delete
- Concrete type provides callbacks

## Schema Mapping

| SDK Field | PF Type | Notes |
|-----------|---------|-------|
| `name` | `StringAttribute` | Required, `RequiresReplace()` |
| `description` | `StringAttribute` | Optional |
| `processors` | `ListAttribute` of `jsontypes.Normalized` | Required, JSON strings |
| `on_failure` | `ListAttribute` of `jsontypes.Normalized` | Optional, JSON strings |
| `metadata` | `StringAttribute` with `jsontypes.NormalizedType{}` | Optional |
| `id` | `StringAttribute` | Computed |

## Callback Design

### `readIngestPipeline`
Decode state → parse composite ID → `GetIngestPipeline` → populate model from API response → return `(model, found, diags)`.

### `deleteIngestPipeline`
Parse composite ID → `DeleteIngestPipeline`.

### `createIngestPipeline` / `updateIngestPipeline`
Both are PUT operations to the same endpoint:
1. Build pipeline body from model (decode JSON strings for processors/on_failure).
2. `PutIngestPipeline(ctx, client, name, body)`.
3. Set composite ID on model.
4. Return model (envelope calls readFunc for readback).

## Model

```go
type Data struct {
    ID                      types.String `tfsdk:"id"`
    Name                    types.String `tfsdk:"name"`
    Description             types.String `tfsdk:"description"`
    Processors              types.List   `tfsdk:"processors"`
    OnFailure               types.List   `tfsdk:"on_failure"`
    Metadata                types.String `tfsdk:"metadata"`
    ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
}
```

Value-receiver methods: `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()`.

## Migration of JSON list fields

SDK uses `[]any` with JSON decode per element. PF list of JSON strings can use `types.List` of `jsontypes.Normalized` or plain strings with custom type. The simplest mapping is `types.List` of `basetypes.StringType` with validation that each element is JSON. Or use `jsontypes.NormalizedType` as element type for the list.

## Registration

- Remove from `provider/provider.go` `ResourcesMap`
- Add to `provider/plugin_framework.go` `resources()` slice

## Testing

- Re-use existing acceptance tests; Terraform interface should be identical.
- Add SDK→PF from_sdk test if state shape differs.

## Open Questions

- Should processors/on_failure use `jsontypes.Normalized` element type? This preserves JSON normalization, which the SDK diff suppressor did. **Decision: Yes**, use `jsontypes.NormalizedType` as the element type for both lists.
