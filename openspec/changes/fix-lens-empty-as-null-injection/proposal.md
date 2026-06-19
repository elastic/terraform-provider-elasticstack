## Why

The shared Lens metric-default normalization injects `empty_as_null = false` into every metric `config_json`, regardless of the metric operation. The Kibana dashboard API only accepts `empty_as_null` on a subset of metric operations (`count`, `sum`, `unique_count`). For other operations — `percentile`, `percentile_rank`, `average`, `min`, `max`, `median`, `standard_deviation`, `last_value` — Kibana rejects the request with HTTP 400 (`Additional properties are not allowed ('empty_as_null' was unexpected)`). This makes those operations unusable in XY and datatable panels (issue [#3707](https://github.com/elastic/terraform-provider-elasticstack/issues/3707)).

## What Changes

- Gate the `empty_as_null` default injection in the shared Lens metric normalization so it is only applied for operations whose Kibana API schema accepts the property (`count`, `sum`, `unique_count`).
- Stop injecting `empty_as_null` for operations whose API schema does not define it (`percentile`, `percentile_rank`, `average`, `min`, `max`, `median`, `standard_deviation`, `last_value`), preventing the HTTP 400 on apply.
- Keep `empty_as_null` semantic-equality / drift handling intact for the supported operations so existing round-trips are unchanged.
- Convert the existing reproduction test `TestAccReproduceIssue3707` from an error-expectation into a passing regression test, and add coverage for at least one other previously-broken operation plus a unit test asserting the gated injection.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `kibana-dashboard`: The panel default normalization requirement (REQ-011) changes so that the `empty_as_null` Lens metric default is injected only for metric operations whose Kibana API schema accepts it, instead of unconditionally for all field-metric operations.

## Impact

- Code: `internal/kibana/dashboard/lenscommon/populate_lens_charts.go` (`PopulateLensMetricDefaults` and the shared field-metric default helper used by XY, datatable, and other Lens charts).
- Affected panels: XY chart (`xy_chart_config`) and datatable (`datatable_config`) `config_json` metrics using the previously-broken operations.
- Tests: `internal/kibana/dashboard/panel/lensxy/issue_3707_acc_test.go` and associated fixture; new unit coverage in `lenscommon`.
- No schema/API changes; behavior-only fix to request shaping. No breaking changes for existing valid configurations.
