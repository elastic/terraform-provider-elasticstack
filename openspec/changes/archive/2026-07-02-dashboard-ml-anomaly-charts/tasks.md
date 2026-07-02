## 1. New panel package scaffold

- [x] 1.1 Create `internal/kibana/dashboard/panel/mlanomalycharts/` package with files: `schema.go`, `model.go`, `api.go` following the `sloburnrate` layout.
- [x] 1.2 Implement `SchemaAttribute()` in `schema.go`:
  - Compose `panelkit.PanelPresentationAttributes()` with `job_ids` (required `ListAttribute(ElementType: types.StringType)`), `max_series_to_plot` (optional `Float64Attribute`), `severity_threshold` (optional `ListNestedAttribute`), and `time_range` (`panelkit.TimeRangeSchema(...)`).
  - `severity_threshold` list item attributes: `severity` (optional string, enum validator: `low`, `warning`, `minor`, `major`, `critical`), `min` (optional int64), `max` (optional int64).
  - Plan-time item-level validator: exactly one of `severity` or `min` must be set per item; `max` may be set only when `min` is set and `severity` is unset; `severity` together with `max` is a plan-time error.
  - Return `panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{BlockName: "ml_anomaly_charts_config", PanelType: "ml_anomaly_charts", ...})`.

## 2. Model structs

- [x] 2.1 Create `internal/kibana/dashboard/models/mlanomalycharts.go` with:
  - `MlAnomalyChartsConfigModel` — fields: `JobIDs types.List`, `MaxSeriesToPlot types.Float64`, `SeverityThreshold []MlAnomalyChartsSeverityThresholdModel`, `TimeRange *TimeRangeModel`, plus the four panelkit presentation fields (`Title`, `Description`, `HideTitle`, `HideBorder`).
  - `MlAnomalyChartsSeverityThresholdModel` — fields: `Severity types.String`, `Min types.Int64`, `Max types.Int64`.
- [x] 2.2 Add `MlAnomalyChartsConfig *MlAnomalyChartsConfigModel` field to `PanelModel` in `internal/kibana/dashboard/models/panels.go` (or equivalent panel model file).

## 3. API mapping

- [x] 3.1 Implement `BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts) diag.Diagnostics` in `mlanomalycharts/api.go`:
  - Map `job_ids`, `max_series_to_plot`, `title`, `description`, `hide_title`, `hide_border`, `time_range` to the `KibanaHTTPAPIsMlAnomalyCharts` struct.
  - For `severity_threshold`: implement a helper `buildSeverityThresholdItems` that converts each `MlAnomalyChartsSeverityThresholdModel` to the correct `json.RawMessage` payload:
    - Named `severity`: marshal the corresponding canonical struct (e.g., `{min: 0, max: 3}` for `low`). The `critical` band has no `max` field — use `SeverityThreshold4{Min: 75}`.
    - Raw range: marshal `struct{ Min int64; Max *int64 }{min, max}`.
- [x] 3.2 Implement `PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsMlAnomalyCharts) diag.Diagnostics` in `mlanomalycharts/api.go`:
  - On import (prior == nil): populate all fields from API unconditionally.
  - On read with prior: apply REQ-009 null-preservation for all optional fields.
  - For `severity_threshold` on read: implement a `severityThresholdFromAPI(apiItem, priorItem *MlAnomalyChartsSeverityThresholdModel)` helper that mirrors `sloStringFromAPIOrPrior` and resolves the **representation form from prior state**, not from the API:
    - **prior named** (`prior.Severity` known, `Min`/`Max` null): store named form; derive the label from the API `{min,max}` via the canonical-band table. If the API value no longer matches any canonical band, fall back to raw `min`/`max` (drift).
    - **prior raw** (`prior.Min`/`Max` known, `Severity` null): store raw `min`/`max` verbatim — do **not** coerce to a named severity even when the pair equals a canonical band.
    - **import** (`priorItem == nil`): default to named form when the API `{min,max}` matches a canonical band, else raw.
  - Form selection is driven exclusively by `prior` (threaded by `panelkit.SimpleFromAPI`); never by the API alone.
  - `time_range` null-preservation: keep `mode` null if it was null in prior state.

