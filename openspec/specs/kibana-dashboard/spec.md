# `elasticstack_kibana_dashboard` — Schema and Functional Requirements

Resource implementation: `internal/kibana/dashboard` (dashboard model and `mapPanelFromAPI` in `models_panels.go` and `models.go`).

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_dashboard` resource, including Kibana Dashboard API usage, composite identity and import, provider-level Kibana OpenAPI client usage, replacement-vs-update behavior, query and panel mapping, section handling, and drift-resistant state reconciliation for dashboard options and panel defaults.

## Schema

```hcl
resource "elasticstack_kibana_dashboard" "example" {
  id           = <computed, string> # canonical state id: "<space_id>/<dashboard_id>"; UseStateForUnknown
  space_id     = <optional, computed, string> # default "default"; RequiresReplace
  dashboard_id = <computed, string> # Kibana-assigned dashboard id; UseStateForUnknown

  title       = <required, string>
  description = <optional, string>

  time_range = {
    from = <required, string>
    to   = <required, string>
    mode = <optional, string> # absolute | relative; preserved when GET omits mode (REQ-009)
  }

  refresh_interval = {
    pause = <required, bool>
    value = <required, int64>
  }

  query = {
    language = <required, string>
    text     = <optional, string> # conflicts with json; string branch of API union
    json     = <optional, json string, normalized> # conflicts with text; object branch
  }

  tags = <optional, list(string)>

  options = <optional, object({
    hide_panel_titles  = <optional, bool>
    use_margins        = <optional, bool>
    sync_colors        = <optional, bool>
    sync_tooltips      = <optional, bool>
    sync_cursor        = <optional, bool>
    auto_apply_filters = <optional, bool>
    hide_panel_borders = <optional, bool>
  })>

  panels = <optional, list(object({
    type = <required, string>
    grid = {
      x = <required, int64>
      y = <required, int64>
      w = <optional, int64>
      h = <optional, int64>
    }
    id = <optional, computed, string> # Terraform id aligned with API id; UseNonNullStateForUnknown

    markdown_config = <optional, object({
      by_value = <optional, object({
        content     = <required, string>
        settings    = <required, object({
          open_links_in_new_tab = <optional, bool>
        })>
        description = <optional, string>
        hide_title  = <optional, bool>
        title       = <optional, string>
        hide_border = <optional, bool>
      })>
      by_reference = <optional, object({
        ref_id      = <required, string>
        description = <optional, string>
        hide_title  = <optional, bool>
        title       = <optional, string>
        hide_border = <optional, bool>
      })>
    })> # only with type = "markdown"; exactly one of by_value or by_reference (REQ-012); conflicts with all other config blocks

    xy_chart_config = <optional, object({
      title       = <optional, string>
      description = <optional, string>
      axis        = <required, object({ x = <optional, object({ domain_json = <optional, json string, normalized>, ... })>, y = <optional, object({ domain_json = <required, json string, normalized>, ... })>, y2 = <optional, object({ domain_json = <required, json string, normalized>, ... })> })>
      decorations = <required, object(...)>
      fitting     = <required, object({ type = <required, string>, dotted = <optional, bool>, end_value = <optional, string> })>
      layers      = <required, list(object({ type = <required, string>, data_layer = <optional, object(...)>, reference_line_layer = <optional, object(...)> }))> # at least 1
      legend      = <required, object(...)>
      query       = <required, object({ language = <optional, string>, query = <required, string> })>
      filters     = <optional, list(object({ filter_json = <required, json string, normalized> }))>
    })> # only with type = "vis"

    treemap_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      data_source_json      = <required, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })> # required for non-ES|QL mode
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      group_by_json         = <required, json string with defaults>
      metrics_json          = <required, json string with defaults>
      label_position        = <optional, string>
      legend                = <required, object(...)>
      value_display         = <optional, object({ mode = <required, string>, percent_decimals = <optional, float64> })>
    })> # only with type = "vis"

    mosaic_config = <optional, object({
      title                   = <optional, string>
      description             = <optional, string>
      data_source_json        = <required, json string, normalized>
      query                   = <optional, object({ language = <optional, string>, query = <required, string> })> # required for non-ES|QL mode
      filters                 = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters   = <optional, bool>
      sampling                = <optional, float64>
      group_by_json           = <required, json string with defaults>
      group_breakdown_by_json = <required, json string with defaults>
      metrics_json            = <required, json string with defaults> # exactly 1 metric in TF model
      legend                  = <required, object(...)>
      value_display           = <optional, object({ mode = <required, string>, percent_decimals = <optional, float64> })>
    })> # only with type = "vis"

    datatable_config = <optional, object({
      no_esql = <optional, object({
        title                 = <optional, string>
        description           = <optional, string>
        data_source_json      = <required, json string, normalized>
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
        data_source_json      = <required, json string, normalized>
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
    })> # only with type = "vis"; `no_esql` and `esql` conflict

    tagcloud_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      data_source_json      = <required, json string, normalized>
      query                 = <required, object({ language = <optional, string>, query = <required, string> })>
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      orientation           = <optional, string>
      font_size             = <optional, object({ min = <optional, float64>, max = <optional, float64> })>
      metric_json           = <required, json string with defaults>
      tag_by_json           = <required, json string with defaults>
    })> # only with type = "vis"

    heatmap_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      data_source_json      = <required, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })> # required for non-ES|QL mode
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      axis                  = <required, object(...)>
      styling               = <required, object({ cells = <required, object(...)> })>
      legend                = <required, object({ visibility = <optional, string>, ... })> # visibility: visible | hidden
      metric_json           = <required, json string with defaults>
      x_axis_json           = <required, json string, normalized>
      y_axis_json           = <optional, json string, normalized>
    })> # only with type = "vis"

    waffle_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      data_source_json      = <required, json string, normalized>
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
    })> # only with type = "vis"; ES|QL-vs-non-ES|QL consistency validator

    region_map_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      data_source_json      = <required, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })>
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      metric_json           = <required, json string with defaults>
      region_json           = <required, json string, normalized>
    })> # only with type = "vis"

    gauge_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      data_source_json      = <required, json string, normalized>
      query                 = <required, object({ language = <optional, string>, query = <required, string> })>
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      metric_json           = <required, json string with defaults>
      shape_json            = <optional, json string, normalized>
    })> # only with type = "vis"

    metric_chart_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      data_source_json      = <required, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })> # non-ES|QL branch
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, bool>
      sampling              = <optional, float64>
      metrics               = <required, list(object({ config_json = <required, json string with defaults> }))> # at most 2
      breakdown_by_json     = <optional, json string, normalized>
    })> # only with type = "vis"

    pie_chart_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      data_source_json      = <optional, json string, normalized>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })>
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      ignore_global_filters = <optional, computed, bool> # default false
      sampling              = <optional, computed, float64> # default 1.0
      donut_hole            = <optional, string>
      label_position        = <optional, string>
      legend = <optional, computed, object({
        nested               = <optional, bool>
        size                 = <required, string> # auto | s | m | l | xl
        truncate_after_lines = <optional, float64>
        visible              = <optional, string> # auto | visible | hidden; maps to API `visibility`
      })> # schema default when omitted (typical size/visibility auto); optional+computed for Terraform
      metrics               = <required, list(object({ config = <required, json string with defaults> }))> # at least 1
      group_by              = <optional, list(object({ config = <required, json string with defaults> }))> # at least 1 when set
    })> # only with type = "vis"

    legacy_metric_config = <optional, object({
      title                 = <optional, string>
      description           = <optional, string>
      data_source_json      = <required, json string, normalized>
      metric_json           = <required, json string with defaults>
      query                 = <optional, object({ language = <optional, string>, query = <required, string> })> # required for supported data_view_reference / data_view_spec sources
      filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
      sampling              = <optional, float64>
      ignore_global_filters = <optional, bool>
    })> # only with type = "vis"

    synthetics_stats_overview_config = <optional, object({
      title       = <optional, string>
      description = <optional, string>
      hide_title  = <optional, bool>
      hide_border = <optional, bool>
      drilldowns  = <optional, list(object({ url = <required, string>, label = <required, string>, encode_url = <optional, bool>, open_in_new_tab = <optional, bool> }))>
      filters     = <optional, object({
        projects      = <optional, list(object({ label = <required, string>, value = <required, string> }))>
        tags          = <optional, list(object({ label = <required, string>, value = <required, string> }))>
        monitor_ids   = <optional, list(object({ label = <required, string>, value = <required, string> }))>
        locations     = <optional, list(object({ label = <required, string>, value = <required, string> }))>
        monitor_types = <optional, list(object({ label = <required, string>, value = <required, string> }))>
        statuses      = <optional, list(object({ label = <required, string>, value = <required, string> }))>
      })>
    })> # only with type = "synthetics_stats_overview"

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
  })>
}
```

Notes:

- The resource is marked as technical preview in its schema description; sections are also marked as technical preview.
- The resource uses only the provider-level Kibana OpenAPI client; there is no resource-local Kibana connection override block.
- The resource does not declare a schema version, custom state upgrader, or resource-level compatibility gate in CRUD logic.
## Requirements
### Requirement: Kibana Dashboard APIs and request shaping (REQ-001)

The resource SHALL manage dashboards through Kibana's Dashboard HTTP APIs for create, get, update, and delete. For non-default spaces it SHALL call those APIs through a space-aware path rooted at `/s/<space_id>`, and for the default space it SHALL use the base dashboard path. Dashboard API requests SHALL include the request shaping used by the implementation: query parameter `allowUnmappedKeys=true`.

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

`dashboard_id` SHALL be `Optional + Computed` (not `Computed`-only). When the practitioner provides `dashboard_id` in configuration, the resource SHALL preserve that value as the dashboard identifier. When `dashboard_id` is omitted, the resource SHALL accept the id generated by Kibana on create. The composite `id` format `<space_id>/<dashboard_id>` is unchanged.

When `dashboard_id` is provided in configuration and later changes, the resource SHALL be destroyed and recreated (`RequiresReplace`), consistent with `space_id` behavior.

#### Scenario: Kibana-generated dashboard id (existing behavior preserved)

- GIVEN configuration does not provide `dashboard_id`
- WHEN create succeeds and Kibana returns dashboard id `abc123`
- THEN state SHALL contain `dashboard_id = "abc123"` and `id = "default/abc123"` unless another `space_id` was configured

#### Scenario: Practitioner-supplied dashboard id on create

- GIVEN configuration provides `dashboard_id = "my-team-overview"` and no dashboard with that id exists
- WHEN create is called
- THEN the resource SHALL call `PUT /api/dashboards/dashboard/my-team-overview`
- AND state SHALL contain `dashboard_id = "my-team-overview"` and `id = "default/my-team-overview"` unless another `space_id` was configured

#### Scenario: Server-returned id matches supplied id

- GIVEN configuration provides `dashboard_id = "my-team-overview"` and create uses PUT
- WHEN the API returns `201 Created` with `id = "my-team-overview"` in the response body
- THEN state SHALL record `dashboard_id = "my-team-overview"` (the server-returned value)

#### Scenario: Changing dashboard_id triggers replacement

- GIVEN state contains `dashboard_id = "old-id"` and the practitioner changes configuration to `dashboard_id = "new-id"`
- WHEN Terraform plans the change
- THEN the plan SHALL show a destroy+create replacement (RequiresReplace; no in-place update of `dashboard_id`)

### Requirement: Import passthrough and composite-id validation (REQ-004)

The resource SHALL support import by passing the imported string directly into the `id` attribute in state. Subsequent read, update, and delete operations SHALL require that stored `id` be parseable as `<space_id>/<dashboard_id>`, and malformed ids SHALL produce the composite-id diagnostic returned by the shared `CompositeID` parser.

#### Scenario: Invalid imported id

- GIVEN an imported id that does not contain exactly one `/`
- WHEN the resource later refreshes, updates, or deletes that instance
- THEN the provider SHALL return an error diagnostic describing the required composite id format

### Requirement: Effective Kibana client selection (REQ-005)

The resource SHALL use the provider's configured Kibana OpenAPI client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for all CRUD operations.

#### Scenario: Standard provider connection

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** all dashboard API calls SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all dashboard API calls SHALL use the scoped Kibana OpenAPI client derived from that block

### Requirement: Replacement fields and schema validation (REQ-006)

Schema validation SHALL enforce that each typed panel config block is only present on a panel whose `type` matches that block's panel type, and that at most one typed config block is present on any panel. This exclusivity requirement now applies to `ml_anomaly_swimlane_config` and `ml_single_metric_viewer_config` in addition to all previously supported typed config blocks:

- `ml_anomaly_swimlane_config` SHALL only be valid when `type = "ml_anomaly_swimlane"` and SHALL conflict with all other typed panel config blocks and with practitioner-authored `config_json`.
- `ml_single_metric_viewer_config` SHALL only be valid when `type = "ml_single_metric_viewer"` and SHALL conflict with all other typed panel config blocks and with practitioner-authored `config_json`.

#### Scenario: ml_anomaly_swimlane_config rejected for non-ml_anomaly_swimlane panel

- GIVEN a panel with `type = "markdown"` and `ml_anomaly_swimlane_config` set
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic

#### Scenario: ml_single_metric_viewer_config rejected for non-ml_single_metric_viewer panel

- GIVEN a panel with `type = "markdown"` and `ml_single_metric_viewer_config` set
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic

#### Scenario: ml_anomaly_swimlane_config conflicts with other typed blocks

- GIVEN a panel with `type = "ml_anomaly_swimlane"` and both `ml_anomaly_swimlane_config` and any other typed config block set
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic

### Requirement: Dashboard root schema API naming (REQ-036)

The resource SHALL expose dashboard-level time selection, refresh, and query using nested attribute objects whose names mirror the Kibana Dashboard API JSON: `time_range` (`from`, `to`, optional `mode`), `refresh_interval` (`pause`, `value`), and `query` (`language`, restricted to `kql` or `lucene`, with exactly one of `text` or `json` for the query expression).

The resource SHALL expose dashboard `options` with the API-aligned flags `auto_apply_filters` and `hide_panel_borders` in addition to the existing option fields.

#### Scenario: Query union uses text branch

- GIVEN `query = { language = "kql" text = "http.response.status_code:200" }`
- WHEN the provider builds the create or update request body
- THEN it SHALL set the API query expression from `query.text` and SHALL set `query.language` from `query.language`

#### Scenario: Query union uses json branch

- GIVEN `query = { language = "kql" json = jsonencode({ ... }) }`
- WHEN the provider builds the create or update request
- THEN it SHALL set the API query expression from `query.json` and SHALL reject configurations where both `text` and `json` are set, or where neither is set

#### Scenario: Options include new flags

- GIVEN `options { hide_panel_borders = true auto_apply_filters = false }`
- WHEN create or update runs
- THEN the provider SHALL include those fields in the API `options` object when known

### Requirement: Create and update request mapping (REQ-007)

On create and update, the resource SHALL map Terraform state to the dashboard API request body using `title`, `description`, nested `time_range`, nested `refresh_interval`, nested `query`, tags, options, panels, and sections when those values are known. `access_control` SHALL be sent on create when known. The current regenerated Kibana `PUT /dashboards/{id}` request body does not expose `access_control`, so updates SHALL preserve prior `access_control` state but SHALL NOT claim to mutate it through the dashboard update request until the API surface supports that field. Query mapping SHALL send `query.text` as the string form of the API query expression and `query.json` as the JSON-object form of the same expression field. On create, the provider SHALL call the `POST /dashboards` API and let Kibana assign the dashboard id. If conversion of query or panel data fails, the operation SHALL return diagnostics and SHALL NOT proceed with the dashboard API call. After a successful create or update, the resource SHALL read the dashboard back and use that read as the authoritative final state; if the dashboard cannot be read back, the operation SHALL fail.

#### Scenario: Post-apply authoritative read

- GIVEN create or update succeeds
- WHEN the provider finalizes state
- THEN it SHALL re-read the dashboard and SHALL fail if the dashboard cannot be retrieved

### Requirement: Read behavior and missing-resource handling (REQ-008)

On refresh, the resource SHALL parse the composite `id`, read the dashboard from Kibana, and repopulate state from the API response. If Kibana returns not found, the resource SHALL remove itself from Terraform state. When a dashboard is found, the resource SHALL map title, description, nested `time_range`, nested `refresh_interval`, nested `query`, tags, options, access control, top-level panels, and sections back into state.

#### Scenario: Dashboard removed outside Terraform

- GIVEN a dashboard recorded in Terraform state
- WHEN refresh runs and Kibana returns not found
- THEN the resource SHALL remove the dashboard from state

#### Scenario: Read maps nested query and time_range

- GIVEN a successful refresh after create with `query = { language = "kql" text = "foo" }` and `time_range = { from = "now-7d" to = "now" }`
- WHEN state is repopulated from the GET response
- THEN the resource SHALL set `query.language`, `query.text`, and `time_range.from` / `time_range.to` from the API payload

The provider SHALL apply intent-preserving null normalization to the root-level `description` attribute on read. When the Kibana API returns an empty-string `description` (`""`) and the prior Terraform plan/state had `description` as null (i.e., the practitioner did not set `description`), the provider SHALL store `null` in state — not `""`. When the prior Terraform plan/state had `description` as `""` (i.e., the practitioner explicitly set `description = ""`), the provider SHALL preserve `""` in state. When the API returns a non-empty `description`, the provider SHALL store that value unchanged.

This normalization fixes a Kibana 9.5 behavior change: previously the API omitted `description` when none was supplied; from 9.5 onward it returns `""`. Without this fix, practitioners who omit `description` see "Provider produced inconsistent result after apply" with the message `.description: was null, but now cty.StringVal("")`.

#### Scenario: Omitted description normalizes to null on read

- GIVEN a dashboard configured without `description` (null in Terraform state/plan)
- AND the Kibana API returns `description: ""`
- WHEN the provider reads the dashboard
- THEN state SHALL contain `description = null`

#### Scenario: Explicit empty description preserved on read

- GIVEN a dashboard configured with `description = ""`
- AND the Kibana API returns `description: ""`
- WHEN the provider reads the dashboard
- THEN state SHALL contain `description = ""`

#### Scenario: Non-empty description preserved unchanged

- GIVEN a dashboard configured with `description = "My dashboard"`
- AND the Kibana API returns `description: "My dashboard"`
- WHEN the provider reads the dashboard
- THEN state SHALL contain `description = "My dashboard"`

### Requirement: State preservation for fields Kibana omits or defaults (REQ-009)

When Kibana omits or defaults fields on read, the resource SHALL preserve prior Terraform intent to avoid inconsistent results and spurious drift where the implementation supports that behavior. The resource preserves the prior `time_range.mode` value already held in state or plan instead of overwriting it from read-back when the GET response does not supply a usable mode. When the GET dashboard API does not supply a usable `access_control.access_mode` value, the resource SHALL clear `access_control` in Terraform state rather than leaving a stale prior value behind. When the options block was omitted in Terraform and Kibana materializes only the default dashboard options matching the implementation's `isDashboardOptionsDefaultSet` helper (including `auto_apply_filters` and `hide_panel_borders` at their API defaults when applicable), the resource SHALL keep the `options` block null in state. When a section's prior `collapsed` value was null and Kibana returns `false`, the resource SHALL preserve null rather than forcing `false` into state.

For panel reads, the provider SHALL seed each panel from prior practitioner intent before finalizing state: from the prior plan on the post-create and post-update read-back, and from prior state on refresh. After that seed, it SHALL apply panel-type-specific alignment so Kibana-injected defaults or omitted optional values do not overwrite practitioner intent. This alignment includes preserving configured titles and descriptions when the API returns blank values, preserving ES|QL control `esql_query`, `title`, and `available_options` when the API omits them, preserving raw `config_json` when the read-back only differs by omitted optional `filters` or `query` keys, and preserving semantically equivalent optional JSON defaults such as `rank_by` in metric and tagcloud configurations.

The resource models only the currently supported Terraform subset of dashboard fields. Fields present in the Kibana dashboard API but not modeled by this resource — for example top-level `project_routing` — are outside this resource contract (see REQ-037 for `filters` and REQ-038 for `pinned_panels`).

The provider SHALL treat an API-returned `""` for `description` as semantically equivalent to an omitted field when prior plan/state had `description` null, restoring null in state rather than propagating the API-echoed empty string. This is an instance of REQ-009 null-preservation applied to the dashboard root `description`. This SHALL be consistent with the null/empty-string normalization already applied to XY chart `fitting.type`, `fitting.end_value`, and panel-level `time_range`.

#### Scenario: Empty-string description treated as null for null-intent practitioners

- GIVEN a practitioner has never set `description` on a dashboard (prior state: null)
- AND Kibana 9.5 returns `description: ""` on a subsequent read or post-apply read-back
- WHEN the provider applies REQ-009 null-preservation to `description`
- THEN state SHALL contain `description = null` and no drift SHALL be reported on the next plan

#### Scenario: Options omitted in config

- GIVEN Terraform configuration omitted the `options` block
- WHEN Kibana read-back contains only the dashboard option defaults
- THEN the resource SHALL keep `options` unset in state

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

The list of panel types that SHALL NOT accept practitioner-authored `config_json` (REQ-010) is extended to include `apm_service_map`. The `apm_service_map` panel type SHALL be managed exclusively through the `apm_service_map_config` block. This extension follows the same enforcement pattern as existing entries in REQ-044A (the registry-driven simple panel handler architecture).
#### Scenario: apm_service_map_config routed by type discriminant

- GIVEN a dashboard API response containing a panel with `"type": "apm_service_map"` and a non-empty config object
- WHEN the resource reads the dashboard
- THEN the provider SHALL populate `apm_service_map_config` in state from the API response
- AND SHALL NOT fall back to `config_json` for that panel

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

### Requirement: Markdown panel behavior (REQ-012)

For `type = "markdown"` panels, the resource SHALL accept a `markdown_config` block whose shape mirrors the API's by-value/by-reference union: `markdown_config = object({ by_value = object({ content, settings, description, hide_title, title, hide_border }), by_reference = object({ ref_id, description, hide_title, title, hide_border }) })`. Exactly one of `by_value` or `by_reference` SHALL be set; setting both or neither SHALL produce an error diagnostic at plan time.

The `by_value` block SHALL require `content` (string) and `settings` (nested object). The `settings` block SHALL accept `open_links_in_new_tab` (bool, optional; when unset, Kibana applies its default of `true`). The `by_value` block SHALL also accept optional `description`, `hide_title`, `title`, and `hide_border`.

The `by_reference` block SHALL require `ref_id` (string) — the unique identifier of an existing markdown library item — and SHALL accept optional `description`, `hide_title`, `title`, and `hide_border`. The resource SHALL NOT validate that the library item exists at plan time; runtime errors from Kibana surface as today.

On write, the resource SHALL build either the `KbnDashboardPanelTypeMarkdownConfig0` (by-value) or `KbnDashboardPanelTypeMarkdownConfig1` (by-reference) API payload according to which sub-block is set. On read, the resource SHALL detect which API branch was returned and populate the matching sub-block, leaving the other branch null. As REQ-010 already specifies, when a markdown panel is managed through `config_json` only, the resource SHALL preserve that JSON-only representation instead of populating typed `markdown_config` sub-blocks.

#### Scenario: By-value markdown panel round-trip

- GIVEN a panel with `type = "markdown"` and `markdown_config = { by_value = { content = "# hi", settings = { open_links_in_new_tab = false }, hide_border = true } }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the same `by_value` shape, `by_reference` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: By-reference markdown panel round-trip

