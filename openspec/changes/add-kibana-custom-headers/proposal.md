## Why

Kibana endpoints protected by access proxies such as Cloudflare Access or Google Identity-Aware Proxy require additional request headers that the provider cannot currently send. This prevents affected users from managing Kibana and Fleet APIs through the provider, including from Terraform Cloud.

## What Changes

- Add optional custom request headers to provider-level `kibana` configuration.
- Add the same custom request-header capability to entity-local `kibana_connection` configuration.
- Send configured custom headers on Kibana-derived API requests, including Fleet API requests that use the effective Kibana connection.
- Match the existing Elasticsearch header contract by accepting a sensitive map with one string value per header and no dedicated environment-variable override.
- Preserve the provider's configured Kibana authentication when a custom header uses the `Authorization` name.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `provider-kibana-connection`: Extend provider-level and entity-local Kibana connections with custom request headers and define their propagation and authentication precedence.

## Impact

- Provider and entity-local Kibana connection schemas and generated documentation.
- Kibana and Fleet client configuration and HTTP request behavior.
- Scoped-connection persistence used by provider lifecycle operations.
- Unit, schema, and client-factory tests for configuration propagation and request behavior.
