## Why

The `elasticstack_kibana_dashboard` resource schema accumulated inconsistencies across panel types during rapid panel expansion: divergent attribute naming (`config` vs `config_json`), mismatched Terraform types (`Float64` vs `Int64` for `truncate_after_lines`), redundant JSON fields alongside typed equivalents (heatmap `x_axis_json`/`y_axis_json`), and fragmented drilldown schemas. Since the resource is graduating from experimental to GA, this is the last opportunity to align the schema with itself and the Kibana Dashboard API before practitioners depend on it.

## What Changes

- **BREAKING** Rename waffle and pie chart nested metric/group-by attribute from `config` to `config_json` to align with metric chart and datatable conventions.
- **BREAKING** Remove heatmap `x_axis_json` and `y_axis_json` in favour of the existing typed `axis` block (plus model-layer mapping to API `x`/`y` breakdown dimensions).
- **BREAKING** Split treemap and mosaic ES|QL configuration from raw `metrics_json`/`group_by_json` into typed nested schemas matching waffle's `esql_metrics`/`esql_group_by` pattern.
- Normalize partition chart legend `truncate_after_lines` to `Int64` consistently (the API uses `float32`, but fractional line truncation is semantically wrong).
- Fix pie chart to use shared `lensChartBaseAttributes()` instead of inline attribute definitions, removing spurious schema-level defaults on `ignore_global_filters` and `sampling`.
- Make pie chart `data_source_json` **Required** (all other charts enforce this; the API requires it).
- Make XY chart `query` **Optional** (ES|QL XY charts have no query in the API).
- Replace synthetics panels' inline drilldown attributes with the shared `urlDrilldownNestedAttributeObject()` helper.
- Extract duplicated synthetics filter-item schema into a shared helper.
- Reuse `getPartitionValueDisplaySchema()` in waffle instead of inline duplicate.
- Move `panelTypeSloOverview` constant definition to `schema.go` alongside all other panel type constants.
- Add ES|QL mode to `gauge_config` and `tagcloud_config`: make `query` Optional, add typed `esql_metric` / `esql_tag_by` blocks, and implement `fromAPIESQL` / `toAPIESQL` model paths. These are the only two chart types with API ES|QL variants (`GaugeESQL`, `TagcloudESQL`) that the provider currently silently blocks.

## Capabilities

### New Capabilities

<!-- None -->

### Modified Capabilities

- `kibana-dashboard`: Schema attribute normalization across panel types. Requirements affected:
  - Attribute naming conventions for JSON metric/group-by blocks.
  - Legend type consistency (`truncate_after_lines`).
  - Required vs optional status of `data_source_json` and `query` on typed Lens charts.
  - Removal of redundant `x_axis_json`/`y_axis_json` on heatmap in favour of typed `axis` block.
  - ES|QL treemap/mosaic typed schema expansion.

## Impact

- `internal/kibana/dashboard/schema.go` and `schema_*.go`: schema edits (attribute types, names, required/optional status, helper extraction).
- `internal/kibana/dashboard/models_*.go`: model layer edits to match renamed attributes, removed fields, and new ES|QL typed schemas.
- `internal/kibana/dashboard/*_test.go` and `testdata/`: acceptance test configs updated to use new attribute names and shapes.
- `internal/kibana/dashboard/panel_config_defaults.go`: update defaulting for renamed attributes.
- No provider registration or API client changes. Breaking only for existing Terraform configurations targeting the experimental dashboard resource.
