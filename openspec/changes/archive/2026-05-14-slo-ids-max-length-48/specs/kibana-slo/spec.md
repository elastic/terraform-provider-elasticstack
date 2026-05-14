## MODIFIED Requirements

### Requirement: Validation — simple schema constraints aligned with the API (REQ-038)
The provider SHALL validate simple SLO schema constraints at plan time to match the current generated SLO API contract and generated union types. The resource SHALL:
- restrict `slo_id` to 8 through 48 characters and `[a-zA-Z0-9_-]+`
- restrict `metric_custom_indicator.{good,total}.metrics[].aggregation` to `sum` or `doc_count`
- restrict metric `name` fields that map to the generated metric unions to `^[A-Z]$`
- restrict `time_window.type` to `rolling` or `calendarAligned`

#### Scenario: Oversized slo_id is rejected
- **WHEN** `slo_id` is set to a value longer than 48 characters
- **THEN** the provider SHALL return a plan-time validation error

#### Scenario: 48-character slo_id is accepted
- **WHEN** `slo_id` is set to a value that is exactly 48 characters long and matches `[a-zA-Z0-9_-]+`
- **THEN** the provider SHALL accept the configuration without a plan-time validation error

#### Scenario: Invalid custom metric aggregation is rejected
- **WHEN** a `metric_custom_indicator` metric entry uses `aggregation = "avg"`
- **THEN** the provider SHALL return a plan-time validation error

#### Scenario: Invalid metric name is rejected
- **WHEN** a metric entry uses `name = "AA"`
- **THEN** the provider SHALL return a plan-time validation error

#### Scenario: Invalid time window type is rejected
- **WHEN** `time_window.type` is set to a value other than `rolling` or `calendarAligned`
- **THEN** the provider SHALL return a plan-time validation error
