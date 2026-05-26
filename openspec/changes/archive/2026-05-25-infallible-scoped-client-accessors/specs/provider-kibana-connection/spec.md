## ADDED Requirements

### Requirement: Provider-level Kibana ↔ Fleet field inheritance is bidirectional
At the provider-level config-building path (when no resource-level `kibana_connection` is in play), the `kibana_oapi` config SHALL inherit unset fields field-by-field from the `fleet { ... }` provider block. This mirrors the existing field-level inheritance where the Fleet config inherits unset fields from the `kibana { ... }` block. The inheritance SHALL apply to URL, username, password, API key, bearer token, CA certificates, and the insecure flag.

#### Scenario: Provider with only fleet block serves Kibana resources
- **GIVEN** a provider configuration with a `fleet { endpoint = "https://kibana.example.com" }` block and no `kibana { ... }` block
- **AND** a covered Kibana resource that resolves its client through `ProviderClientFactory.GetKibanaClient` without a resource-level `kibana_connection`
- **WHEN** the resource invokes `GetKibanaOapiClient()` on the returned scoped client
- **THEN** the scoped client's Kibana OpenAPI client SHALL target the Fleet block's endpoint
- **AND** authentication fields configured only in the `fleet { ... }` block SHALL be used for Kibana OpenAPI requests

#### Scenario: Provider with both blocks uses kibana fields with fleet fallback
- **GIVEN** a provider configuration with `kibana { endpoints = ["K"] }` and `fleet { api_key = "F-KEY" }` blocks
- **AND** no `kibana_connection` resource-level override
- **WHEN** the kibana_oapi config is built
- **THEN** the kibana_oapi config SHALL use `K` as the URL
- **AND** SHALL use `F-KEY` as the API key (inherited from the fleet block because the kibana block does not set it)

#### Scenario: Environment overrides apply after fleet fallback
- **GIVEN** a provider configuration where the fleet-block fallback would supply a Kibana URL
- **AND** the `KIBANA_ENDPOINT` environment variable is set
- **WHEN** the kibana_oapi config is built
- **THEN** the existing environment-override rules (including `TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT` semantics) SHALL apply on top of the resolved value
- **AND** the env-derived value SHALL win over the fleet-block fallback unless the prefer-configured override is set with a configured kibana endpoint already present

### Requirement: Resource-level kibana_connection retains unified-override semantics
A resource-level `kibana_connection` block SHALL continue to supply both the kibana_oapi config and the fleet config in the resulting scoped client. The provider-level Fleet → Kibana inheritance step SHALL NOT apply on the resource-level override path; the resource-level override is the sole source of truth for both kibana_oapi and fleet config built for that scoped client.

#### Scenario: Resource override is the sole source of truth
- **GIVEN** a covered resource that supplies a `kibana_connection { endpoints = ["https://override.example.com"] }` block
- **AND** the provider also has both `kibana { ... }` and `fleet { ... }` provider-level blocks
- **WHEN** the factory builds the scoped client for that resource
- **THEN** both the scoped client's kibana_oapi config and its fleet config SHALL be derived from the resource-level `kibana_connection`
- **AND** the provider-level `kibana { ... }` and `fleet { ... }` blocks SHALL NOT contribute fields to the scoped client
