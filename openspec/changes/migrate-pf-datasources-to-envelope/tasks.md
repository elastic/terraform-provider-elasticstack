## 1. Kibana-backed data sources (Batch 1)

- [ ] 1.1 Migrate `spaces` data source to envelope: embed `KibanaConnectionField` in model, extract schema factory, convert `Read` to callback, update `NewDataSource`, remove `entitycore_contract_test.go`
- [ ] 1.2 Migrate `export_saved_objects` data source to envelope: embed `KibanaConnectionField` in model, extract schema factory, convert `Read` to callback, update `NewDataSource`
- [ ] 1.3 Run acceptance tests for `spaces` and `export_saved_objects`; verify no regressions

## 2. Fleet-backed data sources (Batch 2)

- [ ] 2.1 Migrate `fleet_integration` data source to envelope: embed `KibanaConnectionField` in model, extract schema factory, convert `Read` to callback, update `NewDataSource`
- [ ] 2.2 Migrate `fleet_output` data source to envelope: embed `KibanaConnectionField` in model, extract schema factory, convert `Read` to callback, update `NewDataSource`
- [ ] 2.3 Migrate `enrollment_tokens` data source to envelope: embed `KibanaConnectionField` in model, extract schema factory, convert `Read` to callback, update `NewDataSource`, remove `entitycore_contract_test.go`
- [ ] 2.4 Run acceptance tests for `fleet_integration`, `fleet_output`, and `enrollment_tokens`; verify no regressions

## 3. Elasticsearch-backed data sources (Batch 3)

- [ ] 3.1 Migrate `security_role_mapping` data source to envelope: embed `ElasticsearchConnectionField` in model, extract schema factory, convert `Read` to callback, update `NewRoleMappingDataSource`
- [ ] 3.2 Migrate `enrich_policy` data source to envelope: embed `ElasticsearchConnectionField` in model, extract schema factory, convert `Read` to callback, update `NewEnrichPolicyDataSource`, remove `entitycore_contract_test.go`
- [ ] 3.3 Migrate `indices` data source to envelope: embed `ElasticsearchConnectionField` in model, extract schema factory, convert `Read` to callback, update `NewDataSource`
- [ ] 3.4 Migrate `index_template` data source to envelope: embed `ElasticsearchConnectionField` in model, extract schema factory, convert `Read` to callback, update `NewDataSource`, verify not-found empty-model behavior is preserved
- [ ] 3.5 Run acceptance tests for all Batch 3 data sources; verify no regressions

## 4. Agent Builder agent model cleanup

- [ ] 4.1 Update `agentbuilderagent/models.go` to embed `entitycore.KibanaConnectionField` instead of explicit `KibanaConnection` field
- [ ] 4.2 Remove `GetKibanaConnection()` and `GetVersionRequirements()` methods from `agentDataSourceModel`
- [ ] 4.3 Update any test or struct literal that references the old field name or methods
- [ ] 4.4 Run `agentbuilder_agent` acceptance tests; verify no regressions

## 5. Remove dead code

- [ ] 5.1 Delete `internal/entitycore/data_source_base.go`
- [ ] 5.2 Delete `internal/entitycore/data_source_base_test.go`
- [ ] 5.3 Verify `go build` passes with no references to `DataSourceBase`

## 6. Final verification

- [ ] 6.1 Run `make build` successfully
- [ ] 6.2 Run `make check-openspec` successfully
- [ ] 6.3 Run `make check-lint` successfully
- [ ] 6.4 Confirm all PF data source constructors in `provider/plugin_framework.go` still return valid `datasource.DataSource` values
