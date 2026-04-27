## Why

The `elasticstack_elasticsearch_index_template` resource's `template` block does not support `data_stream_options`, a field introduced in Elasticsearch 9.x. Without this field, practitioners cannot enable the failure store for data streams via Terraform. The only workarounds involve out-of-band API calls (using `local-exec` with `curl`/`jq`), which breaks infrastructure-as-code principles and creates drift between Terraform state and actual cluster configuration.

## What Changes

- Add a `data_stream_options` block inside the `template` block of `elasticstack_elasticsearch_index_template`.
- `data_stream_options` exposes a `failure_store` sub-block with:
  - `enabled` — required boolean to enable or disable the failure store on new data streams.
  - `lifecycle.data_retention` — optional string to set failure store retention duration (e.g. `"30d"`).
- Version-gate `data_stream_options` to Elasticsearch >= 9.1.0.
- Extend the `models.Template` struct with a `DataStreamOptions` field serialized as `data_stream_options` in JSON.
- Update the create/update path (`expandTemplate`) and the read path (`flattenTemplateData`) to handle the new field.
- Add acceptance test coverage for `data_stream_options` including create, update, and read-back scenarios.

## Capabilities

### New Capabilities

None — this change extends an existing resource.

### Modified Capabilities

- `elasticsearch-index-template`: add `data_stream_options` support inside the `template` block, including `failure_store.enabled` and optional `failure_store.lifecycle.data_retention`, with version gating for Elasticsearch >= 9.1.0.

## Impact

- Affected code: `internal/elasticsearch/index/template.go`, `internal/models/models.go`, and acceptance test fixtures in `internal/elasticsearch/index/testdata/`.
- Affected interfaces: the Terraform schema for `elasticstack_elasticsearch_index_template` (new `template.data_stream_options` block), CRUD request construction, and state read/refresh logic.
- The change is additive and backward compatible. Existing configurations without `data_stream_options` are unaffected.

## Assumptions and Open Questions

- **Version introduced**: The Elastic docs describe failure store as "Preview in 9.1" and "Generally available since 9.2". The version gate is conservatively set to >= 9.1.0 to support preview users and align with when the API field first appeared.
- **One-time application**: Per the API documentation, `data_stream_options` inside an index template is applied to a data stream only once at creation time; it does NOT affect existing data streams or backing indices on rollover. This is documented for practitioners in the resource description rather than enforced by the provider.
- **Failure store lifecycle retention**: The docs confirm that `failure_store.lifecycle.data_retention` can be specified in the index template. This will be surfaced as an optional nested `lifecycle` block.
- **Component templates**: The `data_stream_options` field is specific to index templates, not component templates. The `elasticstack_elasticsearch_component_template` resource uses the same `models.Template` struct. Adding `DataStreamOptions` to the shared struct will not break component templates because the field is `omitempty` and the component template resource never populates it.
