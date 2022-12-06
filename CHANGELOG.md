## [Unreleased]

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

[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.3.3...v0.4.0
[0.3.3]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/elastic/terraform-provider-elasticstack/releases/tag/v0.1.0
