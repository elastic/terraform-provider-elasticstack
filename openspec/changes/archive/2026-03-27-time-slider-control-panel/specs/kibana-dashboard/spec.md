## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

Schema validation SHALL enforce that each typed panel config block is only present on a panel whose `type` matches that block's panel type, and that at most one typed config block is present on any panel. This exclusivity requirement now applies to `time_slider_control_config` in addition to all previously supported typed config blocks.

#### Scenario: Reject conflicting time slider config blocks

- GIVEN a panel with `type = "time_slider_control"`
- AND `time_slider_control_config` is set together with another typed panel config block or practitioner-authored `config_json`
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return a plan-time validation error describing the conflict or unsupported configuration
