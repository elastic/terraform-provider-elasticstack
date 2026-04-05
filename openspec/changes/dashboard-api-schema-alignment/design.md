# Design: Dashboard API schema alignment

## Context

The Kibana Dashboard API models dashboard bodies with nested `time_range`, `refresh_interval`, and `query` objects. Panels use `uid` and a `grid` object. Lens inline config (`KbnDashboardPanelLensConfig0`) includes a required `time_range` JSON field distinct from the dashboard-level time range. The provider already maps most of these to Go structs; this change is primarily **Terraform schema and spec alignment**, not new API capabilities.

## Goals

1. Mirror API nesting and names at the Terraform root where practical (`time_range`, `refresh_interval`, `query`).
2. Use `uid` for panel/section identity in HCL to match JSON `uid`.
3. Enforce consistent `*_json` suffixes for JSON-normalized attributes.
4. Align heatmap legend visibility with API enums.
5. Expose remaining dashboard `options` fields that exist in the API client model.
6. Document typed-Lens `time_range` injection without requiring a schema exposure in this change.

## Non-Goals

- Terraform state upgraders or `StateUpgrader` migrations.
- Changing dashboard-level `filters`, `pinned_panels`, or `project_routing` (still out of contract per REQ-009 unless a separate change).
- Exposing Lens panel `time_range` for typed `lens` panels in schema (optional follow-up).
- Renaming panel-internal chart `query` blocks (e.g. `xy_chart_config.query`) to `text`/`json`; those remain the existing structured `query` object until a separate alignment change.

## Decisions

### 1. Nested `time_range`

Single optional/required block (exact optionality should match current semantics: today `time_from` / `time_to` are required at root — preserve requiredness unless product wants optional dashboard time).

```hcl
time_range {
  from = "now-15m"
  to   = "now"
  mode = "relative" # optional; write-only / preserved-in-state behavior unchanged vs REQ-009
}
```

Mapping: unchanged Go paths (`req.TimeRange.*`, read from `data.Data.TimeRange.*`). Update state preservation prose to refer to `time_range.mode` instead of `time_range_mode`.

### 2. Nested `refresh_interval`

```hcl
refresh_interval {
  pause = true
  value = 30
}
```

### 3. Nested `query` with `text` vs `json`

Retain the union semantics of `KbnEsQueryServerQuerySchema`: `language` plus exactly one of string query or object query.

```hcl
query {
  language = "kuery"
  text     = "response.code:200"
}

# OR

query {
  language = "kuery"
  json     = jsonencode({ ... }) # normalized JSON string attribute
}
```

Schema validators SHALL enforce:

- `text` and `json` are mutually exclusive (same as current `query_text` / `query_json`).
- When one branch is set, `language` follows existing rules.

Implementation reuses existing `queryToAPI` / read logic with renamed model fields.

### 4. `uid` on panels and sections

Rename `tfsdk:"id"` → `tfsdk:"uid"` on `panelModel` and `sectionModel` (or equivalent). Composite resource `id` and `dashboard_id` are unchanged. Update REQ-010 wording from “optional `id`” to “optional `uid`”.

### 5. Grid

No attribute renames inside `grid`. Panel `grid` continues to map to API `grid` with `x`, `y`, `w`, `h`. Section `grid` continues to expose only `y` to match API section layout.

### 6. Pie chart: `dataset_json` and `legend_json`

Rename Terraform attributes for `jsontypes.Normalized` fields to `dataset_json` and `legend_json`. Update schema, validators, tests, and REQ-023 examples.

### 7. Heatmap `legend.visibility`

Type: string. Allowed values: `visible`, `hidden` (match `HeatmapLegendVisibility`). Convert to/from API in `heatmapLegendModel`; update REQ-018 and acceptance tests that currently assert bool strings.

### 8. Dashboard `options`

Add:

- `auto_apply_filters` (bool, optional)
- `hide_panel_borders` (bool, optional)

Wire through `optionsModel`, `toAPI`, and `mapOptionsFromAPI`. Extend `isDashboardOptionsDefaultSet` (or equivalent) so that when Terraform omits `options` and Kibana returns only defaults **including** these new flags at their API defaults, state remains a null `options` block where that is the intended behavior—mirror the pattern used for existing option flags.

### 9. Lens `time_range` (API field, not TF for typed panels)

`KbnDashboardPanelLensConfig0` includes `time_range` (JSON). The implementation sets it via `lensPanelTimeRange()` for typed Lens converters. This change:

- **Does not** add Terraform attributes for that field on typed panels.
- **Updates** REQ-013 (and REQ-025 cross-reference) so requirements explicitly state: typed Lens writes send the implementation’s fixed `time_range`; read-back for typed panels does not surface this field as separate Terraform state; `config_json` Lens path preserves API JSON without forced injection.

## Risks

- **Breaking**: All modules and examples must be updated simultaneously with the provider release.
- **Default options**: New option fields may interact with drift suppression; follow existing `isDashboardOptionsDefaultSet` patterns carefully.

## Testing

- Unit tests for renamed model fields and converters.
- Full acceptance test fixture search-replace for new attribute paths.
- Targeted tests for heatmap enum, query XOR, and options round-trip.
