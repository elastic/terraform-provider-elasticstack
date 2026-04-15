## 1. Capture effective endpoint state for accessor validation

- [ ] 1.1 Update `internal/clients/api_client.go` so `APIClient` retains the resolved endpoint values needed to validate Elasticsearch, Kibana, and Fleet accessors after provider configuration and environment overrides are applied.
- [ ] 1.2 Keep Fleet endpoint validation aligned with the existing Fleet-from-Kibana endpoint resolution path so inherited Kibana endpoints continue to work.

## 2. Enforce accessor-level endpoint checks

- [ ] 2.1 Update `GetESClient()` to reject missing effective Elasticsearch endpoints with the new actionable configuration error.
- [ ] 2.2 Update `GetKibanaClient()`, `GetKibanaOapiClient()`, and `GetSloClient()` to reject missing effective Kibana endpoints with component-specific actionable errors and without relying on legacy localhost defaults.
- [ ] 2.3 Update `GetFleetClient()` to reject missing effective Fleet endpoints with the Fleet-specific actionable error while preserving Kibana-derived Fleet endpoint resolution.
- [ ] 2.4 Keep the new validation limited to endpoint presence only; do not add accessor failures for missing auth settings.

## 3. Add focused regression coverage

- [ ] 3.1 Add unit coverage for accessor behavior when Elasticsearch, Kibana, and Fleet endpoints are missing.
- [ ] 3.2 Add focused coverage that proves Kibana endpoint validation blocks the legacy localhost fallback and that Fleet access still succeeds when its endpoint is inherited from Kibana.

