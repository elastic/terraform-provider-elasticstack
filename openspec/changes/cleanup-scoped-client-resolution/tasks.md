## 1. Migrate the remaining production holdout

- [ ] 1.1 Update `internal/apm/agent_configuration` to configure from `ConvertProviderDataToFactory` and store a typed Kibana scoped client instead of `*clients.APIClient`.
- [ ] 1.2 Switch APM create/read/update/delete operations to obtain the Kibana OpenAPI client from the typed scoped client and remove the remaining `ConvertProviderData` production dependency.

## 2. Remove legacy broad-client bridges

- [ ] 2.1 Delete legacy provider-data and resource-scoped broad-client helpers from `internal/clients/api_client.go`, including `ConvertProviderData`, `MaybeNewAPIClientFromFrameworkResource`, `MaybeNewKibanaAPIClientFromFrameworkResource`, `NewAPIClientFromSDKResource`, `NewKibanaAPIClientFromSDKResource`, and `extractDefaultClientFromMeta`.
- [ ] 2.2 Remove the supported factory bridge back to the broad client from `internal/clients/provider_client_factory.go` while preserving private bootstrap logic needed to construct typed scoped clients.

## 3. Privatize the broad client and update compatibility surfaces

- [ ] 3.1 Make the broad client type private within `internal/clients` and remove duplicated exported Elasticsearch- and Kibana-specific helper behavior now owned by scoped clients.
- [ ] 3.2 Update `xpprovider`, acceptance helpers, and tests/mocks to stop exporting or depending on `clients.APIClient` and to construct typed scoped clients directly.

## 4. Sync specs and verify the cleanup

- [ ] 4.1 Update canonical OpenSpec specs under `openspec/specs/` to remove references to deleted broad-client helper paths and to reflect the typed scoped-client contract.
- [ ] 4.2 Run targeted tests, `make build`, and OpenSpec validation/checks needed to confirm the cleanup compiles and the synced requirements stay consistent.
