## 1. index_alias

- [x] 1.1 Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()` to `internal/elasticsearch/index/alias/models.go`
- [x] 1.2 Extract `Read` body to `readAlias` callback
- [x] 1.3 Extract `Delete` body to `deleteAlias` callback
- [x] 1.4 Extract `Create`/`Update` to `createAlias`/`updateAlias` callbacks
- [x] 1.5 Replace `ResourceBase` with `ElasticsearchResource[Data]` in `resource.go`
- [x] 1.6 Remove manual `elasticsearch_connection` from schema factory
- [x] 1.7 Keep `ImportState` on concrete type

## 2. data_stream_lifecycle

- [x] 2.1 Add envelope getters to `internal/elasticsearch/index/datastreamlifecycle/models.go`
- [x] 2.2 Extract Read/Delete/Create/Update to callbacks
- [x] 2.3 Swap ResourceBase for ElasticsearchResource
- [x] 2.4 Remove `elasticsearch_connection` from schema
- [x] 2.5 Preserve ImportState

## 3. enrich_policy

- [x] 3.1 Add envelope getters to `internal/elasticsearch/enrich/` model
- [x] 3.2 Extract Read/Delete/Create/Update to callbacks
- [x] 3.3 Swap ResourceBase for ElasticsearchResource
  - Note: `execute=true` behavior is inside the create/update callback
- [x] 3.4 Remove `elasticsearch_connection` from `GetResourceSchema()`
- [x] 3.5 Preserve custom ImportState (sets `execute=true` after passthrough)

## 4. inference_endpoint

- [x] 4.1 Add envelope getters to `internal/elasticsearch/inference/inferenceendpoint/models.go`
- [x] 4.2 Extract Read/Delete/Create/Update to callbacks
- [x] 4.3 Swap ResourceBase for ElasticsearchResource
- [x] 4.4 Remove `elasticsearch_connection` from schema
- [x] 4.5 Preserve ImportState

## 5. Verification

- [ ] 5.1 `make build`
- [ ] 5.2 `make check-lint`
- [ ] 5.3 Acceptance tests for all four resources
- [ ] 5.4 `openspec validate es-entitycore-migration-phase1-simple-pf --strict`
