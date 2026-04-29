## 1. Envelope generics constructor — `internal/entitycore`

- [x] 1.1 Add `KibanaConnectionField` / `ElasticsearchConnectionField` embeddable connection helpers and `KibanaDataSourceModel` / `ElasticsearchDataSourceModel` interface constraints for the generic constructors (anonymous-field embed of type parameters is not possible in Go as of golang/go#49030, so the concrete model embeds the helper directly)
- [x] 1.2 Add `NewKibanaDataSource[T]()` constructor returning `datasource.DataSource`
- [x] 1.3 Add `NewElasticsearchDataSource[T]()` constructor returning `datasource.DataSource`
- [x] 1.4 Implement schema injection for `kibana_connection` block into datasource schema
- [x] 1.5 Implement schema injection for `elasticsearch_connection` block into datasource schema
- [x] 1.6 Implement generic `Read()` on `genericKibanaDataSource[T]` (config decode → client resolve → callback → state set)
- [x] 1.7 Implement generic `Read()` on `genericElasticsearchDataSource[T]`
- [x] 1.8 Reuse existing datasource-compatible connection block helpers where possible (e.g., `GetKbFWConnectionBlock` / `GetEsFWConnectionBlock`), and if any changes are needed, ensure they target `github.com/hashicorp/terraform-plugin-framework/datasource/schema` block types (typically `dsschema.Block`), not `datasource.Block`
- [x] 1.9 Write unit tests for envelope config decode, client resolution, and state set

## 2. Agent Builder data source migrations

- [x] 2.1 Migrate `kibana/agentbuilderworkflow` data source to `NewKibanaDataSource[workflowDataSourceModel]`
- [x] 2.2 Remove `data_source.go` and `data_source_read.go` orchestration from `agentbuilderworkflow`
- [x] 2.3 Migrate `kibana/agentbuildertool` data source to `NewKibanaDataSource[toolDataSourceModel]`
- [x] 2.4 Remove `data_source.go` and `data_source_read.go` orchestration from `agentbuildertool`
- [x] 2.5 Evaluate `kibana/agentbuilderagent` data source for envelope fit; migrate if straightforward, otherwise extract domain-local read pipeline helper
- [x] 2.6 Extract shared Agent Builder read pipeline helpers (version enforcement, composite ID resolution, space ID fallback) if multiple Agent Builder datasources remain struct-based — **N/A**: after migrations only `agentbuilderagent` remains struct-based, so the condition is false.

## 3. Verification

- [x] 3.1 Run `make build` — no compilation errors
- [x] 3.2 Run `make check-openspec` — OpenSpec validation passes
- [x] 3.3 Run acceptance tests for migrated Agent Builder data sources
- [x] 3.4 Verify existing struct-based data sources (`spaces`, `enrollment_tokens`) remain unaffected
- [x] 3.5 Add `entitycore` unit test covering envelope `Read` with mocked client and config

## 4. Documentation

- [x] 4.1 Add package-level GoDoc for `NewKibanaDataSource` / `NewElasticsearchDataSource` with usage example
- [x] 4.2 Update `internal/entitycore/doc.go` to describe the envelope pattern alongside struct-based embedding
- [x] 4.3 Add ADR or dev-docs note on when to use envelope vs struct-based pattern
