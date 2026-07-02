## ADDED Requirements

### Requirement: ML anomaly charts panel behavior (REQ-047)

The resource SHALL support `type = "ml_anomaly_charts"` panels through the typed `ml_anomaly_charts_config` block. When a panel entry sets `type = "ml_anomaly_charts"`, the resource SHALL require the `ml_anomaly_charts_config` block and SHALL return an error diagnostic when it is absent.

The block accepts the following attributes:

- `job_ids` (required `list(string)`, min 1 item): one or more anomaly-detection job IDs or group IDs whose results are shown. The provider treats these as opaque strings and does not validate their existence against Kibana's ML API at plan time; invalid job IDs surface as Kibana API errors during `terraform apply`.
- `max_series_to_plot` (optional float64): maximum number of anomaly series to plot. When null in state, the attribute is omitted from the API request.
- `severity_threshold` (optional list of objects, min 1 item when present): filters which severity bands are displayed. Each list item is a union — exactly one of the following must be set per item:
  - `severity` (string, one of `low`, `warning`, `minor`, `major`, `critical`): a named severity shortcut. The model layer SHALL expand named severities to their canonical `{min, max}` API pairs at write time.
  - `min` (int64) plus optional `max` (int64): a raw numeric range. `max` MAY be set only when `min` is set (and `severity` is unset). Setting `severity` together with `min` or `max` on the same item SHALL produce an error diagnostic at plan time.
- `title` (optional string): panel title. Subject to REQ-009 null-preservation.
- `description` (optional string): panel description. Subject to REQ-009 null-preservation.
- `hide_title` (optional bool): when true, hides the panel title. Subject to REQ-009 null-preservation.
- `hide_border` (optional bool): when true, hides the panel border. Subject to REQ-009 null-preservation.
- `time_range` (optional object: `from` string required, `to` string required, `mode` string optional): a panel-level time range override, identical in shape to the dashboard root `time_range`. Reuses `panelkit.TimeRangeSchema`. Subject to REQ-009 null-preservation: when prior state has `time_range` null, the provider SHALL keep it null even if the API returns a default; when `mode` is null in prior state, the provider SHALL keep `mode` null.

The model layer SHALL expand named severity values to the following canonical `{min, max}` API pairs:

| `severity` | API `min` | API `max`    |
|---|---|---|
| `low`       | 0         | 2            |
| `warning`   | 3         | 24           |
| `minor`     | 25        | 49           |
| `major`     | 50        | 74           |
| `critical`  | 75        | (omitted — open-ended upper bound) |

On write, the provider SHALL map `ml_anomaly_charts_config` to the `config` object in the `KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts` API schema. Optional fields SHALL be included only when set in state; absent optional fields SHALL NOT be sent to the API.

On read, the provider SHALL repopulate `ml_anomaly_charts_config` from the API response using REQ-009 null-preservation. For `severity_threshold` items: if the API returns a `{min, max}` pair that matches a canonical band, the provider SHALL store it as the named `severity` string; otherwise it SHALL store the raw `min`/`max` integers. The `critical` band (API: `{min: 75}`, no `max` field) SHALL map to `severity = "critical"` on read.

`config_json` SHALL NOT be supported for `ml_anomaly_charts` panels; using `config_json` with `type = "ml_anomaly_charts"` SHALL return an error diagnostic (per REQ-010 policy).

Implementation: new package `internal/kibana/dashboard/panel/mlanomalycharts/` with `schema.go`, `model.go`, and `api.go`; new model file `internal/kibana/dashboard/models/mlanomalycharts.go`; registration in `internal/kibana/dashboard/schema.go` and `internal/kibana/dashboard/registry.go`.

#### Scenario: Creation of ml_anomaly_charts panel with named severities

- GIVEN a panel with `type = "ml_anomaly_charts"` and `ml_anomaly_charts_config` containing `job_ids = ["my-job"]` and `severity_threshold = [{ severity = "critical" }, { severity = "major" }]`
- WHEN create runs
- THEN the provider SHALL send `job_ids = ["my-job"]` and `severity_threshold = [{min: 75}, {min: 50, max: 74}]` in the API request
- AND after the post-apply read, state SHALL represent both items as named severities
- AND a subsequent plan SHALL show no changes

#### Scenario: Round-trip stability for critical (open-ended) severity

- GIVEN a panel with `severity_threshold = [{ severity = "critical" }]` applied and read back
- WHEN the API returns `severity_threshold = [{min: 75}]` (no `max` field)
- THEN the provider SHALL map this to `severity = "critical"` in state
- AND a subsequent plan SHALL show no changes

#### Scenario: Raw range escape hatch

- GIVEN a panel with `severity_threshold = [{ min = 10, max = 20 }]`
- WHEN create runs and the post-apply read returns `{min: 10, max: 20}`
- THEN state SHALL contain `min = 10` and `max = 20` (not coerced to a named severity)
- AND a subsequent plan SHALL show no changes

#### Scenario: Named severity collision round-trip

- GIVEN a panel where the practitioner set `severity_threshold = [{ min = 3, max = 24 }]` (matches `warning` canonical band)
- WHEN the post-apply read occurs
- THEN the provider SHALL store `severity = "warning"` in state (canonical form preferred)
- AND a subsequent plan against the original `{ min = 3, max = 24 }` configuration MAY show a diff (the read normalizes to named form)

#### Scenario: Plan-time validation — both severity and min set

- GIVEN a `severity_threshold` item with both `severity = "major"` and `min = 50`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating that exactly one of `severity` or `min` must be set

#### Scenario: Plan-time validation — max without min

- GIVEN a `severity_threshold` item with `max = 74` but neither `severity` nor `min` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic

#### Scenario: config_json rejected for ml_anomaly_charts

- GIVEN a panel with `type = "ml_anomaly_charts"` and `config_json = "{}"`
- WHEN Terraform validates or applies the configuration
- THEN the resource SHALL return an error diagnostic indicating that `config_json` is not supported for `ml_anomaly_charts` panels

#### Scenario: Optional fields follow null-preservation

- GIVEN an `ml_anomaly_charts_config` that does not set `max_series_to_plot` or `time_range`
- WHEN apply runs and the post-apply read returns server-side defaults for those fields
- THEN state SHALL keep `max_series_to_plot` and `time_range` as null
- AND a subsequent plan SHALL show no changes

#### Scenario: Update job_ids in-place

- GIVEN an existing `ml_anomaly_charts` panel with `job_ids = ["job-a"]`
- WHEN the configuration changes to `job_ids = ["job-a", "job-b"]` and update runs
- THEN the resource SHALL NOT destroy and recreate the dashboard
- AND state SHALL reflect both job IDs after the update
