## Why

`elasticstack_elasticsearch_index` is the most complex Plugin Framework resource in the provider. It currently embeds `*entitycore.ResourceBase` and implements its own `Schema`, `Read`, `Create`, `Update`, and `Delete`. The Read and Delete preludes are standard boilerplate that the entitycore envelope already centralizes. Migrating the resource to the envelope removes this duplication, lets the resource focus on its unique concerns (index creation with adoption, alias/settings/mappings reconciliation, deletion protection, and mappings plan modifiers), and keeps diagnostic handling consistent with other envelope-backed resources.

## What Changes

- Migrate `internal/elasticsearch/index/index` from `*entitycore.ResourceBase` to `*entitycore.ElasticsearchResource[tfModel]`.
- Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` to `tfModel`.
- Convert the schema definition to a factory without `elasticsearch_connection`.
- Provide a real `readIndex` callback (the existing `readIndex` helper already has the correct signature minus the client parameter, which the envelope supplies).
- Provide a real `deleteIndex` callback that checks `deletion_protection` before calling the Delete Index API.
- Use `PlaceholderElasticsearchWriteCallbacks` for create and update because both have custom flows that need access to plan/state details the envelope callback contract does not expose:
  - Create handles `use_existing` adoption, date-math names, and alias/settings/mappings reconciliation for adopted indices.
  - Update derives the concrete index name from state (not plan) and reconciles aliases, settings, and mappings independently.
- Override `Create` and `Update` on the concrete `Resource` type.
- Preserve `ImportState` on the concrete type.
- No schema changes, no external behavior changes.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `elasticsearch-index`: Internal implementation migrated to entitycore envelope.

## Impact

- Refactored code: `internal/elasticsearch/index/index/resource.go`, `read.go`, `create.go`, `update.go`, `delete.go`, `schema.go`, `models.go`.
- Acceptance tests for `index` must continue to pass without modification.
