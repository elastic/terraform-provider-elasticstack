## 1. AIOps log rate analysis panel

- [x] 1.1 Create package `internal/kibana/dashboard/panel/aiopslograteanalysis/`
- [x] 1.2 Add `schema.go`: `SchemaAttribute()` using `panelkit.PanelPresentationAttributes()` plus `data_view_id` (required string); return `panelkit.PanelConfigBlock` with `BlockName: "aiops_log_rate_analysis_config"`, `PanelType: "aiops_log_rate_analysis"`
- [x] 1.3 Add `model.go`: `AiopsLogRateAnalysisConfigModel` with `DataViewID types.String` plus the presentation fields (`Title`, `Description`, `HideTitle`, `HideBorder`, `TimeRange`)
- [x] 1.4 Add `api.go`: `BuildConfig()` mapping model → `kbapi.KibanaHTTPAPIsAiopsLogRateAnalysis`; `PopulateFromAPI()` mapping API → model with REQ-009 null-preservation; `Handler{}` embedding `panelkit.NoopHandlerBase`, implementing `PanelType()`, `SchemaAttribute()`, `FromAPI()` (via `panelkit.SimpleFromAPI`), `ToAPI()` (via `panelkit.SimpleToAPI` + `panelkit.RejectConfigJSON`), `ValidatePanelConfig()` (validate `data_view_id` required)
- [x] 1.5 Extend `PanelModel` in `internal/kibana/dashboard/models/panel.go` with `AiopsLogRateAnalysisConfig *AiopsLogRateAnalysisConfigModel \`tfsdk:"aiops_log_rate_analysis_config"\``
- [x] 1.6 Register `aiopslograteanalysis.Handler{}` in `panelHandlers` slice in `internal/kibana/dashboard/registry.go`

## 2. AIOps pattern analysis panel

- [x] 2.1 Create package `internal/kibana/dashboard/panel/aiopspatternanalysis/`
- [x] 2.2 Add `schema.go`: `SchemaAttribute()` using `panelkit.PanelPresentationAttributes()` plus:
  - `data_view_id` — required string, "The data view ID used for pattern analysis."
  - `field_name` — required string, "The text field on which to run pattern analysis."
  - `minimum_time_range` — optional string, `stringvalidator.OneOf("no_minimum", "1_week", "1_month", "3_months", "6_months")`
  - `random_sampler_mode` — optional string, `stringvalidator.OneOf("off", "on_automatic", "on_manual")`
  - `random_sampler_probability` — optional float64, `float64validator.Between(0.00001, 0.5)`, "Sampling probability, only meaningful when `random_sampler_mode = on_manual`."
  Return `panelkit.PanelConfigBlock` with `BlockName: "aiops_pattern_analysis_config"`, `PanelType: "aiops_pattern_analysis"`
- [x] 2.3 Add `model.go`: `AiopsPatternAnalysisConfigModel` with `DataViewID`, `FieldName types.String`; `MinimumTimeRange`, `RandomSamplerMode types.String`; `RandomSamplerProbability types.Float64`; plus presentation fields
- [x] 2.4 Add `api.go`: `BuildConfig()`, `PopulateFromAPI()`, `Handler{}` following same pattern as task 1.4; `ValidatePanelConfig()` validates `data_view_id` and `field_name` required
- [x] 2.5 Extend `PanelModel` with `AiopsPatternAnalysisConfig *AiopsPatternAnalysisConfigModel \`tfsdk:"aiops_pattern_analysis_config"\``
- [x] 2.6 Register `aiopspatternanalysis.Handler{}` in `panelHandlers`

## 3. AIOps change point chart panel

- [x] 3.1 Create package `internal/kibana/dashboard/panel/aiopschangepointchart/`
- [x] 3.2 Add `schema.go`: `SchemaAttribute()` using `panelkit.PanelPresentationAttributes()` plus:
  - `data_view_id` — required string
  - `metric_field` — required string, "The metric field used by the aggregation function."
  - `aggregation_function` — optional string, `stringvalidator.OneOf("avg", "max", "min", "sum")`
  - `split_field` — optional string, "Field used to split change-point results."
  - `partitions` — optional `schema.SetAttribute` with `ElementType: types.StringType`, "Optional split field values to include (order-insensitive filter set)."
  - `max_series_to_plot` — optional float64, "Maximum number of change points to visualise. Kibana default is 6."
  - `view_type` — optional string, `stringvalidator.OneOf("charts", "table")`
  Return `panelkit.PanelConfigBlock` with `BlockName: "aiops_change_point_chart_config"`, `PanelType: "aiops_change_point_chart"`
- [x] 3.3 Add `model.go`: `AiopsChangePointChartConfigModel` with `DataViewID`, `MetricField`, `AggregationFunction`, `SplitField`, `ViewType types.String`; `Partitions types.Set`; `MaxSeriesToPlot types.Float64`; plus presentation fields
- [x] 3.4 Add `api.go`: `BuildConfig()`, `PopulateFromAPI()`, `Handler{}` following the pattern; map `Partitions` set ↔ `*[]string` in the API; null-preserve `MaxSeriesToPlot`, `Partitions`, and all optional fields; `ValidatePanelConfig()` validates `data_view_id` and `metric_field` required
- [x] 3.5 Extend `PanelModel` with `AiopsChangePointChartConfig *AiopsChangePointChartConfigModel \`tfsdk:"aiops_change_point_chart_config"\``
- [x] 3.6 Register `aiopschangepointchart.Handler{}` in `panelHandlers`

## 4. Tests

- [x] 4.1 Unit test (`api_test.go`) per panel: `BuildConfig` round-trip for required-only and all-optional configurations; `PopulateFromAPI` null-preservation (prior state null → stay null after read)
- [x] 4.2 Acceptance test per panel (`acc_test.go`) covering:
  - Create with required fields only; verify type, grid, and config attrs
  - Import and verify `ImportStateVerify` passes
  - Re-apply with no changes (plan-only); verify no drift
  - Create with all optional fields set; verify each attr
- [x] 4.3 Shared multi-panel acceptance test: a single dashboard containing all three AIOps panels; verify sibling-block mutual exclusion (other `*_config` blocks stay null)
- [x] 4.4 Plan-time error tests:
  - `random_sampler_probability` out of range (e.g., `1.0`) → expect error matching `Between`
  - `aggregation_function` invalid value → expect `OneOf` error
  - `minimum_time_range` invalid value → expect `OneOf` error
  - `view_type` invalid value → expect `OneOf` error

## 5. Build and spec sync

- [x] 5.1 Run `make build` and confirm it compiles
- [x] 5.2 Run `go vet ./internal/kibana/dashboard/...` and fix any issues
- [x] 5.3 Run `go test ./internal/kibana/dashboard/...` (unit tests only, no `TF_ACC=1`)
- [x] 5.4 Run `make check-openspec`
