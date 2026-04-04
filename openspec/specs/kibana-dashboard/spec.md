# `elasticstack_kibana_dashboard` — Schema and Functional Requirements

Resource implementation: `internal/kibana/dashboard`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_dashboard` resource, including Kibana Dashboard API usage, composite identity and import, provider-level Kibana OpenAPI client usage, replacement-vs-update behavior, query and panel mapping, section handling, and drift-resistant state reconciliation for dashboard options and panel defaults.

## Schema

```hcl
resource "elasticstack_kibana_dashboard" "example" {
  id           = <computed, string> # canonical state id: "<space_id>/<dashboard_id>"; UseStateForUnknown
  space_id     = <optional, computed, string> # default "default"; RequiresReplace
  dashboard_id = <optional, computed, string> # Kibana saved object id; RequiresReplace; UseStateForUnknown

  title                  = <required, string>
  description            = <optional, string>
  time_from              = <required, string>
  time_to                = <required, string>
  time_range_mode        = <optional, string> # one of: absolute | relative
  refresh_interval_pause = <required, bool>
  refresh_interval_value = <required, int64>
  query_language         = <required, string>
  query_text             = <optional, string> # conflicts with query_json
  query_json             = <optional, json string, normalized> # conflicts with query_text
  tags                   = <optional, list(string)>

  options = <optional, object({
    hide_panel_titles = <optional, bool>
    use_margins       = <optional, bool>
    sync_colors       = <optional, bool>
    sync_tooltips     = <optional, bool>
    sync_cursor       = <optional, bool>
  })>

  panels = <optional, list(object({
    type = <required, string>
    grid = {
      x = <required, int64>
      y = <required, int64>
      w = <optional, int64>
      h = <optional, int64>
    }
    id = <optional, computed, string> # UseNonNullStateForUnknown

    markdown_config = <optional, object({
      content     = <optional, string>
      description = <optional, string>
      hide_title  = <optional, bool>
      title       = <optional, string>
    })> # only with type = "markdown"; conflicts with all other config blocks

    xy_chart_config = <optional, object({
      title       = <optional, string>
      description = <optional, string>
      axis        = <required, object({ x = <optional, object(...)>, left = <optional, object(...)>, right = <optional, object(...)> })>
      decorations = <required, object(...)>
      fitting     = <required, object({ type = <required, string>, dotted = <optional, bool>, end_value = <optional, string> })>
      layers      = <required, list(object({ type = <required, string>, data_layer = <optional, object(...)>, reference_line_layer = <optional, object(...)> }))> # at least 1
      legend      = <required, object(...)>
      query       = <required, object({ language = <optional, string>, query = <required, string> })>
      filters     = <optional, list(object({ filter_json = <required, json string, normalized> }))>
    })> # only with type = "lens"

    treemap_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      dataset_json          = <required, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })> # required for non-ES|QL mode
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      group_by_json         = <required, json string with defaults>
      metrics_json          = <required, json string with defaults>
      label_position        = <optional, string>
      legend                = <required, object(...)>
      value_display         = <optional, object({ mode = <required, string>, percent_decimals = <optional, float64> })>
    })> # only with type = "lens"

    mosaic_config = <optional, object({
      title                   = <optional, string>
      description             = <optional, string>
      dataset_json            = <required, json string, normalized>
      query                   = <optional, object({ language = <optional, string>, query = <required, string> })> # required for non-ES|QL mode
      filters                 = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters   = <optional, bool>
      sampling                = <optional, float64>
      group_by_json           = <required, json string with defaults>
      group_breakdown_by_json = <required, json string with defaults>
      metrics_json            = <required, json string with defaults> # exactly 1 metric in TF model
      legend                  = <required, object(...)>
      value_display           = <optional, object({ mode = <required, string>, percent_decimals = <optional, float64> })>
    })> # only with type = "lens"

    datatable_config = <optional, object({
      no_esql = <optional, object({
        title                 = <optional, string>
        description           = <optional, string>
        dataset_json          = <required, json string, normalized>
        density               = <required, object(...)>
        query                 = <required, object({ language = <optional, string>, query = <required, string> })>
        filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
        ignore_global_filters = <optional, bool>
        sampling              = <optional, float64>
        metrics               = <required, list(object({ config_json = <required, json string, normalized> }))>
        rows                  = <optional, list(object({ config_json = <required, json string, normalized> }))>
        split_metrics_by      = <optional, list(object({ config_json = <required, json string, normalized> }))>
        sort_by_json          = <optional, json string, normalized>
        paging                = <optional, int64>
      })>
      esql = <optional, object({
        title                 = <optional, string>
        description           = <optional, string>
        dataset_json          = <required, json string, normalized>
        density               = <required, object(...)>
        filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
        ignore_global_filters = <optional, bool>
        sampling              = <optional, float64>
        metrics               = <required, list(object({ config_json = <required, json string, normalized> }))>
        rows                  = <optional, list(object({ config_json = <required, json string, normalized> }))>
        split_metrics_by      = <optional, list(object({ config_json = <required, json string, normalized> }))>
        sort_by_json          = <optional, json string, normalized>
        paging                = <optional, int64>
      })>
    })> # only with type = "lens"; `no_esql` and `esql` conflict

    tagcloud_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      dataset_json          = <required, json string, normalized>
      query                 = <required, object({ language = <optional, string>, query = <required, string> })>
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      orientation           = <optional, string>
      font_size             = <optional, object({ min = <optional, float64>, max = <optional, float64> })>
      metric_json           = <required, json string with defaults>
      tag_by_json           = <required, json string with defaults>
    })> # only with type = "lens"

    heatmap_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      dataset_json          = <required, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })> # required for non-ES|QL mode
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      axes                  = <required, object(...)>
      cells                 = <required, object(...)>
      legend                = <required, object(...)>
      metric_json           = <required, json string with defaults>
      x_axis_json           = <required, json string, normalized>
      y_axis_json           = <optional, json string, normalized>
    })> # only with type = "lens"

    waffle_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      dataset_json          = <required, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })> # required for non-ES|QL mode; omit for ES|QL mode
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      legend                = <required, object(...)>
      value_display         = <optional, object({ mode = <required, string>, percent_decimals = <optional, float64> })>
      metrics               = <optional, list(object({ config = <required, json string with defaults> }))> # non-ES|QL; at least 1
      group_by              = <optional, list(object({ config = <required, json string with defaults> }))> # non-ES|QL
      esql_metrics          = <optional, list(object({ column = <required, string>, operation = <required, string>, label = <optional, string>, format_json = <required, json string, normalized>, color = <required, object(...)> }))> # ES|QL; at least 1
      esql_group_by         = <optional, list(object({ column = <required, string>, operation = <required, string>, collapse_by = <required, string>, color_json = <required, json string, normalized>, format_json = <optional, json string, normalized>, label = <optional, string> }))> # ES|QL
    })> # only with type = "lens"; ES|QL-vs-non-ES|QL consistency validator

    region_map_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      dataset_json          = <required, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })>
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      metric_json           = <required, json string with defaults>
      region_json           = <required, json string, normalized>
    })> # only with type = "lens"

    gauge_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      dataset_json          = <required, json string, normalized>
      query                 = <required, object({ language = <optional, string>, query = <required, string> })>
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      metric_json           = <required, json string with defaults>
      shape_json            = <optional, json string, normalized>
    })> # only with type = "lens"

    metric_chart_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      dataset_json          = <required, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })> # non-ES|QL branch
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      metrics               = <required, list(object({ config_json = <required, json string with defaults> }))> # at most 2
      breakdown_by_json     = <optional, json string, normalized>
    })> # only with type = "lens"

    pie_chart_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      dataset               = <optional, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })>
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, computed, bool> # default false
      sampling              = <optional, computed, float64> # default 1.0
      donut_hole            = <optional, string>
      label_position        = <optional, string>
      legend                = <optional, json string, normalized>
      metrics               = <required, list(object({ config = <required, json string with defaults> }))> # at least 1
      group_by              = <optional, list(object({ config = <required, json string with defaults> }))> # at least 1 when set
    })> # only with type = "lens"

    legacy_metric_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      dataset_json          = <required, json string, normalized>
      metric_json           = <required, json string with defaults>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })> # required for non-ES|QL dataset types
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      sampling              = <optional, float64>
      ignore_global_filters = <optional, bool>
    })> # only with type = "lens"

    config_json          = <optional, computed, json string with default-aware semantic equality> # conflicts with all typed panel config blocks
  }))> # each panel requires at least one config block

  sections = <optional, list(object({
    title     = <required, string>
    id        = <optional, computed, string>
    collapsed = <optional, bool>
    grid = {
      y = <required, int64>
    }
    panels = <optional, list(panel)>
  }))>

  access_control = <optional, object({
    access_mode = <optional, string> # one of: write_restricted | default
    owner       = <optional, string>
  })>
}
```

Notes:

- The resource is marked as technical preview in its schema description; sections are also marked as technical preview.
- The resource uses only the provider-level Kibana OpenAPI client; there is no resource-local Kibana connection override block.
- The resource does not declare a schema version, custom state upgrader, or resource-level compatibility gate in CRUD logic.
## Requirements
### Requirement: Kibana Dashboard APIs and request shaping (REQ-001)

The resource SHALL manage dashboards through Kibana's Dashboard HTTP APIs for create, get, update, and delete. For non-default spaces it SHALL call those APIs through a space-aware path rooted at `/s/<space_id>`, and for the default space it SHALL use the base dashboard path. Dashboard API requests SHALL include the request shaping used by the implementation: header `x-elastic-internal-origin: Kibana` and query parameters `apiVersion=1` and `allowUnmappedKeys=true`.

#### Scenario: Non-default space request

- GIVEN a dashboard with `space_id = "observability"`
- WHEN create, read, update, or delete runs
- THEN the resource SHALL target the dashboard API through the `observability` space-aware path

### Requirement: Client and API error surfacing (REQ-002)

For create, read, update, and delete, when the provider cannot supply a Kibana OpenAPI client, the operation SHALL return an error diagnostic. Transport errors and unexpected HTTP statuses from the dashboard APIs SHALL be surfaced as error diagnostics, except that read not-found and delete not-found SHALL be treated as absence rather than failure. If create returns success without a response body containing a dashboard id, the operation SHALL fail with an error diagnostic.

#### Scenario: Missing Kibana client

- GIVEN the provider configuration does not yield a Kibana OpenAPI client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Delete of already-absent dashboard

- GIVEN a managed dashboard has already been removed from Kibana
- WHEN delete runs and Kibana returns not found
- THEN the provider SHALL treat the delete as successful

### Requirement: Composite identity and computed ids (REQ-003)

The resource SHALL expose a computed canonical `id` in the format `<space_id>/<dashboard_id>`. `space_id` SHALL default to `default` when omitted. On create, if `dashboard_id` is omitted, the resource SHALL accept the id generated by Kibana and store it in both `dashboard_id` and the composite `id`. On read and update, the resource SHALL derive the target dashboard id and space from the composite `id`.

#### Scenario: Generated dashboard id

- GIVEN configuration omits `dashboard_id`
- WHEN create succeeds and Kibana returns dashboard id `abc123`
- THEN state SHALL contain `dashboard_id = "abc123"` and `id = "default/abc123"` unless another `space_id` was configured

### Requirement: Import passthrough and composite-id validation (REQ-004)

The resource SHALL support import by passing the imported string directly into the `id` attribute in state. Subsequent read, update, and delete operations SHALL require that stored `id` be parseable as `<space_id>/<dashboard_id>`, and malformed ids SHALL produce the composite-id diagnostic returned by the shared `CompositeID` parser.

#### Scenario: Invalid imported id

- GIVEN an imported id that does not contain exactly one `/`
- WHEN the resource later refreshes, updates, or deletes that instance
- THEN the provider SHALL return an error diagnostic describing the required composite id format

### Requirement: Provider-level Kibana client only (REQ-005)

The resource SHALL use the provider's configured Kibana OpenAPI client for all CRUD operations. The resource SHALL NOT expose or honor a resource-level connection override block for Kibana requests.

#### Scenario: Standard provider connection

- GIVEN the provider is configured with Kibana access
- WHEN the dashboard resource performs API calls
- THEN all calls SHALL use that provider-level Kibana OpenAPI client

### Requirement: Replacement fields and schema validation (REQ-006)

Schema validation SHALL enforce that each typed panel config block is only present on a panel whose `type` matches that block's panel type, and that at most one typed config block is present on any panel. This exclusivity requirement applies to `time_slider_control_config`, `slo_burn_rate_config`, `slo_error_budget_config`, `esql_control_config`, and `range_slider_control_config` in addition to all previously supported typed config blocks. Schema validation SHALL also enforce that `variable_type` and `control_type` values within `esql_control_config` are restricted to their documented enum values, and that `value` within `range_slider_control_config` contains exactly 2 elements when set.

The existing REQ-006 text is extended. The sentence:

> Each panel SHALL declare at least one panel configuration block, panel configuration blocks SHALL be mutually exclusive, typed panel configuration blocks SHALL only be valid for their supported panel type, and `waffle_config` SHALL enforce its ES|QL-vs-non-ES|QL field consistency rules.

gains the following additions:

- `esql_control_config` SHALL be valid only for panels with `type = "esql_control"`.
- `esql_control_config` SHALL be mutually exclusive with all other panel configuration blocks.
- The `variable_type` attribute within `esql_control_config` SHALL be restricted to the values `fields`, `values`, `functions`, `time_literal`, and `multi_values`; any other value SHALL be rejected at plan time.
- The `control_type` attribute within `esql_control_config` SHALL be restricted to the values `STATIC_VALUES` and `VALUES_FROM_QUERY`; any other value SHALL be rejected at plan time.
- `range_slider_control_config` SHALL be valid only for panels with `type = "range_slider_control"`.
- `range_slider_control_config` SHALL be mutually exclusive with all other panel configuration blocks and with `config_json`.
- The `value` attribute within `range_slider_control_config` SHALL contain exactly 2 elements when set; any other list length SHALL be rejected at plan time.

#### Scenario: esql_control_config rejected for non-esql_control panel (ADDED)

- GIVEN a panel with `type = "lens"` and `esql_control_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: Invalid variable_type value (ADDED)

