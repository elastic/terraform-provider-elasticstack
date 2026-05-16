## Why

Several Plugin Framework Elasticsearch resources duplicate thin `Create` and `Update` methods that only delegate to a shared upsert implementation. Centralizing this lifecycle behavior in `NewElasticsearchResource` keeps the resource envelope responsible for common Terraform request handling and prevents further drift across migrated resources.

## What Changes

- Extend `NewElasticsearchResource[T]` so callers provide required create and update callbacks in addition to schema, read, and delete callbacks.
- Have the Elasticsearch resource envelope implement `Create` and `Update` by decoding the planned model, deriving the write resource ID from the model, resolving the scoped Elasticsearch client, invoking the relevant callback with the resource ID, and persisting the returned model to state.
- Migrate the four duplicated issue examples to pass create and update callbacks to the envelope, using the same callback for both operations where create and update are currently identical.
- Update the `entitycore-resource-envelope` specification to describe the envelope as a complete `resource.Resource` for resources that fit the common Elasticsearch lifecycle.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `entitycore-resource-envelope`: `NewElasticsearchResource` will own Create and Update behavior through required callbacks instead of requiring concrete resources to implement thin wrappers.

## Impact

- Affected code: `internal/entitycore`, the four issue examples in `internal/elasticsearch/security/role`, `internal/elasticsearch/security/rolemapping`, `internal/elasticsearch/security/systemuser`, and `internal/elasticsearch/cluster/script`, plus any existing `NewElasticsearchResource` call sites that need constructor or model constraint updates.
- Affected tests: entitycore resource envelope unit tests and targeted tests for the migrated resources where available.
- Public Terraform behavior should remain unchanged; this is an internal Plugin Framework resource abstraction refactor.
