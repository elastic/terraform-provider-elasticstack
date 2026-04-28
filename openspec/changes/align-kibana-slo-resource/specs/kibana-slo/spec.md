# Delta Spec: `elasticstack_kibana_slo` Alignment

Base spec: `openspec/specs/kibana-slo/spec.md`
Last requirement in base spec: REQ-035
This delta introduces: REQ-036 to REQ-040

---

This delta defines the target behavior introduced by change `align-kibana-slo-resource`. It describes the requirements the implementation will satisfy once the change is applied; it is not intended to claim that the current implementation already meets these behaviors.

## ADDED Requirements

### Requirement: Mapping — `kql_custom_indicator` string and object-form KQL inputs (REQ-036)
The `kql_custom_indicator` block SHALL support both the existing string form and an additive object form for `filter`, `good`, and `total`. The object-form attributes SHALL use the names `filter_kql`, `good_kql`, and `total_kql`, and each SHALL model the Kibana KQL object variant with `kql_query` and `filters`. For each logical field, the provider SHALL allow exactly one representation to be configured: either the existing string attribute or the new `_kql` attribute.

On write, when a string form is configured, the provider SHALL serialize the string arm of the generated kbapi union. When a `_kql` object form is configured, the provider SHALL serialize the object arm of the generated kbapi union, including both `kqlQuery` and `filters` when provided. On read, the provider SHALL round-trip object-form responses without discarding `filters`.

Mutual exclusivity SHALL apply to practitioner-configured values. The `_kql` attributes SHALL support computed state so the provider can retain a richer object-form API response when it cannot be represented losslessly by the legacy string attribute alone.

#### Scenario: String-form KQL input remains supported
- **WHEN** `kql_custom_indicator.good` is configured as a string and `good_kql` is unset
- **THEN** the provider SHALL serialize the string arm of the Kibana KQL union for `good`

#### Scenario: Object-form KQL input is serialized
- **WHEN** `kql_custom_indicator.good_kql` is configured with `kql_query` and one or more `filters`
- **THEN** the provider SHALL serialize the object arm of the Kibana KQL union for `good`

#### Scenario: Multiple representations are rejected
- **WHEN** both `kql_custom_indicator.total` and `kql_custom_indicator.total_kql` are configured
- **THEN** the provider SHALL return a plan-time validation error for conflicting representations of the same logical field

#### Scenario: Object-form response preserves filters
- **WHEN** the Kibana API returns object-form `filter`, `good`, or `total` values containing `filters`
- **THEN** the provider SHALL preserve that object-form information in state rather than degrading it to the string-only representation

#### Scenario: Richer response upgrades state representation
- **WHEN** configuration uses the legacy string form for a KQL field and the API response includes an object-form value with `filters`
- **THEN** the provider SHALL retain the `_kql` state representation for that field and SHALL treat the legacy string attribute as unset in state for that field

#### Scenario: Configured representation is preserved when lossless
- **WHEN** configuration uses the legacy string form for a KQL field and the API response is representable losslessly as the same string form
- **THEN** the provider SHALL keep the string representation in state for that field and SHALL leave the corresponding `_kql` attribute unset

### Requirement: Validation — indicator-specific conditional fields (REQ-037)
The provider SHALL validate indicator-specific required and forbidden fields at plan time where the Terraform Plugin Framework can express the rule. This SHALL include nested indicator blocks that are currently only rejected during model conversion and conditional field rules driven by sibling `aggregation` values.

At minimum, plan-time validation SHALL cover:
- `metric_custom_indicator.good` and `metric_custom_indicator.total`
- `histogram_custom_indicator.good` and `histogram_custom_indicator.total`
- `timeslice_metric_indicator.metric`
- `metric_custom_indicator.{good,total}.metrics[].field` required unless `aggregation = "doc_count"`
- `metric_custom_indicator.{good,total}.metrics[].field` forbidden when `aggregation = "doc_count"`
- `timeslice_metric_indicator.metric.metrics[].field` required for aggregations that require a field
- `timeslice_metric_indicator.metric.metrics[].field` forbidden when `aggregation = "doc_count"`
- `timeslice_metric_indicator.metric.metrics[].percentile` required when `aggregation = "percentile"` and forbidden otherwise

