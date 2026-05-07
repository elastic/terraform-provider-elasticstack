## 1. Snapshot Lifecycle (SLM)

- [x] 1.1 Create `internal/elasticsearch/cluster/slm/resource.go` or restructure existing
- [x] 1.2 Define `Data` struct with PF types and envelope getters
- [x] 1.3 Define PF schema: `name`, `schedule`, `repository`, `config`, `retention`, `indices`, bool flags
- [x] 1.4 Implement `createSlm` callback: build `models.SlmPolicy`, PUT, set ID
- [x] 1.5 Implement `readSlm` callback: GET, populate model, return found
- [x] 1.6 Implement `deleteSlm` callback: DELETE
- [x] 1.7 Implement `updateSlm` callback (same endpoint as create)
- [x] 1.8 Wire into `ElasticsearchResource[Data]`
- [x] 1.9 Add ImportState passthrough

## 2. Snapshot Repository

- [x] 2.1 Define `Data` struct with PF types and envelope getters
- [x] 2.2 Define PF schema with per-type nested objects (`fs`, `url`, `azure`, `gcs`, `s3`)
- [x] 2.3 Add `ValidateConfig` on concrete type enforcing exactly-one type block
- [x] 2.4 Implement `createSnapshotRepository` callback: determine which type block is set, build API model, PUT
- [x] 2.5 Implement `readSnapshotRepository`: GET response, determine type, populate correct nested block
- [x] 2.6 Implement `deleteSnapshotRepository`: DELETE
- [x] 2.7 Implement `updateSnapshotRepository` callback
- [x] 2.8 Wire into envelope
- [x] 2.9 Add ImportState passthrough

## 3. Shared migration of old files

- [x] 3.1 Delete or archive `slm.go` and `snapshot_repository.go` SDK files
- [x] 3.2 Move acceptance tests to new package or update imports

## 4. Provider registration

- [x] 4.1 Remove both from `provider/provider.go`
- [x] 4.2 Add both to `provider/plugin_framework.go`

## 5. Verification

- [x] 5.1 `make build`
- [x] 5.2 `make check-lint`
- [x] 5.3 Acceptance tests: SLM + snapshot repository
- [x] 5.4 `openspec validate es-entitycore-migration-phase2-snapshots --strict`
