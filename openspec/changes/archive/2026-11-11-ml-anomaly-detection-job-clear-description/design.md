## Context

`elasticstack_elasticsearch_ml_anomaly_detection_job` updates through `POST _ml/anomaly_detectors/{id}/_update` with only the changed fields. `description` is Optional (not Computed), so removing it from configuration plans null. `BuildFromPlan` required the planned value to be non-null, so the change was dropped and the read-after-write restored the server value.

## Decisions

**Send an empty string for a removed description.** The Update Job API replaces `description` with any string it receives. The read path already stores an empty description as null (`typeutils.NonEmptyStringishValue`), so the state after apply matches the plan. The guard now skips only unknown values.

**Leave the other skip-on-null fields alone.** `background_persist_interval`, `renormalization_window_days` and `results_retention_days` are numeric or duration values, so an empty string is not a valid way to clear them; that needs its own change.

## Testing

- Unit: `BuildFromPlan` with a null planned description sends `""`; with an unknown one it sends nothing.
- Unit with a fake Elasticsearch (`httptest`): update followed by read, asserting one Update Job call with `{"description":""}` and a null `description` in the refreshed state.
- Acceptance: create a job with `description`, then apply the same configuration without it.

## Risks / Trade-offs

None known. A job created outside Terraform with a description and imported without one in configuration now has its description cleared on the next update that touches the job, which matches the configuration.
