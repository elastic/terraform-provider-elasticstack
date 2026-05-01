## Context

The provider's ML subsystem currently relies on the raw `esapi` client (`GetESClient()`) for every Elasticsearch ML API call. Request bodies are hand-marshaled to JSON, and responses are decoded into custom model structs defined in `internal/models/ml.go`. The typed client (`GetESTypedClient()`) was introduced by `typed-client-bootstrap` and provides strongly-typed `types.*` structs for every ML endpoint.

Files involved:
- **`internal/clients/elasticsearch/ml_job.go`** — 10 helper functions that call raw `esapi` ML endpoints.
- **`internal/elasticsearch/ml/anomalydetectionjob/*.go`** — resource methods that call raw `esClient.ML.PutJob`, `GetJobs`, `UpdateJob`, `CloseJob`, `DeleteJob` directly.
- **`internal/elasticsearch/ml/datafeed/*.go`** — resource methods that consume `PutDatafeed`, `GetDatafeed`, `UpdateDatafeed`, `DeleteDatafeed`, `StopDatafeed`, `StartDatafeed`, and `GetDatafeedStats` helpers.
- **`internal/elasticsearch/ml/jobstate/*.go`** — resource methods that consume `OpenMLJob`, `CloseMLJob`, and `GetMLJobStats` helpers.
- **`internal/elasticsearch/ml/datafeed_state/*.go`** — resource methods that consume `GetDatafeedStats`, `StartDatafeed`, and `StopDatafeed` helpers.
- **`internal/models/ml.go`** — custom ML model structs that shadow typed API structs.

## Goals / Non-Goals

**Goals:**
- Convert every raw `esapi` ML API call in `ml_job.go` to use the typed client (`elasticsearch.TypedClient`).
- Convert every raw `esapi` ML API call in `anomalydetectionjob/*.go` to use the typed client.
- Remove or replace custom model structs in `internal/models/ml.go` with typed API equivalents.
- Ensure all downstream ML resources (`datafeed`, `jobstate`, `datafeed_state`) continue to compile and pass acceptance tests.
- Preserve all existing error-handling behavior (not-found handling, force flags, timeouts, diagnostics).

**Non-Goals:**
- Changing ML resource schemas, Terraform behaviors, or force-new rules.
- Adding new ML resources or data sources.
- Migrating non-ML Elasticsearch helpers or resources.
- Renaming or restructuring package boundaries beyond what the typed client requires.

## Decisions

**1. Replace `GetESClient()` with `GetESTypedClient()` in all ML helpers**
- **Rationale**: The typed client provides the same transport and endpoints but with compile-time type safety. `GetESTypedClient()` is cached (via `sync.Once`) so there is no per-call overhead.
- **Alternative considered**: Keeping `ml_job.go` on the raw client and only migrating resource files. Rejected because the helpers are the primary API surface and keeping them raw would leave the main migration unfinished.

**2. Delete custom model structs and use `types.*` directly**
- **Rationale**: `go-elasticsearch/v8` already defines `types.Datafeed`, `types.Job`, `types.DatafeedStats`, etc. Maintaining parallel structs in `internal/models/ml.go` is redundant and invites drift.
- **Alternative considered**: Keeping custom structs as thin wrappers around typed structs. Rejected because it adds indirection with no benefit; the typed structs already have the correct `json` tags and optional field shapes.

**3. Update `ml_job.go` function signatures to accept typed request structs**
- **Rationale**: Typed API methods accept strongly-typed builders (e.g., `typedapi.ML.PutDatafeed().Request(&types.Datafeed{...})`). Passing raw `[]byte` bodies would defeat the purpose. Helpers will construct the typed request inline or accept typed structs from callers.
- **Alternative considered**: Keeping `[]byte` signatures and doing an extra marshal/unmarshal step. Rejected because it would negate the type-safety benefits.

**4. Keep the same diagnostics and error-handling patterns**
- **Rationale**: The typed client's `.Do(ctx)` returns `(Response, error)`. The first `error` is transport-level (same as raw client). Response-level errors are checked via `.IsError()`. This maps cleanly to existing `diag.Diagnostics` patterns.
- **Alternative considered**: Introducing a new shared wrapper for typed API error handling. Rejected because each helper has slightly different not-found semantics; explicit handling per function is clearer.

**5. Migrate anomaly detection job resource inline with helpers**
- **Rationale**: The anomaly detection job resource currently calls `esClient.ML.PutJob`, `GetJobs`, `UpdateJob`, `CloseJob`, and `DeleteJob` directly, bypassing `ml_job.go`. Migrating these simultaneously avoids a half-typed state in the ML subsystem and keeps the scope of the change coherent.
- **Alternative considered**: Splitting into two changes (helpers first, anomaly detection job second). Rejected because the anomaly detection job resource is tightly coupled to the same APIs and would require a second, nearly identical review pass.

## Risks / Trade-offs

- **[Risk]** Typed API structs may have slightly different field types (pointers, slices, or custom types) compared to custom models, causing subtle serialization differences.
  - **Mitigation**: Compare field-by-field with existing models before removal; run the full ML acceptance test suite to verify round-tripping.
- **[Risk]** Typed API response nesting may differ (e.g., `GetDatafeedsResponse.Datafeeds` vs. custom `[]models.Datafeed`).
  - **Mitigation**: Inspect the generated typed API response struct in `go-elasticsearch` and update index lookups accordingly.
- **[Risk]** `go-elasticsearch` typed API builders may omit fields when set to zero values differently than manual JSON marshaling.
  - **Mitigation**: Add explicit nil checks before setting builder fields; ensure optional+computed fields are omitted when unknown/null.
- **[Risk]** Mixing typed and untyped clients in the same resource during review.
  - **Mitigation**: Verify that every ML file compiles and that no raw `esapi` imports remain in the migrated packages.

## Migration Plan

1. Update `ml_job.go` helpers to use `GetESTypedClient()` and typed API types.
2. Update `anomalydetectionjob/*.go` to use `GetESTypedClient()` and typed API types.
3. Remove unused custom model structs from `internal/models/ml.go`.
4. Ensure `datafeed`, `jobstate`, and `datafeed_state` packages compile against updated helper signatures.
5. Run `make build` and `make check-lint`.
6. Run ML acceptance tests (`go test ./internal/elasticsearch/ml/...`) against a live cluster.

## Open Questions

- None.
