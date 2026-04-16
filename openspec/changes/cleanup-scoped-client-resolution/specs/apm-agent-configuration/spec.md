## MODIFIED Requirements

### Requirement: Kibana client usage (REQ-005)
The resource SHALL obtain its Kibana OpenAPI client through typed scoped-client resolution from the provider-configured `*clients.ProviderClientFactory`. The resource SHALL use `GetKibanaOapiClient()` on the resolved `*clients.KibanaScopedClient` for all API operations. The resource SHALL use the Elastic API version `2023-10-31` in all API requests.

#### Scenario: Resource resolves typed Kibana client from provider defaults
- **WHEN** the resource is configured without an entity-local Kibana override
- **THEN** it SHALL resolve a `*clients.KibanaScopedClient` from the provider client factory and use that typed client for Kibana API operations

#### Scenario: Kibana client acquisition failure
- **WHEN** the provider cannot provide a typed Kibana client or Kibana OpenAPI client
- **THEN** Terraform diagnostics SHALL include an "Unable to get Kibana client" error
