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
- `drilldowns` (list of the shared `url_drilldown` block, max 100 entries, optional). URL drilldown `trigger` is fixed at `"on_open_panel_menu"` for this panel; because `AllowedTriggers` collapses to that single value, the Terraform nested schema omits the `trigger` attribute (same pattern as `slo_burn_rate_config.drilldowns`). Practitioners configure `url`, `label`, and optional `encode_url` / `open_in_new_tab`; the model layer SHALL write `trigger = "on_open_panel_menu"` on the API payload. `encode_url` and `open_in_new_tab` are optional booleans subject to REQ-009 null-preservation.

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

- GIVEN `slo_alerts_config.drilldowns = [{ url = "https://kibana/...", label = "investigate" }]` (no `trigger` attribute in Terraform because the schema omits it when only one trigger is allowed)
- WHEN create runs and the post-apply read returns the same drilldown
- THEN state SHALL contain that drilldown with `encode_url` and `open_in_new_tab` null per REQ-009

#### Scenario: Drilldown trigger fixed by model on write

- GIVEN `slo_alerts_config.drilldowns = [{ url = "https://kibana/...", label = "investigate" }]`
- WHEN create runs and the provider builds the dashboard API request
- THEN each emitted URL drilldown object in the panel payload SHALL include `trigger = "on_open_panel_menu"`

#### Scenario: config_json rejected for slo_alerts panel type

- GIVEN a panel with `type = "slo_alerts"` and `config_json` set
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `slo_alerts`

### Requirement: Discover session panel behavior (REQ-042)

When a panel entry sets `type = "discover_session"`, the resource SHALL accept a `discover_session_config` block and SHALL require that block to be present. The block SHALL expose two mutually exclusive sub-blocks mirroring the API's by-value/by-reference union; exactly one of `by_value` or `by_reference` SHALL be set.

The `discover_session_config` block SHALL accept the typed envelope attributes `title` (string), `description` (string), `hide_title` (bool), `hide_border` (bool), and an optional `drilldowns` list of the shared `url_drilldown` block (max 100 entries). URL drilldown `trigger` is fixed at `"on_open_panel_menu"`; the nested schema omits the Terraform `trigger` attribute when only that trigger is allowed, and the model layer SHALL write `trigger = "on_open_panel_menu"` on the API payload.

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
- `query` (required, the existing typed `query` block: `expression` (required string) and optional `language` with validator `kql` or `lucene`, matching `getFilterSimple()` / dashboard filter query shape)
- `data_source_json` (required string, JSON-encoded object conforming to the API `data_view_reference` or `data_view_spec` discriminator; the resource SHALL validate well-formed JSON at plan time and apply semantic JSON equality on read so key reorderings and Kibana-injected defaults do not produce diffs)
- `filters` (optional list of `object({ filter_json = string })`, max 100). The `filter_json` element shape and normalization SHALL match the dashboard-level `filters` shape defined by REQ-037 (`dashboard-filters`).

The `tab.esql` sub-block SHALL accept:

- `column_order`, `column_settings`, `sort`, `density`, `header_row_height`, `row_height` — same shapes and validators as in `tab.dsl`
- `data_source_json` (required string, JSON-encoded object conforming to the API `esqlDataSource` schema; same normalization semantics as for `tab.dsl.data_source_json`)

#### The `by_reference` sub-block

The `by_reference` sub-block SHALL accept:

- `time_range` — optional, same shape and inheritance semantics as for `by_value`
- `ref_id` — required string identifying the linked Discover session saved object
- `selected_tab_id` — optional string and computed when omitted; the resource SHALL preserve a user-supplied value and SHALL populate it from the API response otherwise. **Acceptance-test gap:** acceptance tests that link a legacy `search` saved-object fixture cannot assert `selected_tab_id` end-to-end when that saved object type does not return `selected_tab_id` from the API; newer `discover-session` saved-object fixtures (when available) will close this gap.
- `overrides` — optional `object` of typed scalars: `column_order`, `column_settings`, `sort`, `density`, `header_row_height`, `row_height`, `rows_per_page`, `sample_size` (same shapes and validators as their `tab.dsl` counterparts)

