# provider-framework-entity-core Specification

## Purpose

Canonical requirements for the shared Plugin Framework **entity core** (`internal/entitycore`): Terraform type-name construction from typed stack components for resources and data sources, provider client-factory wiring via `Configure`, and the rule that the core does not implement entity-kind-specific behavior. Pilot resources and data sources embed this core to avoid duplicated boilerplate.

## Requirements

### Requirement: Embedded entity core constructs Terraform type names from typed namespace parts

The provider SHALL provide a shared Plugin Framework **resource** substrate `entitycore.ResourceBase` and a shared Plugin Framework **data source** substrate `entitycore.DataSourceBase`. Each substrate SHALL construct Terraform type names from the configured provider type name, a typed component namespace, and a literal name suffix. The constructed type name SHALL use the format `<provider_type_name>_<component>_<name>`. Both substrates SHALL share the same `Component` type and SHALL support well-known component constants for `elasticsearch`, `kibana`, `fleet`, and `apm`. The literal name suffix SHALL be passed in unmodified by the caller and SHALL NOT be normalized or derived by the substrate.

#### Scenario: Resource type name is built from component and resource name

- **WHEN** a `ResourceBase` is constructed via `NewResourceBase(ComponentKibana, "agentbuilder_tool")`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_kibana_agentbuilder_tool`

#### Scenario: APM resource type name uses the APM namespace

- **WHEN** a `ResourceBase` is constructed via `NewResourceBase(ComponentAPM, "agent_configuration")`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_apm_agent_configuration`

#### Scenario: Data source type name is built from component and data source name

- **WHEN** a `DataSourceBase` is constructed via `NewDataSourceBase(ComponentElasticsearch, "enrich_policy")`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_elasticsearch_enrich_policy`

#### Scenario: Kibana data source preserves a single-segment literal name

- **WHEN** a `DataSourceBase` is constructed via `NewDataSourceBase(ComponentKibana, "spaces")`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_kibana_spaces`

### Requirement: Embedded entity core provides canonical provider client-factory wiring

Both `ResourceBase` and `DataSourceBase` SHALL store the configured `*clients.ProviderClientFactory` for use by the concrete entity. Each substrate's `Configure` implementation SHALL convert provider data by calling `clients.ConvertProviderDataToFactory` and append any returned diagnostics to the response. If, after appending, the response has any error-level diagnostics, the substrate SHALL not assign a factory from that conversion and SHALL leave unchanged any `*clients.ProviderClientFactory` previously stored by a successful `Configure` call. If there are no error-level diagnostics, it SHALL assign the conversion result (including a nil `*clients.ProviderClientFactory` when `providerData` is nil), replacing any prior stored value. Each substrate SHALL expose access to the stored factory through a `Client()` method rather than a mutable exported field.

#### Scenario: Resource Configure stores the provider client factory

- **WHEN** `ResourceBase.Configure` receives provider data that converts successfully to `*clients.ProviderClientFactory`
- **THEN** the substrate SHALL retain that factory for later access by the concrete resource via `Client()`

#### Scenario: Data source Configure stores the provider client factory

- **WHEN** `DataSourceBase.Configure` receives provider data that converts successfully to `*clients.ProviderClientFactory`
- **THEN** the substrate SHALL retain that factory for later access by the concrete data source via `Client()`

#### Scenario: Resource Configure does not store a client after diagnostic failure

- **WHEN** `ResourceBase.Configure` has appended the conversion diagnostics and the response has error-level diagnostics
- **THEN** the substrate SHALL not assign a factory from that conversion, and SHALL leave unchanged any `*clients.ProviderClientFactory` previously stored by an earlier successful `Configure` call

#### Scenario: Data source Configure does not store a client after diagnostic failure

- **WHEN** `DataSourceBase.Configure` has appended the conversion diagnostics and the response has error-level diagnostics
- **THEN** the substrate SHALL not assign a factory from that conversion, and SHALL leave unchanged any `*clients.ProviderClientFactory` previously stored by an earlier successful `Configure` call

### Requirement: Embedded entity core does not define entity-kind-specific behavior

Neither `ResourceBase` nor `DataSourceBase` SHALL implement behavior beyond `Configure`, `Metadata`, and `Client`. `ResourceBase` SHALL NOT implement `ImportState`, `Schema`, `Create`, `Read`, `Update`, `Delete`, `UpgradeState`, `ValidateConfig`, `ConfigValidators`, `ModifyPlan`, or any other Plugin Framework resource lifecycle method. `DataSourceBase` SHALL NOT implement `Schema`, `Read`, `ConfigValidators`, or `ValidateConfig`. Concrete resources and data sources retain explicit ownership of all such behavior.

#### Scenario: Resource without import remains non-importable

- **WHEN** a concrete resource embeds `*entitycore.ResourceBase` and does not define its own `ImportState`
- **THEN** embedding the substrate SHALL NOT make that resource satisfy `resource.ResourceWithImportState`

