## Why

Kibana and Fleet entities currently rely on provider-level Kibana or Fleet configuration, which blocks the multi-cluster resource-level workflow requested in [#509](https://github.com/elastic/terraform-provider-elasticstack/issues/509). The repository also lacks a real helper-derived implementation for resource-scoped `kibana_connection`, so the feature needs a dedicated provider-level contract before it can be rolled out safely.

## What Changes

- Add provider-level support for resource-scoped `kibana_connection` that is parallel to `elasticsearch_connection` in shape and usage model.
- Define helper behavior for resolving a scoped API client from `kibana_connection` in both Plugin Framework and Plugin SDK code paths.
- Require scoped `kibana_connection` resolution to rebuild the Kibana-derived clients used by Kibana and Fleet entities, rather than reusing provider-level clients unchanged.
- Define how version, flavor, and identity checks behave when an entity uses a scoped `kibana_connection`.

## Capabilities

### New Capabilities
- `provider-kibana-connection`: provider-level requirements for the schema and client-resolution behavior of resource-scoped `kibana_connection`

### Modified Capabilities
<!-- None. -->

## Impact

- `internal/schema/connection.go`
- `internal/clients/api_client.go`
- `internal/clients/config/`
- Shared Kibana and Fleet client-construction paths in Plugin Framework and Plugin SDK code
- Existing Kibana action connector implementation, which should align with the new provider-level contract
