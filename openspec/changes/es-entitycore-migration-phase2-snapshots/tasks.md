## 1. Snapshot Lifecycle (SLM)

- [ ] 1.1 Create `internal/elasticsearch/cluster/slm/resource.go` or restructure existing
- [ ] 1.2 Define `Data` struct with PF types and envelope getters
- [ ] 1.3 Define PF schema: `name`, `schedule`, `repository`, `config`, `retention`, `indices`, bool flags
- [ ] 1.4 Implement `createSlm` callback: build `models.SlmPolicy`, PUT, set ID
- [ ] 1.5 Implement `readSlm` callback: GET, populate model, return found
- [ ] 1.6 Implement `deleteSlm` callback: DELETE
- [ ] 1.7 Implement `updateSlm` callback (same endpoint as create)
- [ ] 1.8 Wire into `ElasticsearchResource[Data]`
- [ ] 1.9 Add ImportState passthrough

## 2. Snapshot Repository

- [ ] 2.1 Define `Data` struct with PF types and envelope getters
- [ ] 2.2 Define PF schema with per-type nested objects (`fs`, `url`, `azure`, `gcs`, `s3`)
- [ ] 2.3 Add `ValidateConfig` on concrete type enforcing exactly-one type block
- [ ] 2.4 Implement `createSnapshotRepository` callback: determine which type block is set, build API model, PUT
- [ ] 2.5 Implement `readSnapshotRepository`: GET response, determine type, populate correct nested block
- [ ] 2.6 Implement `deleteSnapshotRepository`: DELETE
- [ ] 2.7 Implement `updateSnapshotRepository` callback
- [ ] 2.8 Wire into envelope
- [ ] 2.9 Add ImportState passthrough

## 3. Shared migration of old files

- [ ] 3.1 Delete or archive `slm.go` and `snapshot_repository.go` SDK files
- [ ] 3.2 Move acceptance tests to new package or update imports

## 4. Provider registration

- [ ] 4.1 Remove both from `provider/provider.go`
- [ ] 4.2 Add both to `provider/plugin_framework.go`

## 5. Verification

- [ ] 5.1 `make build`
- [ ] 5.2 `make check-lint`
- [ ] 5.3 Acceptance tests: SLM + snapshot repository
- [ ] 5.4 `openspec validate es-entitycore-migration-phase2-snapshots --strict`
