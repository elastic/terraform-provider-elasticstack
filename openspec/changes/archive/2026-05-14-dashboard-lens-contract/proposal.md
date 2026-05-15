## Why

The dashboard package contains 12 Lens visualization converters (`models_lens_panel.go`, `models_*_panel.go` for each chart kind) that convert between Terraform model blocks and `kbapi.KbnDashboardPanelTypeVisConfig0`. These are currently declared in a hard-coded slice `lensVizConverters` and dispatched via a long type-switch in `detectLensVizType`. Adding a new Lens chart requires touching multiple files.

This change extracts a `lenscommon` contract (`VizConverter` interface) and moves each converter into its own isolated package with self-registration. The registry becomes dynamic. Lens-specific shared infrastructure (presentation helpers, drilldown conversion, by_reference read/write) is centralized in `lenscommon`.

## What Changes

### New infrastructure

- `dashboard/lenscommon/iface.go` — `VizConverter` interface, `Resolver` interface for chart time range
- `dashboard/lenscommon/registry.go` — global converter registry with `init()` self-registration
- `dashboard/lenscommon/drilldowns.go` — shared drilldown schema helpers and conversion (moved from `models_drilldowns.go`)
- `dashboard/lenscommon/by_reference.go` — shared by_reference read/write (currently duplicated across vis and lens_dashboard_app paths)

### Migrated converters (12 total)

Each moves to `dashboard/panel/lens{kind}/converter.go`:

- `xy` → `lensxy/`
- `gauge` → `lensgauge/`
- `metric` → `lensmetric/`
- `legacy_metric` → `lenslegacymetric/`
- `pie` → `lenspie/`
- `treemap` → `lenstreemap/`
- `mosaic` → `lensmosaic/`
- `datatable` → `lensdatatable/`
- `tagcloud` → `lenstagcloud/`
- `heatmap` → `lensheatmap/`
- `region_map` → `lensregionmap/`
- `waffle` → `lenswaffle/`

Each converter implements `lenscommon.VizConverter` and self-registers:
```go
func init() { lenscommon.Register(xyConverter{}) }
```

### Refactored files

- `models_lens_panel.go` — deleted; registry lives in `lenscommon/`
- `models_lens_by_value_chart_blocks.go` — chart block struct already lives in `models/lens.go`; remaining helpers relocate to their nearest typed homes (lens dashboard app adapter, waffle panel, lens panel) and the file is deleted.
- `models_plan_state_alignment.go` — lens chart alignment delegates to `converter.AlignStateFromPlan(...)`
- `panel_config_defaults.go` — lens chart defaulting delegates to `converter.PopulateJSONDefaults(...)`

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: internal refactoring; XY pinned-panel state alignment now runs through the same per-chart Lens converters as dashboard panels (see Impact ▸ Behavioral note). Typed schema and read/write contracts remain unchanged aside from that drift-handling improvement.

## Impact

### Source files

- 12 new `dashboard/panel/lens*/` directories
- `dashboard/lenscommon/` new package
- Deletion of `models_lens_panel.go`, `models_lens_by_value_chart_blocks.go`

### Tests

- Converter unit tests co-located with each converter package
- Existing acceptance tests for all Lens chart types pass unchanged

### Behavioral note

Pinned panels now receive XY chart state alignment consistent with non-pinned panels (previously only delivered through the dashboard-level slice aligner).

### Examples

None.

### Dependencies and sequencing

- **Depends on:** `dashboard-extract-models`
- **Independent of:** `dashboard-panel-contract` (lens converters are orthogonal to the panel handler registry)
- **Blocks:** `dashboard-composite-panel-contract` (composite handlers consume the lens registry)
- Can be developed in parallel with `dashboard-panel-contract` after `dashboard-extract-models` merges
