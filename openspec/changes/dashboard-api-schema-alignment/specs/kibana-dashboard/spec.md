# Delta Spec: Dashboard API schema alignment

Base spec: `openspec/specs/kibana-dashboard/spec.md`

This delta introduces **REQ-036** and modifies **REQ-007**, **REQ-008**, **REQ-009**, **REQ-010**, **REQ-013**, **REQ-018**, and **REQ-023** to reflect Terraform schema names and shapes aligned with the Kibana Dashboard API. It replaces portions of the **Schema** appendix in the base spec for the affected attributes.

---

## Schema appendix (REPLACE fragments)

### Top-level dashboard attributes (replaces the corresponding lines in the base schema sketch)

```hcl
resource "elasticstack_kibana_dashboard" "example" {
  id           = <computed, string>
  space_id     = <optional, computed, string>
  dashboard_id = <computed, string>

  title       = <required, string>
  description = <optional, string>

  time_range = <required, object({
    from = <required, string>
    to   = <required, string>
    mode = <optional, string> # absolute | relative; see REQ-009 for read-back preservation
  })>

  refresh_interval = <required, object({
    pause = <required, bool>
    value = <required, int64>
  })>

  query = <required, object({
    language = <required, string>
    # Exactly one of:
    text = <optional, string> # conflicts with json; KQL/Lucene string branch
    json = <optional, json string, normalized> # conflicts with text; object branch
  })>

  tags = <optional, list(string)>

  options = <optional, object({
    hide_panel_titles   = <optional, bool>
    use_margins         = <optional, bool>
    sync_colors         = <optional, bool>
    sync_tooltips       = <optional, bool>
    sync_cursor         = <optional, bool>
    auto_apply_filters  = <optional, bool>
    hide_panel_borders  = <optional, bool>
  })>

  # panels / sections: see REQ-010 delta — uid replaces id on panel and section objects
  # ...
}
```

### Panel / section identity (replaces `id` under `panels` and `sections` in the base schema sketch)

```hcl
panels = <optional, list(object({
  type = <required, string>
  grid = {
    x = <required, int64>
    y = <required, int64>
    w = <optional, int64>
    h = <optional, int64>
  }
  uid = <optional, computed, string> # API uid; replaces former attribute name `id`

  # ... panel config blocks unchanged except pie_chart and heatmap deltas below
})>

sections = <optional, list(object({
  title     = <required, string>
  uid       = <optional, computed, string> # API section uid; replaces former attribute name `id`
  collapsed = <optional, bool>
  grid      = { y = <required, int64> }
  panels    = <optional, list(...)> # same panel object shape as top-level panels
})>
```

### Pie chart (`pie_chart_config`) fragment

`dataset` and `legend` (JSON-normalized) are renamed to **`dataset_json`** and **`legend_json`**.

### Heatmap (`heatmap_config.legend`) fragment

`legend.visibility` SHALL be a string: **`visible`** or **`hidden`**, matching API `HeatmapLegendVisibility`.

---

## ADDED Requirements

### Requirement: Dashboard root schema API naming (REQ-036)

The resource SHALL expose dashboard-level time selection, refresh, and query as object-valued attributes whose names mirror the Kibana Dashboard API JSON: `time_range` (`from`, `to`, optional `mode`), `refresh_interval` (`pause`, `value`), and `query` (`language` with exactly one of `text` or `json` for the query union).

The resource SHALL expose dashboard `options` as an object-valued attribute with the API-aligned flags `auto_apply_filters` and `hide_panel_borders` in addition to the existing option fields.

#### Scenario: Query union uses text branch

- GIVEN `query = { language = "kuery" text = "http.response.status_code:200" }`
- WHEN the provider builds the create or update request body
- THEN it SHALL set the API query to the string branch of `query.query` and SHALL set `query.language` from `query.language`

#### Scenario: Query union uses json branch

- GIVEN `query = { language = "kuery" json = jsonencode({ ... }) }`
- WHEN the provider builds the create or update request
- THEN it SHALL set the API query to the object branch and SHALL reject configurations where both `text` and `json` are set