- GIVEN a panel with `type = "esql_control"` and `esql_control_config.variable_type = "unsupported_type"`
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected at plan time with a diagnostic naming the allowed values

#### Scenario: Invalid control_type value (ADDED)

- GIVEN a panel with `type = "esql_control"` and `esql_control_config.control_type = "UNSUPPORTED"`
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected at plan time with a diagnostic naming the allowed values

#### Scenario: range_slider_control_config rejected for non-range_slider_control panel

- GIVEN a panel with `type = "lens"` and `range_slider_control_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

### Requirement: Create and update request mapping (REQ-007)

On create and update, the resource SHALL map Terraform state to the dashboard API request body using `title`, `description`, `time_range`, `refresh_interval`, query, tags, options, panels, sections, and access control when those values are known. Query mapping SHALL send `query_text` as the string branch of the API union and `query_json` as the object branch of the API union. If conversion of query or panel data fails, the operation SHALL return diagnostics and SHALL NOT proceed with the dashboard API call. After a successful create or update, the resource SHALL read the dashboard back and use that read as the authoritative final state; if the dashboard cannot be read back, the operation SHALL fail.

#### Scenario: Post-apply authoritative read

- GIVEN create or update succeeds
- WHEN the provider finalizes state
- THEN it SHALL re-read the dashboard and SHALL fail if the dashboard cannot be retrieved

### Requirement: Read behavior and missing-resource handling (REQ-008)

On refresh, the resource SHALL parse the composite `id`, read the dashboard from Kibana, and repopulate state from the API response. If Kibana returns not found, the resource SHALL remove itself from Terraform state. When a dashboard is found, the resource SHALL map title, description, time range, refresh interval, query, tags, options, access control, top-level panels, and sections back into state.

#### Scenario: Dashboard removed outside Terraform

- GIVEN a dashboard recorded in Terraform state
- WHEN refresh runs and Kibana returns not found
- THEN the resource SHALL remove the dashboard from state

### Requirement: State preservation for fields Kibana omits or defaults (REQ-009)

When Kibana omits or defaults fields on read, the resource SHALL preserve prior Terraform intent to avoid inconsistent results and spurious drift. Because the GET dashboard API does not return `time_range_mode`, the resource SHALL preserve the prior `time_range_mode` value already held in state or plan. When the GET dashboard API omits `access_control`, the resource SHALL preserve the prior `access_control` value instead of clearing it. When the options block was omitted in Terraform and Kibana materializes only the default dashboard options (`hide_panel_titles = false`, `use_margins = true`, `sync_colors = false`, `sync_tooltips = false`, `sync_cursor = true`), the resource SHALL keep the `options` block null in state. When a section's prior `collapsed` value was null and Kibana returns `false`, the resource SHALL preserve null rather than forcing `false` into state.

#### Scenario: Options omitted in config

- GIVEN Terraform configuration omitted the `options` block
- WHEN Kibana read-back contains only the dashboard option defaults
- THEN the resource SHALL keep `options` unset in state

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

The resource SHALL support top-level `panels`, section-contained `panels`, and `sections` in the order returned by the API and the order given in configuration when building requests. For panel reads, it SHALL distinguish sections from top-level panels and map each panel's `type`, `grid`, optional `id`, and configuration. For typed panel mappings, it SHALL seed from prior state or plan so that optional panel attributes omitted by Kibana on read can be preserved. When a panel is managed through `config_json` only, the resource SHALL preserve that JSON-centric representation and SHALL NOT populate typed configuration blocks from the API for that panel. On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, including `slo_burn_rate`, `slo_error_budget`, `esql_control`, and `range_slider_control`, or omitting all panel configuration blocks, SHALL return an error diagnostic. The `esql_control` panel type SHALL be managed exclusively through the typed `esql_control_config` block. The `range_slider_control` panel type SHALL be managed exclusively through the typed `range_slider_control_config` block.

#### Scenario: config_json rejected for esql_control panel type (ADDED)

- GIVEN a panel with `type = "esql_control"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `esql_control`

