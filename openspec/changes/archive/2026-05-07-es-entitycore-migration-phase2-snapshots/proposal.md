## Why

Migrate the two snapshot-related Plugin SDK resources to the Plugin Framework and entitycore envelope. Both `snapshot_lifecycle` and `snapshot_repository` share the snapshot domain but are functionally independent, making them a natural batch.

## What Changes

- `elasticstack_elasticsearch_snapshot_lifecycle` (`internal/elasticsearch/cluster/slm.go`)
  - SDK→PF. Complex schema for cron schedule, indices, retention, etc.
- `elasticstack_elasticsearch_snapshot_repository` (`internal/elasticsearch/cluster/snapshot_repository.go`)
  - SDK→PF. Polymorphic per-type sub-schemas (fs, url, azure, gcs, s3).

For each:
- Rewrite schema from `*schema.Resource` to PF `schema.Schema`.
- Convert models to PF types with `GetID() / GetResourceID() / GetElasticsearchConnection()`.
- Wire envelope callbacks for Read/Delete/Create/Update.
- Preserve SDK-side list-to-set semantics where they affect behavior.

## Capabilities

### New Capabilities
- `snapshot-lifecycle-resource-via-envelope`
- `snapshot-repository-resource-via-envelope`

### Modified Capabilities
<!-- None. -->

## Impact

- Two files in `internal/elasticsearch/cluster/` plus tests.
- `snapshot_repository` polymorphic sub-schemas are the main complexity risk.
