## Context

After `dashboard-panel-contract` and `dashboard-lens-contract`, all simple panels and all Lens chart converters live in isolated subpackages with self-registration. Three complex panel handlers remain in the monolith: `vis` (the `viz_config` block), `lens_dashboard_app` (the `lens_dashboard_app_config` block), and `discover_session` (the `discover_session_config` block). The first two are Lens composite handlers: they dispatch to the `lenscommon` converter registry for by_value charts and share by_reference presentation logic. `discover_session` is independently composite: it dispatches to a `dsl` vs `esql` tab sub-registry for by_value, and supports a by_reference path with optional overrides.

This change extracts all three into proper `iface.Handler` implementations and completes the central-file cleanup. It also renames `viz_config` to `vis_config` to match Kibana's panel type string `"vis"` and establish the universal block naming convention.

## Goals / Non-Goals

**Goals:**
- `vis` panel handler lives in `dashboard/panel/visconfig/`
- `lens_dashboard_app` panel handler lives in `dashboard/panel/lensdashboardapp/`
- `discover_session` panel handler lives in `dashboard/panel/discoversession/`
- `vis` and `lens_dashboard_app` consume `lenscommon` registry for by_value chart dispatch
- `vis` and `lens_dashboard_app` share `lenscommon.ByReference` for by_reference read/write
- `discover_session` dispatches by_value to a `dsl` vs `esql` tab selector; handles by_reference with overrides
- `viz_config` renamed to `vis_config` in the Terraform schema
- All central switch/case code eliminated
- Schema, validator, and defaults fully assembled from registries

**Non-Goals:**
- No new panel types added
- No new Lens chart types added
- No behavior changes beyond the `viz_config` → `vis_config` rename

## Decisions

### The `vis` → `vis_config` rename

Kibana's API discriminator for this panel type is `"vis"`. The Terraform block was named `viz_config` (following Kibana UI terminology). After the rename:

- `PanelType()` returns `"vis"`
- Block name is `"vis_config"` (derived from PanelType + "_config")
- This makes the `PanelType + "_config"` convention universally true with zero exceptions

All example files, documentation, and acceptance tests referencing `viz_config` are updated.

### Composite handler pattern

`visconfig.Handler` and `lensdashboardapp.Handler` are `iface.Handler` implementations that are internally composite:

```go
type Handler struct{}

func (h Handler) PanelType() string { return "vis" }

func (h Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
    // 1. Extract grid, id, config JSON
    // 2. Classify config: by_value chart? by_reference? ambiguous?
    // 3. If by_value: detect chart type, look up lenscommon converter, delegate PopulateFromAttributes
    // 4. If by_reference: use lenscommon.ByReference reads
    // 5. If config_json path: preserve JSON, skip typed blocks
}

func (h Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
    // 1. If pm.VizConfig.ByValue has a chart block set:
    //    - look up lenscommon converter by block, delegate BuildAttributes
    // 2. If pm.VizConfig.ByReference set:
    //    - use lenscommon.ByReference writes
    // 3. If config_json only:
    //    - unmarshal JSON into kbapi union directly
}
```

The handler owns the branching logic that currently lives in `models_panels.go` and `models_viz_config.go`.

### Shared by_reference logic in lenscommon

```go
package lenscommon

func ByReferenceFromAPI(ctx context.Context, prior *models.VizByReferenceModel, api kbapi.VisByReferenceConfig, pm *models.PanelModel) diag.Diagnostics
func ByReferenceToAPI(refModel models.VizByReferenceModel, dashboard *models.DashboardModel) (kbapi.VisByReferenceConfig, diag.Diagnostics)
```

Both `visconfig` and `lensdashboardapp` handlers call these. The functions are currently scattered across `models_lens_dashboard_app_converters.go` and `models_vis_api.go`.

### visconfig package structure

```
dashboard/panel/visconfig/
  api.go       — Handler implementation
  schema.go    — vis_config block assembled dynamically from lenscommon
  model.go     — vis-specific helpers, config JSON classification
  api_test.go  — unit tests for by_value dispatch, by_reference round-trip, config_json path
```

The `vis_config` block is assembled at init time by iterating the lens converter registry:

