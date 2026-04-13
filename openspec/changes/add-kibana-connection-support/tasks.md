## 1. Shared helper support

- [ ] 1.1 Add dedicated SDK and Plugin Framework helper paths that resolve a scoped `*clients.APIClient` from `kibana_connection`.
- [ ] 1.2 Update client and config construction so scoped `kibana_connection` rebuilds Kibana legacy, Kibana OpenAPI, SLO, and Fleet clients together.
- [ ] 1.3 Ensure scoped version and identity behavior follows the scoped Kibana connection rather than provider-level Elasticsearch identity.

## 2. Align existing support

- [ ] 2.1 Update the shared entity-local `kibana_connection` schema helpers, if needed, so they remain the single source of truth for SDK and Framework entity blocks.
- [ ] 2.2 Align the existing Kibana action connector implementation with the new `kibana_connection` helper path.
- [ ] 2.3 Add focused unit coverage for scoped helper resolution and connector behavior.
