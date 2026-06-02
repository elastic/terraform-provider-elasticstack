## 1. Extract shared validator

- [x] 1.1 Move `ExpandWildcardsValidator` struct + methods from `internal/elasticsearch/cluster/snapshot_validators.go` to new file `internal/utils/validators/expand_wildcards.go`
- [x] 1.2 Move `snapshot_validators_test.go` test cases to `internal/utils/validators/expand_wildcards_test.go`
- [x] 1.3 Update `cluster/slm/schema.go` to import `utils/validators` instead of `cluster`
- [x] 1.4 Update `cluster/snapshot_create/schema.go` to import `utils/validators` instead of `cluster`
- [x] 1.5 Verify `make build` passes after validator move

## 2. Create snapshot domain directory

- [x] 2.1 Create `internal/elasticsearch/snapshot/` directory structure
- [x] 2.2 Add `internal/elasticsearch/snapshot/doc.go` with package comment explaining the snapshot domain

## 3. Move snapshot repository (resource + datasource)

- [x] 3.1 Move `internal/elasticsearch/cluster/snapshot_repository/` → `internal/elasticsearch/snapshot/repository/`
- [x] 3.2 Move `internal/elasticsearch/cluster/snapshot_repository_data_source.go` → `internal/elasticsearch/snapshot/repository/data_source.go`
- [x] 3.3 Move `internal/elasticsearch/cluster/snapshot_repository_data_source_test.go` → `internal/elasticsearch/snapshot/repository/data_source_test.go`
- [x] 3.4 Move `internal/elasticsearch/cluster/snapshot_repository_data_source_internal_test.go` → `internal/elasticsearch/snapshot/repository/data_source_internal_test.go`
- [x] 3.5 Update package declarations in moved snapshot repository files: `snapshot_repository` → `repository`; and in moved datasource files: `cluster` → `repository`
- [x] 3.6 Remove `snapshot_repository.` prefix from all local references within moved datasource files (now in the same `repository` package)
- [x] 3.7 Remove `github.com/elastic/.../internal/elasticsearch/cluster/snapshot_repository` import from moved datasource files
- [x] 3.8 Move `internal/elasticsearch/cluster/descriptions/snapshot_repository_location_mode.md` → `internal/elasticsearch/snapshot/repository/descriptions/`
- [x] 3.9 Update `cluster/descriptions.go` to remove the embed/import for `snapshot_repository_location_mode.md`
- [x] 3.10 Verify `make build` passes after repository move

## 4. Move snapshot create action

- [x] 4.1 Move `internal/elasticsearch/cluster/snapshot_create/` → `internal/elasticsearch/snapshot/create/`
- [x] 4.2 Update package declaration: `snapshot_create` → `create`
- [x] 4.3 Remove `cluster.` prefix from `ExpandWildcardsValidator` reference (now in `utils/validators`)
- [x] 4.4 Verify `make build` passes after create move

## 5. Move snapshot restore action

- [x] 5.1 Move `internal/elasticsearch/cluster/snapshot_restore/` → `internal/elasticsearch/snapshot/restore/`
- [x] 5.2 Update package declaration: `snapshot_restore` → `restore`
- [x] 5.3 Verify `make build` passes after restore move

## 6. Move SLM resource to snapshot/lifecycle

- [x] 6.1 Move `internal/elasticsearch/cluster/slm/` → `internal/elasticsearch/snapshot/lifecycle/`
- [x] 6.2 Update package declaration: `slm` → `lifecycle`
- [x] 6.3 Remove `cluster.` prefix from `ExpandWildcardsValidator` reference (now in `utils/validators`)
- [x] 6.4 Verify `make build` passes after SLM move

## 7. Clean up cluster/ directory

- [x] 7.1 Delete `internal/elasticsearch/cluster/snapshot_validators.go`
- [x] 7.2 Delete `internal/elasticsearch/cluster/snapshot_validators_test.go`
- [x] 7.3 Delete `internal/elasticsearch/cluster/snapshot_repository_data_source.go` (already moved)
- [x] 7.4 Delete `internal/elasticsearch/cluster/snapshot_repository_data_source_test.go` (already moved)
- [x] 7.5 Delete `internal/elasticsearch/cluster/snapshot_repository_data_source_internal_test.go` (already moved)
- [x] 7.6 Verify no snapshot-related files remain under `internal/elasticsearch/cluster/`
- [x] 7.7 Verify `make build` passes after cleanup

## 8. Update provider registration

- [x] 8.1 Update `provider/plugin_framework.go`: change `cluster/snapshot_create` import to `snapshot/create`
- [x] 8.2 Update `provider/plugin_framework.go`: change `cluster/snapshot_restore` import to `snapshot/restore`
- [x] 8.3 Update `provider/plugin_framework.go`: change `cluster/snapshot_repository` import to `snapshot/repository`
- [x] 8.4 Update `provider/plugin_framework.go`: change `cluster/slm` import to `snapshot/lifecycle`
- [x] 8.5 Remove `cluster` datasource import from `plugin_framework.go` (snapshot datasource now lives in repository subpackage)
- [x] 8.6 Update constructor calls in `plugin_framework.go` to use new package aliases
- [x] 8.7 Verify `make build` passes after provider updates

## 9. Validate and verify

- [x] 9.1 Run `make build` — must pass
- [x] 9.2 Run `make test` (or targeted unit tests for moved packages) — must pass
- [x] 9.3 Run `make check-lint` — must pass
- [x] 9.4 Run `make check-openspec` — must pass
- [x] 9.5 Verify no orphaned import paths reference `internal/elasticsearch/cluster/snapshot_*`
- [x] 9.6 Verify `grep -r 'internal/elasticsearch/cluster' provider/plugin_framework.go` shows no snapshot-related imports
