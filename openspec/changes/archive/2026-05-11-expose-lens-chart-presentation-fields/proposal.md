## Why

The Kibana Dashboard API marks `time_range` as a **required** field on every Lens chart panel root (all 12 chart types × DSL + ES|QL variants = 24 schemas). The Terraform resource currently hides this field entirely: every chart panel is hardcoded to `time_range = { from: "now-15m", to: "now" }` via `lensPanelTimeRange()`, with no way for users to override it.

Four additional flat-sibling fields on the same chart roots — `hide_title`, `hide_border`, `drilldowns`, and `references` — are also unexposed, despite being first-class Kibana features that users routinely set in the UI.

This change is unblocked by the recent OpenAPI spec bump, which formalized the field shapes (`kbn-es-query-server-timeRangeSchema`, the discriminated `drilldowns` union, and `kbn-content-management-utils-referenceSchema`) and added them as siblings of the existing typed fields on every chart root.

## What Changes

- **BREAKING** Expose `time_range` as a flat sibling on all twelve typed Lens chart blocks (`xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, `waffle_config`). Shape mirrors the dashboard-level `time_range` block: `{ from, to, mode? }`. Null in config means **inherit the dashboard-level `time_range`** on write; null-preservation on read uses the existing REQ-009 pattern.
- **BREAKING** Drop the hardcoded `lensPanelTimeRange()` default of `now-15m..now`. Replace with inheritance from dashboard-level `time_range` when the chart-level value is null in state.
- Add optional flat siblings `hide_title` (bool), `hide_border` (bool), and `references_json` (normalized JSON string) to all twelve chart blocks.
- Add optional flat sibling `drilldowns` (typed list) to all twelve chart blocks. Each list item is a nested object with three mutually-exclusive variant sub-blocks (`dashboard_drilldown`, `discover_drilldown`, `url_drilldown`) modeling the API discriminated union.
- `dashboard_drilldown.trigger` and `discover_drilldown.trigger` are computed (single-value enums in the API). `url_drilldown.trigger` is required from the user with strict enum validation against the four API-allowed values.
- Resource is unreleased; no migration path needed. Existing state with no `time_range` will start emitting the dashboard-level value on the next apply.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: extends every typed Lens chart block with five new flat-sibling fields (`time_range`, `hide_title`, `hide_border`, `references_json`, `drilldowns`); replaces the hardcoded chart-level time range default with inheritance from dashboard-level `time_range`; adds new requirement scenarios covering inheritance, null-preservation across the chart `time_range`, per-variant drilldown validation, and computed-trigger behavior.

## Impact

- **Code**: schemas in `internal/kibana/dashboard/schema.go` and `schema_xy_chart_panel.go`, `schema_datatable_panel.go`, `schema_tagcloud_panel.go`, `schema_slo_panel.go` (slo blocks unaffected here but the file pattern is the reference); models in all twelve `models_*_panel.go` files for read/write mapping; `models_panels.go` (remove `lensPanelTimeRange()` hardcoded default and route inheritance via the panel mapping path).
- **API client**: no kbapi regeneration needed — the recently bumped client already surfaces `KbnEsQueryServerTimeRangeSchema`, the drilldown discriminated union types, and the chart-root presentation fields.
- **Validators**: reuse `internal/utils/validators/conditional.go` (`AllowedIfDependentPathExpressionOneOf` / `ForbiddenIfDependentPathExpressionOneOf`) for drilldown variant exclusivity within each list item.
- **Tests**: unit tests in each `models_*_panel_test.go`; acceptance tests in `acc_*_panels_test.go`; existing `acc_test.go` round-trip coverage extended for inheritance.
- **Spec**: updated `openspec/specs/kibana-dashboard/spec.md` (schema section + new requirement scenarios) via the delta spec produced in this change.
- **Documentation**: regenerated resource doc under `docs/resources/kibana_dashboard.md` (auto-generated via `make docs-generate`).
