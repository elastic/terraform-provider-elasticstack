## Why

`lens-dashboard-app` by-value panels currently require practitioners to author the full Lens chart payload as opaque `config_json`, even though the provider already exposes rich typed Terraform blocks for many Lens chart shapes on `type = "vis"` panels.

Reusing those typed chart models under `lens_dashboard_app_config.by_value` would make inline `lens-dashboard-app` panels easier to author, validate, and maintain while preserving the raw JSON escape hatch for unsupported or future chart shapes.

## What Changes

- Add typed by-value chart configuration options under `lens_dashboard_app_config.by_value` for Lens chart shapes already supported by the dashboard resource, such as XY, metric, pie, waffle, gauge, heatmap, datatable, treemap, mosaic, tag cloud, region map, and legacy metric where compatible with the generated `lens-dashboard-app` by-value union.
- Keep `by_value.config_json` as an escape hatch for full API-aligned by-value Lens chart JSON.
- Require exactly one by-value source inside `by_value`: either `config_json` or one typed chart block.
- Reuse the existing typed Lens chart schema and converter behavior where the generated `KbnDashboardPanelTypeVisConfig0` and `KbnDashboardPanelTypeLensDashboardAppConfig0` unions share the same chart structs.
- Preserve existing by-reference behavior unchanged.
- Preserve existing `type = "vis"` typed Lens panel behavior unchanged.
- On read, preserve the representation selected by prior Terraform configuration: typed by-value panels remain typed when possible, while imported or raw JSON-authored by-value panels remain represented through `config_json`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: Extend `elasticstack_kibana_dashboard` `lens-dashboard-app` by-value panel requirements to allow typed Lens chart configuration under `lens_dashboard_app_config.by_value` in addition to raw `config_json`.

## Impact

- Affects `internal/kibana/dashboard/schema.go` by extending the nested `lens_dashboard_app_config.by_value` schema.
- Affects dashboard panel models and converters in `internal/kibana/dashboard/models_lens_dashboard_app_panel.go`, `internal/kibana/dashboard/models_lens_dashboard_app_converters.go`, `internal/kibana/dashboard/models_lens_panel.go`, and the existing `models_*_panel.go` Lens chart converters.
- Affects acceptance and unit tests for `elasticstack_kibana_dashboard` lens-dashboard-app panels.
- Requires a delta spec for `openspec/specs/kibana-dashboard/spec.md`.
- No provider dependency or generated client change is expected.
- No state migration is expected because this is additive and `config_json` remains supported.
