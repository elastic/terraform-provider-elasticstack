## Context

The provider already injects `*clients.ProviderClientFactory` into Framework `ProviderData` and SDK `meta`, and most entities now resolve either `*clients.KibanaScopedClient` or `*clients.ElasticsearchScopedClient` from that factory. The remaining gaps are concentrated in `internal/clients/api_client.go`, where a broad `APIClient` still owns duplicated helper behavior and legacy bridge helpers, and in `internal/apm/agent_configuration`, which still converts provider data back into a broad client.

This cleanup spans production code, tests, exported compatibility surfaces, and synced OpenSpec specs. The main constraint is that the factory still needs an internal provider-default client representation to bootstrap typed scoped clients, even after the broad client stops being part of the supported contract.

## Goals / Non-Goals

**Goals:**
- Remove remaining production reliance on broad `APIClient` resolution.
- Make typed factory/scoped-client resolution the only supported provider client contract.
- Shrink `internal/clients` so Kibana- and Elasticsearch-specific helper behavior lives on the corresponding scoped clients.
- Capture the remaining requirement changes in OpenSpec so implementation and specs converge again.

**Non-Goals:**
- Rework provider configuration parsing or connection schema shape.
- Expand the rollout to new resources beyond the existing cleanup target.
- Fully redesign `xpprovider`; this change only narrows or replaces the parts that expose the broad client.
- Eliminate every internal bootstrap helper immediately if a small private helper remains useful during the refactor.

## Decisions

### Decision: Migrate APM agent configuration directly to typed Kibana client resolution

`internal/apm/agent_configuration` only needs Kibana OpenAPI access. It will stop storing `*clients.APIClient` and instead resolve a typed Kibana client from `ProviderClientFactory`, then call `GetKibanaOapiClient()` on that typed client.

This keeps the resource aligned with the same typed contract already used elsewhere and removes the last known Framework production use of `ConvertProviderData`.

Alternatives considered:
- Keep using `ConvertProviderData` and accept one broad-client exception. Rejected because it preserves the exact bridge this cleanup is meant to remove.
- Introduce a new APM-specific provider data adapter. Rejected because it adds another special case instead of converging on the factory pattern.

### Decision: Keep a private provider-default bootstrap client, but remove it from the supported API surface

The factory still needs a provider-default object that contains the clients and configuration necessary to construct typed scoped clients. The broad client type can remain as a private implementation detail inside `internal/clients`, while exported accessors and bridge helpers are removed.

This allows the factory to keep reusing internal construction logic without advertising a public escape hatch back to the broad client.

Alternatives considered:
- Delete the broad client struct entirely in the same change. Rejected because the factory and acceptance bootstrap still need a shared internal assembly point.
- Keep `GetDefaultClient()` exported for internal convenience. Rejected because it contradicts the typed factory contract and invites new broad-client call sites.

### Decision: Move duplicated helper behavior to the owning scoped client and update tests accordingly

Elasticsearch-derived behavior such as `GetESClient`, cluster identity, version/flavor lookup, and minimum-version enforcement belongs on `ElasticsearchScopedClient`. Kibana-derived behavior such as `GetKibanaOapiClient`, `GetFleetClient`, SLO auth context, and Kibana version/flavor checks belongs on `KibanaScopedClient`.

Tests and acceptance helpers will be updated to construct scoped clients directly rather than reaching those helpers through `NewAcceptanceTestingClient()`.

Alternatives considered:
- Leave duplicated methods on both the private broad client and the scoped clients. Rejected because that keeps behavior split across two contracts and weakens compile-time enforcement.
- Replace all test helpers with ad hoc client construction in each package. Rejected because the repo already benefits from shared acceptance bootstrap helpers.

### Decision: Treat `xpprovider` broad-client export removal as an explicit compatibility change

`xpprovider` currently aliases `clients.APIClient`, so privatizing the type changes the exported surface. The change proposal will treat this as a breaking compatibility edge and require the implementation to either expose typed replacements or intentionally narrow the public API.

Alternatives considered:
- Preserve the alias indefinitely. Rejected because it blocks the cleanup goal of making `APIClient` private.
- Defer `xpprovider` until a later cleanup. Rejected because it leaves the most visible external leak of the broad client in place.

## Risks / Trade-offs

- `xpprovider` consumers may rely on `APIClient` today -> Mitigation: call out the break in the proposal, replace it with typed factory/scoped surfaces where possible, and verify downstream compile failures early.
- Acceptance test churn may be larger than production code churn -> Mitigation: centralize replacement helpers first, then migrate package tests mechanically.
- Private bootstrap logic may still temporarily resemble the old broad client -> Mitigation: remove exported access paths in the same change, then trim remaining private-only methods once no callers require them.
- Some synced specs still reference deleted helper names -> Mitigation: include delta specs in this change and update canonical specs when the change is applied and archived.

## Migration Plan

1. Move `internal/apm/agent_configuration` onto factory-resolved typed Kibana client usage.
2. Remove production bridge helpers and `GetDefaultClient()` from the supported factory surface.
3. Privatize the broad client type and delete duplicated exported helper behavior that scoped clients now own.
4. Update `xpprovider`, tests, and acceptance helpers to consume typed surfaces.
5. Sync OpenSpec capability deltas and canonical specs, then validate with the repo's spec checks.

Rollback is low risk because this is a source-level refactor: reverting the change restores the previous bridges and export surface.

## Open Questions

- Whether `xpprovider` should expose the factory directly, expose typed scoped constructors, or deliberately drop the client-construction surface altogether.
- Whether any internal test-only helpers still need a small private bootstrap wrapper after `NewAcceptanceTestingClient()` stops being the default path.
