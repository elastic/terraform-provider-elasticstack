## ADDED Requirements

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

**Kibana version compatibility**: The underlying ML embeddable predates the Dashboard API used by this resource (introduced in Kibana 7.9.0), but no minimum Kibana version specific to its typed representation in the Dashboard API's panel schema could be confirmed from release notes. Since `ml_anomaly_swimlane` is already present in the generated `kbapi` client this resource depends on, and no other typed panel block in this resource carries a bespoke per-panel version gate, the provider does not add one for `ml_anomaly_swimlane_config` either; the Kibana Dashboard API SHALL reject the panel on incompatible stack versions.

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

**Kibana version compatibility**: The underlying ML embeddable predates the Dashboard API used by this resource (introduced in Kibana 8.13.0), but no minimum Kibana version specific to its typed representation in the Dashboard API's panel schema could be confirmed from release notes. Since `ml_single_metric_viewer` is already present in the generated `kbapi` client this resource depends on, and no other typed panel block in this resource carries a bespoke per-panel version gate, the provider does not add one for `ml_single_metric_viewer_config` either; the Kibana Dashboard API SHALL reject the panel on incompatible stack versions.

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

## MODIFIED Requirements

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
