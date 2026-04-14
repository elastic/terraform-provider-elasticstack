## MODIFIED Requirements

### Requirement: Connection (REQ-008–REQ-009)

The resource SHALL use the provider's configured Kibana client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for all API calls of that instance.

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any API call runs
- THEN the provider-configured Kibana client SHALL be used

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any API call runs
- THEN the resource SHALL use the scoped Kibana client derived from that block
