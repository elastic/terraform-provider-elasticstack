## MODIFIED Requirements

### Requirement: Connection (REQ-007)

The resource and data source SHALL use the provider's configured Kibana client by default for all API calls. When `kibana_connection` is configured on the resource or data source, the entity SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for the API call.

#### Scenario: Provider-level Kibana client

- **WHEN** `kibana_connection` is not configured on the resource or data source
- **THEN** the provider-level Kibana client SHALL be used

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource or data source
- **THEN** the scoped Kibana client derived from that block SHALL be used
