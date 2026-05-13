## 1. Create lenscommon package

- [ ] 1.1 Create `dashboard/lenscommon/iface.go` with `VizConverter` and `Resolver` interfaces
- [ ] 1.2 Create `dashboard/lenscommon/registry.go` with global converter map, `Register`, `ForType`, `FirstForBlocks`, `All`
- [ ] 1.3 Create `dashboard/lenscommon/presentation.go` moving `lensChartPresentationReadsFor` and `lensChartPresentationWritesFor` from `dashboard/`
- [ ] 1.4 Create `dashboard/lenscommon/drilldowns.go` moving shared drilldown conversion helpers
- [ ] 1.5 Create `dashboard/lenscommon/by_reference.go` moving shared by_reference read/write logic
- [ ] 1.6 Create `dashboard/lenscommon/slice_alignment.go` with `RegisterSliceAligner` and `ApplySliceAligners`

## 2. Migrate LensByValueChartBlocks to models

- [ ] 2.1 Move `lensByValueChartBlocks` struct from `dashboard/models_lens_by_value_chart_blocks.go` to `models/lens.go` as exported `LensByValueChartBlocks`
- [ ] 2.2 Update all references across dashboard package to use `models.LensByValueChartBlocks`
- [ ] 2.3 Move `lensByValueChartBlocksFromPanel` and related helpers to `panelkit/` or delete if superseded by reflection
- [ ] 2.4 Delete `dashboard/models_lens_by_value_chart_blocks.go`

## 3. Migrate converters (12 total)

For each converter, create `dashboard/panel/lens{kind}/converter.go` with self-registration in `init()`:

- [ ] 3.1 `lensxy/` — XY chart converter (`VizType: "xy"`)
- [ ] 3.2 `lensgauge/` — Gauge converter (`VizType: "gauge"`)
- [ ] 3.3 `lensmetric/` — Metric chart converter (`VizType: "metric"`)
- [ ] 3.4 `lenslegacymetric/` — Legacy metric converter (`VizType: "legacy_metric"`)
- [ ] 3.5 `lenspie/` — Pie chart converter (`VizType: "pie"`)
- [ ] 3.6 `lenstreemap/` — Treemap converter (`VizType: "treemap"`)
- [ ] 3.7 `lensmosaic/` — Mosaic converter (`VizType: "mosaic"`)
- [ ] 3.8 `lensdatatable/` — Datatable converter (`VizType: "datatable"`)
- [ ] 3.9 `lenstagcloud/` — Tagcloud converter (`VizType: "tagcloud"`)
- [ ] 3.10 `lensheatmap/` — Heatmap converter (`VizType: "heatmap"`)
- [ ] 3.11 `lensregionmap/` — Region map converter (`VizType: "region_map"`)
- [ ] 3.12 `lenswaffle/` — Waffle converter (`VizType: "waffle"`)

Each converter must:
- Implement `lenscommon.VizConverter`
- Call `lenscommon.Register(converter{})` in `init()`
- Include `converter_test.go` with existing unit test coverage

## 4. Delete old lens infrastructure

- [ ] 4.1 Delete `dashboard/models_lens_panel.go` (replaced by `lenscommon/`)
- [ ] 4.2 Delete `dashboard/models_lens_dashboard_app_converters.go` (by_reference moves to `lenscommon/`)
- [ ] 4.3 Delete `dashboard/models_lens_dashboard_app_by_value_adapter.go` (absorbed into composites later)
- [ ] 4.4 Delete `dashboard/models_lens_dashboard_app_panel.go` (absorbed into composites later)
- [ ] 4.5 Delete `dashboard/models_lens_by_value_chart_blocks.go` (struct moved to `models/`)
- [ ] 4.6 Update or delete `dashboard/models_lens_panel.go` tests (moved to converter packages)

## 5. Refactor state alignment and defaults

- [ ] 5.1 Refactor `alignPanelStateFromPlan` in `models_plan_state_alignment.go` to delegate lens chart alignment to `lenscommon.All()` converters
- [ ] 5.2 Register XY chart slice aligner in `lensxy/converter.go` `init()`
- [ ] 5.3 Remove explicit per-chart alignment calls from `alignPanelStateFromPlan`
- [ ] 5.4 Refactor `populateLensAttributesDefaults` in `panel_config_defaults.go` to dispatch to `lenscommon.ForType(vizType).PopulateJSONDefaults()`
- [ ] 5.5 Remove hard-coded lens chart type switches from `panel_config_defaults.go`

## 6. Verification

- [ ] 6.1 `go build ./internal/kibana/dashboard/...` passes
- [ ] 6.2 `go vet ./...` passes
- [ ] 6.3 `go test ./internal/kibana/dashboard/...` passes
- [ ] 6.4 All Lens chart acceptance tests pass (XY, gauge, metric, pie, treemap, mosaic, datatable, tagcloud, heatmap, region_map, legacy_metric, waffle)
- [ ] 6.5 `make build` passes
