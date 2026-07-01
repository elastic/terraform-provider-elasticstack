## ADDED Requirements

### Requirement: AIOps log rate analysis panel behavior (REQ-047)

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

### Requirement: AIOps pattern analysis panel behavior (REQ-048)

The `elasticstack_kibana_dashboard` resource SHALL support a panel of `type = "aiops_pattern_analysis"` via an `aiops_pattern_analysis_config` block. The config block SHALL accept:

- `data_view_id` (required string): the data view ID used for pattern analysis.
- `field_name` (required string): the text field on which to run pattern analysis.
- `minimum_time_range` (optional string, enum): one of `no_minimum`, `1_week`, `1_month`, `3_months`, `6_months`. Invalid values SHALL be rejected at plan time.
- `random_sampler_mode` (optional string, enum): one of `off`, `on_automatic`, `on_manual`. Invalid values SHALL be rejected at plan time.
- `random_sampler_probability` (optional float64): the sampling probability, bounded to `[0.00001, 0.5]`. Values outside this range SHALL be rejected at plan time. This field is only meaningful when `random_sampler_mode = "on_manual"`.
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

### Requirement: AIOps change point chart panel behavior (REQ-049)

The `elasticstack_kibana_dashboard` resource SHALL support a panel of `type = "aiops_change_point_chart"` via an `aiops_change_point_chart_config` block. The config block SHALL accept:

- `data_view_id` (required string): the data view ID used for change point detection.
- `metric_field` (required string): the metric field used by the aggregation function.
- `aggregation_function` (optional string, enum): one of `avg`, `max`, `min`, `sum`. Invalid values SHALL be rejected at plan time.
- `split_field` (optional string): the optional field used to split change-point results.
- `partitions` (optional set of strings): optional split field values to include in the panel. Modelled as a set to prevent plan drift from API-returned ordering. Semantically a filter set; duplicate entries are silently deduplicated. An empty set is not meaningful (omit the attribute to disable filtering); a non-null set SHALL contain at least one entry and SHALL be rejected at plan time otherwise.
- `max_series_to_plot` (optional float64): maximum number of change points to visualise. Kibana default is 6. The resource SHALL null-preserve this field when the user omitted it.
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
