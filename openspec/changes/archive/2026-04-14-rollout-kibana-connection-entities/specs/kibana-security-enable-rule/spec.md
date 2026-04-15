## MODIFIED Requirements

### Requirement: Provider configuration and Kibana client

On every CRUD operation, the resource SHALL use the provider's configured Kibana OAPI client by default. If the provider data cannot be converted to a valid API client, the resource SHALL return a configuration error diagnostic. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OAPI client for all operations.

#### Scenario: Unconfigured provider

- GIVEN the provider has not supplied a usable API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with a provider configuration error

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the provider-configured Kibana OAPI client

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the scoped Kibana OAPI client derived from that block
