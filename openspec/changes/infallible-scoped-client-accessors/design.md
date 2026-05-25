## Context

The provider exposes two scoped client types — `*clients.ElasticsearchScopedClient` and `*clients.KibanaScopedClient` — that resources obtain through `*clients.ProviderClientFactory`. The factory's two resolution methods (`GetElasticsearchClient`, `GetKibanaClient`) each return `(scoped, diag.Diagnostics)`. The scoped clients then expose accessors for their underlying typed transports (`GetESClient`, `GetKibanaOapiClient`, `GetFleetClient`) — each of which also returns `(client, diag.Diagnostics)`.

The two diagnostics layers exist for slightly different reasons:

- **Factory layer**: surfaces errors that arise while building the scoped client — bad CA certs, malformed endpoint URLs, configuration normalization failures.
- **Accessor layer**: re-validates "is the endpoint actually configured" because the factory historically always returns a scoped client even when its inner transports are not usable (e.g., a Kibana resource requested a scoped client when the provider was configured for Elasticsearch only).

This means every consumer writes the same boilerplate:

```go
scoped, diags := r.Client().GetKibanaClient(ctx, plan.KibanaConnection)
response.Diagnostics.Append(diags...)
if response.Diagnostics.HasError() {
    return
}

oapi, d := scoped.GetKibanaOapiClient()
response.Diagnostics.Append(d...)
if response.Diagnostics.HasError() {
    return
}
```

The second block is mechanical and identical across ~100 call sites.

Separately, the Kibana and Fleet provider blocks already overlap: the `fleet { ... }` block falls back to the `kibana { ... }` block via `kibanaOapiConfig.toFleetConfig()`, but the reverse direction is not wired. A provider whose only configured block is `fleet { ... }` cannot serve Kibana resources, even though the Fleet API is just the Kibana API and every endpoint in the `fleet` block is a valid Kibana endpoint.

## Goals / Non-Goals

**Goals:**

- Scoped accessors `GetESClient`, `GetKibanaOapiClient`, `GetFleetClient` become single-return getters. Consumers call them once, with no diagnostic check.
- Factory methods `GetElasticsearchClient` and `GetKibanaClient` guarantee on success that the returned scoped client's accessors will return usable clients.
- The same actionable error messages users see today continue to appear — they just appear at factory resolution time instead of accessor call time.
- A provider configured with only a `fleet { ... }` block can serve Kibana resources via field-level inheritance into the `kibana_oapi` config.
- The Fleet-from-Kibana inheritance that exists today is preserved.

**Non-Goals:**

- Renaming `GetElasticsearchClient` / `GetKibanaClient` factory methods. Their names and signatures stay; only their success preconditions change.
- Adding a `fleet_connection` block at the resource level. Resource-level `kibana_connection` continues to be the unified per-resource override that supplies both kibana_oapi and fleet config via `toFleetConfig()`.
- Changing the ES → Kibana → Fleet credential inheritance chain in `newBaseConfigFromFramework` / `newKibanaOapiConfigFromFramework`. That field-level inheritance is the established pattern; this change only adds a new inheritance edge (Fleet → Kibana at provider level).
- Changing acceptance-test client wiring (`NewAcceptanceTestingElasticsearchScopedClient`, `NewAcceptanceTestingKibanaScopedClient`). Those still construct scoped clients directly from env-derived `apiClient`.

## Decisions

### 1. Validation moves to the factory; accessors return single values

`ProviderClientFactory.GetElasticsearchClient` and `GetKibanaClient` perform endpoint-presence validation as a precondition. On success, the returned scoped client's accessors are guaranteed to return non-nil clients.

```go
// Before
func (e *ElasticsearchScopedClient) GetESClient() (*elasticsearch.TypedClient, fwdiag.Diagnostics)
func (k *KibanaScopedClient) GetKibanaOapiClient() (*kibanaoapi.Client, fwdiag.Diagnostics)
func (k *KibanaScopedClient) GetFleetClient() (*fleet.Client, fwdiag.Diagnostics)

// After
func (e *ElasticsearchScopedClient) GetESClient() *elasticsearch.TypedClient
func (k *KibanaScopedClient) GetKibanaOapiClient() *kibanaoapi.Client
func (k *KibanaScopedClient) GetFleetClient() *fleet.Client
```

**Alternative considered**: Keep accessors fallible but make the factory always validate. Rejected — leaves the consumer dance in place even when the factory could provide stronger guarantees.

**Alternative considered**: Make scoped accessors panic on misuse. Rejected — Terraform plugin code surfaces errors via diagnostics. Panicking would crash the provider process instead of producing a CLI error for the user.

### 2. `GetKibanaClient` accepts Kibana OR Fleet endpoint as a configured-state signal

The factory's `GetKibanaClient` returns an error diagnostic only when **both** kibana and fleet endpoint values are empty after all overlays (provider config, resource-level `kibana_connection`, environment overrides) are applied. If either is present, the factory builds and returns a scoped client and `GetKibanaOapiClient` / `GetFleetClient` are both safe to call.

**Why both**: the Fleet API is served by Kibana. A provider configured with only `fleet { endpoint = "..." }` is a fully valid configuration for Kibana resources. Today this case errors at accessor time because the kibana_oapi config inherits no URL.

**Alternative considered**: split factory methods into `GetKibanaClient` and `GetFleetClient` so each resource declares its preference and the factory validates the preferred surface only. Rejected — the user prefers the single factory method, and the per-accessor preference logic delivers the same observable behavior with less surface area.

### 3. Provider-level `kibana_oapi` config gains a Fleet-block fallback

In `internal/clients/config/`, the kibana_oapi config builder gains a field-level fallback from the `fleet { ... }` provider block for any field the kibana block leaves unset. Concretely the flow becomes:

