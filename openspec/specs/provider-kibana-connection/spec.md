# provider-kibana-connection Specification

## Purpose
TBD - created by archiving change add-kibana-connection-support. Update Purpose after archive.
## Requirements
### Requirement: Entity-local `kibana_connection` schema source of truth
The provider SHALL define shared SDK and Plugin Framework schema helpers for entity-local `kibana_connection`. The SDK entity-local helper SHALL be `internal/schema.GetKibanaEntityConnectionSchema()`, and the Plugin Framework entity-local helper SHALL be `internal/schema.GetKbFWConnectionBlock()`. Those helpers SHALL use the same field set and equivalent validation rules as the provider-level Kibana connection block, SHALL be list-shaped with at most one element, and SHALL NOT expose entity-level deprecation metadata. When path-based validation metadata is required, the entity-local helpers SHALL target the entity-local block path rather than the provider-level `kibana` path.

Note: `internal/schema.GetKibanaConnectionSchema()` is the provider-level helper (used for the `kibana` block in `provider.go`). The entity-level helper uses a distinct name to make the two contracts unambiguous and to keep path-scoped validation metadata correct for each call site. Both functions share a common private implementation (`getKibanaConnectionSchema`) parameterised by key name.

#### Scenario: Shared helpers define the entity-local block shape
- **WHEN** an entity-local `kibana_connection` schema or block is defined
- **THEN** it SHALL come from the shared provider schema helpers rather than an entity-specific variant, and any path-based validation metadata SHALL target the entity-local block path

### Requirement: Framework scoped Kibana client resolution
The provider SHALL expose Plugin Framework Kibana-derived client resolution through `*clients.ProviderClientFactory` methods that return a `*clients.KibanaScopedClient`. When an entity-local `kibana_connection` block is not configured, the factory SHALL return a `*clients.KibanaScopedClient` built from provider-level defaults. When the block is configured, the factory SHALL return a `*clients.KibanaScopedClient` whose Kibana legacy client, Kibana OpenAPI client, SLO client, and Fleet client are rebuilt from the scoped `kibana_connection`. Covered Framework resources that need Kibana-derived operations, including `elasticstack_apm_agent_configuration`, SHALL consume that typed scoped client rather than a broad `*clients.APIClient` adapter.

#### Scenario: Framework factory falls back to provider defaults
- **WHEN** a covered Framework entity resolves its effective Kibana client through the factory and `kibana_connection` is absent
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` derived from provider configuration

#### Scenario: Framework factory builds a scoped Kibana-derived client
- **WHEN** a covered Framework entity resolves its effective Kibana client through the factory and `kibana_connection` is configured
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` rebuilt from that connection for Kibana, SLO, and Fleet operations

#### Scenario: Framework entity does not downcast to a broad client
- **WHEN** a covered Framework entity performs Kibana-derived operations
- **THEN** it SHALL use the typed `*clients.KibanaScopedClient` contract rather than converting provider data into a broad `*clients.APIClient`

#### Scenario: APM agent configuration exposes the shared Kibana connection block
- **WHEN** `elasticstack_apm_agent_configuration` defines `kibana_connection`
- **THEN** it SHALL use the shared Plugin Framework Kibana connection block and resolve its effective typed Kibana client through the factory from either provider defaults or that entity-local override

### Requirement: SDK scoped Kibana client resolution
The provider SHALL expose SDK `kibana_connection` resolution through `*clients.ProviderClientFactory` methods that accept resource or data source state and return a `*clients.KibanaScopedClient`. When the block is not configured, the factory SHALL return a `*clients.KibanaScopedClient` built from provider-level defaults. When the block is configured, the factory SHALL return a `*clients.KibanaScopedClient` whose Kibana legacy client, Kibana OpenAPI client, SLO client, and Fleet client are rebuilt from the scoped `kibana_connection`.

#### Scenario: SDK factory falls back to provider defaults
- **WHEN** an SDK entity resolves its effective Kibana client through the factory and `kibana_connection` is absent
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` derived from provider configuration

#### Scenario: SDK factory builds a scoped Kibana-derived client
- **WHEN** an SDK entity resolves its effective Kibana client through the factory and `kibana_connection` is configured
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` rebuilt from that connection for Kibana, SLO, and Fleet operations