#### Scenario: Missing metric field for non-doc_count is rejected
- **WHEN** a `metric_custom_indicator` metric entry has `aggregation = "sum"` and `field` is unset
- **THEN** the provider SHALL return a plan-time validation error

#### Scenario: Field on doc_count metric is rejected
- **WHEN** a `timeslice_metric_indicator.metric.metrics` entry has `aggregation = "doc_count"` and `field` is set
- **THEN** the provider SHALL return a plan-time validation error

#### Scenario: Missing percentile is rejected
- **WHEN** a `timeslice_metric_indicator.metric.metrics` entry has `aggregation = "percentile"` and `percentile` is unset
- **THEN** the provider SHALL return a plan-time validation error

### Requirement: Validation — simple schema constraints aligned with the API (REQ-038)
The provider SHALL validate simple SLO schema constraints at plan time to match the current generated SLO API contract and generated union types. The resource SHALL:
- restrict `slo_id` to 8 through 36 characters and `[a-zA-Z0-9_-]+`
- restrict `metric_custom_indicator.{good,total}.metrics[].aggregation` to `sum` or `doc_count`
- restrict metric `name` fields that map to the generated metric unions to `^[A-Z]$`
- restrict `time_window.type` to `rolling` or `calendarAligned`

#### Scenario: Oversized slo_id is rejected
- **WHEN** `slo_id` is set to a value longer than 36 characters
- **THEN** the provider SHALL return a plan-time validation error

#### Scenario: Invalid custom metric aggregation is rejected
- **WHEN** a `metric_custom_indicator` metric entry uses `aggregation = "avg"`
- **THEN** the provider SHALL return a plan-time validation error

#### Scenario: Invalid metric name is rejected
- **WHEN** a metric entry uses `name = "AA"`
- **THEN** the provider SHALL return a plan-time validation error

#### Scenario: Invalid time window type is rejected
- **WHEN** `time_window.type` is set to a value other than `rolling` or `calendarAligned`
- **THEN** the provider SHALL return a plan-time validation error

### Requirement: Mapping — `artifacts` field (REQ-039)
The resource SHALL expose the SLO `artifacts` field using the shape currently modeled by the filtered Kibana spec. The provider SHALL support `artifacts.dashboards[].id` on create and update, and SHALL round-trip the same structure from read responses.

#### Scenario: Artifacts are sent on create
- **WHEN** the configuration includes `artifacts` with dashboard references
- **THEN** the create request SHALL include those dashboard references in the SLO `artifacts` payload

#### Scenario: Artifacts are updated from read
- **WHEN** the Kibana API returns dashboard references under `artifacts`
- **THEN** the provider SHALL populate the Terraform `artifacts` state with those references

### Requirement: Management — `enabled` state (REQ-040)
The resource SHALL expose SLO enabled state as a managed Terraform attribute. Because the generated update request model does not include `enabled`, the provider SHALL manage write reconciliation through the dedicated Kibana enable and disable SLO APIs rather than by extending the update request body.

If `enabled` is omitted from configuration, the provider SHALL preserve server behavior rather than forcing a value. In that case, it SHALL read the returned enabled state into Terraform state and SHALL NOT call the enable or disable APIs solely to normalize the value.

#### Scenario: Disabled SLO is disabled after create
- **WHEN** configuration sets `enabled = false` and the created SLO is initially enabled
- **THEN** the provider SHALL call the Kibana disable SLO API before the final read-back

#### Scenario: Enabled SLO is re-enabled on update
- **WHEN** configuration sets `enabled = true` for an existing disabled SLO
- **THEN** the provider SHALL call the Kibana enable SLO API and SHALL read the SLO again before writing final state

#### Scenario: Enabled unset preserves server behavior
- **WHEN** `enabled` is omitted from configuration
- **THEN** the provider SHALL record the server-returned enabled value in state and SHALL NOT issue enable or disable API calls only to force a default

