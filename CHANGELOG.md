## [Unreleased]

- Fix handling of `sys_monitoring` in `elasticstack_fleet_agent_policy` ([#792](https://github.com/elastic/terraform-provider-elasticstack/pull/792))
- Migrate `elasticstack_fleet_agent_policy`, `elasticstack_fleet_integration` (both), and `elasticstack_fleet_server_host` to terraform-plugin-framework ([#785](https://github.com/elastic/terraform-provider-elasticstack/pull/785))
- Fix for synthetics http/tcp monitor produces inconsistent result after apply ([#801](https://github.com/elastic/terraform-provider-elasticstack/pull/801))

## [0.11.7] - 2024-09-20

- Add the `alerts_filter` field to the `actions` in the Create Rule API ([#774](https://github.com/elastic/terraform-provider-elasticstack/pull/774))
- Add the `alert_delay` field to the Create Rule API ([#715](https://github.com/elastic/terraform-provider-elasticstack/pull/715))
- Add support for data_stream `lifecycle` template settings ([#724](https://github.com/elastic/terraform-provider-elasticstack/pull/724))
- Fix a provider panic when `elasticstack_kibana_action_connector` reads a non-existant connector ([#729](https://github.com/elastic/terraform-provider-elasticstack/pull/729))
- Add support for `remote_indicies` to `elasticstack_elasticsearch_security_role` & `elasticstack_kibana_security_role` (#723)[https://github.com/elastic/terraform-provider-elasticstack/pull/723]
- Fix error handling in `elasticstack_kibana_import_saved_objects` ([#738](https://github.com/elastic/terraform-provider-elasticstack/pull/738))
- Remove `space_id` parameter from private locations to fix inconsistent state for `elasticstack_kibana_synthetics_private_location` `space_id` ([#733](https://github.com/elastic/terraform-provider-elasticstack/pull/733))
- Add the `Frequency` field to the Create Rule API ([#753](https://github.com/elastic/terraform-provider-elasticstack/pull/753))
- Prevent a provider panic when the repository referenced in an `elasticstack_elasticsearch_snapshot_repository` does not exist ([#758](https://github.com/elastic/terraform-provider-elasticstack/pull/758))
- Add support for `remote_indicies` to `elasticstack_elasticsearch_security_api_key` (#766)[https://github.com/elastic/terraform-provider-elasticstack/pull/766]
- Add support for `icmp` and `browser` monitor types to `elasticstack_kibana_synthetics_monitor` resource (#772)[https://github.com/elastic/terraform-provider-elasticstack/pull/772]
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

[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.7...HEAD
[0.11.6]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.11.6...v0.11.7
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