#### Scenario: config_json rejected for range_slider_control panel type

- GIVEN a panel with `type = "range_slider_control"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `range_slider_control`

---

### Requirement: Panel default normalization and XY-axis drift prevention (REQ-011)

The resource SHALL normalize `config_json` and typed Lens panel data with default-aware semantic equality so Kibana-injected defaults do not cause unnecessary drift. This normalization SHALL include panel-type-specific defaults such as missing empty `filters` arrays and Lens metric/grouping defaults used by the implementation. For XY chart panels, when `axis.x.scale` was unset in configuration and Kibana returns the implicit default `ordinal`, the resource SHALL preserve the unset Terraform value instead of forcing `ordinal` into state.

#### Scenario: Unset XY X-axis scale

- GIVEN an XY chart panel whose configuration left `axis.x.scale` unset
- WHEN read-back from Kibana returns `axis.x.scale = "ordinal"`
- THEN the provider SHALL keep the Terraform state value unset for that field

### Requirement: Markdown panel behavior (REQ-012)

For `type = "markdown"` panels, the resource SHALL accept `markdown_config` with the markdown panel fields `content`, `description`, `hide_title`, and `title`. On write, it SHALL send those fields when they are known. On read, it SHALL repopulate `markdown_config` from the API unless the panel is being managed through `config_json` only, in which case it SHALL preserve the JSON-only representation instead of filling typed markdown fields.

#### Scenario: Markdown panel managed through typed config

- GIVEN a `markdown` panel configured with `markdown_config`
- WHEN create, update, or read runs
- THEN the provider SHALL map the markdown fields between Terraform and the dashboard API

### Requirement: XY chart panel behavior (REQ-013)

For XY chart Lens panels, the resource SHALL require `axis`, `decorations`, `fitting`, `legend`, `query`, and at least one `layers` entry. Each layer SHALL represent either a data layer or a reference-line layer, not both. When the provider builds a typed Lens XY panel, it SHALL send the fixed Lens panel time range used by the implementation for typed Lens panels.

#### Scenario: XY panel requires layers

- GIVEN an XY chart panel configuration
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL require at least one layer and the fixed XY sub-blocks needed by the schema

### Requirement: Treemap panel behavior (REQ-014)

For treemap Lens panels, the resource SHALL require `dataset_json`, `group_by_json`, `metrics_json`, and `legend`. It SHALL treat the panel as non-ES|QL when a real `query` is present, and in that mode `query` SHALL be required. It SHALL treat the panel as ES|QL when `query` is omitted or both `query.query` and `query.language` are null. For semantic equality and read-back reconciliation, treemap `group_by_json` and `metrics_json` SHALL normalize the partition defaults used by the implementation, including terms-style defaults such as `collapse_by`, `format`, `rank_by`, and `size`.

#### Scenario: Treemap mode selection

- GIVEN a treemap panel with no usable `query`
- WHEN the provider converts it to or from the API model
- THEN it SHALL treat the panel as ES|QL mode rather than non-ES|QL mode

### Requirement: Mosaic panel behavior (REQ-015)

For mosaic Lens panels, the resource SHALL require `dataset_json`, `group_by_json`, `group_breakdown_by_json`, `metrics_json`, and `legend`. It SHALL use the same ES|QL-vs-non-ES|QL query rule as treemap panels, and non-ES|QL mosaics SHALL require `query`. `metrics_json` SHALL represent exactly one metric in the Terraform model. On read-back, mosaic partition dimensions SHALL be normalized to drop API-emitted top-level null keys that would otherwise create drift.

#### Scenario: Mosaic requires secondary breakdown

- GIVEN a mosaic panel configuration
- WHEN Terraform validates or the provider builds the API request
- THEN the panel SHALL require `group_breakdown_by_json` in addition to `group_by_json`

### Requirement: Datatable panel behavior (REQ-016)

For datatable Lens panels, the resource SHALL support exactly one of the `no_esql` and `esql` nested configurations. The non-ES|QL branch SHALL require `query` and SHALL map `dataset_json`, `density`, `metrics`, optional `rows`, optional `split_metrics_by`, optional `sort_by_json`, optional `paging`, optional `filters`, and optional `ignore_global_filters` / `sampling`. The ES|QL branch SHALL map the equivalent table configuration without a `query` block.

#### Scenario: Datatable mode blocks are exclusive

- GIVEN a datatable panel configuration
- WHEN Terraform validates the resource schema
- THEN it SHALL reject configurations that set both `no_esql` and `esql`

### Requirement: Tagcloud panel behavior (REQ-017)

For tagcloud Lens panels, the resource SHALL support the non-ES|QL tagcloud shape implemented by the provider. It SHALL require `dataset_json`, `query`, `metric_json`, and `tag_by_json`, with optional `filters`, `ignore_global_filters`, `sampling`, `orientation`, and `font_size`. For semantic equality it SHALL normalize tagcloud metric defaults and the `terms`-operation defaults for `tag_by_json`, including the default `rank_by` value.

#### Scenario: Tagcloud terms defaults

- GIVEN a tagcloud panel whose `tag_by_json` uses the `terms` operation without an explicit `rank_by`
- WHEN state is compared or refreshed
- THEN the provider SHALL treat the default `rank_by` as part of semantic equality

### Requirement: Heatmap panel behavior (REQ-018)

For heatmap Lens panels, the resource SHALL require `dataset_json`, `axes`, `cells`, `legend`, `metric_json`, and `x_axis_json`. It SHALL treat the panel as non-ES|QL when a real `query` is present, and in that mode `query` SHALL be required. It SHALL treat the panel as ES|QL when `query` is omitted or empty by the implementation's mode test. Heatmap metric normalization SHALL use the same metric-default behavior shared with the tagcloud implementation.

#### Scenario: Non-ES|QL heatmap requires query

- GIVEN a heatmap panel using the non-ES|QL branch
- WHEN the provider builds the API request
- THEN it SHALL require `query` to be present

### Requirement: Waffle panel behavior (REQ-019)

For waffle Lens panels, the resource SHALL enforce mutually exclusive non-ES|QL and ES|QL modes. In non-ES|QL mode it SHALL require `query` and at least one `metrics` entry, and it MAY accept `group_by`. In ES|QL mode it SHALL require at least one `esql_metrics` entry, it MAY accept `esql_group_by`, and it SHALL reject `metrics` and `group_by`. On read-back, the provider SHALL preserve the waffle fields that Kibana may omit or materialize differently, including the implementation's merge behavior for `ignore_global_filters`, `sampling`, legend values, visibility, and value-display details. ES|QL number-format JSON for waffle metric formats SHALL normalize the default decimals and compact settings trimmed by the implementation.

#### Scenario: Waffle ES|QL validation

- GIVEN a waffle panel in ES|QL mode
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL require at least one `esql_metrics` entry and SHALL reject `metrics` or `group_by`

### Requirement: Region map panel behavior (REQ-020)

For region-map Lens panels, the resource SHALL require `dataset_json`, `metric_json`, and `region_json`. It SHALL use the implementation's query-based mode selection: when `query.query` is known it SHALL use the non-ES|QL branch, otherwise it SHALL use the ES|QL branch. Region-map metric normalization SHALL apply the region-map metric defaults used by the implementation.

#### Scenario: Region map without known query

- GIVEN a region-map panel whose `query` does not contain a known query string
- WHEN the provider converts the panel
- THEN it SHALL use the ES|QL branch

### Requirement: Gauge panel behavior (REQ-021)

For gauge Lens panels, the resource SHALL support the non-ES|QL gauge shape implemented by the provider. It SHALL require `dataset_json`, `query`, and `metric_json`, and it MAY accept `shape_json`, `filters`, `ignore_global_filters`, and `sampling`. Gauge metric semantic equality SHALL include the implementation's defaults for `empty_as_null`, `hide_title`, and `ticks`.

#### Scenario: Gauge metric defaults

- GIVEN a gauge metric configuration that omits the implementation's defaulted fields
- WHEN the provider compares or refreshes state
- THEN it SHALL normalize those defaults for semantic equality

### Requirement: Metric chart panel behavior (REQ-022)

For metric-chart Lens panels, the resource SHALL map the provider's two metric-chart variants: the non-ES|QL branch when `query` is present, and the ES|QL branch when the implementation detects the query as absent or empty. It SHALL require `dataset_json` and `metrics`, and it MAY accept `breakdown_by_json`, `filters`, `ignore_global_filters`, and `sampling`. Each metric `config_json` SHALL use the shared Lens metric default normalization used by the implementation.

#### Scenario: Metric chart ES|QL read-back

- GIVEN a metric-chart panel in the ES|QL variant
- WHEN the provider reads it back from Kibana
- THEN it SHALL not repopulate a non-ES|QL `query` block into state

### Requirement: Pie chart panel behavior (REQ-023)

For pie Lens panels, the resource SHALL require at least one `metrics` entry and MAY accept `group_by`. It SHALL select the non-ES|QL branch when `query` is present and the ES|QL branch otherwise. When Kibana omits `ignore_global_filters` or `sampling` on read, the provider SHALL treat their default values as `false` and `1.0` respectively. Pie metric and group-by semantic equality SHALL normalize the implementation's pie metric defaults and Lens group-by defaults.

#### Scenario: Pie chart API defaults

- GIVEN a pie panel read from Kibana without explicit `ignore_global_filters` or `sampling`
- WHEN state is refreshed
- THEN the provider SHALL reconcile those fields as `false` and `1.0`

### Requirement: Legacy metric panel behavior (REQ-024)

For legacy-metric Lens panels, the resource SHALL choose its mode from `dataset_json.type`. `dataView` and `index` datasets SHALL use the non-ES|QL branch and SHALL require `query`. `esql` and `table` datasets SHALL use the ES|QL branch and SHALL not require `query`. Legacy metric semantic equality SHALL normalize the implementation's legacy metric defaults, including absent `filters` and the metric/format defaults applied by the resource.

#### Scenario: Legacy metric dataset mode

- GIVEN a legacy metric panel whose `dataset_json.type` is `esql`
- WHEN the provider converts the panel
- THEN it SHALL use the ES|QL branch rather than the non-ES|QL branch

### Requirement: Raw `config_json` panel behavior (REQ-025)

When a panel is authored through `config_json`, the resource SHALL accept only `markdown` and `lens` panel types for write. It SHALL deserialize the raw JSON into the corresponding dashboard panel config and SHALL fail if that JSON cannot be unmarshaled into the supported API config type. For read-back, it SHALL always refresh `config_json` from the API payload using the implementation's default-aware JSON semantics. When a Lens panel is authored through raw `config_json`, the provider SHALL preserve that raw Lens config path rather than re-expressing it through a typed panel block, and it SHALL not apply the typed-Lens `lensPanelTimeRange()` injection path used by the typed converters.

#### Scenario: Unsupported raw config panel type

- GIVEN a panel configured with `config_json` and a panel `type` other than `markdown` or `lens`
- WHEN the provider builds the API request
- THEN it SHALL return an error diagnostic for unsupported `config_json` panel type

### Requirement: Time slider control panel behavior (REQ-029)

For `time_slider_control` panels, the resource SHALL accept an optional `time_slider_control_config` block. The block itself is optional; a panel with `type = "time_slider_control"` and no `time_slider_control_config` block is valid and uses Kibana defaults for the slider position and anchoring behavior. All fields within `time_slider_control_config` are optional.

The `start_percentage_of_time_range` and `end_percentage_of_time_range` attributes are **float32** values in Terraform state, matching Kibana's API type (`float32`). They represent the start and end positions of the slider window as a fraction of the dashboard's global time range. When either attribute is configured, the provider SHALL validate that its value is between 0.0 and 1.0 inclusive, and SHALL return an error diagnostic at plan time if the validation fails.

Using float32 in the provider schema avoids refresh-time plan drift that can occur when values such as `0.1` are authored as float64, serialized to the API as float32, and read back into wider float64 state (binary representation mismatch). Practitioners still author familiar decimal literals in HCL; Terraform coerces them to the schema's float32 type. Because this feature has not been publicly released yet, adopting float32 for these attributes does not require compatibility with any prior released float64 state shape or a migration path.

The `is_anchored` attribute is a bool indicating whether the time window start is anchored. When present, the provider SHALL write it to the API payload. When absent, the provider SHALL omit it from the write payload.

When the provider reads a `time_slider_control` panel back from Kibana, it SHALL preserve the null intent for each config field. If a config field is null in Terraform state (i.e. the practitioner did not configure it), the provider SHALL NOT populate that field from the Kibana read response, even if Kibana returns a value for it. This applies to all three config attributes: `start_percentage_of_time_range`, `end_percentage_of_time_range`, and `is_anchored`.

When Kibana returns an empty or absent `config` object for a `time_slider_control` panel, the provider SHALL treat it as equivalent to an omitted `time_slider_control_config` block and SHALL NOT synthesize a non-null block in state.

The `time_slider_control_config` block SHALL conflict with all other typed panel config blocks and with practitioner-authored `config_json`. This conflict SHALL be enforced by schema-level validators at plan time, consistent with the pattern used by other typed panel config blocks.

Practitioner-authored `config_json` SHALL NOT be used when `type = "time_slider_control"`. The provider SHALL reject that combination at plan time via schema validation on `config_json` (type allowlist). The nested panel object validator documents the rule in its description but SHALL NOT emit a second diagnostic for the same misconfiguration. The `config_json` attribute MAY still appear in Terraform state as a computed read-back of Kibana's serialized panel config; practitioners SHALL use `time_slider_control_config` (optional) or omit panel config instead of setting `config_json` in configuration.

#### Scenario: Time slider control panel with empty config block

- GIVEN a `time_slider_control` panel with an empty `time_slider_control_config = {}` block (block present, all fields omitted)
- WHEN the provider builds the API request
- THEN it SHALL send the panel with an empty `config` object
- AND SHALL NOT return an error diagnostic

#### Scenario: Time slider control panel with percentage fields

- GIVEN a `time_slider_control` panel with `start_percentage_of_time_range = 0.1` and `end_percentage_of_time_range = 0.9`
- WHEN the provider builds the API request
- THEN it SHALL include those values in the `config` object (as float32-compatible values)

#### Scenario: Percentage field out of range

- GIVEN a `time_slider_control` panel with `start_percentage_of_time_range = 1.5`
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return an error diagnostic indicating the value must be between 0.0 and 1.0

#### Scenario: Reject practitioner-authored config_json

- GIVEN a `time_slider_control` panel with `config_json` set in Terraform configuration
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return an error diagnostic (for example `Invalid Configuration` on the `config_json` attribute from the type allowlist validator)

#### Scenario: Null-preservation on read-back

- GIVEN a `time_slider_control` panel where `start_percentage_of_time_range` is null in Terraform state
- AND Kibana returns a value for `start_percentage_of_time_range` in its read response
- WHEN the provider refreshes state
- THEN it SHALL leave `start_percentage_of_time_range` as null in state
- AND SHALL NOT produce a plan diff for that field

#### Scenario: Configured field round-trips on read-back

- GIVEN a `time_slider_control` panel with `start_percentage_of_time_range = 0.1` and `end_percentage_of_time_range = 0.9` in Terraform state
- WHEN the provider refreshes state from Kibana
- THEN it SHALL populate those fields from the Kibana response using the same float32 semantics as the API
- AND SHALL produce no plan diff when the Kibana values match the configured values (including after a subsequent plan-only refresh)

#### Scenario: Empty Kibana config treated as omitted block

- GIVEN a `time_slider_control` panel where Kibana returns an empty or absent `config` object
- AND the practitioner has not configured a `time_slider_control_config` block
- WHEN the provider refreshes state
- THEN it SHALL leave `time_slider_control_config` as null in state

### Requirement: SLO burn rate panel behavior (REQ-032)

The resource SHALL support `type = "slo_burn_rate"` panels through the typed `slo_burn_rate_config` block. The block requires `slo_id` and `duration`, and optionally accepts `slo_instance_id`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns`.

