## 1. Typed Client Infrastructure

- [ ] 1.1 Add `GetESTypedClient() (*elasticsearch.TypedClient, error)` to `ElasticsearchScopedClient` in `internal/clients/elasticsearch_scoped_client.go`
- [ ] 1.2 Update `serverInfo()` and cached field to use `*core.InfoResponse` instead of `*models.ClusterInfo`
- [ ] 1.3 Update `ClusterID()`, `ServerVersion()`, and `ServerFlavor()` to read from `*core.InfoResponse`
- [ ] 1.4 Verify `make build` passes after scoped client changes

## 2. Cluster Helper Migration

- [ ] 2.1 Rewrite `GetClusterInfo` in `internal/clients/elasticsearch/cluster.go` to use `typedClient.Core.Info().Do(ctx)`
- [ ] 2.2 Rewrite `PutSnapshotRepository` to use typed API `Snapshot.CreateRepository().Do(ctx)`
- [ ] 2.3 Rewrite `GetSnapshotRepository` to use typed API `Snapshot.GetRepository().Do(ctx)` with union type handling
- [ ] 2.4 Rewrite `DeleteSnapshotRepository` to use typed API `Snapshot.DeleteRepository().Do(ctx)`
- [ ] 2.5 Rewrite `PutSlm` to use typed API `Slm.PutLifecycle().Do(ctx)`
- [ ] 2.6 Rewrite `GetSlm` to use typed API `Slm.GetLifecycle().Do(ctx)` and return `*types.SLMPolicy`
- [ ] 2.7 Rewrite `DeleteSlm` to use typed API `Slm.DeleteLifecycle().Do(ctx)`
- [ ] 2.8 Rewrite `PutSettings` to use typed API `Cluster.PutSettings().Do(ctx)`
- [ ] 2.9 Rewrite `GetSettings` to use typed API `Cluster.GetSettings().Do(ctx)` and flatten `json.RawMessage` values
- [ ] 2.10 Rewrite `GetScript` to use typed API `Core.GetScript().Do(ctx)` and return `*types.StoredScript`
- [ ] 2.11 Rewrite `PutScript` to use typed API `Core.PutScript().Do(ctx)`
- [ ] 2.12 Rewrite `DeleteScript` to use typed API `Core.DeleteScript().Do(ctx)`
- [ ] 2.13 Remove unused imports (`bytes`, `encoding/json`, `io`, `net/http`, `esapi`) from `cluster.go` after migration
- [ ] 2.14 Verify `make build` passes after helper migration

## 3. Caller Updates — Cluster Info

- [ ] 3.1 Update `internal/elasticsearch/cluster/cluster_info_data_source.go` to use `*core.InfoResponse`
- [ ] 3.2 Adapt `build_date` extraction to use typed API `DateTime.String()`
- [ ] 3.3 Update `internal/elasticsearch/cluster/cluster_info_data_source_test.go` if needed

## 4. Caller Updates — Snapshot Repository

- [ ] 4.1 Update `internal/elasticsearch/cluster/snapshot_repository.go` to adapt to typed API response type
- [ ] 4.2 Add repository union type-switch logic (`types.Repository` → concrete type → `Type` + `Settings`)
- [ ] 4.3 Update `internal/elasticsearch/cluster/snapshot_repository_data_source.go` for typed API response
- [ ] 4.4 Update `internal/elasticsearch/cluster/snapshot_repository_test.go` and `snapshot_repository_data_source_test.go` if needed

## 5. Caller Updates — SLM

- [ ] 5.1 Update `internal/elasticsearch/cluster/slm.go` to use `*types.SLMPolicy` and `*types.Retention`
- [ ] 5.2 Adapt `config` field mapping to typed API `types.Configuration`
- [ ] 5.3 Update `internal/elasticsearch/cluster/slm_test.go` if needed

## 6. Caller Updates — Cluster Settings

- [ ] 6.1 Update `internal/elasticsearch/cluster/settings.go` to handle `getsettings.Response` with `json.RawMessage` values
- [ ] 6.2 Add flattening helper to convert `map[string]json.RawMessage` to `map[string]any`
- [ ] 6.3 Update `internal/elasticsearch/cluster/settings_test.go` if needed

## 7. Caller Updates — Script Resource

- [ ] 7.1 Update `internal/elasticsearch/cluster/script/read.go` to use `*types.StoredScript`
- [ ] 7.2 Update `internal/elasticsearch/cluster/script/update.go` to build request via typed API and preserve `params` behavior
- [ ] 7.3 Update `internal/elasticsearch/cluster/script/delete.go` to call typed `DeleteScript`
- [ ] 7.4 Update `internal/elasticsearch/cluster/script/acc_test.go` if needed

## 8. Model Cleanup

- [ ] 8.1 Remove `ClusterInfo` struct from `internal/models/models.go`
- [ ] 8.2 Remove `SnapshotRepository` struct from `internal/models/models.go`
- [ ] 8.3 Remove `SnapshotPolicy`, `SnapshortRetention`, and `SnapshotPolicyConfig` from `internal/models/models.go`
- [ ] 8.4 Remove `Script` struct from `internal/models/models.go`
- [ ] 8.5 Remove any unused imports from `internal/models/models.go`
- [ ] 8.6 Verify `make build` passes after model removal

## 9. Regression Testing

- [ ] 9.1 Run `make build` and ensure clean compilation
- [ ] 9.2 Run `make test` for unit tests
- [ ] 9.3 Verify `make check-openspec` passes
- [ ] 9.4 Run acceptance tests for affected resources: `script`, `snapshot_repository`, `slm`, `settings`
- [ ] 9.5 Run acceptance test for `cluster_info` data source
