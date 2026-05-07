## Why

Migrate `elasticstack_elasticsearch_ingest_pipeline` from the Terraform Plugin SDK to the Plugin Framework and the entitycore `NewElasticsearchResource` envelope. This eliminates duplicated CRUD boilerplate, centralizes client resolution in the envelope, and aligns with the canonical pattern established by recently-migrated security resources.

## What Changes

- Replace the Plugin SDK resource in `internal/elasticsearch/ingest/pipeline.go` with a Plugin Framework resource that embeds `*entitycore.ElasticsearchResource[Data]`.
- Provide `schemaFactory`, `readFunc`, `deleteFunc`, `createFunc`, and `updateFunc` callbacks for the envelope.
- Remove SDK-specific connection handling, ID parsing, and CRUD framing.
- Convert `metadata`, `description`, `on_failure`, and `processors` fields to Plugin Framework types.
- Add `GetID()` and `GetResourceID()` / `GetElasticsearchConnection()` on the new model type.
- Move provider registration from the SDK `provider/provider.go` `ResourcesMap` to the Plugin Framework resource registry in `provider/plugin_framework.go`.
- Add/import state migration path if any existing state shape differs (likely straight passthrough).

## Capabilities

### New Capabilities
- `ingest-pipeline-resource`: Elasticsearch ingest pipeline resource lifecycle managed via the entitycore Elasticsearch envelope.

### Modified Capabilities
<!-- No existing spec-level behavior changes. -->

## Impact

- Affects `internal/elasticsearch/ingest/pipeline.go` and `pipeline_test.go`.
- No API or schema breaking changes to the public Terraform interface.
- Ingest processor data sources (already PF) are untouched.
