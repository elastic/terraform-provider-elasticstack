# provider-framework-resource-core Specification

## Purpose

Canonical requirements for the shared Plugin Framework **resource core** (`internal/resourcecore`): Terraform type-name construction from typed stack components, provider client-factory wiring via `Configure`, and the rule that the core does not implement import behavior. Pilot resources embed this core to avoid duplicated boilerplate.
## Requirements
### Requirement: Embedded resource core constructs provider resource type names from typed namespace parts
The provider SHALL provide a shared Plugin Framework resource core that constructs Terraform resource type names from the configured provider type name, a typed component namespace, and a literal resource-name suffix. The constructed type name SHALL use the format `<provider_type_name>_<component>_<resource_name>`. The shared core SHALL support well-known component constants for `elasticsearch`, `kibana`, `fleet`, and `apm`.

#### Scenario: Kibana resource type name is built from component and resource name
- **WHEN** a resource core is configured with component `kibana` and resource name `agentbuilder_tool`
- **THEN** `Metadata` SHALL set the type name to `<provider_type_name>_kibana_agentbuilder_tool`

#### Scenario: APM resource type name uses the APM namespace
- **WHEN** a resource core is configured with component `apm` and resource name `agent_configuration`
- **THEN** `Metadata` SHALL set the type name to `<provider_type_name>_apm_agent_configuration`

### Requirement: Embedded resource core provides canonical provider client-factory wiring
The shared Plugin Framework resource core SHALL store the configured `*clients.ProviderClientFactory` for use by the concrete resource. Its `Configure` implementation SHALL convert provider data by calling `clients.ConvertProviderDataToFactory` and append any returned diagnostics to the response. If, after appending, the response has any error-level diagnostics, the core SHALL not assign a factory from that conversion and SHALL leave unchanged any `*clients.ProviderClientFactory` previously stored by a successful `Configure` call. If there are no error-level diagnostics, it SHALL assign the conversion result (including a nil `*clients.ProviderClientFactory` when `providerData` is nil), replacing any prior stored value. The core SHALL expose access to the stored factory through a method rather than a mutable exported field.

#### Scenario: Configure stores the provider client factory
- **WHEN** `Configure` receives provider data that converts successfully to `*clients.ProviderClientFactory`
- **THEN** the core SHALL retain that factory for later access by the concrete resource

#### Scenario: Configure does not store a client after diagnostic failure
- **WHEN** `Configure` has appended the conversion diagnostics and the response has error-level diagnostics
- **THEN** the core SHALL not assign a factory from that conversion, and SHALL leave unchanged any `*clients.ProviderClientFactory` previously stored by an earlier successful `Configure` call

### Requirement: Embedded resource core does not define import behavior
The shared Plugin Framework resource core SHALL NOT implement `ImportState` or otherwise provide default import behavior. Concrete resources SHALL remain responsible for explicitly defining passthrough import, custom import, or no import support according to their own schema and lifecycle behavior.

#### Scenario: Resource without import remains non-importable
- **WHEN** a concrete resource embeds the shared core and does not define its own `ImportState`
- **THEN** embedding the core SHALL NOT make that resource satisfy `resource.ResourceWithImportState`

#### Scenario: Resource with custom import retains explicit ownership
- **WHEN** a concrete resource embeds the shared core and also defines its own `ImportState`
- **THEN** the resource's import behavior SHALL remain defined by the explicit concrete method

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

