## MODIFIED Requirements

### Requirement: Elasticsearch scoped accessor returns a usable client
`(*clients.ElasticsearchScopedClient).GetESClient()` SHALL return a usable `*elasticsearch.TypedClient` without diagnostics. Endpoint-presence validation SHALL be performed by `ProviderClientFactory.GetElasticsearchClient` before the scoped client is returned, so the accessor's contract is to return a non-nil, ready-to-use typed client whenever it is invoked on a scoped client produced by a successful factory call.

#### Scenario: Accessor returns the typed client without diagnostics
- **GIVEN** an `*ElasticsearchScopedClient` produced by a successful `ProviderClientFactory.GetElasticsearchClient` call
- **WHEN** `GetESClient()` is called
- **THEN** the accessor SHALL return the `*elasticsearch.TypedClient` directly with no error diagnostics return value

#### Scenario: Missing Elasticsearch endpoint surfaces at factory resolution
- **GIVEN** provider configuration, `elasticsearch_connection`, and environment overrides that together produce no non-empty Elasticsearch endpoint value
- **WHEN** `ProviderClientFactory.GetElasticsearchClient` is called
- **THEN** the factory SHALL return error diagnostics with the message `elasticsearch client is not configured: set elasticsearch.endpoints, elasticsearch_connection.endpoints, or ELASTICSEARCH_ENDPOINTS`
- **AND** no scoped client SHALL be returned

### Requirement: Kibana scoped accessors return usable clients
`(*clients.KibanaScopedClient).GetKibanaOapiClient()` and `(*clients.KibanaScopedClient).GetFleetClient()` SHALL return usable typed clients without diagnostics. Endpoint-presence validation SHALL be performed by `ProviderClientFactory.GetKibanaClient` before the scoped client is returned, so both accessors' contract is to return a non-nil, ready-to-use client whenever invoked on a scoped client produced by a successful factory call.

#### Scenario: Kibana OpenAPI accessor returns the client without diagnostics
- **GIVEN** a `*KibanaScopedClient` produced by a successful `ProviderClientFactory.GetKibanaClient` call
- **WHEN** `GetKibanaOapiClient()` is called
- **THEN** the accessor SHALL return the `*kibanaoapi.Client` directly with no error diagnostics return value

#### Scenario: Fleet accessor returns the client without diagnostics
- **GIVEN** a `*KibanaScopedClient` produced by a successful `ProviderClientFactory.GetKibanaClient` call
- **WHEN** `GetFleetClient()` is called
- **THEN** the accessor SHALL return the `*fleet.Client` directly with no error diagnostics return value

#### Scenario: Missing Kibana and Fleet endpoint surfaces at factory resolution
- **GIVEN** provider configuration, `kibana_connection`, and environment overrides that together produce no non-empty Kibana endpoint value **and** no non-empty Fleet endpoint value
- **WHEN** `ProviderClientFactory.GetKibanaClient` is called
- **THEN** the factory SHALL return error diagnostics identifying the missing endpoint and naming the configuration paths the user can set (`kibana.endpoints`, `kibana_connection.endpoints`, `KIBANA_ENDPOINT`, `fleet.endpoint`, or `FLEET_ENDPOINT`)
- **AND** no scoped client SHALL be returned

#### Scenario: Single configured endpoint side is sufficient
- **GIVEN** provider configuration where exactly one of the Kibana or Fleet endpoint values is non-empty after all overlays
- **WHEN** `ProviderClientFactory.GetKibanaClient` is called
- **THEN** the factory SHALL return a `*KibanaScopedClient` whose `GetKibanaOapiClient()` and `GetFleetClient()` both return usable clients

### Requirement: Endpoint validation is limited to endpoint presence
Component client validation SHALL enforce endpoint presence only. The factory SHALL NOT reject a configuration solely because username/password, API key, or bearer token values are absent.

#### Scenario: Endpoint-only validation does not add authentication gating
- **GIVEN** a provider configuration whose effective component endpoint is present
- **AND** whose authentication fields are empty
- **WHEN** the corresponding factory method is called
- **THEN** the factory SHALL NOT fail solely because authentication fields are empty
- **AND** the returned scoped client's accessors SHALL return usable clients
