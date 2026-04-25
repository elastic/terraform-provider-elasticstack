# Delta Spec: `lens-dashboard-app` Panel Support

Base spec: `openspec/specs/kibana-dashboard/spec.md` (capability `kibana-dashboard`).

This delta modifies requirements REQ-006 and REQ-025 and adds REQ-035 for `lens-dashboard-app` panel support. Keep this file aligned with `proposal.md` and `design.md` until the change is archived.

---

## Schema additions

The following block is added to the panel object within the `panels` list (and within `sections[*].panels`):

```hcl
lens_dashboard_app_config = <optional, object({
  # Exactly one of by_value or by_reference must be set

  by_value = <optional, object({
    config_json = <required, json string, normalized>  # full API by-value Lens chart config object
  })>

  by_reference = <optional, object({
    ref_id          = <required, string>  # API reference name for the linked library item
    references_json = <optional, json string, normalized>  # array of { id: string, name: string, type: string }
    title           = <optional, string>
    description     = <optional, string>
    hide_title      = <optional, bool>
    hide_border     = <optional, bool>
    drilldowns_json = <optional, json string, normalized>
    time_range = <required, object({
      from = <required, string>
      to   = <required, string>
      mode = <optional, string> # absolute | relative
    })>
  })>
})> # only with type = "lens-dashboard-app"; conflicts with all other config blocks; exactly one of by_value or by_reference must be set
```

**Distinction from existing `vis` Lens panel type**: The `lens_dashboard_app_config` block applies exclusively to panels with `type = "lens-dashboard-app"`. Panels with `type = "vis"` continue to use the existing typed Lens config blocks (`xy_chart_config`, `metric_chart_config`, `waffle_config`, etc.) and supported `config_json` behavior. The type string `lens-dashboard-app` must appear verbatim in the panel `type` attribute; it is not interchangeable with `vis`.

---

## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

Schema validation SHALL enforce that `lens_dashboard_app_config` is valid only for panels with `type = "lens-dashboard-app"`, is mutually exclusive with all other panel configuration blocks, and that exactly one of `by_value` or `by_reference` is set.

The existing REQ-006 text is extended. The sentence:

> Each panel SHALL declare at least one panel configuration block, panel configuration blocks SHALL be mutually exclusive, typed panel configuration blocks SHALL only be valid for their supported panel type, and `waffle_config` SHALL enforce its ES|QL-vs-non-ES|QL field consistency rules.

gains the following additions:

- `lens_dashboard_app_config` SHALL be valid only for panels with `type = "lens-dashboard-app"`.
- `lens_dashboard_app_config` SHALL be mutually exclusive with all other panel configuration blocks.
- Within `lens_dashboard_app_config`, exactly one of `by_value` or `by_reference` SHALL be set; setting both or neither SHALL be rejected at plan time.
- `by_value.config_json` SHALL be required when `by_value` is set.
- `by_reference.ref_id` and `by_reference.time_range` SHALL be required when `by_reference` is set.
- `by_reference.time_range.mode` SHALL be restricted to `absolute` or `relative` when set.

#### Scenario: lens_dashboard_app_config rejected for non-lens-dashboard-app panel (ADDED)

- GIVEN a panel with `type = "vis"` and `lens_dashboard_app_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: Both sub-blocks set simultaneously (ADDED)

- GIVEN a `lens_dashboard_app_config` block with both `by_value` and `by_reference` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that `by_value` and `by_reference` are mutually exclusive

#### Scenario: Neither sub-block set (ADDED)

- GIVEN a `lens_dashboard_app_config` block with neither `by_value` nor `by_reference` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating that exactly one of `by_value` or `by_reference` must be set

---

### Requirement: Raw `config_json` panel behavior (REQ-025)

`config_json` SHALL NOT be supported for `lens-dashboard-app` panels; the `lens-dashboard-app` panel type SHALL be managed exclusively through the typed `lens_dashboard_app_config` block.

The existing REQ-025 text:

> On write, `config_json` SHALL be supported only for `markdown` and `vis` panel types; using `config_json` with any other panel type, or omitting all panel configuration blocks, SHALL return an error diagnostic.

is updated to:

> On write, panel-level `config_json` SHALL be supported only for `markdown` and `vis` panel types; using panel-level `config_json` with any other panel type, including `lens-dashboard-app`, or omitting all panel configuration blocks, SHALL return an error diagnostic. The `lens-dashboard-app` panel type SHALL be managed exclusively through the typed `lens_dashboard_app_config` block.

#### Scenario: config_json rejected for lens-dashboard-app panel type (ADDED)

- GIVEN a panel with `type = "lens-dashboard-app"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `lens-dashboard-app`

---

## ADDED Requirements

### Requirement: `lens-dashboard-app` panel behavior (REQ-035)

