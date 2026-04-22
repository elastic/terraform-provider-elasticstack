## Why

`elasticstack_elasticsearch_index` currently rejects date math index names and assumes the configured `name` is also the stable server-side identity. For date math names that resolve to a concrete index during creation, that assumption causes drift-prone state handling and prevents reliable read, update, and delete operations.

## What Changes

- Add support for plain Elasticsearch date math index names on `elasticstack_elasticsearch_index` without weakening validation for normal static names.
- URI-encode accepted date math names inside the provider before calling the Create Index API.
- Introduce a computed `concrete_name` attribute that tracks the concrete index name Elasticsearch created from the configured `name`.
- Update identity and CRUD behavior so the resource keeps the configured `name` as user intent while targeting the persisted concrete index name for read, update, and delete operations.
- Add focused validation and regression coverage for static names, plain date math names, provider-side encoding, create/read stability, and update behavior after creating from a date math expression.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `elasticsearch-index`: extend index name validation and resource identity handling to support plain date math index names without perpetual drift.

## Impact

- `internal/elasticsearch/index/index/schema.go` for split name validation and the new computed `concrete_name` attribute
- `internal/elasticsearch/index/index/create.go`, `read.go`, `update.go`, `delete.go`, and `models.go` for concrete-name-aware state and CRUD behavior
- `internal/clients/elasticsearch/index.go` for create/get helpers that need the concrete name returned by Elasticsearch
- `internal/elasticsearch/index/index/*test.go` for focused unit and acceptance coverage
- `openspec/specs/elasticsearch-index/spec.md` via the delta spec for the updated resource contract
