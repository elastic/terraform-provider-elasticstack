## Why

The two most complex panel types — `vis` (via `viz_config`) and `lens_dashboard_app` (via `lens_dashboard_app_config`) — remain as monolithic special cases in `models_panels.go` after the simple panel and lens converter migrations. They both consume the lens converter registry from `dashboard-lens-contract`, share by_reference presentation logic, and support composite by_value/by_reference branching. This change completes the contract architecture by extracting them into proper handlers and finalizing all central-file cleanup.

Additionally, the `viz_config` block name is inconsistent with Kibana's wire type `vis`. While the resource has not yet been released, this is the last window to fix the naming before practitioners begin relying on it.

## What Changes

### Composite handler migration

- `dashboard/panel/visconfig/` — new package implementing `iface.Handler` for `type = "vis"`, block name `vis_config`
  - Consumes `lenscommon` registry for by_value chart dispatch
  - Shares `lenscommon.ByReference` for by_reference read/write
- `dashboard/panel/lensdashboardapp/` — new package implementing `iface.Handler` for `type = "lens_dashboard_app"`, block name `lens_dashboard_app_config`
  - Consumes `lenscommon` registry for by_value chart dispatch
  - Shares `lenscommon.ByReference` for by_reference read/write

### User-visible breaking change

- `viz_config` Terraform block renamed to `vis_config` to match Kibana's panel type string `"vis"` and the deriveable block naming convention

### Final cleanup

- `models_panels.go` — switch/case and `toAPI()` cascade fully deleted; replaced by registry-only routing
- `panel_config_validator.go` — hard-coded type switches fully replaced by registry iteration (`ValidatePanelConfig` dispatch)
- `panel_config_defaults.go` — hard-coded lens attribute defaulting fully replaced by registry iteration
- `schema.go` — `getPanelSchema()` assembles attributes from `registry.AllHandlers()`; `getLensDashboardAppByValueNestedAttributes()` and `getVizByValueAttributes()` deleted in favor of registry assembly
- `remove dead code` — all orphaned functions, types, and imports

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: `viz_config` block renamed to `vis_config`. All behavior unchanged.

## Impact

### Source files

- `dashboard/panel/visconfig/`, `dashboard/panel/lensdashboardapp/` new directories
- `models_panels.go` reduced to grid/section helpers and unknown-panel fallback
- `panel_config_validator.go` reduced to registry dispatch + pinned panel logic
- `schema.go` reduced to root schema; panel attributes assembled dynamically

### Tests

- All acceptance tests updated for `vis_config` rename
- Full unit and acceptance test suite must pass

### Examples

- Examples using `viz_config` updated to `vis_config`

### Dependencies and sequencing

- **Depends on:** `dashboard-extract-models`, `dashboard-panel-contract`, `dashboard-lens-contract`
- Must be the **last** change in the sequence; it finalizes the architecture
