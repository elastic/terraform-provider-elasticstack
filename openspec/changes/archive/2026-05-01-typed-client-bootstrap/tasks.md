## 1. Core Implementation

- [x] 1.1 Add `typedClient *elasticsearch.TypedClient` field to `ElasticsearchScopedClient` struct in `internal/clients/elasticsearch_scoped_client.go`
- [x] 1.2 Add `typedClientOnce sync.Once` field to `ElasticsearchScopedClient` struct for thread-safe lazy initialization
- [x] 1.3 Implement `GetESTypedClient() *elasticsearch.TypedClient` method with lazy `ToTyped()` initialization and caching
- [x] 1.4 Add Go doc comments to `GetESTypedClient()` explaining behavior, thread-safety, and that it shares the underlying transport

## 2. Verification

- [x] 2.1 Run `make build` to ensure the project compiles without errors
- [x] 2.2 Run `make check-lint` to ensure code passes lint checks
- [x] 2.3 Verify `GetESClient()` remains unchanged and unaffected by the new code
- [x] 2.4 Confirm `go mod tidy` does not introduce new dependencies (change uses existing `go-elasticsearch/v8`)
