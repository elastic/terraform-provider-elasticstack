## 1. New panel package scaffold

- [ ] 1.1 Create `internal/kibana/dashboard/panel/mlanomalycharts/` package with files: `schema.go`, `model.go`, `api.go` following the `sloburnrate` layout.
- [ ] 1.2 Implement `SchemaAttribute()` in `schema.go`:
  - Compose `panelkit.PanelPresentationAttributes()` with `job_ids` (required `ListAttribute(ElementType: types.StringType)`), `max_series_to_plot` (optional `Float64Attribute`), `severity_threshold` (optional `ListNestedAttribute`), and `time_range` (`panelkit.TimeRangeSchema(...)`).
  - `severity_threshold` list item attributes: `severity` (optional string, enum validator: `low`, `warning`, `minor`, `major`, `critical`), `min` (optional int64), `max` (optional int64).
  - Plan-time item-level validator: exactly one of `severity` or `min` must be set per item; `max` requires `min` (or `severity` is absent).
  - Return `panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{BlockName: "ml_anomaly_charts_config", PanelType: "ml_anomaly_charts", ...})`.

## 2. Model structs

- [ ] 2.1 Create `internal/kibana/dashboard/models/mlanomalycharts.go` with:
  - `MlAnomalyChartsConfigModel` — fields: `JobIDs types.List`, `MaxSeriesToPlot types.Float64`, `SeverityThreshold []MlAnomalyChartsSeverityThresholdModel`, `TimeRange *TimeRangeModel`, plus the four panelkit presentation fields (`Title`, `Description`, `HideTitle`, `HideBorder`).
  - `MlAnomalyChartsSeverityThresholdModel` — fields: `Severity types.String`, `Min types.Int64`, `Max types.Int64`.
- [ ] 2.2 Add `MlAnomalyChartsConfig *MlAnomalyChartsConfigModel` field to `PanelModel` in `internal/kibana/dashboard/models/panels.go` (or equivalent panel model file).

## 3. API mapping

- [ ] 3.1 Implement `BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts) diag.Diagnostics` in `mlanomalycharts/api.go`:
  - Map `job_ids`, `max_series_to_plot`, `title`, `description`, `hide_title`, `hide_border`, `time_range` to the `KibanaHTTPAPIsMlAnomalyCharts` struct.
  - For `severity_threshold`: implement a helper `buildSeverityThresholdItems` that converts each `MlAnomalyChartsSeverityThresholdModel` to the correct `json.RawMessage` payload:
    - Named `severity`: marshal the corresponding canonical struct (e.g., `{min: 0, max: 2}` for `low`). The `critical` band has no `max` field — use `SeverityThreshold4{Min: 75}`.
    - Raw range: marshal `struct{ Min int64; Max *int64 }{min, max}`.
- [ ] 3.2 Implement `PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsMlAnomalyCharts) diag.Diagnostics` in `mlanomalycharts/api.go`:
  - On import (prior == nil): populate all fields from API unconditionally.
  - On read with prior: apply REQ-009 null-preservation for all optional fields.
  - For `severity_threshold` on read: unmarshal each union item; try to match `{min, max}` against the five canonical bands; store as named severity if it matches, raw range otherwise.
  - `time_range` null-preservation: keep `mode` null if it was null in prior state.

## 4. Registration

- [ ] 4.1 Register `mlanomalycharts.SchemaAttribute()` in `internal/kibana/dashboard/schema.go` alongside the other `*_config` blocks.
- [ ] 4.2 Add the handler to `panelHandlers` in `internal/kibana/dashboard/registry.go` (or equivalent handler registration map).
- [ ] 4.3 Add `"ml_anomaly_charts"` to the reject-config_json panel type list in the panel routing logic (REQ-010 enforcement).

## 5. Tests

- [ ] 5.1 Unit tests in `internal/kibana/dashboard/panel/mlanomalycharts/` (file: `api_test.go` or `model_test.go`):
  - Named severity round-trip: write `severity = "major"` → verify API payload `{min: 50, max: 74}`; read back and verify state shows `severity = "major"`.
  - `critical` round-trip (open-ended): write `severity = "critical"` → verify API payload `{min: 75}` (no `max`); read back → `severity = "critical"`.
  - Raw range round-trip: write `min = 10, max = 20` → verify API payload `{min: 10, max: 20}`; read back → `min = 10, max = 20` (not coerced to a named severity).
  - Canonical collision: API returns `{min: 3, max: 24}` → read maps to `severity = "warning"`.
  - Null-preservation: prior state omits `max_series_to_plot`; API returns a server value → state keeps null.
  - `time_range` null-preservation: prior state has `mode = null`; API omits `mode` → state keeps null.
  - Plan-time validator: item with both `severity` and `min` → error; item with `max` but no `min`/`severity` → error; item with neither → error.
- [ ] 5.2 Acceptance test in `internal/kibana/dashboard/panel/mlanomalycharts/acc_test.go`:
  - Test 1: create a dashboard with an `ml_anomaly_charts` panel using named severities (`critical`, `major`); verify plan is stable after apply.
  - Test 2: create a dashboard with a raw-range severity threshold `{min: 10, max = 20}`; verify plan is stable.
  - Test 3: update `job_ids` in-place; verify update does not destroy and recreate the dashboard.

## 6. Delta spec and validation

- [ ] 6.1 Ensure `openspec/changes/dashboard-ml-anomaly-charts/specs/kibana-dashboard/spec.md` defines REQ-047 with scenarios.
- [ ] 6.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate dashboard-ml-anomaly-charts --type change`; fix any reported issues.
- [ ] 6.3 Run `make build` and `go test ./internal/kibana/dashboard/...` (without `TF_ACC` for unit tests).
- [ ] 6.4 Run `make check-openspec`.
