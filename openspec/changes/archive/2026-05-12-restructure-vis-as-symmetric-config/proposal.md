## Why

The `kibana_dashboard` resource exposes the 12 typed Lens chart blocks (`xy_chart_config`, `metric_chart_config`, …) as siblings at the panel level, while the Kibana Dashboard API actually nests them one level deeper inside the `vis` panel type's `config` union. As a result, the TF schema:

- has no way to author **by-reference** Lens panels that use `panel.type = "vis"` (the API's `vis.config` second branch — a saved-object reference with `ref_id`, `references`, `time_range`, `hide_*`, and `drilldowns` — is unreachable except via `panel.config_json`),
- is **structurally asymmetric** with `lens_dashboard_app_config`, even though the API's `vis.config` and `lens-dashboard-app.config` unions are essentially the same shape (12 inline chart kinds + 1 by-reference object),
- inflates `panelConfigNames` to 24 entries spanning two API levels, which makes mutual-exclusion validators and per-block descriptions list 23 unrelated siblings each.

This change introduces a `viz_config = { by_value, by_reference }` block that mirrors `lens_dashboard_app_config`, aligns the TF panel structure with the API panel structure, and unlocks first-class authoring for by-reference Vis panels and full 3-way structured drilldowns. The resource is not yet released, so the structural cutover is taken without a migration path.

## What Changes

- **BREAKING**: Move the 12 typed Lens chart blocks (`xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, `waffle_config`) from panel-level siblings into `viz_config.by_value`.
- Add `viz_config.by_reference` (`ref_id`, `references_json`, `title`, `description`, `hide_title`, `hide_border`, `drilldowns`, required `time_range`) for by-reference Vis panels.
- Add **structured 3-way `drilldowns`** support (`dashboard_drilldown`, `discover_drilldown`, `url_drilldown`) on both `viz_config.by_reference` and `lens_dashboard_app_config.by_reference`.
- **BREAKING**: Replace `lens_dashboard_app_config.by_reference.drilldowns_json` with the new structured `drilldowns` block list (same shape as on `viz_config.by_reference`).
- Make `time_range` **required** on `viz_config.by_reference` (consistent with `lens_dashboard_app_config.by_reference`).
- Extract shared helpers: `getLensByValueAttributes(includeLegacyMetric bool)` (or two helpers — implementation choice) and `getLensByReferenceAttributes()` reused by both `viz_config` and `lens_dashboard_app_config`.
- Shrink `panelConfigNames` from 24 entries to ~13 (one per API panel `type`, plus the universal panel-level `config_json`); simplify `siblingPanelConfigPathsExcept` callers and per-block descriptions to reflect the new two-layer structure.
- Keep panel-level `config_json` as the universal raw-config escape hatch for `vis` and other panel types (no nested `config_json` inside `viz_config.by_value`).
- **Out of scope (separate changes)**: SLO drilldowns (URL-only on the API side; already aligned), per-by-value-chart presentation fields (`hide_title`/`hide_border`/`time_range`/`drilldowns` inside each chart block), the destructive default-branch fix, removal of `lens_dashboard_app_config.by_value.config_json`, new panel types (image / slo_alerts / discover_session), top-level dashboard `filters` / `pinned_panels`, and per-control field gaps.

## Capabilities

### New Capabilities

(none — this change modifies existing dashboard schema requirements)

### Modified Capabilities

- `kibana-dashboard`: Replace panel-level Lens chart blocks with the new nested `viz_config = { by_value, by_reference }` block; add structured 3-way `drilldowns` to by-reference branches on both `viz_config` and `lens_dashboard_app_config`; tighten mutual-exclusion structure to match API panel-type dispatch.

## Impact

- **Code**:
  - `internal/kibana/dashboard/schema.go` — major: add `viz_config`, move 12 chart blocks, extract reusable helpers, shrink `panelConfigNames`, simplify validators, add structured drilldown attribute set.
  - `internal/kibana/dashboard/models_panels.go`, `models_lens_panel.go`, `models_lens_dashboard_app_*` — refactor `case "vis":` in `mapPanelFromAPI` to populate `viz_config.{by_value|by_reference}`; symmetric write path in `toAPI`; add structured drilldown read/write helpers (shared with `lens_dashboard_app_config.by_reference`).
  - `internal/kibana/dashboard/panel_config_validator.go` — simplify type→block lookup to operate at the panel-type level.
  - All existing chart-block model files (`models_xy_chart_panel.go`, `models_metric_panel.go`, …) — no internal shape changes; only the wiring path (parent attribute name) changes.
- **HCL surface**: every test, example, and user dashboard config that uses a typed Vis chart at panel level must wrap the chart block in `viz_config.by_value`. `lens_dashboard_app_config.by_reference.drilldowns_json` callers must migrate to structured `drilldowns`.
- **Tests**: HCL updates across `acc_*_test.go` and `models_*_test.go`; new acceptance coverage for `viz_config.by_reference` and structured drilldowns (3 trigger types × representative panels).
- **Examples**: `examples/resources/elasticstack_kibana_dashboard/` HCL refactored.
- **Docs**: regenerated via `make docs-generate`.
- **OpenSpec**: `openspec/specs/kibana-dashboard/spec.md` requirements and scenarios updated to reflect the new schema and the structured drilldowns; affected REQ IDs touched include the panel-type / chart-block sections and the `lens_dashboard_app_config` by-reference scenarios.
- **API client / generated code**: no impact — uses existing `kbapi` types, just dispatches the `vis` config union differently.
- **Migration**: none — pre-release resource, simple cutover.