The `duration` field SHALL be validated at plan time against the pattern `^\d+[mhd]$`. Any value that does not match SHALL be rejected before any dashboard API call.

On write, the provider SHALL map `slo_burn_rate_config` to the `config` object in the `slo-burn-rate-embeddable` API schema. Optional fields SHALL be included only when set; absent optional fields SHALL NOT be sent to the API. When `drilldowns` is set, each drilldown object SHALL accept practitioner-configured `url` and `label`, SHALL hardcode `trigger = "on_open_panel_menu"` and `type = "url_drilldown"` in the API request, and SHALL include the optional attributes (`encode_url`, `open_in_new_tab`) only when explicitly set.

On read, the `slo_instance_id` field SHALL use null-preservation: if the prior state value was null and the API returns `"*"`, the provider SHALL keep `slo_instance_id` null in state rather than introducing the API sentinel. When `slo_instance_id` is explicitly configured to `"*"`, the provider SHALL round-trip it normally.

#### Scenario: Creation of slo_burn_rate panel with required fields

- GIVEN a dashboard configuration containing an `slo_burn_rate` panel with `slo_burn_rate_config.slo_id = "my-slo-id"` and `slo_burn_rate_config.duration = "72h"`
- WHEN the resource is created
- THEN the provider SHALL send the mapped `config` object to the Kibana dashboard API with `slo_id` and `duration`
- AND `slo_instance_id` SHALL be null in state

