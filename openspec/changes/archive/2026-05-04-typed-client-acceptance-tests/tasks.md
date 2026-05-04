## 1. Transform

- [x] 1.1 Migrate `internal/elasticsearch/transform/transform_test.go` from `GetESClient()` to `GetESTypedClient()` and typed transform APIs

## 2. Enrich

- [x] 2.1 Migrate `internal/elasticsearch/enrich/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed enrich APIs

## 3. Index Lifecycle Management (ILM)

- [x] 3.1 Migrate `internal/elasticsearch/index/ilm/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed ILM APIs

## 4. Index

- [x] 4.1 Migrate `internal/elasticsearch/index/index/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed index APIs
- [x] 4.2 Migrate `internal/elasticsearch/index/component_template_test.go` from `GetESClient()` to `GetESTypedClient()` and typed component-template APIs
- [x] 4.3 Migrate `internal/elasticsearch/index/data_stream_test.go` from `GetESClient()` to `GetESTypedClient()` and typed data-stream APIs
- [x] 4.4 Migrate `internal/elasticsearch/index/template/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed index-template APIs
- [x] 4.5 Migrate `internal/elasticsearch/index/templateilmattachment/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed APIs
- [x] 4.6 Migrate `internal/elasticsearch/index/datastreamlifecycle/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed data-stream-lifecycle APIs
- [x] 4.7 Migrate `internal/elasticsearch/index/alias/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed alias APIs

## 5. Inference

- [x] 5.1 Migrate `internal/elasticsearch/inference/inferenceendpoint/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed inference APIs

## 6. Logstash

- [x] 6.1 Migrate `internal/elasticsearch/logstash/pipeline_test.go` from `GetESClient()` to `GetESTypedClient()` and typed logstash-pipeline APIs

## 7. Security

- [x] 7.1 Migrate `internal/elasticsearch/security/role/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed security-role APIs
- [x] 7.2 Migrate `internal/elasticsearch/security/rolemapping/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed security-role-mapping APIs
- [x] 7.3 Migrate `internal/elasticsearch/security/user/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed security-user APIs

## 8. Cluster

- [x] 8.1 Migrate `internal/elasticsearch/cluster/script_test.go` from `GetESClient()` to `GetESTypedClient()` and typed script APIs
- [x] 8.2 Migrate `internal/elasticsearch/cluster/script/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed script APIs
- [x] 8.3 Migrate `internal/elasticsearch/cluster/settings_test.go` from `GetESClient()` to `GetESTypedClient()` and typed cluster-settings APIs
- [x] 8.4 Migrate `internal/elasticsearch/cluster/slm_test.go` from `GetESClient()` to `GetESTypedClient()` and typed SLM APIs
- [x] 8.5 Migrate `internal/elasticsearch/cluster/snapshot_repository_test.go` from `GetESClient()` to `GetESTypedClient()` and typed snapshot-repository APIs

## 9. Ingest

- [x] 9.1 Migrate `internal/elasticsearch/ingest/pipeline_test.go` from `GetESClient()` to `GetESTypedClient()` and typed ingest-pipeline APIs

## 10. Watcher

- [x] 10.1 Migrate `internal/elasticsearch/watcher/watch/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed watcher APIs

## 11. Kibana Streams

- [x] 11.1 Migrate `internal/kibana/streams/acc_test.go` from `GetESClient()` to `GetESTypedClient()` and typed APIs where equivalents exist

## 12. Client Tests

- [x] 12.1 Migrate `internal/clients/elasticsearch_scoped_client_test.go` from `GetESClient()` to `GetESTypedClient()` where applicable
- [x] 12.2 Migrate `internal/clients/provider_client_factory_test.go` from `GetESClient()` to `GetESTypedClient()` where applicable

## 13. Verification

- [x] 13.1 Run `make build` and confirm zero compile errors across the entire codebase
- [x] 13.2 Run `make check-lint` and resolve any new lint warnings introduced by typed API usage
- [x] 13.3 Run `go test ./internal/...` to verify all unit tests pass
- [x] 13.4 Confirm no remaining `GetESClient()` calls exist in any of the listed test files
- [x] 13.5 Run CI acceptance tests to confirm full test-suite passes with the typed client
