## Context

The dashboard package contains 12 Lens visualization converters that bridge `kbapi.KbnDashboardPanelTypeVisConfig0` and Terraform's `models.LensByValueChartBlocks`. These converters are currently:
- Declared in a hard-coded slice (`lensVizConverters` in `models_panels.go`)
- Dispatched via a long type-switch (`detectLensVizType` in `models_lens_panel.go`)
- Entangled with shared presentation logic (`lensChartPresentationReadsFor/WritesFor`), drilldown helpers, and by_reference read/write scattered across multiple files

This change extracts a `lenscommon` package with a `VizConverter` interface and moves each converter into its own isolated package with `init()` self-registration.

## Goals / Non-Goals

**Goals:**
- Each Lens chart converter is an isolated package implementing `lenscommon.VizConverter`
- Converters self-register via `init()`
- Shared lens infrastructure (presentation, drilldowns, by_reference) lives in `lenscommon`
- `detectLensVizType` and the hard-coded `lensVizConverters` slice are deleted
- State alignment and JSON defaulting for lens charts delegate to converters

**Non-Goals:**
- No panel handler contract changes (that happens in `dashboard-panel-contract` for simple panels and `dashboard-composite-panel-contract` for composites)
- No user-visible schema or behavior changes

## Decisions

### VizConverter interface

```go
package lenscommon

type Resolver interface {
    ResolveChartTimeRange(chartLevel *models.TimeRangeModel) kbapi.KbnEsQueryServerTimeRangeSchema
}

type VizConverter interface {
    VizType() string
    HandlesBlocks(blocks *models.LensByValueChartBlocks) bool

    // SchemaAttribute returns the SingleNestedAttribute for this chart kind
    // inside vis_config.by_value or lens_dashboard_app_config.by_value.
    SchemaAttribute() schema.Attribute

    PopulateFromAttributes(ctx context.Context, resolver Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics
    BuildAttributes(blocks *models.LensByValueChartBlocks, resolver Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics)
    AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks)
    PopulateJSONDefaults(attrs map[string]any) map[string]any
}
```

`VizType()` returns the Kibana type discriminator string (e.g. `"xy"`, `"gauge"`, `"metric"`).

`HandlesBlocks` returns true when the converter's config block pointer is non-nil in the chart blocks.

`PopulateFromAttributes` reads the typed API chart struct from the union and populates the matching `*models.XYChartConfigModel` (etc.).

`BuildAttributes` does the reverse: takes the model and produces the API union value.

`AlignStateFromPlan` fixes Kibana read drift for that specific chart kind (e.g. XY chart layer alignment).

`PopulateJSONDefaults` applies type-specific defaults inside opaque `config_json` for lens attributes.

### Self-registration

Each converter package calls `lenscommon.Register` in `init()`:

```go
package lensxy

func init() { lenscommon.Register(converter{}) }

type converter struct{}

func (c converter) VizType() string { return string(kbapi.XyChartNoESQLTypeXy) }
func (c converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
    return blocks != nil && blocks.XYChartConfig != nil
}
// ...
```

Registration order does not matter. The registry is a `map[string]VizConverter` keyed by `VizType()`.

```go
package lenscommon

var convertersByType map[string]VizConverter

func Register(c VizConverter) {
    if convertersByType == nil {
        convertersByType = make(map[string]VizConverter)
    }
    convertersByType[c.VizType()] = c
}

func ForType(vizType string) VizConverter { return convertersByType[vizType] }
func FirstForBlocks(blocks *models.LensByValueChartBlocks) (VizConverter, bool) { ... }
func All() []VizConverter { ... }
```

### lenscommon shared infrastructure

```
dashboard/lenscommon/
  iface.go          — VizConverter, Resolver
  registry.go       — global converter registry
  presentation.go   — lensChartPresentationReadsFor, lensChartPresentationWritesFor
  drilldowns.go     — shared drilldown conversion helpers
  by_reference.go   — by_reference read/write shared by vis and lens_dashboard_app
```

`presentation.go` moves from `dashboard/` into `lenscommon/`. The `Resolver` interface abstracts the dependency on `dashboardModel` so lens packages don't need to import `dashboard`.

### lenscommon schema factories

Lens chart config blocks share a common presentation surface (title, description, hide_title, hide_border) and some share drilldowns or time_range. These shapes are defined as factory functions in `lenscommon/schema.go` so individual converters compose rather than duplicate them.

```go
// lenscommon/schema.go
package lenscommon

import (
    "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func PresentationAttributes() map[string]schema.Attribute {
    return map[string]schema.Attribute{
        "title":       schema.StringAttribute{Optional: true},
        "description": schema.StringAttribute{Optional: true},
        "hide_title":  schema.BoolAttribute{Optional: true},
        "hide_border": schema.BoolAttribute{Optional: true},
    }
}

func ByReferenceAttributes() map[string]schema.Attribute {
    return map[string]schema.Attribute{
        "ref_id":          schema.StringAttribute{Required: true},
        "references_json": schema.StringAttribute{Optional: true, CustomType: jsontypes.NormalizedType{}},
        "drilldowns": schema.ListNestedAttribute{
            Optional:     true,
            NestedObject: panelkit.URLDrilldownSchema(),
        },
        "time_range": panelkit.TimeRangeSchema(),
        // ... plus PresentationAttributes merged in later
    }
}
```

