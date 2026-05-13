## Context

The `elasticstack_kibana_dashboard` resource was built incrementally across multiple changes (`add-new-panels`, `graduate-kibana-dashboard`). Each new panel type was authored by a different contributor at a different time, using local conventions that diverged from earlier panels:

- Pie chart skipped `lensChartBaseAttributes()` and defined its own `ignore_global_filters` / `sampling` with explicit schema defaults.
- Waffle and pie chart used `metrics[].config` while metric chart and datatable used `metrics[].config_json`.
- Treemap and mosaic ES|QL mode reused the same raw JSON attributes (`metrics_json`, `group_by_json`) as their non-ES|QL mode, while waffle introduced typed ES|QL nested schemas (`esql_metrics`, `esql_group_by`).
- Synthetics panels duplicated an identical inline filter item schema and defined drilldowns as raw inline attributes instead of reusing the shared `urlDrilldownNestedAttributeObject()` helper.
- `panelTypeSloOverview` was defined in `models_slo_overview_panel.go` instead of `schema.go`.

The resource is graduating from experimental to GA (`graduate-kibana-dashboard`). This is the last window to make breaking schema changes without a deprecation cycle.

## Goals / Non-Goals

**Goals:**
- Align metric/group-by JSON attribute naming to `config_json` consistently across all panel types.
- Expand treemap and mosaic ES|QL support to typed nested schemas matching waffle.
- Normalize `truncate_after_lines` to `Int64` in all partition chart legends.
- Fix pie chart to use `lensChartBaseAttributes()`, removing spurious schema defaults.
- Make pie `data_source_json` Required and XY `query` Optional.
- Reuse shared helpers for synthetics filter items, drilldowns, and `value_display`.
- Move `panelTypeSloOverview` constant to `schema.go`.

**Non-Goals:**
- No new panel types.
- No changes to the API client (`kbapi`).
- No changes to dashboard-level attributes (`time_range`, `refresh_interval`, `options`, etc.) beyond what's driven by panel schema alignment.
- No redesign of the `discover_session_config` presentation field placement (API is inconsistent here; changing it would be a large breaking refactor with marginal benefit).

## Decisions

### Decision: Rename `config` → `config_json` for waffle and pie metrics/group-by

**Rationale:** `config_json` is the established convention in datatable and metric chart. Using `config` in waffle and pie was an inconsistency that forces practitioners to remember different names per panel.

**Alternative considered:** Rename metric chart and datatable to `config`. Rejected because `config_json` is more explicit about containing JSON, which is important given the attribute uses a custom JSON type with defaults.

**Implementation:** Update schema definitions in `getWaffleSchema()` and `getPieChart()`, update model structs (`waffleConfigModel`, `pieChartConfigModel`) and their `tfsdk` tags, update `populateLensAttributesDefaults()` in `panel_config_defaults.go`, and update all acceptance test fixtures.

### Decision: Keep heatmap `x_axis_json` / `y_axis_json`

**Rationale:** An earlier iteration of this change proposed removing `x_axis_json` and `y_axis_json` in favor of the typed `axis` block. On closer review, the typed `axis` block only models visual axis configuration (labels, title, orientation), while `x_axis_json` / `y_axis_json` carry the breakdown operation JSON (`terms`, `date_histogram`, `histogram`, `range`, `filters`). The two concerns do not overlap, and removing the raw JSON attributes would leave no way to express the X/Y breakdowns through typed Terraform configuration today. Heatmap therefore retains both blocks as-is; aligning them to a typed dimension schema is deferred to a future change if practitioner demand emerges.

### Decision: Expand treemap/mosaic ES|QL to typed schemas

**Rationale:** Waffle already proved this pattern works. The API's ES|QL types (e.g., `TreemapESQL`) have typed `metrics` and `group_by` arrays with structured fields (`column`, `operation`, `format`, `color`, etc.). Using raw JSON for these in treemap/mosaic loses type safety and forces practitioners to author nested JSON for fields that could be typed.

**Alternative considered:** Collapse waffle ES|QL back to raw JSON. Rejected because typed schemas are more ergonomic and the API supports them.

**Implementation:**
- Create `getTreemapESQLMetricSchema()` and `getTreemapESQLGroupBySchema()` (or generalize waffle's existing helpers).
- Add `esql_metrics` and `esql_group_by` to `getTreemapSchema()` with mutual-exclusion validators against `metrics_json`/`group_by_json`.
- Repeat for mosaic.
- Update `treemapConfigModel`/`mosaicConfigModel`, their `toAPI`/`fromAPI` methods, and `fromAPIESQL` paths.

### Decision: Keep synthetics drilldowns as URL-only but reuse shared helper

**Rationale:** The synthetics panels (stats_overview and monitors) only support URL drilldowns in the API, same as SLO panels. However, the current terraform schema defines drilldown attributes inline rather than using `urlDrilldownNestedAttributeObject()`. Using the shared helper ensures identical descriptions, trigger hardcoding, and null-preservation behaviour.

**Implementation:** Replace inline drilldown definitions in `getSyntheticsStatsOverviewSchema()` and `getSyntheticsMonitorsSchema()` with `urlDrilldownNestedAttributeObject(AllowedTriggers: ["on_open_panel_menu"])`.

### Decision: Keep `discover_session_config` presentation field placement unchanged

**Rationale:** The API places `title`/`description`/`hide_title`/`hide_border` at the root of the `config` object for discover_session panels, outside `by_value`/`by_reference`. The current terraform schema mirrors this by placing them at the root of `discover_session_config`. While this differs from `markdown_config` (which nests them inside `by_value`/`by_reference`), moving them would be a breaking change that affects many test fixtures and doesn't improve API alignment.

**Alternative considered:** Move presentation fields inside `by_value`/`by_reference`. Rejected due to high breaking impact and the fact that the API itself is inconsistent here.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Breaking existing experimental dashboards in practitioner state | Acceptable because the resource is unreleased/graduating; document in CHANGELOG and release notes. |
| Acceptance test fixtures need mass updates | Batch-update test fixtures at the same time as schema changes; run targeted acceptance tests to verify. |
| Model-layer changes for treemap/mosaic ES|QL are non-trivial | Implement one chart at a time; verify with existing waffle ES|QL tests as reference. |

## Migration Plan

1. Schema changes first: rename attributes, add/remove attributes, normalize types.
2. Model layer second: update struct tags, `toAPI`, `fromAPI`, `fromAPIESQL`.
3. Default normalization third: update `panel_config_defaults.go` for renamed attributes.
4. Test fixtures fourth: update all `testdata/` `.tf` files.
5. Acceptance tests fifth: run targeted tests for affected panel types.
6. Build check: `make build`.
7. Archive: mark change complete and sync delta specs to main specs.

No runtime migration needed (state-breaking changes require practitioners to update configs). No database changes.

## Open Questions

- None identified; all decisions are resolved.
