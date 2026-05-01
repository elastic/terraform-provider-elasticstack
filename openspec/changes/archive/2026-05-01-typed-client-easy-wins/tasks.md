## 1. Inference

- [x] 1.1 Migrate `PutInferenceEndpoint` in `internal/clients/elasticsearch/inference.go` to use `GetESTypedClient()` and `typedapi.Inference.Put()` with `types.InferenceEndpoint`
- [x] 1.2 Migrate `GetInferenceEndpoint` in `internal/clients/elasticsearch/inference.go` to use `GetESTypedClient()` and `typedapi.Inference.Get()` with `types.InferenceEndpointInfo`
- [x] 1.3 Migrate `UpdateInferenceEndpoint` in `internal/clients/elasticsearch/inference.go` to use `GetESTypedClient()` and `typedapi.Inference.Update()` with `types.InferenceEndpoint`
- [x] 1.4 Migrate `DeleteInferenceEndpoint` in `internal/clients/elasticsearch/inference.go` to use `GetESTypedClient()` and `typedapi.Inference.Delete()`
- [x] 1.5 Update `internal/elasticsearch/inference/inferenceendpoint/create.go` to compile with migrated `PutInferenceEndpoint` signature
- [x] 1.6 Update `internal/elasticsearch/inference/inferenceendpoint/read.go` to compile with migrated `GetInferenceEndpoint` signature
- [x] 1.7 Update `internal/elasticsearch/inference/inferenceendpoint/update.go` to compile with migrated `UpdateInferenceEndpoint` signature
- [x] 1.8 Update `internal/elasticsearch/inference/inferenceendpoint/delete.go` to compile with migrated `DeleteInferenceEndpoint` signature
- [x] 1.9 Remove now-unused custom `InferenceEndpoint` structs from `inference.go` if fully replaced by typed types
- [x] 1.10 Update `internal/elasticsearch/inference/inferenceendpoint/acc_test.go` to add `SkipFunc` guards for fake API key in update/plan test steps

## 2. Logstash

- [x] 2.1 Migrate `PutLogstashPipeline` in `internal/clients/elasticsearch/logstash.go` to use `GetESTypedClient()` and `typedapi.Logstash.PutPipeline()` with `types.LogstashPipeline`
- [x] 2.2 Migrate `GetLogstashPipeline` in `internal/clients/elasticsearch/logstash.go` to use `GetESTypedClient()` and `typedapi.Logstash.GetPipeline()`
- [x] 2.3 Migrate `DeleteLogstashPipeline` in `internal/clients/elasticsearch/logstash.go` to use `GetESTypedClient()` and `typedapi.Logstash.DeletePipeline()`
- [x] 2.4 Update `internal/elasticsearch/logstash/pipeline.go` to compile with migrated helper signatures
- [x] 2.5 Remove or deprecate `models.LogstashPipeline` if fully replaced by `types.LogstashPipeline`

## 3. Enrich

- [x] 3.1 Migrate `GetEnrichPolicy` in `internal/clients/elasticsearch/enrich.go` to use `GetESTypedClient()` and `typedapi.Enrich.GetPolicy()` with `types.Summary`/`types.EnrichPolicy`
- [x] 3.2 Migrate `PutEnrichPolicy` in `internal/clients/elasticsearch/enrich.go` to use `GetESTypedClient()` and `typedapi.Enrich.PutPolicy()` with `types.EnrichPolicy`
- [x] 3.3 Migrate `DeleteEnrichPolicy` in `internal/clients/elasticsearch/enrich.go` to use `GetESTypedClient()` and `typedapi.Enrich.DeletePolicy()`
- [x] 3.4 Migrate `ExecuteEnrichPolicy` in `internal/clients/elasticsearch/enrich.go` to use `GetESTypedClient()` and `typedapi.Enrich.ExecutePolicy()`
- [x] 3.5 Handle query string↔`types.Query` conversion at the boundary in enrich helpers
- [x] 3.6 Update `internal/elasticsearch/enrich/create.go` to compile with migrated `PutEnrichPolicy` signature
- [x] 3.7 Update `internal/elasticsearch/enrich/data_source.go` to compile with migrated `GetEnrichPolicy` signature
- [x] 3.8 Update `internal/elasticsearch/enrich/delete.go` to compile with migrated `DeleteEnrichPolicy` signature
- [x] 3.9 Remove now-unused custom `enrichPolicyResponse` and `enrichPoliciesResponse` structs from `enrich.go`

## 4. Watch

- [x] 4.1 Migrate `PutWatch` in `internal/clients/elasticsearch/watch.go` to use `GetESTypedClient()` and `typedapi.Watcher.PutWatch()` with `types.Watch` or `types.WatcherAction`/`types.WatcherCondition`/`types.WatcherInput`
- [x] 4.2 Migrate `PutWatchBodyJSON` in `internal/clients/elasticsearch/watch.go` to use typed API, converting raw JSON to typed request fields
- [x] 4.3 Migrate `GetWatch` in `internal/clients/elasticsearch/watch.go` to use `GetESTypedClient()` and `typedapi.Watcher.GetWatch()` with `types.Watch`
- [x] 4.4 Migrate `DeleteWatch` in `internal/clients/elasticsearch/watch.go` to use `GetESTypedClient()` and `typedapi.Watcher.DeleteWatch()`, preserving 404-as-success semantics
- [x] 4.5 Update `internal/elasticsearch/watcher/watch/create.go` to compile with migrated `PutWatch`/`PutWatchBodyJSON` signatures
- [x] 4.6 Update `internal/elasticsearch/watcher/watch/read.go` to compile with migrated `GetWatch` signature
- [x] 4.7 Update `internal/elasticsearch/watcher/watch/delete.go` to compile with migrated `DeleteWatch` signature
- [x] 4.8 Evaluate whether `models.Watch`, `models.PutWatch`, and `models.WatchBody` can be removed or must be retained for schema-layer mapping

## 5. Cleanup and Verification

- [x] 5.1 Run `make build` and confirm zero compile errors across the entire codebase
- [x] 5.2 Run `make check-lint` and resolve any new lint warnings introduced by typed API usage
- [x] 5.3 Run `go test ./internal/clients/elasticsearch/...` to verify unit tests pass
- [x] 5.4 Confirm no remaining `GetESClient()` calls exist in the four migrated helper files
- [x] 5.5 Clean up any unused imports in the migrated helper files
- [x] 5.6 If `models.LogstashPipeline`, `models.Watch`, `models.PutWatch`, or `models.WatchBody` are fully redundant, remove them from `internal/models/models.go`
- [x] 5.7 Run targeted acceptance tests for inference, logstash, enrich, and watcher packages where possible
