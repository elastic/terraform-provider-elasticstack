## ADDED Requirements

### Requirement: Factory validates endpoint presence as a precondition
`ProviderClientFactory.GetElasticsearchClient` and `ProviderClientFactory.GetKibanaClient` SHALL validate that at least one effective endpoint is configured for the requested component before returning a scoped client. On success, the returned scoped client's accessors SHALL be safe to call unconditionally and SHALL return a non-nil, ready-to-use typed client. On failure, the factory SHALL return error diagnostics naming the configuration paths the user can set, and SHALL NOT return a partially-configured scoped client.

#### Scenario: Elasticsearch precondition fails when no ES endpoint is configured
- **GIVEN** provider configuration, `elasticsearch_connection`, and environment overrides that together produce no non-empty Elasticsearch endpoint value
- **WHEN** `GetElasticsearchClient` is called
- **THEN** the factory SHALL return an error diagnostic instructing the user to set `elasticsearch.endpoints`, `elasticsearch_connection.endpoints`, or `ELASTICSEARCH_ENDPOINTS`
- **AND** the factory SHALL NOT return a scoped client

#### Scenario: Kibana precondition fails only when both Kibana and Fleet endpoints are missing
- **GIVEN** provider configuration, `kibana_connection`, and environment overrides that together produce no non-empty Kibana endpoint value **and** no non-empty Fleet endpoint value
- **WHEN** `GetKibanaClient` is called
- **THEN** the factory SHALL return an error diagnostic instructing the user to set one of `kibana.endpoints`, `kibana_connection.endpoints`, `KIBANA_ENDPOINT`, `fleet.endpoint`, or `FLEET_ENDPOINT`
- **AND** the factory SHALL NOT return a scoped client

#### Scenario: Kibana precondition is satisfied by either Kibana or Fleet endpoint
- **GIVEN** provider configuration where exactly one of the Kibana or Fleet endpoint values is non-empty after all overlays
- **WHEN** `GetKibanaClient` is called
- **THEN** the factory SHALL return a `*KibanaScopedClient` whose `GetKibanaOapiClient()` and `GetFleetClient()` both return non-nil clients

#### Scenario: Successful factory call guarantees usable accessors
- **GIVEN** a `*ElasticsearchScopedClient` or `*KibanaScopedClient` returned by a successful factory call
- **WHEN** the consumer calls the scoped client's typed-client accessor (`GetESClient`, `GetKibanaOapiClient`, or `GetFleetClient`)
- **THEN** the accessor SHALL return a non-nil typed client without diagnostics
