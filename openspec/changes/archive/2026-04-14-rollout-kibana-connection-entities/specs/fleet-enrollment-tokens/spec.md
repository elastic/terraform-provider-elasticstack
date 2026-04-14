## ADDED Requirements

### Requirement: Provider-level Fleet client by default with optional scoped override

The `elasticstack_fleet_enrollment_tokens` data source SHALL use the provider-configured Fleet client by default. When `kibana_connection` is configured on the data source, the data source SHALL resolve an effective scoped client from that block and SHALL use the scoped Fleet client for its read operation.

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the data source
- WHEN the data source read runs
- THEN the data source SHALL use the provider-configured Fleet client

#### Scenario: Scoped Fleet connection

- GIVEN `kibana_connection` is configured on the data source
- WHEN the data source read runs
- THEN the data source SHALL use the scoped Fleet client derived from that block
