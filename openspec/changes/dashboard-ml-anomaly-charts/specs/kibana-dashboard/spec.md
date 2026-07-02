## ADDED Requirements

### Requirement: ML anomaly charts panel behavior (REQ-053)

The resource SHALL support `type = "ml_anomaly_charts"` panels through the typed `ml_anomaly_charts_config` block. When a panel entry sets `type = "ml_anomaly_charts"`, the resource SHALL require the `ml_anomaly_charts_config` block and SHALL return an error diagnostic when it is absent.

The block accepts the following attributes:

- `job_ids` (required `list(string)`, min 1 item): one or more anomaly-detection job IDs or group IDs whose results are shown. The provider treats these as opaque strings and does not validate their existence against Kibana's ML API at plan time; invalid job IDs surface as Kibana API errors during `terraform apply`.
- `max_series_to_plot` (optional float64): maximum number of anomaly series to plot. When null in state, the attribute is omitted from the API request.
- `severity_threshold` (optional list of objects, min 1 item when present): filters which severity bands are displayed. Each list item is a union — exactly one of the following must be set per item:
  - `severity` (string, one of `low`, `warning`, `minor`, `major`, `critical`): a named severity shortcut. The model layer SHALL expand named severities to their canonical `{min, max}` API pairs at write time.
  - `min` (int64) plus optional `max` (int64): a raw numeric range. `max` may be set only when `min` is set and `severity` is unset; when `max` is set, `min` must also be set. Setting both `severity` and `min` on the same item SHALL produce an error diagnostic at plan time. Setting `severity` together with `max` SHALL produce an error diagnostic at plan time.
- `title` (optional string): panel title. Subject to REQ-009 null-preservation.
- `description` (optional string): panel description. Subject to REQ-009 null-preservation.
- `hide_title` (optional bool): when true, hides the panel title. Subject to REQ-009 null-preservation.
- `hide_border` (optional bool): when true, hides the panel border. Subject to REQ-009 null-preservation.
- `time_range` (optional object: `from` string required, `to` string required, `mode` string optional): a panel-level time range override, identical in shape to the dashboard root `time_range`. Reuses `panelkit.TimeRangeSchema`. Subject to REQ-009 null-preservation: when prior state has `time_range` null, the provider SHALL keep it null even if the API returns a default; when `mode` is null in prior state, the provider SHALL keep `mode` null.

The model layer SHALL expand named severity values to the following canonical `{min, max}` API pairs (matching the generated Kibana OpenAPI const values in `KibanaHTTPAPIsMlAnomalyChartsSeverityThreshold0`–`SeverityThreshold4`):

| `severity` | API `min` | API `max`    |
|---|---|---|
| `low`       | 0         | 3            |
| `warning`   | 3         | 25           |
| `minor`     | 25        | 50           |
| `major`     | 50        | 75           |
| `critical`  | 75        | (omitted — open-ended upper bound) |

On write, the provider SHALL map `ml_anomaly_charts_config` to the `config` object in the `KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts` API schema. Optional fields SHALL be included only when set in state; absent optional fields SHALL NOT be sent to the API.

On read, the provider SHALL repopulate `ml_anomaly_charts_config` from the API response using REQ-009 null-preservation, extended to the **representation form** of `severity_threshold` items. The API encodes `severity_threshold` as `{min, max}` pairs only; it conveys no information about whether the practitioner authored a named `severity` or a raw numeric range. Therefore the chosen form is recovered from prior state, not inferred by normalizing:

- When the prior item holds a named `severity` (and `min`/`max` are null), the provider SHALL store the named form, deriving the label from the API `{min, max}` pair via the canonical-band table. The `critical` band (API: `{min: 75}`, no `max` field) SHALL map to `severity = "critical"`. If the API value no longer matches any canonical band, the provider SHALL fall back to the raw `min`/`max` form (surfacing as drift).
- When the prior item holds raw `min`/`max` (and `severity` is null), the provider SHALL store the raw `min`/`max` verbatim from the API, even when the pair coincidentally equals a canonical band.
- On import (no prior state), the provider SHALL default to the named form when the API `{min, max}` matches a canonical band, and to the raw form otherwise.

The provider SHALL NOT normalize a practitioner-authored raw range into a named `severity` on read. While the configured values match current state, a subsequent plan SHALL show no changes.

`config_json` SHALL NOT be supported for `ml_anomaly_charts` panels; using `config_json` with `type = "ml_anomaly_charts"` SHALL return an error diagnostic (per REQ-010 policy).

Implementation: new package `internal/kibana/dashboard/panel/mlanomalycharts/` with `schema.go`, `model.go`, and `api.go`; new model file `internal/kibana/dashboard/models/mlanomalycharts.go`; registration in `internal/kibana/dashboard/schema.go` and `internal/kibana/dashboard/registry.go`.

#### Scenario: Creation of ml_anomaly_charts panel with named severities

- GIVEN a panel with `type = "ml_anomaly_charts"` and `ml_anomaly_charts_config` containing `job_ids = ["my-job"]` and `severity_threshold = [{ severity = "critical" }, { severity = "major" }]`
- WHEN create runs
- THEN the provider SHALL send `job_ids = ["my-job"]` and `severity_threshold = [{min: 75}, {min: 50, max: 75}]` in the API request
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

#### Scenario: Raw range coinciding with a canonical band is preserved (no diff)

- GIVEN a panel where the practitioner set `severity_threshold = [{ min = 3, max = 25 }]` (coincides with the `warning` canonical band)
- WHEN create runs and the post-apply read returns `{min: 3, max: 25}`
- THEN the provider SHALL store `min = 3` and `max = 25` in state (NOT coerced to `severity = "warning"`)
- AND a subsequent plan SHALL show no changes

#### Scenario: severity_threshold form is preserved across refresh

- GIVEN state holds `severity_threshold = [{ severity = "major" }, { min = 10, max = 20 }]`
- WHEN a refresh runs and the API returns `[{min: 50, max: 75}, {min: 10, max: 20}]`
- THEN state SHALL retain the first item as `severity = "major"` and the second as `min = 10, max = 20`
- AND a subsequent plan SHALL show no changes

#### Scenario: critical severity preserved in raw form when authored raw

- GIVEN a panel where the practitioner set `severity_threshold = [{ min = 75 }]` (raw form, coincides with the `critical` canonical band)
- WHEN create runs and the post-apply read returns `{min: 75}` (no `max` field)
- THEN the provider SHALL store `min = 75` with `max` null in state (NOT coerced to `severity = "critical"`)
- AND a subsequent plan SHALL show no changes

#### Scenario: Switching severity form is a configuration change

- GIVEN state holds `severity_threshold = [{ severity = "warning" }]`
- WHEN the configuration changes to `{ min = 3, max = 25 }` (same band, raw form)
- THEN the plan SHALL report a change for that item
- AND after apply the state SHALL settle to `{ min = 3, max = 25 }` with a subsequent plan showing no changes

#### Scenario: Import defaults to named form for canonical bands

- GIVEN an existing panel whose API `severity_threshold` is `[{min: 3, max: 25}]` and no prior Terraform state
- WHEN the panel is imported
- THEN state SHALL store `severity = "warning"` (named form preferred only on import, where no practitioner form exists to preserve)

#### Scenario: Plan-time validation — both severity and min set

- GIVEN a `severity_threshold` item with both `severity = "major"` and `min = 50`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating that exactly one of `severity` or `min` must be set

#### Scenario: Plan-time validation — max without min

- GIVEN a `severity_threshold` item with `max = 75` but neither `severity` nor `min` set
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
