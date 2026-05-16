## Why

The Elasticsearch entitycore resource envelope is now the common foundation for many Plugin Framework resources, but its current API is too narrow for a meaningful set of mature Elasticsearch resources. In particular, the current create/update callback contract only exposes the planned model and a write identity, which forces concrete resources to override the envelope when they need prior state, raw config, custom read identity semantics, version requirement enforcement, or post-read private-state side effects.

This shows up repeatedly in existing Elasticsearch resources:

- `elasticstack_elasticsearch_ml_filter` overrides `Update` because it needs prior state for description comparison and performs a remote GET to compute item add/remove diffs.
- `elasticstack_elasticsearch_cluster_settings` overrides `Update` to null removed settings.
- `elasticstack_elasticsearch_ml_anomaly_detection_job` overrides `Update` to build a partial update request from plan vs prior state.
- `elasticstack_elasticsearch_security_user` overrides `Create` and `Update` to access raw config for write-only password handling and to compare password fields with prior state.
- `elasticstack_elasticsearch_security_api_key` overrides `Read` to persist cluster-version private state after a successful refresh.

The current constructor is also overly positional and Elasticsearch-specific boilerplate like `entitycore.ComponentElasticsearch` is repeated at every call site even though the envelope already implies the Elasticsearch namespace.

This change evolves the Elasticsearch envelope into a richer CRUD-oriented abstraction that still preserves the same overall lifecycle shape: decode framework inputs, resolve the scoped client, enforce model-declared version requirements, invoke the concrete callback, perform read-after-write, set state, and run optional post-read side effects. It is intentionally scoped to CRUD-style resources and does not try to absorb state-transition resources such as ML job/datafeed state.

## What Changes

### Entitycore Elasticsearch envelope API

- Replace the positional Elasticsearch envelope constructor with an Elasticsearch-specific options form:
  - `NewElasticsearchResource[T](name string, opts ElasticsearchResourceOptions[T])`
- Hard-code the Elasticsearch component in the Elasticsearch envelope rather than requiring callers to pass `entitycore.ComponentElasticsearch`.
- Introduce structured create/update request objects passed to callbacks:
  - `ElasticsearchCreateRequest[T]` containing `Plan`, `Config`, and `WriteID`
  - `ElasticsearchUpdateRequest[T]` containing `Plan`, `Prior`, `Config`, and `WriteID`
- Introduce `ElasticsearchWriteResult[T]` to carry the callback's returned model reference for the envelope's read-after-write flow.
- Keep the existing read callback shape unchanged.
- Add an optional `PostRead` hook to the Elasticsearch envelope for read-side private-state or other post-refresh side effects.
- Extend the Elasticsearch envelope to honor the same optional `WithVersionRequirements` interface already used by the Kibana envelope and data source envelopes, renaming the shared requirement type from `DataSourceVersionRequirement` to `VersionRequirement`.
- Add an optional `WithReadResourceID` model interface so both ordinary `Read` and read-after-write can use a stable refresh identity that differs from the write identity when needed.

### Resource migrations (proof-of-concept set)

Migrate the following existing Elasticsearch resources to the richer envelope API as proof-of-concept resources:

1. `elasticstack_elasticsearch_ml_filter`
2. `elasticstack_elasticsearch_cluster_settings`
3. `elasticstack_elasticsearch_ml_anomaly_detection_job`
4. `elasticstack_elasticsearch_security_user`
5. `elasticstack_elasticsearch_security_api_key` (read/post-read path)

These five resources exercise the main new capabilities:

- prior-aware update callbacks (`ml_filter`, `cluster_settings`, `ml_anomaly_detection_job`)
- config-aware create/update callbacks (`security_user`)
- post-read hook and private-state persistence (`security_api_key`)

### Envelope scope and non-goals

- This proposal does **not** attempt to migrate or redesign state-transition resources such as `elasticstack_elasticsearch_ml_job_state` and `elasticstack_elasticsearch_ml_datafeed_state`.
- This proposal does **not** reshape the Elasticsearch read callback signature.
- This proposal does **not** introduce a second Elasticsearch CRUD envelope yet.

## Capabilities

### New Capabilities

- `entitycore-resource-envelope`: Elasticsearch resource envelopes support structured create/update callback requests, model-declared read identity, shared version requirement enforcement, and an optional post-read hook.

### Modified Capabilities

- `elasticsearch-ml-filter`: internal implementation migrated from an update override to the richer envelope update callback.
- `elasticsearch-cluster-settings`: internal implementation migrated from an update override to the richer envelope update callback.
- `elasticsearch-ml-anomaly-detection-job`: internal implementation migrated from an update override to the richer envelope update callback.
- `elasticsearch-security-user`: internal implementation migrated from concrete create/update overrides to richer envelope callbacks.
- `elasticsearch-security-api-key`: internal read path migrated from a concrete `Read` override to the envelope read path plus post-read hook.

## Impact

- `internal/entitycore/resource_envelope.go` and related tests will change substantially.
- Elasticsearch envelope call sites across `internal/elasticsearch/...` will be migrated from positional constructor arguments to `ElasticsearchResourceOptions[T]`.
- Shared version requirement types and naming in `internal/entitycore/version_requirements.go`, related tests/specs, and Kibana model packages implementing `GetVersionRequirements()` will change.
- Proof-of-concept resource implementations and tests will be updated in:
  - `internal/elasticsearch/ml/filter`
  - `internal/elasticsearch/cluster/settings`
  - `internal/elasticsearch/ml/anomalydetectionjob`
  - `internal/elasticsearch/security/user`
  - `internal/elasticsearch/security/api_key`
- Exported Elasticsearch envelope callback type names will change shape within `internal/entitycore`; this is safe because all current consumers are internal to the provider repository, but it increases merge-conflict risk for in-flight branches.
- OpenSpec specs for the envelope, the affected Elasticsearch resources, and the Kibana envelope/version-requirement references will need updates to reflect the richer envelope contract and the renamed shared requirement type.
