## ADDED Requirements

### Requirement: Provider-level Fleet client by default with optional scoped override

The `elasticstack_fleet_output` resource and data source SHALL use the provider-configured Fleet client by default. When `kibana_connection` is configured on the resource or data source, that entity SHALL resolve an effective scoped client from the block and SHALL use the scoped Fleet client for its API calls.

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the Fleet output resource or data source
- WHEN an API call runs
- THEN the entity SHALL use the provider-configured Fleet client

#### Scenario: Scoped Fleet connection

- GIVEN `kibana_connection` is configured on the Fleet output resource or data source
- WHEN an API call runs
- THEN the entity SHALL use the scoped Fleet client derived from that block
