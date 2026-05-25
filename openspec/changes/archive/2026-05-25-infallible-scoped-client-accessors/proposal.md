## Why

Every consumer of `*clients.ElasticsearchScopedClient` and `*clients.KibanaScopedClient` has to perform a two-step error-checking dance: first against `ProviderClientFactory.GetElasticsearchClient` / `GetKibanaClient`, then again against the scoped accessors `GetESClient`, `GetKibanaOapiClient`, and `GetFleetClient`. Both checks exist to surface the same condition — "no usable endpoint was configured for this client surface" — and the duplication shows up in ~100 call sites across the provider. Pushing the validation up into the factory lets the scoped accessors become infallible getters that callers can use unconditionally.

A second, smaller asymmetry compounds this: at provider-level config building, the Fleet config already falls back to the Kibana block (Fleet API is just the Kibana API), but the Kibana OpenAPI config has no symmetric fallback from the Fleet block. As a result, a provider configured only with a `fleet { ... }` block cannot serve Kibana resources, even though every endpoint in the `fleet` block is a valid Kibana endpoint. Making the inheritance bidirectional at the provider level removes a surprise.

## What Changes

- **BREAKING** (internal API only): `(*clients.ElasticsearchScopedClient).GetESClient()` returns `*elasticsearch.TypedClient` (no diagnostics).
- **BREAKING** (internal API only): `(*clients.KibanaScopedClient).GetKibanaOapiClient()` returns `*kibanaoapi.Client` (no diagnostics).
- **BREAKING** (internal API only): `(*clients.KibanaScopedClient).GetFleetClient()` returns `*fleet.Client` (no diagnostics).
- `ProviderClientFactory.GetElasticsearchClient` returns an error diagnostic when provider configuration plus environment overrides produce no Elasticsearch endpoint.
- `ProviderClientFactory.GetKibanaClient` returns an error diagnostic when provider configuration plus environment overrides produce no usable Kibana **or** Fleet endpoint.
- Provider-level `kibana_oapi` config gains a field-level fallback step from the `fleet { ... }` block for unset values (URL and other fields), mirroring the existing Fleet-from-Kibana inheritance. Resource-level `kibana_connection` block retains its current unified-override semantics — it supplies both `kibana_oapi` and `fleet` config via `toFleetConfig()` and is not affected by this change.
- Factory method names (`GetElasticsearchClient`, `GetKibanaClient`) are retained — no rename.
- All consumer call sites in `internal/clients/elasticsearch/**`, `internal/kibana/**`, `internal/fleet/**`, `internal/apm/**`, and `internal/acctest/**` drop the second diagnostic check after calling the scoped accessor.

## Capabilities

### New Capabilities

(none — all changes adjust existing capabilities)

### Modified Capabilities

- `provider-component-client-accessors`: scoped accessors become infallible; the endpoint-presence validation contract relocates to the factory.
- `provider-client-factory`: `GetElasticsearchClient` and `GetKibanaClient` gain explicit endpoint-presence preconditions for success and must surface a missing-endpoint error diagnostic when not satisfied. `GetKibanaClient` accepts a usable Kibana **or** Fleet endpoint to satisfy that precondition.
- `provider-kibana-connection`: provider-level `kibana_oapi` config inherits unset fields (URL and credentials) from the `fleet { ... }` provider block, making Kibana ↔ Fleet provider-level inheritance bidirectional. Resource-level `kibana_connection` semantics are unchanged.

## Impact

- **Code**: `internal/clients/provider_client_factory.go`, `internal/clients/elasticsearch_scoped_client.go`, `internal/clients/kibana_scoped_client.go`, `internal/clients/config/kibana_oapi.go`, `internal/clients/config/framework.go`, and ~100 consumer call sites across `internal/clients/elasticsearch/**`, `internal/kibana/**`, `internal/fleet/**`, `internal/apm/**`, `internal/acctest/**`, `internal/entitycore/**`.
- **Tests**: existing unit tests in `internal/clients/elasticsearch_scoped_client_test.go`, `internal/clients/kibana_scoped_client_test.go`, and `internal/clients/provider_client_factory_test.go` that assert error diagnostics from scoped accessors must be reworked to assert the same errors from the factory layer instead. New tests for the Fleet → Kibana inheritance direction at provider level.
- **External API**: `ProviderClientFactory` is exported and consumed by `xpprovider`; the factory method signatures do not change, only their success preconditions and diagnostics. Scoped client method signatures are internal-only and changing them does not affect external consumers.
- **Behavior**: a provider configured with only a `fleet` block (no `kibana`) starts succeeding for Kibana resources where it previously failed.