#### Scenario: slo_instance_id null-preservation after read-back

- GIVEN a dashboard configuration containing an `slo_burn_rate` panel that does not set `slo_instance_id`
- WHEN the resource is created and then read back from Kibana
- AND Kibana returns `slo_instance_id = "*"` in the API response
- THEN the provider SHALL keep `slo_instance_id` as null in state
- AND a subsequent plan SHALL show no changes

#### Scenario: Creation of slo_burn_rate panel with slo_instance_id and drilldowns

- GIVEN a dashboard configuration containing an `slo_burn_rate` panel with `slo_instance_id = "host-a"` and a drilldown entry
- WHEN the resource is created and read back
- THEN all configured attributes SHALL be present in state and a subsequent plan SHALL show no changes

### Requirement: SLO error budget panel behavior (REQ-031)

For `type = "slo_error_budget"` panels, the resource SHALL accept a typed `slo_error_budget_config` block containing the fields of the `slo-error-budget-embeddable` API schema. `slo_id` SHALL be required. `slo_instance_id`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns` SHALL be optional. `slo_error_budget_config` SHALL be mutually exclusive with all other typed panel config blocks and with `config_json`.

On write, the provider SHALL map all configured fields from `slo_error_budget_config` into the Kibana dashboard panel API request for the `slo_error_budget` embeddable type.

On read, the provider SHALL repopulate `slo_error_budget_config` from the API response. For `slo_instance_id`, the provider SHALL preserve the prior Terraform state value when the prior value was null: if the practitioner did not configure `slo_instance_id`, the provider SHALL NOT write the API-returned default `"*"` into state. For `encode_url` and `open_in_new_tab` drilldown fields, the provider SHALL normalize the API default value of `true` so that practitioners who omit those fields do not observe spurious drift after apply. On import, the provider SHALL populate API-returned optional display fields such as `title`, `description`, `hide_title`, and `hide_border`.

`drilldowns` SHALL be represented as a list of typed objects. Each drilldown object SHALL contain required `url` (string) and `label` (string), and optional `encode_url` (bool, default `true`) and `open_in_new_tab` (bool, default `true`). On write, the provider SHALL set Kibana's fixed `trigger = "on_open_panel_menu"` and `type = "url_drilldown"` values in the API request.

#### Scenario: Minimal slo_error_budget panel with only slo_id

- GIVEN a panel with `type = "slo_error_budget"` and `slo_error_budget_config { slo_id = "my-slo-id" }`
- WHEN create and subsequent read run
- THEN the provider SHALL send `slo_id = "my-slo-id"` in the API request
- AND SHALL read it back into state without error

#### Scenario: slo_instance_id null preservation

- GIVEN a panel with `type = "slo_error_budget"` and `slo_error_budget_config` that omits `slo_instance_id`
- WHEN the dashboard is created and subsequently read back from Kibana
- THEN the provider SHALL keep `slo_instance_id` null in state even if Kibana returns `"*"` as the default value
- AND a subsequent plan SHALL show no changes for `slo_instance_id`

#### Scenario: drilldowns configuration

- GIVEN a panel with `type = "slo_error_budget"` and `slo_error_budget_config` containing a `drilldowns` block with `url` and `label`
- WHEN the dashboard is created and subsequently read back from Kibana
- THEN the provider SHALL round-trip the typed drilldown fields
- AND SHALL apply default normalization for `encode_url` and `open_in_new_tab` so that omitting them in configuration does not produce drift
- AND SHALL write Kibana's fixed `trigger` and `type` values into the API request

### Requirement: ES|QL control panel behavior (REQ-026)

For `type = "esql_control"` panels, the resource SHALL accept `esql_control_config` with the required fields `selected_options`, `variable_name`, `variable_type`, `esql_query`, and `control_type`, and the optional fields `title`, `single_select`, `available_options`, and `display_settings`.

On write (create and update), the resource SHALL map `esql_control_config` to the `config` object in the `kbn-dashboard-panel-esql_control` API schema. All required fields SHALL be included in the API request. Optional fields SHALL be included only when they are set in Terraform state; absent optional fields SHALL NOT be sent to the API. The `display_settings` sub-object SHALL be sent only when the `display_settings` block is set; within that block, only attributes that are explicitly set SHALL be included.

On read, the resource SHALL repopulate `esql_control_config` from the API response. Fields that the API response omits SHALL not be forced into state. When `selected_options` is returned by the API, the provider SHALL preserve the API-returned ordering. The provider SHALL NOT apply a typed `config_json` round-trip for `esql_control` panels; such panels are always managed through the typed `esql_control_config` block.

The `esql_control` panel type is a standalone control panel, not a Lens visualization. It does not reference a saved object, and its configuration is fully inline in the dashboard document. As a result, none of the Lens panel converters, Lens time-range injection, or Lens metric default normalization SHALL apply to `esql_control` panels.

#### Scenario: Creation of esql_control panel with required fields using STATIC_VALUES

- GIVEN a dashboard configuration containing an `esql_control` panel with:
  - `type = "esql_control"`
  - `esql_control_config.selected_options = ["option_a", "option_b"]`
  - `esql_control_config.variable_name = "my_var"`
  - `esql_control_config.variable_type = "values"`
  - `esql_control_config.esql_query = "FROM logs-* | STATS count = COUNT(*) BY host.name"`
  - `esql_control_config.control_type = "STATIC_VALUES"`
- WHEN the resource is created
- THEN the provider SHALL send the mapped `config` object to the Kibana dashboard API
- AND the panel SHALL appear in state with all five required fields populated
- AND the provider SHALL NOT populate `config_json` for this panel in state

#### Scenario: Creation of esql_control panel with VALUES_FROM_QUERY control type

- GIVEN a dashboard configuration containing an `esql_control` panel with:
  - `esql_control_config.control_type = "VALUES_FROM_QUERY"`
  - `esql_control_config.variable_type = "fields"`
  - `esql_control_config.variable_name = "target_field"`
  - `esql_control_config.esql_query = "FROM logs-* | KEEP host.name"`
  - `esql_control_config.selected_options = []`
- WHEN the resource is created
- THEN the provider SHALL send the control to Kibana with `control_type = "VALUES_FROM_QUERY"`
- AND on read-back the provider SHALL refresh `selected_options` from the API response without treating an API-returned empty list as drift

#### Scenario: esql_control panel with display_settings

- GIVEN a dashboard panel with `esql_control_config` including a `display_settings` block with `hide_action_bar = true` and `placeholder = "Select a value"`
- WHEN the resource is created or updated
- THEN the provider SHALL include the `display_settings` object in the API request with the set attributes
- AND on read-back the provider SHALL repopulate `display_settings` from the API response

#### Scenario: Read-back of esql_control panel preserves optional fields

- GIVEN a managed `esql_control` panel whose `esql_control_config` omits `title` and `display_settings`
- WHEN Kibana returns the panel without those optional fields
- THEN the provider SHALL keep `title` and `display_settings` as null/unset in state
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: Validation of variable_type enum values

- GIVEN an `esql_control_config` block with `variable_type = "time_literal"`
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be accepted
- GIVEN an `esql_control_config` block with `variable_type = "unsupported"`
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic listing `fields`, `values`, `functions`, `time_literal`, `multi_values` as the valid values

#### Scenario: esql_control panel grid defaults

- GIVEN an `esql_control` panel with only `grid.x` and `grid.y` specified
- WHEN the resource is created
- THEN the provider SHALL apply the API defaults for panel width (`w = 24`) and height (`h = 15`) consistent with the `kbn-dashboard-panel-esql_control` schema

### Requirement: Range slider control panel behavior (REQ-028)

For `type = "range_slider_control"` panels, the resource SHALL accept `range_slider_control_config` with the following attributes:

- **`data_view_id`** (required, string): the ID of the Kibana data view that the slider filter targets.
- **`field_name`** (required, string): the numeric field within the data view that the slider operates on.
- **`title`** (optional, string): a human-readable label displayed above the slider in the dashboard.
- **`use_global_filters`** (optional, bool): when set, controls whether the panel respects dashboard-level global filters.
- **`ignore_validations`** (optional, bool): when set, suppresses validation errors from the control during intermediate states.
- **`value`** (optional, list(string)): the initial min/max range pre-populated on the slider, expressed as a 2-element list `[min, max]`. When set, the list MUST contain exactly 2 elements. The values are strings matching the API representation.
- **`step`** (optional, number): the step size for each increment of the slider.

On write, the resource SHALL send `data_view_id` and `field_name` unconditionally and SHALL include each optional field only when it is set to a known, non-null value. On read, the resource SHALL populate `range_slider_control_config` from the API response for panels with `type = "range_slider_control"` and SHALL leave optional fields null in state when the API does not return them.

The `range_slider_control_config` block is valid only when `type = "range_slider_control"` and MUST NOT appear with any other typed panel config block or with `config_json`.

#### Scenario: Required fields only

- GIVEN a `range_slider_control` panel configured with only `data_view_id` and `field_name`
- WHEN create or update runs
- THEN the API request SHALL include `data_view_id` and `field_name` in the panel config and SHALL omit all unset optional fields

#### Scenario: Optional range pre-selection

- GIVEN a `range_slider_control` panel configured with `value = ["10", "500"]`
- WHEN create or update runs
- THEN the API request SHALL include `value` as a 2-element array matching the configured strings
- AND when read-back occurs, state SHALL reflect `value = ["10", "500"]`

#### Scenario: Invalid value list length

- GIVEN a `range_slider_control_config` block with `value` set to a list with fewer or more than 2 elements
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic stating that `value` must contain exactly 2 elements

#### Scenario: config_json rejected for range_slider_control

- GIVEN a panel with `type = "range_slider_control"` configured with `config_json` instead of `range_slider_control_config`
- WHEN the provider builds the API request
- THEN it SHALL return an error diagnostic for unsupported `config_json` panel type

## Traceability

| Area | Primary files |
|------|----------------|
| Schema | `internal/kibana/dashboard/schema.go` |
| Metadata / Configure / Import | `internal/kibana/dashboard/resource.go` |
| CRUD orchestration | `internal/kibana/dashboard/create.go`, `internal/kibana/dashboard/read.go`, `internal/kibana/dashboard/update.go`, `internal/kibana/dashboard/delete.go` |
| Top-level model mapping | `internal/kibana/dashboard/models.go` |
| Options / access control mapping | `internal/kibana/dashboard/models_options.go`, `internal/kibana/dashboard/models_access_control.go` |
| Panels / sections mapping | `internal/kibana/dashboard/models_panels.go` |
| Visualization-specific panel converters | `internal/kibana/dashboard/models_*_panel.go` |
| Drift normalization | `internal/kibana/dashboard/panel_config_defaults.go`, `internal/kibana/dashboard/models_xy_chart_panel.go` |
| Waffle validation | `internal/kibana/dashboard/waffle_config_validator.go` |
| Dashboard API status handling | `internal/clients/kibanaoapi/dashboards.go` |
| Composite id parsing | `internal/clients/api_client.go` |