### Requirement: Scoped client version and identity behavior
When an entity uses a scoped `*clients.KibanaScopedClient` resolved from `kibana_connection`, version, flavor, and other Kibana-derived client checks SHALL resolve against the scoped connection rather than provider-level Elasticsearch identity. The `*clients.KibanaScopedClient` SHALL therefore avoid exposing provider-level Elasticsearch identity in a way that can make Kibana or Fleet operations target one cluster while version or identity checks target another.

#### Scenario: Scoped version checks follow the scoped Kibana connection
- **WHEN** an entity uses `ServerVersion()`, `ServerFlavor()`, or equivalent behavior through a scoped `kibana_connection`
- **THEN** the result SHALL be derived from the scoped Kibana connection instead of the provider's Elasticsearch connection

#### Scenario: Scoped Kibana client does not expose provider Elasticsearch identity
- **WHEN** a covered Kibana or Fleet entity uses the `*clients.KibanaScopedClient` returned for `kibana_connection`
- **THEN** that `*clients.KibanaScopedClient` SHALL NOT require provider-level Elasticsearch identity to perform Kibana or Fleet operations

### Requirement: In-scope Kibana entity rollout
The rollout's covered Kibana entities SHALL be exactly the Kibana resources and data sources registered in `provider/provider.go` and `provider/plugin_framework.go`: `elasticstack_kibana_action_connector`, `elasticstack_kibana_agentbuilder_export_workflow`, `elasticstack_kibana_agentbuilder_workflow`, `elasticstack_kibana_alerting_rule`, `elasticstack_kibana_dashboard`, `elasticstack_kibana_data_view`, `elasticstack_kibana_default_data_view`, `elasticstack_kibana_export_saved_objects`, `elasticstack_kibana_import_saved_objects`, `elasticstack_kibana_maintenance_window`, `elasticstack_kibana_security_detection_rule`, `elasticstack_kibana_security_enable_rule`, `elasticstack_kibana_security_exception_item`, `elasticstack_kibana_security_exception_list`, `elasticstack_kibana_security_list`, `elasticstack_kibana_security_list_data_streams`, `elasticstack_kibana_security_list_item`, `elasticstack_kibana_security_role`, `elasticstack_kibana_slo`, `elasticstack_kibana_space`, `elasticstack_kibana_spaces`, `elasticstack_kibana_stream`, `elasticstack_kibana_synthetics_monitor`, `elasticstack_kibana_synthetics_parameter`, and `elasticstack_kibana_synthetics_private_location`. Each covered Kibana entity SHALL expose `kibana_connection` using the shared provider schema helper for its implementation style and SHALL use the effective scoped client derived from that block when it is configured.

#### Scenario: Covered Kibana entity exposes and honors `kibana_connection`
- **WHEN** a covered `elasticstack_kibana_*` entity is configured with `kibana_connection`
- **THEN** the entity SHALL execute its API operations against the scoped client derived from that block

### Requirement: In-scope Fleet entity rollout
The rollout's covered Fleet entities SHALL be exactly the Fleet resources and data sources registered in `provider/plugin_framework.go`: `elasticstack_fleet_agent_policy`, `elasticstack_fleet_elastic_defend_integration_policy`, `elasticstack_fleet_enrollment_tokens`, `elasticstack_fleet_integration`, `elasticstack_fleet_integration_policy`, `elasticstack_fleet_output`, and `elasticstack_fleet_server_host`. Each covered Fleet entity SHALL expose `kibana_connection` using the shared provider schema helper for its implementation style and SHALL use the effective scoped client derived from that block when it is configured.

#### Scenario: Covered Fleet entity exposes and honors `kibana_connection`
- **WHEN** a covered `elasticstack_fleet_*` entity is configured with `kibana_connection`
- **THEN** the entity SHALL execute its API operations against the scoped client derived from that block

### Requirement: Provider-client fallback
Covered Kibana and Fleet entities SHALL continue to use the provider-configured client when `kibana_connection` is not configured. The presence of the new block SHALL therefore add an optional override path rather than changing default client-resolution behavior.

#### Scenario: Covered entity uses provider client by default
- **WHEN** a covered Kibana or Fleet entity is configured without `kibana_connection`
- **THEN** the entity SHALL use the provider-configured client for its API operations

### Requirement: Shared schema consistency
Covered Kibana and Fleet entities SHALL use the shared SDK or Plugin Framework `kibana_connection` schema helper rather than entity-specific block variants.

#### Scenario: Covered entity uses the shared block definition
- **WHEN** a covered Kibana or Fleet entity defines `kibana_connection`
- **THEN** that definition SHALL come from the shared provider schema helper for the entity's implementation style

