## ADDED Requirements

### Requirement: Image panel behavior (REQ-040)

When a panel entry sets `type = "image"`, the resource SHALL accept an `image_config` block and SHALL require that block to be present. The block SHALL require `src`, a nested object with mutually exclusive `file = object({ file_id = string })` and `url = object({ url = string })` sub-blocks; exactly one of `src.file` or `src.url` SHALL be set. `file_id` (when `file` is used) and `url` (when `url` is used) SHALL be required non-empty strings.

The `image_config` block SHALL accept the optional attributes `alt_text` (string), `object_fit` (string, validator: one of `"fill"`, `"contain"`, `"cover"`, `"none"`; no Terraform-side default), `background_color` (string), `title` (string), `description` (string), `hide_title` (bool), and `hide_border` (bool). It SHALL also accept an optional `drilldowns` list (max 100 entries) where each entry contains exactly one of `dashboard_drilldown` or `url_drilldown` sub-blocks:

- `dashboard_drilldown = object({ dashboard_id = string (required), label = string (required), trigger = string (validator: must be "on_click_image"), use_filters = optional bool, use_time_range = optional bool, open_in_new_tab = optional bool })`
- `url_drilldown = object({ url = string (required), label = string (required), trigger = string (validator: one of "on_click_image", "on_open_panel_menu"), encode_url = optional bool, open_in_new_tab = optional bool })`

The `image_config` block SHALL conflict with all other typed panel config blocks and with `config_json`. When `type = "image"`, `config_json` SHALL NOT be set on the same panel entry; if it is, the resource SHALL return an error diagnostic indicating `config_json` is not supported for `image`.

On write, the resource SHALL build the `kbn-dashboard-panel-type-image` API payload from the `image_config` block. On read, the resource SHALL repopulate `image_config` from the API response and SHALL apply REQ-009 null-preservation to optional fields and to drilldown sub-fields with API defaults (`use_filters`, `use_time_range`, `open_in_new_tab`, `encode_url`, `object_fit`).

#### Scenario: Image panel with file source round-trip

- GIVEN a panel with `type = "image"` and `image_config = { src = { file = { file_id = "img-1" } }, alt_text = "diagram", object_fit = "cover" }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the same `image_config` shape, `src.url` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: Image panel with url source round-trip

- GIVEN a panel with `type = "image"` and `image_config = { src = { url = { url = "https://example.com/logo.png" } } }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the same `image_config` shape with `src.file` null

#### Scenario: src requires exactly one branch

- GIVEN an `image_config` with both `src.file` and `src.url` set, or with neither set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one of `src.file` or `src.url` must be set

#### Scenario: Invalid object_fit rejected

- GIVEN an `image_config` with `object_fit = "stretch"`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the value must be `"fill"`, `"contain"`, `"cover"`, or `"none"`

#### Scenario: Drilldown discriminator validation

- GIVEN a `drilldowns` entry with both `dashboard_drilldown` and `url_drilldown` set, or with neither set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one drilldown sub-block must be set

#### Scenario: Drilldown trigger validation

- GIVEN a `dashboard_drilldown` with `trigger = "on_open_panel_menu"` (an `image` dashboard drilldown only supports `on_click_image`)
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the trigger must be `"on_click_image"`

#### Scenario: config_json rejected for image panel type

- GIVEN a panel with `type = "image"` and `config_json` set
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `image`

#### Scenario: Drilldown defaults null-preserved on import

- GIVEN an existing image panel with a dashboard drilldown whose `use_filters`, `use_time_range`, and `open_in_new_tab` come from Kibana defaults
- WHEN the resource imports the dashboard
- THEN those drilldown attributes SHALL remain null in state and a subsequent plan against a configuration that omits them SHALL show no changes

### Requirement: SLO alerts panel behavior (REQ-041)

When a panel entry sets `type = "slo_alerts"`, the resource SHALL accept an `slo_alerts_config` block and SHALL require that block to be present. The block SHALL accept the following attributes:

- `slos` (list of `object({ slo_id = string (required), slo_instance_id = optional string })`) — **required**, max 100 entries. The resource SHALL validate at plan time that `slos` contains at least one entry (`len(slos) > 0`); empty `slos` SHALL return an error diagnostic. On read, `slo_instance_id` SHALL apply REQ-009 null-preservation: when prior state had it null, the resource SHALL keep it null even if Kibana returns its server-side default `"*"`.
- `title` (string, optional)
- `description` (string, optional)
- `hide_title` (bool, optional)
- `hide_border` (bool, optional)
- `drilldowns` (list of the shared `url_drilldown` block, max 100 entries, optional). The `url_drilldown.trigger` SHALL be `"on_open_panel_menu"`. `encode_url` and `open_in_new_tab` are optional booleans subject to REQ-009 null-preservation.

The shared `url_drilldown` block referenced here SHALL be the same nested-block schema consumed by `slo_burn_rate_config.drilldowns` and `slo_overview_config.drilldowns` (Go-level consolidation; behavior unchanged for existing SLO panels).

