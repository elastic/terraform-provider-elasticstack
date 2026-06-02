## Context

The provider currently spreads snapshot-related Terraform entities across `internal/elasticsearch/cluster/`:

- `cluster/snapshot_repository/` — resource implementation (registered as `elasticstack_elasticsearch_snapshot_repository`)
- `cluster/snapshot_repository_data_source.go` — datasource implementation (in parent `cluster/` package)
- `cluster/snapshot_create/` — action (`elasticstack_elasticsearch_snapshot_create`)
- `cluster/snapshot_restore/` — action (`elasticstack_elasticsearch_snapshot_restore`)
- `cluster/slm/` — resource (`elasticstack_elasticsearch_slm`)
- `cluster/snapshot_validators.go` — shared `ExpandWildcardsValidator` used by `snapshot_create` and `slm`

This layout makes the snapshot domain hard to discover, splits a single datasource away from its resource, and buries SLM under `cluster` despite it being a snapshot concern.

## Goals / Non-Goals

**Goals:**
- Group all snapshot-related provider entities under a single domain directory: `internal/elasticsearch/snapshot/`
- Colocate each entity's resource/datasource/action implementations
- Reduce the number of import aliases in `provider/plugin_framework.go`
- Make snapshot code discoverable via directory traversal

**Non-Goals:**
- No behavioral changes to any resource, datasource, or action
- No changes to Terraform schema, API calls, or state shape
- No changes to test configurations or acceptance test behavior
- No new resources, datasources, or actions

## Decisions

### Decision 1: New top-level `snapshot/` domain under `elasticsearch/`

**Rationale:** Snapshots are a first-class Elasticsearch API domain (repository, create, restore, lifecycle). They deserve the same top-level treatment as `index/`, `security/`, `ml/`, etc.

**Alternatives considered:**
- Keep everything under `cluster/` but flatten subpackages. Rejected — `cluster/` is already a catch-all for settings, scripting, info, etc. Adding snapshot as a peer makes the problem worse.
- Move only `snapshot_repository` datasource into `snapshot_repository/`. Rejected — still leaves `snapshot_create`, `snapshot_restore`, and `slm` scattered under `cluster/`.

### Decision 2: Subpackage structure: `<domain>/<entity>` for >2 TF kinds, flat for ≤2

**Rationale:** This is the existing hybrid rule the repo already follows. Snapshot has 4 Terraform kinds (repository resource+datasource, create action, restore action, lifecycle resource), so split subpackages are warranted.

**Target tree:**
```
internal/elasticsearch/snapshot/
├── repository/     ← resource + datasource
├── create/         ← action
├── restore/        ← action
└── lifecycle/      ← resource (SLM)
```

### Decision 3: Move `ExpandWildcardsValidator` to `utils/validators/`

**Rationale:** The validator is not snapshot-specific — it validates a generic Elasticsearch API parameter (`expand_wildcards`). Placing it under `utils/validators/` makes it reusable by any package without creating a dependency from `cluster/slm` back into `snapshot/`.

**Alternatives considered:**
- Move to `snapshot/validators.go`. Rejected — `cluster/slm` would import `snapshot`, making SLM depend on a package that isn't semantically its parent.
- Keep in `cluster/`. Rejected — `snapshot/create` importing `cluster/` just for a validator feels wrong; it was only there by historical accident.

### Decision 4: Move `slm` to `snapshot/lifecycle/`

**Rationale:** SLM = Snapshot Lifecycle Management. It is semantically a snapshot concern, not a generic cluster concern. Moving it completes the snapshot domain.

**Alternatives considered:**
- Leave `slm` in `cluster/`. Rejected — inconsistent with the goal of unifying snapshot concerns.

### Decision 5: Rename constructors during move

`NewSnapshotRepositoryResource`, `NewSnapshotRepositoryDataSource`, `NewRestoreAction`, `NewCreateAction` — these stay named as-is. Only their package paths change.

The import aliases in `plugin_framework.go` simplify from:
```go
snapshot_create    ".../cluster/snapshot_create"
snapshot_repository ".../cluster/snapshot_repository"
snapshot_restore   ".../cluster/snapshot_restore"
cluster            ".../cluster" // for NewSnapshotRepositoryDataSource
```
to:
```go
snapshotcreate    ".../snapshot/create"
snapshotrepo      ".../snapshot/repository"
snapshotrestore   ".../snapshot/restore"
snapshotlifecycle ".../snapshot/lifecycle"
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Open PRs touching `cluster/snapshot_*` or `cluster/slm/` will conflict | Land this during a low-activity window; communicate in team channel |
| Import paths in generated docs or external references break | Only internal imports change; no Terraform resource names change |
| `cluster/` becomes sparse and loses its identity | This is desirable — `cluster/` should shrink to `settings`, `script`, `info` |
| Validation build caching invalidated | One-time cost; `make build` confirms correctness |

## Migration Plan

1. Move `ExpandWildcardsValidator` + test → `utils/validators/`
2. Update `cluster/slm/schema.go` and `cluster/snapshot_create/schema.go` to import new validator location
3. Create `internal/elasticsearch/snapshot/` directory
4. Move `cluster/snapshot_create/` → `snapshot/create/`
5. Move `cluster/snapshot_restore/` → `snapshot/restore/`
6. Move `cluster/snapshot_repository/` + datasource files → `snapshot/repository/`
7. Move `cluster/slm/` → `snapshot/lifecycle/`
8. Update `provider/plugin_framework.go` registrations
9. Run `make build` and `make test` to verify
10. Delete empty/obsolete files from `cluster/`

Rollback: Revert the PR. No state or schema changes mean rollback is safe.

## Open Questions

*None.*
