## MODIFIED Requirements

### Requirement: Panel default normalization and XY-axis drift prevention (REQ-011)

The resource SHALL normalize `config_json` and typed `vis` panel data with default-aware semantic equality so Kibana-injected defaults do not cause unnecessary drift. This normalization SHALL include panel-type-specific defaults such as missing empty `filters` arrays and visualization metric/grouping defaults used by the implementation. For XY chart panels, when `axis.x.scale` was unset in configuration and Kibana returns the implicit default `ordinal`, the resource SHALL preserve the unset Terraform value instead of forcing `ordinal` into state.

For XY chart `fitting` round-trips, the resource SHALL treat an empty string returned by Kibana for `fitting.type` (which Kibana emits for some layer kinds such as `bar_stacked`) as semantically null and SHALL restore the practitioner's configured `fitting.type` from the plan. The same null-empty-string treatment SHALL apply to `fitting.end_value`. This prevents "Provider produced inconsistent result after apply" diagnostics when bar-style XY layers are used with an explicit `fitting.type` such as `"none"`.

For XY chart `decorations` round-trips on bar-style layers (e.g. `bar`, `bar_stacked`, `bar_horizontal`), Kibana injects server-side bar-styling defaults — `decorations.show_value_labels = false` and `decorations.minimum_bar_height = 1` — even when the practitioner omitted those fields. When the plan value for such a field is null and the API read-back returns the matching default, the resource SHALL preserve the null plan value in state instead of materializing the server default.

