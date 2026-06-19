## MODIFIED Requirements

### Requirement: Panel default normalization and XY-axis drift prevention (REQ-011)

The resource SHALL normalize `config_json` and typed `vis` panel data with default-aware semantic equality so Kibana-injected defaults do not cause unnecessary drift. This normalization SHALL include panel-type-specific defaults such as missing empty `filters` arrays and visualization metric/grouping defaults used by the implementation. For XY chart panels, when `axis.x.scale` was unset in configuration and Kibana returns the implicit default `ordinal`, the resource SHALL preserve the unset Terraform value instead of forcing `ordinal` into state.

For XY chart `fitting` round-trips, the resource SHALL treat an empty string returned by Kibana for `fitting.type` (which Kibana emits for some layer kinds such as `bar_stacked`) as semantically null and SHALL restore the practitioner's configured `fitting.type` from the plan. The same null-empty-string treatment SHALL apply to `fitting.end_value`. This prevents "Provider produced inconsistent result after apply" diagnostics when bar-style XY layers are used with an explicit `fitting.type` such as `"none"`.

For XY chart `decorations` round-trips on bar-style layers (e.g. `bar`, `bar_stacked`, `bar_horizontal`), Kibana injects server-side bar-styling defaults — `decorations.show_value_labels = false` and `decorations.minimum_bar_height = 1` — even when the practitioner omitted those fields. When the plan value for such a field is null and the API read-back returns the matching default, the resource SHALL preserve the null plan value in state instead of materializing the server default.

For every Lens chart block that exposes `data_source_json` (legacy_metric, region_map, gauge, heatmap, tagcloud, pie, treemap, mosaic, waffle, datatable, and XY data/reference-line layers), Kibana injects `"time_field":"@timestamp"` into the read-back payload when the practitioner omits it. When the practitioner-authored `data_source_json` does not include `time_field`, the resource SHALL strip that injected key from state before semantic comparison and SHALL preserve the practitioner's original JSON payload.

For each Lens chart panel listed below, Kibana materializes hard-coded server defaults for optional fields when the practitioner omits them. The resource SHALL preserve the practitioner's null/unset plan value in state when the API read-back matches the documented default. The known defaults are:
- `gauge_config.styling.shape_json` defaults to `{"type":"bullet","orientation":"horizontal"}`.
- `tagcloud_config.orientation` defaults to `"horizontal"`.
- `tagcloud_config.font_size` defaults to `{min=18, max=72}` (whole block).
- `heatmap_config.axis.{x,y}.labels.visible` default to `true`.
- `heatmap_config.axis.{x,y}.title.visible` default to `false`.
- `heatmap_config.styling.cells.labels.visible` defaults to `false`.
- `heatmap_config.legend.visibility` defaults to `"visible"`.
- `pie_chart_config.label_position` defaults to `"outside"`.
- `treemap_config.legend.visible` and `mosaic_config.legend.visible` default to `"auto"`.
- `treemap_config.value_display` and `mosaic_config.value_display` default to the block `{mode="percentage", percent_decimals=null}` (whole block).

For Lens partition charts (pie `group_by[].config_json`, treemap `group_by_json`, mosaic `group_by_json`/`group_breakdown_by_json`) and Lens datatable (`metrics[].config_json`, `rows[].config_json`, `split_metrics_by[].config_json`), Kibana re-emits each `terms` dimension with the following injected default keys: `rank_by = {type="metric", metric_index=0, direction="desc"}` and `color = {mode="categorical", palette="default", mapping=[]}`. The resource SHALL populate these defaults during semantic-equality comparison so the practitioner's authored JSON round-trips without drift.

When the metric-default normalization injects the `empty_as_null` default into a Lens metric `config_json`, it SHALL inject `empty_as_null = false` ONLY for metric operations whose Kibana API schema accepts the property: `count`, `sum`, and `unique_count`. For all other operations — including `percentile`, `percentile_rank`, `average`, `min`, `max`, `median`, `standard_deviation`, `last_value`, and pipeline operations such as `formula`, `moving_average`, `cumulative_sum`, `differences`, and `counter_rate` — the resource SHALL NOT inject `empty_as_null`, because the corresponding Kibana API metric schema does not define that property and rejects the request with HTTP 400 (`Additional properties are not allowed ('empty_as_null' was unexpected)`). This rule SHALL apply uniformly to every Lens chart family whose metric normalization injects `empty_as_null` — XY (`y[].config_json`), datatable (`metrics[].config_json`), metric chart, pie, gauge, legacy metric, tagcloud, treemap, mosaic, and region map — because all of those families share the same Kibana metric schema in which only `count`, `sum`, and `unique_count` define `empty_as_null`. This gating applies to both the request payload sent to Kibana and the normalization used for semantic-equality comparison, so that operations without `empty_as_null` support neither fail on apply nor produce spurious drift.

