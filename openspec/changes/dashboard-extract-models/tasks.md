## 1. Create dashboard/models package

- [ ] 1.1 Create `internal/kibana/dashboard/models/` directory
- [ ] 1.2 Create `models/panel.go` with `PanelModel`, `PanelGridModel`, `PinnedPanelModel`, `SectionModel`, `SectionGridModel`
- [ ] 1.3 Create `models/dashboard.go` with `DashboardModel`, `DashboardQueryModel`, `RefreshIntervalModel`, `OptionsModel`, `AccessControlValue`, `TimeRangeModel`
- [ ] 1.4 Create `models/slo_burn_rate.go` with `SloBurnRateConfigModel`, `SloBurnRateDrilldownModel`
- [ ] 1.5 Create `models/slo_overview.go` with `SloOverviewConfigModel`, `SloOverviewSingleModel`, `SloOverviewGroupsModel`
- [ ] 1.6 Create `models/slo_error_budget.go` with `SloErrorBudgetConfigModel`
- [ ] 1.7 Create `models/markdown.go` with `MarkdownConfigModel`, `MarkdownConfigByValueModel`, `MarkdownConfigByReferenceModel`, `MarkdownConfigSettingsModel`
- [ ] 1.8 Create `models/controls.go` with `TimeSliderControlConfigModel`, `OptionsListControlConfigModel`, `RangeSliderControlConfigModel`, `EsqlControlConfigModel`
- [ ] 1.9 Create `models/synthetics.go` with `SyntheticsStatsOverviewConfigModel`, `SyntheticsMonitorsConfigModel`
- [ ] 1.10 Create `models/image.go` with `ImagePanelConfigModel`, `ImagePanelSrcModel`, `ImagePanelSrcFileModel`, `ImagePanelSrcURLModel`, `ImagePanelDrilldownModel`, `ImagePanelDashboardDrilldownModel`, `ImagePanelURLDrilldownModel`
- [ ] 1.11 Create `models/slo_alerts.go` with `SloAlertsPanelConfigModel`, `SloAlertsPanelSloModel`, `SloAlertsPanelDrilldownModel`
- [ ] 1.12 Create `models/discover_session.go` with `DiscoverSessionPanelConfigModel`, `DiscoverSessionPanelByValueModel`, `DiscoverSessionTabModel`, `DiscoverSessionDSLTabModel`, `DiscoverSessionESQLTabModel`, `DiscoverSessionPanelByRefModel`, `DiscoverSessionOverridesModel`, `DiscoverSessionSortModel`, `DiscoverSessionColumnSettingModel`, `DiscoverSessionPanelDrilldown`
- [ ] 1.13 Create `models/lens.go` with `LensByValueChartBlocks`, `LensDashboardAppConfigModel`, `VizConfigModel`, `VizByValueModel`, `VizByReferenceModel`, `LensChartPresentationTFModel`
- [ ] 1.14 Create `models/lens_xy.go` with `XYChartConfigModel`, `XYLayerModel`, `XYAxisModel`, etc.
- [ ] 1.15 Create `models/lens_gauge.go` with `GaugeConfigModel`, `GaugeStylingModel`
- [ ] 1.16 Create `models/lens_metric.go` with `MetricChartConfigModel`, `MetricChartMetricModel`
- [ ] 1.17 Create `models/lens_pie.go` with `PieChartConfigModel`
- [ ] 1.18 Create `models/lens_treemap.go` with `TreemapConfigModel`
- [ ] 1.19 Create `models/lens_mosaic.go` with `MosaicConfigModel`
- [ ] 1.20 Create `models/lens_datatable.go` with `DatatableConfigModel`, `DatatableNoESQLConfigModel`, `DatatableESQLConfigModel`
- [ ] 1.21 Create `models/lens_tagcloud.go` with `TagcloudConfigModel`
- [ ] 1.22 Create `models/lens_heatmap.go` with `HeatmapConfigModel`
- [ ] 1.23 Create `models/lens_region_map.go` with `RegionMapConfigModel`
- [ ] 1.24 Create `models/lens_legacy_metric.go` with `LegacyMetricConfigModel`
- [ ] 1.25 Create `models/lens_waffle.go` with `WaffleConfigModel`
- [ ] 1.26 Create `models/filters.go` with `ChartFilterJSONModel`, `FilterSimpleModel`
- [ ] 1.27 Ensure `models/` package imports only TPF types, `customtypes`, and `jsontypes` (no `kbapi`, no `dashboard`)

## 2. Update dashboard package references

- [ ] 2.1 Update `models_panels.go` to import `models/` and use `models.PanelModel` instead of `panelModel`
- [ ] 2.2 Update `models.go` to use `models.DashboardModel` instead of `dashboardModel`
- [ ] 2.3 Update `create.go`/`read.go`/`update.go`/`delete.go`/`resource.go` to use `models.DashboardModel`
- [ ] 2.4 Update all `models_*.go` conversion files to import `models/` and use exported type names
- [ ] 2.5 Update `schema.go` to reference `models.TimeRangeModel` where needed
- [ ] 2.6 Update `panel_config_validator.go` to use `models.PanelModel`
- [ ] 2.7 Update `panel_config_defaults.go` to reference model types
- [ ] 2.8 Update `models_plan_state_alignment.go` to use exported model types
- [ ] 2.9 Update `pinned_panels_mapping.go` to use `models.PinnedPanelModel`
- [ ] 2.10 Remove struct definitions from all original `dashboard/models_*.go` files (keep conversion functions)

## 3. Verify

- [ ] 3.1 `go build ./internal/kibana/dashboard/...` passes
- [ ] 3.2 `go vet ./internal/kibana/dashboard/...` passes
- [ ] 3.3 `go test ./internal/kibana/dashboard/...` passes (all unit tests)
- [ ] 3.4 `make build` passes
- [ ] 3.5 No user-visible schema changes confirmed by diffing generated docs
