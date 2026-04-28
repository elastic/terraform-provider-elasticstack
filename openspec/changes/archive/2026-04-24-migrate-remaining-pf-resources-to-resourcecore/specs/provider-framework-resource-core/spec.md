## ADDED Requirements

### Requirement: Compatible Plugin Framework resources use the shared resource core for bootstrap wiring
For every Plugin Framework resource in this provider whose bootstrap logic is limited to storing a `*clients.ProviderClientFactory`, converting `ProviderData` through the canonical `clients.ConvertProviderDataToFactory` flow, constructing a static Terraform type name from fixed namespace parts, and leaving import/CRUD/state behavior on the concrete resource, the provider SHALL implement that bootstrap wiring by embedding `internal/resourcecore.Core` instead of re-declaring a `client` field plus resource-local `Configure` and `Metadata` methods. Each migrated resource SHALL initialize the core with the component namespace and literal resource-name suffix that preserve its pre-existing Terraform type name exactly, and SHALL keep any explicit `ImportState` behavior on the concrete resource.

#### Scenario: Compatible Fleet resource preserves custom import behavior
- **WHEN** `elasticstack_fleet_agent_download_source` is implemented through the shared resource core
- **THEN** it SHALL configure the core with component `fleet` and resource name `agent_download_source`
- **AND** its explicit composite-ID `ImportState` behavior SHALL remain defined on the concrete resource

#### Scenario: Compatible Kibana resource preserves passthrough import behavior
- **WHEN** `elasticstack_kibana_dashboard` is implemented through the shared resource core
- **THEN** it SHALL configure the core with component `kibana` and resource name `dashboard`
- **AND** its explicit passthrough `ImportState` behavior SHALL remain defined on the concrete resource

#### Scenario: Compatible resource without import remains non-importable
- **WHEN** a compatible Plugin Framework resource without an explicit `ImportState` is migrated to embed the shared resource core
- **THEN** it SHALL continue not to satisfy `resource.ResourceWithImportState`
