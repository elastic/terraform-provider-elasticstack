## 1. Schema fixes — naming, types, and helpers

- [x] 1.1 Rename waffle `metrics[].config` → `metrics[].config_json` and `group_by[].config` → `group_by[].config_json` in `schema.go`
- [x] 1.2 Rename pie `metrics[].config` → `metrics[].config_json` and `group_by[].config` → `group_by[].config_json` in `schema.go`
- [x] 1.3 ~~Remove heatmap `x_axis_json` and `y_axis_json` from `getHeatmapSchema()` in `schema.go`~~ — dropped from this change; heatmap retains raw `x_axis_json`/`y_axis_json` (see design decision).
- [x] 1.4 Change XY chart `query` from Required to Optional in `schema_xy_chart_panel.go`
- [x] 1.5 Change pie `data_source_json` from Optional to Required in `schema.go`
- [x] 1.6 Remove explicit `Default: booldefault.StaticBool(false)` and `Default: float64default.StaticFloat64(1.0)` from pie chart in `schema.go`; refactor `getPieChart()` to start from `lensChartBaseAttributes()`
- [x] 1.7 Normalize partition legend `truncate_after_lines` to `Int64` (treemap, mosaic, pie in `getPartitionLegendSchema()`)
- [x] 1.8 Extract shared synthetics filter item schema into a helper `syntheticsFilterItemSchema()` and replace inline definitions in `getSyntheticsStatsOverviewSchema()` and `getSyntheticsMonitorsSchema()`
- [x] 1.9 Replace synthetics drilldowns inline definitions with `urlDrilldownNestedAttributeObject()` in `getSyntheticsStatsOverviewSchema()`
- [x] 1.10 Replace synthetics drilldowns inline definitions with `urlDrilldownNestedAttributeObject()` in `getSyntheticsMonitorsSchema()`
- [x] 1.11 Refactor waffle `value_display` to use `getPartitionValueDisplaySchema()` instead of inline definition
- [x] 1.12 Move `panelTypeSloOverview` constant from `models_slo_overview_panel.go` to `schema.go`

## 2. Treemap and mosaic ES|QL typed schema expansion

- [x] 2.1 Add `esql_metrics` typed nested attribute to `getTreemapSchema()` with mutual-exclusion validator against `metrics_json`
- [x] 2.2 Add `esql_group_by` typed nested attribute to `getTreemapSchema()` with mutual-exclusion validator against `group_by_json`
- [x] 2.3 Add `esql_metrics` typed nested attribute to `getMosaicSchema()` with mutual-exclusion validator against `metrics_json`
- [x] 2.4 Add `esql_group_by` typed nested attribute to `getMosaicSchema()` with mutual-exclusion validator against `group_by_json`

## 3. Model layer updates

- [x] 3.1 Update `waffleConfigModel` struct tags and `toAPI`/`fromAPI`/`fromAPIESQL` for `config_json` rename
- [x] 3.2 Update `pieChartConfigModel` struct tags and `toAPI`/`fromAPI`/`fromAPIESQL` for `config_json` rename and `data_source_json` required handling
- [x] 3.3 ~~Update `heatmapConfigModel` to remove `XAxisJSON`/`YAxisJSON` fields; update `toAPI`/`fromAPI`/`fromAPIESQL` to map dimensions through internal representation~~ — dropped from this change (see task 1.3 note).
- [x] 3.4 Update `treemapConfigModel` to add ES|QL typed fields; update `toAPI`/`fromAPI`/`fromAPIESQL`
- [x] 3.5 Update `mosaicConfigModel` to add ES|QL typed fields; update `toAPI`/`fromAPI`/`fromAPIESQL`
- [x] 3.6 Update `xyChartConfigModel` to remove required query handling in `fromAPI`; ensure ES|QL path sets `Query = nil`
- [x] 3.7 Update `syntheticsStatsOverviewConfigModel` and `syntheticsMonitorsConfigModel` for shared filter helper and shared drilldown helper

## 4. Default normalization and config_json defaults

- [x] 4.1 Update `populateLensAttributesDefaults()` in `panel_config_defaults.go` for waffle and pie `config_json` rename
- [x] 4.2 ~~Remove heatmap `x_axis_json`/`y_axis_json` default normalization paths if any exist in `panel_config_defaults.go`~~ — dropped from this change (see task 1.3 note).

