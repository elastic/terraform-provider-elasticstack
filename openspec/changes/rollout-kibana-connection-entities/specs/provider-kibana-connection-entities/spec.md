## ADDED Requirements

### Requirement: In-scope Kibana entity rollout
For every Terraform resource or data source registered by the provider whose type name has prefix `elasticstack_kibana_` in the provider's normal constructor set, the entity SHALL expose `kibana_connection` using the shared provider schema helper for its implementation style and SHALL use the effective scoped client derived from that block when it is configured.

#### Scenario: Covered Kibana entity exposes and honors `kibana_connection`
- **WHEN** a covered `elasticstack_kibana_*` entity is configured with `kibana_connection`
- **THEN** the entity SHALL execute its API operations against the scoped client derived from that block

### Requirement: In-scope Fleet entity rollout
For every Terraform resource or data source registered by the provider whose type name has prefix `elasticstack_fleet_` in the provider's normal constructor set, the entity SHALL expose `kibana_connection` using the shared provider schema helper for its implementation style and SHALL use the effective scoped client derived from that block when it is configured.

#### Scenario: Covered Fleet entity exposes and honors `kibana_connection`
- **WHEN** a covered `elasticstack_fleet_*` entity is configured with `kibana_connection`
- **THEN** the entity SHALL execute its API operations against the scoped client derived from that block

### Requirement: Provider-client fallback
Covered Kibana and Fleet entities SHALL continue to use the provider-configured client when `kibana_connection` is not configured. The presence of the new block SHALL therefore add an optional override path rather than changing default client-resolution behavior.

#### Scenario: Covered entity uses provider client by default
- **WHEN** a covered Kibana or Fleet entity is configured without `kibana_connection`
- **THEN** the entity SHALL use the provider-configured client for its API operations

### Requirement: Shared schema consistency
Covered Kibana and Fleet entities SHALL use the shared SDK or Plugin Framework `kibana_connection` schema helper rather than entity-specific block variants.

#### Scenario: Covered entity uses the shared block definition
- **WHEN** a covered Kibana or Fleet entity defines `kibana_connection`
- **THEN** that definition SHALL come from the shared provider schema helper for the entity's implementation style
