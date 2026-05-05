## 1. Update ElasticsearchScopedClient

- [x] 1.1 Remove the raw `*elasticsearch.Client` field from `ElasticsearchScopedClient`
- [x] 1.2 Rename the existing `GetESTypedClient()` method to `GetESClient()` and set its return type to `*elasticsearch.TypedClient`
- [x] 1.3 Remove the old `GetESClient()` method that returned the raw client
- [x] 1.4 Rewrite `serverInfo()` to use `typedClient.Info().Do(ctx)` and unmarshal into `*types.InfoResponse`
- [x] 1.5 Update `internal/clients/elasticsearch_scoped_client_test.go` to assert on `*elasticsearch.TypedClient` returns

## 2. Delete obsolete helpers

- [x] 2.1 Removed `doFWWrite` and `doSDKWrite` from `internal/clients/elasticsearch/helpers.go`; retained `isNotFoundElasticsearchError` and `durationToMsString` which are still used across the codebase
- [x] 2.2 Verified no remaining references to `doFWWrite` or `doSDKWrite` across the codebase

## 3. Delete redundant model files

- [x] 3.1 `internal/models/ml.go` — already deleted in prior merge
- [x] 3.2 `internal/models/transform.go` — retained, types still heavily used by resource layer (out of scope)
- [x] 3.3 `internal/models/enrich.go` — retained, types still used by enrich resource (out of scope)
- [x] 3.4 Remove unused types from `internal/models/models.go` (`TimestampField` removed; `DataStreamLifecycle` retained because it is still actively used by `GetDataStreamLifecycle`; others already removed in prior merges)
- [x] 3.5 Custom types (`BuildDate`, `StringSliceOrCSV`, `Index`, `PutIndexParams`, `IndexAlias`, `LifecycleSettings`, `Downsampling`, `LogstashPipeline`, `Watch`/`PutWatch`/`WatchBody`, `APIKeyRoleDescriptor` etc.) retained
- [x] 3.6 `internal/models/ingest.go` retained — ingest processor structs are custom and have no typedapi equivalent

## 4. Fix imports and references

- [x] 4.1 Replaced remaining raw-client callers (PutIndex, GetDataStreamLifecycle) with typed client APIs
- [x] 4.2 Removed unused `esapi` imports from all Elasticsearch-scoped files
- [x] 4.3 All compilation errors resolved — `go build ./...` passes

## 5. Verify build and quality

- [x] 5.1 Build passes — `go build -buildvcs=false ./...` exits with status 0
- [x] 5.2 Lint passes — `golangci-lint-custom run --max-same-issues=0 ./...` exits with 0 issues
- [x] 5.3 Run targeted acceptance tests for affected resources (cluster settings, security, index, ML, transform, enrich, etc.) to confirm no behavioral regressions
