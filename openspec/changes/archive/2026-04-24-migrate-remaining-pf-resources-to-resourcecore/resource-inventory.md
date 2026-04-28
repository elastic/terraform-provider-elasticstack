# Plugin Framework resource inventory — `resourcecore` rollout

Audit date: aligned with OpenSpec task 1 (`migrate-remaining-pf-resources-to-resourcecore`).

## Method

- **Source of truth**: resources registered from [`provider/plugin_framework.go`](../../../provider/plugin_framework.go) (`Provider.resources` and `Provider.experimentalResources`).
- **Canonical bootstrap** (compatible with [`internal/resourcecore.Core`](../../../internal/resourcecore/core.go)): `ConvertProviderDataToFactory`, append diagnostics, **return without assigning the factory when `resp.Diagnostics.HasError()`**, static `Metadata` type name `"{provider}_{component}_{resourceName}"` equivalent to existing `TypeName`.
- **Approved scope exception for this rollout**: when a resource's only remaining bootstrap mismatch is assigning the converted factory before returning on diagnostics, that difference is **not** treated as a blocker for migration.
- **Already on `resourcecore`**: embed `*resourcecore.Core` and delegate `Configure` / `Metadata` to the core (no local `client` field for bootstrap).
- **State values**: `migrated` = now embeds `*resourcecore.Core` on this branch, `pilot` = pre-existing `resourcecore` adopter, `pending` = in scope but not yet migrated on this branch, `out_of_scope` = intentionally deferred from this change.

## Pilot resources (already embed `*resourcecore.Core`)

| State | Terraform type name | `resourcecore` component | Literal suffix | Import shape |
| --- | --- | --- | --- | --- |
| `pilot` | `elasticstack_apm_agent_configuration` | `apm` | `agent_configuration` | Passthrough `id` |
| `pilot` | `elasticstack_fleet_integration` | `fleet` | `integration` | No import |
| `pilot` | `elasticstack_kibana_agentbuilder_tool` | `kibana` | `agentbuilder_tool` | Passthrough `id` |
| `pilot` | `elasticstack_elasticsearch_ml_job_state` | `elasticsearch` | `ml_job_state` | Passthrough `id` |

## In scope resources migrated on this branch

These resources are in scope for this change and now migrate through `resourcecore` on the current branch.

