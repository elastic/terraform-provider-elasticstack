## MODIFIED Requirements

### Requirement: kbapi SLO models replace generated/slo in domain layer (REQ-032)

`internal/models/slo.go` and all SLO Terraform model and indicator mapping code under
`internal/kibana/slo` SHALL use the `kbapi.KibanaHTTPAPIsSLO*`-prefixed request and response types
produced by the regenerated OpenAPI client (including the renamed equivalents of
`SLOsSloWithSummaryResponse`, `SLOsSloWithSummaryResponse_Indicator`, `SLOsCreateSloRequest`,
`SLOsUpdateSloRequest`, `SLOsGroupBy`, `SLOsSettings`, `SLOsTimeWindow`, `SLOsKqlWithFilters`, and the
related indicator structs) instead of the retired `kbapi.SLOs*` types or `generated/slo` types.
Indicator mapping SHALL preserve the same logical JSON payloads as the pre-migration implementation for
each indicator variant, even where the underlying generated Go type's shape changed (for example
`TimeWindow`, `KqlWithFilters`, and the APM indicator params becoming `oneOf` union wrappers).

#### Scenario: Model compile-time binding to kbapi

- **WHEN** a developer builds the provider after the migration
- **THEN** `internal/models/slo.go` and `internal/kibana/slo` SHALL NOT reference any retired
  `kbapi.SLOs*` type name
- **AND** `internal/models/slo.go` and `internal/kibana/slo` SHALL NOT import `generated/slo`

#### Scenario: Reshaped TimeWindow produces the same wire JSON as before the migration

- **GIVEN** a `time_window` block with a given `duration` and `type`
- **WHEN** the provider builds the create or update request body after the migration
- **THEN** the request body's `timeWindow` field SHALL be structurally and semantically identical to
  the pre-migration request body for the same Terraform configuration

#### Scenario: Reshaped KqlWithFilters union produces the same wire JSON as before the migration

- **GIVEN** a `kql_custom_indicator` configured with either a plain KQL string filter or an object-form
  filter with `kql_query`/`filters`
- **WHEN** the provider builds the create or update request body after the migration
- **THEN** the request body's filter field SHALL be structurally and semantically identical to the
  pre-migration request body for the same Terraform configuration

#### Scenario: Reshaped APM indicator union produces the same wire JSON as before the migration

- **GIVEN** an `apm_latency_indicator` or `apm_availability_indicator` block
- **WHEN** the provider builds the create or update request body after the migration
- **THEN** the request body's indicator `params` SHALL be structurally and semantically identical to the
  pre-migration request body for the same Terraform configuration, despite now being constructed
  through a `oneOf` union wrapper's `From*` accessor instead of a plain struct literal

## ADDED Requirements

### Requirement: Objective and threshold fields preserve float64 precision after the type migration

After migrating to the regenerated `kbapi.KibanaHTTPAPIsSLO*` types, `objective.target`,
`objective.timeslice_target`, APM latency `threshold`, histogram good/total `from`/`to`, and timeslice
metric threshold/percentile fields SHALL remain bound to Go `float64` in the generated client and SHALL
round-trip through create, read, and update without precision loss.

#### Scenario: Fractional objective target round-trips without precision loss

- **GIVEN** an SLO configured with `objective.target = 0.999`
- **WHEN** the provider creates the SLO and then reads it back
- **THEN** the state value for `objective.target` SHALL remain `0.999`, not a value altered by `float32`
  truncation

### Requirement: sli.synthetics.availability indicator type remains unsupported

This migration SHALL NOT add a Terraform schema block, model mapping, or client support for the
`sli.synthetics.availability` SLO indicator type newly exposed by the regenerated OpenAPI spec. The
resource's set of supported indicator blocks SHALL remain unchanged from before this migration.

#### Scenario: No new indicator block is introduced

- **WHEN** a developer inspects `internal/kibana/slo/schema.go` after this migration
- **THEN** it SHALL expose the same set of indicator blocks (`metric_custom_indicator`,
  `histogram_custom_indicator`, `apm_latency_indicator`, `apm_availability_indicator`,
  `kql_custom_indicator`, `timeslice_metric_indicator`) as before this migration, with no
  `synthetics_availability_indicator` block added
