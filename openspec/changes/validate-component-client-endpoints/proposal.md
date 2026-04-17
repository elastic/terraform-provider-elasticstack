## Why

Provider configuration is intentionally component-optional: practitioners should be able to configure only Elasticsearch, only Kibana, or only the components their Terraform resources actually use. Since the provider now resolves typed scoped clients through `*clients.ProviderClientFactory`, entities obtain concrete service clients from `*clients.ElasticsearchScopedClient` and `*clients.KibanaScopedClient` accessors rather than from a supported broad `*clients.APIClient` surface. Those scoped accessors still do not consistently enforce that the required endpoint for a specific component is actually present before returning a client, so missing component configuration leaks through as low-signal downstream failures such as `unsupported protocol scheme ""`, generic `client not found` errors, or misleading localhost behavior from the legacy Kibana client.

Issue [#355](https://github.com/elastic/terraform-provider-elasticstack/issues/355) asks for more relevant errors when the provider is not correctly configured. The narrow fix here is to validate endpoint presence only, at the point where an entity asks for a component client, and return an actionable message before any request is attempted.

## What Changes

- Add endpoint-present validation to typed scoped-client accessors before they return Elasticsearch, Kibana, Kibana OpenAPI, or Fleet clients.
- Require `(*clients.ElasticsearchScopedClient).GetESClient()` to fail with an actionable error when no effective Elasticsearch endpoint is configured from provider defaults, `elasticsearch_connection`, or environment overrides.
- Require `(*clients.KibanaScopedClient).GetKibanaClient()` and `GetKibanaOapiClient()` to fail with an actionable error when no effective Kibana endpoint is configured from provider defaults, `kibana_connection`, or environment overrides.
- Require `(*clients.KibanaScopedClient).GetFleetClient()` to fail with an actionable error when no effective Fleet endpoint is configured, including provider-level Fleet endpoint resolution and Fleet endpoint derivation from Kibana or `kibana_connection`.
- Keep scope limited to endpoint validation only; do not introduce authentication validation or new provider-schema requirements in this change.

## Capabilities

### New Capabilities

- `provider-component-client-accessors`: typed scoped-client accessors SHALL reject missing component endpoints with component-specific, actionable configuration errors before entity operations send requests.

### Modified Capabilities

- _(none)_

## Impact

- Specs: delta spec under `openspec/changes/validate-component-client-endpoints/specs/provider-component-client-accessors/spec.md`
- Typed scoped client accessors and resolved configuration handling in `internal/clients/elasticsearch_scoped_client.go`, `internal/clients/kibana_scoped_client.go`, `internal/clients/provider_client_factory.go`, and related constructor paths in `internal/clients/api_client.go`
- Accessor-level tests and any targeted entity regression coverage needed to verify the new diagnostics