- GIVEN an existing markdown library item with id `md-lib-1` and a panel with `markdown_config = { by_reference = { ref_id = "md-lib-1", title = "shared note" } }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the same `by_reference` shape, `by_value` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: Both sub-blocks set

- GIVEN a panel with both `markdown_config.by_value` and `markdown_config.by_reference` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one sub-block must be set

#### Scenario: Neither sub-block set

- GIVEN a panel with `markdown_config = {}` (no `by_value` and no `by_reference`)
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one sub-block must be set

#### Scenario: open_links_in_new_tab unset preserves API default

- GIVEN a `by_value` markdown panel whose `settings.open_links_in_new_tab` is unset
- WHEN create runs and the API stores Kibana's default `true`
- THEN state SHALL keep `settings.open_links_in_new_tab` null (REQ-009 null-preservation)

#### Scenario: config_json continues to manage markdown panels

- GIVEN a panel with `type = "markdown"` configured through `config_json` only
- WHEN create, update, or read runs
- THEN the resource SHALL preserve the JSON-only representation per REQ-010 and SHALL NOT populate typed `markdown_config` sub-blocks

### Requirement: XY chart panel behavior and typed `vis` `time_range` (REQ-013)

For **typed** `vis` panels (those built through the provider's typed `*_config` blocks and the shared typed visualization write path, not panels managed solely through raw `config_json`), the resource SHALL expose `time_range` as an optional flat sibling attribute on every typed Lens chart block (`xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, `waffle_config`). The attribute SHALL match the dashboard-level `time_range` shape: required `from` (string), required `to` (string), and optional `mode` enum (`absolute` | `relative`).

When the chart-level `time_range` is null in configuration and state, the provider SHALL omit `time_range` from the API payload entirely. The provider SHALL NOT inherit the dashboard-level `time_range` and SHALL NOT use any hardcoded fallback window. Kibana will apply its own default (global dashboard time range) for panels with no panel-level override.

When the chart-level `time_range` is set in configuration, the provider SHALL pass the configured values to the API verbatim, overriding the dashboard-level value for that panel only.

The `vis_config.by_reference` block SHALL expose `time_range` as an **optional** attribute (same shape: required `from`, required `to`, optional `mode`). When `time_range` is null in `by_reference` configuration, the provider SHALL omit it from the API payload. When set, the provider SHALL send it verbatim.

For XY chart `vis` panels specifically, the resource SHALL require `axis`, `decorations`, `fitting`, `legend`, and at least one `layers` entry. The axis object SHALL use `x`, optional primary `y`, and optional secondary `y2`; `axis.x.domain_json` SHALL represent the X-axis domain, and each configured Y axis SHALL require `domain_json`. Each layer SHALL represent either a data layer or a reference-line layer, not both. **`query` SHALL be optional** on the XY chart schema so that ES|QL XY panels (which carry no `query` in the API) are valid without a dummy query block.

REQ-025 governs raw `config_json` `vis` panels; the typed-vs-raw distinction is unchanged.

#### Scenario: Typed `vis` write omits time_range when chart time_range is null

- GIVEN a typed `vis` panel on create or update whose chart-level `time_range` is null in configuration
- AND the dashboard-level `time_range` is `{ from = "now-7d", to = "now" }`
- WHEN the provider builds the visualization payload through the typed converter path
- THEN it SHALL NOT include `time_range` in the API payload for that panel

#### Scenario: Typed `vis` write uses configured chart-level time_range when set

- GIVEN a typed `vis` panel on create or update whose chart-level `time_range` is set to `{ from = "now-30d", to = "now-1d" }` in configuration
- AND the dashboard-level `time_range` is `{ from = "now-7d", to = "now" }`
- WHEN the provider builds the visualization payload through the typed converter path
- THEN it SHALL set `time_range` on the API payload to the chart-level value `{ from = "now-30d", to = "now-1d" }`

#### Scenario: by_reference write omits time_range when not configured

- GIVEN a `vis_config.by_reference` panel on create or update where `time_range` is null in configuration
- WHEN the provider builds the API payload
- THEN it SHALL NOT include `time_range` in the by-reference config payload

#### Scenario: by_reference write sends time_range when configured

- GIVEN a `vis_config.by_reference` panel on create or update where `time_range` is set to `{ from = "now-7d", to = "now" }`
- WHEN the provider builds the API payload
- THEN it SHALL include `time_range = { from = "now-7d", to = "now" }` in the by-reference config payload

#### Scenario: Read preserves null time_range when API returns none

- GIVEN a typed `vis` panel where `time_range` is null in Terraform state
- AND the Kibana API returns no `time_range` field for that panel on read
- WHEN the provider processes the read response
- THEN it SHALL keep `time_range` as null in state (no drift)

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

### Requirement: Datatable panel behavior (REQ-016)

For datatable `vis` panels, the resource SHALL support exactly one of the `no_esql` and `esql` nested configurations. The non-ES|QL branch SHALL require `query` and SHALL map `data_source_json`, `density`, `metrics`, optional `rows`, optional `split_metrics_by`, optional `sort_by_json`, optional `paging`, optional `filters`, and optional `ignore_global_filters` / `sampling`. The ES|QL branch SHALL map the equivalent table configuration without a `query` block.

#### Scenario: Datatable mode blocks are exclusive

