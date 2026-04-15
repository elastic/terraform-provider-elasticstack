## Why

Provider configuration is intentionally component-optional: practitioners should be able to configure only Elasticsearch, only Kibana, or only the components their Terraform resources actually use. The current provider client accessors do not enforce that the required endpoint for a specific component is actually present before returning a client, so missing component configuration leaks through as low-signal downstream failures such as `unsupported protocol scheme ""` or misleading localhost behavior from the legacy Kibana client.

Issue [#355](https://github.com/elastic/terraform-provider-elasticstack/issues/355) asks for more relevant errors when the provider is not correctly configured. The narrow fix here is to validate endpoint presence only, at the point where an entity asks for a component client, and return an actionable message before any request is attempted.

## What Changes

- Add endpoint-present validation to `*clients.APIClient` component accessors before they return Elasticsearch, Kibana, Kibana OpenAPI, SLO, or Fleet clients.
- Require `GetESClient()` to fail with an actionable error when no effective Elasticsearch endpoint is configured.
- Require `GetKibanaClient()`, `GetKibanaOapiClient()`, and `GetSloClient()` to fail with an actionable error when no effective Kibana endpoint is configured.
- Require `GetFleetClient()` to fail with an actionable error when no effective Fleet endpoint is configured, including the case where Fleet relies on Kibana-derived endpoint resolution.
- Keep scope limited to endpoint validation only; do not introduce authentication validation or new provider-schema requirements in this change.

## Capabilities

### New Capabilities

- `provider-component-client-accessors`: provider client accessors SHALL reject missing component endpoints with component-specific, actionable configuration errors before entity operations send requests.

### Modified Capabilities

- _(none)_

## Impact

- Specs: delta spec under `openspec/changes/validate-component-client-endpoints/specs/provider-component-client-accessors/spec.md`
- Provider client accessors and resolved configuration handling in `internal/clients/api_client.go`
- Accessor-level tests and any targeted entity regression coverage needed to verify the new diagnostics

