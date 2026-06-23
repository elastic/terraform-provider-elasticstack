## Why

When the provider is configured with different authentication mechanisms for Elasticsearch and Kibana/Fleet — for example, `api_key` for Elasticsearch and `username`/`password` for Kibana — two `Authorization` headers are sent on every Kibana/Fleet HTTP request. This violates [RFC 7230 §3.2.2](https://www.rfc-editor.org/rfc/rfc7230#section-3.2.2) and has caused HTTP 400 errors in deployments where a reverse proxy sits in front of Kibana. The issue has also been confirmed to affect the Fleet path.

The root cause is two cooperating bugs:

1. **Config layer**: `buildKibanaOapiConfigFromFramework` starts from `base.toKibanaOapiConfig()`, which copies all Elasticsearch credentials (including `APIKey`) as the starting point. When the Kibana block then sets `Username`/`Password`, those fields are written but the inherited `APIKey` is never cleared. The final `kibanaOapiConfig` silently carries both auth methods.

2. **Transport layer**: `transport.RoundTrip` in `kibanaoapi/client.go` applies each auth method independently using a mix of `Header.Set` and `Header.Add`, so multiple auth methods can appear on the wire simultaneously.

The same inheritance path is followed by `NewFromEnv` (used by acceptance tests), so the bug is present in both the provider configuration and acceptance-test paths.

## What Changes

- **Config layer (Kibana)**: `buildKibanaOapiConfigFromFramework` and `withNonURLEnvironmentOverrides` in `kibana_oapi.go` gain method-scoped auth clearing. When a higher-priority source (Kibana provider block or Kibana env vars) introduces an auth method, fields from conflicting auth methods inherited from lower-priority sources are cleared. Same-method fields from different sources are preserved to allow partial auth composition (e.g. `KIBANA_PASSWORD` in env + `username` from the provider schema).

- **Config layer (Fleet)**: `newFleetConfigFromFramework` and `withEnvironmentOverrides` in `fleet.go` receive the same treatment. Fleet block auth and Fleet env vars apply method-scoped clearing against fields inherited from the Kibana config.

- **Diagnostic warnings**: After the final config is assembled in `newProviderKibanaOapiConfigFromFramework`, `newKibanaOapiConfigFromFramework`, and `newFleetConfigFromFramework`, a `diag.AddWarning` is emitted when more than one auth method group is still set. This makes previously-silent precedence decisions visible to operators.

- **Transport safety net**: `transport.RoundTrip` in `kibanaoapi/client.go` is changed to a `switch` statement using `Header.Set` throughout, ensuring exactly one `Authorization` header is sent even if a `Config` struct reaches the transport with multiple auth methods set.

- **Tests**: New unit test scenarios in `kibana_oapi_test.go` and `fleet_test.go` cover mixed-auth configurations at each config-resolution layer.

## Capabilities

### Modified Capabilities

- `kibana-client`: The `kibanaOapiConfig` resolution functions gain method-scoped auth clearing and diagnostic warnings.
- `fleet-client`: The `fleetConfig` resolution functions gain the same treatment as the Kibana path.
- `provider-configuration`: The provider-level auth inheritance semantics are now correctly `ENV > RESOURCE > PROVIDER` and are enforced at the config layer, not just implicitly at the transport layer.

## Non-Goals

- Extending `TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT` semantics to auth fields (deferred follow-up).
- A full source-aware config rewrite that models all sources before resolving priority.
- Changes to Elasticsearch-facing request auth handling (`go-elasticsearch` manages its own auth).
- Migrating remaining resources from Plugin SDK to Plugin Framework.

## Impact

- **Files changed**: `internal/clients/config/kibana_oapi.go`, `internal/clients/config/fleet.go`, `internal/clients/kibanaoapi/client.go`, `internal/clients/config/kibana_oapi_test.go`, `internal/clients/config/fleet_test.go`.
- **Backward compatibility**: No breaking changes. The common case — where ES and Kibana share the same credentials and no explicit Kibana auth block is set — is unaffected (no clearing occurs when there is no higher-priority source).
- **Acceptance tests**: Existing tests that rely on the inherited-ES-auth path continue to work. New tests cover the mixed-auth case.
