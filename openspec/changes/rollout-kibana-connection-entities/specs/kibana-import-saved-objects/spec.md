## MODIFIED Requirements

### Requirement: Provider-level Kibana client only (REQ-004)

The resource SHALL use the provider's configured Kibana legacy client by default for create and update. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana legacy client for create and update.

#### Scenario: Standard provider connection

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** all import saved objects API operations SHALL use the provider-level Kibana legacy client

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all import saved objects API operations SHALL use the scoped Kibana legacy client derived from that block
