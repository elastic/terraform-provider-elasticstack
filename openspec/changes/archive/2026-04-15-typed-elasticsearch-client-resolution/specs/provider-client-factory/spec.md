## MODIFIED Requirements

### Requirement: Factory supports phased migration
In this phase, the `ProviderClientFactory` SHALL provide typed Kibana/Fleet scoped-client resolution and typed Elasticsearch scoped-client resolution, and SHALL NOT expose transitional legacy Elasticsearch resolution methods that return the broad `*clients.APIClient`.

#### Scenario: Elasticsearch entity resolves typed client
- **WHEN** a covered Elasticsearch entity resolves its effective client through the factory
- **THEN** the factory SHALL return a typed Elasticsearch-scoped client rather than a broad `*clients.APIClient`

#### Scenario: Transitional broad Elasticsearch resolution is removed
- **WHEN** implementation code attempts to rely on the factory for legacy broad Elasticsearch resolution in this phase
- **THEN** the provider factory contract SHALL not define that transitional path