- GIVEN a datatable panel configuration
- WHEN Terraform validates the resource schema
- THEN it SHALL reject configurations that set both `no_esql` and `esql`

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

### Requirement: Region map panel behavior (REQ-020)

For region-map `vis` panels, the resource SHALL require `data_source_json`, `metric_json`, and `region_json`. It SHALL use the implementation's query-based mode selection: when `query.query` is known it SHALL use the non-ES|QL branch, otherwise it SHALL use the ES|QL branch. Region-map metric normalization SHALL apply the region-map metric defaults used by the implementation.

#### Scenario: Region map without known query

- GIVEN a region-map panel whose `query` does not contain a known query string
- WHEN the provider converts the panel
- THEN it SHALL use the ES|QL branch

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

### Requirement: Metric chart panel behavior (REQ-022)

For metric-chart `vis` panels, the resource SHALL map the provider's two metric-chart variants: the non-ES|QL branch when `query` is present, and the ES|QL branch when the implementation detects the query as absent or empty. It SHALL require `data_source_json` and `metrics`, and it MAY accept `breakdown_by_json`, `filters`, `ignore_global_filters`, and `sampling`. Each metric `config_json` SHALL use the shared visualization metric default normalization used by the implementation.

#### Scenario: Metric chart ES|QL read-back

- GIVEN a metric-chart panel in the ES|QL variant
- WHEN the provider reads it back from Kibana
- THEN it SHALL not repopulate a non-ES|QL `query` block into state

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

### Requirement: Legacy metric panel behavior (REQ-024)

For legacy-metric `vis` panels, the resource SHALL require `data_source_json.type` to be one of `data_view_reference` or `data_view_spec`. For those supported data source types it SHALL use the non-ES|QL branch and SHALL require `query` and `metric_json`. Data source types such as `esql` or `table` are outside the supported contract and SHALL be rejected with an error diagnostic. Legacy metric semantic equality SHALL normalize the implementation's legacy metric defaults, including absent `filters` and the metric/format defaults applied by the resource.

#### Scenario: Unsupported legacy metric data source type

- GIVEN a legacy metric panel whose `data_source_json.type` is `esql`
- WHEN the provider converts the panel
- THEN it SHALL return an error diagnostic stating that supported data source types are `data_view_reference` and `data_view_spec`

### Requirement: Raw `config_json` panel behavior (REQ-025)

The provider SHALL preserve the unknown-panel round-trip behavior specified in REQ-025 on Kibana
9.5 and later. The acceptance test `TestAccResourceDashboardUnknownPanel_lensDashboardApp` and its
helper `replaceDashboardPanelWithLensDashboardApp` SHALL be removed from
`internal/kibana/dashboard/acc_unknown_panels_test.go` because the test fixture type
(`lens-dashboard-app`) is no longer accepted by the Kibana 9.5+ PUT API.

The provider SHALL continue to satisfy the unknown-panel preservation contract. The following unit
tests in `internal/kibana/dashboard/models_panels_test.go` SHALL remain as the primary test
coverage and are not modified by this change:

- `Test_unknownPanelRoundTrip`
- `Test_mapPanelsFromAPI` / `"unknown panel type preserves id, grid, and config"`
- `Test_panelsToAPI` / `"unknown panel type replays config_json"`

#### Scenario: config_json preserved for unrecognized panel types at read time

- GIVEN a panel with an unknown or unrecognized `type` value (e.g. `custom_unknown_panel`)
- WHEN the provider reads such a panel back from the Kibana API via `dashboardMapPanelsFromAPI`
- THEN the provider SHALL use the unknown-panel fallback and SHALL populate `config_json` in state with the verbatim API config
- AND SHALL NOT return an error diagnostic for the unrecognized panel type
- AND the round-trip through `dashboardPanelsToAPI` SHALL produce API JSON semantically identical to the original input

#### Scenario: Unknown panel type replays config_json on write

- GIVEN a panel model with an unknown `type` and a non-null `config_json`
- WHEN the provider serialises the panel to the API payload via `dashboardPanelsToAPI`
- THEN the API panel `config` SHALL contain the verbatim JSON from `config_json`
- AND the panel `type`, `id`, and `grid` SHALL be preserved unchanged

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

The `esql_control` panel type is a standalone control panel, not a `vis` visualization. It does not reference a saved object, and its configuration is fully inline in the dashboard document. As a result, none of the typed visualization converters, typed visualization time-range injection, or visualization metric default normalization SHALL apply to `esql_control` panels.

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

When a panel entry sets `type = "range_slider_control"`, the resource SHALL accept a `range_slider_control_config` block with the following structure. The block MUST contain exactly one of two mutually exclusive nested blocks: `by_field` or `by_esql`. Having both or neither SHALL produce an error diagnostic at plan time.

#### `by_field` block (Field variant)

The `by_field` nested block represents a range slider sourced from a Kibana data view field:

- `data_view_id` (required, string) — the ID of the data view the slider targets.
- `field_name` (required, string) — the numeric field within the data view.
- `title` (optional, string) — human-readable label displayed above the slider.
- `use_global_filters` (optional, bool) — whether the panel respects dashboard-level global filters.
- `ignore_validations` (optional, bool) — suppresses validation errors from the control during intermediate states.
- `value` (optional, list of string) — the initial min/max range as a 2-element list `[min, max]`; the list MUST contain exactly 2 elements when set.
- `step` (optional, number) — the step size for each increment of the slider (stored as float32 to match the API).

On write, the provider SHALL set `values_source = "field"` automatically; this field SHALL NOT be exposed.

#### `by_esql` block (ES|QL variant)

The `by_esql` nested block represents a range slider sourced from an ES|QL query:

- `esql_query` (required, string) — the ES|QL query that produces the min/max range values.
- `values_source` (required, string) — must be `"esql_query"`. Any other value SHALL produce an error diagnostic at plan time.
- `title` (optional, string) — same as `by_field`.
- `use_global_filters` (optional, bool) — same as `by_field`.
- `ignore_validations` (optional, bool) — same as `by_field`.
- `value` (optional, list of string) — same as `by_field` (2-element list constraint applies).
- `step` (optional, number) — same as `by_field`.

Null-preservation semantics (REQ-009) apply to optional boolean attributes (`use_global_filters`, `ignore_validations`) within both branches. On import, the provider SHALL populate the branch-specific required identifiers from the API response: for `by_field`, `data_view_id` and `field_name`; for `by_esql`, `esql_query` and `values_source`. Optional booleans SHALL be left null.

#### Mutual exclusion and conflict guards

- Exactly one of `by_field` or `by_esql` MUST be set in `range_slider_control_config`.
- `range_slider_control_config` remains mutually exclusive with all other typed panel config blocks and with `config_json`.

#### State migration (v0 → v1)

The same `ResourceWithUpgradeState` upgrader (described in REQ-027 above) SHALL also rewrite existing v0 `range_slider_control_config` flat attributes (`data_view_id`, `field_name`, `title`, `use_global_filters`, `ignore_validations`, `value`, `step`) under a `by_field {}` object.

#### Scenarios

##### Scenario: Field variant round-trip

- GIVEN a `range_slider_control` panel with `by_field = { data_view_id = "orders-view", field_name = "price" }`
- WHEN the provider creates the dashboard and reads it back
- THEN `data_view_id` and `field_name` SHALL be present under `range_slider_control_config.by_field` and a subsequent plan SHALL show no changes

##### Scenario: ES|QL variant round-trip

- GIVEN a `range_slider_control` panel with `by_esql = { esql_query = "FROM orders | STATS ...", values_source = "esql_query" }`
- WHEN the provider creates the dashboard and reads it back
- THEN `esql_query` and `values_source` SHALL be present under `range_slider_control_config.by_esql` and a subsequent plan SHALL show no changes

##### Scenario: Invalid value list length

- GIVEN a `range_slider_control_config` block with `by_field = { ..., value = ["10"] }`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return a validation diagnostic stating that `value` must contain exactly 2 elements

##### Scenario: State upgrade from v0 flat to v1 by_field

- GIVEN a Terraform state containing `range_slider_control_config` with flat attributes (v0 schema: `data_view_id`, `field_name`, etc. at the config root)
- WHEN the provider with the updated schema is applied
- THEN the state upgrader SHALL rewrite the flat attributes under `by_field` and the resulting v1 state SHALL be equivalent to configuring `by_field { data_view_id = ..., field_name = ..., ... }`

##### Scenario: Both branches rejected

- GIVEN a `range_slider_control_config` block that sets both `by_field` and `by_esql`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating the two branches are mutually exclusive

### Requirement: SLO overview panel behavior (REQ-030)

The resource SHALL support `type = "slo_overview"` panels through the typed `slo_overview_config` block. The block SHALL carry exactly one of two mutually exclusive nested blocks: `single` (for single-SLO overview) or `groups` (for grouped SLO overview). On write, the provider SHALL use the presence of the `single` block to select the `slo-single-overview-embeddable` API embeddable type, and the presence of the `groups` block to select the `slo-group-overview-embeddable` API embeddable type. The `overview_mode` discriminant field in the API payload SHALL be set to `"single"` or `"groups"` accordingly and SHALL NOT be exposed as a direct Terraform attribute.

For `single` mode:

- `slo_id` SHALL be required and SHALL be sent as the SLO identifier in the API payload.
- `slo_instance_id` SHALL be optional. When configured, it SHALL be sent; when not configured, the field SHALL be omitted from the write payload. On read, if the prior state value was null and Kibana returns `"*"`, the provider SHALL preserve null rather than force `"*"` into state.
- `remote_name` SHALL be optional and SHALL be sent when configured.

For `groups` mode:

- All fields SHALL be optional.
- `group_filters` SHALL be an optional nested block with the following attributes:
  - `group_by`: optional string, enum-validated as one of `"slo.tags"`, `"status"`, `"slo.indicator.type"`, `"_index"`.
  - `groups`: optional list of strings with a maximum of 100 entries.
  - `kql_query`: optional string.
  - `filters_json`: optional normalized JSON string representing the AS-code filter array; the provider SHALL normalize this field for semantic equality on refresh.

Both modes SHALL support the following shared optional display attributes within their respective nested blocks:

- `title`: optional string.
- `description`: optional string.
- `hide_title`: optional bool.
- `hide_border`: optional bool.

Both modes SHALL support a `drilldowns` optional list of objects with the following attributes:

- `url`: required string.
- `label`: required string.
- `trigger`: required string.
- `type`: required string.
- `encode_url`: optional bool.
- `open_in_new_tab`: optional bool.

On read, the provider SHALL reconstruct the `single` or `groups` sub-block from the API payload's `overview_mode` field. On read, if Kibana omits `hide_border` or any optional display field, the provider SHALL preserve the prior state value rather than forcing a default.

#### Scenario: Single-mode SLO overview panel write and read

- GIVEN a panel with `type = "slo_overview"` and a `single` block with `slo_id = "my-slo-id"` and `slo_instance_id = "instance-1"`
- WHEN the provider builds the API request
- THEN it SHALL send the `slo-single-overview-embeddable` payload with `overview_mode = "single"`, `slo_id = "my-slo-id"`, and `slo_instance_id = "instance-1"`
- AND WHEN the provider reads the panel back
- THEN it SHALL populate `single.slo_id` and `single.slo_instance_id` from the API response

#### Scenario: Groups-mode SLO overview panel with group_filters

- GIVEN a panel with `type = "slo_overview"` and a `groups` block with `group_filters.group_by = "status"` and `group_filters.kql_query = "slo.name: my-*"`
- WHEN the provider builds the API request
- THEN it SHALL send the `slo-group-overview-embeddable` payload with `overview_mode = "groups"` and the configured `group_filters`
- AND WHEN the provider reads the panel back
- THEN it SHALL populate `groups.group_filters.group_by` and `groups.group_filters.kql_query` from the API response

#### Scenario: slo_instance_id null preservation

- GIVEN a panel with `type = "slo_overview"` in `single` mode where `slo_instance_id` was not configured (null in state)
- WHEN the provider reads the panel back and Kibana returns `slo_instance_id = "*"`
- THEN the provider SHALL preserve `slo_instance_id` as null in state rather than updating it to `"*"`

#### Scenario: Invalid slo_overview_config — no sub-block

- GIVEN a panel with `type = "slo_overview"` and an `slo_overview_config` block that contains neither `single` nor `groups`
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: Drilldowns round-trip

- GIVEN a panel with `type = "slo_overview"` in `single` mode with a `drilldowns` entry specifying `url`, `label`, `trigger`, `type`, and `open_in_new_tab = true`
- WHEN the provider builds the API request and reads the panel back
- THEN the `drilldowns` list SHALL reflect the configured values in state

### Requirement: Options list control panel behavior (REQ-027)

When a panel entry sets `type = "options_list_control"`, the resource SHALL accept an `options_list_control_config` block with the following structure. The block MUST contain exactly one of two mutually exclusive nested blocks: `by_field` or `by_esql`. Having both or neither SHALL produce an error diagnostic at plan time.

#### `by_field` block (Field variant)

The `by_field` nested block represents a control sourced from a Kibana data view field:

- `data_view_id` (required, string) — the ID of the Kibana data view the control is tied to.
- `field_name` (required, string) — the name of the field within the data view.
- `title` (optional, string) — human-readable label displayed above the control.
- `use_global_filters` (optional, bool) — whether the control applies the dashboard's global filters to its own query.
- `ignore_validations` (optional, bool) — whether the control skips field-level validation against the data view.
- `single_select` (optional, bool) — when true, only one option may be selected at a time.
- `exclude` (optional, bool) — when true, selected options are used as an exclusion filter rather than an inclusion filter.
- `exists_selected` (optional, bool) — when true, the control filters for documents where the field exists.
- `run_past_timeout` (optional, bool) — when true, the control continues to show results even when the underlying query times out.
- `search_technique` (optional, string) — must be one of `prefix`, `wildcard`, or `exact` when set.
- `selected_options` (optional, list of string) — the initially or persistently selected option values; all values are represented as strings.
- `display_settings` (optional, nested block) — display preferences for the control widget, containing:
  - `placeholder` (string) — placeholder text shown when no option is selected.
  - `hide_action_bar` (bool) — when true, hides the action bar on the control.
  - `hide_exclude` (bool) — when true, hides the exclude toggle.
  - `hide_exists` (bool) — when true, hides the exists filter option.
  - `hide_sort` (bool) — when true, hides the sort control.
- `sort` (optional, nested block) — default sort configuration for the suggestion list, containing:
  - `by` (required, string) — must be one of `_count` or `_key`.
  - `direction` (required, string) — must be one of `asc` or `desc`.

