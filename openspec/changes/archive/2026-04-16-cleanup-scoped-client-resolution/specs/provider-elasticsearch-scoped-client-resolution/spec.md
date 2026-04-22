## MODIFIED Requirements

### Requirement: Elasticsearch scoped client helper behavior
The typed Elasticsearch-scoped client SHALL expose the Elasticsearch client surface and the Elasticsearch-derived helper behavior needed by covered Elasticsearch entities, including composite ID generation, cluster identity lookup, version checks, flavor checks, and minimum-version enforcement. Those behaviors SHALL be available through `*clients.ElasticsearchScopedClient` without requiring a supported broad `*clients.APIClient` Elasticsearch helper surface.

#### Scenario: Scoped client supports Elasticsearch helper behavior
- **WHEN** a covered Elasticsearch entity performs ID generation, cluster identity lookup, version checks, flavor checks, or minimum-version enforcement through the typed scoped client
- **THEN** the typed scoped client SHALL provide that behavior without requiring access to a broad `*clients.APIClient`

#### Scenario: Broad client is not required for Elasticsearch helper behavior
- **WHEN** in-scope Elasticsearch helper behavior has been migrated to `*clients.ElasticsearchScopedClient`
- **THEN** implementation code SHALL not need a supported broad `*clients.APIClient` contract to perform those Elasticsearch-specific operations
