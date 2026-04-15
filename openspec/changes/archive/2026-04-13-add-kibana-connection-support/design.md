## Context

The provider already has shared schema helpers for `kibana_connection`, but it does not have a dedicated helper path for constructing a scoped `*clients.APIClient` from that block. The existing Framework connector implementation currently routes `kibana_connection` through the Elasticsearch-specific helper, which preserves the provider-level Kibana and Fleet clients instead of rebuilding them from the scoped connection.

Kibana and Fleet entities also rely on `*clients.APIClient` helpers for version checks, cluster identity, Kibana legacy APIs, Kibana OpenAPI APIs, SLO APIs, and Fleet APIs. A real provider-level `kibana_connection` contract therefore needs to define how those client surfaces are rebuilt together and how the resulting scoped client behaves when version or identity lookups are performed.

## Goals / Non-Goals

**Goals:**
- Define a provider-level contract for resource-scoped `kibana_connection` that is parallel to the existing `elasticsearch_connection` pattern.
- Add dedicated SDK and Framework helper paths that build scoped Kibana-derived clients from `kibana_connection`.
- Ensure scoped `kibana_connection` affects all Kibana-derived client surfaces used by Kibana and Fleet entities: legacy Kibana client, Kibana OpenAPI client, SLO client, Fleet client, and Kibana-derived version checks.
- Establish a safe foundation that rollout and verification changes can depend on.

**Non-Goals:**
- Rolling the new block out to every Kibana or Fleet entity in this change.
- Defining entity-by-entity schema coverage or lint enforcement in this change.
- Redesigning the fields inside the provider-level `kibana` or `fleet` configuration blocks.

## Decisions

Introduce dedicated `kibana_connection` client-resolution helpers for SDK and Framework code.
The provider should not reuse `NewAPIClientFromSDKResource(...)` or `MaybeNewAPIClientFromFrameworkResource(...)` for `kibana_connection`, because those helpers are Elasticsearch-scoped by contract and by implementation. New helper paths should accept resource-local `kibana_connection` input and return a fully scoped `*clients.APIClient`.

Alternative considered: extend the Elasticsearch helpers to accept both connection block types.
Rejected because it would blur two different contracts and make future lint and coverage rules harder to reason about.

Rebuild all Kibana-derived clients together from the scoped connection.
The scoped client should rebuild the Kibana legacy client, Kibana OpenAPI client, SLO client, and Fleet client from the `kibana_connection` block, rather than swapping only one client surface. This keeps all Kibana/Fleet operations on the same target cluster and avoids mixed-cluster behavior.

Alternative considered: rebuild only the client directly used by the current entity.
Rejected because many entities use multiple `*clients.APIClient` behaviors indirectly, including version checks and Fleet client construction, so partial rebuilding would leave surprising cross-cluster drift.

Make version and identity checks Kibana-derived when the scoped client does not have an Elasticsearch client.
When a scoped `kibana_connection` is in use, the resulting `*clients.APIClient` may intentionally omit the provider-level Elasticsearch client so that `ServerVersion()`, `ServerFlavor()`, and related behavior resolve against the overridden Kibana endpoint instead of the provider's Elasticsearch endpoint.

Alternative considered: carry the provider-level Elasticsearch client through unchanged.
Rejected because it would let a resource-scoped Kibana connection talk to one cluster while version and identity checks come from another.

Keep the schema shape aligned with the existing shared `GetKbFWConnectionBlock()` and `GetKibanaConnectionSchema()` definitions.
This change should define client-resolution semantics, not redesign the schema fields. Any future schema field changes should continue to flow through the existing shared helper builders.

## Risks / Trade-offs

- [Risk] Scoped Kibana-derived clients may diverge from provider-level Elasticsearch identity checks that some code paths implicitly expect -> Mitigation: define the scoped-client behavior explicitly so rollout changes can update entity requirements and tests accordingly.
- [Risk] Fleet client construction depends on Kibana-derived configuration conventions -> Mitigation: build the scoped Fleet client from the same scoped Kibana-derived config path used by provider initialization.
- [Risk] Existing connector behavior may already be relying on the broken helper path -> Mitigation: make the provider-level contract explicit and require connector behavior to align with it as part of implementation.

## Migration Plan

1. Add a new provider-level `provider-kibana-connection` capability spec.
2. Implement dedicated SDK and Framework helper paths that construct scoped Kibana-derived `*clients.APIClient` values from `kibana_connection`.
3. Align the existing action connector implementation with the new helper path so the current `kibana_connection` block becomes behaviorally correct.
4. Leave entity rollout and verification enforcement to the follow-up changes that depend on this provider-level contract.

## Open Questions

- None. The remaining uncertainty is rollout scope, which is intentionally separated into its own change.
