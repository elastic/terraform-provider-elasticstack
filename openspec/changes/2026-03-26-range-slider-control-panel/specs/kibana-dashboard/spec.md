## ADDED Requirements

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
