## MODIFIED Requirements

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
