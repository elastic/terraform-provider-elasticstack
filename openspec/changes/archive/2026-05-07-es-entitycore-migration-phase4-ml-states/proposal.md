## Why

`elasticstack_elasticsearch_ml_job_state` and `elasticstack_elasticsearch_ml_datafeed_state` are state-transition resources (not true CRUD). They currently embed `*entitycore.ResourceBase` and implement their own `Schema`, `Read`, `Create`, `Update`, and `Delete`. Their Read and Delete preludes are identical boilerplate to other resources. Migrating them to the entitycore envelope removes this duplication while keeping their unique state-transition logic on the concrete types.

## What Changes

- Migrate `internal/elasticsearch/ml/jobstate` and `internal/elasticsearch/ml/datafeed_state` to `*entitycore.ElasticsearchResource[T]`.
- Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` to both `MLJobStateData` and `MLDatafeedStateData`.
- Convert schemas to factories without `elasticsearch_connection` blocks.
- Use `PlaceholderElasticsearchWriteCallbacks` for Create and Update on both resources, because their write paths do not fit the standard callback contract (timeouts extracted from model and direct state transitions).
- Provide real read callbacks (`readMLJobState`, `readMLDatafeedState`).
- Override `Create`, `Update`, and `Delete` on both concrete types. Provide non-nil delete callbacks for envelope construction and have concrete Delete methods preserve the existing no-op/conditional-stop behavior.
- Preserve custom `ImportState` on both concrete types.
- No schema changes, no external behavior changes.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `elasticsearch-ml-job-state`: Internal implementation migrated to entitycore envelope.
- `elasticsearch-ml-datafeed-state`: Internal implementation migrated to entitycore envelope.

## Impact

- Refactored code in `internal/elasticsearch/ml/jobstate/` and `internal/elasticsearch/ml/datafeed_state/`.
- Acceptance tests for both resources must continue to pass without modification.
