## Why

The `elasticstack_elasticsearch_index_template` resource's `template` block does not support the `data_stream_options` field introduced in Elasticsearch 9.x. This field controls failure store behavior for data streams. Without it, practitioners cannot enable the failure store via Terraform, forcing out-of-band API calls that create configuration drift between Terraform state and actual cluster configuration.

## What Changes

- Add a `data_stream_options` block inside the `template` block of `elasticstack_elasticsearch_index_template`.
- The `data_stream_options` block supports a nested `failure_store` block with an `enabled` boolean.
- Extend the `models.Template` struct to carry the new `DataStreamOptions` field so the JSON round-trip works correctly.
- Expand and flatten `data_stream_options` in `expandTemplate` and `flattenTemplateData` in `internal/elasticsearch/index/template.go`.
- Apply the same addition to the data source at `internal/elasticsearch/index/template_data_source.go`.
- Regenerate resource and data source documentation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-index-template`: add `data_stream_options` support inside the `template` block, covering failure store management for Elasticsearch 9.x data streams.

## Impact

- Affected code: `internal/elasticsearch/index/template.go`, `internal/elasticsearch/index/template_data_source.go`, `internal/models/models.go`.
- Affected interfaces: Terraform schema for `elasticstack_elasticsearch_index_template` (both resource and data source), model serialization and deserialization, and acceptance test coverage.
- No breaking schema changes. The new block is optional and additive; existing configurations without `data_stream_options` continue to work unchanged.
