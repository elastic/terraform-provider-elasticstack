## Why

The provider currently uses the raw `esapi` (untyped) client for all Elasticsearch API calls. The `go-elasticsearch` library now provides a fully typed API via `elasticsearch.TypedClient` (`ToTyped()`), which eliminates hand-structured request bodies, reduces runtime errors, and enables IDE-driven refactoring. We need a non-breaking bridge so that file-by-file migrations can opt-in to the typed client without disrupting existing code.

## What Changes

- Add a `typedClient *elasticsearch.TypedClient` field to `ElasticsearchScopedClient` (in `internal/clients/elasticsearch_scoped_client.go`).
- Add `GetESTypedClient()` method that lazily converts the existing `*elasticsearch.Client` to `*elasticsearch.TypedClient` via `client.ToTyped()` and caches the result (thread-safe via `sync.Once`).
- `GetESClient()` remains completely untouched — no existing consumer is affected.
- No resources or helpers are migrated yet; this is pure infrastructure/bridge code only.

## Capabilities

### New Capabilities
- `typed-client-bootstrap`: Bridge infrastructure that exposes `GetESTypedClient()` on `ElasticsearchScopedClient`, allowing subsequent migrations to use the typed client.

### Modified Capabilities
- (none — no spec-level behavior changes, purely additive infrastructure)

## Impact

- **Code**: `internal/clients/elasticsearch_scoped_client.go` only.
- **APIs**: No Terraform resource or data source behavior changes.
- **Dependencies**: Relies on existing `go-elasticsearch/v8` `ToTyped()` method.
- **Systems**: None — this is a foundational, backward-compatible change.
