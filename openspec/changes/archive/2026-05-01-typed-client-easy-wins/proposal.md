## Why

The provider-wide typed-client migration is underway. The `go-elasticsearch/v8` Typed API (`elasticsearch.TypedClient` via `ToTyped()`) provides strongly-typed request/response structs that eliminate hand-written JSON marshaling, reduce runtime errors, and enable IDE-driven refactoring. Phase 2 targets four small, self-contained client helper files that have straightforward typed API equivalents, making them ideal early migration candidates.

## What Changes

Migrate the following four client helper files in `internal/clients/elasticsearch/` from raw `esapi` to the typed client (`GetESTypedClient()`):

1. **`inference.go`** — `PutInferenceEndpoint`, `GetInferenceEndpoint`, `UpdateInferenceEndpoint`, `DeleteInferenceEndpoint`
   - Replace custom `InferenceEndpoint` types with `types.InferenceEndpointInfo`/`types.InferenceEndpoint` where possible.
   - Replace `doFWWrite` and manual JSON handling with typed API calls.

2. **`logstash.go`** — `PutLogstashPipeline`, `GetLogstashPipeline`, `DeleteLogstashPipeline`
   - Replace custom `models.LogstashPipeline` usage with `types.LogstashPipeline` where possible.
   - Replace manual JSON decode with typed `Logstash.GetPipeline`/`PutPipeline`/`DeletePipeline`.

3. **`enrich.go`** — `GetEnrichPolicy`, `PutEnrichPolicy`, `DeleteEnrichPolicy`, `ExecuteEnrichPolicy`
   - Replace custom `enrichPolicyResponse` unmarshal with typed `EnrichPolicy` from `types.EnrichPolicy`.
   - Handle query string↔object mapping via `types.Query` or raw JSON as needed.
   - Replace raw `esapi` calls with `typedapi.Enrich.*` equivalents.

4. **`watch.go`** — `PutWatch`, `PutWatchBodyJSON`, `GetWatch`, `DeleteWatch`
   - Replace custom `models.Watch`/`models.PutWatch`/`models.WatchBody` with typed `types.Watch` and `types.WatcherAction` equivalents where possible.
   - Replace manual JSON marshal/unmarshal with typed `Watcher.PutWatch`/`GetWatch`/`DeleteWatch`.

## Capabilities

### New Capabilities
- _(none — no new provider capabilities introduced)_

### Modified Capabilities
- _(none — no spec-level requirement changes; this is a pure implementation migration)_

## Impact

- **Code**: `internal/clients/elasticsearch/inference.go`, `internal/clients/elasticsearch/logstash.go`, `internal/clients/elasticsearch/enrich.go`, `internal/clients/elasticsearch/watch.go`.
- **Callers**: `internal/elasticsearch/inference/inferenceendpoint/*.go`, `internal/elasticsearch/logstash/pipeline.go`, `internal/elasticsearch/enrich/*.go`, `internal/elasticsearch/watcher/watch/*.go`.
- **Models**: `internal/models/models.go` — custom `LogstashPipeline`, `Watch`, `PutWatch`, `WatchBody` may become redundant if fully replaced by typed types.
- **APIs**: No Terraform resource or data source behavior changes; all existing acceptance tests should continue to pass.
- **Dependencies**: Relies on the existing `typed-client-bootstrap` bridge (`GetESTypedClient()`).
