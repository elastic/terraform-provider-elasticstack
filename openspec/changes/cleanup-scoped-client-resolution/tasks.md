## 1. Migrate the remaining production holdout

- [ ] 1.1 Add the shared Plugin Framework `kibana_connection` block to `internal/apm/agent_configuration` and configure the resource from `ConvertProviderDataToFactory` so it can resolve either provider-default or entity-local typed Kibana scoped clients.
- [ ] 1.2 Switch APM create/read/update/delete operations to obtain the Kibana OpenAPI client from the typed scoped client selected by the effective `kibana_connection`, and remove the remaining `ConvertProviderData` production dependency.
- [ ] 1.3 Update `provider/kibana_connection_schema_test.go` so it explicitly owns the registered `elasticstack_kibana_*`, `elasticstack_fleet_*`, and `elasticstack_apm_agent_configuration` entities rather than relying on partial prefix heuristics.
- [ ] 1.4 Update `provider/elasticsearch_connection_schema_test.go` so it explicitly owns the registered `elasticstack_elasticsearch_*` entities in a form that can participate in a shared registry-completeness check.
- [ ] 1.5 Add a shared provider-registry completeness assertion for the two connection-schema fixtures that enumerates entities from `provider.New(...)` and `provider.NewFrameworkProvider(...)` and fails on uncovered or doubly covered entities.
- [ ] 1.6 Add or update acceptance coverage for `elasticstack_apm_agent_configuration` so both provider-default and entity-local `kibana_connection` paths are exercised.

## 2. Remove legacy broad-client bridges

- [ ] 2.1 Delete legacy provider-data and resource-scoped broad-client helpers from `internal/clients/api_client.go`, including `ConvertProviderData`, `MaybeNewAPIClientFromFrameworkResource`, `MaybeNewKibanaAPIClientFromFrameworkResource`, `NewAPIClientFromSDKResource`, `NewKibanaAPIClientFromSDKResource`, and `extractDefaultClientFromMeta`.
- [ ] 2.2 Remove the supported factory bridge back to the broad client from `internal/clients/provider_client_factory.go` while preserving private bootstrap logic needed to construct typed scoped clients.

## 3. Privatize the broad client and update compatibility surfaces

- [ ] 3.1 Make the broad client type private within `internal/clients` and remove duplicated exported Elasticsearch- and Kibana-specific helper behavior now owned by scoped clients.
- [ ] 3.2 Update `xpprovider`, acceptance helpers, and tests/mocks to stop exporting or depending on `clients.APIClient` and to construct typed scoped clients directly.

## 4. Sync specs and verify the cleanup

- [ ] 4.1 Update canonical OpenSpec specs under `openspec/specs/` to remove references to deleted broad-client helper paths and to reflect the typed scoped-client contract.
- [ ] 4.2 Run targeted tests, `make build`, and OpenSpec validation/checks needed to confirm the cleanup compiles and the synced requirements stay consistent.
