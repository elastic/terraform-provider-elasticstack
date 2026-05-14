## Context

The `internal/kibana/dashboard` package contains ~44,000 lines. Terraform model structs (`panelModel`, `dashboardModel`, `sloBurnRateConfigModel`, `xyChartConfigModel`, etc.) are scattered across `models_*.go` files alongside API conversion logic, schema builders, validators, and state alignment. Because these structs carry `tfsdk:` tags, Terraform Plugin Framework populates them via reflection. They must remain in a single Go package, but that package need not be `dashboard`.

Extracting them into `dashboard/models` is a prerequisite for strict package separation of panel logic. The new package sits at the bottom of the import DAG, importing nothing from `dashboard/`, `panel/`, `lenscommon/`, or `kbapi` conversion code.

## Goals / Non-Goals

**Goals:**
- All `tfsdk:`-tagged structs live in `dashboard/models`
- `dashboard/models` imports nothing from `dashboard/` (clean DAG)
- All existing tests pass with zero behavioral changes

**Non-Goals:**
- No interface extraction yet (that happens in `dashboard-panel-contract`)
- No file reorganization within `dashboard/` beyond reference updates
- No logic changes of any kind

## Decisions

### Naming convention

All model structs become exported. The current unexported names (`panelModel`, `dashboardModel`) gain a capital letter. Config model names keep their current shape but are exported:

| Before | After |
|--------|-------|
| `panelModel` | `models.PanelModel` |
| `dashboardModel` | `models.DashboardModel` |
| `sloBurnRateConfigModel` | `models.SloBurnRateConfigModel` |
| `xyChartConfigModel` | `models.XYChartConfigModel` |
| `lensByValueChartBlocks` | `models.LensByValueChartBlocks` |

The shared `timeRangeModel` stays as `models.TimeRangeModel` because other packages will need it.

### What stays in `dashboard/`

