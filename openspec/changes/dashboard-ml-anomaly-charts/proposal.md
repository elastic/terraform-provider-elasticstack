## Why

The `elasticstack_kibana_dashboard` resource cannot embed the `ml_anomaly_charts` panel type. Users who manage ML anomaly detection jobs as infrastructure — and who often already provision those jobs via the provider — cannot surface the resulting anomaly visualizations on their dashboards without falling back to raw panel-level `config_json`. This breaks the typed-configuration story for the most common ML → dashboard workflow.

The Kibana Dashboard API defines a first-class `ml_anomaly_charts` panel type whose configuration is a flat, non-union struct (`KibanaHTTPAPIsMlAnomalyCharts`), making it a straightforward addition that follows the existing panelkit pattern.

## What Changes

- Add a new `ml_anomaly_charts_config` typed block on the `panels[]` schema entry for panels with `type = "ml_anomaly_charts"`.
- The block accepts:
  - `job_ids` (required `list(string)`) — one or more anomaly-detection job IDs or group IDs.
  - `max_series_to_plot` (optional float64) — maximum number of anomaly series to plot.
  - `severity_threshold` (optional list of objects) — filters which severity bands are displayed. Each item is a union of either a named severity shortcut (`severity` string enum: `low`, `warning`, `minor`, `major`, `critical`) OR a raw numeric range (`min` int, optional `max` int); exactly one of the two must be set per item.
  - `title`, `description` (optional strings) — panel presentation fields via `PanelPresentationAttributes()`.
  - `hide_title`, `hide_border` (optional booleans) — panel presentation fields via `PanelPresentationAttributes()`.
  - `time_range` (optional object: `from`, `to`, optional `mode`) — reuses `panelkit.TimeRangeSchema`, identical shape to the dashboard root `time_range`.
- The model layer expands named severities to canonical `{min, max}` API pairs at write time (matching the generated Kibana OpenAPI const values):

| `severity` value | Expanded API payload |
|---|---|
| `low`      | `{min: 0,  max: 3}`  |
| `warning`  | `{min: 3,  max: 25}` |
| `minor`    | `{min: 25, max: 50}` |
| `major`    | `{min: 50, max: 75}` |
| `critical` | `{min: 75}` (open-ended) |

- On read, the provider preserves the practitioner-chosen form of each `severity_threshold` item — storing the named `severity` when the practitioner configured a named severity, and the raw `min`/`max` when they configured limits — recovered from prior state so a stable configuration produces no plan diff. Normalization to the named form is permitted only on import, where no prior form exists to preserve.
- Apply REQ-009 null-preservation semantics on all optional fields (consistent with every other flat-config panel).
- `config_json` is rejected for `ml_anomaly_charts` panels (consistent with REQ-010 policy).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: extend the dashboard panel registry with a new `ml_anomaly_charts` panel handler (REQ-053).

## Impact

- New package `internal/kibana/dashboard/panel/mlanomalycharts/` — schema, model, and API mapping following the `sloburnrate`/`syntheticsmonitors` pattern.
- `internal/kibana/dashboard/schema.go` — register `ml_anomaly_charts_config` SchemaAttribute.
- `internal/kibana/dashboard/registry.go` — register the new handler in `panelHandlers`.
- `internal/kibana/dashboard/models/panels.go` (or equivalent) — add `MlAnomalyChartsConfig *MlAnomalyChartsConfigModel` to `PanelModel`.
- `internal/kibana/dashboard/models/mlanomalycharts.go` — new model structs.
- Unit tests in `mlanomalycharts/` — schema round-trip, severity expansion, null-preservation.
- Acceptance tests — at least one test exercising `ml_anomaly_charts` with named severities and one with raw ranges.
