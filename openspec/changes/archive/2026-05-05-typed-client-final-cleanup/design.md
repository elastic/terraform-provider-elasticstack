## Context

All prior typed-client migration phases have converted every Elasticsearch resource and helper to the `go-elasticsearch` Typed API. The `ElasticsearchScopedClient` currently carries both a raw `*elasticsearch.Client` field and a `typedClient` field (added in `typed-client-bootstrap`), exposing `GetESClient()` (raw) and `GetESTypedClient()` (typed). Because every consumer now uses the typed path, the raw client accessor and its supporting infrastructure are dead code.

In parallel, `internal/models/` contains large hand-rolled structs (`ClusterInfo`, `User`, `Role`, `Datafeed`, `Transform`, etc.) that were created to unmarshal raw `esapi` responses. Since the typed client generates equivalent structs under `go-elasticsearch/v8/typedapi/types`, these provider-local models are also unused.

## Goals / Non-Goals

**Goals:**
- Remove the raw `elasticsearch` field from `ElasticsearchScopedClient` and make `GetESClient()` return `*elasticsearch.TypedClient` directly.
- Delete `internal/clients/elasticsearch/helpers.go` (`doFWWrite`, `doSDKWrite`).
- Delete redundant model files (`models.go` leftovers, `ml.go`, `transform.go`, `enrich.go`) from `internal/models/`.
- Update imports site-wide to remove unused raw-client paths and add `typedapi/types` where needed.
- Keep the build green and all tests passing.

**Non-Goals:**
- Migrating any resource or helper that is not already on the typed client (handled in earlier phases).
- Changing Terraform resource schemas or behavior.
- Modifying Kibana/Fleet/Observability models — those are out of scope for this change.
- Removing `internal/models/ingest.go` — ingest processor structs are custom and have no typedapi equivalent.

## Decisions

**1. Change `GetESClient()` return type to `*elasticsearch.TypedClient` rather than introducing a new method name**
- **Rationale**: By this phase every caller already expects a typed client. Renaming the method would create a larger diff across the entire codebase. Changing the return type keeps the call sites identical; only the method body and the callers' local variable types need adjustment.
- **Alternative considered**: Keep `GetESClient()` returning the raw client and rename `GetESTypedClient()` to `GetESClient()` with a phased deprecation. Rejected because there are zero remaining raw consumers — no deprecation period is needed.

**2. Delete `helpers.go` entirely**
- **Rationale**: `doFWWrite` and `doSDKWrite` exist solely to bridge the raw client (marshal body → call fn → check response). Typed API methods accept structs directly and return typed errors, making both functions obsolete.
- **Alternative considered**: Keep the file but empty it. Rejected because an empty file with a package declaration is noise; Go does not allow completely empty files.

**3. Retain `BuildDate` and a minimal `ClusterInfo` equivalent if needed for `serverInfo()`**
- **Rationale**: `ElasticsearchScopedClient.serverInfo()` currently unmarshals `esapi` `Info()` JSON into `models.ClusterInfo` to extract the cluster UUID and version. After removing the raw client, this method must switch to the typed client's `Info().Do(ctx)`, which returns `*types.InfoResponse`. The `ClusterInfo` model can therefore be removed and the method can use `types.InfoResponse` directly.
- **Alternative considered**: Keeping a slimmed-down `ClusterInfo` struct. Rejected because `types.InfoResponse` from the typed API already has the same fields (`ClusterName`, `ClusterUUID`, `Version.Number`, etc.).

**4. Remove model files in bulk rather than one per capability**
- **Rationale**: The model deletions are tightly coupled — they are all driven by the same "typedapi now provides these types" reason. Splitting them into separate changes would create unnecessary coordination overhead.
- **Alternative considered**: Deleting `ml.go` in `typed-client-ml`, `transform.go` in `typed-client-transform`, etc. Rejected because those earlier phases focused on migrating the API calls; leaving model cleanup until the end is cleaner and prevents compilation breaks if a model is still referenced by another helper.

## Risks / Trade-offs

- **[Risk]** A helper or test file outside the obvious directories may still import a model being deleted.
  - **Mitigation**: Run `make build` and `go test ./...` locally after deletions. The compiler will surface any remaining references immediately.
- **[Risk]** `serverInfo()` uses raw `esapi.Info` to populate `models.ClusterInfo`; removing the raw client requires rewriting this method.
  - **Mitigation**: This is a small, well-scoped change: replace `esClient.Info(...)` with `typedClient.Info().Do(ctx)` and decode into `*types.InfoResponse`. Add unit-test coverage for the new path if not already present.
- **[Risk]** `doFWWrite` / `doSDKWrite` might still be imported by a helper that was supposed to migrate but slipped through.
  - **Mitigation**: Deleting the file is itself the verification — any remaining import will fail compilation.

## Migration Plan

1. Update `ElasticsearchScopedClient` — remove raw field, retarget `GetESClient()` to typed client, rewrite `serverInfo()` to use typed API.
2. Update unit tests in `elasticsearch_scoped_client_test.go` to assert on typed client returns.
3. Delete `internal/clients/elasticsearch/helpers.go`.
4. Delete redundant model files (`internal/models/ml.go`, `internal/models/transform.go`, `internal/models/enrich.go`, strip unused types from `internal/models/models.go`).
5. Run `make build` to identify any remaining broken references.
6. Fix any import or reference issues surfaced by the compiler.
7. Run `make check-lint`.
8. Run targeted acceptance tests for affected resources (cluster settings, security, index, ML, transform, enrich, etc.) to confirm no behavioral regressions.

## Open Questions

- (none — design is straightforward given the prior migration phases)