## 4. Registration

- [x] 4.1 Register `mlanomalycharts.SchemaAttribute()` in `internal/kibana/dashboard/schema.go` alongside the other `*_config` blocks.
- [x] 4.2 Add the handler to `panelHandlers` in `internal/kibana/dashboard/registry.go` (or equivalent handler registration map).
- [x] 4.3 Add `"ml_anomaly_charts"` to the reject-config_json panel type list in the panel routing logic (REQ-010 enforcement).

## 5. Tests

- [x] 5.1 Unit tests in `internal/kibana/dashboard/panel/mlanomalycharts/` (file: `api_test.go` or `model_test.go`):
  - Named severity round-trip: write `severity = "major"` → verify API payload `{min: 50, max: 75}`; read back and verify state shows `severity = "major"`.
  - `critical` round-trip (open-ended): write `severity = "critical"` → verify API payload `{min: 75}` (no `max`); read back → `severity = "critical"`.
  - Raw range round-trip: write `min = 10, max = 20` → verify API payload `{min: 10, max: 20}`; read back → `min = 10, max = 20` (not coerced to a named severity).
  - Form preservation (prior raw): prior item `{min: 3, max: 25}` (coincides with `warning` band); API returns `{min: 3, max: 25}` → state stores `{min: 3, max: 25}` (NOT `severity = "warning"`); plan stable.
  - Form preservation (prior named): prior item `severity = "warning"`; API returns `{min: 3, max: 25}` → state stores `severity = "warning"`; plan stable.
  - Form preservation (`critical` raw): prior item `{min: 75}`; API returns `{min: 75}` (no `max`) → state stores `{min: 75}` raw (NOT `severity = "critical"`); plan stable.
  - Import default (no prior): API returns `{min: 3, max: 25}` → state stores `severity = "warning"` (import-only normalization).
  - Mixed-list form preservation across refresh: prior `[{severity: "major"}, {min:10, max:20}]`, API returns `[{min:50,max:75}, {min:10,max:20}]` → first stays named, second stays raw.
  - Prior-named drift fallback: prior `severity = "major"`, API returns `{min:50, max:80}` → state falls back to raw `min:50, max:80`.
  - `config_json` rejection for `ml_anomaly_charts` panels.
  - Null-preservation: prior state omits `max_series_to_plot`; API returns a server value → state keeps null.
  - `time_range` null-preservation: prior state has `mode = null`; API omits `mode` → state keeps null.
  - Plan-time validator: item with both `severity` and `min` → error; item with `max` but no `min`/`severity` → error; item with `severity` and `max` → error; item with neither → error.
- [x] 5.2 Acceptance test in `internal/kibana/dashboard/panel/mlanomalycharts/acc_test.go`:
  - Test 1: create a dashboard with an `ml_anomaly_charts` panel using named severities (`critical`, `major`); verify plan is stable after apply.
  - Test 2: create a dashboard with a raw-range severity threshold `{min: 10, max = 20}`; verify plan is stable.
  - Test 3: create a dashboard with a raw-range severity threshold `{min: 3, max = 25}` (coincides with the `warning` canonical band); verify plan is stable after apply (locks in the form-preservation no-diff guarantee end-to-end).
  - Test 4: update `job_ids` in-place; verify update does not destroy and recreate the dashboard.
  - Test 5: optional-fields test with `title`, `description`, `hide_title`, `hide_border`, `max_series_to_plot`, and panel-level `time_range` (with `mode`); verify values and stable plan.

## 6. Delta spec and validation

- [x] 6.1 Ensure `openspec/changes/dashboard-ml-anomaly-charts/specs/kibana-dashboard/spec.md` defines REQ-053 with scenarios.
- [x] 6.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate dashboard-ml-anomaly-charts --type change`; fix any reported issues.
- [x] 6.3 Run `make build` and `go test ./internal/kibana/dashboard/...` (without `TF_ACC` for unit tests).
- [x] 6.4 Run `make check-openspec`.
