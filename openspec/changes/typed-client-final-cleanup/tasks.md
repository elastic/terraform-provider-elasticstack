## 1. Update ElasticsearchScopedClient

- [ ] 1.1 Remove the raw `*elasticsearch.Client` field from `ElasticsearchScopedClient`
- [ ] 1.2 Rename the existing `GetESTypedClient()` method to `GetESClient()` and set its return type to `*elasticsearch.TypedClient`
- [ ] 1.3 Remove the old `GetESClient()` method that returned the raw client
- [ ] 1.4 Rewrite `serverInfo()` to use `typedClient.Info().Do(ctx)` and unmarshal into `*types.InfoResponse`
- [ ] 1.5 Update `internal/clients/elasticsearch_scoped_client_test.go` to assert on `*elasticsearch.TypedClient` returns

## 2. Delete obsolete helpers

- [ ] 2.1 Delete `internal/clients/elasticsearch/helpers.go`
- [ ] 2.2 Verify no remaining references to `doFWWrite` or `doSDKWrite` across the codebase

## 3. Delete redundant model files

- [ ] 3.1 Delete `internal/models/ml.go`
- [ ] 3.2 Delete `internal/models/transform.go`
- [ ] 3.3 Delete `internal/models/enrich.go`
- [ ] 3.4 Remove unused types from `internal/models/models.go` (`ClusterInfo`, `User`, `Role`, `RoleMapping`, `APIKey`, `IndexTemplate`, `ComponentTemplate`, `Policy`, `SnapshotRepository`, `SnapshotPolicy`, `DataStream`, `LogstashPipeline`, `Script`, `Watch`, and any other types with typedapi equivalents)
- [ ] 3.5 Retain custom types such as `BuildDate` and any Kibana/Observability-related models that are out of scope
- [ ] 3.6 Retain `internal/models/ingest.go` — ingest processor structs are custom and have no typedapi equivalent

## 4. Fix imports and references

- [ ] 4.1 Replace any remaining direct `github.com/elastic/go-elasticsearch/v8` (raw client) imports with `github.com/elastic/go-elasticsearch/v8/typedapi/types` where typed API types are used
- [ ] 4.2 Remove unused `github.com/elastic/go-elasticsearch/v8/esapi` imports from all files under `internal/clients/elasticsearch/` and any other Elasticsearch-scoped packages
- [ ] 4.3 Resolve any compilation errors uncovered by the build step

## 5. Verify build and quality

- [ ] 5.1 Run `make build` and confirm it exits with status 0
- [ ] 5.2 Run `make check-lint` and confirm no new lint failures
- [ ] 5.3 Run targeted acceptance tests for affected resources (cluster settings, security, index, ML, transform, enrich, etc.) to confirm no behavioral regressions
