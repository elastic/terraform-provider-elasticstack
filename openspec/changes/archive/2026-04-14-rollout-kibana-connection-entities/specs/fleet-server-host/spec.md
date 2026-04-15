## ADDED Requirements

### Requirement: Provider-level Fleet client by default with optional scoped override

The `elasticstack_fleet_server_host` resource SHALL use the provider-configured Fleet client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Fleet client for its API calls.

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN an API call runs
- THEN the resource SHALL use the provider-configured Fleet client

#### Scenario: Scoped Fleet connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN an API call runs
- THEN the resource SHALL use the scoped Fleet client derived from that block
