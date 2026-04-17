## MODIFIED Requirements

### Requirement: Kibana scoped client contract

The typed Kibana-scoped client returned by the factory SHALL expose the Kibana OpenAPI client, SLO client, Fleet client, Kibana auth-context helpers, and Kibana-derived version and flavor checks required by covered Kibana and Fleet entities. The factory contract SHALL use the Kibana OpenAPI configuration surface as the only Kibana connection contract for provider-level and scoped `kibana_connection` resolution and SHALL NOT expose or require `github.com/disaster37/go-kibana-rest` as part of provider wiring.

#### Scenario: Scoped client supports Kibana and Fleet operations
- **WHEN** a covered Kibana or Fleet entity uses a typed Kibana-scoped client
- **THEN** the client SHALL provide the typed client surfaces needed for Kibana HTTP workloads through the OpenAPI client, plus SLO and Fleet API operations as applicable

#### Scenario: Scoped client supports version gating
- **WHEN** a covered Kibana or Fleet entity performs version or flavor checks through the typed Kibana-scoped client
- **THEN** the client SHALL provide `ServerVersion()`, `ServerFlavor()`, or equivalent typed behavior needed for those checks

#### Scenario: Factory does not require a legacy Kibana config surface
- **WHEN** the provider client factory resolves a Kibana-scoped client from provider configuration or `kibana_connection`
- **THEN** it SHALL validate and build that client from the Kibana OpenAPI config surface without relying on a parallel legacy Kibana REST config object