```go
// panel/visconfig/schema.go
package visconfig

import "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"

func byValueAttributes() map[string]schema.Attribute {
    attrs := map[string]schema.Attribute{}
    for _, c := range lenscommon.All() {
        attrs[c.VizType()+"_config"] = c.SchemaAttribute()
    }
    return attrs
}

func schemaAttributes() map[string]schema.Attribute {
    return map[string]schema.Attribute{
        "by_value": schema.SingleNestedAttribute{
            Attributes: byValueAttributes(),
            Validators: []validator.Object{vizByValueSourceValidator{}},
        },
        "by_reference": schema.SingleNestedAttribute{
            Attributes: lenscommon.ByReferenceAttributes(),
        },
    }
}
```

Adding a new Lens chart type in Change 3 automatically surfaces it inside `vis_config.by_value` with zero changes to `visconfig/`. The `vizByValueSourceValidator` still enforces exactly-one-chart-block, but the list of candidate blocks is derived from `lenscommon.All()` rather than a hard-coded slice.

### lensdashboardapp package structure

```
dashboard/panel/lensdashboardapp/
  api.go       — Handler implementation
  schema.go    — lens_dashboard_app_config block assembled dynamically from lenscommon
  model.go     — lens-dashboard-app specific helpers, config classification
  api_test.go  — unit tests
```

Symmetric with `visconfig` but includes a `config_json` escape hatch inside `by_value` alongside the typed chart blocks:

```go
func byValueAttributes() map[string]schema.Attribute {
    attrs := map[string]schema.Attribute{
        "config_json": schema.StringAttribute{
            Optional:   true,
            CustomType: jsontypes.NormalizedType{},
        },
    }
    for _, c := range lenscommon.All() {
        attrs[c.VizType()+"_config"] = c.SchemaAttribute()
    }
    return attrs
}
```

### discoversession package structure

```
dashboard/panel/discoversession/
  api.go       — Handler implementation
  schema.go    — discover_session_config block with by_value and by_reference branches
  model.go     — config classification, tab dispatch helpers
  api_test.go  — unit tests covering by_value DSL, by_value ESQL, and by_reference paths
```

`discover_session` is composite in a different axis from `vis`/`lens_dashboard_app`: its `by_value.tab` block holds either a `dsl` or `esql` sub-block (exactly-one-of enforced by validator), rather than a Lens chart kind. This is analogous to the chart-kind dispatch in `vis`, but does not use the `lenscommon` registry.

```go
type Handler struct{}

func (h Handler) PanelType() string { return "discover_session" }

func (h Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
    // 1. Extract grid, id, config JSON
    // 2. Classify: by_value (DSL tab? ESQL tab?) or by_reference?
    // 3. Delegate to populateByValue or populateByReference
}

func (h Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
    // 1. If by_value with dsl tab set: build DSL by_value API payload
    // 2. If by_value with esql tab set: build ESQL by_value API payload
    // 3. If by_reference: build by_reference API payload
}
```

The `discover_session_config` block schema:

```go
func schemaAttributes() map[string]schema.Attribute {
    return map[string]schema.Attribute{
        "title":       schema.StringAttribute{Optional: true},
        "description": schema.StringAttribute{Optional: true},
        "hide_title":  schema.BoolAttribute{Optional: true},
        "hide_border": schema.BoolAttribute{Optional: true},
        "drilldowns": schema.ListNestedAttribute{
            Optional:     true,
            NestedObject: panelkit.URLDrilldownSchema(),
        },
        "by_value": schema.SingleNestedAttribute{
            Optional: true,
            Attributes: map[string]schema.Attribute{
                "time_range": panelkit.TimeRangeSchema(),
                "tab": schema.SingleNestedAttribute{
                    Required: true,
                    Attributes: map[string]schema.Attribute{
                        "dsl":  dslTabSchema(),   // Optional
                        "esql": esqlTabSchema(),  // Optional
                    },
                    Validators: []validator.Object{discoverSessionTabValidator{}},
                },
            },
        },
        "by_reference": schema.SingleNestedAttribute{
            Optional:   true,
            Attributes: byReferenceAttributes(),
        },
    }
}
```

