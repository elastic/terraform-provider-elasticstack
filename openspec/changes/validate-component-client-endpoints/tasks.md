## 1. Capture effective endpoint state for accessor validation

- [x] 1.1 Update the internal client construction and adapter paths so provider-default `*clients.ElasticsearchScopedClient` and `*clients.KibanaScopedClient` values retain the resolved endpoint values needed to validate Elasticsearch, Kibana, and Fleet accessors after provider configuration and environment overrides are applied.
- [x] 1.2 Update entity-local scoped-client builder paths in `internal/clients/provider_client_factory.go` so `elasticsearch_connection` and `kibana_connection` produce the same endpoint-validation metadata as provider-default clients.
- [x] 1.3 Keep Fleet endpoint validation aligned with the existing Fleet-from-Kibana endpoint resolution path so provider-level and `kibana_connection`-derived Fleet endpoints continue to work.

## 2. Enforce accessor-level endpoint checks

- [x] 2.1 Update `(*clients.ElasticsearchScopedClient).GetESClient()` to reject missing effective Elasticsearch endpoints with the new actionable configuration error.
- [x] 2.2 Update `(*clients.KibanaScopedClient).GetKibanaClient()` and `GetKibanaOapiClient()` to reject missing effective Kibana endpoints with component-specific actionable errors and without relying on legacy localhost defaults.
- [x] 2.3 Update `(*clients.KibanaScopedClient).GetFleetClient()` to reject missing effective Fleet endpoints with the Fleet-specific actionable error while preserving Kibana-derived Fleet endpoint resolution.
- [x] 2.4 Keep the new validation limited to endpoint presence only; do not add accessor failures for missing auth settings.

## 3. Add focused regression coverage

- [x] 3.1 Add unit coverage for typed scoped-client accessor behavior when Elasticsearch, Kibana, and Fleet endpoints are missing.
- [x] 3.2 Add focused coverage that proves Kibana endpoint validation blocks the legacy localhost fallback, entity-local connection blocks produce the same diagnostics, and Fleet access still succeeds when its endpoint is inherited from Kibana.

