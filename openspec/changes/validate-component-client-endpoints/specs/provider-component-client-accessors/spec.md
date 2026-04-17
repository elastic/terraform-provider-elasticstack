## ADDED Requirements

### Requirement: Elasticsearch scoped accessor requires an effective endpoint
`(*clients.ElasticsearchScopedClient).GetESClient()` SHALL validate that an effective Elasticsearch endpoint is configured before returning a client. If provider configuration, `elasticsearch_connection`, and environment overrides together produce no non-empty Elasticsearch endpoint values for the scoped client instance, the accessor SHALL return an error instead of returning a usable client.

#### Scenario: Missing Elasticsearch endpoint returns actionable error
- **GIVEN** an Elasticsearch-scoped client whose effective Elasticsearch endpoint configuration contains no non-empty endpoint values
- **WHEN** `GetESClient()` is called
- **THEN** the accessor SHALL return no client
- **AND** it SHALL return the error `elasticsearch client is not configured: set elasticsearch.endpoints, elasticsearch_connection.endpoints, or ELASTICSEARCH_ENDPOINTS`

### Requirement: Kibana scoped accessors require an effective Kibana endpoint
`(*clients.KibanaScopedClient).GetKibanaOapiClient()` SHALL validate that an effective Kibana endpoint is configured before returning a client. If provider configuration, `kibana_connection`, and environment overrides together produce no non-empty Kibana endpoint values for the scoped client instance, the accessor SHALL return an error instead of returning a usable client.

#### Scenario: Missing Kibana endpoint returns actionable error for the Kibana OpenAPI client
- **GIVEN** a Kibana-scoped client whose effective Kibana endpoint configuration contains no non-empty endpoint values
- **WHEN** `GetKibanaOapiClient()` is called
- **THEN** the accessor SHALL return no client
- **AND** it SHALL return the error `kibana OpenAPI client is not configured: set kibana.endpoints, kibana_connection.endpoints, or KIBANA_ENDPOINT`

### Requirement: Fleet scoped accessor requires an effective Fleet endpoint
`(*clients.KibanaScopedClient).GetFleetClient()` SHALL validate that an effective Fleet endpoint is configured before returning a client. The effective Fleet endpoint MAY come from explicit provider-level Fleet configuration, from the existing Fleet-from-Kibana provider resolution path, or from the `kibana_connection`-derived Fleet config used for scoped Kibana clients. If none of those resolution paths produces any non-empty Fleet endpoint value, the accessor SHALL return an error instead of returning a usable client.

#### Scenario: Fleet accessor accepts Kibana-derived endpoint resolution
- **GIVEN** a Kibana-scoped client whose explicit Fleet endpoint is empty
- **AND** whose effective Kibana endpoint configuration contains at least one non-empty endpoint value and is the source of the Fleet endpoint
- **WHEN** `GetFleetClient()` is called
- **THEN** the accessor SHALL return the Fleet client without a missing-endpoint error

#### Scenario: Missing Fleet endpoint returns actionable error
- **GIVEN** a Kibana-scoped client whose effective Fleet endpoint configuration contains no non-empty endpoint values after explicit Fleet and Kibana-derived resolution are evaluated
- **WHEN** `GetFleetClient()` is called
- **THEN** the accessor SHALL return no client
- **AND** it SHALL return the error `fleet client is not configured: set fleet.endpoint or FLEET_ENDPOINT, or configure kibana.endpoints, kibana_connection.endpoints, or KIBANA_ENDPOINT for inherited Fleet endpoint resolution`

### Requirement: Endpoint validation is limited to endpoint presence
Component accessor validation introduced by this capability SHALL enforce endpoint presence only. Accessors SHALL NOT reject a client solely because username/password, API key, or bearer token values are absent.

#### Scenario: Endpoint-only validation does not add authentication gating
- **GIVEN** a typed scoped client whose effective component endpoint is present
- **AND** whose authentication fields are empty
- **WHEN** the corresponding component accessor is called
- **THEN** the accessor SHALL NOT fail solely because authentication fields are empty

