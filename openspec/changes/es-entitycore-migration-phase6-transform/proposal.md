## Why

`elasticstack_elasticsearch_transform` is the largest remaining Plugin SDK resource in the provider (905 LOC) and one of the most complex. It manages the full lifecycle of Elasticsearch transforms including version-gated settings, start/stop state management, JSON field mapping, and `ExactlyOneOf` validation between `pivot` and `latest`. Rewriting it to Plugin Framework and adopting the entitycore envelope removes SDK tech debt, centralizes standard preludes, and brings the resource in line with modern provider patterns.

## What Changes

- Rewrite `internal/elasticsearch/transform/transform.go` from Plugin SDK to Plugin Framework.
- Introduce a new PF-based model (`tfModel`), schema factory, and conversion helpers in the same package.
- Migrate to `*entitycore.ElasticsearchResource[tfModel]` so the envelope owns Read, Delete, Schema, Configure, and Metadata.
- Provide a real `readTransform` callback that fetches the transform definition and stats.
- Provide a real `deleteTransform` callback that deletes with `force=true`.
- Provide a real `createTransform` callback that calls Put Transform and optionally starts the transform when `enabled` is true.
- Override `Update` on the concrete type because it needs to compare the prior state with the plan to detect `enabled` changes and issue Start/Stop Transform calls after the Update Transform API call.
- Preserve all existing externally-visible behavior: version-gated settings, `pivot`/`latest` mutual exclusivity, defer validation, timeout handling, JSON diff suppression, start/stop lifecycle, and import support.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `elasticsearch-transform`: Implementation rewritten from Plugin SDK to Plugin Framework + entitycore envelope.

## Impact

- Major rewrite of `internal/elasticsearch/transform/transform.go` and supporting files.
- New PF model, schema, and conversion code.
- No Terraform schema changes (all attributes preserved).
- Acceptance tests must continue to pass without modification.
