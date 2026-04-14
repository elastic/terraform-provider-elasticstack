## MODIFIED Requirements

### Requirement: Provider configuration and Kibana client

On create, read, update, and delete, the resource SHALL obtain the Kibana OpenAPI client from the provider by default. If the client cannot be obtained, the resource SHALL return an error diagnostic and SHALL not proceed to the API. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for all operations.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the provider-configured Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the scoped Kibana OpenAPI client derived from that block
