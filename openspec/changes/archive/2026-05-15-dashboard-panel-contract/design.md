## Context

After `dashboard-extract-models`, all Terraform data shapes live in `dashboard/models`. The `dashboard` package still routes panel reads and writes through large switch/case blocks in `models_panels.go` and hard-coded validation in `panel_config_validator.go`. Adding a panel requires scattered changes across 5+ files.

This change introduces a Go interface contract (`iface.Handler`) and migrates all simple panels into isolated subpackages. Simple panels are those with no internal sub-registry (no Lens chart dispatch, no by_value/by_reference composite branching). The router, validator, schema assembly, and pinned panel mapping all become registry-driven.

## Goals / Non-Goals

**Goals:**
- Every simple panel type is implemented as an `iface.Handler` in its own package
- `mapPanelFromAPI` delegates to handler registry; no switch/case for simple panels
- `panelModel.toAPI()` delegates to handler registry; no cascading if/else for simple panels
- `panelConfigValidateDiags` delegates to handler registry
- `pinned_panels_mapping.go` delegates to `handler.PinnedHandler()` for controls
- Panel packages import only `models`, `kbapi`, `panelkit`, `iface`, and `lenscommon` where needed

**Non-Goals:**
- Lens chart converters (`lensVisualizationConverter`) stay as-is until `dashboard-lens-contract`
- `viz_config` / `lens_dashboard_app_config` handlers stay as-is until `dashboard-composite-panel-contract`
- No schema behavior changes; no user-visible changes

## Decisions

### The Handler contract

```go
package iface

import (
    "context"
    "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
    "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/diag"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type Handler interface {
    PanelType() string
    SchemaAttribute() schema.Attribute
    FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics
    ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics)
    ValidatePanelConfig(ctx context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics
    AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel)
    ClassifyJSON(config map[string]any) bool
    PopulateJSONDefaults(config map[string]any) map[string]any
    PinnedHandler() PinnedHandler
}

type PinnedHandler interface {
    FromAPI(ctx context.Context, prior *models.PinnedPanelModel, raw kbapi.DashboardPinnedPanels_Item) (models.PinnedPanelModel, diag.Diagnostics)
    ToAPI(ppm models.PinnedPanelModel) (kbapi.DashboardPinnedPanels_Item, diag.Diagnostics)
}
```

`PanelType()` returns the Kibana API discriminator string (e.g. `"slo_burn_rate"`). The config block name is **always** `"slo_burn_rate_config"` — derived as `PanelType() + "_config"`. This is a hard convention enforced by the registry. The `vis` → `vis_config` rename in Change 4 makes this universally true.

`SchemaAttribute()` returns a `schema.SingleNestedAttribute` for the panel's config block.

`FromAPI` receives the full `kbapi.DashboardPanelItem` (already discriminated by the router) and must populate `pm.Grid`, `pm.ID`, `pm.ConfigJSON`, and the panel's typed config field.

`ToAPI` receives a `models.PanelModel` and the parent dashboard, and returns a `kbapi.DashboardPanelItem`.

`ValidatePanelConfig` returns diagnostics **only** when the panel type matches and the configuration is invalid. The router calls every handler's validator; unmatched types return empty diagnostics.

`AlignStateFromPlan` is a no-op for most panels. It is called by the generic alignment wrapper.

`ClassifyJSON` and `PopulateJSONDefaults` support opaque `config_json` defaulting. Most simple panels return `false` / pass-through.

`PinnedHandler` returns `nil` for panels that never appear in the pinned control bar.

### panelkit — shared utilities

```
dashboard/panelkit/
  grid.go             — GridFromAPI, GridToAPI, IDFromAPI, IDToAPI
  nullpreserve.go     — PreserveString, PreserveBool, PreserveList, etc.
  reflection.go       — HasConfig, ClearConfig, SetConfig via tfsdk tag matching
  time_range.go       — timeRangeModelToAPI, resolveChartTimeRange
  schema.go           — Shared schema factories (TimeRangeSchema, URLDrilldownSchema, FilterJSONElementSchema)
```

`HasConfig` / `ClearConfig` use `reflect.TypeOf(models.PanelModel{})` cached at `init()` time, indexing fields by their `tfsdk` tag value. This eliminates the need for each handler to know the struct field name or type.

```go
func HasConfig(pm *models.PanelModel, blockName string) bool
func ClearConfig(pm *models.PanelModel, blockName string)
```

### Registry

