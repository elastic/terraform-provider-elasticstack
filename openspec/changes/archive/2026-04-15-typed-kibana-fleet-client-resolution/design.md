## Context

Kibana and Fleet entities currently receive the provider's default `*clients.APIClient` through Framework `ProviderData` or SDK `meta`, then optionally rebuild another broad `*clients.APIClient` through `kibana_connection` helpers. That model leaves the provider default client and the correctly scoped client with the same static type, so the compiler cannot force Kibana and Fleet code to resolve a scoped client before reaching Kibana, Kibana OpenAPI, SLO, or Fleet operations.

This matters now because resource-level `kibana_connection` is still being finalized and is directly tied to the multi-cluster Kibana authentication use case tracked in issue [#509](https://github.com/elastic/terraform-provider-elasticstack/issues/509). The next production-ready phase should make Kibana/Fleet behavior correct by construction without waiting for a larger Elasticsearch migration.

## Goals / Non-Goals

**Goals:**
- Introduce a provider-injected factory that becomes the stable provider data contract for both Framework and SDK entities.
- Make Kibana and Fleet code resolve typed scoped clients from `kibana_connection` before calling Kibana, Kibana OpenAPI, SLO, or Fleet sinks.
- Preserve existing Elasticsearch behavior and lint-backed enforcement during this phase so the change can ship independently.
- Keep the existing `kibana_connection` schema surface and rollout scope intact while changing the resolution contract under it.

**Non-Goals:**
- Removing `analysis/esclienthelperplugin` in this phase.
- Converting Elasticsearch sinks or `elasticsearch_connection` handling to typed scoped clients in this phase.
- Redesigning provider-level or entity-level connection block fields.
- Fully deduplicating SDK and Framework decoding paths in `internal/clients/config/`.

## Decisions

Inject a provider client factory for all entities.
The provider should inject `*clients.ProviderClientFactory` into Framework `ProviderData` and SDK `meta` instead of injecting `*clients.APIClient`. This gives the repository one provider data type during the migration and keeps future scoped-client construction centralized.

Alternative considered: keep injecting `*clients.APIClient` for Elasticsearch entities and use a factory only for Kibana/Fleet entities.
Rejected because provider configuration has one injected data contract, and splitting it by entity family would create more migration complexity than it removes.

Ship a transitional factory that supports typed Kibana/Fleet resolution and legacy Elasticsearch resolution.
For this phase, the factory should expose typed Kibana methods that return `*clients.KibanaScopedClient`, while also exposing temporary legacy Elasticsearch resolution methods that preserve current `*clients.APIClient` behavior for unconverted Elasticsearch code paths.

Alternative considered: require phase 1 to convert Kibana/Fleet and Elasticsearch together.
Rejected because the priority is to finalize `kibana_connection` behavior first, while Elasticsearch already has behavior enforcement through the existing lint rule.

Introduce a concrete `KibanaScopedClient` type with Kibana-derived behavior.
The typed Kibana/Fleet client should own the Kibana legacy client, Kibana OpenAPI client, SLO client, Fleet client, Kibana auth context helpers, and Kibana-derived version/flavor checks. It should not expose Elasticsearch identity helpers such as `ClusterID()` because scoped `kibana_connection` intentionally avoids reusing provider-level Elasticsearch identity.

Alternative considered: keep passing `*clients.APIClient` to sinks and only move construction behind the factory.
Rejected because that would still leave sink boundaries too broad for compile-time enforcement.

Move sink boundaries before moving every caller.
Shared helper and sink packages should accept `*clients.KibanaScopedClient` or narrow interfaces derived from it, and Kibana/Fleet resources should resolve the scoped type before calling those helpers. This keeps the compile-time guarantee at the API boundary instead of relying on call-site conventions.

Alternative considered: migrate resources first and leave helper packages broad.
Rejected because the compiler only helps once the sink signatures stop accepting the broad client type.

## Risks / Trade-offs

- [Risk] Introducing the factory for all entities while only typing Kibana/Fleet creates a temporary mixed model -> Mitigation: make the factory's legacy Elasticsearch methods explicit and document that they are transitional only for phase 2 removal.
- [Risk] Kibana/Fleet helper packages currently mix direct client getters with higher-level `APIClient` behaviors -> Mitigation: define the minimum `KibanaScopedClient` surface up front and migrate helper packages to that narrow contract first.
- [Risk] Version checks could accidentally regress to provider-level Elasticsearch behavior during the refactor -> Mitigation: keep `ServerVersion()`, `ServerFlavor()`, and `EnforceMinVersion()` on `KibanaScopedClient` Kibana-derived by contract.
- [Risk] Existing tests and helper utilities may assume `*clients.APIClient` -> Mitigation: add targeted unit coverage for factory resolution and scoped client behavior, and adapt helpers that currently unwrap Kibana clients from the broad type.

## Migration Plan

1. Add `ProviderClientFactory` to `internal/clients/` and update provider configure paths to inject it.
2. Implement typed Kibana/Fleet resolution methods on the factory and add `KibanaScopedClient`.
3. Convert shared Kibana/Fleet helper and sink packages to accept the typed scoped client.
4. Migrate covered Kibana and Fleet resources/data sources to resolve `KibanaScopedClient` from `kibana_connection`.
5. Keep legacy Elasticsearch resolution methods available on the factory until the follow-up Elasticsearch phase replaces them.

## Open Questions

- None. The main design question is sequencing, and this change intentionally fixes that sequence around Kibana/Fleet first.
