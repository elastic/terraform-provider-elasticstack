## Why

The typed Kibana/Fleet and Elasticsearch client-resolution changes are archived, but the codebase still keeps a legacy broad `APIClient` path alive for one Framework resource, several bridge helpers, and test/export surfaces. Finishing that cleanup now will bring the implementation back in line with the scoped-client design, reduce duplicate client behavior, and remove ambiguity about which client contract resources are allowed to use.

## What Changes

- Migrate the remaining Framework holdout, `elasticstack_apm_agent_configuration`, from provider data conversion through a broad `APIClient` to factory-resolved typed scoped clients.
- Remove legacy broad-client bridge helpers from `internal/clients` once no production call sites remain, including `ConvertProviderData`, `MaybeNewAPIClientFromFrameworkResource`, `MaybeNewKibanaAPIClientFromFrameworkResource`, `NewAPIClientFromSDKResource`, and `NewKibanaAPIClientFromSDKResource`.
- Remove broad-client behavior from `APIClient` where the same behavior is already provided by `KibanaScopedClient` or `ElasticsearchScopedClient`.
- **BREAKING** Stop exporting `clients.APIClient` as a supported public surface; keep exported client-resolution APIs focused on `ProviderClientFactory`, `KibanaScopedClient`, and `ElasticsearchScopedClient`.
- Update acceptance helpers, tests, and synced OpenSpec specs so they reference scoped-client resolution rather than the removed broad-client helpers.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `provider-client-factory`: remove the remaining broad-client bridge from the factory contract and make the typed factory/scoped-client surface the only supported provider injection path.
- `provider-kibana-connection`: ensure Framework resources that need Kibana-derived operations consume factory-resolved typed Kibana scoped clients rather than broad `APIClient` adapters.
- `provider-elasticsearch-scoped-client-resolution`: finish the cleanup by removing overlapping Elasticsearch helper behavior from the broad client surface once scoped clients fully own it.
- `apm-agent-configuration`: update the resource contract to acquire its Kibana OpenAPI client through typed scoped-client resolution instead of the broad provider API client.

## Impact

- Affected code is concentrated in `internal/clients`, `internal/apm/agent_configuration`, `xpprovider`, acceptance test helpers, and tests/mocks that still construct or depend on `APIClient`.
- This change narrows the provider's supported client-resolution API surface and may require compatibility decisions for external `xpprovider` consumers.
- Synced OpenSpec requirements under `openspec/specs/` will need cleanup so they stop naming helper paths that are being removed.
