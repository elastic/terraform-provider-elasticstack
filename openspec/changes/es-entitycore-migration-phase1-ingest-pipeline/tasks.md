## 1. Scaffold PF resource package

- [ ] 1.1 Create `internal/elasticsearch/ingest/pipeline_pf.go` (or rename existing)
- [ ] 1.2 Define `Data` struct with PF types and envelope getter methods
- [ ] 1.3 Define `GetSchema` factory returning PF schema without `elasticsearch_connection` block

## 2. Implement envelope callbacks

- [ ] 2.1 Implement `readIngestPipeline(ctx, client, resourceID, state Data) (Data, bool, diag.Diagnostics)`
- [ ] 2.2 Implement `deleteIngestPipeline(ctx, client, resourceID, state Data) diag.Diagnostics`
- [ ] 2.3 Implement `createIngestPipeline(ctx, client, resourceID, state Data) (Data, diag.Diagnostics)`
  - Build pipeline body from model (JSON decode processors/on_failure)
  - Call `elasticsearch.PutIngestPipeline`
  - Set composite ID
- [ ] 2.4 Implement `updateIngestPipeline` (identical to create)

## 3. Wire resource into envelope

- [ ] 3.1 Replace old SDK resource type with `pipelineResource` embedding `*entitycore.ElasticsearchResource[Data]`
- [ ] 3.2 Constructor calls `entitycore.NewElasticsearchResource(...)`
- [ ] 3.3 Add `ImportState` passthrough on concrete type

## 4. Provider registration

- [ ] 4.1 Remove `ingest.ResourceIngestPipeline` from `provider/provider.go` `ResourcesMap`
- [ ] 4.2 Add constructor to `provider/plugin_framework.go` `resources()` slice

## 5. Convert tests

- [ ] 5.1 Update any unit tests that reference old SDK types
- [ ] 5.2 Acceptance tests should pass without modification
- [ ] 5.3 Add `TestAccResourceIngestPipelineFromSDK` if state shape migrated

## 6. Verification

- [ ] 6.1 `make build`
- [ ] 6.2 `make check-lint`
- [ ] 6.3 Acceptance tests against running stack
- [ ] 6.4 `openspec validate es-entitycore-migration-phase1-ingest-pipeline --strict`
