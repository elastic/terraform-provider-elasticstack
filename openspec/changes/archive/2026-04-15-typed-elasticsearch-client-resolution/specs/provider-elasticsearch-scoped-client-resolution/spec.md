## ADDED Requirements

### Requirement: Framework Elasticsearch scoped client resolution
The provider SHALL expose Plugin Framework `elasticsearch_connection` resolution through `ProviderClientFactory` methods that accept an entity-local `elasticsearch_connection` block and return a typed Elasticsearch-scoped client. When the block is not configured, the factory SHALL return a typed client built from provider-level defaults. When the block is configured, the factory SHALL return a typed scoped client rebuilt from that connection for Elasticsearch operations.

#### Scenario: Framework factory falls back to provider defaults
- **WHEN** a Framework Elasticsearch entity resolves its effective client through the factory and `elasticsearch_connection` is absent
- **THEN** the factory SHALL return a typed Elasticsearch-scoped client derived from provider configuration

#### Scenario: Framework factory builds a scoped Elasticsearch client
- **WHEN** a Framework Elasticsearch entity resolves its effective client through the factory and `elasticsearch_connection` is configured
- **THEN** the factory SHALL return a typed Elasticsearch-scoped client rebuilt from that connection for Elasticsearch operations

### Requirement: SDK Elasticsearch scoped client resolution
The provider SHALL expose SDK `elasticsearch_connection` resolution through `ProviderClientFactory` methods that accept resource or data source state and return a typed Elasticsearch-scoped client. When the block is not configured, the factory SHALL return a typed client built from provider-level defaults. When the block is configured, the factory SHALL return a typed scoped client rebuilt from that connection for Elasticsearch operations.

#### Scenario: SDK factory falls back to provider defaults
- **WHEN** an SDK Elasticsearch entity resolves its effective client through the factory and `elasticsearch_connection` is absent
- **THEN** the factory SHALL return a typed Elasticsearch-scoped client derived from provider configuration

#### Scenario: SDK factory builds a scoped Elasticsearch client
- **WHEN** an SDK Elasticsearch entity resolves its effective client through the factory and `elasticsearch_connection` is configured
- **THEN** the factory SHALL return a typed Elasticsearch-scoped client rebuilt from that connection for Elasticsearch operations

### Requirement: Elasticsearch sink type safety
In-scope shared Elasticsearch helpers and sinks SHALL accept the typed Elasticsearch-scoped client, or narrower interfaces derived from it, instead of accepting the broad `*clients.APIClient`.

#### Scenario: Shared sink requires typed client
- **WHEN** code under `internal/clients/elasticsearch/**` exposes a sink or helper that consumes provider-resolved Elasticsearch client state
- **THEN** that sink or helper SHALL require the typed Elasticsearch-scoped client contract rather than `*clients.APIClient`

### Requirement: Elasticsearch scoped client helper behavior
The typed Elasticsearch-scoped client SHALL expose the Elasticsearch client surface and the Elasticsearch-derived helper behavior needed by covered Elasticsearch entities, including composite ID generation, cluster identity lookup, version checks, flavor checks, and minimum-version enforcement.

#### Scenario: Scoped client supports Elasticsearch helper behavior
- **WHEN** a covered Elasticsearch entity performs ID generation, cluster identity lookup, version checks, flavor checks, or minimum-version enforcement through the typed scoped client
- **THEN** the typed scoped client SHALL provide that behavior without requiring access to a broad `*clients.APIClient`
