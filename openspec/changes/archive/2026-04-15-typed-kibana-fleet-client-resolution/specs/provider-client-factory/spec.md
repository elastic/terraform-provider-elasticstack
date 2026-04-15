## ADDED Requirements

### Requirement: Provider injects a client factory
The provider SHALL inject a `*clients.ProviderClientFactory` into Plugin Framework `ProviderData` and SDK `meta` as the provider-scoped client-resolution surface for resources and data sources.

#### Scenario: Framework configure receives a factory
- **WHEN** the Plugin Framework provider configures a resource or data source
- **THEN** the configured provider data SHALL be a `*clients.ProviderClientFactory` rather than a ready-to-use broad `*clients.APIClient`

#### Scenario: SDK configure receives a factory
- **WHEN** the SDK provider configures a resource or data source
- **THEN** the configured `meta` value SHALL be a `*clients.ProviderClientFactory` rather than a ready-to-use broad `*clients.APIClient`

### Requirement: Factory supports phased migration
During the Kibana/Fleet typed-client phase, the `*clients.ProviderClientFactory` SHALL provide typed Kibana/Fleet scoped-client resolution and SHALL also preserve explicit legacy Elasticsearch resolution methods so unconverted Elasticsearch entities continue to behave as they did before the factory migration.

#### Scenario: Kibana entity resolves typed client
- **WHEN** a Kibana or Fleet entity resolves its effective client through the factory
- **THEN** the factory SHALL return a typed Kibana-scoped client for Kibana, Kibana OpenAPI, SLO, and Fleet operations

#### Scenario: Elasticsearch entity uses transitional legacy resolution
- **WHEN** an unconverted Elasticsearch entity resolves its effective client during this phase
- **THEN** the factory SHALL expose a transitional resolution path that preserves the existing broad-client and lint-enforced Elasticsearch behavior

### Requirement: Kibana scoped client contract
The typed Kibana-scoped client returned by the factory SHALL expose the Kibana legacy client, Kibana OpenAPI client, SLO client, Fleet client, Kibana auth-context helpers, and Kibana-derived version and flavor checks required by covered Kibana and Fleet entities.

#### Scenario: Scoped client supports Kibana and Fleet operations
- **WHEN** a covered Kibana or Fleet entity uses a typed Kibana-scoped client
- **THEN** the client SHALL provide the typed client surfaces needed for Kibana, Kibana OpenAPI, SLO, and Fleet API operations

#### Scenario: Scoped client supports version gating
- **WHEN** a covered Kibana or Fleet entity performs version or flavor checks through the typed Kibana-scoped client
- **THEN** the client SHALL provide `ServerVersion()`, `ServerFlavor()`, or equivalent typed behavior needed for those checks
