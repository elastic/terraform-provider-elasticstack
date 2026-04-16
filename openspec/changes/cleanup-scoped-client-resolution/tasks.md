## 1. Migrate the remaining production holdout

- [x] 1.1 Add the shared Plugin Framework `kibana_connection` block to `internal/apm/agent_configuration` and configure the resource from `ConvertProviderDataToFactory` so it can resolve either provider-default or entity-local typed Kibana scoped clients.
- [x] 1.2 Switch APM create/read/update/delete operations to obtain the Kibana OpenAPI client from the typed scoped client selected by the effective `kibana_connection`, and remove the remaining `ConvertProviderData` production dependency.
- [x] 1.3 Replace the split provider connection-schema tests with a single registry-driven test that enumerates all entities registered by `provider.New(...)` and `provider.NewFrameworkProvider(...)`.
- [x] 1.4 In that single test, run one subtest per registered entity and validate that registered `elasticstack_elasticsearch_*` entities expose `elasticsearch_connection` while other registered entities expose `kibana_connection`, with any intentional exception asserted explicitly in the same test.
- [x] 1.5 Track validated entities in that single test and add a final completeness subtest that fails when any registered entity was not validated.
- [x] 1.6 Add or update acceptance coverage for `elasticstack_apm_agent_configuration` so both provider-default and entity-local `kibana_connection` paths are exercised.

## 2. Remove legacy broad-client bridges

- [x] 2.1 Delete legacy provider-data and resource-scoped broad-client helpers from `internal/clients/api_client.go`, including `ConvertProviderData`, `MaybeNewAPIClientFromFrameworkResource`, `MaybeNewKibanaAPIClientFromFrameworkResource`, `NewAPIClientFromSDKResource`, `NewKibanaAPIClientFromSDKResource`, and `extractDefaultClientFromMeta`.
- [x] 2.2 Remove the supported factory bridge back to the broad client from `internal/clients/provider_client_factory.go` while preserving private bootstrap logic needed to construct typed scoped clients.

## 3. Privatize the broad client and update compatibility surfaces

- [x] 3.1 Make the broad client type private within `internal/clients` and remove duplicated exported Elasticsearch- and Kibana-specific helper behavior now owned by scoped clients.
- [x] 3.2 Update `xpprovider`, acceptance helpers, and tests/mocks to stop exporting or depending on `clients.APIClient` and to construct typed scoped clients directly.

## 4. Sync specs and verify the cleanup

- [x] 4.1 Update canonical OpenSpec specs under `openspec/specs/` to remove references to deleted broad-client helper paths and to reflect the typed scoped-client contract.
- [x] 4.2 Run targeted tests, `make build`, and OpenSpec validation/checks needed to confirm the cleanup compiles and the synced requirements stay consistent.
