## 1. Create dashboard/models package

- [x] 1.1 Create `internal/kibana/dashboard/models/` directory
- [x] 1.2 Create `models/panel.go` with `PanelModel`, `PanelGridModel`, `PinnedPanelModel`, `SectionModel`, `SectionGridModel`
- [x] 1.3 Create `models/dashboard.go` with `DashboardModel`, `DashboardQueryModel`, `RefreshIntervalModel`, `OptionsModel`, `AccessControlValue`, `TimeRangeModel`
- [x] 1.4 Create `models/slo_burn_rate.go` with `SloBurnRateConfigModel`, `SloBurnRateDrilldownModel`
- [x] 1.5 Create `models/slo_overview.go` with `SloOverviewConfigModel`, `SloOverviewSingleModel`, `SloOverviewGroupsModel`
- [x] 1.6 Create `models/slo_error_budget.go` with `SloErrorBudgetConfigModel`
- [x] 1.7 Create `models/markdown.go` with `MarkdownConfigModel`, `MarkdownConfigByValueModel`, `MarkdownConfigByReferenceModel`, `MarkdownConfigSettingsModel`
- [x] 1.8 Create `models/controls.go` with `TimeSliderControlConfigModel`, `OptionsListControlConfigModel`, `RangeSliderControlConfigModel`, `EsqlControlConfigModel`
- [x] 1.9 Create `models/synthetics.go` with `SyntheticsStatsOverviewConfigModel`, `SyntheticsMonitorsConfigModel`
- [x] 1.10 Create `models/image.go` with `ImagePanelConfigModel`, `ImagePanelSrcModel`, `ImagePanelSrcFileModel`, `ImagePanelSrcURLModel`, `ImagePanelDrilldownModel`, `ImagePanelDashboardDrilldownModel`, `ImagePanelURLDrilldownModel`
- [x] 1.11 Create `models/slo_alerts.go` with `SloAlertsPanelConfigModel`, `SloAlertsPanelSloModel`, `SloAlertsPanelDrilldownModel`
- [x] 1.12 Create `models/discover_session.go` with `DiscoverSessionPanelConfigModel`, `DiscoverSessionPanelByValueModel`, `DiscoverSessionTabModel`, `DiscoverSessionDSLTabModel`, `DiscoverSessionESQLTabModel`, `DiscoverSessionPanelByRefModel`, `DiscoverSessionOverridesModel`, `DiscoverSessionSortModel`, `DiscoverSessionColumnSettingModel`, `DiscoverSessionPanelDrilldown`
- [x] 1.13 Create `models/lens.go` with `LensByValueChartBlocks`, `LensDashboardAppConfigModel`, `VisConfigModel`, `VisByValueModel`, `VisByReferenceModel`, `LensChartPresentationTFModel`
- [x] 1.14 Create `models/lens_xy.go` with `XYChartConfigModel`, `XYLayerModel`, `XYAxisModel`, etc.
- [x] 1.15 Create `models/lens_gauge.go` with `GaugeConfigModel`, `GaugeStylingModel`
- [x] 1.16 Create `models/lens_metric.go` with `MetricChartConfigModel`, `MetricItemModel`
- [x] 1.17 Create `models/lens_pie.go` with `PieChartConfigModel`
- [x] 1.18 Create `models/lens_treemap.go` with `TreemapConfigModel`
- [x] 1.19 Create `models/lens_mosaic.go` with `MosaicConfigModel`
- [x] 1.20 Create `models/lens_datatable.go` with `DatatableConfigModel`, `DatatableNoESQLConfigModel`, `DatatableESQLConfigModel`
- [x] 1.21 Create `models/lens_tagcloud.go` with `TagcloudConfigModel`
- [x] 1.22 Create `models/lens_heatmap.go` with `HeatmapConfigModel`
- [x] 1.23 Create `models/lens_region_map.go` with `RegionMapConfigModel`
- [x] 1.24 Create `models/lens_legacy_metric.go` with `LegacyMetricConfigModel`
- [x] 1.25 Create `models/lens_waffle.go` with `WaffleConfigModel`
- [x] 1.26 Create `models/filters.go` with `ChartFilterJSONModel`, `FilterSimpleModel`
- [x] 1.27 Ensure `models/` package imports only TPF types, `customtypes`, and `jsontypes` (no `kbapi`, no `dashboard`)

## 2. Update dashboard package references

- [x] 2.1 Update `models_panels.go` to import `models/` and use `models.PanelModel` instead of `panelModel`
- [x] 2.2 Update `models.go` to use `models.DashboardModel` instead of `dashboardModel`
- [x] 2.3 Update `create.go`/`read.go`/`update.go`/`delete.go`/`resource.go` to use `models.DashboardModel`
- [x] 2.4 Update all `models_*.go` conversion files to import `models/` and use exported type names
- [x] 2.5 Update `schema.go` to reference `models.TimeRangeModel` where needed (no change required — file does not reference `tfsdk` model types directly)
- [x] 2.6 Update `panel_config_validator.go` to use `models.PanelModel` (no change required — validators operate on schema/plan paths without concrete model types)
- [x] 2.7 Update `panel_config_defaults.go` to reference model types (no change required — defaults are built from schema without importing `dashboard/models`)
- [x] 2.8 Update `models_plan_state_alignment.go` to use exported model types
- [x] 2.9 Update `pinned_panels_mapping.go` to use `models.PinnedPanelModel`
- [x] 2.10 Remove struct definitions from all original `dashboard/models_*.go` files (keep conversion functions)

## 3. Verify

- [x] 3.1 `go build ./internal/kibana/dashboard/...` passes
- [x] 3.2 `go vet ./internal/kibana/dashboard/...` passes
- [x] 3.3 `go test ./internal/kibana/dashboard/...` passes (all unit tests)
- [x] 3.4 `make build` passes
- [x] 3.5 No user-visible schema changes confirmed by diffing generated docs
