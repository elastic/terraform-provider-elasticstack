## [0.16.3] - 2026-07-22

### Breaking changes

`options_list_control_config` and `range_slider_control_config` are restructured from flat attribute blocks into a two-branch union: `by_field {}` (the existing data-view-field variant) and `by_esql {}` (new: ES|QL query variant). Exactly one of the two must be set. Existing configurations must wrap their current attributes in `by_field { ... }`. A Plugin Framework state upgrader (schema v0 -> v1) automatically migrates existing state on the next `terraform apply`; no manual state surgery is required, but `.tf` files must be updated to the new nested shape afterwards.

### Changes

- defer privilege validation for unknown dynamic block values in `elasticstack_kibana_security_role` ([#4176](https://github.com/elastic/terraform-provider-elasticstack/pull/4176))
- Fix `terraform plan` failure for `elasticstack_elasticsearch_security_api_key` when legacy state contains empty-string `metadata` or `role_descriptors` JSON fields. ([#4173](https://github.com/elastic/terraform-provider-elasticstack/pull/4173))
- Skip the data stream lifecycle DELETE on Elastic Cloud Serverless so destroying `elasticstack_elasticsearch_data_stream_lifecycle` completes instead of failing on the HTTP 410 `api_not_available_exception`. ([#4160](https://github.com/elastic/terraform-provider-elasticstack/pull/4160))
- Fix Fleet Kafka output apply failures when gzip compression level or sasl block are omitted ([#4122](https://github.com/elastic/terraform-provider-elasticstack/pull/4122))
- Normalize empty rule and action `params` values while upgrading alerting-rule state from v0 to v1. ([#4166](https://github.com/elastic/terraform-provider-elasticstack/pull/4166))
- Fix ILM v0→v1 state upgrader invalid JSON error when SDKv2 state stored empty-string values for metadata or allocate routing attributes. ([#4167](https://github.com/elastic/terraform-provider-elasticstack/pull/4167))
- Suppress Kibana 9.5+ server-injected `ENVIRONMENT_ALL` default on APM service-map dashboard panels when the practitioner omits `environment`. ([#4128](https://github.com/elastic/terraform-provider-elasticstack/pull/4128))
- add links panel support to Kibana dashboard resource ([#4078](https://github.com/elastic/terraform-provider-elasticstack/pull/4078))
- Reject ML anomaly detection job `results_index_name` values starting with `custom-` to prevent plan/apply drift. ([#4107](https://github.com/elastic/terraform-provider-elasticstack/pull/4107))
- Fall back to full package uninstall when Fleet 9.5 rejects DeleteKibanaAssets in the install space during integration destroy. ([#4073](https://github.com/elastic/terraform-provider-elasticstack/pull/4073))
- Add `field_stats_table_config` typed panel block to `elasticstack_kibana_dashboard` ([#4074](https://github.com/elastic/terraform-provider-elasticstack/pull/4074))
- Fix `elasticstack_fleet_integration_policy` post-apply inconsistency errors on Elastic Stack 9.5 caused by Fleet injecting server-managed `data_stream.*` keys into stream vars and populating the `defaults` block. ([#4055](https://github.com/elastic/terraform-provider-elasticstack/pull/4055))
- Fix Security Entity Store flakiness — wait for uninstall/started state, retry transient HTTP 500s on install/link/entity create, and isolate acceptance tests. ([#4062](https://github.com/elastic/terraform-provider-elasticstack/pull/4062))
- Add `elasticstack_fleet_agentless_policy` resource for managing Fleet agentless policies (Elastic Cloud Hosted / Serverless, Kibana 9.3.0+). ([#4034](https://github.com/elastic/terraform-provider-elasticstack/pull/4034))
- Fix elasticstack_elasticsearch_security_role read failure on Elasticsearch 9.5+ by fetching the role via raw transport and decoding the global privilege object as opaque JSON, and strip the server-injected empty data_source array from state. ([#4059](https://github.com/elastic/terraform-provider-elasticstack/pull/4059))
- Add ES|QL-sourced variant support for `options_list_control` and `range_slider_control` dashboard panels via new `by_field`/`by_esql` branches. ([#4039](https://github.com/elastic/terraform-provider-elasticstack/pull/4039))
- Fixes "Provider produced inconsistent result after apply" for `elasticstack_kibana_dashboard` on Kibana 9.5+ when the `description` attribute is omitted. The Kibana 9.5 dashboard API now echoes back `description: ""` instead of omitting the field; the provider now preserves `null` in state when the practitioner omitted `description`, while still preserving an explicit `description = ""`. No action required for existing configurations. ([#4057](https://github.com/elastic/terraform-provider-elasticstack/pull/4057))
- Add `ml_anomaly_charts_config` panel type support to `elasticstack_kibana_dashboard`. ([#4037](https://github.com/elastic/terraform-provider-elasticstack/pull/4037))
- Added `aiops_log_rate_analysis_config`, `aiops_pattern_analysis_config`, and `aiops_change_point_chart_config` typed panel blocks to `elasticstack_kibana_dashboard`, enabling AIOps panels with typed validation and drift-safe planning instead of raw `config_json`. ([#4026](https://github.com/elastic/terraform-provider-elasticstack/pull/4026))
- Add typed `apm_service_map_config` block to `elasticstack_kibana_dashboard` resource ([#4025](https://github.com/elastic/terraform-provider-elasticstack/pull/4025))
- Add `ml_anomaly_swimlane_config` and `ml_single_metric_viewer_config` typed panel blocks to `elasticstack_kibana_dashboard`. ([#4017](https://github.com/elastic/terraform-provider-elasticstack/pull/4017))
- Fix perpetual diff and apply crash for component template settings and alias routing drift ([#3998](https://github.com/elastic/terraform-provider-elasticstack/pull/3998))
- Prevent "Provider produced inconsistent result after apply" errors when optional string attributes (chunk_size, max_snapshot_bytes_per_sec, max_restore_bytes_per_sec) are set to empty string in the snapshot repository resource. ([#3832](https://github.com/elastic/terraform-provider-elasticstack/pull/3832))
- Fix Fleet agent policy create when policy_id is omitted and add plan-time validation for explicit policy_id values ([#3937](https://github.com/elastic/terraform-provider-elasticstack/pull/3937))
- Add `elasticstack_kibana_tag` resource and `elasticstack_kibana_tags` data source for managing Kibana tags and listing tags in a space. ([#3921](https://github.com/elastic/terraform-provider-elasticstack/pull/3921))
- Fix component and index template state upgrade when SDK stored empty strings for mappings, settings, or metadata. ([#3914](https://github.com/elastic/terraform-provider-elasticstack/pull/3914))
- Add `elasticstack_kibana_osquery_pack` resource and data source for managing user-defined Osquery query packs and reading packs (including prebuilt read-only packs). ([#3893](https://github.com/elastic/terraform-provider-elasticstack/pull/3893))
- Document supported `check` and `response` sub-keys (including `check.response.status`) on the `elasticstack_kibana_synthetics_monitor` HTTP attribute. ([#3895](https://github.com/elastic/terraform-provider-elasticstack/pull/3895))
- Fix empty string settings in snapshot repository blocks by preserving prior state when Elasticsearch omits those keys from the API response. ([#3719](https://github.com/elastic/terraform-provider-elasticstack/pull/3719))
- Add advanced_settings map to elasticstack_fleet_elastic_defend_integration_policy for Elastic Defend advanced policy configuration. ([#3845](https://github.com/elastic/terraform-provider-elasticstack/pull/3845))
- Add `elasticstack_kibana_osquery_saved_query` resource and data source for managing Kibana Osquery saved queries. ([#3883](https://github.com/elastic/terraform-provider-elasticstack/pull/3883))

## [0.16.2] - 2026-06-23

### Breaking changes

`elasticstack_elasticsearch_ml_anomaly_detection_job`: `timeouts` is now an attribute instead of a block. Replace block syntax (`timeouts { delete = "20m" }`) with attribute syntax (`timeouts = { delete = "20m" }`).

### Changes

- Add `elasticstack_kibana_osquery_pack` resource for managing user-defined Osquery query packs and data source for reading packs (including prebuilt read-only packs).
- Fix duplicate Authorization headers when ES and Kibana/Fleet use different auth methods ([#3722](https://github.com/elastic/terraform-provider-elasticstack/pull/3722))
- Add Elasticsearch CA fingerprint connection support via `ca_fingerprint` and `ELASTICSEARCH_CA_FINGERPRINT`. ([#3837](https://github.com/elastic/terraform-provider-elasticstack/pull/3837))
- preserve prior input values across ES redaction on read for `elasticstack_elasticsearch_watch` resources with HTTP basic auth ([#3839](https://github.com/elastic/terraform-provider-elasticstack/pull/3839))
- Adds allow_restricted_indices to remote_indices on elasticstack_elasticsearch_security_role and elasticstack_kibana_security_role ([#3701](https://github.com/elastic/terraform-provider-elasticstack/pull/3701))
- Relax provider-side length validation for ML datafeed and filter IDs ([#3798](https://github.com/elastic/terraform-provider-elasticstack/pull/3798))
- Stop `elasticstack_kibana_dashboard` from injecting `empty_as_null` into Lens metric `config_json` for operations the Kibana API rejects (e.g. `percentile`), which previously caused an HTTP 400 on apply. ([#3720](https://github.com/elastic/terraform-provider-elasticstack/pull/3720))
- Fix "Provider produced inconsistent result after apply" for `.slack_api` action connectors by normalizing the planned `config` value the same way the read response is normalized. ([#3749](https://github.com/elastic/terraform-provider-elasticstack/pull/3749))
- The `elasticstack_kibana_dashboard` resource now supports an optional user-supplied `dashboard_id`, creating the dashboard via PUT upsert so practitioners can assign stable, human-readable identifiers. ([#3788](https://github.com/elastic/terraform-provider-elasticstack/pull/3788))
- Promote the Kibana Streams resource from experimental to tech preview, making it publicly available without requiring the `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` environment variable. ([#3782](https://github.com/elastic/terraform-provider-elasticstack/pull/3782))
- Add `elasticstack_elasticsearch_ccr_follower_index` and `elasticstack_elasticsearch_ccr_auto_follow_pattern` resources for managing Cross-Cluster Replication. ([#3615](https://github.com/elastic/terraform-provider-elasticstack/pull/3615))
- Add uniform `timeouts` support to all entitycore-envelope-backed Elasticsearch and Kibana resources; `elasticsearch_ml_anomaly_detection_job` `timeouts` changes from block to attribute syntax. ([#3607](https://github.com/elastic/terraform-provider-elasticstack/pull/3607))
- Adds Terraform resource and data source for managing the Kibana Security Entity Store resolution links and resolution groups. ([#3514](https://github.com/elastic/terraform-provider-elasticstack/pull/3514))
- Add `elasticstack_elasticsearch_ml_trained_model_deployment` resource for managing ML trained model deployments ([#3532](https://github.com/elastic/terraform-provider-elasticstack/pull/3532))
- Add Kibana Security Entity Store entity resource and data source ([#3529](https://github.com/elastic/terraform-provider-elasticstack/pull/3529))
- Preserve space_ids value when Fleet API omits the field, preventing "inconsistent result after apply" error ([#3566](https://github.com/elastic/terraform-provider-elasticstack/pull/3566))
- Stop `elasticstack_kibana_security_exception_item` and `elasticstack_kibana_security_exception_list` from tripping "produced inconsistent result after apply" when `tags` or `os_types` is configured as an empty set. ([#3552](https://github.com/elastic/terraform-provider-elasticstack/pull/3552))
- Add elasticstack_elasticsearch_ml_trained_model_alias resource ([#3533](https://github.com/elastic/terraform-provider-elasticstack/pull/3533))
- Add data source for reading Elasticsearch ML trained model metadata. ([#3531](https://github.com/elastic/terraform-provider-elasticstack/pull/3531))
- Extended plan-time `params` validation for `elasticstack_kibana_alerting_rule` to cover all 35 rule types known to the generated kbapi client via discriminator dispatch, replacing the previous 12-entry hand-maintained map. This provides stricter validation for previously pass-through types such as `observability.rules.custom_threshold`. ([#3510](https://github.com/elastic/terraform-provider-elasticstack/pull/3510))
- Adds Terraform resource and data source for managing the Kibana Security Entity Store. ([#3497](https://github.com/elastic/terraform-provider-elasticstack/pull/3497))
- Fix 404 error on fleet_server_host update when host_id is omitted or changed ([#3508](https://github.com/elastic/terraform-provider-elasticstack/pull/3508))
- Add agent_policy_ids support to elasticstack_fleet_elastic_defend_integration_policy and change agent_policy_id from Required to Optional. ([#3492](https://github.com/elastic/terraform-provider-elasticstack/pull/3492))

## [0.16.1] - 2026-06-01

### Changes

- Fix 404 error on update and destroy of `elasticstack_fleet_server_host` when `host_id` is omitted from config by adding `UseStateForUnknown()` and `RequiresReplace()` plan modifiers to `host_id`. Changing `host_id` explicitly now triggers destroy-and-recreate instead of a broken update. Fixes [#864](https://github.com/elastic/terraform-provider-elasticstack/issues/864).
- Add elasticstack_elasticsearch_connector resource and data source plus elasticstack_elasticsearch_connector_sync_job_create action for Elasticsearch content connectors. ([#3435](https://github.com/elastic/terraform-provider-elasticstack/pull/3435))
- Preserve S3 endpoint and path_style_access in snapshot repository PUT bodies ([#3447](https://github.com/elastic/terraform-provider-elasticstack/pull/3447))
- Fix inconsistent state when metadata is set to jsonencode({}) on elasticsearch_security_user ([#3448](https://github.com/elastic/terraform-provider-elasticstack/pull/3448))
- Make Kibana dashboard panel-level time_range optional so panels can use the dashboard global time range ([#3436](https://github.com/elastic/terraform-provider-elasticstack/pull/3436))
- Fix plan-time `Value Conversion Error` when the whole `analysis_config` block (or `analysis_config.per_partition_categorization`) is sourced from a Terraform variable or `for_each`. ([#3425](https://github.com/elastic/terraform-provider-elasticstack/pull/3425))
- Allow ``elasticstack_kibana_space`` to manage the default Kibana space without errors and return an actionable import diagnostic on create 409. ([#3423](https://github.com/elastic/terraform-provider-elasticstack/pull/3423))
- Fixed false-positive validation error for cluster_settings when persistent/transient blocks are populated via dynamic blocks driven by local values. ([#3411](https://github.com/elastic/terraform-provider-elasticstack/pull/3411))
- Fix `Provider produced inconsistent result after apply` and plan drift on Kibana Lens dashboard panels caused by Kibana-injected server defaults (issue #3402 and related across every Lens panel type) ([#3404](https://github.com/elastic/terraform-provider-elasticstack/pull/3404))
- Fix `elasticstack_kibana_dashboard` round-trip drift for Lens XY/gauge/heatmap/image panels, ES|QL controls, and empty `panels` lists; add getting-started, operations, and advanced Kibana dashboard guides with example configs and screenshots. ([#3391](https://github.com/elastic/terraform-provider-elasticstack/pull/3391))
- Add elasticsearch_snapshot_create and elasticsearch_snapshot_restore provider-defined actions for on-demand snapshot creation and restore. ([#3376](https://github.com/elastic/terraform-provider-elasticstack/pull/3376))
- Add `elasticstack_elasticsearch_query_ruleset` resource and data source for the Elasticsearch Query Rules API. ([#3365](https://github.com/elastic/terraform-provider-elasticstack/pull/3365))
- improve security role documentation and add automated drift detection for Kibana feature privilege docs ([#3339](https://github.com/elastic/terraform-provider-elasticstack/pull/3339))
- webhook connector with sensitive config no longer fails with "inconsistent values for sensitive attribute" when `method` is omitted ([#3358](https://github.com/elastic/terraform-provider-elasticstack/pull/3358))
- Hydrate all `elasticstack_elasticsearch_index` settings fields on import so plans no longer show spurious drift. ([#3360](https://github.com/elastic/terraform-provider-elasticstack/pull/3360))
- Surface Kibana Boom error message when import saved objects returns non-200/400 HTTP responses ([#3359](https://github.com/elastic/terraform-provider-elasticstack/pull/3359))
- `elasticstack_kibana_security_exception_list` now accepts `type=rule_default`, allowing terraform to explicitly manage per-rule exception list containers that were previously expected to auto-create from a detection rule POST (which does not actually happen). ([#3348](https://github.com/elastic/terraform-provider-elasticstack/pull/3348))
- Add `elasticstack_elasticsearch_synonym_set` resource and data source for full CRUD management of Elasticsearch synonym sets via the Synonyms API. ([#3335](https://github.com/elastic/terraform-provider-elasticstack/pull/3335))
- Make Elasticsearch resources serverless-aware by routing all version- and flavor-gating through serverless-safe primitives. ([#3325](https://github.com/elastic/terraform-provider-elasticstack/pull/3325))
- Prevent panic in `elasticstack_fleet_agent_policy` when a `global_data_tags` entry has neither string nor number value set. ([#3322](https://github.com/elastic/terraform-provider-elasticstack/pull/3322))
- Add write-only `secrets_wo` and `secrets_wo_version` attributes to `elasticstack_kibana_action_connector` for ephemeral secret sources. ([#3323](https://github.com/elastic/terraform-provider-elasticstack/pull/3323))
- Provider with only a fleet block can serve Kibana resources; missing-endpoint errors surface earlier at plan time. ([#3316](https://github.com/elastic/terraform-provider-elasticstack/pull/3316))

## [0.16.0] - 2026-05-25

### Breaking changes

- The `lens-dashboard-app` panel type is no longer supported in `elasticstack_kibana_dashboard`. Use `type = "vis"` instead.

`elasticstack_kibana_security_detection_rule` `actions.alerts_filter` is now a structured nested attribute with `query` (`kql`, `filters_json`) and optional `timeframe` (`days`, `timezone`, `hours_start`, `hours_end`), replacing the broken `map(string)` shape.

### Changes

- Add `elasticstack_elasticsearch_ml_calendar_job` to assign one ML anomaly detection job to a calendar; entry added under Unreleased in `CHANGELOG.md`. ([#2933](https://github.com/elastic/terraform-provider-elasticstack/pull/2933))
- guard nil package list decoding failures in Fleet package listing ([#3275](https://github.com/elastic/terraform-provider-elasticstack/pull/3275))
- Allow `elasticstack_kibana_space.disabled_features` be set when `solution` is `classic`, unset, or unknown. ([#3217](https://github.com/elastic/terraform-provider-elasticstack/pull/3217))
- Add ephemeral `elasticstack_elasticsearch_security_api_key` resource for in-memory API key credentials ([#3176](https://github.com/elastic/terraform-provider-elasticstack/pull/3176))
- `exceptions_list[].type` on `elasticstack_kibana_security_detection_rule` now accepts `rule_default` and `endpoint_trusted_devices`, fixing plan-time failures for rules with user-defined rule-local exception lists. ([#3000](https://github.com/elastic/terraform-provider-elasticstack/pull/3000))
- Remove `lens-dashboard-app` panel type from `elasticstack_kibana_dashboard`; migrate to `type = "vis"`. ([#3209](https://github.com/elastic/terraform-provider-elasticstack/pull/3209))
- Reject empty mappings in indexmappings resource at plan time ([#3186](https://github.com/elastic/terraform-provider-elasticstack/pull/3186))
- Add `elasticstack_elasticsearch_ml_calendar` and `elasticstack_elasticsearch_ml_calendar_event` resources for managing Elasticsearch ML calendars, scheduled events, and job associations in Terraform. ([#1969](https://github.com/elastic/terraform-provider-elasticstack/pull/1969))
- Normalise empty-object mappings/settings on read for component templates to prevent inconsistent state after apply (issue #609). ([#3175](https://github.com/elastic/terraform-provider-elasticstack/pull/3175))
- Omit ILM allocate `number_of_replicas` and `total_shards_per_node` from API requests when not explicitly configured, so routing-filter-only policies no longer override index template settings. ([#3174](https://github.com/elastic/terraform-provider-elasticstack/pull/3174))
- Fix perpetual plan diff on role mapping `rules` when field values use single-element arrays (e.g. `groups = ["project1"]`). ([#3172](https://github.com/elastic/terraform-provider-elasticstack/pull/3172))
- Fix perpetual plan diff for Fleet input-type integrations (e.g., gcp_pubsub) by extracting package-level variable defaults. ([#3145](https://github.com/elastic/terraform-provider-elasticstack/pull/3145))
- Preserve explicit `start`/`end` on `elasticstack_elasticsearch_ml_datafeed_state` and expose ES-effective search bounds via `effective_search_start` / `effective_search_end`. ([#3151](https://github.com/elastic/terraform-provider-elasticstack/pull/3151))
- `elasticstack_kibana_data_view` namespace updates now apply correctly for data views in non-default Kibana spaces. ([#3150](https://github.com/elastic/terraform-provider-elasticstack/pull/3150))
- Fix "Provider produced inconsistent result after apply" for elasticstack_elasticsearch_index_template and elasticstack_elasticsearch_component_template when template.settings contains keys not modeled by the go-elasticsearch typed client (e.g. index.search.slowlog.include) or string-encoded scalars coerced by typed structs (e.g. index.lifecycle.parse_origination_date). ([#3126](https://github.com/elastic/terraform-provider-elasticstack/pull/3126))
- Fix `elasticstack_kibana_security_detection_rule` `actions.alerts_filter` with structured nested blocks; migrate `actions` and `frequency` to block syntax. ([#3123](https://github.com/elastic/terraform-provider-elasticstack/pull/3123))
- Add support for configuring Agent Builder skills ([#3006](https://github.com/elastic/terraform-provider-elasticstack/pull/3006))
- Add `elasticstack_elasticsearch_index_mappings` resource for managing a subset of mappings on an existing index ([#3121](https://github.com/elastic/terraform-provider-elasticstack/pull/3121))
- Fix drift in elasticstack_elasticsearch_ingest_pipeline when a processor contains fields not modeled by the go-elasticsearch typed client (e.g. override on rename processor) ([#3122](https://github.com/elastic/terraform-provider-elasticstack/pull/3122))

## [0.15.2] - 2026-05-18

### Changes

- Fix Read and Delete failures on ElasticsearchResource when the state ID is a plain (non-composite) identifier, e.g. after import with only the resource name. ([#3084](https://github.com/elastic/terraform-provider-elasticstack/pull/3084))
- Migrate `elasticstack_kibana_security_role` resource and data source implementation to Terraform Plugin Framework while preserving the existing schema and behavior. ([#3071](https://github.com/elastic/terraform-provider-elasticstack/pull/3071))
- Migrate `elasticstack_kibana_space` to Plugin Framework while preserving the resource schema, defaults, and observable behavior. ([#3073](https://github.com/elastic/terraform-provider-elasticstack/pull/3073))
- Fix "Provider produced inconsistent result after apply" when nested list attributes are set to empty lists in detection rules. ([#3074](https://github.com/elastic/terraform-provider-elasticstack/pull/3074))
- Migrate `elasticstack_kibana_action_connector` data source to Terraform Plugin Framework; behavior is preserved for existing configurations. ([#3072](https://github.com/elastic/terraform-provider-elasticstack/pull/3072))
- Add `allow_auto_create` optional attribute to the `elasticstack_elasticsearch_index_template` resource and data source ([#3059](https://github.com/elastic/terraform-provider-elasticstack/pull/3059))
- Fix space-aware URL construction to correctly handle Kibana base path configurations ([#3053](https://github.com/elastic/terraform-provider-elasticstack/pull/3053))
- Fix state consistency error when runtime_field_map is omitted from configuration after previously being set ([#2592](https://github.com/elastic/terraform-provider-elasticstack/pull/2592))
- Fix serverless version gating for data_stream_options in elasticsearch_index_template and elasticsearch_index_component_template, and for  ignore_missing_component_templates in elasticsearch_index_template. ([#3023](https://github.com/elastic/terraform-provider-elasticstack/pull/3023))
- Fix spurious plan drift for component/index templates using boolean scalar mappings or null settings after Elasticsearch echoes them as strings ([#3014](https://github.com/elastic/terraform-provider-elasticstack/pull/3014))
- Add `template.data_stream_options` block to `elasticstack_elasticsearch_component_template` to configure the failure store on data streams composed from this component. Requires Elasticsearch >= 9.1.0. ([#2963](https://github.com/elastic/terraform-provider-elasticstack/pull/2963))
- Fix unknown-value handling in the required-if validator for custom SLO metric indicators. ([#3001](https://github.com/elastic/terraform-provider-elasticstack/pull/3001))
- Suppress drift from Kibana field popularity counts on `elasticstack_kibana_data_view.field_attrs` and apply field metadata updates in place instead of replacing the data view. ([#2964](https://github.com/elastic/terraform-provider-elasticstack/pull/2964))

## [0.15.1] - 2026-05-15

### Changes

- Fix plan-time Value Conversion Error regression when analysis_config.detectors is sourced from a Terraform variable or module input ([#2981](https://github.com/elastic/terraform-provider-elasticstack/pull/2981))
- Fix "Provider produced inconsistent result after apply" for ingest pipelines when processors use single-string fields that the typed client returns as arrays. ([#2979](https://github.com/elastic/terraform-provider-elasticstack/pull/2979))
- increase supported Kibana SLO ID length from 36 to 48 characters ([#2973](https://github.com/elastic/terraform-provider-elasticstack/pull/2973))
- Allow duplicate action group IDs in `kibana_alerting_rule`. This reverts an incorrect new validation from 0.15.0 ([#2969](https://github.com/elastic/terraform-provider-elasticstack/pull/2969))
- Fix "Provider produced inconsistent result after apply" for component templates with nested object mappings. ([#2968](https://github.com/elastic/terraform-provider-elasticstack/pull/2968))

## [0.15.0] - 2026-05-13

### Breaking changes

Removed top-level `enabled` from `elasticstack_fleet_integration_policy`. In practice this field was unusable, causing state consistency issues unless it was `true`. Kibana doesn't support enabling/disabling an integration policy directly.

The documented minimum supported Elastic Stack version is now 8.0. 7.x is no longer included in the acceptance test matrix or officially supported. Compatibility branches and version gates for pre-8.0 Elasticsearch behavior have been removed from the transform and ILM resources.

#### `elasticstack_kibana_security_detection_rule` action `params` format change

Previously `elasticstack_kibana_security_detection_rule` used a map of strings for action parameters. This caused issues with actions requiring non-string based parameters (see <https://github.com/elastic/terraform-provider-elasticstack/issues/2339> for an example). This has been changed to a single JSON string value which supports arbitrary param values.

Previously

```hcl
resource "elasticstack_kibana_security_detection_rule" "test" {
...

  actions = [
    {
...
      params = {
        message = "Test state upgrade alert"
      }
...
  ]
}
```

becomes

```hcl
resource "elasticstack_kibana_security_detection_rule" "test" {
...

  actions = [
    {
...
      params = jsonencode({
        message = "Test state upgrade alert"
      })
...
  ]
}
```

### Changes

- Fleet resources now retry on HTTP 409 Conflict with exponential backoff, resolving failures when running `terraform apply` with `parallelism > 1`. ([#2911](https://github.com/elastic/terraform-provider-elasticstack/pull/2911))
- Add `elasticstack_kibana_dashboard` resource ([#2902](https://github.com/elastic/terraform-provider-elasticstack/pull/2902))
- Add `elasticstack_elasticsearch_ml_filter` resource for managing Elasticsearch ML filters (used with anomaly detection `custom_rules`). ([#1970](https://github.com/elastic/terraform-provider-elasticstack/pull/1970))
- Improve docs on ML Anomaly job results_index_name ([#2919](https://github.com/elastic/terraform-provider-elasticstack/pull/2919))
- Add optional `scope` on detector `custom_rules` for ML anomaly detection jobs (map analysis field names to ML `filter_id` and optional `filter_type`). ([#2877](https://github.com/elastic/terraform-provider-elasticstack/pull/2877))
- Added plan-time validation for kibana_slo time_window.duration based on window type. ([#2914](https://github.com/elastic/terraform-provider-elasticstack/pull/2914))
- Remove top-level `enabled` field from `elasticstack_fleet_integration_policy`. ([#2773](https://github.com/elastic/terraform-provider-elasticstack/pull/2773))
- Adds sort nested block to elasticstack_elasticsearch_index resource with deprecation of sort_field/sort_order and seamless migration ([#2851](https://github.com/elastic/terraform-provider-elasticstack/pull/2851))
- Fix crash when role_descriptors is not set in elasticstack_elasticsearch_security_api_key ([#2855](https://github.com/elastic/terraform-provider-elasticstack/pull/2855))
- Migrate `elasticstack_elasticsearch_security_user` data source to Plugin Framework. ([#2854](https://github.com/elastic/terraform-provider-elasticstack/pull/2854))
- Migrate `elasticsearch_security_role` data source from Plugin SDK to Plugin Framework. ([#2847](https://github.com/elastic/terraform-provider-elasticstack/pull/2847))
- Migrate `elasticstack_elasticsearch_cluster_settings` to the Terraform plugin framework ([#2755](https://github.com/elastic/terraform-provider-elasticstack/pull/2755))
- Migrate elasticstack_elasticsearch_info data source to Plugin Framework ([#2796](https://github.com/elastic/terraform-provider-elasticstack/pull/2796))
- Migrate `elasticstack_elasticsearch_snapshot_lifecycle` and `elasticstack_elasticsearch_snapshot_repository` to the Plugin Framework. ([#2752](https://github.com/elastic/terraform-provider-elasticstack/pull/2752))
- Migrated the `elasticstack_elasticsearch_transform` resource to the Plugin Framework. ([#2757](https://github.com/elastic/terraform-provider-elasticstack/pull/2757))
- Migrate `elasticstack_elasticsearch_component_template` to the Terraform plugin framework ([#2749](https://github.com/elastic/terraform-provider-elasticstack/pull/2749))
- Migrated `elasticstack_elasticsearch_logstash_pipeline` resource to the Terraform plugin framework. ([#2750](https://github.com/elastic/terraform-provider-elasticstack/pull/2750))
- Migrate the `elasticstack_elasticsearch_snapshot_repository` data source to the Terraform plugin framework. ([#2761](https://github.com/elastic/terraform-provider-elasticstack/pull/2761))
- Store nil watch metadata as JSON null instead of empty object ([#2759](https://github.com/elastic/terraform-provider-elasticstack/pull/2759))
- Migrates `elasticstack_elasticsearch_ingest_pipeline` to the Terraform plugin framework ([#2745](https://github.com/elastic/terraform-provider-elasticstack/pull/2745))
- Fix ILM policy delete failures when the policy is still referenced by indices (e.g. Fleet-managed data stream backing indices) ([#2714](https://github.com/elastic/terraform-provider-elasticstack/pull/2714))
- Migrated `elasticstack_elasticsearch_data_stream` resource from Plugin SDK to Plugin Framework ([#2744](https://github.com/elastic/terraform-provider-elasticstack/pull/2744))
- Add elasticstack_apm_source_map resource for managing APM source maps via Kibana API ([#2712](https://github.com/elastic/terraform-provider-elasticstack/pull/2712))
- Fixed enrich policy resource recreation on every apply when `query` is not configured, by treating marshaled-null API responses as equivalent to an absent query. ([#2691](https://github.com/elastic/terraform-provider-elasticstack/pull/2691))
- Duplicate `actions` block `group` values are now rejected at plan time with a clear error instead of failing with an opaque HTTP 400 at apply time. ([#2656](https://github.com/elastic/terraform-provider-elasticstack/pull/2656))
- Poll for job closed state before deleting ML anomaly detection job to eliminate HTTP 409 version_conflict_engine_exception on teardown ([#2669](https://github.com/elastic/terraform-provider-elasticstack/pull/2669))
- Adds `elasticstack_fleet_proxy` resource for managing fleet proxies ([#2364](https://github.com/elastic/terraform-provider-elasticstack/pull/2364))
- `elasticstack_fleet_integration` now syncs `space_id` from Fleet on both create and read, preventing state drift that caused unexpected forced replacements. ([#2582](https://github.com/elastic/terraform-provider-elasticstack/pull/2582))
- Add space-aware Kibana asset management for elasticstack_fleet_integration on Kibana >= 8.15.0 ([#2608](https://github.com/elastic/terraform-provider-elasticstack/pull/2608))
- Internal migration of ingest processor data sources to Plugin Framework. Add missing common fields to geoip and user_agent processors. ([#2609](https://github.com/elastic/terraform-provider-elasticstack/pull/2609))
- Add optional `use_existing` on `elasticstack_elasticsearch_index` to adopt an existing index at create instead of failing on duplicate. ([#2589](https://github.com/elastic/terraform-provider-elasticstack/pull/2589))
- Fix plan-time params validation in `elasticstack_kibana_alerting_rule` for xpack.uptime.alerts.monitorStatus by using the correct generated struct and expanding legacy filter fields. ([#2573](https://github.com/elastic/terraform-provider-elasticstack/pull/2573))
- Fix perpetual plan diff for `indices_options.expand_wildcards = ["all"]` in ML datafeed resource ([#2572](https://github.com/elastic/terraform-provider-elasticstack/pull/2572))
- add tamper protection option to agent policy ressource ([#2086](https://github.com/elastic/terraform-provider-elasticstack/pull/2086))
- Drop Elastic Stack 7.x support floor. The provider now documents and tests against Elastic Stack 8.0+. ([#2554](https://github.com/elastic/terraform-provider-elasticstack/pull/2554))
- Fix perpetual plan drift on elasticstack_elasticsearch_index mappings when an index template injects additional mapping content. ([#2542](https://github.com/elastic/terraform-provider-elasticstack/pull/2542))
- Align Kibana SLO KQL schema and API mapping with object-form filters, settings, artifacts, and enabled state. ([#2495](https://github.com/elastic/terraform-provider-elasticstack/pull/2495))
- `elasticstack_kibana_space` now correctly clears `description`, `initials`, `color`, and `image_url` when the configuration sets them to an empty string. Previously those explicit empty-string assignments were silently dropped from the outbound API request and Kibana retained the prior value. ([#2452](https://github.com/elastic/terraform-provider-elasticstack/pull/2452))
- elasticstack_fleet_agent_policy no longer errors with "Provider produced inconsistent result" when the Fleet API returns an empty description for a policy whose description is unset in the Terraform configuration. ([#2448](https://github.com/elastic/terraform-provider-elasticstack/pull/2448))
- Migrate `elasticstack_elasticsearch_index_template` resource and data source to the Terraform Plugin Framework. Existing state is upgraded automatically (v0 → v1); attribute names, paths, block syntax, identity, and import behavior are preserved. ([#2515](https://github.com/elastic/terraform-provider-elasticstack/pull/2515))
- `elasticstack_fleet_integration_policy` can now be imported from non-default Kibana spaces using a composite `<space_id>/<policy_id>` import ID ([#2522](https://github.com/elastic/terraform-provider-elasticstack/pull/2522))
- Add `template.data_stream_options` block to `elasticstack_elasticsearch_index_template` to configure the failure store for new data streams via Terraform ([#2509](https://github.com/elastic/terraform-provider-elasticstack/pull/2509))
- elasticstack_fleet_integration now detects out-of-band package upgrades and downgrades during refresh by consulting InstallationInfo.Version; terraform plan surfaces the drift instead of reporting "No changes". ([#2447](https://github.com/elastic/terraform-provider-elasticstack/pull/2447))
- elasticstack_elasticsearch_security_role now detects out-of-band drift on description, metadata, and other attributes during refresh; terraform plan no longer silently returns "No changes" when a role is modified outside Terraform. ([#2446](https://github.com/elastic/terraform-provider-elasticstack/pull/2446))
- New `elasticstack_fleet_custom_integration` resource for uploading and managing locally-built Fleet integration packages via the EPM binary upload API ([#2387](https://github.com/elastic/terraform-provider-elasticstack/pull/2387))
- Change `elasticstack_kibana_security_detection_rule.actions[].params` to a JSON string rather than a map of string values. This allows setting arbitrary, nested param values ([#2340](https://github.com/elastic/terraform-provider-elasticstack/pull/2340))
- Add import support to the elasticstack_elasticsearch_enrich_policy resource ([#2427](https://github.com/elastic/terraform-provider-elasticstack/pull/2427))
- Add ssl.verification_mode attribute to the elasticstack_fleet_output ssl block ([#2415](https://github.com/elastic/terraform-provider-elasticstack/pull/2415))

## [0.14.5] - 2026-04-21

- Fix `elasticstack_kibana_slo.metric_custom_indicator` to support doc_count aggregation by making field optional and sending the no-field API variant ([#2394](https://github.com/elastic/terraform-provider-elasticstack/pull/2394))
- Fix "provider produced inconsistent result after apply" for SLO resources when objective target, timeslice target, or histogram range from/to values are not exactly representable in float32 ([#2401](https://github.com/elastic/terraform-provider-elasticstack/pull/2401))
- Add `elasticstack_fleet_agent_download_source` resource ([#2081](https://github.com/elastic/terraform-provider-elasticstack/pull/2081))

## [0.14.4] - 2026-04-20

### Changes

- Add `elasticstack_kibana_agentbuilder_agent` resource and data source ([#2295](https://github.com/elastic/terraform-provider-elasticstack/pull/2295))
- Add `create_new_copies` and `compatibility_mode` attributes to `elasticstack_kibana_import_saved_objects` ([#2289](https://github.com/elastic/terraform-provider-elasticstack/pull/2289))
- Fix `elasticstack_elasticsearch_watch` updates when Watcher redacts nested action secrets on refresh. ([#2296](https://github.com/elastic/terraform-provider-elasticstack/pull/2296))
- Migrated `elasticstack_elasticsearch_watch` to the Terraform Plugin Framework ([#2287](https://github.com/elastic/terraform-provider-elasticstack/pull/2287))
- Add `elasticstack_kibana_agentbuilder_tool` resource and data source ([#2111](https://github.com/elastic/terraform-provider-elasticstack/pull/2111))
- Fix state consistency with semantic text types in `elasticstack_elasticsearch_index` ([#2112](https://github.com/elastic/terraform-provider-elasticstack/pull/2112))
- Add `elasticstack_elasticsearch_inference_endpoint` resource. ([#1955](https://github.com/elastic/terraform-provider-elasticstack/pull/1955))
- `elasticstack_kibana_data_view`: Support in-place updates of the `namespaces` field via the Kibana Spaces API (`POST /api/spaces/_update_objects_spaces`), preventing data view recreation when sharing across spaces. Previously, any change to `namespaces` would force replacement of the resource, breaking dependent alerting rules that reference the data view by ID. ([#2129](https://github.com/elastic/terraform-provider-elasticstack/pull/2129))
- Add `elasticstack_kibana_agentbuilder_workflow` resource and data source ([#1923](https://github.com/elastic/terraform-provider-elasticstack/pull/1923))
- Add `elasticstack_fleet_output` data source. ([#1762](https://github.com/elastic/terraform-provider-elasticstack/pull/1762))
- Fix `termField` validation for ESQL `.es-query` alert rules in `elasticstack_kibana_alerting_rule`. ([#1914](https://github.com/elastic/terraform-provider-elasticstack/pull/1914))
- Fix perpetual diff in `elasticstack_elasticsearch_index_template` if `search_routing` or `index_routing` was unset but `routing` was set ([#1841](https://github.com/elastic/terraform-provider-elasticstack/pull/1841))
- Fix provider panic in `elasticstack_fleet_integration_policy` when the integration version is no longer available in the package registry. ([#1913](https://github.com/elastic/terraform-provider-elasticstack/pull/1913))
- Add `elasticstack_elasticsearch_ingest_processor_inference` data source ([#1956](https://github.com/elastic/terraform-provider-elasticstack/pull/1956))
- Add an experimental flag to skip synthetics location validation. ([#1924](https://github.com/elastic/terraform-provider-elasticstack/pull/1924))
- Add flapping detection to `elasticstack_kibana_alerting_rule`. ([#1966](https://github.com/elastic/terraform-provider-elasticstack/pull/1966))
- Attempt recovery when data view creation fails ([#2024](https://github.com/elastic/terraform-provider-elasticstack/pull/2024))
- Fix several "Provider produced inconsistent result after apply" errors in the anomaly detection job resource. ([#2034](https://github.com/elastic/terraform-provider-elasticstack/pull/2034))
- Remove deprecation warning on the `elasticsearch_connection` attribute provider wide ([#2100](https://github.com/elastic/terraform-provider-elasticstack/pull/2100))
- Migrate `elasticstack_elasticsearch_index_lifecycle` to the Terraform Plugin Framework ([#2002](https://github.com/elastic/terraform-provider-elasticstack/pull/2002))
- Add `space_id` to `elasticstack_kibana_synthetics_private_location` ([#2142](https://github.com/elastic/terraform-provider-elasticstack/pull/2142))
- Use space-scoped endpoints for Fleet resources to allow space-restricted roles to properly manage Fleet resources. ([#2084](https://github.com/elastic/terraform-provider-elasticstack/pull/2084))
- Fix perpetual `id` "known after apply" diff in `elasticstack_elasticsearch_security_role` by adding `UseStateForUnknown` plan modifier to the computed `id` attribute. ([#2160](https://github.com/elastic/terraform-provider-elasticstack/pull/2160))
- Add `elasticstack_fleet_elastic_defend_integration_policy` resource ([#2147](https://github.com/elastic/terraform-provider-elasticstack/pull/2147))

## [0.14.3] - 2026-03-02

### Changes

- Add `elasticstack_fleet_output` data source. ([#1762](https://github.com/elastic/terraform-provider-elasticstack/pull/1762))
- Add `elasticstack_elasticsearch_index_template_ilm_attachment` resource to attach ILM policies to Fleet-managed or externally-managed index templates via the `@custom` component template. ([#1641](https://github.com/elastic/terraform-provider-elasticstack/pull/1641))
- Fix `elasticstack_kibana_slo` `timeslice_metric_indicator` to support `last_value`, `cardinality`, and `std_deviation` aggregations which are valid in the Kibana SLO API but were previously rejected by the provider. ([#1749](https://github.com/elastic/terraform-provider-elasticstack/pull/1749))
- Add `elasticstack_kibana_security_enable_rule` resource ([#1710](https://github.com/elastic/terraform-provider-elasticstack/pull/1710))
- Fix value conversion error in `elasticstack_elasticsearch_index_alias` when indices are unknown at plan time. ([#1755](https://github.com/elastic/terraform-provider-elasticstack/pull/1755))
- Fix state consistency error in `elasticstack_kibana_security_exception_list` when `os_types` are used in Elastic Stack 9.2 ([#1740](https://github.com/elastic/terraform-provider-elasticstack/pull/1740))
- Fix state consistency error in `elasticstack_elasticsearch_security_role` when `description` is empty (`""`) ([#1780](https://github.com/elastic/terraform-provider-elasticstack/pull/1780))
- Fix state consistency issue in `elasticstack_kibana_slo` when `group_by` is set to an empty list ([#1776](https://github.com/elastic/terraform-provider-elasticstack/pull/1776))
- Fix state consistency error in `elasticstack_kibana_security_detection_rule` when `threat_filter` is supplied ([#1758](https://github.com/elastic/terraform-provider-elasticstack/pull/1758))
- Fix state consistency error in `elasticstack_fleet_integration_policy` when the policy was updated outside of the Terraform workflow ([#1616](https://github.com/elastic/terraform-provider-elasticstack/pull/1616))

## [0.14.2] - 2026-02-19

### Changes

<<<<<<< HEAD

=======
>>>>>>> 7346053b2 (docs(changelog): add entity store timing/isolation fix entry (#4062))

- Add parameter validation and default normalization for `elasticstack_kibana_alerting_rule` to prevent inconsistent state errors caused by API-injected defaults. ([#1648](https://github.com/elastic/terraform-provider-elasticstack/pull/1648))
- Fix JSON marshaling error in `elasticstack_kibana_slo` when `good` or `total` fields in `kql_custom_indicator` are empty or null. ([#1729](https://github.com/elastic/terraform-provider-elasticstack/pull/1729))

## [0.14.1] - 2026-02-18

### Changes

<<<<<<< HEAD

=======
>>>>>>> 7346053b2 (docs(changelog): add entity store timing/isolation fix entry (#4062))

- Fix provider panic in `elasticstack_kibana_slo` when SLO updates error without a HTTP response. ([#1725](https://github.com/elastic/terraform-provider-elasticstack/pull/1725))
- Fix inconsistent state error in `elasticstack_kibana_alerting_rule` when `alert_delay` is not specified. ([#1726](https://github.com/elastic/terraform-provider-elasticstack/pull/1726))

## [0.14.0] - 2026-02-16

### Breaking changes

#### `elasticstack_fleet_integration` `space_ids` attribute has been reduced to a single `space_id`

The provider was only considering the first entry in the `space_ids` set ([#1642](https://github.com/elastic/terraform-provider-elasticstack/issues/1642)). Extending the resource to correctly handle multiple spaces would not make sense as a single Terraform resource. Instead this attribute is being reduced to a single string, with practitioners able to manage the installation of an integration across multiple spaces through multiple instances of this resource.

Existing usage of the `space_ids` attribute must be migrated to `space_id`:

```hcl
resource "elasticstack_fleet_integration" "tcp" {
  name = "tcp"
  version = "1.16.0"
  space_ids = ["default", "o11y"]
}
```

becomes:

```hcl
resource "elasticstack_fleet_integration" "tcp-default" {
  name = "tcp"
  version = "1.16.0"
  space_id = "default"
}

resource "elasticstack_fleet_integration" "tcp-o11y" {
  name = "tcp"
  version = "1.16.0"
  space_id = "o11y"
}
```

#### `elasticstack_fleet_integration_policy` input block has changed to a map attribute

The `input` block in the `elasticstack_fleet_integration_policy` resource has been restructured into the `inputs` map attribute.
This transition:
<<<<<<< HEAD

=======
>>>>>>> 7346053b2 (docs(changelog): add entity store timing/isolation fix entry (#4062))

- Allows the provider to implement semantic equality checking across all inputs within the integration policy. This change:
  - Prevents several state consistency errors experienced whilst using this resource
  - Allow practitioners to only define configuration for the inputs, streams, and variables that differ from the package defined defaults.
- Reduces the scope of the large `streams_json` string. Instead allowing each stream to be defined as it's own object for Terraform drift checking.

Existing usage of the `input` block must be migrated to the attribute syntax. Some [examples of this migration](https://github.com/elastic/terraform-provider-elasticstack/pull/1482/files) can be seen in the changes to the provider automated tests. As a step-by-step guide however:

1. `input` blocks are merged together into a single `inputs` attribute
2. The `input_id` attribute is removed, and instead used as the map key when defining an input
3. `streams_json` is removed, with the contents becoming a `streams` map attribute

Combined, this looks like:

```hcl
input {
  input_id = "tcp-tcp"
  enabled  = false
  streams_json = jsonencode({
    "tcp.generic" : {
      "enabled" : false
      "vars" : {
        "listen_address" : "localhost"
        "listen_port" : 8085
        "data_stream.dataset" : "tcp.generic"
        "tags" : []
        "syslog_options" : "field: message"
        "ssl" : ""
        "custom" : ""
      }
    }
  })
}
```

becoming

```hcl
inputs = {
  "tcp-tcp" = {
    enabled = false
    streams = {
      "tcp.generic" = {
        enabled = false
        vars = jsonencode({
          "listen_address" : "localhost"
          "listen_port" : 8085
          "data_stream.dataset" : "tcp.generic"
          "tags" : []
          "syslog_options" : "field: message"
          "ssl" : ""
          "custom" : ""
        })
      }
    }
  }
}
```

### Changes

- Add import support for `elasticstack_elasticsearch_script` resource ([#1637](https://github.com/elastic/terraform-provider-elasticstack/pull/1637))
- Migrate `elasticstack_kibana_alerting_rule` to use plugin framework ([#1664](https://github.com/elastic/terraform-provider-elasticstack/pull/1664))
- Migrate `elasticstack_kibana_slo` resource to the Terraform plugin framework ([#1647](https://github.com/elastic/terraform-provider-elasticstack/pull/1647))
- Prevent a provider error with `elasticstack_fleet_integration_policy` when moving between a single `policy_id` and multiple `policy_ids` ([#1644](https://github.com/elastic/terraform-provider-elasticstack/pull/1644))
- Fix concurrent map write errors with `elasticstack_fleet_integration_policy` ([#1629](https://github.com/elastic/terraform-provider-elasticstack/pull/1629))
- Add support for Fleet API installation parameters to `elasticstack_fleet_integration` resource: `prerelease`, `ignore_mapping_update_errors` (8.11.0+), `skip_data_stream_rollover` (8.11.0+), and `ignore_constraints`. These parameters provide full control over package installation behavior and enable installation of prerelease (beta, non-GA) packages.
- Correctly handle 404 responses when reading `elasticstack_fleet_integration` resources ([#1608](https://github.com/elastic/terraform-provider-elasticstack/pull/1608))
- Fix handling custom `policy_id` attributes in `elasticstack_fleet_integration_policy` resources ([#1594](https://github.com/elastic/terraform-provider-elasticstack/pull/1594))
- Add `advanced_settings` to `elasticstack_fleet_agent_policy` to configure agent logging, CPU limits, and download settings ([#1545](https://github.com/elastic/terraform-provider-elasticstack/pull/1545))
- Prevent provider panic when importing a non-existant `elasticstack_elasticsearch_ml_datafeed`. ([#1579](https://github.com/elastic/terraform-provider-elasticstack/pull/1579))
- Fix handling of empty `except` attributes in `elasticstack_elasticsearch_security_role` ([#1581](https://github.com/elastic/terraform-provider-elasticstack/pull/1581))
- Fix the enabled property being ignored in `elasticstack_kibana_alerting_rule` ([#1527](https://github.com/elastic/terraform-provider-elasticstack/pull/1527))
- Add `advanced_monitoring_options` to `elasticstack_fleet_agent_policy` to configure HTTP monitoring endpoint and diagnostics settings ([#1537](https://github.com/elastic/terraform-provider-elasticstack/pull/1537))
- Move the `input` block to an `inputs` map in `elasticstack_fleet_integration_policy` ([#1482](https://github.com/elastic/terraform-provider-elasticstack/pull/1482))
- Fix `elasticstack_elasticsearch_ml_anomaly_detection_job` import to be resilient to sparse state values
- Fix a state consistency issue when an `elasticstack_elasticsearch_ml_datafeed_state` resource without `start` configured is started after being stopped. ([#1563](https://github.com/elastic/terraform-provider-elasticstack/pull/1563))
- Fix a state consistency issue when `elasticstack_elasticsearch_ml_datafeed_state` `start` and `end` times are specified in a timezone that is not the server timezone `elasticstack_elasticsearch_ml_datafeed_state` resource without `start` configured is started after being stopped. ([#1563](https://github.com/elastic/terraform-provider-elasticstack/pull/1563))
- Fix an issue where `elasticstack_elasticsearch_ml_datafeed_state` `start` and `end` times where treated by the provider as unix seconds, but by the API as unix milliseconds.
- Only require input parameters in `elasticstack_fleet_integration_policy` to be specified if they differ from integration defaults ([#1558](https://github.com/elastic/terraform-provider-elasticstack/pull/1558))
- Only require vars in `elasticstack_fleet_integration_policy` to be specified if they differ from integration defaults ([#1593](https://github.com/elastic/terraform-provider-elasticstack/pull/1593))
- Allow space restricted roles to manage `elasticstack_fleet_agent_policy` resources. ([#1597](https://github.com/elastic/terraform-provider-elasticstack/pull/1597))
- Fix missing timeslice's metric-scoped `filter` parameter for doc_count aggregations ([#1636](https://github.com/elastic/terraform-provider-elasticstack/pull/1636))
- Collapse `space_ids` to a single `space_id` in `elasticstack_fleet_integration` ([#1645](https://github.com/elastic/terraform-provider-elasticstack/pull/1645))
- Add `bearer_token` authentication support to Kibana and Fleet provider configurations. Bearer tokens configured in the `elasticsearch` block are now propagated to `kibana` and `fleet` blocks as fallback credentials, consistent with the existing behavior for `username`, `password`, and `api_key`. New environment variables `KIBANA_BEARER_TOKEN` and `FLEET_BEARER_TOKEN` are also supported. ([#1690](https://github.com/elastic/terraform-provider-elasticstack/pull/1690))

## [0.13.1] - 2025-12-12

- Fix handling empty types in `elasticstack_elasticsearch_ml_anomaly_detection_job` ([#1544](https://github.com/elastic/terraform-provider-elasticstack/pull/1544))
- Fix handling empty `clusters` and `run_as` attributes in `elasticstack_elasticsearch_security_role` resource ([#1542](https://github.com/elastic/terraform-provider-elasticstack/pull/1542))

## [0.13.0] - 2025-12-10

### Breaking changes

#### `elasticstack_elasticsearch_index.alias` block has changed to a set attribute

The `alias` block in the `elasticstack_elasticsearch_index` resource has been moved to an attribute.
This transition provides better support for future changes in both the provider and the underlying Terraform framework.

Existing usage of the `alias` block must be migrated to the attribute syntax. For example:

```hcl
alias {
  name = "my_alias_1"
}

alias {
  name = "my_alias_2"
  filter = jsonencode({
    term = { "user.id" = "developer" }
  })
}
```

becomes

```hcl
alias = [
  {
    name = "my_alias_1"
  },
  {
    name = "my_alias_2"
    filter = jsonencode({
      term = { "user.id" = "developer" }
    })
  }
]
```

### Changes

- Fix `elasticstack_kibana_action_connector` failing with "inconsistent result after apply" when config contains null values ([#1524](https://github.com/elastic/terraform-provider-elasticstack/pull/1524))
- Add `host_name_format` to `elasticstack_fleet_agent_policy` to configure host name format (hostname or FQDN) ([#1312](https://github.com/elastic/terraform-provider-elasticstack/pull/1312))
- Create `elasticstack_kibana_prebuilt_rule` resource ([#1296](https://github.com/elastic/terraform-provider-elasticstack/pull/1296))
- Add `required_versions` to `elasticstack_fleet_agent_policy` ([#1436](https://github.com/elastic/terraform-provider-elasticstack/pull/1436))
- Migrate `elasticstack_elasticsearch_security_role` resource to Terraform Plugin Framework ([#1330](https://github.com/elastic/terraform-provider-elasticstack/pull/1330))
- Fix an issue where the `elasticstack_fleet_output` resource would error due to inconsistent state after an ouptut was edited in the Kibana UI ([#1506](https://github.com/elastic/terraform-provider-elasticstack/pull/1506))
- Allow `index` and `data_view_id` values to both be unknown during planning in `elasticstack_kibana_security_detection_rule` ([#1499](https://github.com/elastic/terraform-provider-elasticstack/pull/1499))
- Support `.bedrock` and `.gen-ai` connectors ([#1467](https://github.com/elastic/terraform-provider-elasticstack/pull/1467))
- Support the `solution` attribute in `elasticstack_kibana_space` from 8.16 ([#1486](https://github.com/elastic/terraform-provider-elasticstack/pull/1486))
- Add `elasticstack_elasticsearch_alias` resource ([#1343](https://github.com/elastic/terraform-provider-elasticstack/pull/1343))
- Add `mapping_total_fields_limit` to `elasticstack_elasticsearch_index` ([#1494](https://github.com/elastic/terraform-provider-elasticstack/pull/1494))
- Add `elasticstack_kibana_default_data_view` resource ([#1379](https://github.com/elastic/terraform-provider-elasticstack/pull/1379))
- Add support for [Security Exceptions](https://github.com/elastic/terraform-provider-elasticstack/issues/1332)
  - Add `elasticstack_kibana_security_exception_item` resource ([#1496](https://github.com/elastic/terraform-provider-elasticstack/pull/1496))
  - Add `elasticstack_kibana_security_exception_list` resource ([#1495](https://github.com/elastic/terraform-provider-elasticstack/pull/1495))
  - Add `elasticstack_kibana_security_list` resource ([#1489](https://github.com/elastic/terraform-provider-elasticstack/pull/1489))
  - Add `elasticstack_kibana_security_list_item` resource ([#1492](https://github.com/elastic/terraform-provider-elasticstack/pull/1492))
  - Add `elasticstack_kibana_security_list_data_streams` resource ([#1525](https://github.com/elastic/terraform-provider-elasticstack/pull/1525))

## [0.12.2] - 2025-11-19

- Fix `elasticstack_elasticsearch_snapshot_lifecycle` metadata type conversion causing terraform apply to fail ([#1409](https://github.com/elastic/terraform-provider-elasticstack/issues/1409))
- Add new `elasticstack_elasticsearch_ml_anomaly_detection_job` resource ([#1329](https://github.com/elastic/terraform-provider-elasticstack/pull/1329))
- Add new `elasticstack_elasticsearch_ml_datafeed` resource ([#1340](https://github.com/elastic/terraform-provider-elasticstack/pull/1340))
- Add `space_ids` attribute to all Fleet resources to support space-aware Fleet resource management ([#1390](https://github.com/elastic/terraform-provider-elasticstack/pull/1390))
- Add back missing import support for `elasticstack_elasticsearch_security_role_mapping` ([#1441](https://github.com/elastic/terraform-provider-elasticstack/pull/1441))
- Add new `elasticstack_elasticsearch_ml_job_state` resource ([#1337](https://github.com/elastic/terraform-provider-elasticstack/pull/1337))
- Add new `elasticstack_elasticsearch_ml_datafeed_state` resource ([#1422](https://github.com/elastic/terraform-provider-elasticstack/pull/1422))
- Add `output_id` to `elasticstack_fleet_integration_policy` resource ([#1445](https://github.com/elastic/terraform-provider-elasticstack/pull/1445))
- Make `hosts` attribute required in `elasticstack_fleet_output` resource ([#1450](https://github.com/elastic/terraform-provider-elasticstack/pull/1450/files))
- Fix `elasticstack_kibana_security_detection_rule` to properly respect `space_id`

## [0.12.1] - 2025-10-22

- Fix regression restricting the characters in an `elasticstack_elasticsearch_role_mapping` `name`. ([#1373](https://github.com/elastic/terraform-provider-elasticstack/pull/1373))
- Add schema validations to require either (but not both) `index` and `data_view_id` is set for relevant Security Detection Rules ([#1381](https://github.com/elastic/terraform-provider-elasticstack/pull/1381))

## [0.12.0] - 2025-10-15

- Fix provider crash with `elasticstack_kibana_action_connector` when `config` or `secrets` was unset in 0.11.17 ([#1355](https://github.com/elastic/terraform-provider-elasticstack/pull/1355))
- Added `labels` field to `elasticstack_kibana_synthetics_monitor` resource for associating key-value pairs with monitors ([#1360](https://github.com/elastic/terraform-provider-elasticstack/pull/1360))
- Fixes provider crash with `elasticstack_kibana_slo` when using `kql_custom_indicator` with no `filter` set. ([#1354](https://github.com/elastic/terraform-provider-elasticstack/pull/1354))
- Updates for Security Detection Rules ([#1361](https://github.com/elastic/terraform-provider-elasticstack/pull/1361)
  - Add support for `threat` property
  - Gracefully support `query` property not being set
  - Add esql specific validations to reject unsupported fields `index` and `filters`
  - Gracefully handle response action with no provided `frequency`
  - Add validation for required `anomaly_threshold` field in anomaly detection rules
  - Add support for `timeline_id` / `timeline_title` fields
  - Gracefully handle `threat_query` not being provided for `threat_match` rule

## [0.11.19] - 2025-10-22

Version 0.11.19 is equivalent to 0.12.1. It is being released to help mitigate impact from 0.11.18 being inadvertently released ahead of schedule. This version contained a breaking change and defects related to internal refactors. While 0.11.19 still contains a breaking change from 0.11.17 it does fix defects (see details below) for any users relying on the latest 0.11.x version.

- Fix regression restricting the characters in an `elasticstack_elasticsearch_role_mapping` `name`. ([#1373](https://github.com/elastic/terraform-provider-elasticstack/pull/1373))
- Add schema validations to require either (but not both) `index` and `data_view_id` is set for relevant Security Detection Rules ([#1381](https://github.com/elastic/terraform-provider-elasticstack/pull/1381))
- Fix provider crash with `elasticstack_kibana_action_connector` when `config` or `secrets` was unset in 0.11.17 ([#1355](https://github.com/elastic/terraform-provider-elasticstack/pull/1355))
- Added `labels` field to `elasticstack_kibana_synthetics_monitor` resource for associating key-value pairs with monitors ([#1360](https://github.com/elastic/terraform-provider-elasticstack/pull/1360))
- Fixes provider crash with `elasticstack_kibana_slo` when using `kql_custom_indicator` with no `filter` set. ([#1354](https://github.com/elastic/terraform-provider-elasticstack/pull/1354))
- Updates for Security Detection Rules ([#1361](https://github.com/elastic/terraform-provider-elasticstack/pull/1361)
  - Add support for `threat` property
  - Gracefully support `query` property not being set
  - Add esql specific validations to reject unsupported fields `index` and `filters`
  - Gracefully handle response action with no provided `frequency`
  - Add validation for required `anomaly_threshold` field in anomaly detection rules
  - Add support for `timeline_id` / `timeline_title` fields
  - Gracefully handle `threat_query` not being provided for `threat_match` rule

## [0.11.18] - 2025-10-10

### Breaking changes

The `ssl` field on the `elasticstack_fleet_output` resource has been changes from a block to an attribute. This change ensures ongoing consistency within the resource schema for this resource, and aligns with Terraform best practices.

Existing `elasticstack_fleet_output` resources defining `ssl` will have to update the declaration to an attribute style. For example:

```hcl
resource "elasticstack_fleet_output" "output" {
  ...
  ssl {
    ...
  }
}
```

becomes

```hcl
resource "elasticstack_fleet_output" "output" {
  ...
  ssl = {  # Note the equals sign here. 
    ...
  }
}
```

### Changes

- Create `elasticstack_kibana_security_detection_rule` resource. ([#1290](https://github.com/elastic/terraform-provider-elasticstack/pull/1290))
- Add `elasticstack_kibana_export_saved_objects` data source ([#1293](https://github.com/elastic/terraform-provider-elasticstack/pull/1293))
- Create `elasticstack_kibana_maintenance_window` resource. ([#1224](https://github.com/elastic/terraform-provider-elasticstack/pull/1224))
- Add support for `solution` field in `elasticstack_kibana_space` resource and data source ([#1102](https://github.com/elastic/terraform-provider-elasticstack/issues/1102))
- Add `slo_id` validation to `elasticstack_kibana_slo` ([#1221](https://github.com/elastic/terraform-provider-elasticstack/pull/1221))
- Add `ignore_missing_component_templates` to `elasticstack_elasticsearch_index_template` ([#1206](https://github.com/elastic/terraform-provider-elasticstack/pull/1206))
- Migrate `elasticstack_elasticsearch_enrich_policy` resource and data source to Terraform Plugin Framework ([#1220](https://github.com/elastic/terraform-provider-elasticstack/pull/1220))
- Prevent provider panic when a script exists in state, but not in Elasticsearch ([#1218](https://github.com/elastic/terraform-provider-elasticstack/pull/1218))
- Add support for managing cross_cluster API keys in `elasticstack_elasticsearch_security_api_key` ([#1252](https://github.com/elastic/terraform-provider-elasticstack/pull/1252))
- Allow version changes without a destroy/create cycle with `elasticstack_fleet_integration` ([#1255](https://github.com/elastic/terraform-provider-elasticstack/pull/1255)). This fixes an issue where it was impossible to upgrade integrations which are used by an integration policy.
- Add `namespace` attribute to `elasticstack_kibana_synthetics_monitor` resource to support setting data stream namespace independently from `space_id` ([#1247](https://github.com/elastic/terraform-provider-elasticstack/pull/1247))
- Support setting an explit `connector_id` in `elasticstack_kibana_action_connector`. This attribute already existed, but was being ignored by the provider. Setting the attribute will return an error in Elastic Stack v8.8 and lower since creating a connector with an explicit ID is not supported. ([1260](https://github.com/elastic/terraform-provider-elasticstack/pull/1260))
- Migrate `elasticstack_kibana_action_connector` to the Terraform plugin framework ([#1269](https://github.com/elastic/terraform-provider-elasticstack/pull/1269))
- Migrate `elasticstack_elasticsearch_security_role_mapping` resource and data source to Terraform Plugin Framework ([#1279](https://github.com/elastic/terraform-provider-elasticstack/pull/1279))
- Add support for `inactivity_timeout` in `elasticstack_fleet_agent_policy` ([#641](https://github.com/elastic/terraform-provider-elasticstack/issues/641))
- Migrate `elasticstack_elasticsearch_script` resource to Terraform Plugin Framework ([#1297](https://github.com/elastic/terraform-provider-elasticstack/pull/1297))
- Add support for `kafka` output types in `elasticstack_fleet_output` ([#1302](https://github.com/elastic/terraform-provider-elasticstack/pull/1302))
- Add support for `prevent_initial_backfill` to `elasticstack_kibana_slo` ([#1071](https://github.com/elastic/terraform-provider-elasticstack/pull/1071))
- [Refactor] Regenerate the SLO client using the current OpenAPI spec ([#1303](https://github.com/elastic/terraform-provider-elasticstack/pull/1303))
- Add support for `data_view_id` in the `elasticstack_kibana_slo` resource ([#1305](https://github.com/elastic/terraform-provider-elasticstack/pull/1305))
- Add support for `unenrollment_timeout` in `elasticstack_fleet_agent_policy` ([#1169](https://github.com/elastic/terraform-provider-elasticstack/issues/1169))
- Handle default value for `allow_restricted_indices` in `elasticstack_elasticsearch_security_api_key` ([#1315](https://github.com/elastic/terraform-provider-elasticstack/pull/1315))
- Fixed `nil` reference in kibana synthetics API client in case of response errors ([#1320](https://github.com/elastic/terraform-provider-elasticstack/pull/1320))
- Add support for `agent_policy_ids` in `elasticstack_fleet_integration_policy` ([#1131](https://github.com/elastic/terraform-provider-elasticstack/pull/1311))

## [0.11.17] - 2025-07-21

- Add `elasticstack_apm_agent_configuration` resource ([#1196](https://github.com/elastic/terraform-provider-elasticstack/pull/1196))
- Add support for `timeslice_metric_indicator` in `elasticstack_kibana_slo` ([#1195](https://github.com/elastic/terraform-provider-elasticstack/pull/1195))
- Add `elasticstack_elasticsearch_ingest_processor_reroute` data source ([#678](https://github.com/elastic/terraform-provider-elasticstack/issues/678))
- Add support for `supports_agentless` to `elasticstack_fleet_agent_policy` ([#1197](https://github.com/elastic/terraform-provider-elasticstack/pull/1197))
- Ignore `master_timeout` when targeting Serverless projects ([#1207](https://github.com/elastic/terraform-provider-elasticstack/pull/1207))

## [0.11.16] - 2025-07-09

- Add `headers` for the provider connection ([#1057](https://github.com/elastic/terraform-provider-elasticstack/pull/1057))
- Migrate `elasticstack_elasticsearch_system_user` resource to Terraform plugin framework ([#1154](https://github.com/elastic/terraform-provider-elasticstack/pull/1154))
- Add custom `endpoint` configuration support for snapshot repository setup ([#1158](https://github.com/elastic/terraform-provider-elasticstack/pull/1158))
- Add `description` to `elasticstack_kibana_security_role` ([#1172](https://github.com/elastic/terraform-provider-elasticstack/issues/1172))
- Add `elasticstack_kibana_synthetics_parameter` resource ([#1155](https://github.com/elastic/terraform-provider-elasticstack/pull/1155))

## [0.11.15] - 2025-04-23

- Add `global_data_tags` to fleet agent policies. ([#1044](https://github.com/elastic/terraform-provider-elasticstack/pull/1044))

## [0.11.14] - 2025-03-17

- Fix a provider crash when interacting with elasticstack_kibana_data_view resources created with 0.11.0. ([#979](https://github.com/elastic/terraform-provider-elasticstack/pull/979))
- Add `max_primary_shard_docs` condition to ILM rollover ([#845](https://github.com/elastic/terraform-provider-elasticstack/pull/845))
- Add missing entries to `data_view.field_formats.params` ([#1001](https://github.com/elastic/terraform-provider-elasticstack/pull/1001))
- Fix namespaces inconsistency when creating elasticstack_kibana_data_view resources ([#1011](https://github.com/elastic/terraform-provider-elasticstack/pull/1011))
- Update rule ID documentation. ([#1047](https://github.com/elastic/terraform-provider-elasticstack/pull/1047))
- Mark `elasticstack_kibana_action_connector.secrets` as sensitive. ([#1045](https://github.com/elastic/terraform-provider-elasticstack/pull/1045))

## [0.11.13] - 2025-01-09

- Support 8.15.5 in acc tests ([#963](https://github.com/elastic/terraform-provider-elasticstack/pull/963)).
- Support 8.16.2 in acc tests ([#964](https://github.com/elastic/terraform-provider-elasticstack/pull/964)).
- Support 8.17.0 in acc tests ([#969](https://github.com/elastic/terraform-provider-elasticstack/pull/969)).
- Support 9.0.0 in acc tests ([#954](https://github.com/elastic/terraform-provider-elasticstack/pull/954)).
- Support several ssl fields in `elasticstack_kibana_synthetics_monitor` ([#967](https://github.com/elastic/terraform-provider-elasticstack/pull/967)).
- HTTP 400 Bad Request When Creating elasticstack_kibana_security_role ([933](https://github.com/elastic/terraform-provider-elasticstack/issues/933)).

## [0.11.12] - 2024-12-16

### Breaking changes

- Support multiple group by fields in SLOs ([#870](https://github.com/elastic/terraform-provider-elasticstack/pull/878)). This changes to type of the `group_by` attribute of the `elasticstack_kibana_slo` resource from a String to a list of Strings. Any existing SLO defintions will need to update `group_by = "field"` to `group_by = ["field"]`.

### Changes

- Handle NPE in integration policy secrets ([#946](https://github.com/elastic/terraform-provider-elasticstack/pull/946))
- Use the auto-generated OAS schema from elastic/kibana for the Fleet API. ([#834](https://github.com/elastic/terraform-provider-elasticstack/issues/834))
- Support description in `elasticstack_elasticsearch_security_role` data sources. ([#884](https://github.com/elastic/terraform-provider-elasticstack/pull/884))
- Prevent spurious recreation of `elasticstack_fleet_agent_policy` resources due to 'changing' policy ids ([#885](https://github.com/elastic/terraform-provider-elasticstack/pull/885))
- Support `elasticstack_kibana_alerting_rule` resources with only one of `kql` or `timeframe` attributes set ([#886](https://github.com/elastic/terraform-provider-elasticstack/pull/886))
- Rename generated/fleet to generated/kibana, add data_view APIs. Keep libs/go-kibana-rest until migration can be completed. Clean and simplify the `elasticstack_kibana_data_view` resource to match the styling of Fleet resources. ([#881](https://github.com/elastic/terraform-provider-elasticstack/issues/881))
- Exposes internal objects needed to build a Crossplane Elasticstack provider ([#949](https://github.com/elastic/terraform-provider-elasticstack/pull/949))

## [0.11.11] - 2024-10-25

- Allow `elasticstack_kibana_alerting_rule` to be used without Elasticsearch being configured. ([#869](https://github.com/elastic/terraform-provider-elasticstack/pull/869))
- Add resource `elasticstack_elasticsearch_data_stream_lifecycle` ([#838](https://github.com/elastic/terraform-provider-elasticstack/issues/838))
- Ensure API keys are not replaced when upgrading from 0.11.9 or earlier. ([#875](https://github.com/elastic/terraform-provider-elasticstack/pull/875))

## [0.11.10] - 2024-10-23

- Fix bug updating alert delay ([#859](https://github.com/elastic/terraform-provider-elasticstack/pull/859))
- Support updating `elasticstack_elasticsearch_security_api_key` when supported by the backing cluster ([#843](https://github.com/elastic/terraform-provider-elasticstack/pull/843))
- Fix validation of `throttle`, and `interval` attributes in `elasticstack_kibana_alerting_rule` allowing all Elastic duration values ([#846](https://github.com/elastic/terraform-provider-elasticstack/pull/846))
- Fix boolean setting parsing for `elasticstack_elasticsearch_indices` data source. ([#842](https://github.com/elastic/terraform-provider-elasticstack/pull/842))
- Update all Fleet and utils/tfsdk instances of diagnostics parameters to pass by pointer instead of pass by value. Added upgrader for fleet_integration_policy v0 to handle empty string vars_json/streams_json. ([#855](https://github.com/elastic/terraform-provider-elasticstack/pull/855))
- Fix handling of EPM packages when uninstalled outside Terraform, and diags in create/update. ([#854](https://github.com/elastic/terraform-provider-elasticstack/pull/854))

## [0.11.9] - 2024-10-14

### Breaking changes

- Remove support for specifying `include_type_name` from the `elasticstack_elasticsearch_index` resource. This parameter has been deprecated from 7.0, with indices restricted to a single type since 6.0. ([#832](https://github.com/elastic/terraform-provider-elasticstack/pull/832))

### Changes

- Fix inconsistent output errors in `elasticstack_fleet_output` for `default_integrations` and `default_monitoring`. ([#841](https://github.com/elastic/terraform-provider-elasticstack/pull/841))
- Fix secret handling `elasticstack_fleet_integration_policy` resource. ([#821](https://github.com/elastic/terraform-provider-elasticstack/pull/821))
- Fix merge values for `elasticstack_kibana_synthetics_monitor` monitor locations ([#823](https://github.com/elastic/terraform-provider-elasticstack/pull/823))
- Migrate to a v8 Elasticsearch client ([#832](https://github.com/elastic/terraform-provider-elasticstack/pull/832))
- Add support for the `.gemini` connector type for Kibana action connectors ([#819](https://github.com/elastic/terraform-provider-elasticstack/pull/819))
- Add `aliases` attribute to `elasticstack_elasticsearch_transform` resource. ([#825](https://github.com/elastic/terraform-provider-elasticstack/pull/825))
- Add `description` attribute to `elasticstack_elasticsearch_security_role` resource. ([#824](https://github.com/elastic/terraform-provider-elasticstack/pull/824))
- Fix merge values for `elasticstack_kibana_synthetics_monitor` monitor locations ([#823](https://github.com/elastic/terraform-provider-elasticstack/pull/823)
- Add `elasticstack_elasticsearch_index_template` data source ([#828](https://github.com/elastic/terraform-provider-elasticstack/pull/828))

## [0.11.8] - 2024-10-02

- Add key_id to the `elasticstack_elasticsearch_api_key` resource. ([#789](https://github.com/elastic/terraform-provider-elasticstack/pull/789))
- Fix handling of `sys_monitoring` in `elasticstack_fleet_agent_policy` ([#792](https://github.com/elastic/terraform-provider-elasticstack/pull/792))
- Migrate `elasticstack_fleet_agent_policy`, `elasticstack_fleet_integration` (both), and `elasticstack_fleet_server_host` to terraform-plugin-framework ([#785](https://github.com/elastic/terraform-provider-elasticstack/pull/785))
- Fix for synthetics http/tcp monitor produces inconsistent result after apply ([#801](https://github.com/elastic/terraform-provider-elasticstack/pull/801))
- Migrate `elasticstack_fleet_integration_policy` to terraform-plugin-framework. Fix drift in integration policy secrets. ([#797](https://github.com/elastic/terraform-provider-elasticstack/pull/797))
- Migrate `elasticstack_fleet_output` to terraform-plugin-framework. ([#811](https://github.com/elastic/terraform-provider-elasticstack/pull/811))

## [0.11.7] - 2024-09-20

- Add the `alerts_filter` field to the `actions` in the Create Rule API ([#774](https://github.com/elastic/terraform-provider-elasticstack/pull/774))
- Add the `alert_delay` field to the Create Rule API ([#715](https://github.com/elastic/terraform-provider-elasticstack/pull/715))
- Add support for data_stream `lifecycle` template settings ([#724](https://github.com/elastic/terraform-provider-elasticstack/pull/724))
- Fix a provider panic when `elasticstack_kibana_action_connector` reads a non-existant connector ([#729](https://github.com/elastic/terraform-provider-elasticstack/pull/729))
- Add support for `remote_indicies` to `elasticstack_elasticsearch_security_role` & `elasticstack_kibana_security_role` [#723](https://github.com/elastic/terraform-provider-elasticstack/pull/723)
- Fix error handling in `elasticstack_kibana_import_saved_objects` ([#738](https://github.com/elastic/terraform-provider-elasticstack/pull/738))
- Remove `space_id` parameter from private locations to fix inconsistent state for `elasticstack_kibana_synthetics_private_location` `space_id` ([#733](https://github.com/elastic/terraform-provider-elasticstack/pull/733))
- Add the `Frequency` field to the Create Rule API ([#753](https://github.com/elastic/terraform-provider-elasticstack/pull/753))
- Prevent a provider panic when the repository referenced in an `elasticstack_elasticsearch_snapshot_repository` does not exist ([#758](https://github.com/elastic/terraform-provider-elasticstack/pull/758))
- Add support for `remote_indicies` to `elasticstack_elasticsearch_security_api_key` [#766](https://github.com/elastic/terraform-provider-elasticstack/pull/766)
- Add support for `icmp` and `browser` monitor types to `elasticstack_kibana_synthetics_monitor` resource [#772](https://github.com/elastic/terraform-provider-elasticstack/pull/772)
- Migrate `elasticstack_fleet_enrollment_tokens` to terraform-plugin-framework ([#778](https://github.com/elastic/terraform-provider-elasticstack/pull/778))

## [0.11.6] - 2024-08-20

- Improve validation for index settings and mappings ([#719](https://github.com/elastic/terraform-provider-elasticstack/pull/719))
- Add support for Kibana synthetics http and tcp monitors ([#699](https://github.com/elastic/terraform-provider-elasticstack/pull/699))
- Add `elasticstack_kibana_spaces` data source ([#682](https://github.com/elastic/terraform-provider-elasticstack/pull/682))

## [0.11.5] - 2024-08-12

- Fix setting `id` for Fleet outputs and servers ([#666](https://github.com/elastic/terraform-provider-elasticstack/pull/666))
- Fix `elasticstack_fleet_enrollment_tokens` returning empty tokens in some case ([#683](https://github.com/elastic/terraform-provider-elasticstack/pull/683))
- Add support for Kibana synthetics private locations ([#696](https://github.com/elastic/terraform-provider-elasticstack/pull/696))
- Support setting `restriction` in `elasticstack_elasticsearch_security_api_key` role definitions ([#577](https://github.com/elastic/terraform-provider-elasticstack/pull/577))
- Fix type of `group_by` attribute in the `kibana_slo` resource to be compatible with versions 8.14+ ([#701](https://github.com/elastic/terraform-provider-elasticstack/pull/701))

## [0.11.4] - 2024-06-13

### Breaking changes

- The `title` attribute is now required in the elasticstack_kibana_data_view resource. In practice the resource didn't work without this set, the schema now enforces it's correctly configured.

### Fixed

- Populate policy_id when importing fleet policies and integrations ([#646](https://github.com/elastic/terraform-provider-elasticstack/pull/646))
- Fix alerting rule update crash when backend responds with HTTP 4xx. ([#649](https://github.com/elastic/terraform-provider-elasticstack/pull/649))
- Fix the elasticstack_kibana_data_view resource when not specifying an `id` and running against Kibana 8.14 ([#663](https://github.com/elastic/terraform-provider-elasticstack/pull/663))
- Support allow_write_after_shrink when managing ILM policies ([#662](https://github.com/elastic/terraform-provider-elasticstack/pull/662))
- Support managing image_url in Kibana spaces ([#664](https://github.com/elastic/terraform-provider-elasticstack/pull/664))

## [0.11.3] - 2024-05-16

### Fixed

- Prevent a provider panic when an `elasticstack_elasticsearch_template` or `elasticstack_elasticsearch_component_template` includes an empty `template` (`template {}`) block. ([#598](https://github.com/elastic/terraform-provider-elasticstack/pull/598))
- Prevent `elasticstack_kibana_space` to attempt the space recreation if `initials` and `color` are not provided. ([#606](https://github.com/elastic/terraform-provider-elasticstack/pull/606))
- Prevent a provider panic in `elasticstack_kibana_data_view` when a `field_format` does not include a `pattern`. ([#619](https://github.com/elastic/terraform-provider-elasticstack/pull/619/files))
- Fixed a bug where the `id` attribute for `elasticstack_kibana_slo` resources was ignored by renaming the attribute to `slo_id`. ([#622](https://github.com/elastic/terraform-provider-elasticstack/pull/622))
- Fixed a bug where the `rule_id` attribute for `elasticstack_kibana_alerting_rule` was ignored. ([#626](https://github.com/elastic/terraform-provider-elasticstack/pull/626))
- Fixed a bug with incorrect HTTP header name for API key for alerting client. ([#633](https://github.com/elastic/terraform-provider-elasticstack/pull/633))
- Fix provider crash when running against Serverless projects ([#630](https://github.com/elastic/terraform-provider-elasticstack/pull/630))

### Added

- Added datasource for alerting connectors. ([#607](https://github.com/elastic/terraform-provider-elasticstack/pull/607))

## [0.11.2] - 2024-03-13

### Fixed

- Fix authentication for fleet API (using ApiKey instead of Bearer keyword) ([#576](https://github.com/elastic/terraform-provider-elasticstack/pull/576))
- Ensure all Kibana resources use the supplied `ca_certs` value. ([#585](https://github.com/elastic/terraform-provider-elasticstack/pull/585))
- Don't panic when SLM indices are specified as a CSV string rather than an array ([#593](https://github.com/elastic/terraform-provider-elasticstack/pull/593))

## [0.11.1] - 2024-02-17

### Added

- Add downsample section to ILMs [#538](https://github.com/elastic/terraform-provider-elasticstack/pull/538)
- Add new optional `ca_certs` attribute for Kibana ([#507](https://github.com/elastic/terraform-provider-elasticstack/pull/507))
- Support API key authentication in `elasticstack_kibana_data_view` resource ([#549](https://github.com/elastic/terraform-provider-elasticstack/pull/549))

### Fixed

- Handle nil LastExecutionDate's in Kibana alerting rules. ([#508](https://github.com/elastic/terraform-provider-elasticstack/pull/508))
- Import all relevant attributes during `elasticstack_fleet_output` import ([#522](https://github.com/elastic/terraform-provider-elasticstack/pull/522))
- Fix issue when setting `override` in `elasticstack_kibana_data_view` resource ([#550](https://github.com/elastic/terraform-provider-elasticstack/pull/550))
- Fixup typos in `elasticstack_elasticsearch_transform` and `elasticstack_kibana_security_role` docs ([#551](https://github.com/elastic/terraform-provider-elasticstack/pull/551))
- Fix issue when setting `field_attrs` in `elasticstack_kibana_data_view` resource ([#552](https://github.com/elastic/terraform-provider-elasticstack/pull/552))
- Fixup support for managing `elasticstack_kibana_data_view` resources in non-default spaces ([#559](https://github.com/elastic/terraform-provider-elasticstack/pull/559))
- Add an example resource and import command to the `elasticstack_kibana_data_view` docs ([#560](https://github.com/elastic/terraform-provider-elasticstack/pull/560))

## [0.11.0] - 2023-12-12

### Added

- Switch to Terraform [protocol version 6](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-6) that is compatible with Terraform CLI version 1.0 and later.
- Add 'elasticstack_fleet_package' data source ([#469](https://github.com/elastic/terraform-provider-elasticstack/pull/469))
- Add `tags` option to Kibana's SLOs ([#495](https://github.com/elastic/terraform-provider-elasticstack/pull/495))
- Add support for Authorization header - Bearer Token and ES-Client-Authentication fields added.([#500](https://github.com/elastic/terraform-provider-elasticstack/pull/500))
- Add support for managing Kibana Data Views ([#502](https://github.com/elastic/terraform-provider-elasticstack/pull/502))
- Support Logstash SSL fields in Fleet output ([#498](https://github.com/elastic/terraform-provider-elasticstack/pull/498))
- Support for Kibana API Key authentication ([#372](https://github.com/elastic/terraform-provider-elasticstack/pull/372))

### Fixed

- Rename fleet package objects to `elasticstack_fleet_integration` and `elasticstack_fleet_integration_policy` ([#476](https://github.com/elastic/terraform-provider-elasticstack/pull/476))
- Fix a provider crash when managing SLOs outside of the default Kibana space. ([#485](https://github.com/elastic/terraform-provider-elasticstack/pull/485))
- Make input optional for `elasticstack_fleet_integration_policy` ([#493](https://github.com/elastic/terraform-provider-elasticstack/pull/493))
- Sort Fleet integration policy inputs to ensure consistency ([#494](https://github.com/elastic/terraform-provider-elasticstack/pull/494))
- Updated Elasticsearch role_mapping.go to enforce the replacement/updates of role mapping resources when the name field is altered. ([#503](https://github.com/elastic/terraform-provider-elasticstack/pull/503))

## [0.10.0] - 2023-11-02

### Added

- Add support for Kibana security role ([#435](https://github.com/elastic/terraform-provider-elasticstack/pull/435))
- Introduce `elasticstack_kibana_import_saved_objects` resource as an additive only way to manage Kibana saved objects ([#343](https://github.com/elastic/terraform-provider-elasticstack/pull/343)).
- Add support for Terraform Plugin Framework ([#343](https://github.com/elastic/terraform-provider-elasticstack/pull/343)).
- Fix fleet resources not having ID set on import ([#447](https://github.com/elastic/terraform-provider-elasticstack/pull/447))
- Fix Fleet Agent Policy monitoring settings ([#448](https://github.com/elastic/terraform-provider-elasticstack/pull/448))
- Add `elasticstack_elasticsearch_info` data source. ([#467](https://github.com/elastic/terraform-provider-elasticstack/pull/467))
- Add `elasticstack_fleet_package` and `elasticstack_fleet_package_policy` resources ([#454](https://github.com/elastic/terraform-provider-elasticstack/pull/454))

## [0.9.0] - 2023-10-09

### Added

- Update `elasticstack_fleet_output` to use new API schema format ([#434](https://github.com/elastic/terraform-provider-elasticstack/pull/434))

### Fixed

- Fix mapping of webhook connectors that is stored in tfstate ([#433](https://github.com/elastic/terraform-provider-elasticstack/pull/433))

## [0.8.0] - 2023-09-26

### Added

- Add support for the `.slack_api` connector type for Kibana action connectors ([#419](https://github.com/elastic/terraform-provider-elasticstack/pull/419))
- resource `elasticstack_kibana_slo`: Update `histogram_custom_indicator` `from` and `to` fields to float ([#430](https://github.com/elastic/terraform-provider-elasticstack/pull/430))

## [0.7.0] - 2023-08-22

### Added

- Add support for Kibana SLOs ([#385](https://github.com/elastic/terraform-provider-elasticstack/pull/385))
- Document all available environment variables ([#405](https://github.com/elastic/terraform-provider-elasticstack/pull/405))

## [0.6.2] - 2023-06-19

### Added

- Logging of Kibana action connectors HTTP requests and responses when [Terraform logs are enabled](https://developer.hashicorp.com/terraform/internals/debugging).
- Add `skip_destroy` flag to `elasticstack_fleet_agent_policy` resource ([#357](https://github.com/elastic/terraform-provider-elasticstack/pull/357))

## [0.6.1] - 2023-05-30

### Added

- Add `path_style_access` setting to `elasticstack_elasticsearch_snapshot_repository` on s3 repositories to enable path style access pattern ([#331](https://github.com/elastic/terraform-provider-elasticstack/pull/331))
- Add `transform` field to `elasticstack_elasticsearch_watch` to allow for payload transforms to be defined ([#340](https://github.com/elastic/terraform-provider-elasticstack/pull/340))

### Fixed

- Fix error presented by incorrect handling of `disabled_features` field in `elasticstack_kibana_space` resource ([#340](https://github.com/elastic/terraform-provider-elasticstack/pull/340))

## [0.6.0] - 2023-05-24

### Added

- New resource `elasticstack_elasticsearch_enrich_policy` to manage enrich policies ([#286](https://github.com/elastic/terraform-provider-elasticstack/pull/286)) ([Enrich API](https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-apis.html))
- New data source `elasticstack_elasticsearch_enrich_policy` to read enrich policies ([#293](https://github.com/elastic/terraform-provider-elasticstack/pull/293)) ([Enrich API](https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-apis.html))
- Add 'mapping_coerce' field to index resource ([#229](https://github.com/elastic/terraform-provider-elasticstack/pull/229))
- Add 'min\_\*' conditions to ILM rollover ([#250](https://github.com/elastic/terraform-provider-elasticstack/pull/250))
- Add support for Kibana connections ([#226](https://github.com/elastic/terraform-provider-elasticstack/pull/226))
- **[Breaking Change] Add 'deletion_protection' field to index resource** to avoid unintentional deletion. ([#167](https://github.com/elastic/terraform-provider-elasticstack/pull/167))

  - To delete index resource, you'll need to explicitly set `deletion_protection = false` as follows.

  ```terraform
  resource "elasticstack_elasticsearch_index" "example" {
    name = "example"
    mappings = jsonencode({
      properties = {
        field1    = { type = "text" }
      }
    })
    deletion_protection = false
  }
  ```

- Add `elasticstack_kibana_space` for managing Kibana spaces ([#272](https://github.com/elastic/terraform-provider-elasticstack/pull/272))
- Add `elasticstack_elasticsearch_transform` for managing Elasticsearch transforms ([#284](https://github.com/elastic/terraform-provider-elasticstack/pull/284))
- Add `elasticstack_elasticsearch_watch` for managing Elasticsearch Watches ([#155](https://github.com/elastic/terraform-provider-elasticstack/pull/155))
- Add `elasticstack_kibana_alerting_rule` for managing Kibana alerting rules ([#292](https://github.com/elastic/terraform-provider-elasticstack/pull/292))
- Add client for communicating with the Fleet APIs ([#311](https://github.com/elastic/terraform-provider-elasticstack/pull/311)])
- Add `elasticstack_fleet_enrollment_tokens` and `elasticstack_fleet_agent_policy` for managing Fleet enrollment tokens and agent policies ([#322](https://github.com/elastic/terraform-provider-elasticstack/pull/322))
- Add `elasticstack_fleet_output` and `elasticstack_fleet_server_host` for managing Fleet outputs and server hosts ([#327](https://github.com/elastic/terraform-provider-elasticstack/pull/327)])
- Add `elasticstack_kibana_action_connector` for managing Kibana action connectors ([#306](https://github.com/elastic/terraform-provider-elasticstack/pull/306))

### Fixed

- Updated unsupported queue_max_bytes_number and queue_max_bytes_units with queue.max_bytes ([#266](https://github.com/elastic/terraform-provider-elasticstack/issues/266))
- Respect `ignore_unavailable` and `include_global_state` values when configuring SLM policies ([#224](https://github.com/elastic/terraform-provider-elasticstack/pull/224))
- Refactor API client functions and return diagnostics ([#220](https://github.com/elastic/terraform-provider-elasticstack/pull/220))
- Fix not to recreate index when field is removed from mapping ([#232](https://github.com/elastic/terraform-provider-elasticstack/pull/232))
- Add query params fields to index resource ([#244](https://github.com/elastic/terraform-provider-elasticstack/pull/244))
- Properly handle errors which occur during provider execution ([#262](https://github.com/elastic/terraform-provider-elasticstack/pull/262))
- Correctly handle empty logstash pipeline metadata in plan diffs ([#256](https://github.com/elastic/terraform-provider-elasticstack/pull/256))
- Fix error when logging API requests in debug mode ([#259](https://github.com/elastic/terraform-provider-elasticstack/pull/259))
- **[Breaking Change] Change `pipeline_metadata` type from schema.TypeMap to schema.TypeString**. This is to fix an error caused by updates to Logstash Pipelines outside of TF ([#278](https://github.com/elastic/terraform-provider-elasticstack/issues/278))

  - To use the updated `pipeline_metadata` field, you'll need to encapsulate any Terraform configuration with **jsonencode{}** as follows:

    ```terraform
    resource "elasticstack_elasticsearch_logstash_pipeline" "example" {
      name = "example"
      pipeline = <<-EOF
        input{}
        filter{}
        output{}
    EOF
      pipeline_metadata = jsonencode({
        type = "logstash_pipeline"
        version = 1
      })
    }
    ```

  - If migrating existing resources in state from a previous version of the provider, then you will need to remove the reference to the resources in state before reapplying / reimporting:
    - Run `terraform state rm` against your logstash pipelines (<https://developer.hashicorp.com/terraform/cli/commands/state/rm>)
    - Ensure any definitions of the `pipeline_metadata` field in your resource definitions have been encapsulated with `jsonencode()` as mentioned above.
    - EITHER
      - run `terraform plan`
      - run `terraform apply`
    - OR
      - reimport the resources into state using `terraform import` (<https://developer.hashicorp.com/terraform/cli/import>)

- Fix order of `indices` field in SLM ([#326](https://github.com/elastic/terraform-provider-elasticstack/pull/326))

## [0.5.0] - 2022-12-07

### Added

- New resource `elasticstack_elasticsearch_logstash_pipeline` to manage Logstash pipelines ([Centralized Pipeline Management](https://www.elastic.co/guide/en/logstash/current/logstash-centralized-pipeline-management.html)) ([#151](https://github.com/elastic/terraform-provider-elasticstack/pull/151))
- Add `elasticstack_elasticsearch_script` resource ([#173](https://github.com/elastic/terraform-provider-elasticstack/pull/173))
- Add `elasticstack_elasticsearch_security_role` data source ([#177](https://github.com/elastic/terraform-provider-elasticstack/pull/177))
- Add `elasticstack_elasticsearch_security_role_mapping` data source ([#178](https://github.com/elastic/terraform-provider-elasticstack/pull/178))
- Apply `total_shards_per_node` setting in `allocate` action in ILM. Supported from Elasticsearch version **7.16** ([#112](https://github.com/elastic/terraform-provider-elasticstack/issues/112))
- Add `elasticstack_elasticsearch_security_api_key` resource ([#193](https://github.com/elastic/terraform-provider-elasticstack/pull/193))
- Add `elasticstack_elasticsearch_security_system_user` resource to manage built-in user ([#188](https://github.com/elastic/terraform-provider-elasticstack/pull/188))
- Add `unassigned_node_left_delayed_timeout` to index resource ([#196](https://github.com/elastic/terraform-provider-elasticstack/pull/196))
- Add support for Client certificate based authentication ([#191](https://github.com/elastic/terraform-provider-elasticstack/pull/191))
- Deprecate `elasticsearch_connection` blocks on individual resources/data sources. Connection configuration should be configured directly on the provider with multiple provider instances used to connect to different clusters. ([#218](https://github.com/elastic/terraform-provider-elasticstack/pull/218))

### Fixed

- Remove unnecessary unsetting id on delete ([#174](https://github.com/elastic/terraform-provider-elasticstack/pull/174))
- Fix not found handling for snapshot repository ([#175](https://github.com/elastic/terraform-provider-elasticstack/pull/175))
- Add warn log when to remove resource from state ([#185](https://github.com/elastic/terraform-provider-elasticstack/pull/185))
- Import snapshot repository name when importing ([#187](https://github.com/elastic/terraform-provider-elasticstack/pull/187))

## [0.4.0] - 2022-10-07

### Added

- Add ca_data field to provider schema ([#145](https://github.com/elastic/terraform-provider-elasticstack/pull/145))
- Add individual setting fields ([#137](https://github.com/elastic/terraform-provider-elasticstack/pull/137))
- Allow use of `api_key` instead of `username`/`password` for authentication ([#130](https://github.com/elastic/terraform-provider-elasticstack/pull/130))
- Add `allow_restricted_indices` setting to security role ([#125](https://github.com/elastic/terraform-provider-elasticstack/issues/125))
- Add conditional to only set `password` and `password_hash` when a new value is defined ([#127](https://github.com/elastic/terraform-provider-elasticstack/pull/128))
- Add support for ELASTICSEARCH_INSECURE environment variable as the default of the `insecure` config value ([#127](https://github.com/elastic/terraform-provider-elasticstack/pull/128))
- Add elasticstack_elasticsearch_security_role_mapping resource ([148](https://github.com/elastic/terraform-provider-elasticstack/pull/148))

### Fixed

- Refactor main function not to use deprecated debug method ([#149](https://github.com/elastic/terraform-provider-elasticstack/pull/149))
- Expose provider package ([#142](https://github.com/elastic/terraform-provider-elasticstack/pull/142))
- Upgrade Go version to 1.19 and sdk version to v2.22.0 ([#139](https://github.com/elastic/terraform-provider-elasticstack/pull/139))
- Make API calls context aware to be able to handle timeouts ([#138](https://github.com/elastic/terraform-provider-elasticstack/pull/138))
- Correctly identify a missing security user ([#101](https://github.com/elastic/terraform-provider-elasticstack/issues/101))
- Support **7.x** Elasticsearch < **7.15** by removing the default `media_type` attribute in the Append processor ([#118](https://github.com/elastic/terraform-provider-elasticstack/pull/118))

## [0.3.3] - 2022-03-22

### Fixed

- Make sure it is possible to set priority to `0` in ILM template ([#88](https://github.com/elastic/terraform-provider-elasticstack/issues/88))
- Set the ILM name on read operation ([#87](https://github.com/elastic/terraform-provider-elasticstack/issues/87))
- Always use data_stream setting if it's present ([#91](https://github.com/elastic/terraform-provider-elasticstack/issues/91))

## [0.3.2] - 2022-02-28

### Fixed

- Properly apply `number_of_replicas` setting in `allocate` action in ILM ([#80](https://github.com/elastic/terraform-provider-elasticstack/issues/80))

## [0.3.1] - 2022-02-24

### Fixed

- Add new field `allow_custom_routing` in `data_stream` section of [`index_template`](https://www.elastic.co/guide/en/elasticsearch/reference/8.0/indices-put-template.html#put-index-template-api-request-body), which appears only in Elasticsearch version **8.0.0**. Make sure `index_template` resource can work with both **7.x** and **8.x** versions ([#72](https://github.com/elastic/terraform-provider-elasticstack/pull/72))
- Panic using `elasticstack_elasticsearch_security_user_data_source` when the user absent or there is not enough permissions to fetch users from ES API ([#73](https://github.com/elastic/terraform-provider-elasticstack/issues/73))
- Fix typo in the index alias model ([#78](https://github.com/elastic/terraform-provider-elasticstack/issues/78))

## [0.3.0] - 2022-02-17

### Added

- New resource `elasticstack_elasticsearch_data_stream` to manage Elasticsearch [data streams](https://www.elastic.co/guide/en/elasticsearch/reference/current/data-streams.html) ([#45](https://github.com/elastic/terraform-provider-elasticstack/pull/45))
- New resource `elasticstack_elasticsearch_ingest_pipeline` to manage Elasticsearch [ingest pipelines](https://www.elastic.co/guide/en/elasticsearch/reference/7.16/ingest.html) ([#56](https://github.com/elastic/terraform-provider-elasticstack/issues/56))
- New resource `elasticstack_elasticsearch_component_template` to manage Elasticsearch [component templates](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-component-template.html) ([#39](https://github.com/elastic/terraform-provider-elasticstack/pull/39))
- New helper data sources to create [processorts](https://www.elastic.co/guide/en/elasticsearch/reference/current/processors.html) for ingest pipelines ([#67](https://github.com/elastic/terraform-provider-elasticstack/pull/67))

### Fixed

- Update only changed index settings ([#52](https://github.com/elastic/terraform-provider-elasticstack/issues/52))
- Enable import of index settings ([#53](https://github.com/elastic/terraform-provider-elasticstack/issues/53))
- Handle `allocate` action in ILM policy ([#59](https://github.com/elastic/terraform-provider-elasticstack/issues/59))
- Send only initialized values to Elasticsearch API when managing the users ([#66](https://github.com/elastic/terraform-provider-elasticstack/issues/66))

## [0.2.0] - 2022-01-27

### Added

- New resource to manage Elasticsearch indices ([#38](https://github.com/elastic/terraform-provider-elasticstack/pull/38))
- The `insecure` option to the Elasticsearch provider configuration to ignore certificate verification ([#36](https://github.com/elastic/terraform-provider-elasticstack/pull/36))
- The `ca_file` option to the Elasticsearch provider configuration to provide path to the custom root certificate file ([#35](https://github.com/elastic/terraform-provider-elasticstack/pull/35))
- Documentation templates for all the existing resources ([#32](https://github.com/elastic/terraform-provider-elasticstack/pull/32))

### Fixed

- Identify missing or deleted resources in the Elasticsearch cluster and make sure Terraform can re-create them ([#40](https://github.com/elastic/terraform-provider-elasticstack/pull/40))

### Changed

- **[Breaking]** Rename `aliases` configuration option in
  `elasticstack_elasticsearch_index_template` resource to singular `alias`

## [0.1.0] - 2021-12-20

### Added

- Initial set of the resources to gather feedback from the community
- Initial set of docs
- CI integration

[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.16.3...HEAD
[0.16.3]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.16.2...v0.16.3
[0.16.2]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.16.1...v0.16.2
[0.16.1]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.16.0...v0.16.1
[0.16.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.15.2...v0.16.0
[0.15.2]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.15.1...v0.15.2
[0.15.1]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.15.0...v0.15.1
[0.15.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.5...v0.15.0
[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5
[0.14.4]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.3...v0.14.4
[0.14.3]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.2...v0.14.3
[0.14.2]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.1...v0.14.2
[0.14.1]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.0...v0.14.1
[0.14.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.13.1...v0.14.0
[0.13.1]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.13.0...v0.13.1
[0.13.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.12.2...v0.13.0
[0.12.2]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.12.1...v0.12.2
[0.12.1]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.12.0...v0.12.1
[0.12.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.18...v0.12.0
[0.11.19]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.18...v0.11.19
[0.11.18]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.17...v0.11.18
[0.11.17]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.16...v0.11.17
[0.11.16]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.15...v0.11.16
[0.11.15]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.14...v0.11.15
[0.11.14]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.13...v0.11.14
[0.11.13]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.12...v0.11.13
[0.11.12]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.11...v0.11.12
[0.11.11]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.10...v0.11.11
[0.11.10]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.9...v0.11.10
[0.11.9]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.8...v0.11.9
[0.11.8]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.7...v0.11.8
[0.11.7]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.6...v0.11.7
[0.11.6]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.5...v0.11.6
[0.11.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.4...v0.11.5
[0.11.4]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.3...v0.11.4
[0.11.3]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.2...v0.11.3
[0.11.2]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.1...v0.11.2
[0.11.1]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.0...v0.11.1
[0.11.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.6.2...v0.7.0
[0.6.2]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.6.1...v0.6.2
[0.6.1]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.3.3...v0.4.0
[0.3.3]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/elastic/terraform-provider-elasticstack/releases/tag/v0.1.0
