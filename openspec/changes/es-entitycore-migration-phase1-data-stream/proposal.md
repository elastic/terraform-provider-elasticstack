## Why

Migrate `elasticstack_elasticsearch_data_stream` from the Terraform Plugin SDK to the Plugin Framework and the entitycore `NewElasticsearchResource` envelope. The data stream resource is a simple PUT-on-create / GET-read / DELETE-delete resource with no update behavior, making it an ideal early SDK-to-PF migration target.

## What Changes

- Replace the Plugin SDK resource in `internal/elasticsearch/index/data_stream.go` with a Plugin Framework resource.
- Embed `*entitycore.ElasticsearchResource[Data]` and provide envelope callbacks.
- Convert model fields (`name`, `timestamp_field`, `indices`, `generation`, `metadata`, `status`, `template`, `ilm_policy`, `hidden`, `system`, `replicated`) to Plugin Framework types.
- Remove SDK connection handling, `schema.Resource` definition, and SDK CRUD functions.
- Move provider registration from the SDK `provider/provider.go` `ResourcesMap` to the Plugin Framework resource registry in `provider/plugin_framework.go`.

## Capabilities

### New Capabilities
- `data-stream-resource`: Elasticsearch data stream resource managed via the entitycore envelope.

### Modified Capabilities
<!-- No existing spec-level behavior changes. -->

## Impact

- Affects `internal/elasticsearch/index/data_stream.go` and `data_stream_test.go`.
- No Terraform interface changes.
