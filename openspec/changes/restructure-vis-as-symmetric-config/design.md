## Context

The `elasticstack_kibana_dashboard` resource (Terraform Plugin Framework, code under `internal/kibana/dashboard/`) currently exposes 12 typed Lens chart blocks (`xy_chart_config`, `metric_chart_config`, …) plus `markdown_config`, `lens_dashboard_app_config`, controls, SLO, synthetics, and `config_json` as **flat panel-level siblings** (`panelConfigNames`, schema.go). Mutual exclusion across all of them is enforced via `siblingPanelConfigPathsExcept` and `objectvalidator.ConflictsWith` chains.

The Kibana Dashboard API (verified against `generated/kbapi/dashboards.json`) takes a different shape:

```
panel.type ∈ { "vis", "lens-dashboard-app", "markdown", "esql_control",
               "options_list_control", "range_slider_control", "time_slider_control",
               "slo_burn_rate", "slo_overview", "slo_error_budget",
               "synthetics_monitors", "synthetics_stats_overview",
               "discover_session", "image", "slo_alerts" }

panel.type = "vis"                  → config: by_value(12 chart kinds) | by_reference
panel.type = "lens-dashboard-app"   → config: by_value(11 chart kinds) | by_reference   (no legacy_metric)
```

The two `config` unions are essentially identical: a `by_value` `anyOf` of 12 (for vis) or 11 (for lens-app) inline chart kinds, and an object `by_reference` branch with `ref_id`, `references`, `time_range`, `title`/`description`/`hide_title`/`hide_border`, and `drilldowns` (3-way: `dashboard_drilldown`, `discover_drilldown`, `url_drilldown`).

`lens_dashboard_app_config` already mirrors that nested shape in TF (with `by_value` and `by_reference` sub-blocks, and a `lensDashboardAppConfigModeValidator` enforcing exactly-one). The `vis` panel-type does not — its 12 chart kinds and the by-reference branch live at the panel level (chart blocks) or are inaccessible (by_reference).

This change makes the Vis side symmetric with `lens_dashboard_app_config` and tightens the typed-drilldown surface across both panel types.

The resource has not yet shipped a stable release. The user has explicitly approved breaking changes without a migration path.

## Goals / Non-Goals

**Goals:**

- Introduce `viz_config = { by_value, by_reference }` mirroring `lens_dashboard_app_config`.
- Move the 12 typed Lens chart blocks under `viz_config.by_value` (12 chart kinds; legacy_metric included for vis only).
- Provide first-class authoring of by-reference Vis panels via `viz_config.by_reference` (no longer requires the `config_json` escape hatch).
- Replace `lens_dashboard_app_config.by_reference.drilldowns_json` with structured `drilldowns` and add the same structured `drilldowns` to `viz_config.by_reference`.
- Cover the full API drilldown surface (3-way union: dashboard / discover / URL).
- Reduce schema noise: `panelConfigNames` collapses from 24 to ~13 entries; mutual-exclusion blast radius shrinks from 23 siblings per chart block to ~11 within `viz_config.by_value`.
- Keep behavior of all 12 chart blocks (XY, metric, gauge, heatmap, tagcloud, region map, datatable, pie, mosaic, treemap, waffle, legacy metric) functionally unchanged — they only move location.
- Maintain the panel-level `config_json` escape hatch unchanged for `vis` and other panel types.

**Non-Goals:**

- No new panel types (image, slo_alerts, discover_session land separately).
- No additional presentation fields (`hide_title`/`hide_border`/`time_range`/`drilldowns`) **inside** the per-chart by_value blocks — that is a separate change.
- No SLO drilldowns extension. SLO embeddables are URL-only on the API side (verified across `slo-burn-rate`, `slo-overview`, `slo-error-budget` schemas in `dashboards.json`); no backport applies. SLO `drilldowns` blocks remain as-is.
- No removal of `lens_dashboard_app_config.by_value.config_json` (deferred to a separate "tidy" change).
- No fix for the destructive default branch in `mapPanelFromAPI` (separate change; lands first to unblock other panel-type additions).
- No migration path or compatibility shim for the relocated chart blocks.

## Decisions

### D1. Block name: `viz_config`

`viz_config` (chosen) over `vis_config`, `vis`, or `inline_vis_config`. Mirrors `lens_dashboard_app_config` in suffix convention. `viz_` chosen by user preference for readability.

### D2. Symmetric `by_value` / `by_reference` sub-blocks

`viz_config = { by_value = {...}, by_reference = {...} }` with a `vizConfigModeValidator` enforcing exactly-one (mirrors `lensDashboardAppConfigModeValidator`).

