## 1. Entitycore Version Requirements

- [ ] 1.1 Add `DataSourceVersionRequirement` to `internal/entitycore/data_source_envelope.go` with `MinVersion *version.Version` and `ErrorMessage string`.
- [ ] 1.2 Add optional interface `KibanaDataSourceWithVersionRequirements` with `GetVersionRequirements() ([]DataSourceVersionRequirement, diag.Diagnostics)`.
- [ ] 1.3 Update `genericKibanaDataSource.Read` to detect the optional interface after scoped Kibana client resolution and before `readFunc`.
- [ ] 1.4 For each returned requirement, call `client.EnforceMinVersion`, append converted SDK diagnostics, and add `Unsupported server version` with the requirement's error message when unsupported.
- [ ] 1.5 Ensure data sources whose models do not implement the optional interface continue through the existing read path unchanged.

## 2. Entitycore Tests

- [ ] 2.1 Add tests proving `NewKibanaDataSource` still works when the model does not implement version requirements.
- [ ] 2.2 Add tests for models that implement version requirements, including successful enforcement and short-circuit behavior when diagnostics/errors occur where feasible with existing test helpers.
- [ ] 2.3 Verify existing envelope tests for schema injection, metadata, configure, and unconfigured-client diagnostics still pass.

## 3. Agent Builder Data Source Constructor and Schema

- [ ] 3.1 Replace the concrete `DataSource` implementation in `internal/kibana/agentbuilderagent/data_source.go` with `entitycore.NewKibanaDataSource[agentDataSourceModel](entitycore.ComponentKibana, "agentbuilder_agent", getDataSourceSchema, readAgentDataSource)`.
- [ ] 3.2 Remove local data source `client` storage, `Configure`, and `Metadata` methods.
- [ ] 3.3 Convert `Schema` in `internal/kibana/agentbuilderagent/data_source_schema.go` into `getDataSourceSchema() dsschema.Schema`.
- [ ] 3.4 Remove the explicit `kibana_connection` block from the Agent Builder schema factory so the envelope owns connection block injection.

## 4. Agent Builder Model and Read Callback

- [ ] 4.1 Add `GetKibanaConnection() types.List` to `agentDataSourceModel`, or embed `entitycore.KibanaConnectionField` if that produces a cleaner implementation.
- [ ] 4.2 Add `GetVersionRequirements()` to `agentDataSourceModel`, returning the static `minKibanaAgentBuilderAPIVersion` requirement with the current Agent Builder unsupported-version message.
- [ ] 4.3 Refactor `func (d *DataSource) Read(...)` into `readAgentDataSource(ctx context.Context, kbClient *clients.KibanaScopedClient, config agentDataSourceModel) (agentDataSourceModel, diag.Diagnostics)`.
- [ ] 4.4 Remove config decode, `GetKibanaClient`, static `minKibanaAgentBuilderAPIVersion` enforcement, and `resp.State.Set` from the callback because the envelope owns those steps.
- [ ] 4.5 Preserve composite ID parsing, default-space handling, agent fetch, model population, `include_dependencies` normalization, tool expansion, workflow lookup, and conditional `minVersionAdvancedAgentConfig` enforcement.

## 5. Agent Builder Tests

- [ ] 5.1 Add or update tests confirming `NewDataSource()` implements `datasource.DataSource` and `datasource.DataSourceWithConfigure` through the envelope.
- [ ] 5.2 Add or update tests confirming metadata remains `elasticstack_kibana_agentbuilder_agent`.
- [ ] 5.3 Add or update tests confirming the final data source schema still includes `kibana_connection` and the existing Agent Builder attributes.
- [ ] 5.4 Keep existing acceptance coverage for default space, explicit space, explicit `kibana_connection`, `include_dependencies`, and workflow-tool export unchanged.

## 6. Verification

- [ ] 6.1 Run `go test ./internal/entitycore ./internal/kibana/agentbuilderagent`.
- [ ] 6.2 Run `make check-openspec` once the OpenSpec CLI is available.
- [ ] 6.3 If an Elastic Stack test environment is available, run the targeted Agent Builder data source acceptance tests.
- [ ] 6.4 Run `make build` before considering implementation complete.