```
base creds (from ES block + env)
   │
   ▼
+ kibana block overlays         ← only fields the user set in `kibana { ... }`
   │
   ▼
+ fleet block fallback overlays ← only fields still unset after the kibana step
   │  (URL, Username, Password, APIKey, BearerToken, CACerts, Insecure)
   │
   ▼
+ KIBANA_* env overrides        ← unchanged
```

The fleet block fallback applies **only** at the provider-level path (`newKibanaOapiConfigFromFramework`). It does **not** apply when a resource-level `kibana_connection` block is in play — that path (`NewFromFrameworkKibanaResource`) is a unified override that supplies both kibana_oapi and fleet config via `toFleetConfig()`.

**Field-level vs block-level fallback**: this matches the existing pattern (ES → Kibana → Fleet inheritance is field-level). A user can set `kibana { endpoints = [...] }` and `fleet { api_key = "..." }` and get a kibana_oapi config that uses the kibana endpoint with the fleet API key, just like today's Fleet → Kibana inheritance does.

**Alternative considered**: whole-block fallback (use kibana block if present, else use fleet block as a unit). Rejected — diverges from the established ES → Kibana → Fleet pattern and would silently disable parts of a partial `kibana { ... }` block when the user added a `fleet { ... }` block.

### 4. Validation logic lives in the factory, not in the config layer

Endpoint-presence checks today live in two places: in the accessors (post-construction) and indirectly via `cfg.KibanaOapi == nil` / `cfg.Elasticsearch == nil` checks in `newAPIClientFromConfig`. After this change, the canonical place to detect "no usable endpoint" is `ProviderClientFactory.GetElasticsearchClient` / `GetKibanaClient`. The accessors trust the factory.

For `GetKibanaClient`, the check is:

```go
hasKibana := scoped.kibanaEndpoint != ""
hasFleet  := scoped.fleetEndpoint != ""
if !hasKibana && !hasFleet {
    return nil, /* error: configure kibana.endpoints, fleet.endpoint, KIBANA_ENDPOINT, or FLEET_ENDPOINT */
}
```

For `GetElasticsearchClient`, the check is `len(esEndpoints) > 0` with at least one non-empty value — the same check the accessor performs today.

### 5. Error messages preserve user guidance

The factory-level error diagnostics use the same actionable wording the accessors emit today:

- `elasticsearch client is not configured: set elasticsearch.endpoints, elasticsearch_connection.endpoints, or ELASTICSEARCH_ENDPOINTS`
- For `GetKibanaClient`, a combined message: `kibana/fleet client is not configured: set kibana.endpoints, kibana_connection.endpoints, KIBANA_ENDPOINT, fleet.endpoint, or FLEET_ENDPOINT`

Tests that pin the existing strings update to assert the same wording from the factory.

### 6. Accessor signatures change is internal-only

`ElasticsearchScopedClient` and `KibanaScopedClient` are exposed by the `clients` package, but their accessor methods are only consumed within this module. `xpprovider`'s public surface only depends on `ProviderClientFactory` constructors and the existing factory method signatures (which are unchanged). The breaking nature of the accessor signature change is therefore contained within the provider tree.

## Risks / Trade-offs

**[Risk]** A resource path uses a scoped client accessor without first calling the factory in this same operation (e.g., a long-lived scoped client from a cache). → **Mitigation**: the factory builds a fresh scoped client per Create/Read/Update — there is no caching layer that hands out scoped clients across the boundary. The provider component tests already pin this lifecycle.

**[Risk]** A new failure case appears: a Kibana resource on a provider with no kibana/fleet block fails at factory resolution rather than at the kibana API call. → **Mitigation**: this is a strict improvement — earlier failure means clearer Terraform plan output. Documented in the proposal as the substantive contract change.

**[Risk]** The Fleet → Kibana inheritance lets a provider with a single `fleet { ... }` block start succeeding for Kibana resources, where users may have been relying on the old error to remind them to add a `kibana { ... }` block. → **Mitigation**: low risk because the Fleet API is in fact the Kibana API; if the fleet endpoint works, Kibana operations against it work too. Surfacing this as a release-note item is sufficient.

**[Risk]** Tests that lock in the current "accessor returns error" behavior (in `internal/clients/elasticsearch_scoped_client_test.go`, `internal/clients/kibana_scoped_client_test.go`) must be rewritten. → **Mitigation**: explicit task in `tasks.md`; the new expected errors come from the factory layer with the same message strings.

**[Trade-off]** The factory does more work upfront (endpoint validation) on every resource Create/Read/Update. → It is cheap (string comparisons against the resolved config) and runs once per operation regardless.

## Migration Plan

1. Add the Fleet → Kibana provider-level inheritance step in `internal/clients/config/kibana_oapi.go` (or a small helper in `framework.go`).
2. Move endpoint validation into `GetElasticsearchClient` / `GetKibanaClient` in `provider_client_factory.go`. Keep the existing diagnostics from `buildEsClient` / `buildKibanaScopedClientFromConfig`.
3. Change scoped accessor signatures in `elasticsearch_scoped_client.go` and `kibana_scoped_client.go` to single-return.
4. Update the three scoped-client unit-test files and `provider_client_factory_test.go` to assert errors from the factory layer.
5. Mechanically update consumer call sites — drop the second diagnostic check after the scoped accessor call. The set of files is enumerable: `internal/clients/elasticsearch/**`, `internal/kibana/**`, `internal/fleet/**`, `internal/apm/**`, `internal/acctest/**`, `internal/entitycore/**`.
6. Run `make build` and the full unit-test suite. Run targeted acceptance tests for one Kibana resource, one Fleet resource, and one Elasticsearch resource to confirm end-to-end behavior is unchanged.
