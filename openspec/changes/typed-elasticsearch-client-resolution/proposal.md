## Why

`elasticsearch_connection` already has behavior protected by custom lint, but that protection is expensive because the compiler still sees provider-default clients, scoped clients, and bypassed clients as the same `*clients.APIClient` type. Once `kibana_connection` is stabilized with a factory-based typed model, Elasticsearch should move to the same compile-time enforcement pattern and remove the custom provenance analyzer entirely.

## What Changes

- Extend the provider client factory to produce typed Elasticsearch scoped clients for SDK and Plugin Framework entities.
- Change Elasticsearch helper and sink APIs to accept typed Elasticsearch scoped clients instead of the broad `*clients.APIClient`.
- Update Elasticsearch client-resolution requirements to describe factory-based typed resolution from `elasticsearch_connection`.
- Remove the custom Elasticsearch client-resolution lint capability and its repository wiring once typed sink boundaries make the misuse unrepresentable at compile time.

## Capabilities

### New Capabilities
- `provider-elasticsearch-scoped-client-resolution`: provider-level typed Elasticsearch scoped-client requirements for entity-local `elasticsearch_connection` resolution and Elasticsearch sink usage

### Modified Capabilities
- `provider-client-factory`: extend the phased provider factory contract so Elasticsearch entities also consume typed scoped clients rather than legacy broad-client access
- `elasticsearch-client-resolution-lint`: retire the custom lint capability after typed Elasticsearch sink boundaries replace provenance analysis

## Impact

- Provider configure paths in `provider/provider.go` and `provider/plugin_framework.go`
- Shared client construction and config code under `internal/clients/` and `internal/clients/config/`
- Elasticsearch helper and sink code under `internal/clients/elasticsearch/` and Elasticsearch entity code under `internal/elasticsearch/`
- `analysis/esclienthelperplugin`, `.golangci.yaml`, and lint-related Makefile wiring
- OpenSpec requirements for provider factory behavior and Elasticsearch enforcement
