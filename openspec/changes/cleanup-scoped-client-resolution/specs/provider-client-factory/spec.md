## MODIFIED Requirements

### Requirement: Provider injects a client factory
The provider SHALL inject a `*clients.ProviderClientFactory` into Plugin Framework `ProviderData` and SDK `meta` as the provider-scoped client-resolution surface for resources and data sources. Covered consumers SHALL resolve typed scoped clients through factory methods rather than converting provider data or meta back into a broad `*clients.APIClient`.

#### Scenario: Framework configure receives a factory
- **WHEN** the Plugin Framework provider configures a resource or data source
- **THEN** the configured provider data SHALL be a `*clients.ProviderClientFactory` rather than a ready-to-use broad `*clients.APIClient`

#### Scenario: Framework consumer resolves typed client from factory
- **WHEN** a covered Framework resource or data source needs Elasticsearch- or Kibana-derived operations
- **THEN** it SHALL obtain a typed scoped client from `*clients.ProviderClientFactory` instead of converting provider data into a broad `*clients.APIClient`

#### Scenario: SDK configure receives a factory
- **WHEN** the SDK provider configures a resource or data source
- **THEN** the configured `meta` value SHALL be a `*clients.ProviderClientFactory` rather than a ready-to-use broad `*clients.APIClient`

### Requirement: Factory supports phased migration
In this phase, the `ProviderClientFactory` SHALL provide typed Kibana/Fleet scoped-client resolution and typed Elasticsearch scoped-client resolution, and SHALL NOT expose a supported bridge back to the broad `*clients.APIClient`.

#### Scenario: Kibana entity resolves typed client
- **WHEN** a Kibana, Fleet, or other covered Kibana-derived entity resolves its effective client through the factory
- **THEN** the factory SHALL return a typed Kibana-scoped client for Kibana, Kibana OpenAPI, SLO, and Fleet operations

#### Scenario: Elasticsearch entity resolves typed client
- **WHEN** a covered Elasticsearch entity resolves its effective client through the factory
- **THEN** the factory SHALL return a typed Elasticsearch-scoped client rather than a broad `*clients.APIClient`

#### Scenario: Broad-client bridge is not part of the factory contract
- **WHEN** implementation code attempts to rely on the factory for a supported broad-client escape hatch
- **THEN** the provider factory contract SHALL not define that path