#### Scenario: Resource with custom import retains explicit ownership

- **WHEN** a concrete resource embeds `*entitycore.ResourceBase` and also defines its own `ImportState`
- **THEN** the resource's import behavior SHALL remain defined by the explicit concrete method

#### Scenario: Data source schema and read remain on the concrete data source

- **WHEN** a concrete data source embeds `*entitycore.DataSourceBase`
- **THEN** the concrete data source SHALL define its own `Schema` and `Read`, and the substrate SHALL NOT provide defaults for either

### Requirement: Compatible Plugin Framework resources use ResourceBase for bootstrap wiring

For every Plugin Framework resource in this provider whose bootstrap logic is limited to storing a `*clients.ProviderClientFactory`, converting `ProviderData` through the canonical `clients.ConvertProviderDataToFactory` flow, constructing a static Terraform type name from fixed namespace parts, and leaving import/CRUD/state behavior on the concrete resource, the provider SHALL implement that bootstrap wiring by embedding `*entitycore.ResourceBase` instead of re-declaring a `client` field plus resource-local `Configure` and `Metadata` methods. Each migrated resource SHALL initialize the substrate with the component namespace and literal resource-name suffix that preserve its pre-existing Terraform type name exactly, and SHALL keep any explicit `ImportState` behavior on the concrete resource.

#### Scenario: Compatible Fleet resource preserves custom import behavior

- **WHEN** `elasticstack_fleet_agent_download_source` is implemented through the shared resource substrate
- **THEN** it SHALL configure `ResourceBase` with component `fleet` and resource name `agent_download_source`
- **AND** its explicit composite-ID `ImportState` behavior SHALL remain defined on the concrete resource

#### Scenario: Compatible Kibana resource preserves passthrough import behavior

- **WHEN** `elasticstack_kibana_dashboard` is implemented through the shared resource substrate
- **THEN** it SHALL configure `ResourceBase` with component `kibana` and resource name `dashboard`
- **AND** its explicit passthrough `ImportState` behavior SHALL remain defined on the concrete resource

#### Scenario: Compatible resource without import remains non-importable

- **WHEN** a compatible Plugin Framework resource without an explicit `ImportState` is migrated to embed `*entitycore.ResourceBase`
- **THEN** it SHALL continue not to satisfy `resource.ResourceWithImportState`

### Requirement: Compatible Plugin Framework data sources use DataSourceBase for bootstrap wiring

For every Plugin Framework data source in this provider whose bootstrap logic is limited to storing a `*clients.ProviderClientFactory`, converting `ProviderData` through the canonical `clients.ConvertProviderDataToFactory` flow, and constructing a static Terraform type name from fixed namespace parts, the provider SHALL implement that bootstrap wiring by embedding `*entitycore.DataSourceBase` instead of re-declaring a `client` field plus data-source-local `Configure` and `Metadata` methods. Each migrated data source SHALL initialize the substrate with the component namespace and literal data-source-name suffix that preserve its pre-existing Terraform type name exactly, and SHALL keep `Schema` and `Read` defined on the concrete data source. The change `rename-resourcecore-to-entitycore` migrates one data source per stack component (excluding APM, which has no Plugin Framework data sources today). Subsequent changes MAY migrate additional Plugin Framework data sources under this requirement.

#### Scenario: Compatible Elasticsearch data source uses enrich_policy as the literal name suffix

- **WHEN** `elasticstack_elasticsearch_enrich_policy` (data source) is implemented through `DataSourceBase`
- **THEN** it SHALL configure `DataSourceBase` with component `elasticsearch` and data-source name `enrich_policy`
- **AND** its `Schema` and `Read` SHALL remain defined on the concrete data source

#### Scenario: Compatible Kibana data source uses spaces as the literal name suffix

- **WHEN** `elasticstack_kibana_spaces` (data source) is implemented through `DataSourceBase`
- **THEN** it SHALL configure `DataSourceBase` with component `kibana` and data-source name `spaces`
- **AND** its `Schema` and `Read` SHALL remain defined on the concrete data source

#### Scenario: Compatible Fleet data source uses enrollment_tokens as the literal name suffix

- **WHEN** `elasticstack_fleet_enrollment_tokens` (data source) is implemented through `DataSourceBase`
- **THEN** it SHALL configure `DataSourceBase` with component `fleet` and data-source name `enrollment_tokens`
- **AND** its `Schema` and `Read` SHALL remain defined on the concrete data source

#### Scenario: Concrete data source obtains the client through the substrate accessor

- **WHEN** a compatible Plugin Framework data source embeds `*entitycore.DataSourceBase`
- **THEN** its `Read` implementation SHALL obtain the `*clients.ProviderClientFactory` via the promoted `Client()` accessor rather than a duplicated `client` field on the concrete data source
