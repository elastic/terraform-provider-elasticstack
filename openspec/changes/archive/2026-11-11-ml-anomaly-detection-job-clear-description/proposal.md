## Why

Removing `description` from an `elasticstack_elasticsearch_ml_anomaly_detection_job` configuration fails on apply:

```
Error: Provider produced inconsistent result after apply
.description: was null, but now cty.StringVal("...")
```

`description` is Optional only, so the plan holds null. `UpdateAPIModel.BuildFromPlan` skipped null values, so the Update Job API was never called (the resource logged "No changes detected to updatable fields"), Elasticsearch kept the old description, and the read-after-write put it back into state.

## What Changes

- When the prior state has a `description` and the plan has null, the update sends `"description": ""`. Elasticsearch stores the empty string and the read already maps it to null.
- An unknown planned `description` is still omitted from the update body.

Non-goals:

- `background_persist_interval`, `renormalization_window_days` and `results_retention_days` have the same skip-on-null shape. They are numeric or duration values, so clearing them needs a separate decision on what to send (explicit null or the server default); out of scope here.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-ml-anomaly-detection-job`: REQ-016–REQ-018 state that removing `description` from configuration clears it through the Update Job API.

## Impact

- `internal/elasticsearch/ml/anomalydetectionjob/models_api.go`: `BuildFromPlan`.
- `internal/elasticsearch/ml/anomalydetectionjob/update_test.go`: unit coverage with a fake Elasticsearch for update followed by read.
- `internal/elasticsearch/ml/anomalydetectionjob/acc_test.go` and testdata: acceptance coverage for removing `description`.
- Users: no migration.
