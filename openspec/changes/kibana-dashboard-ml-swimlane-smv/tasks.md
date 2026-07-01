## 1. Spec

- [x] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-dashboard-ml-swimlane-smv --type change` (or `make check-openspec` after sync).
- [x] 1.2 Resolve open question on minimum Kibana version for `ml_anomaly_swimlane` and `ml_single_metric_viewer` panels (see `design.md`); update delta spec with a version compatibility note if confirmed.
- [ ] 1.3 On completion of implementation, **sync** delta into `openspec/specs/kibana-dashboard/spec.md` or **archive** the change per project workflow.

## 2. Implementation — `ml_anomaly_swimlane`

- [ ] 2.1 Add `MlAnomalySwimlaneConfigModel` to `internal/kibana/dashboard/models/` (file `ml_anomaly_swimlane.go`): `SwimlaneType types.String`, `JobIds []types.String`, `ViewBy types.String`, `PerPage types.Float32`, `Title types.String`, `Description types.String`, `HideTitle types.Bool`, `HideBorder types.Bool`, `TimeRange *TimeRangeModel`.
- [ ] 2.2 Add `MlAnomalySwimlaneConfig *MlAnomalySwimlaneConfigModel` field to `PanelModel` in `internal/kibana/dashboard/models/panel.go` with tfsdk tag `ml_anomaly_swimlane_config`.
- [ ] 2.3 Create `internal/kibana/dashboard/panel/mlanomalyswimlane/schema.go`: `SchemaAttribute()` using `panelkit.PanelConfigBlock` with `PanelPresentationAttributes()` plus `swimlane_type` (required string, `stringvalidator.OneOf("overall", "viewBy")`), `job_ids` (required list of strings, `listvalidator.SizeAtLeast(1)`), `view_by` (optional string), `per_page` (optional float32 — use `schema.Float32Attribute` if supported, else `schema.Float64Attribute` with API coercion note), `time_range` via `panelkit.TimeRangeSchema()`. Add cross-field object validators: `view_by` required when `swimlane_type = "viewBy"`; `view_by` forbidden when `swimlane_type = "overall"`.
- [ ] 2.4 Create `internal/kibana/dashboard/panel/mlanomalyswimlane/api.go`: implement `Handler` embedding `panelkit.NoopHandlerBase`, `PanelType()`, `SchemaAttribute()`, `FromAPI()` (using `panelkit.SimpleFromAPI` + `item.AsKibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalySwimlane`), `ToAPI()` (using `panelkit.SimpleToAPI`, dispatching to branch 0 or 1 based on `swimlane_type`), `ValidatePanelConfig()`.
- [ ] 2.5 Create `internal/kibana/dashboard/panel/mlanomalyswimlane/model.go`: `BuildConfig()` (state → API payload, dispatches to `MlAnomalySwimlane0` or `MlAnomalySwimlane1`), `PopulateFromAPI()` (API → state, union branch detection via `AsKibanaHTTPAPIsMlAnomalySwimlane1` check on `swimlane_type`), null-preserve all optional fields via `panelkit.Preserve*` helpers.
- [ ] 2.6 Register `mlanomalyswimlane.Handler{}` in `panelHandlers` slice in `internal/kibana/dashboard/registry.go`.
- [ ] 2.7 Add `ml_anomaly_swimlane_config` schema attribute to `internal/kibana/dashboard/schema.go` panel object.

## 3. Implementation — `ml_single_metric_viewer`

- [ ] 3.1 Add `MlSingleMetricViewerConfigModel` to `internal/kibana/dashboard/models/` (file `ml_single_metric_viewer.go`): `JobIds []types.String`, `SelectedDetectorIndex types.Float32`, `ForecastId types.String`, `FunctionDescription types.String`, `SelectedEntities map[string]MlSingleMetricViewerEntityModel` (or `types.Map` with nested attribute type), `Title types.String`, `Description types.String`, `HideTitle types.Bool`, `HideBorder types.Bool`, `TimeRange *TimeRangeModel`. Add `MlSingleMetricViewerEntityModel` with `StringValue types.String` and `NumericValue types.Number`.
- [ ] 3.2 Add `MlSingleMetricViewerConfig *MlSingleMetricViewerConfigModel` field to `PanelModel` in `internal/kibana/dashboard/models/panel.go` with tfsdk tag `ml_single_metric_viewer_config`.
- [ ] 3.3 Create `internal/kibana/dashboard/panel/mlsinglemetricviewer/schema.go`: `SchemaAttribute()` using `panelkit.PanelConfigBlock` with `PanelPresentationAttributes()` plus `job_ids` (required list of strings, `listvalidator.SizeAtLeast(1)`, `listvalidator.SizeAtMost(1)`), `selected_detector_index` (optional float32/float64), `forecast_id` (optional string), `function_description` (optional string, `stringvalidator.OneOf("min", "max", "mean")`), `selected_entities` (`schema.MapNestedAttribute`, optional, nested object with optional `string_value types.String` and optional `numeric_value types.Number`, plan-time object validator enforcing exactly one of the two per entry), `time_range` via `panelkit.TimeRangeSchema()`.
- [ ] 3.4 Create `internal/kibana/dashboard/panel/mlsinglemetricviewer/api.go`: implement `Handler`, `FromAPI()`, `ToAPI()`, `ValidatePanelConfig()` analogous to the swimlane handler.
- [ ] 3.5 Create `internal/kibana/dashboard/panel/mlsinglemetricviewer/model.go`: `BuildConfig()` (state → API including `selected_entities` serialization: string_value → `string` union branch, numeric_value → `float32` union branch), `PopulateFromAPI()` (API → state including `selected_entities` deserialization via `AsKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities0` / `...1`, null-preserve all optional fields).
- [ ] 3.6 Register `mlsinglemetricviewer.Handler{}` in `panelHandlers` in `registry.go`.
- [ ] 3.7 Add `ml_single_metric_viewer_config` schema attribute to `internal/kibana/dashboard/schema.go` panel object.

## 4. Testing — `ml_anomaly_swimlane`

- [ ] 4.1 Add acceptance test (`internal/kibana/dashboard/panel/mlanomalyswimlane/acc_test.go`): create/read/update round-trip with `swimlane_type = "overall"`, verify state matches.
- [ ] 4.2 Add acceptance test: create/read/update round-trip with `swimlane_type = "viewBy"` and `view_by` set, verify state matches.
- [ ] 4.3 Add acceptance test: plan-time rejection when `swimlane_type = "viewBy"` and `view_by` is absent (expect diagnostic).
- [ ] 4.4 Add acceptance test: plan-time rejection when `swimlane_type = "overall"` and `view_by` is set (expect diagnostic).
- [ ] 4.5 Add unit tests for `BuildConfig()` and `PopulateFromAPI()` covering both branches, null-preservation on optional fields, and `per_page` float32 round-trip.

## 5. Testing — `ml_single_metric_viewer`

- [ ] 5.1 Add acceptance test: create/read/update round-trip with `selected_entities` containing both string and numeric entries, verify state matches.
- [ ] 5.2 Add acceptance test: plan-time rejection when a `selected_entities` entry has both `string_value` and `numeric_value` set.
- [ ] 5.3 Add acceptance test: plan-time rejection when a `selected_entities` entry has neither `string_value` nor `numeric_value` set.
- [ ] 5.4 Add acceptance test: plan-time rejection when `job_ids` has more than one entry (length-1 validator).
- [ ] 5.5 Add acceptance test: round-trip with `forecast_id` and `function_description` set.
- [ ] 5.6 Add unit tests for `BuildConfig()` and `PopulateFromAPI()` covering `selected_entities` serialization/deserialization (string branch, numeric branch, null-preservation when key absent in prior state).