```go
// dashboard/registry.go
var panelHandlers = []iface.Handler{
    sloburnrate.Handler{},
    slooverview.Handler{},
    sloerrorbudget.Handler{},
    sloalerts.Handler{},
    syntheticsstatsoverview.Handler{},
    syntheticsmonitors.Handler{},
    timeslider.Handler{},
    optionslist.Handler{},
    rangeslider.Handler{},
    esqlcontrol.Handler{},
    markdown.Handler{},
    image.Handler{},
    // visconfig and lensdashboardapp added in Change 4
}

var panelTypeToHandler map[string]iface.Handler
var panelConfigNames []string

func init() {
    panelTypeToHandler = make(map[string]iface.Handler, len(panelHandlers))
    panelConfigNames = append(panelConfigNames, "config_json")
    for _, h := range panelHandlers {
        panelTypeToHandler[h.PanelType()] = h
        panelConfigNames = append(panelConfigNames, h.PanelType()+"_config")
    }
}

func LookupHandler(panelType string) iface.Handler { return panelTypeToHandler[panelType] }
func AllHandlers() []iface.Handler                  { return panelHandlers }
func ConfigNames() []string                         { return panelConfigNames }
```

### Router

```go
// dashboard/router.go
func mapPanelFromAPI(ctx context.Context, dashboard *models.DashboardModel, tfPanel *models.PanelModel, item kbapi.DashboardPanelItem) (models.PanelModel, diag.Diagnostics) {
    discriminator, _ := item.Discriminator()

    var pm models.PanelModel
    if tfPanel != nil {
        pm = *tfPanel
    }
    pm.Type = types.StringValue(discriminator)

    handler := registry.LookupHandler(discriminator)
    if handler == nil {
        return unknownPanelFromAPI(&pm, item)
    }

    diags := handler.FromAPI(ctx, &pm, tfPanel, item)
    alignPanelStateFromPlan(ctx, tfPanel, &pm)
    return pm, diags
}

func panelModelToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
    for _, h := range registry.AllHandlers() {
        if panelkit.HasConfig(&pm, h.PanelType()+"_config") {
            return h.ToAPI(pm, dashboard)
        }
    }
    return fallbackPanelToAPI(pm)
}
```

The unknown-panel fallback and `config_json` paths remain in `dashboard/` as shared infrastructure.

### Panel package structure

Each simple panel gets a `dashboard/panel/{type}/` package:

```go
// panel/sloburnrate/api.go
package sloburnrate

import (
    "context"
    "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
    "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
    "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
    "github.com/hashicorp/terraform-plugin-framework/diag"
)

type Handler struct{}

func (h Handler) PanelType() string       { return "slo_burn_rate" }
func (h Handler) ClassifyJSON(config map[string]any) bool { return false }
func (h Handler) PopulateJSONDefaults(config map[string]any) map[string]any { return config }
func (h Handler) PinnedHandler() iface.PinnedHandler { return nil }

func (h Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
    apiPanel, _ := item.AsKbnDashboardPanelTypeSloBurnRate()
    pm.Grid = panelkit.GridFromAPI(apiPanel.Grid.X, apiPanel.Grid.Y, apiPanel.Grid.W, apiPanel.Grid.H)
    pm.ID = panelkit.IDFromAPI(apiPanel.Id)
    pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
    return populateFromAPI(pm, prior, apiPanel.Config)
}

func (h Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
    grid := panelkit.GridToAPI(pm.Grid)
    id := panelkit.IDToAPI(pm.ID)
    panel := kbapi.KbnDashboardPanelTypeSloBurnRate{Grid: grid, Id: id}
    diags := buildConfig(pm, &panel)
    var item kbapi.DashboardPanelItem
    item.FromKbnDashboardPanelTypeSloBurnRate(panel)
    return item, diags
}

func (h Handler) ValidatePanelConfig(ctx context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
    // ...
}

func (h Handler) AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
    // no-op for slo_burn_rate
}
```

### panelkit schema factories

Cross-panel sub-schema shapes live as pure factory functions in `panelkit/schema.go`. Panel handlers import and compose them rather than duplicating attribute definitions.

```go
// panelkit/schema.go
package panelkit

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

func TimeRangeSchema() schema.SingleNestedAttribute {
    return schema.SingleNestedAttribute{
        MarkdownDescription: "Time range with from, to, and optional mode.",
        Optional:            true,
        Attributes: map[string]schema.Attribute{
            "from": schema.StringAttribute{Required: true},
            "to":   schema.StringAttribute{Required: true},
            "mode": schema.StringAttribute{Optional: true},
        },
    }
}

func URLDrilldownSchema() schema.NestedAttributeObject {
    return schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{
            "url":           schema.StringAttribute{Required: true},
            "label":         schema.StringAttribute{Required: true},
            "encode_url":    schema.BoolAttribute{Optional: true},
            "open_in_new_tab": schema.BoolAttribute{Optional: true},
        },
    }
}

func FilterJSONElementSchema() schema.NestedAttributeObject {
    return schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{
            "filter_json": schema.StringAttribute{
                CustomType: jsontypes.NormalizedType{},
                Required:   true,
            },
        },
    }
}
```