| State | Terraform type name | `resourcecore` component | Literal suffix | Package / constructor | Import shape | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `migrated` | `elasticstack_fleet_agent_download_source` | `fleet` | `agent_download_source` | `internal/fleet/agentdownloadsource` | **Custom**: `<source_id>` or `<space_id>/<source_id>` → sets `id`, `source_id`, `space_ids` | — |
| `migrated` | `elasticstack_fleet_agent_policy` | `fleet` | `agent_policy` | `internal/fleet/agentpolicy` | **Custom**: composite or bare policy id → `policy_id`, optional `space_ids` | — |
| `migrated` | `elasticstack_fleet_custom_integration` | `fleet` | `custom_integration` | `internal/fleet/customintegration` | **No import** | — |
| `migrated` | `elasticstack_fleet_elastic_defend_integration_policy` | `fleet` | `elastic_defend_integration_policy` | `internal/fleet/elastic_defend_integration_policy` | **Custom**: passthrough raw import id to `id`; parse with `CompositeIDFromStrFw` — plain id sets `policy_id`; composite id sets `policy_id` and `space_ids` | — |
| `migrated` | `elasticstack_fleet_integration_policy` | `fleet` | `integration_policy` | `internal/fleet/integration_policy` | Passthrough `policy_id` | — |
| `migrated` | `elasticstack_fleet_output` | `fleet` | `output` | `internal/fleet/output` | Passthrough `output_id` | — |
| `migrated` | `elasticstack_fleet_server_host` | `fleet` | `server_host` | `internal/fleet/serverhost` | Passthrough `host_id` | — |
| `migrated` | `elasticstack_kibana_action_connector` | `kibana` | `action_connector` | `internal/kibana/connectors` | **Custom**: sets `id` from import id (no passthrough helper) | — |
| `migrated` | `elasticstack_kibana_agentbuilder_agent` | `kibana` | `agentbuilder_agent` | `internal/kibana/agentbuilderagent` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_agentbuilder_workflow` | `kibana` | `agentbuilder_workflow` | `internal/kibana/agentbuilderworkflow` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_alerting_rule` | `kibana` | `alerting_rule` | `internal/kibana/alertingrule` | **Custom**: composite id → `rule_id`, `space_id`, `id` | — |
| `migrated` | `elasticstack_kibana_data_view` | `kibana` | `data_view` | `internal/kibana/dataview` | **Custom**: composite id → structured state (`SpaceID`, etc.) | — |
| `migrated` | `elasticstack_kibana_default_data_view` | `kibana` | `default_data_view` | `internal/kibana/defaultdataview` | **No import** | — |
| `migrated` | `elasticstack_kibana_import_saved_objects` | `kibana` | `import_saved_objects` | `internal/kibana/import_saved_objects` | **No import** | — |
| `migrated` | `elasticstack_kibana_install_prebuilt_rules` | `kibana` | `install_prebuilt_rules` | `internal/kibana/prebuilt_rules` | **No import** | — |
| `migrated` | `elasticstack_kibana_maintenance_window` | `kibana` | `maintenance_window` | `internal/kibana/maintenance_window` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_security_detection_rule` | `kibana` | `security_detection_rule` | `internal/kibana/security_detection_rule` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_security_enable_rule` | `kibana` | `security_enable_rule` | `internal/kibana/security_enable_rule` | **No import** | — |
| `migrated` | `elasticstack_kibana_security_exception_item` | `kibana` | `security_exception_item` | `internal/kibana/security_exception_item` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_security_exception_list` | `kibana` | `security_exception_list` | `internal/kibana/securityexceptionlist` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_security_list` | `kibana` | `security_list` | `internal/kibana/securitylist` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_security_list_data_streams` | `kibana` | `security_list_data_streams` | `internal/kibana/security_list_data_streams` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_security_list_item` | `kibana` | `security_list_item` | `internal/kibana/securitylistitem` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_slo` | `kibana` | `slo` | `internal/kibana/slo` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_synthetics_monitor` | `kibana` | `synthetics_monitor` | `internal/kibana/synthetics/monitor` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_synthetics_parameter` | `kibana` | `synthetics_parameter` | `internal/kibana/synthetics/parameter` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_synthetics_private_location` | `kibana` | `synthetics_private_location` | `internal/kibana/synthetics/privatelocation` | Passthrough `id` | — |
| `migrated` | `elasticstack_kibana_dashboard` | `kibana` | `dashboard` | `internal/kibana/dashboard` | Passthrough `id` *(experimental registration)* | — |
| `migrated` | `elasticstack_kibana_stream` | `kibana` | `stream` | `internal/kibana/streams` | Passthrough `id` *(experimental registration)* | — |
| `migrated` | `elasticstack_elasticsearch_index` | `elasticsearch` | `index` | `internal/elasticsearch/index/index` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_data_stream_lifecycle` | `elasticsearch` | `data_stream_lifecycle` | `internal/elasticsearch/index/datastreamlifecycle` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_index_lifecycle` | `elasticsearch` | `index_lifecycle` | `internal/elasticsearch/index/ilm` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_index_template_ilm_attachment` | `elasticsearch` | `index_template_ilm_attachment` | `internal/elasticsearch/index/templateilmattachment` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_index_alias` | `elasticsearch` | `index_alias` | `internal/elasticsearch/index/alias` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_script` | `elasticsearch` | `script` | `internal/elasticsearch/cluster/script` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_enrich_policy` | `elasticsearch` | `enrich_policy` | `internal/elasticsearch/enrich` | **Custom**: passthrough `id`, then sets `execute = true` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_security_api_key` | `elasticsearch` | `security_api_key` | `internal/elasticsearch/security/api_key` | No import | Previously treated as out of scope because of `configuredResources`; that package-level slice was unused and removed, so the resource now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_security_system_user` | `elasticsearch` | `security_system_user` | `internal/elasticsearch/security/systemuser` | No import | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_security_user` | `elasticsearch` | `security_user` | `internal/elasticsearch/security/user` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_security_role` | `elasticsearch` | `security_role` | `internal/elasticsearch/security/role` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_security_role_mapping` | `elasticsearch` | `security_role_mapping` | `internal/elasticsearch/security/rolemapping` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_watch` | `elasticsearch` | `watch` | `internal/elasticsearch/watcher/watch` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_inference_endpoint` | `elasticsearch` | `inference_endpoint` | `internal/elasticsearch/inference/inferenceendpoint` | Passthrough `id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_ml_anomaly_detection_job` | `elasticsearch` | `ml_anomaly_detection_job` | `internal/elasticsearch/ml/anomalydetectionjob` | **Custom**: composite id → `id`, `job_id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_ml_datafeed` | `elasticsearch` | `ml_datafeed` | `internal/elasticsearch/ml/datafeed` | **Custom**: passthrough `id`, then sets `datafeed_id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |
| `migrated` | `elasticstack_elasticsearch_ml_datafeed_state` | `elasticsearch` | `ml_datafeed_state` | `internal/elasticsearch/ml/datafeed_state` | Passthrough `datafeed_id` | Prior local `Configure` assigned factory before return on diagnostics; now uses `resourcecore` |

## Open questions (from design, unchanged)

- Provider registry test scope: `Provider.resources` only vs including `Provider.experimentalResources` (this inventory lists experimental Kibana resources explicitly).
