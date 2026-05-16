## Why

The `elasticstack_elasticsearch_ml_anomaly_detection_job` delete path calls `CloseJob` then immediately calls `DeleteJob` without waiting for the close transition to commit in Elasticsearch's internal `.ml-config` index. The ES `DeleteJob` API uses optimistic concurrency control: if the document's sequence number is incremented by the close operation between when `DeleteJob` reads and writes it, the call fails with HTTP 409 `version_conflict_engine_exception`. This happens deterministically whenever a job is deleted from `opened` state — which is exactly what occurs during the post-test teardown of `TestAccResourceMLJobStateExplicitConnection` (whose final step leaves the job open).

## What Changes

- Add a `WaitForMLJobClosed` polling helper to `internal/clients/elasticsearch/ml_job.go` that blocks until a job's stats report `closed` state (or the job is gone).
- Call `WaitForMLJobClosed` in `internal/elasticsearch/ml/anomalydetectionjob/delete.go` between the `CloseJob` and `DeleteJob` API calls.
- Add a `timeouts` block to the `elasticstack_elasticsearch_ml_anomaly_detection_job` resource supporting a configurable `delete` timeout (default 20 minutes), so the polling wait bound is user-controllable.
- Update REQ-021 in the `elasticsearch-ml-anomaly-detection-job` spec to require polling for `closed` state after `CloseJob` returns and before `DeleteJob` is called.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-ml-anomaly-detection-job`: REQ-021 (Delete — close before delete) must be strengthened to require that the provider polls the job's state until `closed` is confirmed before calling `DeleteJob`. This closes a gap where `CloseJob` returns before its internal `.ml-config` document write is durable, causing a version conflict on the subsequent delete.

## Impact

- `internal/clients/elasticsearch/ml_job.go`: new exported `WaitForMLJobClosed` function.
- `internal/elasticsearch/ml/anomalydetectionjob/delete.go`: insert wait between close and delete; read the configurable delete timeout.
- `internal/elasticsearch/ml/anomalydetectionjob/schema.go` + `models_tf.go`: add `timeouts` block (Delete only).
- `openspec/specs/elasticsearch-ml-anomaly-detection-job/spec.md`: update REQ-021.
- No new dependencies — reuses `asyncutils.WaitForStateTransition`, `GetMLJobStats`, and the existing `terraform-plugin-framework-timeouts` library already in the codebase.