The `slo_alerts_config` block SHALL conflict with all other typed panel config blocks and with `config_json`. When `type = "slo_alerts"`, `config_json` SHALL NOT be set on the same panel entry; if it is, the resource SHALL return an error diagnostic indicating `config_json` is not supported for `slo_alerts`.

On write, the resource SHALL build the `kbn-dashboard-panel-type-slo_alerts` API payload from the `slo_alerts_config` block. On read, the resource SHALL repopulate `slo_alerts_config` from the API response.

#### Scenario: Minimal slo_alerts panel round-trip

- GIVEN a panel with `type = "slo_alerts"` and `slo_alerts_config = { slos = [{ slo_id = "slo-1" }] }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the same single SLO entry with `slo_instance_id` null and a subsequent plan SHALL show no changes

#### Scenario: Empty slos rejected at plan time

- GIVEN an `slo_alerts_config` with `slos = []`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating `slos` must contain at least one entry

#### Scenario: slo_instance_id null preservation

- GIVEN an existing slo_alerts panel where Kibana stored `slo_instance_id = "*"` server-side and prior state had it null
- WHEN the resource imports or refreshes the dashboard
- THEN `slo_instance_id` SHALL remain null in state and a subsequent plan against a configuration that omits it SHALL show no changes

#### Scenario: Drilldown round-trip

- GIVEN `slo_alerts_config.drilldowns = [{ url_drilldown = { url = "https://kibana/...", label = "investigate", trigger = "on_open_panel_menu" } }]`
- WHEN create runs and the post-apply read returns the same drilldown
- THEN state SHALL contain that drilldown with `encode_url` and `open_in_new_tab` null per REQ-009

#### Scenario: Invalid drilldown trigger

- GIVEN a drilldown with `trigger = "on_click_image"`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the trigger must be `"on_open_panel_menu"`

#### Scenario: config_json rejected for slo_alerts panel type

- GIVEN a panel with `type = "slo_alerts"` and `config_json` set
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `slo_alerts`

### Requirement: Discover session panel behavior (REQ-042)

When a panel entry sets `type = "discover_session"`, the resource SHALL accept a `discover_session_config` block and SHALL require that block to be present. The block SHALL expose two mutually exclusive sub-blocks mirroring the API's by-value/by-reference union; exactly one of `by_value` or `by_reference` SHALL be set.

The `discover_session_config` block SHALL accept the typed envelope attributes `title` (string), `description` (string), `hide_title` (bool), `hide_border` (bool), and an optional `drilldowns` list of the shared `url_drilldown` block (max 100 entries). The `url_drilldown.trigger` SHALL be `"on_open_panel_menu"`.

#### The `by_value` sub-block

The `by_value` sub-block SHALL accept:

- `time_range` — optional, shape `object({ from = string, to = string, mode? = string })`. When null, the model layer SHALL materialize the dashboard-root `time_range` into the API payload at write time so the API request always carries a value. Read-back SHALL apply REQ-009 null-preservation: when prior state had `time_range` null and the API echoes the dashboard-inherited value, state SHALL remain null.
- `tab` — required `object`. The Discover API restricts inline tab configuration to one entry today; the resource exposes it as a single object. A future `tabs = list(...)` shape MAY be added additively if Kibana lifts the cardinality limit. The `tab` object SHALL contain two mutually exclusive sub-blocks `dsl` and `esql`; exactly one SHALL be set.

The `tab.dsl` sub-block SHALL accept:

- `column_order` (optional list of strings, max 100)
- `column_settings` (optional map of `object({ width = optional number })`)
- `sort` (optional list of `object({ name = string (required), direction = string (required, enum "asc"|"desc") })`, max 100)
- `density` (optional string, enum `"compact"`/`"expanded"`/`"normal"`)
- `header_row_height` (optional string, validator: a decimal integer in `"1".."5"` or the literal `"auto"`)
- `row_height` (optional string, validator: a decimal integer in `"1".."20"` or the literal `"auto"`)
- `rows_per_page` (optional number, `1..10000`)
- `sample_size` (optional number, `10..10000`)
- `view_mode` (optional string, enum `"documents"`/`"patterns"`/`"aggregated"`)
- `query` (required, the existing typed `query` block: `object({ language = string, text? = string, json? = string })`)
- `data_source_json` (required string, JSON-encoded object conforming to the API `data_view_reference` or `data_view_spec` discriminator; the resource SHALL validate well-formed JSON at plan time and apply semantic JSON equality on read so key reorderings and Kibana-injected defaults do not produce diffs)
- `filters` (optional list of `object({ filter_json = string })`, max 100). The `filter_json` element shape and normalization SHALL match the dashboard-level `filters` shape defined by REQ-037 (`dashboard-filters`).

The `tab.esql` sub-block SHALL accept:

- `column_order`, `column_settings`, `sort`, `density`, `header_row_height`, `row_height` — same shapes and validators as in `tab.dsl`
- `data_source_json` (required string, JSON-encoded object conforming to the API `esqlDataSource` schema; same normalization semantics as for `tab.dsl.data_source_json`)

#### The `by_reference` sub-block

The `by_reference` sub-block SHALL accept:

- `time_range` — optional, same shape and inheritance semantics as for `by_value`
- `ref_id` — required string identifying the linked Discover session saved object
- `selected_tab_id` — optional string and computed when omitted; the resource SHALL preserve a user-supplied value and SHALL populate it from the API response otherwise
- `overrides` — optional `object` of typed scalars: `column_order`, `column_settings`, `sort`, `density`, `header_row_height`, `row_height`, `rows_per_page`, `sample_size` (same shapes and validators as their `tab.dsl` counterparts)

The `by_reference` sub-block SHALL NOT include a `references` attribute in v1. If subsequent verification shows the Dashboard API requires client-side `references` for this panel, a follow-on change SHALL add `references_json` additively before archival of this change.

#### Cross-cutting

The `discover_session_config` block SHALL conflict with all other typed panel config blocks and with `config_json`. When `type = "discover_session"`, `config_json` SHALL NOT be set on the same panel entry; if it is, the resource SHALL return an error diagnostic indicating `config_json` is not supported for `discover_session`.

On write, the resource SHALL build the `kbn-dashboard-panel-type-discover_session` API payload from the `discover_session_config` block, choosing the by-value or by-reference branch according to which sub-block is set and, within `by_value.tab`, choosing the DSL or ES|QL tab variant according to which sub-block is set. On read, the resource SHALL detect the API branch and populate the matching sub-block, leaving the other null.

REQ-009 null-preservation SHALL apply to all optional fields, drilldown defaults, and panel-level `time_range`.

#### Scenario: By-value DSL tab round-trip

- GIVEN a panel with `discover_session_config = { by_value = { tab = { dsl = { query = { language = "kuery", text = "host.name : \"web-01\"" }, data_source_json = jsonencode({ type = "data_view_reference", id = "logs-*" }), column_order = ["@timestamp", "message"] } } } }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the same `by_value.tab.dsl` shape, `by_value.tab.esql` SHALL be null, `by_reference` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: By-value ES|QL tab round-trip

