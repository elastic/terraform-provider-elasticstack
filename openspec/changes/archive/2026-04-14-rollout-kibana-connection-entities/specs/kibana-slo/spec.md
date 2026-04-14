## MODIFIED Requirements

### Requirement: Connection — provider default Kibana client with optional scoped override (REQ-013)

The resource SHALL use the provider's configured Kibana client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for all API calls and version checks.

#### Scenario: All operations use provider Kibana client by default

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** the resource SHALL use the provider's Kibana client for all API calls

#### Scenario: Scoped Kibana connection drives operations and version checks

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** the resource SHALL use the scoped Kibana client derived from that block for API calls and stack-version-dependent behavior
