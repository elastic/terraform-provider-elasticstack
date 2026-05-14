## ADDED Requirements

### Requirement: Lens visualization converter registry preserves typed chart behavior (REQ-045)

The `elasticstack_kibana_dashboard` resource implementation SHALL support a Lens visualization converter registry for typed Lens chart handling while preserving the Terraform-facing behavior of all supported Lens chart blocks. Each supported Lens chart kind SHALL be handled by a dedicated converter that participates in shared Lens chart read, write, defaulting, and state-alignment flows through a common registry.

The registry SHALL be the authoritative source used to classify typed Lens by-value chart payloads, dispatch API-to-state conversion for supported Lens chart kinds, dispatch state-to-API conversion for supported Lens chart kinds, and apply Lens chart defaulting or state-alignment behavior that is specific to a chart kind.

This architectural change SHALL NOT alter the user-visible schema or runtime behavior of existing supported Lens chart blocks under dashboard panel configurations.

#### Scenario: Typed Lens chart read conversion is resolved through the converter registry

- GIVEN a dashboard panel API payload for a supported by-value Lens chart kind
- WHEN the provider reads that panel into Terraform state
- THEN the provider SHALL resolve the chart kind through the Lens converter registry
- AND the matching converter SHALL populate the corresponding typed Lens chart block in state

#### Scenario: Typed Lens chart write conversion is resolved through the converter registry

- GIVEN a Terraform configuration that selects a supported typed Lens chart block
- WHEN the provider builds the dashboard API payload for that panel
- THEN the provider SHALL resolve the configured chart block through the Lens converter registry
- AND the matching converter SHALL build the corresponding by-value Lens chart payload

#### Scenario: Lens chart behavior remains unchanged after converter extraction

- GIVEN a dashboard using a supported typed Lens chart block
- WHEN the provider plans, applies, reads, and refreshes that dashboard after the converter migration
- THEN the typed Lens chart block SHALL preserve the same user-visible schema, validation behavior, defaulting behavior, null-preservation behavior, and API round-tripping as before the migration
