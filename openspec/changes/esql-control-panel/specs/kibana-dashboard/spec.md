# Delta Spec: ES|QL Control Panel Support

Base spec: `openspec/specs/kibana-dashboard/spec.md`
Last requirement in base spec: REQ-025

---

## Schema additions

The following block is added to the panel object within the `panels` list (and within `sections[*].panels`):

```hcl
esql_control_config = <optional, object({
  # Required
  selected_options  = <required, list(string)>
  variable_name     = <required, string>
  variable_type     = <required, string>  # enum: fields | values | functions | time_literal | multi_values
  esql_query        = <required, string>
  control_type      = <required, string>  # enum: STATIC_VALUES | VALUES_FROM_QUERY

  # Optional
  title             = <optional, string>
  single_select     = <optional, bool>
  available_options = <optional, list(string)>

  display_settings  = <optional, object({
    placeholder     = <optional, string>
    hide_action_bar = <optional, bool>
    hide_exclude    = <optional, bool>
    hide_exists     = <optional, bool>
    hide_sort       = <optional, bool>
  })>
})> # only with type = "esql_control"; conflicts with all other config blocks
```

---

## MODIFIED: Replacement fields and schema validation (REQ-006)

The existing REQ-006 text is extended. The sentence:

> Each panel SHALL declare at least one panel configuration block, panel configuration blocks SHALL be mutually exclusive, typed panel configuration blocks SHALL only be valid for their supported panel type, and `waffle_config` SHALL enforce its ES|QL-vs-non-ES|QL field consistency rules.

gains the following additions:

- `esql_control_config` SHALL be valid only for panels with `type = "esql_control"`.
- `esql_control_config` SHALL be mutually exclusive with all other panel configuration blocks.
- The `variable_type` attribute within `esql_control_config` SHALL be restricted to the values `fields`, `values`, `functions`, `time_literal`, and `multi_values`; any other value SHALL be rejected at plan time.
- The `control_type` attribute within `esql_control_config` SHALL be restricted to the values `STATIC_VALUES` and `VALUES_FROM_QUERY`; any other value SHALL be rejected at plan time.

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

---

## MODIFIED: Panels and `config_json` round-trip behavior (REQ-010)

The existing REQ-010 text:

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, or omitting all panel configuration blocks, SHALL return an error diagnostic.

is updated to:

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, including `esql_control`, or omitting all panel configuration blocks, SHALL return an error diagnostic. The `esql_control` panel type SHALL be managed exclusively through the typed `esql_control_config` block.

#### Scenario: config_json rejected for esql_control panel type (ADDED)

- GIVEN a panel with `type = "esql_control"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `esql_control`

---

## ADDED: ES|QL control panel behavior (REQ-026)

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
