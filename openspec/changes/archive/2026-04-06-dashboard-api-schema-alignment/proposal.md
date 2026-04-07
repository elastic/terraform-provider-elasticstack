# Proposal: Dashboard resource schema alignment with the Kibana Dashboard API

## Why

The `elasticstack_kibana_dashboard` resource evolved Terraform-first attribute names and flat structures (`time_from` / `time_to`, `query_language` + `query_text` / `query_json`, `panels[].id`, pie chart `dataset` / `legend` without the `*_json` suffix, heatmap legend visibility as a bool, and a partial `options` object). Practitioners mapping from OpenAPI or Kibana JSON must mentally translate between API and Terraform. Tight alignment with the generated API models (`PostDashboardsJSONBody` / `GetDashboardsId`, panel structs, and enum values) reduces cognitive load, makes documentation mirror Kibana’s vocabulary, and keeps escape-hatch JSON attributes consistently named.

## What Changes

- **Top-level time range**: Replace `time_from`, `time_to`, and `time_range_mode` with a nested `time_range` block: `from`, `to`, and optional `mode` (maps to `KbnEsQueryServerTimeRangeSchema`).
- **Top-level refresh**: Replace `refresh_interval_pause` and `refresh_interval_value` with a nested `refresh_interval` block: `pause`, `value` (maps to `KbnDataServiceServerRefreshIntervalSchema`).
- **Top-level query**: Replace root-level `query_language`, `query_text`, and `query_json` with a nested `query` block containing `language` and **exactly one of** `text` (string branch of the API union) or `json` (normalized JSON for the object branch). Mutual exclusivity and semantics match today’s `query_text` / `query_json` behavior.
- **Panel and section identity**: Rename Terraform attributes `panels[].id` and `sections[].id` to `uid` to match API `uid` (JSON) while leaving the resource’s composite `id` and `dashboard_id` unchanged.
- **Grid**: No rename of `grid` or its `x` / `y` / `w` / `h` attributes for panels; section `grid` remains `y`-only. This change documents parity with the API object shape (no behavioral change beyond `uid`).
- **JSON-backed attributes naming**: Rename pie chart `dataset` and `legend` (both `jsontypes.Normalized`) to `dataset_json` and `legend_json`. Audit other `jsontypes.Normalized` / JSON custom types in this resource for the same `*_json` convention.
- **Heatmap legend visibility**: Replace bool with a string enum `visible` | `hidden` matching `HeatmapLegendVisibility` in the API.
- **Dashboard `options`**: Add `auto_apply_filters` and `hide_panel_borders` to the `options` object, mapped to the API options model on create and update.
- **Typed Lens `time_range` behavior**: Requirements SHALL document that structured Lens panels (`type = "lens"` with typed config blocks) are written with a provider-defined Lens config `time_range` (`KbnDashboardPanelLensConfig0.time_range`); this field remains **unmodeled** in Terraform for typed panels in this change (same as today, but explicit in spec). Raw `config_json` Lens panels remain exempt from typed `lensPanelTimeRange()` injection per REQ-025.

## Capabilities

After this change, practitioners can:

- Author dashboard root attributes in a shape that mirrors Kibana’s `time_range`, `refresh_interval`, and `query` objects.
- Use API-consistent names `uid`, `query.text` / `query.json`, pie `dataset_json` / `legend_json`, heatmap `legend.visibility` enum, and full dashboard `options` flags where implemented.
- Rely on requirements text for Lens typed vs `config_json` time-range behavior.

## Impact

- **Breaking change**: Existing configurations MUST be updated; attribute paths and some types change. **No Terraform state migration** is proposed; practitioners refresh or rewrite configuration to match the new schema.
- **Documentation and tests**: All resource docs, examples, and acceptance test fixtures under `internal/kibana/dashboard/testdata/` must follow the new attribute names.
- **OpenSpec**: Delta updates under this change modify the canonical `kibana-dashboard` spec (schema sketch and affected requirements). Sync or archive after implementation.