On write, the provider SHALL set `values_source = "field"` automatically on the API payload for the Field branch; this field SHALL NOT be exposed as a Terraform attribute.

#### `by_esql` block (ES|QL variant)

The `by_esql` nested block represents a control sourced from an ES|QL query:

- `esql_query` (required, string) — the ES|QL query that produces the available option values.
- `values_source` (required, string) — the source discriminator; MUST be `"esql_query"`. Any other value SHALL produce an error diagnostic at plan time.
- `title` (optional, string) — same as `by_field`.
- `use_global_filters` (optional, bool) — same as `by_field`.
- `ignore_validations` (optional, bool) — same as `by_field`.
- `single_select` (optional, bool) — same as `by_field`.
- `exclude` (optional, bool) — same as `by_field`.
- `exists_selected` (optional, bool) — same as `by_field`.
- `run_past_timeout` (optional, bool) — same as `by_field`.
- `search_technique` (optional, string) — same as `by_field` (must be one of `prefix`, `wildcard`, or `exact` when set).
- `selected_options` (optional, list of string) — same as `by_field`.
- `display_settings` (optional, nested block) — same structure as `by_field`.
- `sort` (optional, nested block) — same structure as `by_field`.

#### Null-preservation and import semantics

During import (no prior state), the provider SHALL populate the branch-specific required identifiers from the API response: for `by_field`, populate `data_view_id` and `field_name`; for `by_esql`, populate `esql_query` and `values_source`. In both branches, `title`, `search_technique`, `selected_options`, and `display_settings` SHALL be populated where present in the API response; optional booleans and `sort` SHALL be left null.

#### Mutual exclusion and conflict guards

- Exactly one of `by_field` or `by_esql` MUST be set in `options_list_control_config`.
- `options_list_control_config` SHALL remain mutually exclusive with all other typed panel config blocks and with `config_json` (unchanged from existing REQ-027).

#### State migration (v0 → v1)

The change from a flat `options_list_control_config` schema to the two-branch nested schema constitutes a schema version increment. The resource SHALL implement a Plugin Framework `ResourceWithUpgradeState` upgrader that rewrites existing v0 state by moving all flat `options_list_control_config` attributes under a `by_field {}` object. The upgrader SHALL handle both in-grid `panels[]` entries and `pinned_panels` entries.

#### Scenarios

##### Scenario: Field variant round-trip

- GIVEN a panel with `type = "options_list_control"` whose `options_list_control_config` sets `by_field = { data_view_id = "logs-view", field_name = "service.name", search_technique = "prefix", single_select = true }`
- WHEN the provider creates the dashboard and reads it back
- THEN all configured attributes SHALL be present in state under `options_list_control_config.by_field` and a subsequent plan SHALL show no changes

##### Scenario: ES|QL variant round-trip

- GIVEN a panel with `type = "options_list_control"` whose `options_list_control_config` sets `by_esql = { esql_query = "FROM logs | STATS ...", values_source = "esql_query", title = "Service" }`
- WHEN the provider creates the dashboard and reads it back
- THEN `options_list_control_config.by_esql.esql_query` and `values_source` SHALL be present in state and a subsequent plan SHALL show no changes

##### Scenario: Field variant null-preservation on import

- GIVEN an existing dashboard whose options-list control is a field-backed control with Kibana server-side defaults for `use_global_filters`, `ignore_validations`, `exclude`, `exists_selected`, `run_past_timeout`, and `sort`
- WHEN the provider imports the dashboard resource
- THEN `data_view_id` and `field_name` SHALL be populated under `by_field` in state
- AND optional booleans and `sort` SHALL remain null in state
- AND a subsequent plan against a configuration that omits them SHALL show no changes

##### Scenario: Both branches rejected

- GIVEN an `options_list_control_config` block that sets both `by_field` and `by_esql`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating the two branches are mutually exclusive

##### Scenario: Neither branch rejected

- GIVEN an `options_list_control_config` block with neither `by_field` nor `by_esql` set
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating exactly one branch must be configured

##### Scenario: Invalid values_source on by_esql

- GIVEN `by_esql = { esql_query = "...", values_source = "field" }`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating `values_source` must be `"esql_query"`

##### Scenario: State upgrade from v0 flat to v1 by_field

- GIVEN a Terraform state that contains `options_list_control_config` with flat attributes (v0 schema: `data_view_id`, `field_name`, etc. at the config root)
- WHEN the provider with the updated schema is applied
- THEN the state upgrader SHALL rewrite the flat attributes under `by_field` and the resulting v1 state SHALL be equivalent to configuring `by_field { data_view_id = ..., field_name = ..., ... }`

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

### Requirement: Dashboard-level saved filters round-trip (REQ-037)

The resource SHALL expose dashboard-level saved filters at the dashboard root as `filters = list(object({ filter_json = string }))`. Each `filter_json` value SHALL be a JSON-encoded object that conforms to the Kibana Dashboard API `kbn-as-code-filters-schema_*` discriminated union (or DSL/spatial filter shape) and SHALL be normalized for diff comparison using the same semantic JSON equality applied to per-panel `filter_json` values.

On create and update, the resource SHALL include each filter's JSON object in the dashboard API request `filters` array in the order given in configuration. On read, the resource SHALL repopulate `filters` from the API response in the order returned. When `filters` is unset in configuration and the API returns either no `filters` field or an empty list, the resource SHALL keep the Terraform attribute unset (not coerce it to an empty list).

#### Scenario: Saved filters round-trip across create and read

- GIVEN `filters = [{ filter_json = jsonencode({ operator = "is", field = "host.name", value = "web-01" }) }]`
- WHEN create runs and the post-apply read returns the same filter
- THEN state SHALL contain a single `filters` entry whose `filter_json` is semantically equal to the configured JSON

#### Scenario: Filter JSON normalization avoids spurious diffs

- GIVEN a `filter_json` value whose key ordering or whitespace differs from the API response
- WHEN refresh runs
- THEN the provider SHALL not produce a diff for that filter

#### Scenario: Unset filters preserved when API returns empty

- GIVEN `filters` is unset in configuration
- WHEN read runs and the API returns no `filters` field or `filters: []`
- THEN the Terraform `filters` attribute SHALL remain unset rather than being set to an empty list

#### Scenario: Multiple filters preserved in order

- GIVEN `filters = [a, b, c]` with three distinct filter JSON values
- WHEN create or update runs and the post-apply read returns the same three filters in the same order
- THEN state SHALL contain those three filters in the same order

### Requirement: Dashboard-level pinned controls round-trip (REQ-038)

The resource SHALL expose dashboard-level pinned controls at the dashboard root as `pinned_panels = list(object({ type, options_list_control_config, range_slider_control_config, time_slider_control_config, esql_control_config }))`. Each entry SHALL declare exactly one of the four typed `*_control_config` blocks, and the chosen block SHALL match the entry's `type` value (for example `type = "options_list_control"` requires `options_list_control_config` and forbids the other three). `pinned_panels` entries SHALL NOT include a `grid` attribute, and the resource SHALL NOT populate one on read.

The four `*_control_config` block schemas SHALL be identical to the schemas used for the same control types under `panels[]`; any future change to those typed schemas applies to both placements.

