## 1. index_alias

- [ ] 1.1 Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()` to `internal/elasticsearch/index/alias/models.go`
- [ ] 1.2 Extract `Read` body to `readAlias` callback
- [ ] 1.3 Extract `Delete` body to `deleteAlias` callback
- [ ] 1.4 Extract `Create`/`Update` to `createAlias`/`updateAlias` callbacks
- [ ] 1.5 Replace `ResourceBase` with `ElasticsearchResource[Data]` in `resource.go`
- [ ] 1.6 Remove manual `elasticsearch_connection` from schema factory
- [ ] 1.7 Keep `ImportState` on concrete type

## 2. data_stream_lifecycle

- [ ] 2.1 Add envelope getters to `internal/elasticsearch/index/datastreamlifecycle/models.go`
- [ ] 2.2 Extract Read/Delete/Create/Update to callbacks
- [ ] 2.3 Swap ResourceBase for ElasticsearchResource
- [ ] 2.4 Remove `elasticsearch_connection` from schema
- [ ] 2.5 Preserve ImportState

## 3. enrich_policy

- [ ] 3.1 Add envelope getters to `internal/elasticsearch/enrich/` model
- [ ] 3.2 Extract Read/Delete/Create/Update to callbacks
- [ ] 3.3 Swap ResourceBase for ElasticsearchResource
  - Note: `execute=true` behavior is inside the create/update callback
- [ ] 3.4 Remove `elasticsearch_connection` from `GetResourceSchema()`
- [ ] 3.5 Preserve custom ImportState (sets `execute=true` after passthrough)

## 4. inference_endpoint

- [ ] 4.1 Add envelope getters to `internal/elasticsearch/inference/inferenceendpoint/models.go`
- [ ] 4.2 Extract Read/Delete/Create/Update to callbacks
- [ ] 4.3 Swap ResourceBase for ElasticsearchResource
- [ ] 4.4 Remove `elasticsearch_connection` from schema
- [ ] 4.5 Preserve ImportState

## 5. Verification

- [ ] 5.1 `make build`
- [ ] 5.2 `make check-lint`
- [ ] 5.3 Acceptance tests for all four resources
- [ ] 5.4 `openspec validate es-entitycore-migration-phase1-simple-pf --strict`
