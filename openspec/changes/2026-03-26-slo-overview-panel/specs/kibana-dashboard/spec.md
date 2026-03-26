## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

Changes to `space_id` or `dashboard_id` SHALL require replacement rather than in-place update. The schema SHALL reject `time_range_mode` values other than `absolute` or `relative`, reject `access_control.access_mode` values other than `write_restricted` or `default`, and reject configurations that set both `query_text` and `query_json`. Each panel SHALL declare at least one panel configuration block, panel configuration blocks SHALL be mutually exclusive, typed panel configuration blocks SHALL only be valid for their supported panel type, and `waffle_config` SHALL enforce its ES|QL-vs-non-ES|QL field consistency rules. For `slo_overview_config`, the schema SHALL additionally enforce that exactly one of the `single` or `groups` nested blocks is present, that `single.slo_id` is provided when the `single` block is configured, and that `groups.group_filters.group_by` is one of the documented enum values when set.

#### Scenario: Query conflict

- GIVEN configuration sets both `query_text` and `query_json`
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: Panel configuration mismatch

- GIVEN a panel with `type = "markdown"` and `xy_chart_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before apply

#### Scenario: SLO overview config mode conflict

- GIVEN a panel with `type = "slo_overview"` and both `single` and `groups` blocks configured inside `slo_overview_config`
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: SLO overview single mode missing slo_id

- GIVEN a panel with `type = "slo_overview"` whose `slo_overview_config` contains a `single` block without `slo_id`
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: Invalid group_by value

- GIVEN a panel with `type = "slo_overview"` whose `groups.group_filters.group_by` is set to a value not in `["slo.tags", "status", "slo.indicator.type", "_index"]`
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

The resource SHALL support top-level `panels`, section-contained `panels`, and `sections` in the order returned by the API and the order given in configuration when building requests. For panel reads, it SHALL distinguish sections from top-level panels and map each panel's `type`, `grid`, optional `id`, and configuration. For typed panel mappings, the resource SHALL seed from prior state or plan so that optional panel attributes omitted by Kibana on read can be preserved. When a panel is managed through `config_json` only, the resource SHALL preserve that JSON-centric representation and SHALL NOT populate typed configuration blocks from the API for that panel. On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, or omitting all panel configuration blocks, SHALL return an error diagnostic. The `slo_overview` panel type SHALL NOT be supported through `config_json`; it SHALL only be managed through the typed `slo_overview_config` block.

#### Scenario: JSON-only lens panel

- GIVEN a panel with `type = "lens"` configured only through `config_json`
- WHEN the dashboard is read back from Kibana
- THEN the provider SHALL keep `config_json` as the round-tripped panel representation and SHALL leave typed panel blocks unset for that panel

#### Scenario: config_json rejected for slo_overview

- GIVEN a panel with `type = "slo_overview"` configured through `config_json`
- WHEN the provider builds the API request
- THEN it SHALL return an error diagnostic for unsupported `config_json` panel type

## ADDED Requirements

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
