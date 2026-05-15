## 1. Envelope API redesign

- [ ] 1.1 Rename `DataSourceVersionRequirement` to `VersionRequirement` across `internal/entitycore`, its tests, Kibana model packages implementing `GetVersionRequirements()` (for example `internal/kibana/agentbuilderagent`, `internal/kibana/maintenance_window`, and `internal/kibana/streams`), and related spec references
- [ ] 1.2 Add `ElasticsearchResourceOptions[T]` and change the constructor to `NewElasticsearchResource[T](name string, opts ElasticsearchResourceOptions[T])`
- [ ] 1.3 Remove the `component` parameter from the Elasticsearch envelope and hard-code the Elasticsearch namespace for Metadata/type naming
- [ ] 1.4 Add `ElasticsearchCreateRequest[T]`, `ElasticsearchUpdateRequest[T]`, and `ElasticsearchWriteResult[T]`
- [ ] 1.5 Change Elasticsearch create/update callback types to use the new request/result structs
- [ ] 1.6 Add optional `WithReadResourceID` support and centralize read identity resolution for ordinary `Read` and read-after-write
- [ ] 1.7 Extend the Elasticsearch envelope to enforce `WithVersionRequirements` with the same Create/Read/Update semantics as Kibana
- [ ] 1.7.1 Generalize `enforceVersionRequirements` (or add `enforceElasticsearchVersionRequirements`) so Elasticsearch envelopes can enforce version requirements using `*clients.ElasticsearchScopedClient`
- [ ] 1.8 Add optional `PostRead` hook support to the Elasticsearch envelope, including passing `resp.Private` to the hook after a successful state set
- [ ] 1.9 Update `internal/entitycore/resource_envelope_test.go` and related tests to cover the new constructor, callback requests, read identity resolution, version requirements, and post-read hook behavior

## 2. Mechanical call-site migration

- [ ] 2.1 Update all Elasticsearch envelope constructor call sites to use `NewElasticsearchResource(name, opts)`
- [ ] 2.2 Remove now-unneeded `entitycore.ComponentElasticsearch` arguments from Elasticsearch envelope call sites
- [ ] 2.3 Run `make build` after the constructor migration to confirm all Elasticsearch resources compile against the new API

## 3. Proof-of-concept resource migrations

### 3.1 `elasticstack_elasticsearch_ml_filter`
- [ ] 3.1.1 Replace the concrete `Update` override with an envelope update callback using `ElasticsearchUpdateRequest[TFModel]`
- [ ] 3.1.2 Preserve existing diff semantics for `items` and `description`, including the current remote GET used to compute item add/remove diffs
- [ ] 3.1.3 Update focused tests for the resource if needed

### 3.2 `elasticstack_elasticsearch_cluster_settings`
- [ ] 3.2.1 Replace the concrete `Update` override with an envelope update callback using `ElasticsearchUpdateRequest[tfModel]`
- [ ] 3.2.2 Preserve removed-setting nulling behavior and read-after-write refresh semantics
- [ ] 3.2.3 Update focused tests for helper behavior and envelope integration if needed

### 3.3 `elasticstack_elasticsearch_ml_anomaly_detection_job`
- [ ] 3.3.1 Replace the concrete `Update` override with an envelope update callback using `ElasticsearchUpdateRequest[TFModel]`
- [ ] 3.3.2 Preserve partial-update body construction and no-updatable-change behavior
- [ ] 3.3.3 Update focused tests for the resource if needed

### 3.4 `elasticstack_elasticsearch_security_user`
- [ ] 3.4.1 Replace the concrete `Create` override with an envelope create callback using `ElasticsearchCreateRequest[Data]`
- [ ] 3.4.2 Replace the concrete `Update` override with an envelope update callback using `ElasticsearchUpdateRequest[Data]`
- [ ] 3.4.3 Preserve write-only password handling and prior-state password change detection using `Config` and `Prior`
- [ ] 3.4.4 Ensure the create/update callback sets `model.ID = types.StringValue(id.String())` before returning so envelope read-after-write persists a model carrying the composite ID
- [ ] 3.4.5 Update focused tests for the resource if needed

### 3.5 `elasticstack_elasticsearch_security_api_key`
- [ ] 3.5.1 Replace the concrete `Read` override with the envelope read path plus `PostRead` hook
- [ ] 3.5.2 Preserve cluster-version private-state persistence semantics after successful refresh
- [ ] 3.5.3 Update focused tests for the resource if needed

## 4. OpenSpec updates

- [ ] 4.1 Update `openspec/specs/entitycore-resource-envelope/spec.md` to describe the new Elasticsearch constructor, request/result callback API, `WithReadResourceID`, Elasticsearch version requirements, and `PostRead`
- [ ] 4.2 Update `openspec/specs/elasticsearch-ml-filter/spec.md` to reflect migration to the richer envelope update callback
- [ ] 4.3 Update `openspec/specs/elasticsearch-cluster-settings/spec.md` to reflect migration to the richer envelope update callback
- [ ] 4.4 Update `openspec/specs/elasticsearch-ml-anomaly-detection-job/spec.md` to reflect migration to the richer envelope update callback
- [ ] 4.5 Update `openspec/specs/elasticsearch-security-user/spec.md` to reflect migration to richer envelope create/update callbacks using config and prior state
- [ ] 4.6 Update `openspec/specs/elasticsearch-security-api-key/spec.md` to reflect envelope-owned read plus post-read side effects
- [ ] 4.7 Update `openspec/specs/entitycore-kibana-resource-envelope/spec.md` to reference `VersionRequirement` after the shared type rename

## 5. Validation

- [ ] 5.1 Run focused unit tests for `internal/entitycore/...`
- [ ] 5.2 Run focused tests for the migrated resources:
  - `go test ./internal/elasticsearch/ml/filter/...`
  - `go test ./internal/elasticsearch/cluster/settings/...`
  - `go test ./internal/elasticsearch/ml/anomalydetectionjob/...`
  - `go test ./internal/elasticsearch/security/user/...`
  - `go test ./internal/elasticsearch/security/api_key/...`
- [ ] 5.3 Run `make build`
- [ ] 5.4 Run `make check-lint`
- [ ] 5.5 Run `make check-openspec`
- [ ] 5.6 If Elasticsearch test infrastructure is available, run targeted acceptance tests for the five proof-of-concept resources
