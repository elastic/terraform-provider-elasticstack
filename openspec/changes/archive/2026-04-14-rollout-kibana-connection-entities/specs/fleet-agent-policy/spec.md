## MODIFIED Requirements

### Requirement: Connection — provider client (REQ-008)

The resource SHALL use the provider-level Fleet client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the Fleet client derived from the scoped connection for all API calls.

#### Scenario: Provider client used by default

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** the resource SHALL obtain its Fleet client from the provider configuration

#### Scenario: Scoped Fleet client used when overridden

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** the resource SHALL obtain its effective Fleet client from the scoped connection for all CRUD operations
