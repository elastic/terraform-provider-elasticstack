## Why

The provider-wide typed-client migration is in its final phase. All Elasticsearch helper and resource files have already been moved from the raw `esapi` client to `go-elasticsearch` Typed API (`elasticsearch.TypedClient`). The only remaining `GetESClient()` callers are inside acceptance-test files, where raw API methods are used directly for preflight setup, post-test cleanup, and existence checks. Eliminating these last untyped callers completes the migration, removes hand-structured request/response boilerplate from the test suite, and ensures the entire Elasticsearch surface is consistent.

## What Changes

- Migrate every remaining acceptance-test file that calls `GetESClient()` to `GetESTypedClient()` with typed API equivalents.
- Replace raw `esapi` request patterns (e.g. `esClient.Indices.Get`, `esClient.TransformGetTransform`, `esClient.EnrichGetPolicy`, `esClient.Security.GetRole`) with their `TypedClient` counterparts.
- Remove manual `io.ReadAll` + `json.Unmarshal` boilerplate where typed responses provide strongly-typed structs.
- Keep test behavior identical — no Terraform resource schemas, no provider configuration changes, and no test assertions are modified.

Files migrated:
- `internal/elasticsearch/transform/transform_test.go`
- `internal/elasticsearch/enrich/acc_test.go`
- `internal/elasticsearch/index/ilm/acc_test.go`
- `internal/elasticsearch/index/index/acc_test.go`
- `internal/elasticsearch/index/component_template_test.go`
- `internal/elasticsearch/index/data_stream_test.go`
- `internal/elasticsearch/index/template/acc_test.go`
- `internal/elasticsearch/index/templateilmattachment/acc_test.go`
- `internal/elasticsearch/index/datastreamlifecycle/acc_test.go`
- `internal/elasticsearch/index/alias/acc_test.go`
- `internal/elasticsearch/inference/inferenceendpoint/acc_test.go`
- `internal/elasticsearch/logstash/pipeline_test.go`
- `internal/elasticsearch/security/role/acc_test.go`
- `internal/elasticsearch/security/rolemapping/acc_test.go`
- `internal/elasticsearch/security/user/acc_test.go`
- `internal/elasticsearch/cluster/script_test.go`
- `internal/elasticsearch/cluster/script/acc_test.go`
- `internal/elasticsearch/cluster/settings_test.go`
- `internal/elasticsearch/cluster/slm_test.go`
- `internal/elasticsearch/cluster/snapshot_repository_test.go`
- `internal/elasticsearch/ingest/pipeline_test.go`
- `internal/elasticsearch/watcher/watch/acc_test.go`
- `internal/kibana/streams/acc_test.go`
- `internal/clients/elasticsearch_scoped_client_test.go`
- `internal/clients/provider_client_factory_test.go`

## Capabilities

### New Capabilities
_(none — this is an internal test-suite refactoring with no new user-visible capabilities)_

### Modified Capabilities
_(none — no spec-level requirements change; all tests preserve their existing behavior and assertions)_

## Impact

- **Test code only**: No provider runtime behavior changes.
- **APIs**: All acceptance tests continue to exercise the same Elasticsearch/Kibana APIs, only through the typed client surface.
- **Dependencies**: Relies on existing `go-elasticsearch/v8` (`ToTyped()`) already exposed via `ElasticsearchScopedClient.GetESTypedClient()`.
- **Build / CI**: `go test` compilation and acceptance-test execution paths are affected; no new dependencies introduced.