On create and update, the resource SHALL include `pinned_panels` in the dashboard API request body in the order given in configuration. On read, the resource SHALL repopulate `pinned_panels` from the API `pinned_panels` array in the order returned. When `pinned_panels` is unset in configuration and the API returns an empty list (Kibana's default), the resource SHALL keep the Terraform attribute unset.

#### Scenario: Pinned options-list control round-trip

- GIVEN `pinned_panels = [{ type = "options_list_control", options_list_control_config = { ... } }]`
- WHEN create runs and the post-apply read returns the same control
- THEN state SHALL contain a single `pinned_panels` entry with the matching typed config and no `grid`

#### Scenario: Mismatched type and config block

- GIVEN a `pinned_panels` entry with `type = "range_slider_control"` and only `options_list_control_config` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the typed block does not match `type`

#### Scenario: Multiple typed config blocks set on one entry

- GIVEN a `pinned_panels` entry with both `options_list_control_config` and `range_slider_control_config` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one typed block must be set

#### Scenario: Empty pinned_panels preserved as unset

- GIVEN `pinned_panels` is unset in configuration
- WHEN read runs and the API returns `pinned_panels: []`
- THEN the Terraform `pinned_panels` attribute SHALL remain unset

### Requirement: Lens chart presentation fields on typed `vis` panels under `vis_config.by_value` (REQ-039)

For every typed Lens chart block reachable at `panels[].vis_config.by_value` on `type = "vis"` panels (`xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, `waffle_config`), the resource SHALL expose the following optional attributes on that chart block that mirror the corresponding fields on the Kibana chart-root API schemas:

- `hide_title` (bool): when set, the API payload SHALL include `hide_title` on the chart root; when null in state, the payload SHALL omit it.
- `hide_border` (bool): when set, the API payload SHALL include `hide_border` on the chart root; when null in state, the payload SHALL omit it.
- `references_json` (normalized JSON string): when set, the API payload SHALL include `references` on the chart root as the parsed JSON array (`kbn-content-management-utils-referenceSchema[]`); when null in state, the payload SHALL omit it. Read-back SHALL normalize the returned `references` array into the canonical JSON form used by the resource.
- `drilldowns` (typed list of variant sub-blocks per REQ-041): when set, the API payload SHALL include `drilldowns` on the chart root as a typed array conforming to the API discriminated union; when null in state, the payload SHALL omit it.

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

### Requirement: Chart-level `time_range` null-preservation (REQ-040)

The resource SHALL preserve practitioner intent for the chart-level `time_range` block on every typed Lens chart reachable under `panels[].vis_config.by_value.<chart>_config` (for `type = "vis"`), using the same null-preservation pattern as REQ-009 for `time_range.mode`.

When the Kibana API response omits chart-level `time_range` (or returns an empty/zero-valued time range struct), the provider SHALL leave state's chart-level `time_range` as null. When the API returns a populated chart-level `time_range`, the provider SHALL populate state from the API response (subject to the `time_range.mode` null-preservation rule below).

The chart-level `time_range.mode` attribute SHALL follow the same null-preservation rule as the dashboard-level `time_range.mode` in REQ-009: when prior state has `mode = null` and the API response omits or returns no usable mode, state SHALL preserve null rather than overwriting with a default.

#### Scenario: Chart time_range stays null when API omits it

- GIVEN a `vis` panel with a typed Lens chart block under `vis_config.by_value` whose prior state has `time_range = null`
- AND the Kibana API response omits `time_range` on that chart root
- WHEN the provider reads the panel
- THEN state SHALL preserve `time_range = null` on the chart panel

#### Scenario: Chart time_range mode null-preservation

- GIVEN a typed Lens chart panel whose prior state has `time_range = { from = "now-7d", to = "now", mode = null }`
- AND the Kibana API response omits `mode` on the chart-root `time_range`
- WHEN the provider reads the panel
- THEN state SHALL preserve `time_range.mode = null`

### Requirement: Structured drilldown list and per-variant validation (REQ-041)

The resource SHALL validate and serialize the `drilldowns` attribute whenever it is set on any of the following placements:

- chart blocks under `panels[].vis_config.by_value.<chart>_config`, or
- `panels[].vis_config.by_reference`.

For each such list, each list item SHALL be an object containing three mutually-exclusive optional sub-blocks modeling the API discriminated union: `dashboard_drilldown`, `discover_drilldown`, and `url_drilldown`. Each list item SHALL set exactly one variant sub-block; setting zero or multiple variants SHALL produce a plan-time validation error that identifies the offending list item and lists the allowable variants.

The Terraform attribute names SHALL match the implementations above (`dashboard_drilldown`, `discover_drilldown`, `url_drilldown` nested objects). Behavior and serialization rules SHALL be identical for embedded chart drilldowns and by-reference panel drilldowns.

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

### Requirement: `vis_config` block for `vis` panels (REQ-042)

For `type = "vis"` panels, the resource SHALL accept a `vis_config` block with exactly one of `by_value` or `by_reference` sub-blocks set. `vis_config` SHALL be valid only on panels with `type = "vis"`, SHALL be mutually exclusive with all other panel-type configuration blocks, and SHALL be mutually exclusive with panel-level `config_json`.

**`vis_config.by_value`** SHALL accept exactly one of the 12 supported typed Lens chart blocks: `xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, or `waffle_config`. The internal shape and behavior of each chart block SHALL be unchanged except for nesting under `vis_config.by_value` (REQ-013 through REQ-024 continue to describe chart semantics, now reached via `vis_config.by_value.<block_name>`). `vis_config.by_value` SHALL NOT contain a nested `config_json` attribute; raw `vis` configuration SHALL be authored using panel-level `config_json` with `type = "vis"` instead.

**`vis_config.by_reference`** SHALL accept:

- `ref_id` (required string) — saved-object reference name; maps to the API `config.ref_id` field.
- `references_json` (optional normalized JSON string) — array of `{ id, name, type }` saved-object references; maps to the API `config.references` array. A saved Lens visualization reference SHALL typically have a reference whose `name` matches `ref_id`, whose `type` is `lens`, and whose `id` is the saved object ID.
- `time_range` (optional object with `from`, `to`, optional `mode` ∈ `absolute`/`relative`) — omitted from the API payload when null in configuration; sent verbatim when set.
- `title`, `description` (optional strings).
- `hide_title`, `hide_border` (optional booleans).
- `drilldowns` (optional list of structured drilldown blocks per REQ-041).

**On write (create and update):**

For `vis_config.by_value` panels, the resource SHALL convert the typed chart block into the matching generated by-value Lens chart schema and SHALL send the resulting object directly as the panel API `config`. The provider SHALL set the panel discriminator to `vis`.

For `vis_config.by_reference` panels, the resource SHALL set the API `config.ref_id` from `by_reference.ref_id`, set the API `config.time_range` object from `by_reference.time_range`, and include `references`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns` only when their corresponding Terraform attributes are set. The provider SHALL set the panel discriminator to `vis`.

**On read:**

The resource SHALL classify the API `config` JSON object in this order: (1) **By-reference**: if the object omits the by-value chart discriminator (`type` at the top level of the chart config) and has a non-empty `ref_id`, the resource SHALL populate `by_reference` and leave `by_value` unset. (2) **By-value**: otherwise, if the object has a non-empty top-level chart `type`, the resource SHALL populate `by_value.<chart_block>` from the API read using the matching typed chart block. (3) When prior plan or state had `by_reference`, the resource SHALL preserve that prior `by_reference` block per REQ-009 and SHALL NOT silently mode-flip to `by_value`.

#### Scenario: Creation of a typed by-value `vis` panel via `vis_config`

- GIVEN a panel with `type = "vis"` and `vis_config.by_value.xy_chart_config = {...}`
- WHEN the resource is created
- THEN the provider SHALL convert the typed chart block into the API by-value chart schema and send it as the panel `config`
- AND the panel SHALL appear in state with `vis_config.by_value.xy_chart_config` populated, all other by-value chart blocks null, and `vis_config.by_reference` null

#### Scenario: Creation of a by-reference `vis` panel

- GIVEN a panel with `type = "vis"` and:
  - `vis_config.by_reference.ref_id = "panel_0"`
  - `vis_config.by_reference.references_json = "[{\"id\":\"abc-123\",\"name\":\"panel_0\",\"type\":\"lens\"}]"`
  - `vis_config.by_reference.time_range.from = "now-15m"`
  - `vis_config.by_reference.time_range.to = "now"`
- WHEN the resource is created
- THEN the provider SHALL send a panel payload with `config.ref_id`, `config.references`, and `config.time_range` to the Kibana dashboard API with the panel discriminator set to `vis`
- AND the panel SHALL appear in state with `vis_config.by_reference.ref_id = "panel_0"` and `vis_config.by_value` null

#### Scenario: Both `vis_config` sub-blocks set simultaneously

- GIVEN a `vis_config` block with both `by_value` and `by_reference` set
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that `by_value` and `by_reference` are mutually exclusive

#### Scenario: Neither `vis_config` sub-block set

- GIVEN a `vis_config` block with neither `by_value` nor `by_reference` set
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that exactly one of `by_value` or `by_reference` must be set

#### Scenario: Multiple typed by-value chart blocks set simultaneously

- GIVEN a `vis_config.by_value` block with two typed chart blocks set (for example `xy_chart_config` and `metric_chart_config`)
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that exactly one chart block must be set

#### Scenario: No chart block in `vis_config.by_value`

- GIVEN a `vis_config.by_value` block with no typed chart block set
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that exactly one chart block must be set

#### Scenario: `vis_config` rejected for non-vis panel

- GIVEN a panel with `type = "markdown"` and `vis_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: `vis_config` mutually exclusive with panel-level `config_json`

- GIVEN a panel with `type = "vis"` that sets both `vis_config` and panel-level `config_json`
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that `vis_config` and `config_json` are mutually exclusive

#### Scenario: `time_range` optional on `vis_config.by_reference`

- GIVEN a `vis_config.by_reference` block with `ref_id` set and `time_range` omitted
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be accepted without a `time_range` block

#### Scenario: Read-back detects by-reference mode

- GIVEN a managed `vis` panel authored in by-reference mode
- WHEN Kibana returns the panel `config` with `ref_id` and `time_range` and no top-level chart `type`
- THEN the provider SHALL populate `vis_config.by_reference` in state and leave `vis_config.by_value` null
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: Read-back populates typed by-value chart block

- GIVEN a managed `vis` panel authored with `vis_config.by_value.xy_chart_config`
- WHEN Kibana returns the panel `config` with the matching XY chart shape
- THEN the provider SHALL populate `vis_config.by_value.xy_chart_config` in state
- AND SHALL leave the other 11 by-value chart blocks null
- AND SHALL leave `vis_config.by_reference` null

#### Scenario: Raw `vis` panel authored via panel-level `config_json`

- GIVEN a panel with `type = "vis"` and panel-level `config_json = "<raw vis config JSON>"` and no `vis_config` block
- WHEN the resource is created
- THEN the provider SHALL unmarshal `config_json` and send the decoded JSON object directly as the panel API `config`
- AND the panel SHALL appear in state with panel-level `config_json` populated and `vis_config` null

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

### Requirement: Registry-driven simple panel handler architecture preserves dashboard behavior (REQ-044A)

The `elasticstack_kibana_dashboard` resource implementation SHALL support a registry-driven handler architecture for simple panel types while preserving the user-visible dashboard behavior defined elsewhere in this capability. A simple panel type is one whose typed configuration does not require internal Lens chart dispatch or by-value/by-reference composite branching.

For each supported simple panel type, the implementation SHALL provide a dedicated handler responsible for that panel type's schema attribute construction, API-to-state mapping, state-to-API mapping, configuration validation, and any panel-specific state alignment. The registry SHALL be the authoritative source used to assemble simple panel schema attributes, route panel reads from API discriminator to handler, route typed panel writes from configured block to handler, and dispatch panel-specific validation.

This architectural change SHALL preserve existing Terraform-facing behavior for the migrated simple panel types, including schema shape, validation outcomes, null-preservation behavior, pinned-panel behavior where applicable, and API round-tripping.

#### Scenario: Simple panel read routing is resolved through the registry

- GIVEN a dashboard API response containing a supported simple panel type
- WHEN the provider reads the panel
- THEN the provider SHALL resolve the panel type through the handler registry
- AND the matching handler SHALL populate the Terraform panel state for that panel type

#### Scenario: Simple panel write routing is resolved through the registry

- GIVEN a Terraform panel configuration for a supported simple panel type
- WHEN the provider builds the dashboard API request
- THEN the provider SHALL resolve the configured typed panel block through the handler registry
- AND the matching handler SHALL build the panel API payload

#### Scenario: Migrated simple panels preserve prior Terraform behavior

- GIVEN a dashboard that uses a migrated simple panel type
- WHEN the provider plans, applies, reads, and refreshes that dashboard after the handler migration
- THEN the panel SHALL preserve the same user-visible schema and runtime behavior as before the migration

#### Scenario: Pinned control panels continue to round-trip through typed handlers

- GIVEN a dashboard with pinned control panels of a migrated control type
- WHEN the provider reads or writes the dashboard
- THEN pinned panel conversion SHALL delegate through the migrated typed handler path
- AND pinned panel Terraform behavior SHALL remain unchanged

### Requirement: SLO overview constant definition site (REQ-045)

The `panelTypeSloOverview` constant SHALL be defined in `schema.go` alongside all other dashboard panel type constants. It SHALL NOT be defined in a panel-specific model file.

#### Scenario: Constant location

- GIVEN a code review of `schema.go`
- WHEN the reviewer searches for `panelTypeSloOverview`
- THEN the constant SHALL be found in `schema.go` together with `panelTypeSloAlerts`, `panelTypeSloBurnRate`, and other panel type constants

### Requirement: Provider registration (REQ-040)

The `elasticstack_kibana_dashboard` resource SHALL be registered through the provider's standard Plugin Framework resource set returned by `Provider.resources(...)` in `provider/plugin_framework.go`. It SHALL NOT be returned from `Provider.experimentalResources(...)`, and practitioners SHALL NOT be required to set `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` to use the resource.

#### Scenario: Default provider surface includes the resource

- **GIVEN** a released provider build (no `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` override)
- **WHEN** Terraform requests the provider's resource set
- **THEN** `elasticstack_kibana_dashboard` SHALL be present in the resources returned by `Provider.Resources(ctx)`

#### Scenario: Experimental resource set excludes the dashboard resource

- **GIVEN** the provider's experimental Plugin Framework resource set returned by `Provider.experimentalResources(ctx)`
- **WHEN** that set is enumerated
- **THEN** it SHALL NOT contain `dashboard.NewResource`

#### Scenario: Practitioner does not need the experimental opt-in

- **GIVEN** a Terraform configuration declaring `resource "elasticstack_kibana_dashboard" "example" { ... }`
- **WHEN** Terraform plans or applies against a released provider build with `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` unset
- **THEN** the provider SHALL recognize and operate the resource without requiring the environment variable

### Requirement: Dashboard model package extraction preserves behavior (REQ-043)

The `elasticstack_kibana_dashboard` resource implementation SHALL isolate Terraform model structs in `internal/kibana/dashboard/models` while preserving the resource's externally observable behavior. All Terraform model structs used by the dashboard resource and its panel/config submodels SHALL move to the `models` package and SHALL be exported so conversion, validation, schema, and lifecycle code in `internal/kibana/dashboard` can reference them without import cycles.

This extraction SHALL be mechanical only: it SHALL NOT change the Terraform schema, API payload shapes, state alignment rules, plan semantics, import behavior, or read/write behavior described by existing dashboard requirements.

#### Scenario: Existing dashboard configuration remains behaviorally unchanged after model extraction

- GIVEN a dashboard configuration that was accepted before the model extraction
- WHEN the provider plans, applies, reads, and refreshes that configuration after the extraction
- THEN the resource SHALL expose the same schema and SHALL produce the same observable behavior and state transitions as before

#### Scenario: Dashboard logic references exported models package types

- GIVEN dashboard resource implementation code that performs schema, validation, lifecycle, or API conversion work
- WHEN it references dashboard Terraform model structs
- THEN those structs SHALL be imported from `internal/kibana/dashboard/models`
- AND the extraction SHALL avoid introducing Go import cycles between the models package and dashboard logic packages

### Requirement: Lens visualization converter registry preserves typed chart behavior (REQ-045)

The `elasticstack_kibana_dashboard` resource implementation SHALL support a Lens visualization converter registry for typed Lens chart handling while preserving the Terraform-facing behavior of all supported Lens chart blocks. Each supported Lens chart kind SHALL be handled by a dedicated converter that participates in shared Lens chart read, write, defaulting, and state-alignment flows through a common registry.

The registry SHALL be the authoritative source used to classify typed Lens by-value chart payloads, dispatch API-to-state conversion for supported Lens chart kinds, dispatch state-to-API conversion for supported Lens chart kinds, and apply Lens chart defaulting or state-alignment behavior that is specific to a chart kind.

This architectural change SHALL NOT alter the user-visible schema or runtime behavior of existing supported Lens chart blocks under dashboard panel configurations.

#### Scenario: Typed Lens chart read conversion is resolved through the converter registry

- GIVEN a dashboard panel API payload for a supported by-value Lens chart kind
- WHEN the provider reads that panel into Terraform state
- THEN the provider SHALL resolve the chart kind through the Lens converter registry
- AND the matching converter SHALL populate the corresponding typed Lens chart block in state

#### Scenario: Typed Lens chart write conversion is resolved through the converter registry

- GIVEN a Terraform configuration that selects a supported typed Lens chart block
- WHEN the provider builds the dashboard API payload for that panel
- THEN the provider SHALL resolve the configured chart block through the Lens converter registry
- AND the matching converter SHALL build the corresponding by-value Lens chart payload

#### Scenario: Lens chart behavior remains unchanged after converter extraction

- GIVEN a dashboard using a supported typed Lens chart block
- WHEN the provider plans, applies, reads, and refreshes that dashboard after the converter migration
- THEN the typed Lens chart block SHALL preserve the same user-visible schema, validation behavior, defaulting behavior, null-preservation behavior, and API round-tripping as before the migration

### Requirement: Composite Lens panel handlers and `vis_config` naming (REQ-046)

For composite Lens-backed dashboard panel types, the `elasticstack_kibana_dashboard` resource SHALL use dedicated typed handler implementations while preserving the behaviors already defined for those panel types, except where this requirement explicitly changes the Terraform schema name.

For `type = "vis"` panels, the Terraform typed configuration block SHALL be named `vis_config`. The provider SHALL treat `vis_config` as the typed configuration entry point for `vis` panels and SHALL align routing and validation with the panel type discriminator `"vis"`.

The handler architecture for these composite panel types SHALL consume the shared Lens converter registry for by-value Lens chart handling and shared by-reference conversion logic for by-reference handling.

Except for the `viz_config` to `vis_config` rename, this architectural change SHALL preserve the previously defined Terraform-facing behavior for `vis` panels, including read/write semantics, validation, null-preservation, typed by-value chart handling, and by-reference handling.

#### Scenario: `vis` panels use `vis_config` as the typed block name

- GIVEN a dashboard panel with `type = "vis"`
- WHEN Terraform validates or the provider processes the typed panel configuration
- THEN the typed configuration block name SHALL be `vis_config`
- AND routing and validation for that panel SHALL use the `"vis"` panel type contract

#### Scenario: Composite by-value Lens chart handling uses shared registry dispatch

- GIVEN a `vis` panel configured with a supported typed by-value Lens chart block
- WHEN the provider reads or writes that panel
- THEN the provider SHALL dispatch the by-value Lens chart conversion through the shared Lens converter registry

#### Scenario: Composite by-reference handling remains behaviorally unchanged

- GIVEN a `vis` panel configured in by-reference mode
- WHEN the provider plans, applies, reads, and refreshes that panel after the composite handler migration
- THEN the panel SHALL preserve the same by-reference behavior already defined for that panel type

#### Scenario: `viz_config` is no longer the accepted typed block name

- GIVEN a Terraform configuration that uses `viz_config` on a `type = "vis"` panel
- WHEN Terraform validates the configuration against the updated schema
- THEN the configuration SHALL be rejected because the typed block name is `vis_config`

### Requirement: Create method selection (REQ-003a)

The resource's create function SHALL select the HTTP method based on whether `dashboard_id` is known and non-null at plan time:

- When `dashboard_id` IS known and non-null at plan time, the resource SHALL call `PUT /api/dashboards/dashboard/{dashboard_id}` (using the `UpdateDashboard` kbapi wrapper).
- When `dashboard_id` IS NOT known or IS null at plan time, the resource SHALL call `POST /api/dashboards/dashboard` (using the `CreateDashboard` kbapi wrapper).

The create function SHALL extract the resulting `dashboard_id` from the API response body: `JSON201.Id` when the API returns `201 Created`, or `JSON200.Id` when the API returns `200 OK` (possible for PUT when the server treats an existing id as an overwrite).

#### Scenario: POST path when dashboard_id is absent

- GIVEN plan has `dashboard_id` unknown or null
- WHEN `createDashboard` is called
- THEN it SHALL call `POST /api/dashboards/dashboard` (not PUT)
- AND state SHALL contain the Kibana-assigned UUID in `dashboard_id`

#### Scenario: PUT path when dashboard_id is known at plan time

- GIVEN plan has `dashboard_id = "my-id"` (known, non-null)
- WHEN `createDashboard` is called
- THEN it SHALL call `PUT /api/dashboards/dashboard/my-id`
- AND the `{id}` path parameter SHALL equal `"my-id"`

### Requirement: `ml_anomaly_swimlane` panel support (REQ-047)

The `elasticstack_kibana_dashboard` resource SHALL accept an optional `ml_anomaly_swimlane_config` block on panel entries whose `type` is `ml_anomaly_swimlane`. When the panel type is `ml_anomaly_swimlane`, the block is **required**; omitting it SHALL produce a plan-time error.

The `ml_anomaly_swimlane_config` block exposes a **flat** schema with the following attributes:

| Attribute | Type | Required/Optional | Notes |
|-----------|------|-------------------|-------|
| `swimlane_type` | string | Required | Enum: `"overall"` or `"viewBy"`. Discriminates the API union. |
| `job_ids` | list(string) | Required | At least one entry. Maps to `config.job_ids` in the API. |
| `view_by` | string | Conditional | Required when `swimlane_type = "viewBy"`; forbidden when `swimlane_type = "overall"`. Maps to `config.view_by` in the `viewBy` branch. |
| `per_page` | float32 | Optional | Number of rows per page in a view-by swim lane. Maps to `config.per_page`. |
| `title` | string | Optional | Panel title. |
| `description` | string | Optional | Panel description. |
| `hide_title` | bool | Optional | When true, hides the panel title. |
| `hide_border` | bool | Optional | When true, hides the panel border. |
| `time_range` | object | Optional | Panel-level time range with required `from` and `to` and optional `mode` (`"absolute"` \| `"relative"`). |

The `ml_anomaly_swimlane_config` block SHALL conflict with all other typed panel config blocks and with practitioner-authored `config_json`, consistent with REQ-006.

**Kibana version compatibility**: The underlying ML anomaly swim lane embeddable (introduced in Kibana 7.9.0) predates the Dashboard API panel schema used by this resource, but no minimum Kibana version specific to its typed representation in the Dashboard API's panel schema could be confirmed from release notes. Empirical verification during implementation: Kibana **9.4.0** rejects a panel with `type = "ml_anomaly_swimlane"` with a 400 (`"type" to be one of [...]` — the type is absent from the Dashboard API's panel-type enum), while Kibana **9.5.0** (snapshot, unreleased as of writing — matching the `kibana/main` branch the generated `kbapi` client is built from) accepts it and round-trips correctly. Since no other typed panel block in this resource carries a bespoke per-panel version gate, the provider does not add one for `ml_anomaly_swimlane_config` either; the Kibana Dashboard API SHALL reject the panel with a 400 on incompatible (pre-9.5) stack versions.

**Write path**: When `swimlane_type = "overall"`, the provider SHALL serialize the panel config using `KibanaHTTPAPIsMlAnomalySwimlane0` (omitting `view_by`). When `swimlane_type = "viewBy"`, the provider SHALL serialize using `KibanaHTTPAPIsMlAnomalySwimlane1` (including the required `view_by` field).

**Read path**: The provider SHALL detect the union branch from the API response. For optional fields (`per_page`, `title`, `description`, `hide_title`, `hide_border`, `time_range`), the provider SHALL apply null-preservation: if a field is null in Terraform state, it SHALL remain null after read even if the API returns a value for it.

#### Scenario: Create overall swim lane

- GIVEN a panel with `type = "ml_anomaly_swimlane"` and `ml_anomaly_swimlane_config.swimlane_type = "overall"` and `job_ids = ["my-job"]`
- WHEN the provider executes create
- THEN the request body SHALL include a panel whose config has `swimlane_type = "overall"` and `job_ids = ["my-job"]`, and SHALL NOT include a `view_by` field

#### Scenario: Create viewBy swim lane

- GIVEN a panel with `swimlane_type = "viewBy"` and `job_ids = ["my-job"]` and `view_by = "host.name"`
- WHEN the provider executes create
- THEN the request body SHALL include a panel whose config has `swimlane_type = "viewBy"`, `job_ids = ["my-job"]`, and `view_by = "host.name"`

#### Scenario: Reject view_by absent on viewBy swim lane

- GIVEN a panel with `swimlane_type = "viewBy"` and `view_by` absent
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic indicating `view_by` is required for `swimlane_type = "viewBy"`

#### Scenario: Reject view_by present on overall swim lane

- GIVEN a panel with `swimlane_type = "overall"` and `view_by` set to a non-null string
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic indicating `view_by` must not be set for `swimlane_type = "overall"`

#### Scenario: Reject missing config block

- GIVEN a panel with `type = "ml_anomaly_swimlane"` and no `ml_anomaly_swimlane_config` block
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic

#### Scenario: per_page null-preservation on read

- GIVEN a panel where `per_page` is null in Terraform state
- AND the API returns a value for `per_page` in its response
- WHEN the provider refreshes state
- THEN `per_page` SHALL remain null in state

#### Scenario: Round-trip with per_page

- GIVEN a panel with `per_page = 10`
- WHEN the provider applies and then refreshes state
- THEN `per_page` SHALL be `10` in state after refresh

#### Scenario: Reject config_json for ml_anomaly_swimlane panel

- GIVEN a panel with `type = "ml_anomaly_swimlane"` and `config_json` set in Terraform configuration
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic

---

### Requirement: `ml_single_metric_viewer` panel support (REQ-048)

The `elasticstack_kibana_dashboard` resource SHALL accept an optional `ml_single_metric_viewer_config` block on panel entries whose `type` is `ml_single_metric_viewer`. When the panel type is `ml_single_metric_viewer`, the block is **required**; omitting it SHALL produce a plan-time error.

The `ml_single_metric_viewer_config` block exposes the following attributes:

| Attribute | Type | Required/Optional | Notes |
|-----------|------|-------------------|-------|
| `job_ids` | list(string) | Required | Exactly one entry (length-1 validator). Maps to `config.job_ids`. |
| `selected_detector_index` | float32 | Optional | Zero-based detector index within the job. Maps to `config.selected_detector_index`. |
| `forecast_id` | string | Optional | Forecast identifier to overlay. Maps to `config.forecast_id`. |
| `function_description` | string | Optional | When set, MUST be one of `"min"`, `"max"`, or `"mean"` (plan-time validation). For `metric` detectors only; ignored for other detector functions. Maps to `config.function_description`. |
| `selected_entities` | map(object) | Optional | Map keyed by partition/by/over field name. Each value has optional `string_value (string)` and optional `numeric_value (number)`, with a plan-time validator requiring exactly one. |
| `title` | string | Optional | Panel title. |
| `description` | string | Optional | Panel description. |
| `hide_title` | bool | Optional | When true, hides the panel title. |
| `hide_border` | bool | Optional | When true, hides the panel border. |
| `time_range` | object | Optional | Panel-level time range with required `from` and `to` and optional `mode`. |

The `ml_single_metric_viewer_config` block SHALL conflict with all other typed panel config blocks and with practitioner-authored `config_json`.

**Kibana version compatibility**: The underlying ML single metric viewer embeddable (introduced in Kibana 8.13.0) predates the Dashboard API panel schema used by this resource, but no minimum Kibana version specific to its typed representation in the Dashboard API's panel schema could be confirmed from release notes. Empirical verification during implementation, mirroring `ml_anomaly_swimlane`: Kibana **9.4.0** rejects the Dashboard API panel type `ml_single_metric_viewer`, while Kibana **9.5.0** (snapshot, unreleased as of writing) accepts it and round-trips correctly. Since no other typed panel block in this resource carries a bespoke per-panel version gate, the provider does not add one for `ml_single_metric_viewer_config` either; the Kibana Dashboard API SHALL reject the panel with a 400 on incompatible (pre-9.5) stack versions.

**`selected_entities` serialization**: The attribute is a `MapNestedAttribute` keyed by field name. Each value object carries two optional attributes: `string_value` (Terraform `String`) and `numeric_value` (Terraform `Number`). A plan-time object validator SHALL enforce that exactly one of `string_value` or `numeric_value` is set on each value entry. On write, if `string_value` is set, the provider SHALL emit the entity value as the string union branch (`KibanaHTTPAPIsMlSingleMetricViewerSelectedEntities0`); if `numeric_value` is set, it SHALL emit as the numeric union branch (`KibanaHTTPAPIsMlSingleMetricViewerSelectedEntities1`, a `float32`). On read, the provider SHALL detect the union branch and populate the corresponding attribute; the other attribute SHALL remain null.

**`job_ids` length constraint**: A list-length validator (`listvalidator.SizeAtMost(1)`) SHALL enforce that practitioners cannot supply more than one job ID. This matches the Single Metric Viewer's single-job API semantics while preserving schema uniformity with the sibling ML panel family.

**Write path**: The provider SHALL construct `KibanaHTTPAPIsMlSingleMetricViewer` from state, including all optional fields when set.

**Read path**: Null-preservation applies to all optional attributes (`selected_detector_index`, `forecast_id`, `function_description`, `selected_entities`, and presentation attributes). If null in Terraform state, the attribute SHALL remain null after read even if the API returns a value.

#### Scenario: Create with string and numeric selected_entities

- GIVEN a panel with `selected_entities = { airline = { string_value = "AAL" }, region_code = { numeric_value = 4 } }`
- WHEN the provider executes create
- THEN the request body SHALL include `selected_entities.airline` as the string `"AAL"` and `selected_entities.region_code` as the numeric value `4`

#### Scenario: Reject both string_value and numeric_value on same entity

- GIVEN a `selected_entities` entry with both `string_value` and `numeric_value` set
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic

#### Scenario: Reject neither string_value nor numeric_value on same entity

- GIVEN a `selected_entities` entry with both `string_value` and `numeric_value` absent or null
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic

#### Scenario: Reject job_ids with more than one entry

- GIVEN a panel with `job_ids = ["job-a", "job-b"]`
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic indicating `job_ids` must contain exactly one entry

#### Scenario: selected_entities null-preservation on read

- GIVEN a panel where `selected_entities` is null in Terraform state
- AND the API returns non-empty `selected_entities` in its response
- WHEN the provider refreshes state
- THEN `selected_entities` SHALL remain null in state

#### Scenario: selected_entities round-trip

- GIVEN a panel with `selected_entities = { host = { string_value = "web-01" } }`
- WHEN the provider applies and then refreshes state
- THEN `selected_entities.host.string_value` SHALL be `"web-01"` in state after refresh
- AND `selected_entities.host.numeric_value` SHALL be null

#### Scenario: Reject config_json for ml_single_metric_viewer panel

- GIVEN a panel with `type = "ml_single_metric_viewer"` and `config_json` set in Terraform configuration
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic

#### Scenario: Reject missing config block

- GIVEN a panel with `type = "ml_single_metric_viewer"` and no `ml_single_metric_viewer_config` block
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time error diagnostic

### Requirement: APM service map panel support (REQ-049)

The `elasticstack_kibana_dashboard` resource SHALL support `type = "apm_service_map"` panels through a typed `apm_service_map_config` block. The block exposes the full flat configuration surface of `KibanaHTTPAPIsApmServiceMapEmbeddable`.

#### Schema attributes

All attributes within `apm_service_map_config` are optional unless stated otherwise.

**Service selectors** (all optional strings — freely combinable, no mutual exclusion):
- `environment` — APM service environment (e.g. `"production"`).
- `service_name` — Focus the map on a specific service.
- `service_group_id` — Reference to a saved APM service group (opaque string; no foreign-key validation by the provider).

**Query**:
- `kuery` — KQL query string (plain `StringAttribute`; always KQL, not an object).

**Layout**:
- `map_orientation` — String enum: `horizontal` or `vertical`. The resource SHALL return an error diagnostic at plan time when a value outside this set is supplied.
- `sync_with_dashboard_filters` — Boolean; when null, the attribute is omitted from the API payload.

**Filter lists** (each a set of validated strings; order does not affect plan stability):
- `alert_status_filter` — Set of strings; allowed values: `active`, `delayed`, `recovered`, `untracked`.
- `anomaly_severity_filter` — Set of strings; allowed values: `low`, `warning`, `minor`, `major`, `critical`, `unknown`.
- `connection_filter` — Set of strings; allowed values: `connected`, `orphaned`.
- `slo_status_filter` — Set of strings; allowed values: `degrading`, `healthy`, `noData`, `violated`.

Invalid values for any filter set attribute SHALL produce an error diagnostic at plan time.

**Presentation passthroughs** (reuse `panelkit.PanelPresentationAttributes()`):
- `title`, `description`, `hide_title`, `hide_border`

**Time range**:
- `time_range` — Optional sub-block `{ from: string, to: string, mode: optional string ("absolute" | "relative") }`.

#### Write (ToAPI) behaviour

When `apm_service_map_config` is set, the provider SHALL:
- Set `type` to `"apm_service_map"` in the API panel payload.
- Map each set attribute that is non-null and non-empty to the corresponding slice in the `config` object; omit null/empty sets from the payload.
- Map scalar optional attributes (strings, bools) only when non-null.
- Map `time_range` only when the block is non-null.

`config_json` SHALL NOT be accepted for `apm_service_map` panels; the registry guard (REQ-044A) SHALL return an error diagnostic if `config_json` is set on a panel with `type = "apm_service_map"`. The `apm_service_map` panel type SHALL be managed exclusively through the typed `apm_service_map_config` block.

#### Read (FromAPI) behaviour and null-preservation

On read, the provider SHALL apply REQ-009 null-preservation for every optional field:
- When prior state had a field null, the provider SHALL keep it null in state even if the API returns a value.
- When prior state had a field set, the provider SHALL update it from the API response.
- For filter set attributes: when prior state had the attribute null, the provider SHALL keep it null regardless of the API response. When prior state had the attribute set (including empty set), the provider SHALL reconstruct the `types.Set` from the API slice; the set implementation guarantees that element order is ignored for plan comparison, so re-ordered API responses SHALL produce no plan diff.
- `time_range` — when prior state had it null and the API echoes a value (e.g. the dashboard-level time range), state SHALL remain null.
- On import (no prior state): when the API returns a non-empty config object, populate all non-null API fields into state; when the API returns a nil or empty config, leave `apm_service_map_config` null in state.

The `apm_service_map_config` block SHALL be mutually exclusive with all other typed panel config blocks and with `config_json`. The registry-driven mutual-exclusion guard (REQ-044A) enforces this.

#### Scenarios

##### Scenario: Create apm_service_map panel with environment selector

- GIVEN a dashboard configuration with a panel of `type = "apm_service_map"` and `apm_service_map_config = { environment = "production" }`
- WHEN the resource creates the dashboard
- THEN the API payload SHALL include `"config": { "environment": "production" }` in the panel body
- AND a subsequent plan SHALL show no changes

##### Scenario: Create apm_service_map panel with service_name selector

- GIVEN a dashboard configuration with `apm_service_map_config = { service_name = "checkout" }`
- WHEN the resource creates the dashboard
- THEN the API payload SHALL include `"config": { "service_name": "checkout" }` in the panel body
- AND a subsequent plan SHALL show no changes

##### Scenario: Create apm_service_map panel with service_group_id selector

- GIVEN a dashboard configuration with `apm_service_map_config = { service_group_id = "group-abc" }`
- WHEN the resource creates the dashboard
- THEN the API payload SHALL include `"config": { "service_group_id": "group-abc" }` in the panel body

##### Scenario: Create apm_service_map panel with all three service selectors combined

- GIVEN a dashboard configuration with `apm_service_map_config = { environment = "staging", service_name = "checkout", service_group_id = "group-abc" }`
- WHEN the resource creates the dashboard
- THEN all three fields SHALL appear in the API payload
- AND no mutual-exclusion error SHALL be returned

##### Scenario: Filter sets with multiple values and order independence

- GIVEN a dashboard with `apm_service_map_config.alert_status_filter = ["active", "delayed"]`
- WHEN the resource reads back the dashboard and the API returns `["delayed", "active"]` (reversed order)
- THEN the provider SHALL produce no plan diff
- AND state SHALL contain a set with values `"active"` and `"delayed"`

##### Scenario: All filter sets populated

- GIVEN a panel with all four filter attributes set with multiple valid enum values
- WHEN the resource creates the dashboard and reads it back
- THEN each filter set in state SHALL contain the expected values
- AND a subsequent plan SHALL show no changes

##### Scenario: Invalid alert_status_filter value rejected

- GIVEN a panel with `apm_service_map_config = { alert_status_filter = ["invalid_value"] }`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the value is not an allowed enum member

##### Scenario: Invalid map_orientation value rejected

- GIVEN a panel with `apm_service_map_config = { map_orientation = "diagonal" }`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the value must be `horizontal` or `vertical`

##### Scenario: config_json rejected for apm_service_map panel

- GIVEN a panel with `type = "apm_service_map"` and `config_json` also set
- WHEN Terraform plans the configuration
- THEN the resource SHALL return an error diagnostic indicating `config_json` is unsupported for the `apm_service_map` panel type

##### Scenario: Null-preservation on optional scalars

- GIVEN a prior state where `apm_service_map_config.environment` is null
- WHEN the API read returns an `environment` value
- THEN state SHALL keep `environment` null
- AND the subsequent plan SHALL show no changes

##### Scenario: Import null-preservation

- GIVEN an existing dashboard with `apm_service_map` panels that have API-side defaults for optional fields
- WHEN the resource imports the dashboard
- THEN optional fields not explicitly configured SHALL remain null in state
- AND a subsequent plan against a configuration that omits those fields SHALL show no changes

##### Scenario: Full configuration round-trip

- GIVEN an `apm_service_map_config` block with every attribute populated
- WHEN the resource creates the dashboard and reads it back
- THEN all attribute values SHALL appear in state
- AND a subsequent plan SHALL show no changes

### Requirement: AIOps log rate analysis panel behavior (REQ-050)

The `elasticstack_kibana_dashboard` resource SHALL support a panel of `type = "aiops_log_rate_analysis"` via an `aiops_log_rate_analysis_config` block. The config block SHALL accept:

- `data_view_id` (required string): the data view ID used to run log rate analysis.
- Standard panelkit presentation passthroughs (all optional): `title`, `description`,
  `hide_title`, `hide_border`, `time_range`.

On create and update the resource SHALL serialize these fields into the
`KibanaHTTPAPIsKbnDashboardPanelTypeAiopsLogRateAnalysis` API panel type. On read the resource
SHALL apply REQ-009 null-preservation: optional fields that were null in prior state SHALL remain
null after read even if Kibana returns server-side defaults. On import (no prior state) the
resource SHALL populate `data_view_id` from the API; optional presentation fields SHALL be
populated only when the API returns non-nil values.

The resource SHALL reject simultaneous `aiops_log_rate_analysis_config` and `config_json` on
the same panel with a plan-time error. The resource SHALL require `data_view_id` to be set when
`type = "aiops_log_rate_analysis"` and `aiops_log_rate_analysis_config` is provided.

No drilldowns are supported on this panel type (the API model does not expose them).

#### Scenario: Required-only log rate analysis panel round-trip

- GIVEN a panel with `type = "aiops_log_rate_analysis"` and `aiops_log_rate_analysis_config = { data_view_id = "logs-*" }`
- WHEN the resource creates the dashboard and reads state
- THEN `panels.0.type` SHALL equal `"aiops_log_rate_analysis"`, `panels.0.aiops_log_rate_analysis_config.data_view_id` SHALL equal `"logs-*"`, and optional presentation fields SHALL be null
- AND a subsequent plan SHALL show no changes

#### Scenario: Import preserves data_view_id

- GIVEN an existing Kibana dashboard containing an `aiops_log_rate_analysis` panel with `data_view_id = "logs-*"`
- WHEN the resource imports the dashboard
- THEN `data_view_id` SHALL equal `"logs-*"` in state
- AND a plan with a config specifying only `data_view_id` SHALL show no changes

#### Scenario: Optional presentation fields round-trip

- GIVEN `aiops_log_rate_analysis_config` with `title = "Log spikes"`, `hide_title = true`, and `hide_border = false`
- WHEN the resource creates the dashboard and reads state
- THEN each presentation field SHALL appear in state with its specified value
- AND a plan with no config changes SHALL show no changes

#### Scenario: `config_json` conflict rejected

- GIVEN a panel with `type = "aiops_log_rate_analysis"` that sets both `aiops_log_rate_analysis_config` and `config_json`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic

---

### Requirement: AIOps pattern analysis panel behavior (REQ-051)

The `elasticstack_kibana_dashboard` resource SHALL support a panel of `type = "aiops_pattern_analysis"` via an `aiops_pattern_analysis_config` block. The config block SHALL accept:

- `data_view_id` (required string): the data view ID used for pattern analysis.
- `field_name` (required string): the text field on which to run pattern analysis.
- `minimum_time_range` (optional string, enum): one of `no_minimum`, `1_week`, `1_month`, `3_months`, `6_months`. Invalid values SHALL be rejected at plan time.
- `random_sampler_mode` (optional string, enum): one of `off`, `on_automatic`, `on_manual`. Invalid values SHALL be rejected at plan time.
- `random_sampler_probability` (optional float32): the sampling probability, bounded to `[0.00001, 0.5]`. Values outside this range SHALL be rejected at plan time. This field is only meaningful when `random_sampler_mode = "on_manual"`.
- Standard panelkit presentation passthroughs (all optional): `title`, `description`, `hide_title`, `hide_border`, `time_range`.

On create and update the resource SHALL serialize non-null optional enum and float fields into the `KibanaHTTPAPIsKbnDashboardPanelTypeAiopsPatternAnalysis` API panel type. On read the resource SHALL apply REQ-009 null-preservation. On import the resource SHALL populate `data_view_id` and `field_name` from the API.

The resource SHALL reject simultaneous `aiops_pattern_analysis_config` and `config_json` on the same panel. No drilldowns are supported.

#### Scenario: Required-only pattern analysis panel round-trip

- GIVEN `aiops_pattern_analysis_config = { data_view_id = "logs-*", field_name = "message" }`
- WHEN the resource creates and reads state
- THEN `data_view_id` SHALL equal `"logs-*"`, `field_name` SHALL equal `"message"`, and all optional fields SHALL be null
- AND a subsequent plan SHALL show no changes

#### Scenario: All optional fields round-trip

- GIVEN `aiops_pattern_analysis_config` with `minimum_time_range = "1_week"`, `random_sampler_mode = "on_manual"`, `random_sampler_probability = 0.01`
- WHEN the resource creates and reads state
- THEN each field SHALL appear in state with its specified value

#### Scenario: Probability out of range rejected

- GIVEN `aiops_pattern_analysis_config` with `random_sampler_probability = 1.0`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the value must be between `0.00001` and `0.5`

#### Scenario: Invalid enum rejected at plan time

- GIVEN `aiops_pattern_analysis_config` with `minimum_time_range = "2_weeks"` or `random_sampler_mode = "maybe"`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the valid enum values

#### Scenario: Null-preservation of optional fields on update

- GIVEN a dashboard created with `minimum_time_range` and `random_sampler_mode` omitted
- WHEN Kibana returns those fields with server-side defaults on read
- THEN state SHALL keep them null and a plan SHALL show no changes

---

### Requirement: AIOps change point chart panel behavior (REQ-052)

The `elasticstack_kibana_dashboard` resource SHALL support a panel of `type = "aiops_change_point_chart"` via an `aiops_change_point_chart_config` block. The config block SHALL accept:

- `data_view_id` (required string): the data view ID used for change point detection.
- `metric_field` (required string): the metric field used by the aggregation function.
- `aggregation_function` (optional string, enum): one of `avg`, `max`, `min`, `sum`. Invalid values SHALL be rejected at plan time.
- `split_field` (optional string): the optional field used to split change-point results.
- `partitions` (optional set of strings): optional split field values to include in the panel. Modelled as a set to prevent plan drift from API-returned ordering. Semantically a filter set; duplicate entries are silently deduplicated. An empty set is not meaningful (omit the attribute to disable filtering); a non-null set SHALL contain at least one entry and SHALL be rejected at plan time otherwise.
- `max_series_to_plot` (optional float32): maximum number of change points to visualise. Kibana default is 6. The resource SHALL null-preserve this field when the user omitted it.
- `view_type` (optional string, enum): one of `charts`, `table`. Invalid values SHALL be rejected at plan time.
- Standard panelkit presentation passthroughs (all optional): `title`, `description`, `hide_title`, `hide_border`, `time_range`.

On create and update the resource SHALL serialize non-null optional fields into the `KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart` API panel type. The `partitions` set SHALL be serialized to `*[]string` in the API body. On read the resource SHALL apply REQ-009 null-preservation. On import the resource SHALL populate `data_view_id` and `metric_field` from the API.

The resource SHALL reject simultaneous `aiops_change_point_chart_config` and `config_json` on the same panel. No drilldowns are supported.

#### Scenario: Required-only change point chart round-trip

- GIVEN `aiops_change_point_chart_config = { data_view_id = "metrics-*", metric_field = "system.cpu.total.pct" }`
- WHEN the resource creates and reads state
- THEN `data_view_id` and `metric_field` SHALL be present in state; all optional fields (including `partitions`) SHALL be null
- AND a subsequent plan SHALL show no changes

#### Scenario: Partitions set is order-insensitive

- GIVEN `aiops_change_point_chart_config` with `partitions = ["host-b", "host-a", "host-c"]`
- WHEN the resource creates the dashboard and Kibana returns the partitions in a different order
- THEN state SHALL reflect the set (regardless of order) and a plan SHALL show no changes

#### Scenario: Empty partitions set rejected at plan time

- GIVEN `aiops_change_point_chart_config` with `partitions = []`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating `partitions` must contain at least one entry; the user SHALL omit the attribute instead

#### Scenario: All optional fields round-trip

- GIVEN `aiops_change_point_chart_config` with `aggregation_function = "avg"`, `split_field = "host.name"`, `partitions = ["host-a"]`, `max_series_to_plot = 6`, `view_type = "charts"`
- WHEN the resource creates and reads state
- THEN each field SHALL appear in state with its specified value

#### Scenario: Invalid enum rejected at plan time

- GIVEN `aiops_change_point_chart_config` with `aggregation_function = "median"` or `view_type = "grid"`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the valid enum values

#### Scenario: Multi-panel AIOps dashboard — sibling mutual exclusion

- GIVEN a dashboard containing three panels each of a different AIOps type (log rate analysis, pattern analysis, change point chart)
- WHEN the resource creates and reads state
- THEN each panel's config block SHALL be non-null only for its own type, and all sibling config blocks SHALL be null
- AND a plan SHALL show no changes

### Requirement: ML anomaly charts panel behavior (REQ-053)

The resource SHALL support `type = "ml_anomaly_charts"` panels through the typed `ml_anomaly_charts_config` block. When a panel entry sets `type = "ml_anomaly_charts"`, the resource SHALL require the `ml_anomaly_charts_config` block and SHALL return an error diagnostic when it is absent.

The block accepts the following attributes:

- `job_ids` (required `list(string)`, min 1 item): one or more anomaly-detection job IDs or group IDs whose results are shown. The provider treats these as opaque strings and does not validate their existence against Kibana's ML API at plan time; invalid job IDs surface as Kibana API errors during `terraform apply`.
- `max_series_to_plot` (optional int64): maximum number of anomaly series to plot. When null in state, the attribute is omitted from the API request. The Kibana API represents this field as a JSON number (`*float32` in the generated client); the provider exposes it as an integer since a series count cannot be fractional, converting to/from the API's numeric type at the boundary.
- `severity_threshold` (optional list of objects, min 1 item when present): filters which severity bands are displayed. Each list item is a union — exactly one of the following must be set per item:
  - `severity` (string, one of `low`, `warning`, `minor`, `major`, `critical`): a named severity shortcut. The model layer SHALL expand named severities to their canonical `{min, max}` API pairs at write time.
  - `min` (int64) plus optional `max` (int64): a raw numeric range. `max` may be set only when `min` is set and `severity` is unset; when `max` is set, `min` must also be set. Setting both `severity` and `min` on the same item SHALL produce an error diagnostic at plan time. Setting `severity` together with `max` SHALL produce an error diagnostic at plan time.
- `title` (optional string): panel title. Subject to REQ-009 null-preservation.
- `description` (optional string): panel description. Subject to REQ-009 null-preservation.
- `hide_title` (optional bool): when true, hides the panel title. Subject to REQ-009 null-preservation.
- `hide_border` (optional bool): when true, hides the panel border. Subject to REQ-009 null-preservation.
- `time_range` (optional object: `from` string required, `to` string required, `mode` string optional): a panel-level time range override, identical in shape to the dashboard root `time_range`. Reuses `panelkit.TimeRangeSchema`. Subject to REQ-009 null-preservation: when prior state has `time_range` null, the provider SHALL keep it null even if the API returns a default; when `mode` is null in prior state, the provider SHALL keep `mode` null.

The model layer SHALL expand named severity values to the following canonical `{min, max}` API pairs (matching the generated Kibana OpenAPI const values in `KibanaHTTPAPIsMlAnomalyChartsSeverityThreshold0`–`SeverityThreshold4`):

| `severity` | API `min` | API `max`    |
|---|---|---|
| `low`       | 0         | 3            |
| `warning`   | 3         | 25           |
| `minor`     | 25        | 50           |
| `major`     | 50        | 75           |
| `critical`  | 75        | (omitted — open-ended upper bound) |

On write, the provider SHALL map `ml_anomaly_charts_config` to the `config` object in the `KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts` API schema. Optional fields SHALL be included only when set in state; absent optional fields SHALL NOT be sent to the API.

On read, the provider SHALL repopulate `ml_anomaly_charts_config` from the API response using REQ-009 null-preservation, extended to the **representation form** of `severity_threshold` items. The API encodes `severity_threshold` as `{min, max}` pairs only; it conveys no information about whether the practitioner authored a named `severity` or a raw numeric range. Therefore the chosen form is recovered from prior state, not inferred by normalizing:

- When the prior item holds a named `severity` (and `min`/`max` are null), the provider SHALL store the named form, deriving the label from the API `{min, max}` pair via the canonical-band table. The `critical` band (API: `{min: 75}`, no `max` field) SHALL map to `severity = "critical"`. If the API value no longer matches any canonical band, the provider SHALL fall back to the raw `min`/`max` form (surfacing as drift).
- When the prior item holds raw `min`/`max` (and `severity` is null), the provider SHALL store the raw `min`/`max` verbatim from the API, even when the pair coincidentally equals a canonical band.
- On import (no prior state), the provider SHALL default to the named form when the API `{min, max}` matches a canonical band, and to the raw form otherwise.

The provider SHALL NOT normalize a practitioner-authored raw range into a named `severity` on read. While the configured values match current state, a subsequent plan SHALL show no changes.

`config_json` SHALL NOT be supported for `ml_anomaly_charts` panels; using `config_json` with `type = "ml_anomaly_charts"` SHALL return an error diagnostic (per REQ-010 policy).

Implementation: new package `internal/kibana/dashboard/panel/mlanomalycharts/` with `schema.go`, `model.go`, and `api.go`; new model file `internal/kibana/dashboard/models/mlanomalycharts.go`; registration in `internal/kibana/dashboard/schema.go` and `internal/kibana/dashboard/registry.go`.

#### Scenario: Creation of ml_anomaly_charts panel with named severities

- GIVEN a panel with `type = "ml_anomaly_charts"` and `ml_anomaly_charts_config` containing `job_ids = ["my-job"]` and `severity_threshold = [{ severity = "critical" }, { severity = "major" }]`
- WHEN create runs
- THEN the provider SHALL send `job_ids = ["my-job"]` and `severity_threshold = [{min: 75}, {min: 50, max: 75}]` in the API request
- AND after the post-apply read, state SHALL represent both items as named severities
- AND a subsequent plan SHALL show no changes

#### Scenario: Round-trip stability for critical (open-ended) severity

- GIVEN a panel with `severity_threshold = [{ severity = "critical" }]` applied and read back
- WHEN the API returns `severity_threshold = [{min: 75}]` (no `max` field)
- THEN the provider SHALL map this to `severity = "critical"` in state
- AND a subsequent plan SHALL show no changes

#### Scenario: Raw range escape hatch

- GIVEN a panel with `severity_threshold = [{ min = 10, max = 20 }]`
- WHEN create runs and the post-apply read returns `{min: 10, max: 20}`
- THEN state SHALL contain `min = 10` and `max = 20` (not coerced to a named severity)
- AND a subsequent plan SHALL show no changes

#### Scenario: Raw range coinciding with a canonical band is preserved (no diff)

- GIVEN a panel where the practitioner set `severity_threshold = [{ min = 3, max = 25 }]` (coincides with the `warning` canonical band)
- WHEN create runs and the post-apply read returns `{min: 3, max: 25}`
- THEN the provider SHALL store `min = 3` and `max = 25` in state (NOT coerced to `severity = "warning"`)
- AND a subsequent plan SHALL show no changes

#### Scenario: severity_threshold form is preserved across refresh

- GIVEN state holds `severity_threshold = [{ severity = "major" }, { min = 10, max = 20 }]`
- WHEN a refresh runs and the API returns `[{min: 50, max: 75}, {min: 10, max: 20}]`
- THEN state SHALL retain the first item as `severity = "major"` and the second as `min = 10, max = 20`
- AND a subsequent plan SHALL show no changes

#### Scenario: critical severity preserved in raw form when authored raw

- GIVEN a panel where the practitioner set `severity_threshold = [{ min = 75 }]` (raw form, coincides with the `critical` canonical band)
- WHEN create runs and the post-apply read returns `{min: 75}` (no `max` field)
- THEN the provider SHALL store `min = 75` with `max` null in state (NOT coerced to `severity = "critical"`)
- AND a subsequent plan SHALL show no changes

#### Scenario: Switching severity form is a configuration change

- GIVEN state holds `severity_threshold = [{ severity = "warning" }]`
- WHEN the configuration changes to `{ min = 3, max = 25 }` (same band, raw form)
- THEN the plan SHALL report a change for that item
- AND after apply the state SHALL settle to `{ min = 3, max = 25 }` with a subsequent plan showing no changes

#### Scenario: Import defaults to named form for canonical bands

- GIVEN an existing panel whose API `severity_threshold` is `[{min: 3, max: 25}]` and no prior Terraform state
- WHEN the panel is imported
- THEN state SHALL store `severity = "warning"` (named form preferred only on import, where no practitioner form exists to preserve)

#### Scenario: Plan-time validation — both severity and min set

- GIVEN a `severity_threshold` item with both `severity = "major"` and `min = 50`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating that exactly one of `severity` or `min` must be set

#### Scenario: Plan-time validation — max without min

- GIVEN a `severity_threshold` item with `max = 75` but neither `severity` nor `min` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic

#### Scenario: config_json rejected for ml_anomaly_charts

- GIVEN a panel with `type = "ml_anomaly_charts"` and `config_json = "{}"`
- WHEN Terraform validates or applies the configuration
- THEN the resource SHALL return an error diagnostic indicating that `config_json` is not supported for `ml_anomaly_charts` panels

#### Scenario: Optional fields follow null-preservation

- GIVEN an `ml_anomaly_charts_config` that does not set `max_series_to_plot` or `time_range`
- WHEN apply runs and the post-apply read returns server-side defaults for those fields
- THEN state SHALL keep `max_series_to_plot` and `time_range` as null
- AND a subsequent plan SHALL show no changes

#### Scenario: Update job_ids in-place

- GIVEN an existing `ml_anomaly_charts` panel with `job_ids = ["job-a"]`
- WHEN the configuration changes to `job_ids = ["job-a", "job-b"]` and update runs
- THEN the resource SHALL NOT destroy and recreate the dashboard
- AND a subsequent plan SHALL show no changes

### Requirement: Field statistics table panel behavior (REQ-054)

When a panel entry sets `type = "field_stats_table"`, the resource SHALL accept a `field_stats_table_config` block and SHALL require that block to be present when the panel type is `field_stats_table`. The block SHALL expose two mutually exclusive sub-blocks mirroring the API's `view_type` discriminated union:

- `by_dataview` — backed by a Kibana data view (API `view_type = "dataview"`).
- `by_esql` — backed by an ES|QL query (API `view_type = "esql"`).

Exactly one of `by_dataview` or `by_esql` SHALL be set; setting both or neither SHALL produce an error diagnostic at plan time.

The `field_stats_table_config` block SHALL conflict with all other typed panel config blocks and with panel-level `config_json`. When `type = "field_stats_table"`, `config_json` SHALL NOT be set on the same panel entry; if it is, the resource SHALL return an error diagnostic indicating `config_json` is not supported for `field_stats_table`.

#### The `by_dataview` sub-block

The `by_dataview` sub-block SHALL accept:

- `data_view_id` (required, string) — the identifier of the source data view.
- `show_distributions` (optional, bool) — whether to show distribution mini-charts in the table; null-preserved on read per REQ-009: when prior state has it null, the provider keeps it null even if Kibana returns a server-side default.
- `title` (optional, string) — panel display title; null-preserved on read.
- `description` (optional, string) — panel description; null-preserved on read.
- `hide_title` (optional, bool) — whether to hide the panel title; null-preserved on read.
- `hide_border` (optional, bool) — whether to hide the panel border; null-preserved on read.
- `time_range` (optional, object `{ from = required string, to = required string, mode = optional string }`) — panel-level time range override; null-preserved on read: when prior state has `time_range` null, the provider keeps it null even if Kibana returns values.

On write, the provider SHALL set `view_type = "dataview"` internally and map `data_view_id` and optional fields to the API payload. The `view_type` field is not exposed as a user-facing attribute.

#### The `by_esql` sub-block

The `by_esql` sub-block SHALL accept:

- `query` (required, string) — the ES|QL query string; mapped to `query.esql` in the API payload.
- `show_distributions` (optional, bool) — null-preserved on read per REQ-009.
- `title` (optional, string) — null-preserved on read.
- `description` (optional, string) — null-preserved on read.
- `hide_title` (optional, bool) — null-preserved on read.
- `hide_border` (optional, bool) — null-preserved on read.
- `time_range` (optional, object `{ from = required string, to = required string, mode = optional string }`) — null-preserved on read.

On write, the provider SHALL set `view_type = "esql"` internally and map `query` to `query.esql` and optional fields to the API payload.

#### Read behavior

On read, the resource SHALL detect the `view_type` field in the API response and populate the matching sub-block (`by_dataview` or `by_esql`), leaving the other sub-block null. For each optional attribute, the resource SHALL apply REQ-009 null-preservation: if prior state had the attribute null, the provider SHALL keep it null even if the API response contains a value for it.

#### Scenario: by_dataview branch create/read round-trip

- GIVEN a panel with `type = "field_stats_table"` and `field_stats_table_config.by_dataview = { data_view_id = "logs-view", show_distributions = true, time_range = { from = "now-24h", to = "now" } }`
- WHEN create runs and the post-apply read returns the panel
- THEN state SHALL contain `by_dataview` populated with the same values, `by_esql` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: by_esql branch create/read round-trip

- GIVEN a panel with `type = "field_stats_table"` and `field_stats_table_config.by_esql = { query = "FROM logs | STATS count = COUNT(*) BY service.name", show_distributions = false }`
- WHEN create runs and the post-apply read returns the panel
- THEN state SHALL contain `by_esql` populated with the same values, `by_dataview` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: Exactly one of by_dataview or by_esql

- GIVEN a panel with both `field_stats_table_config.by_dataview` and `field_stats_table_config.by_esql` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one of `by_dataview` or `by_esql` must be set

#### Scenario: Neither branch set

- GIVEN a panel with `field_stats_table_config = {}` (neither `by_dataview` nor `by_esql` set)
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one of `by_dataview` or `by_esql` must be set

#### Scenario: time_range null-preservation

- GIVEN a `field_stats_table_config.by_dataview` panel whose prior state has `time_range = null`
- WHEN the post-apply read returns a panel where Kibana populated `time_range` with values
- THEN state SHALL keep `time_range` null and a subsequent plan against configuration that omits `time_range` SHALL show no changes

#### Scenario: show_distributions null-preservation

- GIVEN a `field_stats_table_config.by_esql` panel whose prior state has `show_distributions = null`
- WHEN the post-apply read returns a panel where Kibana populated `show_distributions`
- THEN state SHALL keep `show_distributions` null and a subsequent plan against configuration that omits it SHALL show no changes

#### Scenario: config_json rejected for field_stats_table panel type

- GIVEN a panel with `type = "field_stats_table"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `field_stats_table`

#### Scenario: Drift detection — Kibana returns branch data intact

- GIVEN an existing dashboard with a `field_stats_table` panel in state
- WHEN Kibana returns the same panel configuration on a subsequent read
- THEN a plan SHALL show no changes

---

### Requirement: Dashboard resource schema version upgrade (REQ-040)

The `elasticstack_kibana_dashboard` resource SHALL implement `terraform-plugin-framework`'s `ResourceWithUpgradeState` interface with a single state upgrader for version 0 → 1.

The v0 → v1 upgrader SHALL:

1. Inspect every entry in `panels[]` and every entry in `pinned_panels[]`.
2. For each entry whose `type` is `"options_list_control"`: move all flat attributes from `options_list_control_config` (i.e. `data_view_id`, `field_name`, `title`, `use_global_filters`, `ignore_validations`, `single_select`, `exclude`, `exists_selected`, `run_past_timeout`, `search_technique`, `selected_options`, `display_settings`, `sort`) into a nested `by_field {}` object within `options_list_control_config`.
3. For each entry whose `type` is `"range_slider_control"`: move all flat attributes from `range_slider_control_config` (`data_view_id`, `field_name`, `title`, `use_global_filters`, `ignore_validations`, `value`, `step`) into a nested `by_field {}` object within `range_slider_control_config`.
4. Leave all other panel types unchanged.

The resource schema version SHALL be incremented to 1. No data SHALL be lost during the upgrade; the resulting state SHALL be functionally equivalent to the original state.

#### Scenario: State upgrade preserves all field-branch attributes

- GIVEN a v0 state containing both `options_list_control` and `range_slider_control` panels with all optional attributes set (e.g. `sort`, `display_settings`, `value`, `step`)
- WHEN the state upgrader runs
- THEN all attribute values SHALL be present under the `by_field {}` sub-object in v1 state and no attributes SHALL be dropped

#### Scenario: Non-control panels are unaffected by the upgrader

- GIVEN a v0 state containing a mix of `options_list_control`, `range_slider_control`, and `markdown` panels
- WHEN the state upgrader runs
- THEN the `markdown` panel entries SHALL be unchanged in v1 state

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
| `viz` panel / `vis_config` / REQ-042 | `internal/kibana/dashboard/models_panels.go`, `internal/kibana/dashboard/models_lens_panel.go`, `internal/kibana/dashboard/schema.go` |
| Drift normalization | `internal/kibana/dashboard/panel_config_defaults.go`, `internal/kibana/dashboard/models_plan_state_alignment.go`, `internal/kibana/dashboard/models_xy_chart_panel.go` |
| Waffle validation | `internal/kibana/dashboard/waffle_config_validator.go` |
| Dashboard API status handling | `internal/clients/kibanaoapi/dashboards.go` |
| Composite id parsing | `internal/clients/api_client.go` |
