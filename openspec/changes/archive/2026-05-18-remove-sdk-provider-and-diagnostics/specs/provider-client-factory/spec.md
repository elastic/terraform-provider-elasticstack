## REMOVED Requirements

### Requirement: SDK configure receives a factory
**Reason**: The SDK provider is removed entirely; there is no SDK configure path.
**Migration**: All resources and data sources already use Plugin Framework. No consumer migration is required.

#### Scenario: SDK meta injection is absent
- **GIVEN** the codebase after this change
- **WHEN** searching for references to SDK `meta` injection into `*schema.ResourceData`
- **THEN** there SHALL be none in `internal/clients/provider_client_factory.go`
- **AND** `ConvertMetaToFactory` SHALL be removed

### Requirement: Factory supports phased migration — SDK Elasticsearch resolution
**Reason**: The SDK provider and all SDK resources are removed; transitional legacy resolution is no longer needed.
**Migration**: All Elasticsearch entities already use Plugin Framework and call `GetElasticsearchClient`.

#### Scenario: SDK Elasticsearch client resolution is absent
- **GIVEN** the codebase after this change
- **WHEN** searching for `GetElasticsearchClientFromSDK`
- **THEN** the function SHALL NOT exist

#### Scenario: SDK Kibana client resolution is absent
- **GIVEN** the codebase after this change
- **WHEN** searching for `GetKibanaClientFromSDK`
- **THEN** the function SHALL NOT exist

## MODIFIED Requirements

### Requirement: Provider injects a client factory
The provider SHALL inject a `*clients.ProviderClientFactory` into Plugin Framework `ProviderData` and `ResourceData` as the provider-scoped client-resolution surface for resources and data sources. Covered consumers SHALL resolve typed scoped clients through factory methods rather than converting provider data or meta back into a broad `*clients.APIClient`.

#### Scenario: Framework configure receives a factory
- **WHEN** the Plugin Framework provider configures a resource or data source
- **THEN** the configured provider data SHALL be a `*clients.ProviderClientFactory` rather than a ready-to-use broad `*clients.APIClient`

#### Scenario: Framework consumer resolves typed client from factory
- **WHEN** a covered Framework resource or data source needs Elasticsearch- or Kibana-derived operations
- **THEN** it SHALL obtain a typed scoped client from `*clients.ProviderClientFactory` instead of converting provider data into a broad `*clients.APIClient`

### Requirement: Factory supports phased migration
During the Kibana/Fleet typed-client phase, the `*clients.ProviderClientFactory` SHALL provide typed Kibana/Fleet scoped-client resolution and SHALL also preserve explicit legacy Elasticsearch resolution methods so unconverted Elasticsearch entities continue to behave as they did before the factory migration.

#### Scenario: Kibana entity resolves typed client
- **WHEN** a Kibana or Fleet entity resolves its effective client through the factory
- **THEN** the factory SHALL return a typed Kibana-scoped client whose surfaces include the Kibana OpenAPI client, SLO client, and Fleet client for their respective operations

#### Scenario: Elasticsearch entity uses transitional legacy resolution
- **WHEN** an unconverted Elasticsearch entity resolves its effective client during this phase
- **THEN** the factory SHALL expose a transitional resolution path that preserves the existing broad-client and lint-enforced Elasticsearch behavior
