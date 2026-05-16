## Why

`elasticstack_elasticsearch_watch` is already implemented on the Terraform Plugin Framework but still embeds `*entitycore.ResourceBase` directly. It duplicates the standard Read/Delete prelude (state decode → composite ID parse → scoped client resolution → API call) and manually declares the `elasticsearch_connection` block in its schema. Migrating it to the `entitycore.ElasticsearchResource[Data]` envelope eliminates that duplication and brings it in line with other envelope-adopted resources.

The watch resource has straightforward CRUD: Create and Update both call `PutWatch` then `read`; `read` already exists as a helper that can become the envelope `readFunc`; Delete calls `DeleteWatch`. No config-derived post-processing is needed after read-back beyond the existing actions redaction logic, which lives inside `fromAPIModel` and is already isolated.

## What Changes

- Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]` in `internal/elasticsearch/watcher/watch/resource.go`.
- Add `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` to the `Data` struct.
- Convert the existing `read(ctx, data Data)` helper into a package-level `readWatch` callback with the envelope signature `(ctx, client, resourceID, state) (Data, bool, diag.Diagnostics)`.
- Extract `createWatch` and `updateWatch` callbacks from the existing `Create` and `Update` method bodies. Both callbacks call `PutWatch`, compute the composite id, and return the model. The envelope handles the subsequent read-back.
- Extract `deleteWatch` from the existing `Delete` method body.
- Strip the `elasticsearch_connection` block from the schema factory; the envelope injects it.
- Preserve `ImportState` passthrough on the concrete `watchResource` type.
- Preserve actions redaction logic in `fromAPIModel`; no changes needed.

## Capabilities

### New Capabilities
- `elasticsearch-watch-via-envelope`: Migrate `elasticstack_elasticsearch_watch` to the entitycore resource envelope.

### Modified Capabilities
<!-- None. External schema, APIs, and behavior are preserved. -->

## Impact

- Refactored code: `internal/elasticsearch/watcher/watch/{resource.go,read.go,delete.go,create.go,update.go,schema.go,models.go}`.
- No public API changes. No Terraform schema changes.
- Acceptance tests for watch must continue to pass without modification.
