# Delta Spec: `lens-dashboard-app` Panel Support

Base spec: `openspec/specs/kibana-dashboard/spec.md`
Last requirement in base spec: REQ-025
This delta introduces: REQ-035

---

## Schema additions

The following block is added to the panel object within the `panels` list (and within `sections[*].panels`):

```hcl
lens_dashboard_app_config = <optional, object({
  # Exactly one of by_value or by_reference must be set

  by_value = <optional, object({
    attributes_json = <required, json string, normalized>  # full Lens chart attributes object
    references_json = <optional, json string, normalized>  # array of { id: string, name: string, type: string }
  })>

  by_reference = <optional, object({
    saved_object_id = <required, string>  # ID of the saved Lens visualization saved object
    overrides_json  = <optional, json string, normalized>  # JSON object for overrides to the saved Lens object
  })>

  # Shared optional fields
  title       = <optional, string>
  description = <optional, string>
  hide_title  = <optional, bool>
  hide_border = <optional, bool>

  time_range = <optional, object({
    from = <required, string>
    to   = <required, string>
  })>
})> # only with type = "lens-dashboard-app"; conflicts with all other config blocks; exactly one of by_value or by_reference must be set
```

**Distinction from existing `lens` panel type**: The `lens_dashboard_app_config` block applies exclusively to panels with `type = "lens-dashboard-app"`. Panels with `type = "lens"` continue to use the existing typed config blocks (`xy_chart_config`, `metric_chart_config`, `waffle_config`, etc.) and `config_json`. The type string `lens-dashboard-app` must appear verbatim in the panel `type` attribute; it is not interchangeable with `lens`.

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
- `by_value.attributes_json` SHALL be required when `by_value` is set.
- `by_reference.saved_object_id` SHALL be required when `by_reference` is set.

#### Scenario: lens_dashboard_app_config rejected for non-lens-dashboard-app panel (ADDED)

- GIVEN a panel with `type = "lens"` and `lens_dashboard_app_config` set
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

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, or omitting all panel configuration blocks, SHALL return an error diagnostic.

is updated to:

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, including `lens-dashboard-app`, or omitting all panel configuration blocks, SHALL return an error diagnostic. The `lens-dashboard-app` panel type SHALL be managed exclusively through the typed `lens_dashboard_app_config` block.

#### Scenario: config_json rejected for lens-dashboard-app panel type (ADDED)

- GIVEN a panel with `type = "lens-dashboard-app"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `lens-dashboard-app`

---

## ADDED Requirements

### Requirement: `lens-dashboard-app` panel behavior (REQ-035)

For `type = "lens-dashboard-app"` panels, the resource SHALL accept `lens_dashboard_app_config` with exactly one of the `by_value` or `by_reference` sub-blocks set. Within `by_value`, the `attributes_json` attribute is required. Within `by_reference`, the `saved_object_id` attribute is required. The optional shared attributes `title`, `description`, `hide_title`, `hide_border`, and `time_range` MAY be set in either mode.

**On write (create and update):**

For by-value panels, the resource SHALL map `by_value.attributes_json` to the `attributes` field in the API payload, and SHALL include `references` from `by_value.references_json` when set. For by-reference panels, the resource SHALL set `saved_object_id` in the API payload from `by_reference.saved_object_id`, and SHALL include `overrides` from `by_reference.overrides_json` when set.

In both modes, the resource SHALL include the shared optional fields (`title`, `description`, `hide_title`, `hide_border`, `time_range`) in the API payload only when they are set in Terraform state. Absent optional fields SHALL NOT be sent to the API.

**On read:**

The resource SHALL determine the panel mode by inspecting the API response: the presence of an `attributes` key indicates by-value mode; the presence of a `saved_object_id` key indicates by-reference mode. The resource SHALL populate the corresponding sub-block in state and leave the other sub-block as null. Fields absent from the API response SHALL not be forced into state.

`attributes_json`, `references_json`, and `overrides_json` SHALL use default-aware semantic JSON equality for plan comparison. API-injected field ordering or default field additions SHALL NOT create spurious plan diffs.

The `lens-dashboard-app` panel type is distinct from the `lens` panel type. None of the typed Lens panel converters (e.g. `xy_chart_config`, `metric_chart_config` converters), Lens time-range injection via `lensPanelTimeRange()`, or Lens metric default normalization SHALL apply to `lens-dashboard-app` panels. The `lens_dashboard_app_config` block uses its own read and write converters.

#### Scenario: Creation of a by-reference lens-dashboard-app panel

- GIVEN a dashboard configuration containing a `lens-dashboard-app` panel with:
  - `type = "lens-dashboard-app"`
  - `lens_dashboard_app_config.by_reference.saved_object_id = "abc-123"`
  - `lens_dashboard_app_config.title = "My Shared Visualization"`
- WHEN the resource is created
- THEN the provider SHALL send a panel payload with `saved_object_id = "abc-123"` and `title = "My Shared Visualization"` to the Kibana dashboard API
- AND the panel SHALL appear in state with `by_reference.saved_object_id = "abc-123"` and `by_value` as null
- AND the provider SHALL NOT populate `config_json` for this panel in state

#### Scenario: Creation of a by-value lens-dashboard-app panel

- GIVEN a dashboard configuration containing a `lens-dashboard-app` panel with:
  - `type = "lens-dashboard-app"`
  - `lens_dashboard_app_config.by_value.attributes_json = "<valid Lens chart JSON>"`
  - `lens_dashboard_app_config.by_value.references_json = "[{\"id\": \"dv-1\", \"name\": \"indexpattern-datasource-layer-abc\", \"type\": \"index-pattern\"}]"`
- WHEN the resource is created
- THEN the provider SHALL send a panel payload with the `attributes` object and `references` array to the Kibana dashboard API
- AND the panel SHALL appear in state with `by_value.attributes_json` and `by_value.references_json` populated and `by_reference` as null

#### Scenario: by-reference panel with time_range and overrides_json

- GIVEN a `lens-dashboard-app` panel in by-reference mode with:
  - `lens_dashboard_app_config.by_reference.saved_object_id = "xyz-456"`
  - `lens_dashboard_app_config.by_reference.overrides_json = "{\"timeRange\": {\"from\": \"now-7d\", \"to\": \"now\"}}"`
  - `lens_dashboard_app_config.time_range.from = "now-7d"`
  - `lens_dashboard_app_config.time_range.to = "now"`
- WHEN the resource is created or updated
- THEN the provider SHALL include the `time_range` object and `overrides` object in the API payload
- AND on read-back the provider SHALL repopulate both from the API response

#### Scenario: Invalid mixed configuration — both sub-blocks set

- GIVEN a `lens_dashboard_app_config` block with both `by_value` and `by_reference` configured
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating mutual exclusivity

#### Scenario: Read-back detects by-reference mode from API response

- GIVEN a managed `lens-dashboard-app` panel authored in by-reference mode
- WHEN Kibana returns the panel with a `saved_object_id` field and no `attributes` field
- THEN the provider SHALL populate `by_reference` in state and leave `by_value` as null
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: Read-back preserves absent optional shared fields

- GIVEN a managed `lens-dashboard-app` panel in by-reference mode that omits `description` and `time_range`
- WHEN Kibana returns the panel without those optional fields
- THEN the provider SHALL keep `description` and `time_range` as null/unset in state
- AND SHALL NOT create a spurious diff on the next plan
