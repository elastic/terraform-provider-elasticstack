## Why

Snapshot-related Terraform entities (`snapshot_repository` resource/data source, `snapshot_create` action, `snapshot_restore` action, and `slm` resource) are currently scattered across `internal/elasticsearch/cluster/` subpackages. This violates the code organization principle that packages should map to API domains, not to arbitrary nesting under a catch-all `cluster` directory. Finding snapshot-related code requires navigating multiple unrelated subpaths, and the `snapshot_repository` data source lives directly in `cluster/` rather than alongside its resource implementation.

## What Changes

- **Move `internal/elasticsearch/cluster/snapshot_repository/`** (resource + existing datasource files) → `internal/elasticsearch/snapshot/repository/`
- **Move `internal/elasticsearch/cluster/snapshot_create/`** → `internal/elasticsearch/snapshot/create/`
- **Move `internal/elasticsearch/cluster/snapshot_restore/`** → `internal/elasticsearch/snapshot/restore/`
- **Move `internal/elasticsearch/cluster/slm/`** → `internal/elasticsearch/snapshot/lifecycle/`
- **Move `ExpandWildcardsValidator`** from `cluster/snapshot_validators.go` → `utils/validators/expand_wildcards.go` (generic validator used by both snapshot and SLM code)
- **Remove orphaned datasource files** from `cluster/` (`snapshot_repository_data_source.go`, test files) after moving them into `snapshot/repository/`
- **Update provider registration** in `provider/plugin_framework.go` to use new package paths

No behavioral changes to any resource, data source, or action. This is a pure package-level refactor.

## Capabilities

### New Capabilities

_None — this is an internal reorganization with no new Terraform capabilities._

### Modified Capabilities

_None — no spec-level requirements change. Existing specs (`elasticsearch-snapshot-repository`, `elasticsearch-snapshot-lifecycle`, etc.) remain valid._

## Impact

- `provider/plugin_framework.go` — import paths and constructor calls updated for 4 registration sites (resource, datasource, 2 actions)
- `internal/elasticsearch/cluster/` — snapshot-related subpackages removed; validator moved out; datasource files removed
- `internal/elasticsearch/snapshot/` — new top-level domain directory created
- `internal/utils/validators/` — receives `ExpandWildcardsValidator`
- `internal/elasticsearch/cluster/slm/` — import path changes to reference new validator location
- Any open PRs touching `cluster/snapshot_*` or `cluster/slm/` will have merge conflicts