### Panel handler using shared schemas

The `slo_burn_rate` handler composes its schema from shared factories:

```go
// panel/sloburnrate/schema.go
package sloburnrate

import (
    "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func sloBurnRateSchemaAttributes() map[string]schema.Attribute {
    return map[string]schema.Attribute{
        "slo_id":          schema.StringAttribute{Required: true},
        "duration":        schema.StringAttribute{Required: true},
        "slo_instance_id": schema.StringAttribute{Optional: true},
        "title":           schema.StringAttribute{Optional: true},
        "description":     schema.StringAttribute{Optional: true},
        "hide_title":      schema.BoolAttribute{Optional: true},
        "hide_border":     schema.BoolAttribute{Optional: true},
        "drilldowns": schema.ListNestedAttribute{
            Optional:     true,
            NestedObject: panelkit.URLDrilldownSchema(),
        },
    }
}
```

`panelkit.URLDrilldownSchema()` is also used by `slo_overview`, `slo_alerts`, and the `lenscommon.ByReferenceAttributes()` factory. No duplication of the 4–6 attribute URL drilldown shape.

The `image` panel has a distinct drilldown shape: each drilldown entry carries either a `dashboard_drilldown` sub-block or a `url_drilldown` sub-block (not the URL-only shape). A separate `panelkit.ImageDrilldownSchema()` factory is added for this shape; no other current panel reuses it.

### Null-preservation pattern

The three-branch null-preservation pattern stays in each panel's `populateFromAPI`:

```go
func populateFromAPI(pm, prior *models.PanelModel, api kbapi.SloBurnRateEmbeddable) diag.Diagnostics {
    if prior != nil && prior.SloBurnRateConfig == nil {
        return nil // preserve nil intent
    }
    existing := pm.SloBurnRateConfig
    if existing == nil {
        existing = &models.SloBurnRateConfigModel{}
    }
    // required fields
    existing.SloID = types.StringValue(api.SloId)
    existing.Duration = types.StringValue(api.Duration)
    // optional fields with null-preservation
    if typeutils.IsKnown(existing.SloInstanceID) { ... }
    if typeutils.IsKnown(existing.Title) { ... }
    pm.SloBurnRateConfig = existing
    return nil
}
```

This is not generated. It is declared and maintained per panel. The contract gives it a clean home.

### Control panels and pinned panels

Control panels (`time_slider`, `options_list`, `range_slider`, `esql_control`) implement `PinnedHandler`:

```go
func (h Handler) PinnedHandler() iface.PinnedHandler { return pinnedHandler{} }

type pinnedHandler struct{}
func (p pinnedHandler) FromAPI(...) (models.PinnedPanelModel, diag.Diagnostics) { ... }
func (p pinnedHandler) ToAPI(ppm models.PinnedPanelModel) (kbapi.DashboardPinnedPanels_Item, diag.Diagnostics) { ... }
```

`dashboard/pinned_panels_mapping.go` replaces its hard-coded control switches with:

```go
for _, raw := range *api {
    handler := registry.LookupHandler(discriminator)
    if handler == nil || handler.PinnedHandler() == nil {
        // error: unsupported pinned panel type
    }
    pinnedHandler := handler.PinnedHandler()
    // ... delegate to pinnedHandler.FromAPI / ToAPI
}
```

### Schema assembly

`getPanelSchema()` iterates `registry.AllHandlers()`:

```go
func getPanelSchema() schema.NestedAttributeObject {
    attrs := basePanelAttributes()
    for _, h := range registry.AllHandlers() {
        attrs[h.PanelType()+"_config"] = h.SchemaAttribute()
    }
    return schema.NestedAttributeObject{
        Validators: []validator.Object{panelConfigValidator{}},
        Attributes: attrs,
    }
}
```

### Test organization

- Unit tests for conversion logic live in the panel package: `panel/sloburnrate/api_test.go`
- Integration / acceptance tests stay in `dashboard/acc_slo_burn_rate_panels_test.go`
- The panel package tests exercise exported functions only (`BuildConfig`, `PopulateFromAPI`)
- `dashboard/` tests exercise the full resource lifecycle
- Contract tests use the `contracttest` harness (see below)

