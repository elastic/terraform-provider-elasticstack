## MODIFIED Requirements

### Requirement: Framework scoped Kibana client resolution
The provider SHALL expose Plugin Framework `kibana_connection` resolution through `*clients.ProviderClientFactory` methods that accept an entity-local `kibana_connection` block and return a `*clients.KibanaScopedClient`. When the block is not configured, the factory SHALL return a `*clients.KibanaScopedClient` built from provider-level defaults. When the block is configured, the factory SHALL return a `*clients.KibanaScopedClient` whose Kibana legacy client, Kibana OpenAPI client, SLO client, and Fleet client are rebuilt from the scoped `kibana_connection`.

#### Scenario: Framework factory falls back to provider defaults
- **WHEN** a Framework entity resolves its effective Kibana client through the factory and `kibana_connection` is absent
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` derived from provider configuration

#### Scenario: Framework factory builds a scoped Kibana-derived client
- **WHEN** a Framework entity resolves its effective Kibana client through the factory and `kibana_connection` is configured
- **THEN** the factory SHALL return a `*clients.KibanaScopedClient` rebuilt from that connection for Kibana, SLO, and Fleet operations

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
