## MODIFIED Requirements

### Requirement: Connection (REQ-003)

The data source SHALL use the provider's configured Kibana client (`KibanaSpaces` API client) by default. When `kibana_connection` is configured on the data source, the data source SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for the spaces API call.

#### Scenario: Provider-level Kibana client

- **WHEN** `kibana_connection` is not configured on the data source
- **THEN** the provider-level Kibana client SHALL be used

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the data source
- **THEN** the scoped Kibana client derived from that block SHALL be used