## 5. Test fixtures

- [x] 5.1 Update all waffle acceptance test `.tf` fixtures to use `config_json`
- [x] 5.2 Update all pie chart acceptance test `.tf` fixtures to use `config_json` and ensure `data_source_json` is present
- [x] 5.3 ~~Update all heatmap acceptance test `.tf` fixtures to remove `x_axis_json`/`y_axis_json`~~ — dropped from this change (see task 1.3 note).
- [x] 5.4 Update all treemap acceptance test `.tf` fixtures for ES|QL typed schemas (add new ES|QL test fixtures)
- [x] 5.5 Update all mosaic acceptance test `.tf` fixtures for ES|QL typed schemas (add new ES|QL test fixtures)
- [x] 5.6 Update XY chart acceptance test `.tf` fixtures to remove required `query` where testing ES|QL mode
- [x] 5.7 Update synthetics acceptance test `.tf` fixtures for shared drilldown schema shape changes

## 6. Unit and integration tests

- [x] 6.1 Update `models_waffle_panel_test.go` for `config_json` rename
- [x] 6.2 Update `models_pie_chart_panel_test.go` for `config_json` rename and `data_source_json` required handling
- [x] 6.3 ~~Update `models_heatmap_panel_test.go` for removed `x_axis_json`/`y_axis_json`~~ — dropped from this change (see task 1.3 note).
- [x] 6.4 Update `models_treemap_panel_test.go` and `models_mosaic_panel_test.go` for ES|QL typed schemas
- [x] 6.5 Update `models_xy_chart_panel_test.go` for optional query
- [x] 6.6 Update `panel_config_defaults_test.go` for renamed attributes
- [x] 6.7 Run `make build` and fix compilation errors
- [x] 6.8 Run targeted acceptance tests for affected panel types: waffle, pie, heatmap, treemap, mosaic, XY, synthetics

## 7. ES|QL support for gauge and tagcloud

- [x] 7.1 Make `query` Optional in `getGaugeSchema()` and `getTagcloudSchema()` in `schema.go`
- [x] 7.2 Add `esql_metric` typed nested attribute to `getGaugeSchema()` (column, format_json, optional label/color_json/subtitle/goal/max/min/ticks/title) with mutual-exclusion validator against `metric_json`
- [x] 7.3 Add `esql_metric` typed nested attribute to `getTagcloudSchema()` (column, format_json, optional label) with mutual-exclusion validator against `metric_json`
- [x] 7.4 Add `esql_tag_by` typed nested attribute to `getTagcloudSchema()` (column, format_json, color_json, optional label) with mutual-exclusion validator against `tag_by_json`
- [x] 7.5 Implement `fromAPIESQL()` on `gaugeConfigModel` using `kbapi.GaugeESQL`; set `Query = nil` and populate typed `esql_metric` block
- [x] 7.6 Implement `fromAPIESQL()` on `tagcloudConfigModel` using `kbapi.TagcloudESQL`; set `Query = nil` and populate typed `esql_metric` / `esql_tag_by` blocks
- [x] 7.7 Update gauge and tagcloud `fromPanelAPI` dispatch to route ES|QL API payloads to the new `fromAPIESQL` path
- [x] 7.8 Implement ES|QL write path (`toAPI` guard or separate `toAPIESQL`) for gauge and tagcloud that emits `kbapi.GaugeESQL` / `kbapi.TagcloudESQL`
- [x] 7.9 Add gauge ES|QL acceptance test fixture and test case
- [x] 7.10 Add tagcloud ES|QL acceptance test fixture and test case
- [x] 7.11 Add unit tests for `fromAPIESQL` and `toAPIESQL` paths in `models_gauge_panel_test.go` and `models_tagcloud_panel_test.go`

## 8. Finalization

- [x] 8.1 Run `make build` and fix compilation errors
- [x] 8.2 Run `make check-lint` and fix any lint issues
- [x] 8.3 Run `make check-openspec` and fix any spec validation issues
- [x] 8.4 Update CHANGELOG.md with breaking changes summary
- [x] 8.5 Verify all acceptance test fixtures compile with `terraform fmt`
- [x] 8.6 OpenSpec verify change completeness (`openspec validate fix-dashboard-schema-inconsistencies` — note: `openspec verify` subcommand does not exist; used `openspec validate` as the closest valid verification command)
