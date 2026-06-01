## Context

The Elasticsearch ML trained models API (`GET _ml/trained_models/{model_id}`) returns a `TrainedModelConfig` object containing metadata about a trained model: its type, size, tags, creation info, inference configuration, and input definition. The generated typed client exposes this as:

```go
client.Ml.GetTrainedModels().ModelId(modelID).Do(ctx)  // returns []types.TrainedModelConfig
```

The `TrainedModelConfig` Go type is defined in `typedapi/types/trainedmodelconfig.go`. No Terraform data source currently wraps this API.

## Goals

- Expose a read-only data source `elasticstack_elasticsearch_ml_trained_model` that retrieves a single trained model by ID.
- Surface the metadata fields practitioners need to reference in downstream resources (deployments, inference processors).
- Handle gracefully the case where a model does not exist (404 → not found, no error).
- Follow the project's Plugin Framework (PF) data source patterns, consistent with the `entitycore` envelope.

## Non-Goals

- Creating, uploading, or deleting trained models (deferred to a separate resource).
- Exposing model definition payloads (`definition`, `compressed_definition`) — these are large write-only fields not useful in a data source.
- Returning multiple models from a single data source invocation; `model_id` is a singular lookup.

## Decisions

| Topic | Decision |
|-------|-----------|
| Package location | `internal/elasticsearch/ml/trainedmodel/` — mirrors the convention for other ML entities (`ml/filter`, `ml/anomalydetectionjob`, etc.) |
| Schema type | Plugin Framework data source via the entitycore `NewElasticsearchDataSource` envelope |
| Identity | `model_id` (required); composite `id` = `<cluster_uuid>/<model_id>` via `client.ID(clusterUUID, modelID)` |
| API call | `client.Ml.GetTrainedModels().ModelId(modelID).Do(ctx)` — requests exactly one model; response is `[]TrainedModelConfig`; take `[0]` |
| 404 / not found | When the API returns 404 or the results array is empty, signal not-found to the framework (no error); mirrors ML filter read behavior |
| JSON computed fields | `input_json`, `inference_config_json`, `metadata_json` — marshal the corresponding struct fields to JSON string; null when the field is absent from the API response |
| `default_field_map` | Expose as `map(string)` — matches `TrainedModelConfig.DefaultFieldMap` which is `map[string]string` |
| `tags` | Expose as `set(string)` |
| `model_size_bytes` | Expose as `int64` (bytes); null when absent |
| `create_time` | Store as string; format as ISO-8601 if the API returns an `EpochTime` or `DateTime` |
| Connection block | No explicit `elasticsearch_connection` block in the schema — it is injected transparently by the entitycore envelope |
| Client helper | A small helper function in `internal/clients/elasticsearch/` that wraps `GetTrainedModels` for testability; alternatively inline in `read.go` if the pattern is simple enough |
| Minimum ES version | 8.0+; no version gate is expected beyond what the provider already enforces. Confirm during implementation. |

## Non-Goals (implementation)

- Do not implement or stub a write resource; this data source is strictly read-only.
- Do not expose `allow_no_match`, `from`, `size`, or other query parameters — the data source reads exactly one model by ID.

## Risks / Trade-offs

- **Model not in cluster**: If a model was removed outside Terraform, the data source will return not-found. Downstream `depends_on` or references will fail at plan time if the model is absent. This is expected behavior for a data source.
- **Large models**: Some PyTorch models can be very large. Because definition fields are excluded, the response payload is small; no performance concern.
- **Alias resolution**: The API accepts model aliases as `model_id`. If an alias is used, the API returns the canonical model; `model_id` in state will reflect the alias passed by the practitioner, not the resolved model ID. This may cause a drift if the user later uses the canonical ID. Document this behavior.
- **`create_time` format**: The API returns time as an epoch ms integer or an ISO string depending on ES version. Implementation should normalize to a consistent string format.

## Open Questions

1. **`create_time` wire format**: Does `TrainedModelConfig.CreateTime` come back as an epoch-ms integer or an ISO-8601 string? The Go type may be `types.DateTime` or `int64`. Confirm during implementation and normalize to ISO-8601 string.
2. **Alias in state vs. canonical `model_id`**: Should the data source echo back the input `model_id` (alias or canonical) or always store the canonical model ID returned by the API? The issue body suggests echoing the input; confirm expected behavior and document.
3. **Minimum Elasticsearch version**: Confirm the exact minimum version for the `GET _ml/trained_models` API (expected 8.0+); add a version gate if an older stack must be supported.