#### Scenario: Options include new flags

- GIVEN `options = { hide_panel_borders = true auto_apply_filters = false }`
- WHEN create or update runs
- THEN the provider SHALL include those fields in the API `options` object when known

---

## MODIFIED Requirements

### Requirement: Create and update request mapping (REQ-007)

On create and update, the resource SHALL map Terraform state to the dashboard API request body using `title`, `description`, **`time_range`**, **`refresh_interval`**, **`query`**, tags, **extended `options`**, panels, and sections when those values are known. `access_control` SHALL be sent on create when known. The current regenerated Kibana `PUT /dashboards/{id}` request body does not expose `access_control`, so updates SHALL preserve prior `access_control` state but SHALL NOT claim to mutate it through the dashboard update request until the API surface supports that field.

**Query mapping** SHALL send `query.text` as the string branch of the API union and `query.json` as the object branch of the API union. **`query_language` / `query_text` / `query_json` root attributes are removed** in favor of the nested `query` block.

#### Scenario: Post-apply authoritative read (unchanged intent)

- GIVEN create or update succeeds
- WHEN the provider finalizes state
- THEN it SHALL re-read the dashboard and SHALL fail if the dashboard cannot be retrieved

---

### Requirement: Read behavior and missing-resource handling (REQ-008)

On refresh, the resource SHALL parse the composite `id`, read the dashboard from Kibana, and repopulate state from the API response. If Kibana returns not found, the resource SHALL remove itself from Terraform state. When a dashboard is found, the resource SHALL map title, description, **nested `time_range`**, **nested `refresh_interval`**, **nested `query`**, tags, **extended `options`**, access control, top-level panels, and sections back into state.

#### Scenario: Read maps nested query and time_range

- GIVEN a successful refresh after create with `query { language = "kuery" text = "foo" }` and `time_range { from = "now-7d" to = "now" }`
- WHEN state is repopulated from the GET response
- THEN the resource SHALL set `query.language`, `query.text`, and `time_range.from` / `time_range.to` from the API payload

---

### Requirement: State preservation for fields Kibana omits or defaults (REQ-009)

When Kibana omits or defaults fields on read, the resource SHALL preserve prior Terraform intent to avoid inconsistent results and spurious drift. The resource currently preserves the prior **`time_range.mode`** value already held in state or plan instead of overwriting it from read-back, because the implementation does not currently map the API's optional `time_range.mode` field into state from GET responses. (Legacy attribute name `time_range_mode` at root is removed in favor of `time_range.mode`.)

When the GET dashboard API omits `access_control`, the resource SHALL preserve the prior `access_control` value instead of clearing it. When the options block was omitted in Terraform and Kibana materializes only the default dashboard options, the resource SHALL keep the `options` block null in state. For this preservation and drift-suppression behavior, "default dashboard options" SHALL be interpreted consistently with the implementation's **`isDashboardOptionsDefaultSet`** helper for **all** modeled option fields, including **`auto_apply_filters = true`** and **`hide_panel_borders = false`** as the API defaults when those fields are present or materialized by Kibana.

The resource models only the currently supported Terraform subset of dashboard fields. Fields present in the Kibana dashboard API but not modeled by this resource, including `filters`, `pinned_panels`, and `project_routing`, are outside this resource contract and are not guaranteed to round-trip through Terraform updates.

#### Scenario: time_range.mode preservation when GET omits mode

- GIVEN prior state or plan holds `time_range.mode = "relative"`
- WHEN GET does not return a usable `time_range.mode` for state mapping
- THEN the resource SHALL preserve the prior `time_range.mode` value instead of clearing it

---

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

