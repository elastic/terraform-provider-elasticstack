# Design: Migrate Elasticsearch data stream to entitycore envelope

## Overview

Move `elasticstack_elasticsearch_data_stream` from Plugin SDK to Plugin Framework with the entitycore envelope. The resource is create-once (ForceNew on `name`): PUT creates, GET reads, DELETE deletes. No true update.

## Current State

- File: `internal/elasticsearch/index/data_stream.go`
- SDK resource; all fields except `name` are Computed
- `ResourceDataStreamPut` does `PutDataStream` then readback

## Target State

- PF resource in `internal/elasticsearch/index/datastream/` or keep existing `internal/elasticsearch/index/` package
- Embed `*entitycore.ElasticsearchResource[Data]`
- `name` remains ForceNew (use `RequiresReplace` plan modifier)

## Schema Mapping

| SDK Field | PF Type | Notes |
|-----------|---------|-------|
| `name` | `StringAttribute` | Required, `RequiresReplace()` |
| `timestamp_field` | `StringAttribute` | Computed |
| `indices` | `ListAttribute` of object | Computed |
| `generation` | `IntAttribute` | Computed |
| `metadata` | `StringAttribute` | Computed (JSON) |
| `status` | `StringAttribute` | Computed |
| `template` | `StringAttribute` | Computed |
| `ilm_policy` | `StringAttribute` | Computed |
| `hidden` | `BoolAttribute` | Computed |
| `system` | `BoolAttribute` | Computed |
| `replicated` | `BoolAttribute` | Computed |

## Callback Design

### `createDataStream`
1. `client.ID(ctx, data.Name.ValueString())` → set `data.ID`
2. `PutDataStream(ctx, client, name)`
3. Return `data` (envelope does readback)

### `updateDataStream`
Same as create (ForceNew resource conceptually has no update, but envelope still needs a callback). Since `name` is `RequiresReplace()`, Terraform will Destroy+Create on name change. The update callback can simply call create logic or be identical to it.

### `readDataStream`
1. Parse composite ID
2. `GetDataStream(ctx, client, name)`
3. If nil → `(_, false, nil)`
4. Populate all computed fields from response
5. Return `(data, true, nil)`

### `deleteDataStream`
1. Parse composite ID
2. `DeleteDataStream`

## Model

```go
type Data struct {
    ID                      types.String `tfsdk:"id"`
    Name                    types.String `tfsdk:"name"`
    TimestampField          types.String `tfsdk:"timestamp_field"`
    Indices                 types.List   `tfsdk:"indices"`
    Generation              types.Int64  `tfsdk:"generation"`
    Metadata                types.String `tfsdk:"metadata"`
    Status                  types.String `tfsdk:"status"`
    Template                types.String `tfsdk:"template"`
    ILMPolicy               types.String `tfsdk:"ilm_policy"`
    Hidden                  types.Bool   `tfsdk:"hidden"`
    System                  types.Bool   `tfsdk:"system"`
    Replicated              types.Bool   `tfsdk:"replicated"`
    ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
}
```

## Registration

- Remove from `provider/provider.go` `ResourcesMap`
- Add to `provider/plugin_framework.go` `resources()`

## Indices Nested Object

The `indices` computed field is a list of objects with `index_name` and `index_uuid`. In PF, define a nested object type or use `types.List` of a struct with `tfsdk` tags.

## Testing

- Re-use `data_stream_test.go` acceptance tests.
- No interface changes expected.