### contracttest harness

`panelkit/contracttest` is a test-only package that verifies any `iface.Handler` implementation satisfies the full behavioral contract: round-trip stability, null-preservation of every optional field, `HasConfig`/`ClearConfig` correctness, and schema consistency. Handlers declare a raw Kibana JSON fixture; the harness derives all scenarios from `handler.SchemaAttribute()`.

```
dashboard/panelkit/
  contracttest/
    harness.go       — contracttest.Run entry point and Config struct
    roundtrip.go     — FromAPI → ToAPI stability check
    nullpreserve.go  — schema-driven null-preservation scenario generation
    reflection.go    — HasConfig/ClearConfig post-condition assertions
    schema.go        — SchemaAttribute structural assertions
    parse.go         — raw JSON → kbapi.DashboardPanelItem adapter
```

The harness is imported only from `_test.go` files. It has zero runtime footprint.

#### Usage

Each panel package adds a single call in its `api_test.go`:

```go
// panel/sloburnrate/api_test.go
func TestContract(t *testing.T) {
    contracttest.Run(t, sloburnrate.Handler{}, contracttest.Config{
        // Copied directly from Kibana DevTools, no kbapi construction required
        FullAPIResponse: `{
            "type": "slo_burn_rate",
            "grid": {"x": 0, "y": 0, "w": 24, "h": 6},
            "config": {
                "sloId": "my-slo-id",
                "duration": "5m",
                "title": "Burn Rate",
                "sloInstanceId": "*"
            }
        }`,
    })
}
```

That is the complete test file for the contract. Handwritten assertions are for additional scenarios only.

#### What the harness verifies

**Round-trip stability** (`roundtrip.go`)

```
raw JSON → kbapi.DashboardPanelItem → FromAPI → panelModel → ToAPI → DashboardPanelItem
                                                                        ↓
                                                               semantically ≡ original JSON?
```

Catches missing field mappings and serialization bugs with zero written assertions.

**Schema-driven null-preservation** (`nullpreserve.go`)

The harness walks `handler.SchemaAttribute()` and collects every `Optional: true` leaf. For each it runs three sub-tests:

```
field F declared Optional in SchemaAttribute():
  ① prior.F = null → FromAPI → assert state.F = null     (preserve user's null intent)
  ② prior.F = known → FromAPI → assert state.F = api val  (update when present in API)
  ③ prior = nil    → FromAPI → assert state.F populated   (fresh import path)
```

The handler author writes no null-preservation scenarios. The schema is the specification. Adding a new optional attribute to `SchemaAttribute()` automatically generates a failing test if the handler's `populateFromAPI` forgot to guard it with `typeutils.IsKnown`.

**`HasConfig`/`ClearConfig` correctness** (`reflection.go`)

```
After FromAPI  → assert panelkit.HasConfig(pm, handler.PanelType()+"_config") == true
After ClearConfig → assert panelkit.HasConfig(pm, ...) == false
```

Verifies the reflection cache is correctly wired for every handler without requiring a dashboard resource instantiation.

**Schema structural assertions** (`schema.go`)

```
SchemaAttribute() → assert outer attribute is Optional (never Required)
                 → assert no Required nested attribute is absent from FullAPIResponse fixture
                 → assert ValidatePanelConfig returns error when a Required field is zeroed
```

#### Why this is the right place for null-preservation logic

The three-branch null-preservation pattern in `populateFromAPI` is the most error-prone part of adding a new panel. It is subtle, has no compiler check, and fails silently in plan diffs rather than errors. Current acceptance tests can only catch it if someone happened to write a scenario where an optional field starts null and the user doesn't set it.

With the harness, the null-preservation contract is derived from the schema at test time. A handler that forgets to guard `SloInstanceID` fails `TestContract/NullPreserve/slo_instance_id/prior_null` in under one second, with a field-level error message, without a running Kibana.

## Risks / Trade-offs

- [Risk] Handler `FromAPI` signature is stable for ~10 panels but might need extension for composites ➝ *Mitigation:* composites (Change 4) validate the interface before it's locked; if extension is needed, it happens in Change 2
- [Risk] Reflection in `panelkit.HasConfig` panics on startup if a `tfsdk` tag is missing ➝ *Mitigation:* the `init()` cache iterates all fields; a missing tag for a registered handler panics immediately during package init, caught by `go test`
- [Risk] `ValidatePanelConfig` called for every panel on every plan, O(n*m) where n=panels, m=handlers ➝ *Mitigation:* n is bounded by dashboard panel count (typically <100), m is <20; negligible cost

## Open questions

None.
