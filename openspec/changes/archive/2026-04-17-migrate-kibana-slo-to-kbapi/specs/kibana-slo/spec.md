## ADDED Requirements

### Requirement: SLO HTTP via shared kbapi client (REQ-031)

The `elasticstack_kibana_slo` implementation SHALL perform Find SLOs (search/list), Get SLO, Create SLO, Update SLO, and Delete SLO HTTP calls using the generated OpenAPI Kibana package `github.com/elastic/terraform-provider-elasticstack/generated/kbapi` and helper functions under `internal/clients/kibanaoapi` colocated with other Kibana entity helpers. The implementation SHALL NOT use `github.com/elastic/terraform-provider-elasticstack/generated/slo` types or `slo.SloAPI` for those operations after this migration.

#### Scenario: Reads and mutations use kbapi transport

- **WHEN** the provider performs create, read, update, or delete for `elasticstack_kibana_slo`
- **THEN** each corresponding Kibana SLO HTTP request SHALL be executed through the configured `kibanaoapi.Client` (`kbapi.ClientWithResponses`) for the effective Kibana connection (provider default or `kibana_connection` scoped client)

#### Scenario: Legacy generated/slo client not used for SLO

- **WHEN** the provider issues any SLO HTTP request for this resource
- **THEN** the code path SHALL NOT depend on `GetSloClient`, `buildSloClient`, or `SetSloAuthContext` for authentication or transport

### Requirement: kbapi SLO models replace generated/slo in domain layer (REQ-032)

`internal/models/slo.go` and all SLO Terraform model and indicator mapping code under `internal/kibana/slo` SHALL use kbapi SLO request and response types (including `SLOsSloWithSummaryResponse`, `SLOsSloWithSummaryResponse_Indicator`, `SLOsCreateSloRequest`, `SLOsUpdateSloRequest`, `SLOsGroupBy`, `SLOsSettings`, and related indicator structs) instead of types from `generated/slo`. Indicator mapping SHALL preserve the same logical JSON payloads as the pre-migration implementation for each indicator variant.

#### Scenario: Model compile-time binding to kbapi

- **WHEN** a developer builds the provider after the migration
- **THEN** `internal/models/slo.go` and `internal/kibana/slo` SHALL not import `generated/slo`

### Requirement: kibanaoapi SLO helper surface (REQ-033)

The provider SHALL expose focused `kibanaoapi` functions (or methods on `*kibanaoapi.Client`) for at minimum: **Get** SLO by space and id, **Create** SLO, **Update** SLO, **Delete** SLO, and **Find** SLOs (paginated search using the kbapi `FindSlos` operation). Helpers SHALL apply space-aware path editing consistent with other `kibanaoapi` modules and SHALL map HTTP status codes to Terraform diagnostics in line with existing patterns (including treating get-by-id not found as absence where the resource layer expects it).

#### Scenario: Find helper is available for search operations

- **WHEN** code requests a paginated SLO search within a space via the new helper
- **THEN** the helper SHALL call the kbapi find-SLOs operation with `SpaceAwarePathRequestEditor` for that space and return typed results or diagnostics on failure

### Requirement: Version gates and group_by wire format unchanged (REQ-034)

After migration, the resource SHALL continue to satisfy existing compatibility and mapping requirements **without weakening minimum stack versions or error messages**: REQ-014 (`group_by` when non-empty), REQ-015 (multiple `group_by` elements), REQ-016 (`settings.prevent_initial_backfill` when set to a known value), REQ-017 (non-empty `data_view_id` on indicators), REQ-023 (`group_by` JSON string vs array encoding by stack version), and REQ-007 (version must be obtainable before create/update). Version evaluation SHALL use the same effective scoped Kibana client and version source as today.

#### Scenario: Multi group_by still blocked below 8.14

- **WHEN** the resolved stack version is below 8.14.0 and the user configures more than one `group_by` element
- **THEN** the provider SHALL return an unsupported-version error diagnostic and SHALL NOT successfully persist an SLO that violates REQ-015

#### Scenario: prevent_initial_backfill still gated below 8.15

- **WHEN** the resolved stack version is below 8.15.0 and `settings.prevent_initial_backfill` is set to a known value
- **THEN** the provider SHALL return an unsupported-version error diagnostic consistent with REQ-016

#### Scenario: data_view_id still gated below 8.15

- **WHEN** the resolved stack version is below 8.15.0 and any indicator has a non-empty `data_view_id`
- **THEN** the provider SHALL return an unsupported-version error diagnostic consistent with REQ-017

#### Scenario: group_by wire encoding by version

- **WHEN** the stack version is in [8.10.0, 8.14.0) and a single-element `group_by` is sent on create or update
- **THEN** the JSON body SHALL encode `group_by` as a single string as required by REQ-023

- **WHEN** the stack version is at least 8.14.0 and multiple `group_by` elements are sent
- **THEN** the JSON body SHALL encode `group_by` as a JSON array of strings as required by REQ-023
