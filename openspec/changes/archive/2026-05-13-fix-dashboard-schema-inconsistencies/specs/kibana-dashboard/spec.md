## MODIFIED Requirements

### Requirement: XY chart panel behavior and typed `vis` `time_range` (REQ-013)

For **typed** `vis` panels (those built through the provider's typed `*_config` blocks and the shared typed visualization write path, not panels managed solely through raw `config_json`), the resource SHALL expose `time_range` as an optional flat sibling attribute on every typed Lens chart block (`xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, `waffle_config`). The attribute SHALL match the dashboard-level `time_range` shape: required `from` (string), required `to` (string), and optional `mode` enum (`absolute` | `relative`).

When the chart-level `time_range` is null in configuration and state, the provider SHALL inherit the dashboard-level `time_range` when assembling the visualization payload, copying the dashboard-level `from`, `to`, and `mode` into the API request. The hardcoded `lensPanelTimeRange()` window (`now-15m..now`) SHALL NOT be used; it is removed in favor of inheritance.

When the chart-level `time_range` is set in configuration, the provider SHALL pass the configured values to the API verbatim, overriding the dashboard-level value for that panel only.

For XY chart `vis` panels specifically, the resource SHALL require `axis`, `decorations`, `fitting`, `legend`, and at least one `layers` entry. The axis object SHALL use `x`, optional primary `y`, and optional `secondary_y`; `axis.x.domain_json` SHALL represent the X-axis domain, and each configured Y axis SHALL require `domain_json`. Each layer SHALL represent either a data layer or a reference-line layer, not both. **`query` SHALL be optional** on the XY chart schema so that ES|QL XY panels (which carry no `query` in the API) are valid without a dummy query block.

REQ-025 governs raw `config_json` `vis` panels; the typed-vs-raw distinction is unchanged.

#### Scenario: Typed `vis` write inherits dashboard time_range when chart time_range is null

- GIVEN a typed `vis` panel on create or update whose chart-level `time_range` is null in configuration
- AND the dashboard-level `time_range` is `{ from = "now-7d", to = "now" }`
- WHEN the provider builds the visualization payload through the typed converter path
- THEN it SHALL set `time_range` on the API payload to `{ from = "now-7d", to = "now" }` copied from the dashboard-level value

#### Scenario: Typed `vis` write uses configured chart-level time_range when set

- GIVEN a typed `vis` panel on create or update whose chart-level `time_range` is set to `{ from = "now-30d", to = "now-1d" }` in configuration
- AND the dashboard-level `time_range` is `{ from = "now-7d", to = "now" }`
- WHEN the provider builds the visualization payload through the typed converter path
- THEN it SHALL set `time_range` on the API payload to the chart-level value `{ from = "now-30d", to = "now-1d" }`

#### Scenario: XY panel requires layers

- GIVEN an XY chart panel configuration
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL require at least one layer and the fixed XY sub-blocks needed by the schema

#### Scenario: ES|QL XY panel omits query

- GIVEN an XY chart panel configured for ES|QL mode (no usable query expression)
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be accepted without a `query` block

### Requirement: Treemap panel behavior (REQ-014)

For treemap `vis` panels, the resource SHALL require `data_source_json`, `group_by_json`, `metrics_json`, and `legend`. It SHALL treat the panel as non-ES|QL when a real `query` is present, and in that mode `query` SHALL be required. It SHALL treat the panel as ES|QL when `query` is omitted or both `query.query` and `query.language` are null. For semantic equality and read-back reconciliation, treemap `group_by_json` and `metrics_json` SHALL normalize the partition defaults used by the implementation, including terms-style defaults such as `collapse_by`, `format`, `rank_by`, and `size`.

When the panel is in ES|QL mode, the resource SHALL expose typed nested schemas for `esql_metrics` and `esql_group_by` matching the structure used by waffle `esql_metrics` and `esql_group_by`. These typed schemas SHALL be mutually exclusive with the non-ES|QL `metrics_json` and `group_by_json` fields respectively.

#### Scenario: Treemap mode selection

- GIVEN a treemap panel with no usable `query`
- WHEN the provider converts it to or from the API model
- THEN it SHALL treat the panel as ES|QL mode rather than non-ES|QL mode

#### Scenario: Treemap ES|QL typed metrics round-trip

- GIVEN a treemap panel in ES|QL mode with `esql_metrics` configured with at least one entry
- WHEN the provider builds the API request and reads the panel back
- THEN the typed `esql_metrics` entries SHALL round-trip without drift

### Requirement: Mosaic panel behavior (REQ-015)

For mosaic `vis` panels, the resource SHALL require `data_source_json`, `group_by_json`, `group_breakdown_by_json`, `metrics_json`, and `legend`. It SHALL use the same ES|QL-vs-non-ES|QL query rule as treemap panels, and non-ES|QL mosaics SHALL require `query`. `metrics_json` SHALL represent exactly one metric in the Terraform model. On read-back, mosaic partition dimensions SHALL be normalized to drop API-emitted top-level null keys that would otherwise create drift.

When the panel is in ES|QL mode, the resource SHALL expose typed nested schemas for `esql_metrics` and `esql_group_by` matching the structure used by waffle `esql_metrics` and `esql_group_by`. These typed schemas SHALL be mutually exclusive with the non-ES|QL `metrics_json` and `group_by_json` fields respectively.

#### Scenario: Mosaic requires secondary breakdown

- GIVEN a mosaic panel configuration
- WHEN Terraform validates or the provider builds the API request
- THEN the panel SHALL require `group_breakdown_by_json` in addition to `group_by_json`

#### Scenario: Mosaic ES|QL typed metrics round-trip

- GIVEN a mosaic panel in ES|QL mode with `esql_metrics` configured with exactly one entry
- WHEN the provider builds the API request and reads the panel back
- THEN the typed `esql_metrics` entries SHALL round-trip without drift

### Requirement: Heatmap panel behavior (REQ-018)

For heatmap `vis` panels, the resource SHALL require `data_source_json`, `axis`, `styling.cells`, `legend`, `metric_json`, and `x_axis_json` (with optional `y_axis_json`). **`legend.visibility` SHALL use the string values `visible` or `hidden`,** matching the API enum. It SHALL treat the panel as non-ES|QL when a real `query` is present, and in that mode `query` SHALL be required. It SHALL treat the panel as ES|QL when `query` is omitted or empty by the implementation's mode test. Heatmap metric normalization SHALL use the same metric-default behavior shared with the tagcloud implementation.

The resource SHALL retain `x_axis_json` and `y_axis_json` as raw JSON attributes for the X and Y breakdown dimensions; this change does not remove them in favor of the typed `axis` block. The typed `axis` block continues to represent visual axis configuration (labels, title, orientation), while `x_axis_json` / `y_axis_json` carry the breakdown operation JSON (e.g. `terms`, `date_histogram`).

#### Scenario: Non-ES|QL heatmap requires query

- GIVEN a heatmap panel using the non-ES|QL branch
- WHEN the provider builds the API request
- THEN it SHALL require `query` to be present

#### Scenario: Heatmap legend visibility enum

- GIVEN `heatmap_config.legend.visibility = "hidden"`
- WHEN the provider builds the API request
- THEN it SHALL set API heatmap legend visibility to the `hidden` enum value

### Requirement: Waffle panel behavior (REQ-019)

For waffle `vis` panels, the resource SHALL enforce mutually exclusive non-ES|QL and ES|QL modes. In non-ES|QL mode it SHALL require `query` and at least one `metrics` entry, and it MAY accept `group_by`. In ES|QL mode it SHALL require at least one `esql_metrics` entry, it MAY accept `esql_group_by`, and it SHALL reject `metrics` and `group_by`. On read-back, the provider SHALL preserve the waffle fields that Kibana may omit or materialize differently, including the implementation's merge behavior for `ignore_global_filters`, `sampling`, legend values, visibility, and value-display details. ES|QL number-format JSON for waffle metric formats SHALL normalize the default decimals and compact settings trimmed by the implementation.

Non-ES|QL waffle metric and group-by entries SHALL use the attribute name **`config_json`** (not `config`) to align with the datatable and metric chart conventions. Each entry SHALL be a JSON string with defaults.

#### Scenario: Waffle ES|QL validation

- GIVEN a waffle panel in ES|QL mode
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL require at least one `esql_metrics` entry and SHALL reject `metrics` or `group_by`

#### Scenario: Waffle non-ES|QL uses config_json

- GIVEN a waffle panel in non-ES|QL mode with `metrics = [{ config_json = jsonencode({ operation = "count" }) }]`
- WHEN the provider builds the API request and reads the panel back
- THEN the metrics SHALL round-trip using the `config_json` attribute name with no plan diff

### Requirement: Pie chart panel behavior (REQ-023)

For pie `vis` panels, the resource SHALL require `data_source_json`, at least one `metrics` entry, and MAY accept `group_by`. It SHALL select the non-ES|QL branch when `query` is present and the ES|QL branch otherwise. When Kibana omits `ignore_global_filters` or `sampling` on read, the provider SHALL treat their default values as `false` and `1.0` respectively. Pie metric and group-by semantic equality SHALL normalize the implementation's pie metric defaults and visualization group-by defaults.

Pie chart attributes SHALL derive from the shared `lensChartBaseAttributes()` helper, so `ignore_global_filters` and `sampling` SHALL be `Optional: true, Computed: true` without explicit Terraform schema defaults. **`data_source_json` SHALL be Required** on `pie_chart_config` to align with all other typed Lens chart blocks.

Pie metric and group-by entries SHALL use the attribute name **`config_json`** (not `config`) to align with the datatable and metric chart conventions.

The resource SHALL expose an optional structured **`legend`** block matching treemap and mosaic legends (attributes `nested`, required `size`, optional `truncate_after_lines`, optional `visible`). The Terraform attribute `legend.visible` SHALL map to the API field `legend.visibility`. When the `legend` block is absent from practitioner configuration, the provider SHALL still build a valid API pie legend by supplying the implementation default legend size `auto`. The Terraform schema SHALL use an optional computed **`legend`** with a default object (typically size and visibility `auto`) so plan-time defaults align with typical Kibana read-back when the block is omitted.

#### Scenario: Pie chart API defaults

- GIVEN a pie panel read from Kibana without explicit `ignore_global_filters` or `sampling`
- WHEN state is refreshed
- THEN the provider SHALL reconcile those fields as `false` and `1.0`

#### Scenario: Pie chart requires data_source_json

- GIVEN a `pie_chart_config` block with no `data_source_json` set
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return an error diagnostic indicating that `data_source_json` is required

#### Scenario: Pie chart uses structured legend

- GIVEN `pie_chart_config.legend` with `size = "auto"` and `visible = "visible"`
- WHEN the provider builds the visualization attributes
- THEN it SHALL encode the pie legend using the API pie legend shape
- AND it SHALL map Terraform `visible` to API `visibility`

#### Scenario: Pie chart legend omitted

- GIVEN `pie_chart_config` with no `legend` block
- WHEN the provider builds the visualization attributes
- THEN it SHALL still produce a valid pie legend object for the API
- AND it SHALL use the implementation default legend size `auto`

#### Scenario: Pie chart read-back uses legend block

- GIVEN a managed pie chart whose API payload contains a legend object
- WHEN the provider refreshes state
- THEN it SHALL populate `pie_chart_config.legend`
- AND it SHALL NOT populate `pie_chart_config.legend_json`

#### Scenario: Pie chart config_json naming

- GIVEN a pie panel with `metrics = [{ config_json = jsonencode({ operation = "count" }) }]`
- WHEN the provider builds the API request and reads the panel back
- THEN the metrics SHALL round-trip using the `config_json` attribute name with no plan diff

### Requirement: Synthetics stats overview panel behavior (REQ-033)

For `type = "synthetics_stats_overview"` panels, the resource SHALL accept an optional `synthetics_stats_overview_config` block. All fields within the block are optional; the panel is valid with an entirely absent or empty config block, in which case it displays monitoring statistics for all Elastic Synthetics monitors visible within the space.

The `synthetics_stats_overview_config` block SHALL expose the following optional attributes:

- `title` (string): display title shown in the panel header.
- `description` (string): descriptive text for the panel.
- `hide_title` (bool): when true, suppresses the panel title in the dashboard.
- `hide_border` (bool): when true, suppresses the panel border in the dashboard.
- `drilldowns` (list of objects): URL drilldown actions attached to the panel. Each drilldown object SHALL contain:
  - `url` (string): the URL template for the drilldown action.
  - `label` (string): the human-readable label for the drilldown action.
  - `encode_url` (bool, optional): whether to URL-encode the drilldown target; defaults to `true` at the API level.
  - `open_in_new_tab` (bool, optional): whether to open the drilldown in a new browser tab; defaults to `true` at the API level.
  - The API fields `trigger` and `type` each accept only one value (`on_open_panel_menu` and `url_drilldown` respectively). These SHALL be hardcoded in the write converter and SHALL NOT be exposed as user-configurable Terraform attributes.
- `filters` (nested block, optional): Synthetics-specific monitor filter constraints. Each filter category within the block is optional and accepts a `list(object({ label = string, value = string }))`:
  - `projects`: filter by Synthetics project.
  - `tags`: filter by monitor tag.
  - `monitor_ids`: filter by monitor ID (the API accepts up to 5000 entries).
  - `locations`: filter by monitor location.
  - `monitor_types`: filter by monitor type (e.g. `browser`, `http`).
  - `statuses`: filter by monitor status.

On write, the resource SHALL include only the attributes that are known and non-null in the Terraform configuration. Absent optional fields SHALL NOT be sent in the API request.

On read-back, when Kibana returns an empty or absent config object for a `synthetics_stats_overview` panel, the resource SHALL preserve `null` in state for the `synthetics_stats_overview_config` block rather than materializing an empty block. Individual optional fields absent from the API response SHALL remain null in state rather than being forced to default values.

On read-back, when Kibana returns a nil or empty `filters` object, the resource SHALL treat it as equivalent to an absent `filters` block and SHALL NOT populate the `filters` block in state.

The `synthetics_stats_overview_config` block SHALL be mutually exclusive with all other typed panel config blocks and with `config_json`. The resource SHALL return an error diagnostic if `config_json` is used with `type = "synthetics_stats_overview"`.

#### Scenario: Panel with no config

- GIVEN a panel with `type = "synthetics_stats_overview"` and no `synthetics_stats_overview_config` block
- WHEN create or update runs
- THEN the resource SHALL send a valid panel API payload without a config body and SHALL NOT return an error

#### Scenario: Panel with filter constraints

- GIVEN a `synthetics_stats_overview_config` block with a `filters` sub-block containing one or more filter category lists
- WHEN create or update runs
- THEN the resource SHALL include those filter constraints in the panel config payload sent to Kibana

#### Scenario: Read-back with empty API config

- GIVEN a `synthetics_stats_overview` panel whose API response contains no config fields
- WHEN read runs
- THEN the resource SHALL keep `synthetics_stats_overview_config` null in state

#### Scenario: Read-back with empty filters

- GIVEN a `synthetics_stats_overview` panel whose API response contains a `filters` object with no entries (non-nil but empty)
- WHEN read runs (including refresh, not just import)
- THEN the resource SHALL set `filters` to null in state regardless of the prior state value

#### Scenario: Panel config block exclusivity

- GIVEN a panel with `type = "synthetics_stats_overview"` and `config_json` also set
- WHEN the provider builds the API request
- THEN it SHALL return an error diagnostic for unsupported `config_json` panel type

### Requirement: Synthetics monitors panel behavior (REQ-034)

For `type = "synthetics_monitors"` panels, the resource SHALL accept an optional `synthetics_monitors_config` block. The block, if present, may contain an optional `filters` nested block. Within `filters`, all six filter dimensions (`projects`, `tags`, `monitor_ids`, `locations`, `monitor_types`, `statuses`) are optional lists of `{ label, value }` objects.

The `synthetics_monitors` panel type is a standalone panel, not a `vis` visualization. It does not reference a saved object, and its configuration is fully inline in the dashboard document. None of the typed visualization converters, typed visualization time-range injection, or visualization metric default normalization SHALL apply to `synthetics_monitors` panels.

**On write (create and update):**

When `synthetics_monitors_config` is set, the resource SHALL map the config block to the panel's `config` object in the API request. When the `filters` block is set, the resource SHALL include the `filters` sub-object with only the filter dimensions that are explicitly configured. Filter dimensions that are not set SHALL be omitted from the API request rather than sent as empty arrays. When `synthetics_monitors_config` is omitted entirely, the resource SHALL send an empty `config` object `{}` or omit `config` from the panel payload, consistent with how other all-optional panel config blocks are handled.

**On read:**

When Kibana returns a `synthetics_monitors` panel with an empty or absent `config` object, the provider SHALL keep `synthetics_monitors_config` null in state. When Kibana returns a present `config` with an empty or absent `filters` object, the provider SHALL keep the `filters` block null in state. When Kibana returns individual filter dimension arrays that are empty, the provider SHALL treat them as equivalent to omitted dimensions and SHALL NOT force empty lists into state.

The provider SHALL seed `synthetics_monitors_config` from prior state or plan on read-back, so that filter dimensions omitted by Kibana do not overwrite Terraform-authored values with null.

**Shared filter model:**

The filter structure used by `synthetics_monitors_config` (lists of `{ label, value }` pairs for each filter dimension) is identical to the filter structure used by `synthetics_stats_overview_config` (REQ-033). Both panel types SHALL consume the same shared nested-block schema function for filter items, eliminating the current inline duplication.

#### Scenario: Synthetics monitors panel with no config block

- GIVEN a dashboard configuration containing a `synthetics_monitors` panel with no `synthetics_monitors_config` block
- WHEN the resource is created
- THEN the provider SHALL send a valid API request for the panel without a populated `config` object
- AND state SHALL record `synthetics_monitors_config` as null
- AND a subsequent plan SHALL show no changes

#### Scenario: Synthetics monitors panel with filters

- GIVEN a dashboard configuration containing a `synthetics_monitors` panel with:
  - `type = "synthetics_monitors"`
  - `synthetics_monitors_config.filters.projects = [{ label = "My Project", value = "my-project" }]`
  - `synthetics_monitors_config.filters.statuses = [{ label = "Up", value = "up" }, { label = "Down", value = "down" }]`
- WHEN the resource is created
- THEN the provider SHALL send the mapped `config.filters` object to the Kibana dashboard API with the `projects` and `statuses` dimensions populated
- AND the panel SHALL appear in state with those filter dimensions populated
- AND omitted filter dimensions (`tags`, `monitor_ids`, `locations`, `monitor_types`) SHALL remain null in state

#### Scenario: Read-back null preservation when config is empty

- GIVEN a managed `synthetics_monitors` panel whose `synthetics_monitors_config` is null (no config block)
- WHEN Kibana returns the panel with an empty `config` object `{}`
- THEN the provider SHALL keep `synthetics_monitors_config` null in state
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: Read-back null preservation when filters is empty

- GIVEN a managed `synthetics_monitors` panel with `synthetics_monitors_config` set but `filters` omitted
- WHEN Kibana returns the panel with a `config` containing an empty `filters` object `{}`
- THEN the provider SHALL keep the `filters` block null in state
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: All filter dimensions set

- GIVEN a `synthetics_monitors` panel with all six filter dimensions configured in `filters`
- WHEN the resource is created and read back
- THEN all six filter dimensions SHALL be present in state
- AND a subsequent plan SHALL show no changes

#### Scenario: monitor_ids large list (API constraint documentation)

- GIVEN a `synthetics_monitors` panel with `filters.monitor_ids` containing more than 5000 items
- WHEN the provider sends the API request
- THEN the API MAY return an error; the provider SHALL surface that error as a diagnostic
- AND the provider SHALL NOT enforce a plan-time validator for the 5000-item limit (this is an API-side constraint)

### Requirement: Tagcloud panel behavior (REQ-017)

For tagcloud `vis` panels, the resource SHALL support both non-ES|QL and ES|QL modes.

Non-ES|QL mode requires a `query` block with non-null `expression` and `language`; in that mode the resource SHALL require `data_source_json`, `query`, `metric_json`, and `tag_by_json`, with optional `filters`, `ignore_global_filters`, `sampling`, `orientation`, and `font_size`. For semantic equality it SHALL normalize tagcloud metric defaults and the `terms`-operation defaults for `tag_by_json`, including the default `rank_by` value.

ES|QL mode is selected when `query` is omitted or both `expression` and `language` are null; in ES|QL mode the resource SHALL require `data_source_json` and typed `esql_metric` and `esql_tag_by` blocks instead of `metric_json` and `tag_by_json`. The `query` attribute SHALL be Optional on the schema so that ES|QL configurations are valid. The `esql_metric` block SHALL contain required `column` (string) and `format_json` (normalized JSON for the format type), and optional `label` (string). The `esql_tag_by` block SHALL contain required `column` (string), `format_json` (normalized JSON), and `color_json` (normalized JSON for the color mapping), and optional `label` (string).

#### Scenario: Tagcloud terms defaults

- GIVEN a tagcloud panel whose `tag_by_json` uses the `terms` operation without an explicit `rank_by`
- WHEN state is compared or refreshed
- THEN the provider SHALL treat the default `rank_by` as part of semantic equality

#### Scenario: Tagcloud ES|QL round-trip

- GIVEN a tagcloud panel in ES|QL mode with `esql_metric.column = "count"` and `esql_tag_by.column = "host"`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the typed `esql_metric` and `esql_tag_by` blocks and `metric_json`/`tag_by_json` SHALL be null

### Requirement: Gauge panel behavior (REQ-021)

For gauge `vis` panels, the resource SHALL support both non-ES|QL and ES|QL modes.

Non-ES|QL mode requires a `query` block with non-null `expression` and `language`; in that mode the resource SHALL require `data_source_json`, `query`, and `metric_json`, and it MAY accept `shape_json`, `filters`, `ignore_global_filters`, and `sampling`. Gauge metric semantic equality SHALL include the implementation's defaults for `empty_as_null`, `hide_title`, and `ticks`.

ES|QL mode is selected when `query` is omitted or both `expression` and `language` are null; in ES|QL mode the resource SHALL require `data_source_json` and a typed `esql_metric` block instead of `metric_json`. The `query` attribute SHALL be Optional on the schema so that ES|QL configurations are valid. The `esql_metric` block SHALL contain required `column` (string) and `format_json` (normalized JSON for the format type), and optional `label` (string), `color_json` (normalized JSON for the gauge fill color), `subtitle` (string), `goal` (object: required `column` string, optional `label` string), `max` (object: required `column` string, optional `label` string), `min` (object: required `column` string, optional `label` string), `ticks` (object: optional `mode` string, optional `visible` bool), and `title` (object: optional `text` string, optional `visible` bool).

#### Scenario: Gauge metric defaults

- GIVEN a gauge metric configuration that omits the implementation's defaulted fields
- WHEN the provider compares or refreshes state
- THEN it SHALL normalize those defaults for semantic equality

#### Scenario: Gauge ES|QL round-trip

- GIVEN a gauge panel in ES|QL mode with `esql_metric.column = "revenue"` and `esql_metric.format_json` set to a number format
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the typed `esql_metric` block and `metric_json` SHALL be null

## ADDED Requirements

### Requirement: Typed partition chart legends (REQ-043)

Partition chart legends (treemap, mosaic, pie, waffle) SHALL expose `truncate_after_lines` as an `Int64` attribute in the Terraform schema. The API specifies `float32` for this field, but the value represents an integer count of lines before truncation. All partition chart legend schemas SHALL use `Int64` consistently.

#### Scenario: Legend truncate_after_lines accepts integer values

- GIVEN a treemap panel with `legend.truncate_after_lines = 5`
- WHEN the provider builds the API request
- THEN it SHALL encode the value as `5` in the API payload
- AND read-back SHALL preserve the integer value without drift

#### Scenario: Legend truncate_after_lines rejects fractional practitioner input

- GIVEN a mosaic panel with `legend.truncate_after_lines = 5.5`
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return an error diagnostic because the attribute type is `Int64`

### Requirement: Waffle value_display reuse (REQ-044)

Waffle `value_display` SHALL use the shared `getPartitionValueDisplaySchema()` helper, ensuring the same attribute names, types, and documentation text as treemap and mosaic `value_display`.

#### Scenario: Waffle value_display shape matches treemap

- GIVEN a waffle panel with `value_display = { mode = "percentage", percent_decimals = 2 }`
- WHEN Terraform validates the configuration
- THEN the validation SHALL pass and the attribute descriptions SHALL match those used by `treemap_config.value_display`

### Requirement: SLO overview constant definition site (REQ-045)

The `panelTypeSloOverview` constant SHALL be defined in `schema.go` alongside all other dashboard panel type constants. It SHALL NOT be defined in a panel-specific model file.

#### Scenario: Constant location

- GIVEN a code review of `schema.go`
- WHEN the reviewer searches for `panelTypeSloOverview`
- THEN the constant SHALL be found in `schema.go` together with `panelTypeSloAlerts`, `panelTypeSloBurnRate`, and other panel type constants
