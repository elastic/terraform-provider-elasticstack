## Why

`kibana_connection` is still being finalized for Kibana and Fleet entities, and the current helper-based `*clients.APIClient` model does not let the compiler enforce that those entities actually rebuild scoped clients from the resource-local override. That leaves the highest-priority new behavior depending on conventions rather than types at the same time that users want reliable multi-cluster Kibana authentication support.

## What Changes

- Introduce a provider-injected client factory that becomes the supported provider data surface for resources and data sources.
- Add typed Kibana/Fleet scoped-client resolution so Kibana and Fleet code must resolve a scoped client from `kibana_connection` before reaching Kibana, Kibana OpenAPI, SLO, or Fleet sinks.
- Update `kibana_connection` requirements to describe factory-based, typed scoped-client behavior instead of helper-returned broad `*clients.APIClient` values.
- Keep Elasticsearch on its current broad-client and lint-enforced path for this phase so Kibana/Fleet enforcement can ship independently.

## Capabilities

### New Capabilities
- `provider-client-factory`: provider-level factory requirements for injecting typed client resolution into Framework `ProviderData` and SDK `meta`, while supporting phased migration from the current broad API client model

### Modified Capabilities
- `provider-kibana-connection`: replace helper-returned broad client resolution with factory-based typed Kibana/Fleet scoped client requirements for covered Kibana and Fleet entities

## Impact

- Provider configure paths in `provider/provider.go` and `provider/plugin_framework.go`
- Shared client construction and config code under `internal/clients/` and `internal/clients/config/`
- Kibana and Fleet helper and sink code under `internal/kibana/`, `internal/fleet/`, and supporting shared packages
- Existing `provider-kibana-connection` OpenSpec requirements