Alternative considered: keep chart blocks at panel level and add a sibling `vis_by_reference_config` block. Rejected because it perpetuates the API-mismatch and keeps mutual-exclusion validators sprawling.

### D3. Two helper functions instead of a flag

`getVizByValueAttributes()` (12 chart kinds, includes `legacy_metric_config`) and `getLensDashboardAppByValueAttributes()` (11 chart kinds, no `legacy_metric_config`).

Alternative considered: one helper with `includeLegacyMetric bool` flag. Rejected: 5 lines of duplicated wiring is clearer than a parameter that's only ever called with two distinct values.

The shared `getLensByReferenceAttributes()` helper (returning the by-reference attribute map: `ref_id`, `references_json`, `title`, `description`, `hide_title`, `hide_border`, `drilldowns`, required `time_range`) is single-implementation and used by both panel-type configs.

### D4. No `config_json` inside `viz_config.by_value`

Panel-level `config_json` continues to serve as the universal raw-config escape hatch for any panel type. `viz_config.by_value` is **typed-only** — exactly one of the 12 chart blocks must be set.

Alternative considered: mirror `lens_dashboard_app_config.by_value.config_json` for full symmetry. Rejected: panel-level `config_json` already covers raw `vis` config (set `panel.type = "vis"` and `panel.config_json = "..."`); a nested `viz_config.by_value.config_json` would be a second way to do the same thing. The asymmetry with `lens_dashboard_app_config.by_value.config_json` is acknowledged and deferred to a separate cleanup change that may remove the lens-app variant for the same reason.

### D5. Required `time_range` in `viz_config.by_reference`

Matches the established UX choice in `lens_dashboard_app_config.by_reference` (where `time_range` is required even though the API marks it optional). By-reference panels without an explicit time range are rarely intentional; failing fast at plan time prevents broken dashboards.

### D6. Structured 3-way `drilldowns` (with full trigger surface)

`drilldowns = list(object({ dashboard = ..., discover = ..., url = ... }))` with a per-item validator enforcing exactly one of the three sub-blocks set.

Each variant matches the API schema exactly:

| Variant | Required | Optional | Notes |
|---|---|---|---|
| `dashboard` | `dashboard_id`, `label` | `use_filters` (default true), `use_time_range` (default true), `open_in_new_tab` (default false) | API `trigger` and `type` are constants set by the writer (`on_apply_filter`, `dashboard_drilldown`); not surfaced |
| `discover` | `label` | `open_in_new_tab` (default true) | API `trigger` and `type` constants set by writer |
| `url` | `url`, `label` | `trigger` (∈ `on_click_row`/`on_click_value`/`on_open_panel_menu`/`on_select_range`), `encode_url` (default true), `open_in_new_tab` (default true) | URL drilldowns are the only variant where the API exposes a multi-value `trigger` enum; we expose it as an optional string with `OneOf` validator |

Alternative considered: keep `drilldowns_json`. Rejected because the user specifically requested structured 3-way support; structured drilldowns deliver plan-diff visibility for a meaningful authoring concern.

Alternative considered: omit URL `trigger` (matches today's SLO blocks which don't expose it). Rejected because the API supports it and it's the sole way to author menu-only or range-only drilldowns.

### D7. Migrate `lens_dashboard_app_config.by_reference.drilldowns_json` → structured `drilldowns`

Same shape as `viz_config.by_reference.drilldowns`. The shared `getLensByReferenceAttributes()` helper guarantees both stay in lockstep.

This is **breaking** for any user authoring `lens_dashboard_app_config.by_reference.drilldowns_json`. Pre-release; acceptable.

### D8. Read-back round-trip for structured drilldowns

When the API returns drilldowns the resource cannot losslessly represent in the structured form (e.g., a future API extension with new properties), the read path produces a plan-time error at refresh and surfaces a clear diagnostic. There is no `drilldowns_json` fallback. This is consistent with the resource's general "typed-when-supported, fail loud otherwise" stance for breaking spec changes.

### D9. Shrink `panelConfigNames` and reshape mutual-exclusion

`panelConfigNames` becomes (in order): `config_json`, `markdown_config`, `viz_config`, `lens_dashboard_app_config`, `esql_control_config`, `options_list_control_config`, `range_slider_control_config`, `time_slider_control_config`, `slo_burn_rate_config`, `slo_overview_config`, `slo_error_budget_config`, `synthetics_monitors_config`, `synthetics_stats_overview_config` (13 entries; 1:1 with API panel types + the universal `config_json`).

