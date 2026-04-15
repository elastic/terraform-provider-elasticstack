## 1. Elasticsearch Factory And Scoped Client

- [ ] 1.1 Extend `ProviderClientFactory` with typed Elasticsearch resolution methods for SDK and Plugin Framework `elasticsearch_connection`
- [ ] 1.2 Add `ElasticsearchScopedClient` and unit tests for provider-default versus scoped Elasticsearch resolution behavior
- [ ] 1.3 Remove the factory's temporary legacy Elasticsearch resolution methods once all Elasticsearch consumers are migrated

## 2. Elasticsearch Sink And Entity Migration

- [ ] 2.1 Update `internal/clients/elasticsearch/` helper and sink packages to accept `ElasticsearchScopedClient` or narrow typed interfaces instead of `*clients.APIClient`
- [ ] 2.2 Migrate covered Plugin Framework Elasticsearch resources/data sources to store the injected factory and resolve `ElasticsearchScopedClient` from `elasticsearch_connection`
- [ ] 2.3 Migrate covered SDK Elasticsearch resources/data sources to store the injected factory and resolve `ElasticsearchScopedClient` from `elasticsearch_connection`
- [ ] 2.4 Update unit and acceptance tests for Elasticsearch helper behavior, version checks, cluster identity, and scoped override handling

## 3. Custom Lint Removal

- [ ] 3.1 Delete `analysis/esclienthelperplugin`, its tests, and any supporting wrapper/export packages
- [ ] 3.2 Remove `.golangci.yaml`, Makefile, and related lint workflow wiring for the Elasticsearch client-resolution custom analyzer
- [ ] 3.3 Verify there are no remaining in-scope Elasticsearch sinks or resources that depend on broad `*clients.APIClient` provenance analysis

## 4. Verification

- [ ] 4.1 Run OpenSpec validation for the new change artifacts
- [ ] 4.2 Run targeted Go tests for updated Elasticsearch client and entity packages plus `make build`
- [ ] 4.3 Run repository lint after analyzer removal to confirm the new compile-time enforcement model replaces the deleted custom lint task
