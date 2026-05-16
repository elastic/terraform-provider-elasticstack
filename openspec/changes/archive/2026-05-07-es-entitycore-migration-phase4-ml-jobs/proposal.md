## Why

`elasticstack_elasticsearch_ml_anomaly_detection_job` and `elasticstack_elasticsearch_ml_datafeed` are two mature Plugin Framework ML resources that both embed `*entitycore.ResourceBase` and repeat the standard Read/Delete/Schema prelude. Centralizing these preludes in the entitycore envelope removes duplicated boilerplate across the ML package and keeps diagnostic handling, client resolution, and state persistence consistent with other envelope-backed resources.

## What Changes

- Migrate `internal/elasticsearch/ml/anomalydetectionjob` and `internal/elasticsearch/ml/datafeed` to `*entitycore.ElasticsearchResource[T]`.
- Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` to both `TFModel` (anomaly detection job) and `Datafeed` (datafeed).
- Convert schemas to factories without `elasticsearch_connection` blocks (envelope injects them).
- Extract existing `read`, `delete`, `create`, and `update` bodies into package-level callbacks with the envelope signatures.
- For `anomaly_detection_job`: use real callbacks for create, read, and delete; override `Update` on the concrete type because the update body builder needs access to both plan and state.
- For `datafeed`: use real callbacks for all four operations (create, read, update, delete).
- Preserve custom `ImportState` on both concrete types.
- No schema changes, no external behavior changes, no acceptance test changes.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `elasticsearch-ml-anomaly-detection-job`: Internal implementation migrated to entitycore envelope.
- `elasticsearch-ml-datafeed`: Internal implementation migrated to entitycore envelope.

## Impact

- Refactored code: `internal/elasticsearch/ml/anomalydetectionjob/resource.go`, `read.go`, `create.go`, `update.go`, `delete.go`, `schema.go`, `models.go` (or equivalent).
- Refactored code: `internal/elasticsearch/ml/datafeed/resource.go`, `read.go`, `create.go`, `update.go`, `delete.go`, `schema.go`, `models.go`.
- Acceptance tests for both resources must continue to pass without modification.
