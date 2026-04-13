## MODIFIED Requirements

### Requirement: Provider-level Kibana client only (REQ-005)

The resource SHALL use the provider's configured Kibana OpenAPI client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for create, read, update, and delete.

#### Scenario: Standard provider connection

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** all data view API operations SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all data view API operations SHALL use the scoped Kibana OpenAPI client derived from that block
