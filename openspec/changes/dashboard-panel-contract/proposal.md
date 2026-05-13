## Why

After `dashboard-extract-models` decouples Terraform data shapes from logic, the dashboard package still contains ~900 lines of switch/case dispatch in `models_panels.go` for panel read/write routing, plus hard-coded panel-type mapping in `panel_config_validator.go`. Adding a new panel requires touching 5+ central files.

This change introduces a `panel/iface` contract and migrates all "simple" panels (non-Lens, non-composite) into isolated subpackages. Each panel implements `iface.Handler` with schema, read, write, validation, and state alignment logic co-located. The router, validator, and schema assembly become registry-driven. Adding a panel in the future requires only a new package and one registration line.

## What Changes

### New infrastructure

- `dashboard/panel/iface/iface.go` — `Handler` and `PinnedHandler` interfaces
- `dashboard/panelkit/` — shared utilities: grid conversion, null-preservation helpers, `tfsdk`-tag-based reflection for config field access
- `dashboard/registry.go` — `panelHandlers` slice, lookup map, `panelConfigNames` derivation
- `dashboard/router.go` — `mapPanelFromAPI` delegates to handler registry; `panelModel.ToAPI` delegates to handler registry

### Migrated panels (10 total)

All panels that are simple (no internal sub-registry, no by_value/by_reference composite branching):

- `slo_burn_rate`
- `slo_overview`
- `slo_error_budget`
- `synthetics_stats_overview`
- `synthetics_monitors`
- `time_slider_control`
- `options_list_control`
- `range_slider_control`
- `esql_control`
- `markdown`

Each gets a `dashboard/panel/{type}/` package with:
- `api.go` — `Handler` implementation, `FromAPI()` and `ToAPI()`
- `model.go` — panel-specific model helpers (exported, testable)
- `schema.go` — `SchemaAttribute()` builder

### Refactored central files

- `models_panels.go` — simple panel cases removed from `mapPanelFromAPI` switch; simple panel branches removed from `panelModel.toAPI()` cascade
- `panel_config_validator.go` — simple panel cases removed from `panelConfigValidateDiags`; replaced with registry iteration
- `schema.go` — `panelConfigNames` derived from registry; no longer hard-coded
- `pinned_panels_mapping.go` — control panel cases delegate to `handler.PinnedHandler()`

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: internal refactoring only. User-visible schema and behavior are unchanged.

## Impact

### Source files

- 10 new `dashboard/panel/{type}/` directories
- `dashboard/panel/iface/`, `dashboard/panelkit/` new packages
- `dashboard/registry.go`, `dashboard/router.go` new files
- Deletion of most of `models_panels.go` switch/case and `toAPI()` cascade for simple panels
- `panel_config_validator.go` simplified

### Tests

- Unit tests co-located with each panel package (`panel/sloburnrate/api_test.go`, etc.)
- Integration tests remain in `dashboard/` (`acc_*_panels_test.go`)
- All existing acceptance tests pass unchanged

### Examples

None.

### Dependencies and sequencing

- **Depends on:** `dashboard-extract-models` (must have `models.PanelModel` in separate package)
- **Blocks:** nothing directly; but other panel migrations should wait to avoid duplicate effort
- Can be developed in parallel with `dashboard-lens-contract` once `dashboard-extract-models` is merged
