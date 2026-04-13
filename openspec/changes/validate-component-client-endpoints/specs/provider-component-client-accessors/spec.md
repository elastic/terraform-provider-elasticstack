## ADDED Requirements

### Requirement: Elasticsearch accessor requires an effective endpoint
`(*clients.APIClient).GetESClient()` SHALL validate that an effective Elasticsearch endpoint is configured before returning a client. If neither provider configuration nor environment overrides produce an Elasticsearch endpoint, the accessor SHALL return an error instead of returning a usable client.

#### Scenario: Missing Elasticsearch endpoint returns actionable error
- **GIVEN** a provider client whose effective Elasticsearch endpoint configuration is empty
- **WHEN** `GetESClient()` is called
- **THEN** the accessor SHALL return no client
- **AND** it SHALL return the error `provider Elasticsearch client is not configured: set elasticsearch.endpoints or ELASTICSEARCH_ENDPOINTS`

### Requirement: Kibana-family accessors require an effective Kibana endpoint
`(*clients.APIClient).GetKibanaClient()`, `(*clients.APIClient).GetKibanaOapiClient()`, and `(*clients.APIClient).GetSloClient()` SHALL validate that an effective Kibana endpoint is configured before returning a client. If neither provider configuration nor environment overrides produce a Kibana endpoint, each accessor SHALL return an error instead of returning a usable client.

#### Scenario: Missing Kibana endpoint returns actionable error for the legacy Kibana client
- **GIVEN** a provider client whose effective Kibana endpoint configuration is empty
- **WHEN** `GetKibanaClient()` is called
- **THEN** the accessor SHALL return no client
- **AND** it SHALL return the error `provider Kibana client is not configured: set kibana.endpoints or KIBANA_ENDPOINT`

#### Scenario: Missing Kibana endpoint returns actionable error for the Kibana OpenAPI client
- **GIVEN** a provider client whose effective Kibana endpoint configuration is empty
- **WHEN** `GetKibanaOapiClient()` is called
- **THEN** the accessor SHALL return no client
- **AND** it SHALL return the error `provider Kibana OpenAPI client is not configured: set kibana.endpoints or KIBANA_ENDPOINT`

#### Scenario: Missing Kibana endpoint returns actionable error for the SLO client
- **GIVEN** a provider client whose effective Kibana endpoint configuration is empty
- **WHEN** `GetSloClient()` is called
- **THEN** the accessor SHALL return no client
- **AND** it SHALL return the error `provider SLO client is not configured: set kibana.endpoints or KIBANA_ENDPOINT`

#### Scenario: Legacy Kibana accessor does not fall back to localhost when unconfigured
- **GIVEN** a provider client whose effective Kibana endpoint configuration is empty
- **WHEN** `GetKibanaClient()` is called
- **THEN** the accessor SHALL fail before a request can target a default localhost endpoint

### Requirement: Fleet accessor requires an effective Fleet endpoint
`(*clients.APIClient).GetFleetClient()` SHALL validate that an effective Fleet endpoint is configured before returning a client. The effective Fleet endpoint MAY come from explicit Fleet configuration or from the existing Fleet-from-Kibana endpoint resolution path. If neither resolution path produces a Fleet endpoint, the accessor SHALL return an error instead of returning a usable client.

#### Scenario: Fleet accessor accepts Kibana-derived endpoint resolution
- **GIVEN** a provider client whose explicit Fleet endpoint is empty
- **AND** whose effective Kibana endpoint configuration is present and is the source of the Fleet endpoint
- **WHEN** `GetFleetClient()` is called
- **THEN** the accessor SHALL return the Fleet client without a missing-endpoint error

#### Scenario: Missing Fleet endpoint returns actionable error
- **GIVEN** a provider client whose effective Fleet endpoint configuration is empty after explicit Fleet and Kibana-derived resolution are evaluated
- **WHEN** `GetFleetClient()` is called
- **THEN** the accessor SHALL return no client
- **AND** it SHALL return the error `provider Fleet client is not configured: set fleet.endpoint or FLEET_ENDPOINT, or configure kibana.endpoints or KIBANA_ENDPOINT for inherited Fleet endpoint resolution`

### Requirement: Endpoint validation is limited to endpoint presence
Component accessor validation introduced by this capability SHALL enforce endpoint presence only. Accessors SHALL NOT reject a client solely because username/password, API key, or bearer token values are absent.

#### Scenario: Endpoint-only validation does not add authentication gating
- **GIVEN** a provider client whose effective component endpoint is present
- **AND** whose authentication fields are empty
- **WHEN** the corresponding component accessor is called
- **THEN** the accessor SHALL NOT fail solely because authentication fields are empty