`siblingPanelConfigPathsExcept` callers and `panelConfigDescription` strings shrink correspondingly. Inside `viz_config.by_value`, a separate, scoped mutual-exclusion list of 12 chart-kind names is used (similar pattern already exists for `lens_dashboard_app_config.by_value` via `lensDashboardAppByValueSourceValidator`).

### D10. `mapPanelFromAPI` dispatch for `case "vis":`

The switch case inspects the `vis.config` discriminator (effectively: try by_reference object shape first; fall through to by_value chart-kind detection). The chart-kind detection logic (currently in `models_lens_panel.go`'s `detectLensVizType`) is reused unchanged — only the assignment target moves from a panel-level field to `viz_config.by_value.<chart_kind_field>`.

`toAPI` symmetric: when `viz_config.by_value` is set, write inline chart config; when `viz_config.by_reference` is set, write the by-reference object. Otherwise (legacy `panel.config_json` with `type = "vis"`), unmarshal `config_json` directly into `KbnDashboardPanelTypeVis_Config` (existing behavior preserved).

## Risks / Trade-offs

- **[Risk] Large HCL surface area updates.** Every test, example, and downstream user dashboard config that uses a typed Vis chart block must wrap it in `viz_config.by_value`.
  - **Mitigation**: Pre-release status removes external migration concerns. Internally, the touch is mechanical (one extra wrapping level); a single sed-style rewrite covers most cases. A draft of the rewrite is included as a tasks.md item with explicit acceptance test inventory.

- **[Risk] Doc churn explosion.** ~12 chart-block descriptions gain context paths (`viz_config.by_value.xy_chart_config` instead of `xy_chart_config`).
  - **Mitigation**: `make docs-generate` regenerates everything; spot-check key examples. Description boilerplate (`Mutually exclusive with …`) actually shrinks because the sibling list is now scoped.

- **[Risk] Read-back round-trip for structured drilldowns may surface latent API drift.** If a real Kibana instance returns a drilldown shape the resource cannot model losslessly (extra fields, future drilldown variant), refresh produces a plan-time error rather than silently degrading.
  - **Mitigation**: Acceptance test coverage for all three variants × representative panels. Document the strict round-trip behavior in the spec scenarios. If brittleness emerges, a follow-up change can add a `drilldowns_json` opt-out at the by_reference level.

- **[Risk] Subtle breakage in mutual-exclusion validators if `panelConfigNames` is incompletely updated.** With chart blocks no longer in the top-level list, any leftover validator referencing the old names compiles but fails at plan time.
  - **Mitigation**: Compile-time enforcement via constant references; targeted unit test that asserts every entry in `panelConfigNames` corresponds to a registered top-level attribute and vice versa.

- **[Risk] User dashboards already imported into state with old shape become unreadable after upgrade.**
  - **Mitigation**: Pre-release; documented in proposal and CHANGELOG. Users re-import or re-author.

- **[Trade-off] Asymmetry between `viz_config.by_value` (no `config_json`) and `lens_dashboard_app_config.by_value` (has `config_json`) is intentional and tracked.** Acknowledged in D4. A follow-up change may remove the lens-app variant for full symmetry.

## Migration Plan

Pre-release resource; no migration shim. CHANGELOG entry highlights:

1. Move every panel-level chart block (e.g. `xy_chart_config = {...}`) to `viz_config.by_value.xy_chart_config = {...}`.
2. Replace any `lens_dashboard_app_config.by_reference.drilldowns_json = jsonencode([...])` with structured `drilldowns = [{ url = {...} }, ...]` (or `dashboard`, `discover` variants).
3. Re-import dashboards that fail refresh after upgrade.

Implementation deploys as a single PR. Rollback is a `git revert` (no schema migrations or external state to undo).

## Open Questions

- **Q1.** `getLensByReferenceAttributes()` is shared between `viz_config.by_reference` and `lens_dashboard_app_config.by_reference`. The latter currently has its own `time_range` validator and description style. The shared helper will normalize both — verify no acceptance scenario depends on the existing description text.
- **Q2.** When the API returns a `vis` panel whose `config` matches *both* a chart-kind shape and the by-reference shape (theoretically impossible per the union, but malformed payloads happen), which branch wins? Proposed: by_reference if `ref_id` is present, otherwise by_value. Document explicitly in spec scenario.
- **Q3.** Should we add a deprecation notice in the panel-level `config_json` description when `panel.type = "vis"` to nudge users toward `viz_config.by_value` for typed authoring? Lean **no** — `config_json` is a legitimate escape hatch and should not be discouraged.