- Resource lifecycle (`create.go`, `read.go`, `update.go`, `delete.go`, `resource.go`)
- Schema assembly (`schema.go` — references `models.TimeRangeModel` but doesn't define it)
- Validation logic (`panel_config_validator.go`, `drilldown_validators.go`, etc.)
- State alignment (`models_plan_state_alignment.go`)
- Config JSON defaulting (`panel_config_defaults.go`)
- API conversion logic (to be refactored in subsequent changes)

### What moves to `dashboard/models/`

- `panel.go` — `PanelModel`, `PanelGridModel`, `PinnedPanelModel`, `SectionModel`, `SectionGridModel`
- `dashboard.go` — `DashboardModel`, `DashboardQueryModel`, `RefreshIntervalModel`, `OptionsModel`, `AccessControlValue`, etc.
- `slo_burn_rate.go` — `SloBurnRateConfigModel`, `SloBurnRateDrilldownModel`
- `markdown.go` — `MarkdownConfigModel`, `MarkdownConfigByValueModel`, etc.
- `lens.go` — `LensByValueChartBlocks`, `LensDashboardAppConfigModel`, `VizConfigModel`, etc.
- `lens_xy.go` — `XYChartConfigModel`, `XYLayerModel`, etc.
- `lens_gauge.go` — `GaugeConfigModel`, etc.
- And all remaining `*ConfigModel` structs

### File organization in `dashboard/models/`

One file per conceptual grouping. The exact file names are not critical because the package is data-only; contributors can add new config models to whichever file is most natural. Suggested initial structure:

```
dashboard/models/
  panel.go              — PanelModel, PanelGridModel, PinnedPanelModel, SectionModel
  dashboard.go          — DashboardModel, DashboardQueryModel, RefreshIntervalModel, OptionsModel
  time_range.go         — TimeRangeModel
  slo_burn_rate.go      — SloBurnRateConfigModel
  slo_overview.go       — SloOverviewConfigModel
  slo_error_budget.go   — SloErrorBudgetConfigModel
  markdown.go           — MarkdownConfigModel
  controls.go           — TimeSliderControlConfigModel, OptionsListControlConfigModel, etc.
  synthetics.go         — SyntheticsStatsOverviewConfigModel, SyntheticsMonitorsConfigModel
  image.go              — ImagePanelConfigModel, ImagePanelSrcModel, ImagePanelSrcFileModel, ImagePanelSrcURLModel, ImagePanelDrilldownModel, ImagePanelDashboardDrilldownModel, ImagePanelURLDrilldownModel
  slo_alerts.go         — SloAlertsPanelConfigModel, SloAlertsPanelSloModel, SloAlertsPanelDrilldownModel
  discover_session.go   — DiscoverSessionPanelConfigModel, DiscoverSessionPanelByValueModel, DiscoverSessionTabModel, DiscoverSessionDSLTabModel, DiscoverSessionESQLTabModel, DiscoverSessionPanelByRefModel, DiscoverSessionOverridesModel, DiscoverSessionSortModel, DiscoverSessionColumnSettingModel, DiscoverSessionPanelDrilldown
  lens.go               — LensByValueChartBlocks, LensDashboardAppConfigModel, VizConfigModel, VizByValueModel, VizByReferenceModel
  lens_xy.go            — XYChartConfigModel, XYLayerModel, …
  lens_gauge.go         — GaugeConfigModel, GaugeStylingModel
  lens_metric.go        — MetricChartConfigModel, …
  lens_pie.go           — PieChartConfigModel, …
  lens_treemap.go       — TreemapConfigModel, …
  lens_mosaic.go        — MosaicConfigModel, …
  lens_datatable.go     — DatatableConfigModel, …
  lens_tagcloud.go      — TagcloudConfigModel, …
  lens_heatmap.go       — HeatmapConfigModel, …
  lens_region_map.go    — RegionMapConfigModel, …
  lens_legacy_metric.go — LegacyMetricConfigModel, …
  lens_waffle.go        — WaffleConfigModel, …
  filters.go            — ChartFilterJSONModel, FilterSimpleModel
```

### Import DAG after extraction

```
dashboard/models                ← bottom; imports utils, customtypes, tpf types only
    ↑
dashboard                       ← imports models, kbapi, clients, panelkit (future)
    ↑
provider
```

All model files import only:
- `github.com/hashicorp/terraform-plugin-framework/types`
- `github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes`
- `github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes`

No `kbapi`, no `diag`, no schema imports.

Decoupling the data layer from the logic layer also enables shared schema factories in `panelkit/` and `lenscommon/` in downstream changes. Because `panelkit` can import `models` for reflection without creating a cycle, it can provide `HasConfig(pm, blockName)` and shared `schema.Attribute` factories that panel packages compose.

### Reference update strategy

Every `dashboardModel` in `dashboard/` becomes `models.DashboardModel`. Every `panelModel` becomes `models.PanelModel`. Because this is purely mechanical:

1. Create `dashboard/models/` with all structs
2. Delete struct definitions from existing `dashboard/models_*.go` files (leave conversion functions)
3. Add `import ".../dashboard/models"` to affected files
4. Run `go build ./internal/kibana/dashboard/...` and fix every compiler error

The compiler is the source of truth. No manual search required.

### `models_panels.go` and `models.go` during transition

`models_panels.go` currently contains both `panelModel` struct and `mapPanelFromAPI` / `panelsToAPI` methods. After extraction:
- The `panelModel` struct moves to `models/panel.go`
- The methods on `panelModel` (`toAPI`, `mapPanelFromAPI`, etc.) stay in `dashboard/models_panels.go` but change receiver type from `panelModel` to `models.PanelModel`
- Similarly, `dashboardModel.populateFromAPI` and `.toAPICreateRequest` stay in `dashboard/models.go` with updated types

These methods will be refactored into the handler registry in `dashboard-panel-contract`. This change only moves the data.

## Risks / Trade-offs

- [Risk] `go build` will produce hundreds of errors after struct removal; the fix is mechanical but tedious ➝ *Mitigation:* use `gopls` or IDE rename refactoring if available; otherwise batch-fix compiler errors
- [Risk] Missed references in test files (`_test.go`) that use unexported types ➝ *Mitigation:* `go test ./...` catches all; no unexported types remain in `dashboard/` that tests need to access from `models`
- [Risk] Accidental logic creep (developer starts moving conversion functions too) ➝ *Mitigation:* restrict scope in code review; only structs with `tfsdk` tags move

## Open questions

None.
