## 1. Unit Tests

- [ ] 1.1 Add `buildGoodItemDocCount` and `buildTotalItemDocCount` helpers (using `Metrics1`) inside `TestMetricCustomIndicator_PopulateFromAPI` in `models_metric_custom_indicator_test.go`
- [ ] 1.2 Add subtest "uses Metrics1 for doc_count aggregation" to `TestMetricCustomIndicator_ToAPI` — builds a model with `aggregation = "doc_count"` and `Field: types.StringNull()`, calls `metricCustomIndicatorToAPI()`, asserts the result decodes as `GoodMetrics1`/`TotalMetrics1` with correct name, aggregation, and filter
- [ ] 1.3 Add subtest "maps doc_count metrics without field" to `TestMetricCustomIndicator_PopulateFromAPI` — constructs `Metrics1` API items using the helpers from 1.1, calls `populateFromMetricCustomIndicator`, asserts `Field.IsNull()` for both good and total

## 2. Fix

- [ ] 2.1 In `schema.go` `metricCustomIndicatorSchema()`, change `field` from `Required: true` to `Optional: true` (with description) for both `good.metrics` and `total.metrics`
- [ ] 2.2 In `models_metric_custom_indicator.go`, update `buildGoodMetricItem` to dispatch on `aggregation = "doc_count"`: use `SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1` (no field) for doc_count, keep existing `Metrics0` path for all other aggregations
- [ ] 2.3 In `models_metric_custom_indicator.go`, apply the same dispatch to `buildTotalMetricItem` using `SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1`

## 3. Acceptance Test

- [ ] 3.1 Create `internal/kibana/slo/testdata/TestAccResourceSlo_metric_custom_indicator_doc_count/test/test.tf` — SLO with `metric_custom_indicator` using `doc_count` for good (with filter) and total (no filter, no field)
- [ ] 3.2 Add `TestAccResourceSlo_metric_custom_indicator_doc_count` to `acc_test.go` (after `TestAccResourceSlo_timeslice_metric_indicator_multiple_mixed_metrics`) — single-step test with `SkipFunc: versionutils.CheckIfVersionIsUnsupported(sloTimesliceMetricsMinVersion)`, asserting index, metric names, aggregations, filter presence, field absence, and equations

## 4. Verification

- [ ] 4.1 Run `make build` and confirm it compiles cleanly
- [ ] 4.2 Run unit tests: `go test ./internal/kibana/slo/... -run TestMetricCustomIndicator` and confirm all subtests pass
