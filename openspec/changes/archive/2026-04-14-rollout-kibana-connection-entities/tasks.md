## 1. Add schema and model support for the full registered rollout set

- [x] 1.1 Add `kibana_connection` to the SDK-registered Kibana entities from `provider/provider.go`: `elasticstack_kibana_action_connector`, `elasticstack_kibana_security_role`, and `elasticstack_kibana_space`.
- [x] 1.2 Add `kibana_connection` to the Plugin Framework Kibana entities from `provider/plugin_framework.go`: `elasticstack_kibana_action_connector`, `elasticstack_kibana_agentbuilder_export_workflow`, `elasticstack_kibana_agentbuilder_workflow`, `elasticstack_kibana_alerting_rule`, `elasticstack_kibana_dashboard`, `elasticstack_kibana_data_view`, `elasticstack_kibana_default_data_view`, `elasticstack_kibana_export_saved_objects`, `elasticstack_kibana_import_saved_objects`, `elasticstack_kibana_maintenance_window`, `elasticstack_kibana_security_detection_rule`, `elasticstack_kibana_security_enable_rule`, `elasticstack_kibana_security_exception_item`, `elasticstack_kibana_security_exception_list`, `elasticstack_kibana_security_list`, `elasticstack_kibana_security_list_data_streams`, `elasticstack_kibana_security_list_item`, `elasticstack_kibana_slo`, `elasticstack_kibana_spaces`, `elasticstack_kibana_stream`, `elasticstack_kibana_synthetics_monitor`, `elasticstack_kibana_synthetics_parameter`, and `elasticstack_kibana_synthetics_private_location`.
- [x] 1.3 Add `kibana_connection` to the Plugin Framework Fleet entities from `provider/plugin_framework.go`: `elasticstack_fleet_agent_policy`, `elasticstack_fleet_elastic_defend_integration_policy`, `elasticstack_fleet_enrollment_tokens`, `elasticstack_fleet_integration`, `elasticstack_fleet_integration_policy`, `elasticstack_fleet_output`, and `elasticstack_fleet_server_host`.
- [x] 1.4 Add or update entity models so resource and data source state carries `kibana_connection` where required across the full rollout set.

## 2. Adopt effective scoped clients everywhere the provider registers Kibana or Fleet entities

- [x] 2.1 Update the SDK-registered Kibana entities so they resolve an effective client from `kibana_connection` and use the scoped Kibana client surfaces when configured.
- [x] 2.2 Update the Plugin Framework Kibana entities so they resolve an effective client from `kibana_connection` and use the scoped Kibana client surfaces when configured, including the conditionally registered `elasticstack_kibana_dashboard` and `elasticstack_kibana_stream` resources.
- [x] 2.3 Update the Plugin Framework Fleet entities so they resolve an effective client from `kibana_connection` and use the scoped Fleet client when configured.
- [x] 2.4 Ensure version checks, space-aware operations, import flows, and read-after-write paths continue to run against the effective client for every adopted entity.

## 3. Finish rollout docs and regression checks

- [x] 3.1 Update the affected entity documentation and examples so `kibana_connection` appears everywhere this rollout adds support.
- [x] 3.2 Add or update focused entity tests that confirm default provider behavior still works and scoped `kibana_connection` is plumbed through the adopted code paths for the registered rollout set.
