## Context

The `typed-client-bootstrap` change already added `GetESTypedClient()` to `ElasticsearchScopedClient`, returning a cached `*elasticsearch.TypedClient`. All Elasticsearch API calls in the provider currently go through the raw `esapi` client returned by `GetESClient()`, which requires manual request construction, JSON marshaling, and response body parsing.

This change targets four helper files in `internal/clients/elasticsearch/` that are small, self-contained, and have direct typed API equivalents in `go-elasticsearch/v8`. Migrating these first establishes patterns for larger, more complex migrations in subsequent phases.

## Goals / Non-Goals

**Goals:**
- Convert `inference.go`, `logstash.go`, `enrich.go`, and `watch.go` to use `GetESTypedClient()` instead of `GetESClient()`.
- Eliminate manual `json.Marshal`/`json.Unmarshal` in these helpers by using typed request/response structs.
- Reduce or eliminate custom model structs where typed API types (`typedapi/types.*`) provide equivalent or better coverage.
- Update all downstream resource callers so they compile and behave identically.
- Ensure every migrated helper preserves existing error semantics (404 = not found, etc.).

**Non-Goals:**
- Do not migrate resource files themselves beyond the minimal changes needed to compile with updated helper signatures.
- Do not remove `GetESClient()` or change `ElasticsearchScopedClient` construction.
- Do not introduce new Terraform resources, data sources, or schema changes.
- Do not modify acceptance test behavior (tests may be updated in `typed-client-acceptance-tests`).

## Decisions

### 1. Migrate helper signatures to accept `*elasticsearch.TypedClient` directly

**Chosen:** Each helper function will obtain the typed client from `apiClient.GetESTypedClient()` at the top of the function, then use it for all API calls.

**Rationale:** Keeps the change localized to each helper. Callers in resource files don't need to change how they pass `apiClient`.

**Alternative considered:** Changing resource files to call `GetESTypedClient()` themselves and pass the typed client down. Rejected because it would proliferate changes across many files instead of keeping them in the four targeted helpers.

### 2. Use typed API request/response types where they match provider needs

**Chosen:**
- For **inference**: `types.InferenceEndpoint` (PUT/update request) and `types.InferenceEndpointInfo` (GET response).
- For **logstash**: `types.LogstashPipeline` (GET/PUT response/request body is `map[string]types.LogstashPipeline`).
- For **enrich**: `types.EnrichPolicy` (GET/PUT) and `types.Summary` (GET response wrapper).
- For **watch**: `types.Watch` (GET response) and `types.WatcherAction`/`types.WatcherCondition`/`types.WatcherInput` (PUT request fields).

**Rationale:** Eliminates custom structs and manual JSON handling. The typed types are generated from the Elasticsearch specification and stay in sync with the server.

**Alternative considered:** Keeping custom structs and only using the typed client for transport. Rejected because it defeats the purpose of the migration — we want type safety end-to-end.

### 3. Preserve `map[string]any` fields where typed API is too rigid

**Chosen:** For deeply nested or highly dynamic fields (e.g., watch `metadata`, inference `service_settings`), continue using `map[string]any` or JSON strings at the Terraform schema layer, and convert to/from typed types at the API boundary.

**Rationale:** The Terraform provider's schema uses `types.String` with JSON normalization for these fields. Converting the entire provider to strongly-typed nested structs would require schema changes, which are out of scope for this migration.

**Alternative considered:** Fully typed nested structs throughout. Rejected because it would require breaking schema changes and massive resource file rework.

### 4. Map 404 semantics identically

**Chosen:** Where raw code checked `res.StatusCode == http.StatusNotFound`, use the typed API's error handling pattern. Most typed `Get`/`Delete` calls return an error on 404; helpers should detect this and return `nil` (for Get) or empty diagnostics (for Delete), preserving existing behavior.

**Rationale:** Terraform resources rely on "not found" being a non-error during read and delete operations. Changing this would break state refresh and destroy.

**Alternative considered:** Letting typed API errors propagate unchanged. Rejected because typed client 404s may surface as structured errors that resources don't expect.

### 5. Keep `InferenceEndpoint` custom type if typed API lacks task-type support

**Chosen:** The current `InferenceEndpoint` struct includes `TaskType` and `InferenceID` as top-level fields, but `types.InferenceEndpoint` does not have a `TaskType` field (it's a path parameter in the typed API). The helper functions will continue to accept task type and inference ID as function parameters, and construct the typed request with the appropriate path parameter.

**Rationale:** The typed API separates path parameters from the request body. The provider's model currently conflates them for convenience.

**Alternative considered:** Changing resource schemas to match typed API structure. Rejected because it would be a breaking schema change.

## Risks / Trade-offs

- **[Risk]** Typed API `types.EnrichPolicy.Query` is `*types.Query` (strongly typed), but the provider stores query as a JSON string.
  - **Mitigation:** Use `json.Marshal`/`json.Unmarshal` at the boundary to convert between `*types.Query` and `string`. This is a small, localized conversion in the enrich helper.

- **[Risk]** Typed API `types.Watch` has different field shapes than the provider's `models.Watch` (e.g., `WatchID` is not part of the typed struct).
  - **Mitigation:** Extract `WatchID` from the function parameter or response wrapper (`GetWatchResponse.Id_`), and map fields individually rather than whole-struct replacement.

- **[Risk]** `DeleteWatch` raw code returns success on 404; typed API may return an error.
  - **Mitigation:** Check the typed API error for 404 semantics (e.g., `elastic.IsNotFound(err)` or status-code extraction) and return empty diagnostics, matching existing behavior.

- **[Risk]** Logstash pipeline typed response is `map[string]types.LogstashPipeline`, which doesn't include `PipelineID` as a struct field.
  - **Mitigation:** Map the map key to `PipelineID` after decoding, just as the current code does.

- **[Risk]** Compilation errors in downstream resource files due to changed helper return types or model structs.
  - **Mitigation:** After migrating each helper, immediately update and compile its callers. Run `make build` after each file group.

## Migration Plan

1. Migrate `inference.go` helpers and compile `internal/elasticsearch/inference/inferenceendpoint/*.go`.
2. Migrate `logstash.go` helpers and compile `internal/elasticsearch/logstash/pipeline.go`.
3. Migrate `enrich.go` helpers and compile `internal/elasticsearch/enrich/*.go`.
4. Migrate `watch.go` helpers and compile `internal/elasticsearch/watcher/watch/*.go`.
5. Run `make build` and `make check-lint` for the full project.
6. Remove any now-unused custom model structs from `internal/models/models.go`.
7. Verify no remaining references to the old helper signatures.