- GIVEN a panel with `discover_session_config = { by_value = { tab = { esql = { data_source_json = jsonencode({ type = "esql", query = "FROM logs-*" }) } } } }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain `by_value.tab.esql` populated and `by_value.tab.dsl` null

#### Scenario: By-reference panel round-trip

- GIVEN an existing Discover session saved object with id `discover-1` and a panel with `discover_session_config = { by_reference = { ref_id = "discover-1", title = "saved errors" } }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the same `by_reference` shape with `selected_tab_id` populated from the API response and `by_value` SHALL be null

#### Scenario: Exactly one of by_value or by_reference

- GIVEN a panel with both `by_value` and `by_reference` set, or with neither set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one of `by_value` or `by_reference` must be set

#### Scenario: Exactly one of tab.dsl or tab.esql

- GIVEN `by_value.tab` with both `dsl` and `esql` set, or with neither set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one of `tab.dsl` or `tab.esql` must be set

#### Scenario: Panel-level time_range inherits from dashboard

- GIVEN a panel with `discover_session_config.by_value.time_range = null` and the dashboard root `time_range = { from = "now-15m", to = "now" }`
- WHEN create runs
- THEN the API request SHALL include `time_range = { from = "now-15m", to = "now" }` on the panel payload
- AND the post-apply read SHALL preserve `discover_session_config.by_value.time_range` as null in state

#### Scenario: Filter JSON shape matches dashboard-level filters

- GIVEN a `tab.dsl.filters` entry whose `filter_json` value semantically equals a dashboard-level `filters` entry's `filter_json` value
- WHEN refresh runs
- THEN both filters SHALL round-trip with the same semantic JSON equality and no diff SHALL be produced for either

#### Scenario: Invalid row_height rejected

- GIVEN `tab.dsl.row_height = "25"` (above the API maximum of 20)
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the value must be a decimal integer in `"1".."20"` or the literal `"auto"`

#### Scenario: data_source_json normalization avoids spurious diffs

- GIVEN a `tab.dsl.data_source_json` value whose key ordering or whitespace differs from the API response
- WHEN refresh runs
- THEN the provider SHALL not produce a diff for that field

#### Scenario: Drilldown trigger validation

- GIVEN a `drilldowns` entry with `url_drilldown.trigger = "on_click_image"`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the trigger must be `"on_open_panel_menu"`

#### Scenario: selected_tab_id computed when omitted

- GIVEN a `by_reference` panel without `selected_tab_id` in configuration
- WHEN create runs and the API returns `selected_tab_id = "tab-uuid-1"` in the response
- THEN state SHALL contain `selected_tab_id = "tab-uuid-1"` and a subsequent plan against the same configuration SHALL show no changes

#### Scenario: config_json rejected for discover_session panel type

- GIVEN a panel with `type = "discover_session"` and `config_json` set
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `discover_session`