For `type = "lens-dashboard-app"` panels, the resource SHALL accept `lens_dashboard_app_config` with exactly one of the `by_value` or `by_reference` sub-blocks set. Within `by_value`, the `config_json` attribute is required and SHALL contain a JSON object that maps directly to the generated by-value `KbnDashboardPanelTypeLensDashboardApp.config` union. Within `by_reference`, the `ref_id` and `time_range` attributes are required. The optional by-reference attributes `references_json`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns_json` MAY be set.

**On write (create and update):**

For by-value panels, the resource SHALL map `by_value.config_json` directly to the panel `config` object without wrapping it in an `attributes` object and without splitting out references. The JSON object SHALL be expected to match one of the current generated by-value Lens chart schemas, including that schema's required fields such as chart `type` and `time_range` where applicable.

For by-reference panels, the resource SHALL set the API `config.ref_id` field from `by_reference.ref_id`, set the API `config.time_range` object from `by_reference.time_range`, and include `references`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns` only when their corresponding Terraform attributes are set. `references_json` SHALL map to the API `references` array of `{ id, name, type }` objects. A saved Lens visualization reference SHALL be represented through `references_json`, typically with a reference whose `name` matches `ref_id`, whose `type` is `lens`, and whose `id` is the saved object ID.

**On read:**

The resource SHALL determine the panel mode from the API `config` union. When the config decodes as the generated by-reference object with `ref_id` and `time_range`, the resource SHALL populate `by_reference` in state and leave `by_value` null. Otherwise, the resource SHALL preserve the API `config` JSON under `by_value.config_json` and leave `by_reference` null. Fields absent from the API response SHALL not be forced into state.

`by_value.config_json`, `by_reference.references_json`, and `by_reference.drilldowns_json` SHALL use semantic JSON equality for plan comparison. API-injected field ordering SHALL NOT create spurious plan diffs.

The `lens-dashboard-app` panel type is distinct from the existing `vis` Lens panel path. None of the typed Lens panel converters (e.g. `xy_chart_config`, `metric_chart_config` converters), Lens time-range injection via `lensPanelTimeRange()`, or Lens metric default normalization SHALL apply to `lens-dashboard-app` panels. The `lens_dashboard_app_config` block uses its own read and write converters.

#### Scenario: Creation of a by-reference lens-dashboard-app panel

- GIVEN a dashboard configuration containing a `lens-dashboard-app` panel with:
  - `type = "lens-dashboard-app"`
  - `lens_dashboard_app_config.by_reference.ref_id = "panel_0"`
  - `lens_dashboard_app_config.by_reference.references_json = "[{\"id\":\"abc-123\",\"name\":\"panel_0\",\"type\":\"lens\"}]"`
  - `lens_dashboard_app_config.by_reference.time_range.from = "now-15m"`
  - `lens_dashboard_app_config.by_reference.time_range.to = "now"`
  - `lens_dashboard_app_config.by_reference.title = "My Shared Visualization"`
- WHEN the resource is created
- THEN the provider SHALL send a panel payload with `config.ref_id = "panel_0"`, the references array, the time range object, and `title = "My Shared Visualization"` to the Kibana dashboard API
- AND the panel SHALL appear in state with `by_reference.ref_id = "panel_0"` and `by_value` as null
- AND the provider SHALL NOT populate panel-level `config_json` for this panel in state

#### Scenario: Creation of a by-value lens-dashboard-app panel

- GIVEN a dashboard configuration containing a `lens-dashboard-app` panel with:
  - `type = "lens-dashboard-app"`
  - `lens_dashboard_app_config.by_value.config_json = "<valid generated API Lens chart config JSON>"`
- WHEN the resource is created
- THEN the provider SHALL send the decoded JSON object directly as the panel API `config`
- AND the panel SHALL appear in state with `by_value.config_json` populated and `by_reference` as null

#### Scenario: by-reference panel with required time_range and optional drilldowns_json

- GIVEN a `lens-dashboard-app` panel in by-reference mode with:
  - `lens_dashboard_app_config.by_reference.ref_id = "panel_0"`
  - `lens_dashboard_app_config.by_reference.time_range.from = "now-7d"`
  - `lens_dashboard_app_config.by_reference.time_range.to = "now"`
  - `lens_dashboard_app_config.by_reference.time_range.mode = "relative"`
  - `lens_dashboard_app_config.by_reference.drilldowns_json = "[{\"type\":\"url_drilldown\",\"trigger\":\"on_click_value\",\"label\":\"Open\",\"url\":\"https://example.com\"}]"`
- WHEN the resource is created or updated
- THEN the provider SHALL include the `time_range` object and `drilldowns` array in the API payload
- AND on read-back the provider SHALL repopulate both from the API response

#### Scenario: Invalid mixed configuration — both sub-blocks set

- GIVEN a `lens_dashboard_app_config` block with both `by_value` and `by_reference` configured
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating mutual exclusivity

#### Scenario: Read-back detects by-reference mode from API response

- GIVEN a managed `lens-dashboard-app` panel authored in by-reference mode
- WHEN Kibana returns the panel config with `ref_id` and `time_range`
- THEN the provider SHALL populate `by_reference` in state and leave `by_value` as null
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: Read-back preserves absent optional by-reference fields

- GIVEN a managed `lens-dashboard-app` panel in by-reference mode that omits `description`, `hide_title`, and `hide_border`
- WHEN Kibana returns the panel without those optional fields
- THEN the provider SHALL keep those optional fields null/unset in state
- AND SHALL NOT create a spurious diff on the next plan