The `by_reference` sub-block SHALL NOT include a `references` or `references_json` attribute in v1. Empirical verification on Kibana **9.4.0** (`openspec/changes/add-new-panels/design.md`, “Open questions”): creating a dashboard via `POST /api/dashboards` with a `discover_session` panel whose `config` contains only `ref_id` (and required envelope fields such as `time_range`) succeeds **without** any client-side references; a top-level dashboard `references` property is **rejected** by the Dashboard API (400 — additional properties not allowed). If a future Kibana version changes this contract, a follow-on change MAY add `references_json` or equivalent additively.

#### Cross-cutting

The `discover_session_config` block SHALL conflict with all other typed panel config blocks and with `config_json`. When `type = "discover_session"`, `config_json` SHALL NOT be set on the same panel entry; if it is, the resource SHALL return an error diagnostic indicating `config_json` is not supported for `discover_session`.

On write, the resource SHALL build the `kbn-dashboard-panel-type-discover_session` API payload from the `discover_session_config` block, choosing the by-value or by-reference branch according to which sub-block is set and, within `by_value.tab`, choosing the DSL or ES|QL tab variant according to which sub-block is set. On read, the resource SHALL detect the API branch and populate the matching sub-block, leaving the other null.

REQ-009 null-preservation SHALL apply to all optional fields, drilldown defaults, and panel-level `time_range`.

#### Scenario: By-value DSL tab round-trip

- GIVEN a panel with `discover_session_config = { by_value = { tab = { dsl = { query = { expression = "host.name : \"web-01\"", language = "kql" }, data_source_json = jsonencode({ type = "data_view_reference", ref_id = "logs-*" }), column_order = ["@timestamp", "message"] } } } }`
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

#### Scenario: Drilldown trigger fixed by model on write

- GIVEN `discover_session_config.drilldowns = [{ url = "https://kibana/...", label = "investigate" }]`
- WHEN create runs and the provider builds the dashboard API request
- THEN each emitted URL drilldown object in the panel payload SHALL include `trigger = "on_open_panel_menu"`

#### Scenario: selected_tab_id computed when omitted

- GIVEN a `by_reference` panel without `selected_tab_id` in configuration
- WHEN create runs and the API returns `selected_tab_id = "tab-uuid-1"` in the response
- THEN state SHALL contain `selected_tab_id = "tab-uuid-1"` and a subsequent plan against the same configuration SHALL show no changes

#### Scenario: config_json rejected for discover_session panel type

- GIVEN a panel with `type = "discover_session"` and `config_json` set
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `discover_session`

## MODIFIED Requirements

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

The resource SHALL support top-level `panels`, section-contained `panels`, and `sections` in the order returned by the API and the order given in configuration when building requests. For panel reads, it SHALL distinguish sections from top-level panels and map each panel's `type`, `grid`, optional **`id`**, and configuration. For typed panel mappings, the resource SHALL seed from prior state or plan so that optional panel attributes omitted by Kibana on read can be preserved. When a panel is managed through `config_json` only, the resource SHALL preserve that JSON-centric representation and SHALL NOT populate typed configuration blocks from the API for that panel.

On write, practitioner-authored panel-level `config_json` SHALL be supported only for `markdown` and `vis` panel types; using practitioner-authored panel-level `config_json` with any other panel type, including `slo_burn_rate`, `slo_error_budget`, `esql_control`, `lens-dashboard-app`, `image`, `slo_alerts`, and `discover_session`, or omitting all panel configuration blocks, SHALL return an error diagnostic. Exception: when a panel's `config_json` value was populated by the provider during a prior read to preserve an unknown panel type (see "Unknown panel-type preservation" below), the provider SHALL re-emit that preserved payload on write without error, because the value originates from the API rather than from the practitioner. The `esql_control` panel type SHALL be managed exclusively through the typed `esql_control_config` block. The `lens-dashboard-app` panel type SHALL be managed exclusively through the typed `lens_dashboard_app_config` block.

`config_json` SHALL NOT be supported for `options_list_control` panels; the `options_list_control` panel type SHALL be managed exclusively through the typed `options_list_control_config` block; using `config_json` with `type = "options_list_control"` SHALL return an error diagnostic.

`config_json` SHALL NOT be supported for `synthetics_monitors` panels; the `synthetics_monitors` panel type SHALL be managed exclusively through the typed `synthetics_monitors_config` block; using `config_json` with `type = "synthetics_monitors"` SHALL return an error diagnostic.

