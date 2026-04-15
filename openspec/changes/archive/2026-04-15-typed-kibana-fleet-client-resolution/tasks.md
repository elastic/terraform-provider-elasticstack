## 1. Provider Factory Foundation

- [x] 1.1 Add `ProviderClientFactory` under `internal/clients/` and update `provider/provider.go` and `provider/plugin_framework.go` to inject it instead of `*clients.APIClient`
- [x] 1.2 Implement typed Kibana/Fleet resolution methods plus temporary legacy Elasticsearch resolution methods on the factory
- [x] 1.3 Add `KibanaScopedClient` and unit tests covering provider-default and scoped `kibana_connection` resolution for both SDK and Plugin Framework entry points

## 2. Kibana And Fleet Sink Migration

- [x] 2.1 Update shared Kibana/Fleet helper and sink packages to accept `KibanaScopedClient` or narrow typed interfaces instead of `*clients.APIClient`
- [x] 2.2 Refactor helper utilities that currently unwrap Kibana, Kibana OpenAPI, SLO, or Fleet clients from `*clients.APIClient` to use the typed scoped client contract
- [x] 2.3 Add or update unit tests for typed Kibana/Fleet sink usage and Kibana-derived version/flavor behavior

## 3. Entity Adoption

- [x] 3.1 Migrate covered Plugin Framework Kibana and Fleet resources/data sources to store the injected factory and resolve `KibanaScopedClient` from `kibana_connection`
- [x] 3.2 Migrate covered SDK Kibana resources/data sources to store the injected factory and resolve `KibanaScopedClient` from `kibana_connection`
- [x] 3.3 Update acceptance and regression tests so provider-default and scoped `kibana_connection` behavior are both verified for covered Kibana/Fleet entities

## 4. Verification

- [x] 4.1 Run OpenSpec validation for the new change artifacts
- [x] 4.2 Run targeted Go tests for updated Kibana/Fleet client and entity packages plus `make build`
- [x] 4.3 Confirm Elasticsearch entities still work through the factory's temporary legacy path and leave `analysis/esclienthelperplugin` unchanged for the follow-up phase
