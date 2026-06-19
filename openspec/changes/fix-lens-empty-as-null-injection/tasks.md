## 1. Verify API schema scope

- [x] 1.1 Confirm in `generated/kbapi/kibana.gen.go` which XY `KibanaHTTPAPIsXyY*` metric structs define an `EmptyAsNull` field (expected: Count, Sum, UniqueCount only)
- [x] 1.2 Confirm in `generated/kbapi/kibana.gen.go` which datatable `KibanaHTTPAPIsDatatableMetric*` structs define `EmptyAsNull` (expected: Count, Sum, UniqueCount only)
- [x] 1.3 Check the tagcloud / region-map / partition metric API structs reached via `populateFieldMetricLensDefaults` to determine whether the same gate applies to those families (confirmed: gauge, pie, legacy-metric, metric-chart, tagcloud, treemap, mosaic all share the same Count/Sum/UniqueCount-only support set)

## 2. Implement the gated injection

- [x] 2.1 Add `operationSupportsEmptyAsNull(operation string) bool` to `internal/kibana/dashboard/lenscommon/populate_lens_charts.go` returning true only for `count`, `sum`, `unique_count`
- [x] 2.2 Gate the `empty_as_null` injection in `PopulateLensMetricDefaults` on `operationSupportsEmptyAsNull(model["operation"])`
- [x] 2.3 Apply the same gate in `populateFieldMetricLensDefaults` for the chart families confirmed in 1.3 to omit `empty_as_null` for the affected operations; leave families whose schema legitimately supports it unchanged
- [x] 2.4 Confirm no other injector (`PopulateGaugeMetricDefaults`, `PopulatePieChartMetricDefaults`, `PopulateLegacyMetricMetricDefaults`) emits `empty_as_null` for an operation whose API schema rejects it; adjust only where the generated type confirms the omission (all three confirmed and gated)

## 3. Update the reproduction test into a regression test

- [x] 3.1 Convert `TestAccReproduceIssue3707` in `internal/kibana/dashboard/panel/lensxy/issue_3707_acc_test.go` from `ExpectError` to a successful apply step (no error) with a follow-up plan-only step asserting no diff
- [x] 3.2 Keep the `percentile_bar_horizontal` fixture; ensure the data view dependency and required variables resolve for a successful apply (removed the explicit `empty_as_null = true` so the provider's gated injection is exercised)
- [x] 3.3 Add an acceptance case (new step or fixture) for another previously-broken operation (`average`) confirming a clean apply with no diff
- [x] 3.4 Add an acceptance case for `count` confirming `empty_as_null` is still emitted and round-trips without drift

## 4. Unit coverage

- [x] 4.1 Add unit tests in `internal/kibana/dashboard/lenscommon` asserting `PopulateLensMetricDefaults` injects `empty_as_null` for `count`/`sum`/`unique_count`
- [x] 4.2 Add unit tests asserting `PopulateLensMetricDefaults` does NOT inject `empty_as_null` for `percentile`, `percentile_rank`, `average`, `median`, `standard_deviation`, `last_value`
- [x] 4.3 Add a unit test for `operationSupportsEmptyAsNull` covering supported and unsupported operations

## 5. Validate

- [x] 5.1 `make build` succeeds
- [x] 5.2 Run `go test` for the `lenscommon` unit tests
- [x] 5.3 Run the targeted acceptance tests (`TestAccReproduceIssue3707` and new cases) against the local Elastic stack
- [x] 5.4 Update `CHANGELOG.md` Unreleased section with the bug fix entry referencing issue #3707
- [x] 5.5 Run `make check-openspec` / `openspec validate` for the change