`config_json` SHALL NOT be supported for `synthetics_stats_overview` panels; the write-path dispatcher SHALL return an error diagnostic if `config_json` is set on a panel with `type = "synthetics_stats_overview"`. The error message SHALL indicate that `config_json` is unsupported for the configured panel type.

`config_json` SHALL NOT be supported for `image` panels; the `image` panel type SHALL be managed exclusively through the typed `image_config` block; using `config_json` with `type = "image"` SHALL return an error diagnostic.

`config_json` SHALL NOT be supported for `slo_alerts` panels; the `slo_alerts` panel type SHALL be managed exclusively through the typed `slo_alerts_config` block; using `config_json` with `type = "slo_alerts"` SHALL return an error diagnostic.

`config_json` SHALL NOT be supported for `discover_session` panels; the `discover_session` panel type SHALL be managed exclusively through the typed `discover_session_config` block; using `config_json` with `type = "discover_session"` SHALL return an error diagnostic.

**Unknown panel-type preservation**: When the read path encounters a panel whose `type` does not match any typed configuration block, the resource SHALL preserve the panel's `id`, `grid`, `type`, and the panel's full raw API configuration payload in state. The preserved payload SHALL be stored in the panel's existing `config_json` attribute (which is `Optional: true, Computed: true`), reusing that attribute for storage rather than introducing a new private field. This is an intentional implementation decision: it avoids introducing a new unexposed attribute and leverages the existing semantic-equality normalization path already in place for `config_json`. The `config_json` value in state is provider-populated, not practitioner-authored. On subsequent writes, the resource SHALL re-emit the preserved payload verbatim through the `config_json` write codepath. If a configuration declares a panel whose `type` matches no typed block and no preserved payload exists in `config_json` state, the resource SHALL return the "unsupported panel type" error diagnostic, clarifying that the type is not yet supported and was not preserved from the API.

**Panel and section identity**: The Terraform attributes **`panels[].id`** and **`sections[].id`** SHALL align directly with the generated Kibana API field **`id`** for panels and sections.

#### Scenario: Panel id round-trip

- GIVEN a panel with `id = "panel-a"` in configuration
- WHEN create or update runs
- THEN the API request SHALL include the generated API `id` field (or equivalent panel identity) consistent with `panel-a` for that panel

#### Scenario: config_json rejected for options_list_control panel type

- GIVEN a panel with `type = "options_list_control"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `options_list_control`

#### Scenario: config_json rejected for synthetics_monitors panel type

- GIVEN a panel with `type = "synthetics_monitors"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `synthetics_monitors`

#### Scenario: config_json rejected for synthetics_stats_overview panel type

- GIVEN a panel with `type = "synthetics_stats_overview"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic indicating that `config_json` is not supported for the configured panel type

#### Scenario: config_json rejected for image panel type

- GIVEN a panel with `type = "image"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `image`

#### Scenario: config_json rejected for slo_alerts panel type

- GIVEN a panel with `type = "slo_alerts"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `slo_alerts`

#### Scenario: config_json rejected for discover_session panel type

- GIVEN a panel with `type = "discover_session"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `discover_session`

#### Scenario: Unknown panel type preserved on read

- GIVEN a dashboard managed by the resource that contains a panel whose `type` does not map to any typed configuration block in this resource (any panel type not yet modeled by the provider)
- WHEN refresh, import, or post-apply read runs
- THEN the resource SHALL store the panel's `id`, `grid`, `type`, and full raw API config payload in state without populating any typed config block, and a subsequent plan against unchanged configuration SHALL produce no diff

#### Scenario: Unknown panel type round-trips on write

- GIVEN state contains a panel with an unknown `type` and a preserved raw API payload from a prior read
- WHEN create or update runs and the user has not modified the panel
- THEN the provider SHALL re-emit the preserved payload verbatim in the API request body

#### Scenario: Practitioner cannot author unknown panel types

- GIVEN a Terraform configuration declaring `panels[].type = "<not-typed-and-no-prior-state>"` with no typed config block and no `config_json`
- WHEN validate or plan runs
- THEN the resource SHALL return an error diagnostic indicating the panel type is unsupported
