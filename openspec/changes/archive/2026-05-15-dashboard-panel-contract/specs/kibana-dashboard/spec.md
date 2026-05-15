## ADDED Requirements

### Requirement: Registry-driven simple panel handler architecture preserves dashboard behavior (REQ-044)

The `elasticstack_kibana_dashboard` resource implementation SHALL support a registry-driven handler architecture for simple panel types while preserving the user-visible dashboard behavior defined elsewhere in this capability. A simple panel type is one whose typed configuration does not require internal Lens chart dispatch or by-value/by-reference composite branching.

For each supported simple panel type, the implementation SHALL provide a dedicated handler responsible for that panel type's schema attribute construction, API-to-state mapping, state-to-API mapping, configuration validation, and any panel-specific state alignment. The registry SHALL be the authoritative source used to assemble simple panel schema attributes, route panel reads from API discriminator to handler, route typed panel writes from configured block to handler, and dispatch panel-specific validation.

This architectural change SHALL preserve existing Terraform-facing behavior for the migrated simple panel types, including schema shape, validation outcomes, null-preservation behavior, pinned-panel behavior where applicable, and API round-tripping.

#### Scenario: Simple panel read routing is resolved through the registry

- GIVEN a dashboard API response containing a supported simple panel type
- WHEN the provider reads the panel
- THEN the provider SHALL resolve the panel type through the handler registry
- AND the matching handler SHALL populate the Terraform panel state for that panel type

#### Scenario: Simple panel write routing is resolved through the registry

- GIVEN a Terraform panel configuration for a supported simple panel type
- WHEN the provider builds the dashboard API request
- THEN the provider SHALL resolve the configured typed panel block through the handler registry
- AND the matching handler SHALL build the panel API payload

#### Scenario: Migrated simple panels preserve prior Terraform behavior

- GIVEN a dashboard that uses a migrated simple panel type
- WHEN the provider plans, applies, reads, and refreshes that dashboard after the handler migration
- THEN the panel SHALL preserve the same user-visible schema and runtime behavior as before the migration

#### Scenario: Pinned control panels continue to round-trip through typed handlers

- GIVEN a dashboard with pinned control panels of a migrated control type
- WHEN the provider reads or writes the dashboard
- THEN pinned panel conversion SHALL delegate through the migrated typed handler path
- AND pinned panel Terraform behavior SHALL remain unchanged
