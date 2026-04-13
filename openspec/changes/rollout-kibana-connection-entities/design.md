## Context

After provider-level `kibana_connection` support exists, Kibana and Fleet entities still need explicit adoption work. Today many entity specs say the resource or data source uses only the provider-level Kibana or Fleet client, and the code reflects several different client-access patterns: direct `GetKibanaOapiClient()` calls, legacy Kibana client helpers, SLO operations through `*clients.APIClient`, and Fleet operations through `GetFleetClient()`.

Because this rollout touches many entities, the design needs a consistent adoption rule that is broad enough to cover the in-scope registered entities while still allowing entity-specific specs to retain their own space handling, version gates, and import behavior.

## Goals / Non-Goals

**Goals:**
- Define one provider-level rollout capability that identifies the in-scope Kibana and Fleet Terraform entities and requires them to expose and honor `kibana_connection`.
- Update entity specs that currently forbid a resource-level override so their client-resolution requirements match the rollout.
- Preserve entity-specific business behavior while changing only how the effective client is selected.
- Keep the block shape consistent across adopted entities by requiring shared schema helpers instead of per-entity variants.

**Non-Goals:**
- Redefining the internals of the `kibana_connection` helper contract itself.
- Introducing coverage or lint enforcement in this change.
- Expanding the rollout to entities that are intentionally outside the provider's normal registered set for the targeted constructors.

## Decisions

Use a cross-entity provider capability to define rollout scope.
Rather than duplicating the entire rollout rule in every single entity spec, the change should add a provider-level capability that defines which Kibana and Fleet resources and data sources are in scope, the shared schema source of truth, and the default-versus-scoped client-resolution behavior. Entity specs then only need targeted deltas where current requirements explicitly forbid the override.

Alternative considered: update every entity spec independently without a cross-entity capability.
Rejected because it would scatter the same rollout contract across many files and make future additions harder to keep consistent.

Apply `kibana_connection` uniformly to resources and data sources in scope.
The rollout should cover both resources and data sources for in-scope Kibana and Fleet entities, matching the provider's existing multi-entity connection patterns and reducing special cases for users.

Alternative considered: resources only.
Rejected because Fleet and Kibana data sources also make client-scoped API calls and would otherwise remain inconsistent with the resource story.

Retain entity-specific lifecycle and API semantics while changing only the effective client source.
Space handling, import behavior, version gates, and request/response mapping should remain owned by each entity spec. The rollout only changes whether those operations run against the provider client or a scoped client derived from `kibana_connection`.

Alternative considered: restate broader entity behavior in the rollout capability.
Rejected because it would duplicate entity-owned requirements and create drift risk.

Target the provider's normal registered entity set for the chosen constructors.
The rollout capability should describe the same class of entities that the provider exposes through its normal constructors, so the change matches the user's visible provider surface and stays aligned with the verification change.

Alternative considered: include experimental entities unconditionally.
Rejected because the default provider constructors do not always register them, and that would complicate rollout and verification scoping.

## Risks / Trade-offs

- [Risk] Some entities use different client surfaces, so rollout can become mechanically repetitive -> Mitigation: define the shared rollout contract once and keep entity deltas narrowly focused on connection-resolution behavior.
- [Risk] Existing specs that forbid overrides could become inconsistent with the new rollout capability -> Mitigation: explicitly modify every contradictory spec in this change.
- [Risk] Normal-registered-set scoping may leave experimental entities for later follow-up -> Mitigation: treat experimental adoption as an intentional future change rather than implicit scope creep.

## Migration Plan

1. Add a new provider-level rollout capability describing the in-scope Kibana and Fleet entities and the shared `kibana_connection` contract they adopt.
2. Update the contradictory Kibana and Fleet entity specs so they no longer require provider-only client resolution.
3. Leave verification and implementation details to the adjacent support and verification changes.

## Open Questions

- None. The main remaining choice is implementation sequencing, which is handled outside this proposal.
