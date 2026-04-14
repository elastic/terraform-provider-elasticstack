## MODIFIED Requirements

### Requirement: Connection (REQ-008–REQ-009)

The resource SHALL use the provider's configured Kibana client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for all API calls of that instance.

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any resource API call runs
- THEN the resource SHALL use the provider-configured Kibana client

#### Scenario: Scoped Kibana connection on resource

- GIVEN `kibana_connection` is configured on the resource
- WHEN any resource API call runs
- THEN the resource SHALL use the scoped Kibana client derived from that block

### Requirement: Connection (REQ-DS-008)

The data source SHALL use the provider's configured Kibana client by default. When `kibana_connection` is configured on the data source, the data source SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for its read operation.

#### Scenario: Provider client used by default for data source

- GIVEN `kibana_connection` is not configured on the data source
- WHEN the data source read runs
- THEN the data source SHALL use the provider-configured Kibana client

#### Scenario: Scoped Kibana connection on data source

- GIVEN `kibana_connection` is configured on the data source
- WHEN the data source read runs
- THEN the data source SHALL use the scoped Kibana client derived from that block
