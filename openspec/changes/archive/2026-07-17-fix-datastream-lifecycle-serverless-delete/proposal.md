## Why

Destroying `elasticstack_elasticsearch_data_stream_lifecycle` on Elastic Cloud Serverless fails because the Delete Data Lifecycle API is unavailable. Serverless manages data stream lifecycle and retention automatically, so the provider must not attempt to remove them.

## What Changes

- Detect Elastic Cloud Serverless before deleting a data stream lifecycle.
- On serverless, skip the Delete Data Lifecycle API request, emit a warning diagnostic, and permit Terraform to remove the resource from state.
- Preserve Delete API behavior on stateful Elasticsearch deployments and preserve error behavior when deployment detection fails.
- Document the serverless-specific delete semantics in the canonical data stream lifecycle requirements.

## Capabilities

### Modified Capabilities

- `elasticsearch-data-stream-lifecycle`: Define delete behavior for Elastic Cloud Serverless, including skipping the unavailable API, returning a warning, and removing state without changing the server-side data stream lifecycle.

## Impact

- `internal/clients/elasticsearch/datastream.go`
- `internal/clients/elasticsearch/datastream_test.go`
- `openspec/specs/elasticsearch-data-stream-lifecycle/spec.md`
