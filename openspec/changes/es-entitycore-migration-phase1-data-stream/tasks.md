## 1. Scaffold PF resource

- [ ] 1.1 Create/replace `internal/elasticsearch/index/data_stream.go` with PF types
- [ ] 1.2 Define `Data` struct with PF types and envelope getters
- [ ] 1.3 Define PF schema factory (no `elasticsearch_connection` block)

## 2. Implement envelope callbacks

- [ ] 2.1 `readDataStream`: GET `/data_stream/{name}`, populate fields, return found flag
- [ ] 2.2 `deleteDataStream`: DELETE `/data_stream/{name}`
- [ ] 2.3 `createDataStream`: PUT `/data_stream/{name}`, set composite ID
- [ ] 2.4 `updateDataStream`: same as create (name is ForceNew)

## 3. Wire resource

- [ ] 3.1 Embed `*entitycore.ElasticsearchResource[Data]`
- [ ] 3.2 Add `ImportState` passthrough

## 4. Provider registration

- [ ] 4.1 Remove from `provider/provider.go`
- [ ] 4.2 Add to `provider/plugin_framework.go`

## 5. Convert tests

- [ ] 5.1 Acceptance tests should require no modification
- [ ] 5.2 Add from-SDK state migration test if needed

## 6. Verification

- [ ] 6.1 `make build`
- [ ] 6.2 `make check-lint`
- [ ] 6.3 Acceptance tests