#### Scenario: Unset XY X-axis scale

- GIVEN an XY chart panel whose configuration left `axis.x.scale` unset
- WHEN read-back from Kibana returns `axis.x.scale = "ordinal"`
- THEN the provider SHALL keep the Terraform state value unset for that field

#### Scenario: Bar-stacked XY layer with fitting.type = "none"

- GIVEN an XY chart panel with a `bar_stacked` data layer and `fitting = { type = "none" }`
- WHEN create runs and Kibana's read-back returns `fitting.type = ""` (empty string)
- THEN the provider SHALL preserve `fitting.type = "none"` in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: Bar-stacked XY layer omits decorations.show_value_labels and minimum_bar_height

- GIVEN an XY chart panel with a `bar_stacked` data layer whose `decorations` block omits `show_value_labels` and `minimum_bar_height`
- WHEN create runs and Kibana's read-back returns `decorations.show_value_labels = false` and `decorations.minimum_bar_height = 1`
- THEN the provider SHALL keep both fields null in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: data_source_json without time_field round-trips on every Lens chart

- GIVEN a Lens chart panel of any supported type whose `data_source_json` omits `time_field`
- WHEN create runs and Kibana's read-back returns the same payload with `"time_field":"@timestamp"` injected
- THEN the provider SHALL preserve the practitioner's JSON in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal gauge panel preserves null styling.shape_json

- GIVEN a gauge panel whose `gauge_config.styling` block omits `shape_json`
- WHEN create runs and Kibana's read-back returns `styling.shape_json = {"type":"bullet","orientation":"horizontal"}`
- THEN the provider SHALL keep `styling.shape_json` null in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal tagcloud panel preserves null orientation and font_size

- GIVEN a tagcloud panel whose `tagcloud_config` omits `orientation` and `font_size`
- WHEN create runs and Kibana's read-back returns `orientation = "horizontal"` and `font_size = {min=18, max=72}`
- THEN the provider SHALL keep both fields null/unset in state and a subsequent plan SHALL show no changes

#### Scenario: Minimal heatmap panel preserves null axis, styling, and legend defaults

- GIVEN a heatmap panel whose `axis.{x,y}.labels.visible`, `axis.{x,y}.title.visible`, `styling.cells.labels.visible`, and `legend.visibility` are unset
- WHEN create runs and Kibana's read-back returns the documented defaults (`labels.visible=true`, `title.visible=false`, `cells.labels.visible=false`, `legend.visibility="visible"`)
- THEN the provider SHALL keep each of those fields null in state and a subsequent plan SHALL show no changes

#### Scenario: Minimal pie panel preserves null label_position and group_by JSON defaults

- GIVEN a pie panel whose `pie_chart_config.label_position` is unset and whose `group_by[].config_json` for a `terms` operation omits `rank_by` and `color`
- WHEN create runs and Kibana's read-back returns `label_position = "outside"` and injects the partition default keys into `group_by[].config_json`
- THEN the provider SHALL keep `label_position` null in state and SHALL preserve the practitioner's `group_by[].config_json` payload
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal treemap / mosaic panel preserves partition legend and value_display defaults

- GIVEN a treemap or mosaic panel whose `legend.visible` is unset and whose `value_display` block is omitted
- WHEN create runs and Kibana's read-back returns `legend.visible = "auto"` and a default `value_display = {mode="percentage", percent_decimals=null}` block
- THEN the provider SHALL keep `legend.visible` null and SHALL drop the injected `value_display` block from state
- AND a subsequent plan SHALL show no changes

#### Scenario: Datatable terms metrics preserve injected JSON defaults

- GIVEN a datatable panel whose `metrics[].config_json` omits `color`, `empty_as_null`, and `format`
- WHEN create runs and Kibana's read-back re-emits those keys with their documented defaults
- THEN the provider SHALL preserve the practitioner's `metrics[].config_json` payload via semantic-equality comparison
- AND a subsequent plan SHALL show no changes

#### Scenario: XY percentile metric does not inject empty_as_null

- GIVEN an XY `bar_horizontal` panel whose `y[].config_json` uses `operation = "percentile"` with a numeric `percentile` value and omits `empty_as_null`
- WHEN create runs and the provider builds the Kibana API request
- THEN the request payload SHALL NOT contain `empty_as_null` for that metric and Kibana SHALL accept the request (no HTTP 400)
- AND a subsequent plan SHALL show no changes

#### Scenario: XY count metric still injects empty_as_null

- GIVEN an XY panel whose `y[].config_json` uses `operation = "count"` and omits `empty_as_null`
- WHEN create runs and the provider builds the Kibana API request and reads the panel back
- THEN the provider SHALL inject the `empty_as_null = false` default for that metric and the metric SHALL round-trip without drift
