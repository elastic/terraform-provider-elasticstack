## Why

The typed Kibana/Fleet and Elasticsearch client-resolution changes are archived, but the codebase still keeps a legacy broad `APIClient` path alive for one Framework resource, several bridge helpers, and test/export surfaces. Finishing that cleanup now will bring the implementation back in line with the scoped-client design, reduce duplicate client behavior, and remove ambiguity about which client contract resources are allowed to use.

## What Changes

- Add an optional entity-local `kibana_connection` block to `elasticstack_apm_agent_configuration` using the shared Plugin Framework Kibana connection schema helper so the resource can either inherit provider defaults or target a scoped Kibana connection.
- Migrate the remaining Framework holdout, `elasticstack_apm_agent_configuration`, from provider data conversion through a broad `APIClient` to factory-resolved typed scoped clients resolved from that effective Kibana connection.
- Replace the split connection-schema fixture ownership model with a single provider-registry test that enumerates all entities registered by `provider.New(...)` and `provider.NewFrameworkProvider(...)`, runs one subtest per entity, validates the expected connection block contract for that entity using the registered type name plus explicit documented exceptions, and then asserts the full registered inventory was validated.
- Remove legacy broad-client bridge helpers from `internal/clients` once no production call sites remain, including `ConvertProviderData`, `MaybeNewAPIClientFromFrameworkResource`, `MaybeNewKibanaAPIClientFromFrameworkResource`, `NewAPIClientFromSDKResource`, and `NewKibanaAPIClientFromSDKResource`.
- Remove broad-client behavior from `APIClient` where the same behavior is already provided by `KibanaScopedClient` or `ElasticsearchScopedClient`.
- **BREAKING** Stop exporting `clients.APIClient` as a supported public surface; keep exported client-resolution APIs focused on `ProviderClientFactory`, `KibanaScopedClient`, and `ElasticsearchScopedClient`.
- Update acceptance helpers, tests, and synced OpenSpec specs so they reference scoped-client resolution rather than the removed broad-client helpers.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `provider-client-factory`: remove the remaining broad-client bridge from the factory contract and make the typed factory/scoped-client surface the only supported provider injection path.
- `provider-kibana-connection`: extend the typed `kibana_connection` contract to `elasticstack_apm_agent_configuration`, and ensure Framework resources that need Kibana-derived operations consume factory-resolved typed Kibana scoped clients rather than broad `APIClient` adapters.
- `provider-kibana-connection-coverage`: replace split fixture ownership with a single registry-driven provider coverage test that validates the expected connection block for every non-Elasticsearch registered entity and proves the full registered inventory was exercised.
- `provider-elasticsearch-connection`: cover registered Elasticsearch entities through the same single registry-driven provider coverage test so the expected Elasticsearch connection block is verified from the provider registry rather than a hand-maintained ownership inventory.
- `provider-elasticsearch-scoped-client-resolution`: finish the cleanup by removing overlapping Elasticsearch helper behavior from the broad client surface once scoped clients fully own it.
- `apm-agent-configuration`: update the resource schema and client-resolution contract so it accepts an optional `kibana_connection` override and acquires its Kibana OpenAPI client through typed scoped-client resolution instead of the broad provider API client.

## Impact

- Affected code is concentrated in `internal/clients`, `internal/apm/agent_configuration`, shared Kibana connection schema helpers, `xpprovider`, acceptance test helpers, and tests/mocks that still construct or depend on `APIClient`.
- Affected code also includes the provider connection-schema coverage tests and any shared test helpers used to enumerate registered provider entities.
- This change narrows the provider's supported client-resolution API surface and may require compatibility decisions for external `xpprovider` consumers.
- `elasticstack_apm_agent_configuration` gains a new optional `kibana_connection` block, so acceptance coverage and requirements must verify both provider-default and entity-local override paths.
- Connection-schema coverage can no longer rely on hand-maintained ownership inventories; the single provider coverage test must enumerate the complete registered entity inventory, apply the documented connection-block classification rules, and fail fast if future additions are not validated.
- Synced OpenSpec requirements under `openspec/specs/` will need cleanup so they stop naming helper paths that are being removed.