For every Lens chart block that exposes `data_source_json` (legacy_metric, region_map, gauge, heatmap, tagcloud, pie, treemap, mosaic, waffle, datatable, metric, and XY data/reference-line layers), Kibana injects certain optional keys into the read-back payload when the practitioner omits them: `"time_field":"@timestamp"`, and — as of Kibana 9.5.0 GA — `"name"` (echoing the underlying data view's display name for `data_view_spec` sources; earlier Kibana versions omitted this key). When the practitioner-authored `data_source_json` does not include one of these injected keys, the resource SHALL strip that key from state before semantic comparison and SHALL preserve the practitioner's original JSON payload.

For each Lens chart panel listed below, Kibana materializes hard-coded server defaults for optional fields when the practitioner omits them. The resource SHALL preserve the practitioner's null/unset plan value in state when the API read-back matches the documented default. The known defaults are:

- `gauge_config.styling.shape_json` defaults to `{"type":"bullet","orientation":"horizontal"}`.
- `tagcloud_config.orientation` defaults to `"horizontal"`.
- `tagcloud_config.font_size` defaults to `{min=18, max=72}` (whole block).
- `heatmap_config.axis.{x,y}.labels.visible` default to `true`.
- `heatmap_config.axis.{x,y}.title.visible` default to `false`.
- `heatmap_config.axis.x.labels.orientation` defaults to `"horizontal"` (X only; Y has no orientation in the schema or API) (Kibana 9.6.0-SNAPSHOT and later).
- `heatmap_config.styling.cells.labels.visible` defaults to `false`.
- `heatmap_config.legend.visibility` defaults to `"visible"`.
- `heatmap_config.legend.truncate_after_lines` defaults to `1` (Kibana 9.6.0-SNAPSHOT and later).
- `pie_chart_config.label_position` defaults to `"outside"`.
- `pie_chart_config.legend.truncate_after_lines` defaults to `1` (Kibana 9.6.0-SNAPSHOT and later).
- `pie_chart_config.legend.nested` defaults to `false` (Kibana 9.6.0-SNAPSHOT and later).
- `waffle_config.legend.truncate_after_lines` defaults to `1` (Kibana 9.6.0-SNAPSHOT and later).
- `treemap_config.legend.visible` and `mosaic_config.legend.visible` default to `"auto"`.
- `treemap_config.legend.truncate_after_lines` and `mosaic_config.legend.truncate_after_lines` default to `1` (Kibana 9.6.0-SNAPSHOT and later).
- `treemap_config.legend.nested` and `mosaic_config.legend.nested` default to `false` (Kibana 9.6.0-SNAPSHOT and later).
- `treemap_config.value_display` and `mosaic_config.value_display` default to the block `{mode="percentage", percent_decimals=2}` (Kibana 9.6.0-SNAPSHOT and later). The older shape `{mode="percentage", percent_decimals=null}` SHALL also be treated as a Kibana default. When the practitioner omitted `percent_decimals` and Kibana returns `2`, the resource SHALL preserve the null plan value regardless of `mode` (including `mode="absolute"`).
- `xy_chart_config.layers[].reference_line_layer.thresholds[].axis` defaults to `"y"` (Kibana 9.6.0-SNAPSHOT and later).
- When the typed `xy_chart_config.layers[].reference_line_layer.thresholds[].operation` is omitted, Kibana echoes `value_json.operation` onto that sibling (Kibana 9.6.0-SNAPSHOT and later). The `_layers_reference` evidence is a `value_json` blob that already carries `"operation":"static_value"`. The resource SHALL restore the null plan value when that echo is `"static_value"`. Non-`static_value` omitted-operation paths are unprobed and SHALL NOT be assumed to use the same restore.
- `xy_chart_config.layers[].reference_line_layer.thresholds[].color_json` defaults to `{"type":"auto"}` (Kibana 9.6.0-SNAPSHOT and later).

These preservations are plan-relative. On `terraform import` there is no plan value, so Kibana-injected defaults (for example `percent_decimals = 2`, `truncate_after_lines = 1`, `thresholds[].axis = "y"`) SHALL be written to state and a subsequent plan MAY show changes for the affected attributes.

For Lens partition charts (pie `group_by[].config_json`, treemap `group_by_json`, mosaic `group_by_json`/`group_breakdown_by_json`) and Lens datatable (`metrics[].config_json`, `rows[].config_json`, `split_metrics_by[].config_json`), Kibana re-emits each `terms` dimension with the following injected default keys: `rank_by = {type="metric", metric_index=0, direction="desc"}` and `color = {mode="categorical", palette="default", mapping=[]}`. The resource SHALL populate these defaults during semantic-equality comparison so the practitioner's authored JSON round-trips without drift.

When the metric-default normalization injects the `empty_as_null` default into a Lens metric `config_json`, it SHALL inject `empty_as_null = false` ONLY for metric operations whose Kibana API schema accepts the property: `count`, `sum`, and `unique_count`. For all other operations — including `percentile`, `percentile_rank`, `average`, `min`, `max`, `median`, `standard_deviation`, `last_value`, and pipeline operations such as `formula`, `moving_average`, `cumulative_sum`, `differences`, and `counter_rate` — the resource SHALL NOT inject `empty_as_null`, because the corresponding Kibana API metric schema does not define that property and rejects the request with HTTP 400 (`Additional properties are not allowed ('empty_as_null' was unexpected)`). This rule SHALL apply uniformly to every Lens chart family whose metric normalization injects `empty_as_null` — XY (`y[].config_json`), datatable (`metrics[].config_json`), metric chart, pie, gauge, legacy metric, tagcloud, treemap, mosaic, and region map — because all of those families share the same Kibana metric schema in which only `count`, `sum`, and `unique_count` define `empty_as_null`. This gating applies to both the request payload sent to Kibana and the normalization used for semantic-equality comparison, so that operations without `empty_as_null` support neither fail on apply nor produce spurious drift.

As of Kibana 9.6.0-SNAPSHOT, XY `y[].config_json` read-back injects `"axis":"y"` when the practitioner's metric configuration omits `axis`. When the practitioner already set `axis` (for example `"y2"`), Kibana preserves that value. The resource SHALL treat omitted metric `axis` as the default `"y"` on XY `y[]` metric-population paths only, SHALL preserve an explicit metric `axis`, and SHALL NOT derive `"y2"` (or any other axis) from chart-level `axis.y2`. The `color: {type:"auto"}` omit-default, when the practitioner omitted `color`, is already covered by existing metric-default normalization. This axis behavior SHALL NOT apply to grouping/dimension JSON paths such as datatable `rows[]` / `split_metrics_by[]`.

As of Kibana 9.6.0-SNAPSHOT, XY reference-line `thresholds[]` inject omit-defaults when the practitioner omits them: `axis` defaults to `"y"` and `color_json` defaults to `{"type":"auto"}`. When the typed `operation` is omitted, Kibana echoes `value_json.operation` onto that sibling; the resource SHALL restore the null plan value when that echo is `"static_value"` (the only probed shape). When `value_json` is a `static_value` object, Kibana also read-backs a scalar equal to that object's `value` key; the resource SHALL restore the practitioner's authored object when the scalar matches. These defaults SHALL NOT apply to XY `y[]` metric `config_json` (they are sibling threshold attributes). When the practitioner already set `thresholds[].axis` to a value Kibana accepts (for example `"y2"`), Kibana preserves that value and the resource SHALL NOT overwrite it with `"y"`. The Terraform schema currently validates `thresholds[].axis` as `bottom|left|right`, all of which Kibana 9.6 rejects (HTTP 400); reconciling the enum is tracked separately (https://github.com/elastic/terraform-provider-elasticstack/issues/4958). Until then only the omitted-axis default path is reachable.

Datatable `metrics[].config_json` extras injected by Kibana 9.6.0-SNAPSHOT when omitted are `visible:true`, `alignment:"right"`, and `color: {type:"auto"}`. Datatable metrics SHALL NOT assume an `axis` default.

When omitted, the provider SHALL inject `legacy_metric_config.metric_json` extras `color: {type:"auto"}` and `size:"m"` only for field-metric operations (`count`, `sum`, `unique_count`, `min`, `max`, `average`, `median`, `standard_deviation`, `last_value`, `percentile`, `percentile_rank`). The provider SHALL NOT inject these extras for pipeline operations such as `formula`. Legacy metrics SHALL NOT assume an `axis` default.

The resource SHALL NOT treat a practitioner-authored static Y-metric color as semantically equal to Kibana's `{type:"auto"}` read-back. Kibana 9.6 ES|QL XY overwrites practitioner static Y color; that overwrite is out of scope for this requirement.

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

#### Scenario: data_source_json without name round-trips on every Lens chart (Kibana 9.5.0 GA)

- GIVEN a Lens chart panel of any supported type whose `data_source_json` uses a `data_view_spec` source and omits `name`
- WHEN create runs against Kibana 9.5.0 GA (or later) and its read-back returns the same payload with a `"name"` key injected (echoing the data view's display name)
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

- GIVEN a heatmap panel whose `axis.{x,y}.labels.visible`, `axis.x.labels.orientation`, `axis.{x,y}.title.visible`, `styling.cells.labels.visible`, `legend.visibility`, and `legend.truncate_after_lines` are unset
- WHEN create runs and Kibana's read-back returns the documented defaults (`labels.visible=true`, `axis.x.labels.orientation="horizontal"`, `title.visible=false`, `cells.labels.visible=false`, `legend.visibility="visible"`, `legend.truncate_after_lines=1`)
- THEN the provider SHALL keep each of those fields null in state and a subsequent plan SHALL show no changes

#### Scenario: Minimal pie panel preserves null label_position and group_by JSON defaults

- GIVEN a pie panel whose `pie_chart_config.label_position` is unset and whose `group_by[].config_json` for a `terms` operation omits `rank_by` and `color`
- WHEN create runs and Kibana's read-back returns `label_position = "outside"` and injects the partition default keys into `group_by[].config_json`
- THEN the provider SHALL keep `label_position` null in state and SHALL preserve the practitioner's `group_by[].config_json` payload
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal pie panel preserves null legend.truncate_after_lines and legend.nested

- GIVEN a pie panel whose `pie_chart_config.legend.truncate_after_lines` and `legend.nested` are unset
- WHEN create runs against Kibana 9.6.0-SNAPSHOT (or later) and its read-back returns `legend.truncate_after_lines = 1` and `legend.nested = false`
- THEN the provider SHALL keep both `legend.truncate_after_lines` and `legend.nested` null in state
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal waffle panel preserves null legend.truncate_after_lines

- GIVEN a waffle panel whose `waffle_config.legend.truncate_after_lines` is unset
- WHEN create runs and Kibana's read-back returns `legend.truncate_after_lines = 1`
- THEN the provider SHALL keep `legend.truncate_after_lines` null in state and a subsequent plan SHALL show no changes

#### Scenario: Minimal treemap / mosaic panel preserves partition legend and value_display defaults

- GIVEN a treemap or mosaic panel whose `legend.visible` is unset and whose `value_display` block is omitted
- WHEN create runs and Kibana's read-back returns `legend.visible = "auto"` and the default `value_display` block `{mode="percentage", percent_decimals=2}` (or the older shape `{mode="percentage", percent_decimals=null}`)
- THEN the provider SHALL keep `legend.visible` null and SHALL drop the injected `value_display` block from state
- AND a subsequent plan SHALL show no changes

#### Scenario: Treemap / mosaic omitted percent_decimals is preserved when Kibana returns 2

- GIVEN a treemap or mosaic panel whose `value_display.percent_decimals` is unset, including when `value_display.mode = "absolute"`
- WHEN create runs against Kibana 9.6.0-SNAPSHOT (or later) and its read-back returns `percent_decimals = 2`
- THEN the provider SHALL keep `percent_decimals` null in state
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal treemap / mosaic panel preserves null legend.truncate_after_lines and legend.nested

- GIVEN a treemap or mosaic panel whose `legend.truncate_after_lines` and `legend.nested` are unset
- WHEN create runs against Kibana 9.6.0-SNAPSHOT (or later) and its read-back returns `legend.truncate_after_lines = 1` and `legend.nested = false`
- THEN the provider SHALL keep both fields null in state
- AND a subsequent plan SHALL show no changes

#### Scenario: Datatable terms metrics preserve injected JSON defaults

- GIVEN a datatable panel whose `metrics[].config_json` omits `color`, `visible`, `alignment`, `empty_as_null`, and `format`
- WHEN create runs and Kibana's read-back re-emits those keys with their documented defaults (`color: {type:"auto"}`, `visible:true`, `alignment:"right"`, and the existing `empty_as_null`/format defaults) and does not inject `axis`
- THEN the provider SHALL preserve the practitioner's `metrics[].config_json` payload via semantic-equality comparison
- AND a subsequent plan SHALL show no changes

#### Scenario: Legacy metric_json preserves injected color and size defaults

- GIVEN a legacy metric panel whose `metric_json` uses a field-metric `operation` such as `"count"` and omits `color` and `size`
- WHEN create runs against Kibana 9.6.0-SNAPSHOT (or later) and its read-back returns `color: {type:"auto"}` and `size:"m"` without an `axis` key
- THEN the provider SHALL preserve the practitioner's `metric_json` payload in state
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

#### Scenario: XY count metric round-trips on Kibana 9.6 with injected axis and color

- GIVEN an XY panel whose `y[].config_json` uses `operation = "count"`, sets `empty_as_null = true`, and omits both `axis` and `color`
- WHEN create runs against Kibana 9.6.0-SNAPSHOT (or later) and its read-back returns `{"operation":"count","empty_as_null":true,"axis":"y","color":{"type":"auto"}}`
- THEN the provider SHALL preserve the practitioner's original `config_json` payload in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: XY y[] explicit axis is preserved

- GIVEN an XY panel whose `y[].config_json` already sets `axis = "y2"`
- WHEN create runs against Kibana 9.6.0-SNAPSHOT (or later) and its read-back returns `axis = "y2"`
- THEN the provider SHALL preserve the practitioner's `axis` value and SHALL NOT overwrite it with `"y"`
- AND a subsequent plan SHALL show no changes

#### Scenario: XY omitted metric axis is not derived from chart-level axis.y2

- GIVEN an XY panel whose chart-level `axis.y2` is configured and whose `y[].config_json` omits `axis`
- WHEN create runs against Kibana 9.6.0-SNAPSHOT (or later) and its read-back returns `axis = "y"`
- THEN the provider SHALL treat `"y"` as the omitted-axis default and SHALL NOT infer `"y2"` from chart-level `axis.y2`
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal XY reference-line threshold preserves omitted axis, operation, color, and value_json object

- GIVEN an XY chart with a reference-line layer whose `thresholds[]` omit `axis`, `operation`, and `color_json`, and whose `value_json` is a `static_value` object with a numeric `value`
- WHEN create runs against Kibana 9.6.0-SNAPSHOT (or later) and its read-back returns `axis = "y"`, `operation = "static_value"`, `color_json = {"type":"auto"}`, and a scalar `value_json` equal to that object's `value`
- THEN the provider SHALL keep `axis`, `operation`, and `color_json` null in state and SHALL restore the practitioner's `value_json` object
- AND a subsequent plan SHALL show no changes

#### Scenario: Import writes Kibana-injected omit-defaults into state

- GIVEN a Lens panel whose configuration omitted a field listed above (for example treemap `value_display.percent_decimals`, waffle `legend.truncate_after_lines`, or XY `thresholds[].axis`)
- WHEN `terraform import` runs against Kibana 9.6.0-SNAPSHOT (or later)
- THEN state SHALL hold Kibana's injected default for that field
- AND a subsequent plan MAY show changes relative to a configuration that still omits the field

#### Scenario: Kibana ES|QL XY static Y color overwrite is not treated as a default

- GIVEN an ES|QL XY panel whose `y[].config_json` sets a static color
- WHEN Kibana's read-back replaces that color with `{type:"auto"}`
- THEN the provider SHALL NOT treat the practitioner static color and `{type:"auto"}` as semantically equal
