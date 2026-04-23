## ADDED Requirements

### Requirement: Import state (REQ-023)

The resource SHALL implement `resource.ResourceWithImportState`. The import ID SHALL be the full resource `id` in the format `<cluster_uuid>/<policy_name>`. On import, the resource SHALL set `id` to the provided import ID and SHALL set `execute` to `true` in state (the computed default), so that a subsequent `terraform plan` shows no diff when `execute` is not explicitly configured.

#### Scenario: Import with valid id sets state

- **GIVEN** a valid import ID of the form `<cluster_uuid>/<policy_name>`
- **WHEN** `terraform import` runs
- **THEN** `id` in state SHALL equal the provided import ID
- **AND** `execute` in state SHALL be `true`
- **AND** all other resource attributes SHALL be populated from the API by the subsequent Read call

#### Scenario: Import followed by plan shows no diff

- **GIVEN** a resource was successfully imported with no `execute` configured
- **WHEN** `terraform plan` runs after the import
- **THEN** no attribute differences SHALL be shown and no replacement SHALL be planned

### Requirement: Import acceptance test (REQ-024)

The resource acceptance test suite SHALL include an import step that verifies round-trip import integrity: after importing an existing policy, `terraform plan` SHALL show no diff and all policy attributes SHALL match the pre-import configuration.

#### Scenario: Acceptance test import step

- **GIVEN** an existing enrich policy created in a prior acceptance test step
- **WHEN** an `ImportState: true, ImportStateVerify: true` step runs with the resource's composite ID
- **THEN** all attributes in the imported state SHALL match the originally configured attributes
