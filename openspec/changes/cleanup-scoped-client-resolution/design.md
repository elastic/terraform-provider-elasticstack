## Context

The provider already injects `*clients.ProviderClientFactory` into Framework `ProviderData` and SDK `meta`, and most entities now resolve either `*clients.KibanaScopedClient` or `*clients.ElasticsearchScopedClient` from that factory. The remaining gaps are concentrated in `internal/clients/api_client.go`, where a broad `APIClient` still owns duplicated helper behavior and legacy bridge helpers, and in `internal/apm/agent_configuration`, which still converts provider data back into a broad client.

This cleanup spans production code, tests, exported compatibility surfaces, shared entity-local connection schema helpers, and synced OpenSpec specs. The main constraint is that the factory still needs an internal provider-default client representation to bootstrap typed scoped clients, even after the broad client stops being part of the supported contract.

## Goals / Non-Goals

**Goals:**
- Add `kibana_connection` to `elasticstack_apm_agent_configuration` without inventing a resource-specific connection schema.
- Make provider connection-schema coverage follow the real provider registry instead of partial naming heuristics or hand-maintained ownership inventories, so future entities cannot be omitted silently.
- Remove remaining production reliance on broad `APIClient` resolution.
- Make typed factory/scoped-client resolution the only supported provider client contract.
- Shrink `internal/clients` so Kibana- and Elasticsearch-specific helper behavior lives on the corresponding scoped clients.
- Capture the remaining requirement changes in OpenSpec so implementation and specs converge again.

**Non-Goals:**
- Rework provider-level configuration parsing or invent a new APM-specific connection schema shape.
- Expand the rollout to new resources beyond the existing cleanup target.
- Fully redesign `xpprovider`; this change only narrows or replaces the parts that expose the broad client.
- Eliminate every internal bootstrap helper immediately if a small private helper remains useful during the refactor.

## Decisions

### Decision: Migrate APM agent configuration directly to typed Kibana client resolution

`internal/apm/agent_configuration` only needs Kibana OpenAPI access. It will add the shared Plugin Framework `kibana_connection` block so the resource can either inherit provider defaults or scope Kibana-derived operations per resource. The resource will stop storing `*clients.APIClient` and instead resolve a typed Kibana client from `ProviderClientFactory`, then call `GetKibanaOapiClient()` on that typed client.

This keeps the resource aligned with the same typed contract already used elsewhere, closes the requirements gap called out in review, and removes the last known Framework production use of `ConvertProviderData`.

Alternatives considered:
- Keep the resource on provider-default resolution only and just remove the misleading override wording from the spec. Rejected because the cleanup is already touching APM client resolution, and adding the shared block now keeps the resource aligned with the broader `kibana_connection` model instead of leaving APM as a special case.
- Keep using `ConvertProviderData` and accept one broad-client exception. Rejected because it preserves the exact bridge this cleanup is meant to remove.
- Introduce a new APM-specific provider data adapter. Rejected because it adds another special case instead of converging on the factory pattern.

### Decision: Use a single registry-driven connection-schema coverage test

The provider coverage tests will stop acting like independent prefix scans or hand-maintained ownership fixtures. Instead, a single test will enumerate every entity registered by `provider.New(...)` and `provider.NewFrameworkProvider(...)`, run one subtest per registered entity, validate the expected connection block contract for that entity, record that the entity was exercised, and finish with a completeness assertion that every registered entity was validated.

For the current provider surface, registered `elasticstack_elasticsearch_*` entities are expected to expose `elasticsearch_connection`, while other registered entities are expected to expose `kibana_connection`. The test will keep the provider constructors as the source of truth for what exists, and the final completeness subtest will ensure no registered entity escapes validation. Where a registered entity intentionally lacks a connection block, that exception must be asserted explicitly in the same test so it remains visible and reviewable. The current documented exception is the SDK `elasticstack_elasticsearch_ingest_processor_*` data sources, which build ingest processor payloads only and therefore assert the absence of both connection blocks.

This keeps test coverage aligned with the actual provider surface, eliminates duplicated ownership inventories, and turns future registry changes into immediate test failures instead of silent omissions while keeping the current no-connection carve-out explicit.

Alternatives considered:
- Keep prefix-based selection split across multiple tests. Rejected because it still separates the per-entity assertions from the completeness guarantee and makes it easier for the two to drift.
- Maintain separate hand-written ownership inventories in test code. Rejected because drift between those inventories and the provider registrations is exactly the class of failure this change is trying to avoid.

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
- Connection-schema coverage can drift from the real provider registry -> Mitigation: derive the full registered entity inventory from provider constructors, validate each entity in a single test loop, and fail if the final completeness check finds anything unvalidated.
- Private bootstrap logic may still temporarily resemble the old broad client -> Mitigation: remove exported access paths in the same change, then trim remaining private-only methods once no callers require them.
- Some synced specs still reference deleted helper names -> Mitigation: include delta specs in this change and update canonical specs when the change is applied and archived.

## Migration Plan

1. Add the shared `kibana_connection` block to `internal/apm/agent_configuration` and move it onto factory-resolved typed Kibana client usage.
2. Remove production bridge helpers and `GetDefaultClient()` from the supported factory surface.
3. Privatize the broad client type and delete duplicated exported helper behavior that scoped clients now own.
4. Update `xpprovider`, tests, and acceptance helpers to consume typed surfaces.
5. Sync OpenSpec capability deltas and canonical specs, then validate with the repo's spec checks.

Rollback is low risk because this is a source-level refactor: reverting the change restores the previous bridges and export surface.

## Open Questions

- Whether `xpprovider` should expose the factory directly, expose typed scoped constructors, or deliberately drop the client-construction surface altogether.
- Whether any internal test-only helpers still need a small private bootstrap wrapper after `NewAcceptanceTestingClient()` stops being the default path.