Each `VizConverter.SchemaAttribute()` merges `lenscommon.PresentationAttributes()` with chart-specific attributes:

```go
// panel/lensgauge/converter.go (schema method only)
func (c converter) SchemaAttribute() schema.Attribute {
    attrs := lenscommon.PresentationAttributes()
    attrs["data_source_json"] = schema.StringAttribute{Required: true, CustomType: jsontypes.NormalizedType{}}
    attrs["metric_json"] = schema.StringAttribute{Required: true, CustomType: customtypes.NewJSONWithDefaultsType(populateGaugeMetricDefaults)}
    // ... gauge-specific attrs
    return schema.SingleNestedAttribute{
        MarkdownDescription: "Gauge chart configuration.",
        Optional:            true,
        Attributes:          attrs,
    }
}
```

The composite handlers (`visconfig` and `lensdashboardapp`) in Change 4 assemble their `by_value` blocks by iterating `lenscommon.All()` and calling `c.SchemaAttribute()` for each converter. This eliminates `vizByValueSourceAttrNames` and `lensDashboardAppByValueSourceAttrNames` — the source of truth is the live converter registry.

### Converter package layout

```
dashboard/panel/
  lensxy/
    converter.go  — implements lenscommon.VizConverter for XY charts
    converter_test.go
  lensgauge/
    converter.go
    converter_test.go
  lensmetric/
    converter.go
    converter_test.go
  lenslegacymetric/
    converter.go
    converter_test.go
  lenspie/
    converter.go
    converter_test.go
  lenstreemap/
    converter.go
    converter_test.go
  lensmosaic/
    converter.go
    converter_test.go
  lensdatatable/
    converter.go
    converter_test.go
  lenstagcloud/
    converter.go
    converter_test.go
  lensheatmap/
    converter.go
    converter_test.go
  lensregionmap/
    converter.go
    converter_test.go
  lenswaffle/
    converter.go
    converter_test.go
```

### Deleted files

- `models_lens_panel.go` — `detectLensVizType`, `lensVizConverterForType`, the `lensVisualizationConverter` interface all move to `lenscommon/`
- `models_lens_by_value_chart_blocks.go` — the `lensByValueChartBlocks` struct moves to `models/lens.go`; helper functions use `panelkit` reflection or move to `lenscommon/`

### State alignment delegation

`models_plan_state_alignment.go` currently calls per-chart alignment directly:

```go
alignGaugeStateFromPlan(ctx, planBlocks.GaugeConfig, stateBlocks.GaugeConfig)
```

After migration:

```go
for _, c := range lenscommon.All() {
    c.AlignStateFromPlan(ctx, planBlocks, stateBlocks)
}
```

Each converter's `AlignStateFromPlan` checks if its block is present and aligns only if so. This is a small dispatch loop instead of 12 explicit calls.

Similarly, `panel_config_defaults.go` dispatches JSON defaulting:

```go
func populateLensAttributesDefaults(attrs map[string]any) map[string]any {
    vizType, _ := attrs["type"].(string)
    c := lenscommon.ForType(vizType)
    if c != nil {
        return c.PopulateJSONDefaults(attrs)
    }
    return attrs
}
```

### Cross-panel alignment (XY charts)

`alignXYChartStateFromPlanPanels` currently iterates the full panel slice looking for XY charts inside both `viz_config.by_value` and `lens_dashboard_app_config.by_value`. This is a cross-cutting concern that needs access to the full panel list.

After migration, `lenscommon` exposes a registry of "slice aligners":

```go
func RegisterSliceAligner(f func(planPanels, statePanels []models.PanelModel))
```

XY chart converter registers one in `init()`:

```go
func init() {
    lenscommon.Register(converter{})
    lenscommon.RegisterSliceAligner(alignXYChartSlices)
}
```

The generic alignment wrapper in `dashboard/` calls all registered slice aligners after per-panel alignment. Currently only XY needs this; the mechanism is there for future chart types.

## Risks / Trade-offs

- [Risk] `init()` registration order across packages is undefined ➝ *Mitigation:* the registry is a map; order doesn't matter. No converter depends on another.
- [Risk] Missing `init()` call (developer creates new converter package but forgets to register) ➝ *Mitigation:* converter packages are small and follow a template. A missing registration means the converter is never found, caught immediately by tests that exercise the chart type.
- [Risk] `lenscommon` becomes a large package if all shared logic moves there ➝ *Mitigation:* `lenscommon` contains only interfaces, registry, and truly shared helpers. Presentation logic is still substantial but was already substantial in `dashboard/`.

## Open questions

None.
