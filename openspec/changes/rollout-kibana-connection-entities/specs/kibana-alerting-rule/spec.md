## MODIFIED Requirements

### Requirement: Provider configuration and Kibana client (REQ-006)

On create, read, update, and delete, if the provider did not supply a usable API client for this resource, the resource SHALL return a configuration error diagnostic. The resource SHALL use the provider's configured Kibana HTTP client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana HTTP client for all operations.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with a provider configuration error

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the provider-configured Kibana HTTP client

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the scoped Kibana HTTP client derived from that block
