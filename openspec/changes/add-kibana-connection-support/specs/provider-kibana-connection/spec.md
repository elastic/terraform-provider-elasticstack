## ADDED Requirements

### Requirement: Entity-local `kibana_connection` schema source of truth
The provider SHALL define shared SDK and Plugin Framework schema helpers for entity-local `kibana_connection`. Those helpers SHALL use the same field set and validation rules as the provider-level Kibana connection block, SHALL be list-shaped with at most one element, and SHALL NOT expose entity-level deprecation metadata.

#### Scenario: Shared helpers define the entity-local block shape
- **WHEN** an entity-local `kibana_connection` schema or block is defined
- **THEN** it SHALL come from the shared provider schema helpers rather than an entity-specific variant

### Requirement: Framework scoped Kibana client resolution
The provider SHALL expose a Plugin Framework helper that accepts an entity-local `kibana_connection` block and a default `*clients.APIClient`. When the block is not configured, the helper SHALL return the default client. When the block is configured, the helper SHALL return a scoped `*clients.APIClient` whose Kibana legacy client, Kibana OpenAPI client, SLO client, and Fleet client are rebuilt from the scoped `kibana_connection`.

#### Scenario: Framework helper falls back to provider client
- **WHEN** a Framework entity resolves its effective client and `kibana_connection` is absent
- **THEN** the helper SHALL return the provider-configured default `*clients.APIClient`

#### Scenario: Framework helper builds a scoped Kibana-derived client
- **WHEN** a Framework entity resolves its effective client and `kibana_connection` is configured
- **THEN** the helper SHALL return a scoped `*clients.APIClient` rebuilt from that connection for Kibana, SLO, and Fleet operations

### Requirement: SDK scoped Kibana client resolution
The provider SHALL expose an SDK helper that accepts resource or data source state and provider meta and resolves an effective `*clients.APIClient` from entity-local `kibana_connection`. When the block is not configured, the helper SHALL use the provider client. When the block is configured, the helper SHALL build a scoped `*clients.APIClient` whose Kibana legacy client, Kibana OpenAPI client, SLO client, and Fleet client are rebuilt from the scoped `kibana_connection`.

#### Scenario: SDK helper falls back to provider client
- **WHEN** an SDK entity resolves its effective client and `kibana_connection` is absent
- **THEN** the helper SHALL return the provider-configured default `*clients.APIClient`

#### Scenario: SDK helper builds a scoped Kibana-derived client
- **WHEN** an SDK entity resolves its effective client and `kibana_connection` is configured
- **THEN** the helper SHALL return a scoped `*clients.APIClient` rebuilt from that connection for Kibana, SLO, and Fleet operations

### Requirement: Scoped client version and identity behavior
When an entity uses a scoped `kibana_connection`, version, flavor, and other Kibana-derived client checks SHALL resolve against the scoped connection rather than the provider-level Elasticsearch client. The scoped `*clients.APIClient` SHALL therefore avoid reusing provider-level Elasticsearch identity in a way that can make Kibana or Fleet operations target one cluster while version or identity checks target another.

#### Scenario: Scoped version checks follow the scoped Kibana connection
- **WHEN** an entity uses `ServerVersion()`, `ServerFlavor()`, or equivalent behavior through a scoped `kibana_connection`
- **THEN** the result SHALL be derived from the scoped Kibana connection instead of the provider's Elasticsearch connection
