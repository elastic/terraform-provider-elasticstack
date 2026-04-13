## MODIFIED Requirements

### Requirement: Provider-level Kibana client only (REQ-005)

The resource SHALL use the provider's configured Kibana clients by default (OpenAPI client for create, read, and update; legacy client for delete). When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI and legacy clients for the corresponding operations.

#### Scenario: Standard provider connection

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** all parameter API operations SHALL use the provider-level Kibana clients

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all parameter API operations SHALL use the scoped Kibana clients derived from that block
