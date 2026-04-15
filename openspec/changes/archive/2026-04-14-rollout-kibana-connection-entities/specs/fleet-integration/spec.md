## MODIFIED Requirements

### Requirement: Connection (REQ-008)

The resource and data source SHALL use the provider-level Fleet client obtained from provider configuration by default. When `kibana_connection` is configured on the resource or data source, the entity SHALL resolve an effective scoped client from that block and SHALL use the Fleet client derived from the scoped connection for the API call.

#### Scenario: Provider Fleet client used by default

- **WHEN** `kibana_connection` is not configured on the resource or data source
- **THEN** the entity SHALL obtain its Fleet client from the provider configuration

#### Scenario: Scoped Fleet client used when overridden

- **WHEN** `kibana_connection` is configured on the resource or data source
- **THEN** the entity SHALL obtain its effective Fleet client from the scoped connection for the read or lifecycle operation