## MODIFIED Requirements

### Requirement: Update flow (REQ-019)
On update, the resource SHALL convert the Terraform plan to an API model and call the Kibana Update SLO API using the `slo_id` and `space_id` from the current composite `id`. If the planned `enabled` value differs from the server state after the definition update, the resource SHALL call the dedicated Kibana enable or disable SLO API as needed. The resource SHALL perform a read-back after successful write operations to populate computed fields into state.

#### Scenario: Update calls API and reads back
- **WHEN** an existing SLO has a changed `name` in the Terraform plan
- **THEN** the provider SHALL call the Kibana Update SLO API and SHALL perform a subsequent get to populate computed fields in state

#### Scenario: Update reconciles enabled state through dedicated APIs
- **WHEN** an existing SLO changes only its `enabled` value in the Terraform plan
- **THEN** the provider SHALL call the Kibana enable or disable SLO API instead of attempting to send `enabled` in the update request body

### Requirement: Read flow (REQ-020)
On read, the resource SHALL parse the composite `id` from state to extract `space_id` and `slo_id`, then call the Kibana Get SLO API. If the API returns HTTP 404, the resource SHALL remove itself from state without error. On a successful response, the resource SHALL update all state attributes from the API response, including `enabled`, `artifacts`, and any supported `settings` members.

#### Scenario: Successful read maps all attributes
- **WHEN** a valid get-SLO API response is returned
- **THEN** all supported attributes, including `name`, `description`, `budgeting_method`, `time_window`, `objective`, `indicator`, `settings`, `group_by`, `tags`, `slo_id`, `space_id`, `enabled`, and `artifacts`, SHALL be updated in state from the response

### Requirement: Mapping — `settings` block (REQ-024)
The `settings` block uses `UseStateForUnknown` on its object plan modifier. When the `settings` block is configured, the resource SHALL send `sync_delay`, `frequency`, `prevent_initial_backfill`, and `sync_field` where those values are known. When the `settings` block is not configured, no settings SHALL be sent. After reading from the API, if the `settings` block was previously configured in state, the resource SHALL update the `settings` object in state from the API response; if it was not configured, `settings` SHALL remain null in state.

#### Scenario: Settings omitted when not configured
- **WHEN** create runs with no `settings` block in configuration
- **THEN** the create request SHALL NOT include a `settings` payload

#### Scenario: Sync field is sent when configured
- **WHEN** `settings.sync_field` is configured with a known value
- **THEN** the provider SHALL include `syncField` in the SLO settings payload

### Requirement: Mapping — `metric_custom_indicator` doc_count aggregation (REQ-035)
When a `metric_custom_indicator` metric entry uses `aggregation = "doc_count"`, the `field` attribute SHALL be optional and the provider SHALL NOT send `field` to the Kibana API. For all other supported `metric_custom_indicator` aggregation types, `field` SHALL be required. The schema SHALL declare `field` as optional for `metric_custom_indicator.{good,total}.metrics`, SHALL validate `aggregation` as `sum` or `doc_count`, and SHALL validate metric `name` values against the generated union contract.

On write, when `aggregation = "doc_count"` the provider SHALL serialise the metric using the no-field API variant. For all other supported aggregations the provider SHALL use the field-bearing API variant. After a successful read-back, when the API returns a `doc_count` metric the provider SHALL store `field = null` in state.

#### Scenario: Doc count omits field
- **WHEN** a `metric_custom_indicator` good or total metric has `aggregation = "doc_count"` and `field` is not set
- **THEN** the provider SHALL serialize the no-field API variant and SHALL accept the configuration

#### Scenario: Doc count round-trips with null field
- **WHEN** the Kibana API returns a `metric_custom_indicator` metric with `aggregation = "doc_count"`
- **THEN** the provider SHALL store `field = null` in state for that metric

#### Scenario: Non-doc_count requires field
- **WHEN** a `metric_custom_indicator` metric has `aggregation != "doc_count"` and a non-null `field`
- **THEN** the provider SHALL serialize the field-bearing API variant for that metric
