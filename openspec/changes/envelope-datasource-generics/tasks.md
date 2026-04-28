## 1. Envelope generics constructor — `internal/entitycore`

- [ ] 1.1 Add `kibanaEnvelope[T any]` and `elasticsearchEnvelope[T any]` types with anonymous field embed
- [ ] 1.2 Add `NewKibanaDataSource[T]()` constructor returning `datasource.DataSource`
- [ ] 1.3 Add `NewElasticsearchDataSource[T]()` constructor returning `datasource.DataSource`
- [ ] 1.4 Implement schema injection for `kibana_connection` block into datasource schema
- [ ] 1.5 Implement schema injection for `elasticsearch_connection` block into datasource schema
- [ ] 1.6 Implement generic `Read()` on `genericKibanaDataSource[T]` (config decode → client resolve → callback → state set)
- [ ] 1.7 Implement generic `Read()` on `genericElasticsearchDataSource[T]`
- [ ] 1.8 Reuse existing datasource-compatible connection block helpers where possible (e.g., `GetKbFWConnectionBlock` / `GetEsFWConnectionBlock`), and if any changes are needed, ensure they target `github.com/hashicorp/terraform-plugin-framework/datasource/schema` block types (typically `dsschema.Block`), not `datasource.Block`
- [ ] 1.9 Write unit tests for envelope config decode, client resolution, and state set

## 2. Agent Builder data source migrations

- [ ] 2.1 Migrate `kibana/agentbuilderworkflow` data source to `NewKibanaDataSource[workflowDataSourceModel]`
- [ ] 2.2 Remove `data_source.go` and `data_source_read.go` orchestration from `agentbuilderworkflow`
- [ ] 2.3 Migrate `kibana/agentbuildertool` data source to `NewKibanaDataSource[toolDataSourceModel]`
- [ ] 2.4 Remove `data_source.go` and `data_source_read.go` orchestration from `agentbuildertool`
- [ ] 2.5 Evaluate `kibana/agentbuilderagent` data source for envelope fit; migrate if straightforward, otherwise extract domain-local read pipeline helper
- [ ] 2.6 Extract shared Agent Builder read pipeline helpers (version enforcement, composite ID resolution, space ID fallback) if multiple Agent Builder datasources remain struct-based

## 3. Verification

- [ ] 3.1 Run `make build` — no compilation errors
- [ ] 3.2 Run `make check-openspec` — OpenSpec validation passes
- [ ] 3.3 Run acceptance tests for migrated Agent Builder data sources
- [ ] 3.4 Verify existing struct-based data sources (`spaces`, `enrollment_tokens`) remain unaffected
- [ ] 3.5 Add `entitycore` unit test covering envelope `Read` with mocked client and config

## 4. Documentation

- [ ] 4.1 Add package-level GoDoc for `NewKibanaDataSource` / `NewElasticsearchDataSource` with usage example
- [ ] 4.2 Update `internal/entitycore/doc.go` to describe the envelope pattern alongside struct-based embedding
- [ ] 4.3 Add ADR or dev-docs note on when to use envelope vs struct-based pattern
