## 1. Scaffold PF resource

- [x] 1.1 Create/replace `internal/elasticsearch/logstash/pipeline.go` with PF types
- [x] 1.2 Define `Data` struct with all pipeline fields and envelope getters
- [x] 1.3 Define PF schema factory (no `elasticsearch_connection`)

## 2. Settings helpers

- [x] 2.1 Create `expandSettings(data Data) map[string]any` for Create/Update
- [x] 2.2 Create `flattenSettings(apiMap map[string]any, data *Data)` for Read
- [x] 2.3 Preserve the known-keys validation list (static Go constant)

## 3. Implement envelope callbacks

- [x] 3.1 `readLogstashPipeline`: GET, populate fields including settings flatten
- [x] 3.2 `deleteLogstashPipeline`: DELETE
- [x] 3.3 `createLogstashPipeline`: PUT with pipeline body + settings, set ID
- [x] 3.4 `updateLogstashPipeline`: same as create (full replace)

## 4. Wire resource

- [x] 4.1 Embed `*entitycore.ElasticsearchResource[Data]`
- [x] 4.2 Add ImportState passthrough

## 5. Provider registration

- [x] 5.1 Remove from `provider/provider.go`
- [x] 5.2 Add to `provider/plugin_framework.go`

## 6. Verification

- [x] 6.1 `make build`
- [ ] 6.2 `make check-lint`
- [ ] 6.3 Acceptance tests
- [ ] 6.4 `openspec validate es-entitycore-migration-phase2-logstash --strict`
