## [Unreleased]
### Added
- New resource `elasticstack_elasticsearch_data_stream` to manage Elasticsearch [data streams](https://www.elastic.co/guide/en/elasticsearch/reference/current/data-streams.html) ([#45](https://github.com/elastic/terraform-provider-elasticstack/pull/45))
- New resource `elasticstack_elasticsearch_ingest_pipeline` to manage Elasticsearch [ingest pipelines](https://www.elastic.co/guide/en/elasticsearch/reference/7.16/ingest.html) ([#56](https://github.com/elastic/terraform-provider-elasticstack/issues/56))
- New resource `elasticstack_elasticsearch_component_template` to manage Elasticsearch [component templates](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-component-template.html) ([#39](https://github.com/elastic/terraform-provider-elasticstack/pull/39))

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
- Documentation templates for all the exisiting resources ([#32](https://github.com/elastic/terraform-provider-elasticstack/pull/32))

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

[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/elastic/terraform-provider-elasticstack/releases/tag/v0.1.0
