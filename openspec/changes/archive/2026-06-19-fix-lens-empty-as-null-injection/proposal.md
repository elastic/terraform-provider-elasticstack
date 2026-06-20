## Why

The Lens metric-default normalization injects `empty_as_null = false` into every metric `config_json`, regardless of the metric operation. The Kibana dashboard API only accepts `empty_as_null` on a subset of metric operations (`count`, `sum`, `unique_count`). For other operations — `percentile`, `percentile_rank`, `average`, `min`, `max`, `median`, `standard_deviation`, `last_value`, and pipeline operations like `formula`/`moving_average` — Kibana rejects the request with HTTP 400 (`Additional properties are not allowed ('empty_as_null' was unexpected)`). This was reported for XY `percentile` (issue [#3707](https://github.com/elastic/terraform-provider-elasticstack/issues/3707)), but verification against the generated `kbapi` types shows the same defect across every Lens chart family that injects `empty_as_null` (XY, datatable, metric chart, pie, gauge, legacy metric, tagcloud, treemap, mosaic, region map), since all share the same metric schema where only `count`/`sum`/`unique_count` define the property.

## What Changes

- Add a single `operationSupportsEmptyAsNull` allowlist (`count`, `sum`, `unique_count`) in `lenscommon`.
- Gate the `empty_as_null` default injection on that allowlist at every injection site: `PopulateLensMetricDefaults` (XY, datatable, metric chart), `populateFieldMetricLensDefaults` (tagcloud, region map, partition metrics), `PopulateGaugeMetricDefaults`, `PopulatePieChartMetricDefaults`, and `PopulateLegacyMetricMetricDefaults`.
- Stop injecting `empty_as_null` for operations whose API schema does not define it, preventing the HTTP 400 on apply.
- Keep `empty_as_null` semantic-equality / drift handling intact for the supported operations so existing round-trips are unchanged.
- Convert the existing reproduction test `TestAccReproduceIssue3707` from an error-expectation into a passing regression test, and add coverage for at least one other previously-broken operation plus unit tests asserting the gated injection.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `kibana-dashboard`: The panel default normalization requirement (REQ-011) changes so that the `empty_as_null` Lens metric default is injected only for metric operations whose Kibana API schema accepts it, instead of unconditionally for all field-metric operations.

## Impact

- Code: `internal/kibana/dashboard/lenscommon/populate_lens_charts.go` — new `operationSupportsEmptyAsNull` helper and gated `empty_as_null` injection in `PopulateLensMetricDefaults`, `populateFieldMetricLensDefaults`, `PopulateGaugeMetricDefaults`, `PopulatePieChartMetricDefaults`, and `PopulateLegacyMetricMetricDefaults`.
- Affected panels: every typed Lens chart (`xy_chart_config`, `datatable_config`, `metric_chart_config`, `pie_chart_config`, `gauge_config`, `legacy_metric_config`, `tagcloud_config`, `region_map_config`, `treemap_config`, `mosaic_config`) whose metric `config_json` uses a previously-broken operation.
- Tests: `internal/kibana/dashboard/panel/lensxy/issue_3707_acc_test.go` and associated fixture; new unit coverage in `lenscommon`.
- No schema/API changes; behavior-only fix to request shaping. No breaking changes for existing valid configurations (`count`/`sum`/`unique_count` behavior is unchanged).