`discover_session` does not use the `lenscommon.ByReference` shared helper because its by_reference model includes a `selected_tab_id` field and an `overrides` sub-block not present in the Lens by_reference shape.

### Final cleanup

After visconfig, lensdashboardapp, and discoversession handlers are registered, the following files are deleted or stripped to shells:

| File | Fate |
|------|------|
| `models_panels.go` | Delete all switch/case and cascading if/else; keep only `unknownPanelFromAPI`, `fallbackPanelToAPI`, and section helpers |
| `models_lens_dashboard_app_converters.go` | Move by_reference logic to `lenscommon/`, delete file |
| `models_lens_dashboard_app_by_value_adapter.go` | Delete; logic absorbed into `lensdashboardapp.Handler` |
| `models_lens_dashboard_app_panel.go` | Delete; absorbed into `lensdashboardapp/` |
| `models_viz_config.go` | Delete; absorbed into `visconfig/` |
| `models_vis_panel_test.go` | Move tests to `panel/visconfig/api_test.go` |
| `models_vis_api.go` | Delete; absorbed into `lenscommon/by_reference.go` |
| `models_discover_session_panel.go` | Delete; absorbed into `panel/discoversession/` |
| `models_discover_session_panel_test.go` | Move tests to `panel/discoversession/api_test.go` |
| `schema_discover_session_panel.go` | Delete; absorbed into `panel/discoversession/schema.go` |
| `panel_config_validator.go` | Remove all hard-coded panel types; keep only registry dispatch loop and pinned panel logic |
| `panel_config_defaults.go` | Remove hard-coded lens chart dispatch; keep only top-level entry point that delegates to registries |
| `schema.go` | Remove `panelConfigNames` hard-coded slice and the `getLensDashboardAppByValueNestedAttributes` / `getVizByValueAttributes` factory functions; panel attributes assemble from `registry.AllHandlers()`; lens by_value blocks assemble from `lenscommon.All()` |

### Schema assembly

The final `getPanelSchema()` assembles panel attributes from two registries:

```go
func getPanelSchema() schema.NestedAttributeObject {
    attrs := basePanelAttributes() // type, grid, id, config_json

    // Panel handler registry (simple panels + composites)
    for _, h := range registry.AllHandlers() {
        attrs[h.PanelType()+"_config"] = h.SchemaAttribute()
    }

    return schema.NestedAttributeObject{
        Validators: []validator.Object{panelConfigValidator{}},
        Attributes: attrs,
    }
}
```

The Lens composite handlers (`visconfig` and `lensdashboardapp`) in turn assemble their inner `by_value` blocks from the `lenscommon` converter registry. `discoversession` assembles its own static `by_value.tab` schema. The result is a fully dynamic schema: adding a panel handler adds a top-level config block; adding a Lens converter adds a typed chart block inside both Lens composite handlers, all without touching `schema.go`.

The `panelConfigValidator`:

```go
func (v panelConfigValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
    attrs := req.ConfigValue.Attributes()
    typeValue, _ := attrs["type"].(interface{ ValueString() string }).ValueString()

    for _, h := range registry.AllHandlers() {
        resp.Diagnostics.Append(h.ValidatePanelConfig(ctx, typeValue, attrs, req.Path)...)
    }
}
```

No hard-coded panel types remain.

### Dependency chain

```
dashboard-extract-models
    ├──► dashboard-panel-contract
    │        (simple panels + markdown)
    └──► dashboard-lens-contract
             (12 lens converters)
                  └──► dashboard-composite-panel-contract
                            (vis + lens_dashboard_app + final cleanup)
```

## Risks / Trade-offs

- [Risk] `vis_config` rename is a breaking schema change for any non-released practitioner configs ➝ *Mitigation:* resource is not released; no migration needed. Change is documented in changelog.
- [Risk] Composite handlers are the most complex part of the refactor; errors here affect all lens charts ➝ *Mitigation:* comprehensive acceptance test coverage before merge; all lens chart tests must pass
- [Risk] Dead code elimination is error-prone (deleting something still referenced via reflection or late binding) ➝ *Mitigation:* `go build` and `go test` catch all references; no dynamic loading exists

## Open questions

None.
