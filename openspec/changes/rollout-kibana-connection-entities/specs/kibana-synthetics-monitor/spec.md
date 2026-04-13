## MODIFIED Requirements

### Requirement: Provider configuration and Kibana client (REQ-003)

On create, read, update, and delete, if the provider did not supply a usable API client, the resource SHALL return an "Unconfigured Client" error diagnostic and SHALL NOT proceed to the Kibana API. The resource SHALL use the provider-configured Kibana client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for all operations.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an "Unconfigured Client" error diagnostic

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the provider-configured Kibana client

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the scoped Kibana client derived from that block
