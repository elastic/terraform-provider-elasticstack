## Why

The provider is incrementally migrating Elasticsearch API calls from the raw `esapi` (untyped) client to the `go-elasticsearch` Typed API (`elasticsearch.TypedClient` via `ToTyped()`). The cluster helpers in `internal/clients/elasticsearch/cluster.go` are the next surface to migrate. Using the typed client eliminates hand-structured JSON request bodies, removes custom model types that duplicate upstream structs, and enables compile-time API shape checking.

## What Changes

- Migrate `internal/clients/elasticsearch/cluster.go` helper functions to use the typed client:
  - `GetClusterInfo` → `typedapi.Core.Info().Do(ctx)` — return `*info.Response` instead of `*models.ClusterInfo`
  - `PutSnapshotRepository` / `GetSnapshotRepository` / `DeleteSnapshotRepository` → snapshot repository typed APIs
  - `PutSlm` / `GetSlm` / `DeleteSlm` → SLM typed APIs
  - `PutSettings` / `GetSettings` → cluster settings typed APIs
  - `GetScript` / `PutScript` / `DeleteScript` → script typed APIs (uses `types.StoredScript`)
- Remove or deprecate custom models: `ClusterInfo`, `SnapshotRepository`, `SnapshotPolicy`, `Script`
- Update all callers in `internal/elasticsearch/cluster/` resources and tests that consume the migrated helpers:
  - `script` resource + tests
  - `script` acceptance tests
  - `settings` tests
  - `slm` tests
  - `snapshot_repository` tests
- Ensure zero behavioral changes from the Terraform user perspective; this is a pure implementation refactor.

## Capabilities

### New Capabilities
- `typed-client-cluster`: Internal contract for cluster helper functions migrated to the typed client. Defines typed API usage, return types, and not-found/error handling behavior for `GetClusterInfo`, snapshot repository, SLM, cluster settings, and script helpers.

### Modified Capabilities
<!-- No spec-level behavior changes for existing Terraform resources; this is an internal implementation refactor -->
(none)

## Impact

- **Code**: `internal/clients/elasticsearch/cluster.go`, `internal/models/models.go`, and resource/test files under `internal/elasticsearch/cluster/`
- **APIs**: No Terraform resource or data source behavior changes
- **Dependencies**: Relies on existing `go-elasticsearch/v8` Typed API support
- **Systems**: None — purely internal refactor
