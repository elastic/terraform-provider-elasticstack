## MODIFIED Requirements

### Requirement: XY chart panel behavior and typed `vis` `time_range` (REQ-013)

For **typed** `vis` panels (those built through the provider's typed `*_config` blocks and the shared typed visualization write path, not panels managed solely through raw `config_json`), the resource SHALL expose `time_range` as an optional flat sibling attribute on every typed Lens chart block (`xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, `waffle_config`). The attribute SHALL match the dashboard-level `time_range` shape: required `from` (string), required `to` (string), and optional `mode` enum (`absolute` | `relative`).

When the chart-level `time_range` is null in configuration and state, the provider SHALL inherit the dashboard-level `time_range` when assembling the visualization payload, copying the dashboard-level `from`, `to`, and `mode` into the API request. The hardcoded `lensPanelTimeRange()` window (`now-15m..now`) SHALL NOT be used; it is removed in favor of inheritance.

When the chart-level `time_range` is set in configuration, the provider SHALL pass the configured values to the API verbatim, overriding the dashboard-level value for that panel only.

For XY chart `vis` panels specifically, the resource SHALL require `axis`, `decorations`, `fitting`, `legend`, `query`, and at least one `layers` entry. The axis object SHALL use `x`, optional primary `y`, and optional `secondary_y`; `axis.x.domain_json` SHALL represent the X-axis domain, and each configured Y axis SHALL require `domain_json`. Each layer SHALL represent either a data layer or a reference-line layer, not both.

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

## ADDED Requirements

### Requirement: Lens chart presentation fields on typed `vis` panels (REQ-037)

For every typed Lens chart block on `vis` panels (`xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, `waffle_config`), the resource SHALL expose the following optional flat-sibling attributes that mirror the corresponding fields on the Kibana chart-root API schemas:

- `hide_title` (bool): when set, the API payload SHALL include `hide_title` on the chart root; when null in state, the payload SHALL omit it.
- `hide_border` (bool): when set, the API payload SHALL include `hide_border` on the chart root; when null in state, the payload SHALL omit it.
- `references_json` (normalized JSON string): when set, the API payload SHALL include `references` on the chart root as the parsed JSON array (`kbn-content-management-utils-referenceSchema[]`); when null in state, the payload SHALL omit it. Read-back SHALL normalize the returned `references` array into the canonical JSON form used by the resource.
- `drilldowns` (typed list of variant sub-blocks per REQ-038): when set, the API payload SHALL include `drilldowns` on the chart root as a typed array conforming to the API discriminated union; when null in state, the payload SHALL omit it.

On read-back, the provider SHALL populate each attribute from the API response when present, and SHALL preserve null in state when the API omits the field (consistent with REQ-009 null-preservation semantics).

#### Scenario: hide_title round-trip

- GIVEN a typed Lens chart panel with `hide_title = true` in configuration
- WHEN the provider applies the configuration and reads it back
- THEN the API payload SHALL include `hide_title: true` on the chart root
- AND state SHALL show `hide_title = true`

#### Scenario: hide_title null-preservation on read

- GIVEN a typed Lens chart panel whose prior state has `hide_border = null`
- AND the Kibana API response omits `hide_border` on the chart root
- WHEN the provider reads the panel
- THEN state SHALL preserve `hide_border = null`

#### Scenario: references_json round-trip

- GIVEN a typed Lens chart panel with `references_json = jsonencode([{ name = "foo", type = "index-pattern", id = "abc" }])` in configuration
- WHEN the provider applies the configuration and reads it back
- THEN the API payload SHALL include the parsed `references` array on the chart root
- AND state SHALL show the normalized JSON form

### Requirement: Chart-level `time_range` null-preservation and inheritance from dashboard (REQ-038)

The resource SHALL preserve practitioner intent for the chart-level `time_range` block on every typed Lens chart panel using the same null-preservation pattern as REQ-009 for `time_range.mode`.

When prior state has `panel.<chart>_config.time_range = null` AND the API-returned chart-level `time_range` equals the dashboard-level `time_range` (compared by literal `from`, `to`, and `mode` string equality, treating both nulls as equal), the provider SHALL preserve null in state. Otherwise, the provider SHALL populate state with the API-returned chart-level `time_range`.

The chart-level `time_range.mode` attribute SHALL follow the same null-preservation rule as the dashboard-level `time_range.mode` in REQ-009: when prior state has `mode = null` and the API response omits or returns no usable mode, state SHALL preserve null rather than overwriting with a default.

#### Scenario: Chart time_range null-preserved when equal to dashboard

- GIVEN a typed Lens chart panel whose prior state has `time_range = null`
- AND the dashboard-level `time_range` is `{ from = "now-7d", to = "now" }`
- AND the Kibana API response returns `time_range = { from = "now-7d", to = "now" }` on that chart root
- WHEN the provider reads the panel
- THEN state SHALL preserve `time_range = null` on the chart panel

#### Scenario: Chart time_range populated when not equal to dashboard

- GIVEN a typed Lens chart panel whose prior state has `time_range = null`
- AND the dashboard-level `time_range` is `{ from = "now-7d", to = "now" }`
- AND the Kibana API response returns `time_range = { from = "now-30d", to = "now-1d" }` on that chart root
- WHEN the provider reads the panel
- THEN state SHALL populate `time_range = { from = "now-30d", to = "now-1d" }` on the chart panel

#### Scenario: Chart time_range mode null-preservation

- GIVEN a typed Lens chart panel whose prior state has `time_range = { from = "now-7d", to = "now", mode = null }`
- AND the Kibana API response omits `mode` on the chart-root `time_range`
- WHEN the provider reads the panel
- THEN state SHALL preserve `time_range.mode = null`

### Requirement: Drilldown structured list and per-variant validation (REQ-039)

When a typed Lens chart panel includes the `drilldowns` attribute, each list item SHALL be an object containing three mutually-exclusive optional sub-blocks modeling the API discriminated union: `dashboard_drilldown`, `discover_drilldown`, and `url_drilldown`. Each list item SHALL set exactly one variant sub-block; setting zero or multiple variants SHALL produce a plan-time validation error that identifies the offending list item and lists the allowable variants.

The `dashboard_drilldown` sub-block SHALL expose: required `dashboard_id` (string), required `label` (string), computed `trigger` (string, always `"on_apply_filter"`), optional `use_filters` (bool, default `true`), optional `use_time_range` (bool, default `true`), and optional `open_in_new_tab` (bool, default `false`).

The `discover_drilldown` sub-block SHALL expose: required `label` (string), computed `trigger` (string, always `"on_apply_filter"`), and optional `open_in_new_tab` (bool, default `true`).

The `url_drilldown` sub-block SHALL expose: required `url` (string), required `label` (string), required `trigger` (string) validated against the four API-allowed values (`on_click_row`, `on_click_value`, `on_open_panel_menu`, `on_select_range`), optional `encode_url` (bool, default `true`), and optional `open_in_new_tab` (bool, default `true`).

The provider SHALL implement inter-variant exclusivity using the conditional validators in `internal/utils/validators/conditional.go` (sibling-path expression-based). Each variant sub-block SHALL be decorated with sibling-relative forbidden conditions ensuring the other two variants are unset when this one is set.

The API write path SHALL serialize each list item according to the variant set: `dashboard_drilldown` items SHALL map to the API `dashboard_drilldown` object including `type: "dashboard_drilldown"`; `discover_drilldown` items SHALL map to the API `discover_drilldown` object including `type: "discover_drilldown"`; `url_drilldown` items SHALL map to the API `url_drilldown` object including `type: "url_drilldown"`. On read-back the provider SHALL detect the variant from the API `type` discriminator and populate the corresponding sub-block.

#### Scenario: url_drilldown with valid trigger round-trip

- GIVEN a typed Lens chart panel with a single `drilldowns` entry whose `url_drilldown` is set to `{ url = "https://example.com", label = "Open", trigger = "on_click_value" }`
- WHEN the provider applies the configuration and reads it back
- THEN the API payload SHALL include a `drilldowns` array of one element with `type = "url_drilldown"`, `url = "https://example.com"`, `label = "Open"`, `trigger = "on_click_value"`
- AND state SHALL show the same configuration with `url_drilldown.trigger = "on_click_value"`

#### Scenario: dashboard_drilldown computed trigger

- GIVEN a typed Lens chart panel with a `drilldowns` entry whose `dashboard_drilldown` is set to `{ dashboard_id = "abc", label = "Drill" }` (trigger not set in config)
- WHEN the provider applies the configuration and reads it back
- THEN the API payload SHALL include `trigger = "on_apply_filter"` for that drilldown
- AND state SHALL show `dashboard_drilldown.trigger = "on_apply_filter"` as the computed value

#### Scenario: url_drilldown trigger invalid value rejected at plan

- GIVEN a typed Lens chart panel with a `drilldowns` entry whose `url_drilldown.trigger = "on_invalid"`
- WHEN Terraform validates the configuration at plan time
- THEN validation SHALL produce an error attributed to that list item indicating the trigger value is not one of the allowed values (`on_click_row`, `on_click_value`, `on_open_panel_menu`, `on_select_range`)

#### Scenario: Multiple drilldown variants set on one list item rejected at plan

- GIVEN a typed Lens chart panel with a `drilldowns` list item that sets both `dashboard_drilldown` and `url_drilldown`
- WHEN Terraform validates the configuration at plan time
- THEN validation SHALL produce an error attributed to that list item indicating only one variant sub-block may be set

#### Scenario: No drilldown variant set on a list item rejected at plan

- GIVEN a typed Lens chart panel with a `drilldowns` list item that sets none of `dashboard_drilldown`, `discover_drilldown`, or `url_drilldown`
- WHEN Terraform validates the configuration at plan time
- THEN validation SHALL produce an error attributed to that list item indicating exactly one variant sub-block must be set

#### Scenario: Setting computed trigger in configuration rejected

- GIVEN a typed Lens chart panel with a `dashboard_drilldown` entry that explicitly sets `trigger = "on_apply_filter"` in configuration
- WHEN Terraform validates the configuration at plan time
- THEN validation SHALL produce an error indicating `trigger` is a computed-only attribute on the `dashboard_drilldown` and `discover_drilldown` variants
