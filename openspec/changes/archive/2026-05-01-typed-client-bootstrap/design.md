## Context

`ElasticsearchScopedClient` (in `internal/clients/elasticsearch_scoped_client.go`) is the canonical typed surface for Elasticsearch operations. It currently exposes `GetESClient()` which returns `*elasticsearch.Client` — the raw, untyped `esapi` client. All provider resources and helpers build requests manually using this client.

The `go-elasticsearch/v8` library now provides `client.ToTyped()` which returns `*elasticsearch.TypedClient`. This client uses strongly-typed structs for every request/response, eliminating `map[string]interface{}` boilerplate and runtime shape mismatches. However, switching the entire provider at once would be high-risk and hard to review. We need an incremental migration path.

## Goals / Non-Goals

**Goals:**
- Provide a bridge method `GetESTypedClient()` on `ElasticsearchScopedClient` that returns `*elasticsearch.TypedClient`.
- Ensure the typed client is lazily initialized and cached (thread-safe).
- Keep `GetESClient()` untouched so existing code continues to work without modification.
- Enable subsequent changes (e.g., `typed-client-index`, `typed-client-cluster`) to migrate individual files incrementally.

**Non-Goals:**
- Migrating any existing resource or helper to the typed client (future changes).
- Removing or deprecating `GetESClient()`.
- Changing provider configuration, client construction, or authentication flows.
- Adding new Terraform resources or data sources.

## Decisions

**1. Lazy initialization with `sync.Once`**
- **Rationale**: `ToTyped()` is cheap but not free. `sync.Once` guarantees exactly-once initialization without holding a mutex for every call. The typed client is stateless and safe for concurrent use once created.
- **Alternative considered**: `sync.Mutex` + nil check. Rejected because `sync.Once` is simpler and idiomatic for this pattern.

**2. Cache the typed client as a struct field**
- **Rationale**: Once created, the typed client is bound to the same underlying `*elasticsearch.Client` (same transport, same endpoints). Caching avoids repeated `ToTyped()` calls and keeps the API ergonomic.
- **Alternative considered**: Computing `ToTyped()` on every `GetESTypedClient()` call. Rejected because it creates unnecessary garbage for no benefit.

**3. Add the field directly to `ElasticsearchScopedClient` rather than a wrapper**
- **Rationale**: The scoped client is already the abstraction layer. Adding a field keeps the surface simple and avoids proliferating types.
- **Alternative considered**: A separate `TypedElasticsearchScopedClient` wrapper. Rejected because it would require updating every call site that constructs or accepts the scoped client.

**4. `GetESTypedClient()` returns `*elasticsearch.TypedClient` (no error)**
- **Rationale**: `client.ToTyped()` returns `*TypedClient` with no error — it is a pure wrapper around the existing client. Adding an error return to `GetESTypedClient()` would be misleading (there is no failure mode distinct from `GetESClient()`). Callers that need error handling can validate the underlying client first via `GetESClient()`.
- **Alternative considered**: Returning `(*elasticsearch.TypedClient, error)` for API symmetry with `GetESClient()`. Rejected because it forces all callers to handle an error that can never occur.

## Risks / Trade-offs

- **[Risk]** `ToTyped()` re-runs the product check on the typed client's first request, which adds marginal latency on first use.
  - **Mitigation**: The product check is fast and only runs once per cached typed client. Document this behavior in the method comment.
- **[Risk]** Future `go-elasticsearch` versions could change `ToTyped()` behavior or signature.
  - **Mitigation**: The provider pins `go-elasticsearch/v8`. Any upgrade would be a dedicated change with regression testing.
- **[Risk]** Developers might inconsistently mix typed and untyped APIs in the same resource.
  - **Mitigation**: This is expected during the transition. Each migration change should convert a resource fully. Code review enforces consistency.
