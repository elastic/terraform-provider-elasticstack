# provider-client-factory Specification

## Purpose
TBD - created by archiving change typed-kibana-fleet-client-resolution. Update Purpose after archive.
## Requirements
### Requirement: Provider injects a client factory
The provider SHALL inject a `*clients.ProviderClientFactory` into Plugin Framework `ProviderData` and SDK `meta` as the provider-scoped client-resolution surface for resources and data sources.

#### Scenario: Framework configure receives a factory
- **WHEN** the Plugin Framework provider configures a resource or data source
- **THEN** the configured provider data SHALL be a `*clients.ProviderClientFactory` rather than a ready-to-use broad `*clients.APIClient`

#### Scenario: SDK configure receives a factory
- **WHEN** the SDK provider configures a resource or data source
- **THEN** the configured `meta` value SHALL be a `*clients.ProviderClientFactory` rather than a ready-to-use broad `*clients.APIClient`

### Requirement: Factory supports phased migration
In this phase, the `ProviderClientFactory` SHALL provide typed Kibana/Fleet scoped-client resolution and typed Elasticsearch scoped-client resolution, and SHALL NOT expose transitional legacy Elasticsearch resolution methods that return the broad `*clients.APIClient`.

#### Scenario: Kibana entity resolves typed client
- **WHEN** a Kibana or Fleet entity resolves its effective client through the factory
- **THEN** the factory SHALL return a typed Kibana-scoped client for Kibana, Kibana OpenAPI, SLO, and Fleet operations

#### Scenario: Elasticsearch entity resolves typed client
- **WHEN** a covered Elasticsearch entity resolves its effective client through the factory
- **THEN** the factory SHALL return a typed Elasticsearch-scoped client rather than a broad `*clients.APIClient`

#### Scenario: Transitional broad Elasticsearch resolution is removed
- **WHEN** implementation code attempts to rely on the factory for legacy broad Elasticsearch resolution in this phase
- **THEN** the provider factory contract SHALL not define that transitional path

### Requirement: Kibana scoped client contract
The typed Kibana-scoped client returned by the factory SHALL expose the Kibana legacy client, Kibana OpenAPI client, SLO client, Fleet client, Kibana auth-context helpers, and Kibana-derived version and flavor checks required by covered Kibana and Fleet entities.

#### Scenario: Scoped client supports Kibana and Fleet operations
- **WHEN** a covered Kibana or Fleet entity uses a typed Kibana-scoped client
- **THEN** the client SHALL provide the typed client surfaces needed for Kibana, Kibana OpenAPI, SLO, and Fleet API operations

#### Scenario: Scoped client supports version gating
- **WHEN** a covered Kibana or Fleet entity performs version or flavor checks through the typed Kibana-scoped client
- **THEN** the client SHALL provide `ServerVersion()`, `ServerFlavor()`, or equivalent typed behavior needed for those checks

