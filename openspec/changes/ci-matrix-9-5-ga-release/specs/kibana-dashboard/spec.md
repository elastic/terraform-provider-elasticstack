## MODIFIED Requirements

### Requirement: ML anomaly charts panel behavior (REQ-053)

The resource SHALL support `type = "ml_anomaly_charts"` panels through the typed `ml_anomaly_charts_config` block. When a panel entry sets `type = "ml_anomaly_charts"`, the resource SHALL require the `ml_anomaly_charts_config` block and SHALL return an error diagnostic when it is absent.

The block accepts the following attributes:

- `job_ids` (required `list(string)`, min 1 item): one or more anomaly-detection job IDs or group IDs whose results are shown. The provider treats these as opaque strings and does not validate their existence against Kibana's ML API at plan time; invalid job IDs surface as Kibana API errors during `terraform apply`.
- `max_series_to_plot` (optional int64): maximum number of anomaly series to plot. When null in state, the attribute is omitted from the API request. The Kibana API represents this field as a JSON number (`*float32` in the generated client); the provider exposes it as an integer since a series count cannot be fractional, converting to/from the API's numeric type at the boundary.
- `severity_threshold` (optional list of objects, min 1 item when present): filters which severity bands are displayed. Each list item is a union — exactly one of the following must be set per item:
  - `severity` (string, one of `low`, `warning`, `minor`, `major`, `critical`): a named severity shortcut. The model layer SHALL expand named severities to their canonical `{min, max}` API pairs at write time.
  - `min` (int64) plus optional `max` (int64): an alternative, numeric spelling of one of the canonical severity bands. The Kibana API constrains `min` to one of the canonical band start values (`0`, `3`, `25`, `50`, or `75` — see the canonical-band table below); the provider does not validate this constraint client-side, so a `min` value that does not match one of these boundaries SHALL be rejected by the Kibana API at `terraform apply` time with an HTTP 400 error, not caught earlier as a plan-time diagnostic. `max` may be set only when `min` is set and `severity` is unset; when `max` is set, `min` must also be set. Setting both `severity` and `min` on the same item SHALL produce an error diagnostic at plan time. Setting `severity` together with `max` SHALL produce an error diagnostic at plan time.
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

The raw `min`/`max` form exists so a practitioner can spell one of these same five bands numerically instead of via the `severity` enum; it is not a general-purpose custom-range escape hatch. Kibana's Dashboard API validates `min` against the canonical band list server-side — a `min` outside `{0, 3, 25, 50, 75}` is rejected with HTTP 400 regardless of stack version, and the provider has no client-side equivalent of that validation, so the failure surfaces only at `terraform apply`.

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

#### Scenario: Raw range with a non-canonical min is rejected by the API

- GIVEN a panel with `severity_threshold = [{ min = 10, max = 20 }]` (a `min` value that does not match any canonical band boundary)
- WHEN create runs
- THEN the Kibana API SHALL reject the request with an HTTP 400 error identifying `severity_threshold` as invalid
- AND `terraform apply` SHALL fail with that error, since the provider does not validate the canonical-boundary constraint client-side

#### Scenario: Raw range coinciding with a canonical band is preserved (no diff)

- GIVEN a panel where the practitioner set `severity_threshold = [{ min = 3, max = 25 }]` (coincides with the `warning` canonical band)
- WHEN create runs and the post-apply read returns `{min: 3, max: 25}`
- THEN the provider SHALL store `min = 3` and `max = 25` in state (NOT coerced to `severity = "warning"`)
- AND a subsequent plan SHALL show no changes
