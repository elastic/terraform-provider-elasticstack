## Why

The `metric_custom_indicator` schema marks `field` as required on every metric entry, but the Kibana API rejects `field` when `aggregation = "doc_count"` — `doc_count` counts documents and has no field operand. Users who configure a `doc_count` metric receive a 400 error ("Excess keys are not allowed: ...field"), making this aggregation type entirely unusable from Terraform. The `timeslice_metric_indicator` already handles this correctly; `metric_custom_indicator` needs the same treatment.

## What Changes

- `metric_custom_indicator.good.metrics.field`: change from required to optional.
- `metric_custom_indicator.total.metrics.field`: change from required to optional.
- Write path (`buildGoodMetricItem` / `buildTotalMetricItem`): when `aggregation = "doc_count"`, send the no-field API variant (`Metrics1`) instead of the field-bearing variant (`Metrics0`).
- Add unit tests covering the `doc_count` write path and read-path round-trip.
- Add an acceptance test (`TestAccResourceSlo_metric_custom_indicator_doc_count`) with a testdata TF config that exercises `doc_count` on both good and total metrics.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `kibana-slo`: `metric_custom_indicator.{good,total}.metrics.field` changes from required to optional — must not be set when `aggregation = "doc_count"`, must be set for all other aggregations.

## Impact

- `internal/kibana/slo/schema.go` — two attribute definitions
- `internal/kibana/slo/models_metric_custom_indicator.go` — write-path helpers
- `internal/kibana/slo/models_metric_custom_indicator_test.go` — unit tests
- `internal/kibana/slo/acc_test.go` — acceptance test function
- `internal/kibana/slo/testdata/TestAccResourceSlo_metric_custom_indicator_doc_count/test/test.tf` — new testdata config
- No API client changes, no schema version bump, no state migration needed (adding optional attribute is backwards-compatible with existing state).
