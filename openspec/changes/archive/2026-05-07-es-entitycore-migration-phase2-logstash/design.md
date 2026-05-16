# Design: Migrate logstash pipeline to entitycore envelope

## Overview

Move `elasticstack_elasticsearch_logstash_pipeline` from Plugin SDK to Plugin Framework with the entitycore envelope. The resource manages centralized pipeline management via the Logstash APIs.

## Current State

- File: `internal/elasticsearch/logstash/pipeline.go` (~330 LOC)
- SDK resource with dynamic pipeline settings (key/value pairs with known type mappings in `allSettingsKeys`)
- Schema fields: `pipeline_id` (ForceNew), `description`, `pipeline`, `pipeline_metadata` (JSON), and many optional pipeline settings

## Schema Mapping

| SDK Field | PF Type | Notes |
|-----------|---------|-------|
| `pipeline_id` | `StringAttribute`, Required, RequiresReplace | |
| `description` | `StringAttribute`, Optional | |
| `last_modified` | `StringAttribute`, Computed | |
| `pipeline` | `StringAttribute`, Required | |
| `pipeline_metadata` | `StringAttribute`, Optional, json | |
| `pipeline_batch_delay` | `Int64Attribute`, Optional | |
| `pipeline_batch_size` | `Int64Attribute`, Optional | |
| `pipeline_workers` | `Int64Attribute`, Optional | |
| ... (all settings) | matching types | |

The `allSettingsKeys` map in the SDK maps setting names to `schema.ValueType` for conversion. In PF, each field is typed directly, so no dynamic type map is needed at runtime. However, the API response returns all settings in a flat map, so `read` must convert API map values to the correct Terraform types.

## Settings Conversion

SDK `expandSettings` and `flattenSettings` functions map between the flat API map and the typed schema. In PF:
- **Expand (Create/Update)**: iterate known setting keys, if value is non-nil in model, add to API map with the correct Go type.
- **Flatten (Read)**: iterate API map, set corresponding model field. Unknown keys are ignored (or optionally warned).

## Callback Design

### `createLogstashPipeline` / `updateLogstashPipeline`
1. Build `models.Pipeline` from model: `description`, `pipeline`, `pipeline_metadata` → JSON.
2. Gather non-nil settings into a `map[string]any`.
3. `PutLogstashPipeline(ctx, client, pipelineID, pipeline, settings)`.
4. Set composite ID on model.
5. Return model.

### `readLogstashPipeline`
1. Parse composite ID.
2. `GetLogstashPipeline(ctx, client, pipelineID)`.
3. If not found → `(_, false, nil)`.
4. Populate model fields from response.
5. For settings: iterate `allSettingsKeys` (or a PF equivalent constant mapping), look up in API response, set model field with correct type conversion.

### `deleteLogstashPipeline`
1. Parse composite ID.
2. `DeleteLogstashPipeline`.

## Model

```go
type Data struct {
    ID                      types.String `tfsdk:"id"`
    PipelineID              types.String `tfsdk:"pipeline_id"`
    Description             types.String `tfsdk:"description"`
    LastModified            types.String `tfsdk:"last_modified"`
    Pipeline                types.String `tfsdk:"pipeline"`
    PipelineMetadata        types.String `tfsdk:"pipeline_metadata"`
    PipelineBatchDelay      types.Int64  `tfsdk:"pipeline_batch_delay"`
    PipelineBatchSize       types.Int64  `tfsdk:"pipeline_batch_size"`
    ... // remaining settings
    ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
}
```

## Registration
- Remove from `provider/provider.go`
- Add to `provider/plugin_framework.go`

## Testing
- Re-use `pipeline_test.go` acceptance tests.
- Settings round-trip (especially queue type booleans/ints/strings) is the main regression risk.
