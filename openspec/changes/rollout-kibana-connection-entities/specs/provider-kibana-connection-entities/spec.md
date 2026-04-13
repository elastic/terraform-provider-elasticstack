## ADDED Requirements

### Requirement: In-scope Kibana entity rollout
The rollout's covered Kibana entities SHALL be exactly the Kibana resources and data sources registered in `provider/provider.go` and `provider/plugin_framework.go`: `elasticstack_kibana_action_connector`, `elasticstack_kibana_agentbuilder_export_workflow`, `elasticstack_kibana_agentbuilder_workflow`, `elasticstack_kibana_alerting_rule`, `elasticstack_kibana_dashboard`, `elasticstack_kibana_data_view`, `elasticstack_kibana_default_data_view`, `elasticstack_kibana_export_saved_objects`, `elasticstack_kibana_import_saved_objects`, `elasticstack_kibana_maintenance_window`, `elasticstack_kibana_security_detection_rule`, `elasticstack_kibana_security_enable_rule`, `elasticstack_kibana_security_exception_item`, `elasticstack_kibana_security_exception_list`, `elasticstack_kibana_security_list`, `elasticstack_kibana_security_list_data_streams`, `elasticstack_kibana_security_list_item`, `elasticstack_kibana_security_role`, `elasticstack_kibana_slo`, `elasticstack_kibana_space`, `elasticstack_kibana_spaces`, `elasticstack_kibana_stream`, `elasticstack_kibana_synthetics_monitor`, `elasticstack_kibana_synthetics_parameter`, and `elasticstack_kibana_synthetics_private_location`. Each covered Kibana entity SHALL expose `kibana_connection` using the shared provider schema helper for its implementation style and SHALL use the effective scoped client derived from that block when it is configured.

#### Scenario: Covered Kibana entity exposes and honors `kibana_connection`
- **WHEN** a covered `elasticstack_kibana_*` entity is configured with `kibana_connection`
- **THEN** the entity SHALL execute its API operations against the scoped client derived from that block

### Requirement: In-scope Fleet entity rollout
The rollout's covered Fleet entities SHALL be exactly the Fleet resources and data sources registered in `provider/plugin_framework.go`: `elasticstack_fleet_agent_policy`, `elasticstack_fleet_elastic_defend_integration_policy`, `elasticstack_fleet_enrollment_tokens`, `elasticstack_fleet_integration`, `elasticstack_fleet_integration_policy`, `elasticstack_fleet_output`, and `elasticstack_fleet_server_host`. Each covered Fleet entity SHALL expose `kibana_connection` using the shared provider schema helper for its implementation style and SHALL use the effective scoped client derived from that block when it is configured.

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