The resource SHALL support top-level `panels`, section-contained `panels`, and `sections` in the order returned by the API and the order given in configuration when building requests. For panel reads, it SHALL distinguish sections from top-level panels and map each panel's `type`, `grid`, optional **`uid`**, and configuration. For typed panel mappings, the resource SHALL seed from prior state or plan so that optional panel attributes omitted by Kibana on read can be preserved. When a panel is managed through `config_json` only, the resource SHALL preserve that JSON-centric representation and SHALL NOT populate typed configuration blocks from the API for that panel. On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, including `slo_burn_rate`, `slo_error_budget`, and `esql_control`, or omitting all panel configuration blocks, SHALL return an error diagnostic. The `esql_control` panel type SHALL be managed exclusively through the typed `esql_control_config` block.

**Panel and section identity**: The Terraform attributes **`panels[].uid`** and **`sections[].uid`** replace **`panels[].id`** and **`sections[].id`** to match API `uid`.

#### Scenario: Panel uid round-trip

- GIVEN a panel with `uid = "panel-a"` in configuration
- WHEN create or update runs
- THEN the API request SHALL include `uid` (or equivalent panel identity) consistent with `panel-a` for that panel

---

### Requirement: XY chart panel behavior (REQ-013)

For XY chart Lens panels, the resource SHALL require `axis`, `decorations`, `fitting`, `legend`, `query`, and at least one `layers` entry. Each layer SHALL represent either a data layer or a reference-line layer, not both. When the provider builds a typed Lens XY panel, it SHALL send the fixed Lens panel time range used by the implementation for typed Lens panels (see **`KbnDashboardPanelLensConfig0.time_range`** in the API). The Terraform schema for typed XY panels does not expose that Lens-level `time_range` as a separate configurable attribute; it remains implementation-defined for typed panels.

#### Scenario: Typed XY write includes Lens time_range

- GIVEN a typed `xy_chart_config` panel on create or update
- WHEN the provider builds the Lens panel payload
- THEN it SHALL set `time_range` on `KbnDashboardPanelLensConfig0` to the implementation’s fixed window for typed converters

---

### Requirement: Heatmap panel behavior (REQ-018)

For heatmap Lens panels, the resource SHALL require `dataset_json`, `axes`, `cells`, `legend`, `metric_json`, and `x_axis_json`. **`legend.visibility` SHALL use the string values `visible` or `hidden`**, matching the API enum. The resource SHALL treat the panel as non-ES|QL when a real `query` is present, and in that mode `query` SHALL be required. It SHALL treat the panel as ES|QL when `query` is omitted or empty by the implementation's mode test. Heatmap metric normalization SHALL use the same metric-default behavior shared with the tagcloud implementation.

#### Scenario: Heatmap legend visibility enum

- GIVEN `heatmap_config.legend.visibility = "hidden"`
- WHEN the provider builds the API request
- THEN it SHALL set API heatmap legend visibility to the `hidden` enum value

---

### Requirement: Pie chart panel behavior (REQ-023)

For pie Lens panels, the resource SHALL require at least one `metrics` entry and MAY accept `group_by`. It SHALL select the non-ES|QL branch when `query` is present and the ES|QL branch otherwise. When Kibana omits `ignore_global_filters` or `sampling` on read, the provider SHALL treat their default values as `false` and `1.0` respectively. Pie metric and group-by semantic equality SHALL normalize the implementation's pie metric defaults and Lens group-by defaults.

**JSON attributes**: The resource SHALL use **`dataset_json`** and **`legend_json`** (normalized JSON strings) for pie chart dataset and legend objects, replacing the former attribute names `dataset` and `legend`.

#### Scenario: Pie chart uses dataset_json

- GIVEN `pie_chart_config` with `dataset_json` set to a normalized JSON string for the pie dataset
- WHEN the provider builds the Lens attributes
- THEN it SHALL decode `dataset_json` into the API pie dataset shape

---

## Notes

- **Breaking change**: No compatibility shim or state migration; practitioners update module source to the new attribute paths.
- **Lens `time_range`**: REQ-025 continues to govern `config_json` Lens panels (no `lensPanelTimeRange()` injection). REQ-013 and REQ-007 deltas clarify typed Lens behavior vs dashboard-level `time_range`.
