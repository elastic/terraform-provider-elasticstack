## Why

The `elasticstack_elasticsearch_index_template` resource's `template` block does not expose `data_stream_options`, a field introduced in Elasticsearch 9.x that controls data-stream-specific behavior such as enabling the failure store. Without this, operators cannot configure failure-store enablement via Terraform, forcing them to use out-of-band API calls, which creates drift between Terraform state and actual cluster configuration.

## What Changes

- Add a `data_stream_options` nested block inside `template` that supports a `failure_store` sub-block with a boolean `enabled` attribute.
- Extend the `models.Template` Go struct with a `DataStreamOptions` field serialized as `data_stream_options`.
- Extend `expandTemplate` to read and populate the new block during create/update.
- Extend `flattenTemplateData` to write the new block back into state during read.
- Gate the new field on Elasticsearch >= 9.0.0 and return an error diagnostic when configured against an older cluster.
- Add a delta spec under the `elasticsearch-index-template` capability that documents the new requirements.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-index-template`: add `data_stream_options.failure_store.enabled` to the `template` block, with a version gate requiring Elasticsearch >= 9.0.0.

## Impact

- Affected code is limited to `internal/elasticsearch/index/template.go` and `internal/models/models.go`.
- No schema changes are required to the data source (`template_data_source.go`) unless data-source read also needs to surface the field — the data source should be updated for symmetry.
- The change is additive; existing configurations that omit `data_stream_options` are unaffected.
