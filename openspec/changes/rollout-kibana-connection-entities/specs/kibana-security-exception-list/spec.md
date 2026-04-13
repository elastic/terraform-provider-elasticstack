## MODIFIED Requirements

### Requirement: Provider configuration and Kibana client (REQ-006)

On create, read, update, and delete, the resource SHALL use the provider's configured Kibana OAPI HTTP client by default. If that client cannot be obtained, the resource SHALL return an error diagnostic with summary "Failed to get Kibana client" and SHALL not proceed to the API. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OAPI HTTP client for all operations.

#### Scenario: Unconfigured provider

- GIVEN the provider did not supply a usable Kibana API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic before making any API call

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the provider-configured Kibana OAPI HTTP client

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the scoped Kibana OAPI HTTP client derived from that block
