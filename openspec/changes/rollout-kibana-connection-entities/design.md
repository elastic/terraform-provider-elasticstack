## Context

After provider-level `kibana_connection` support exists, Kibana and Fleet entities still need explicit adoption work. Today many entity specs say the resource or data source uses only the provider-level Kibana or Fleet client, and the code reflects several different client-access patterns: direct `GetKibanaOapiClient()` calls, legacy Kibana client helpers, SLO operations through `*clients.APIClient`, and Fleet operations through `GetFleetClient()`. The current rollout artifacts also undercount the registered Kibana and Fleet entities, so the change needs an explicit registration-based scope that matches both `provider/provider.go` and `provider/plugin_framework.go`.

Because this rollout touches many entities, the design needs a consistent adoption rule that is broad enough to cover the in-scope registered entities while still allowing entity-specific specs to retain their own space handling, version gates, and import behavior.

## Goals / Non-Goals

**Goals:**
- Define one provider-level rollout capability that identifies the in-scope Kibana and Fleet Terraform entities directly from the provider registration lists and requires them to expose and honor `kibana_connection`.
- Update entity specs that currently forbid a resource-level override so their client-resolution requirements match the rollout.
- Preserve entity-specific business behavior while changing only how the effective client is selected.
- Keep the block shape consistent across adopted entities by requiring shared schema helpers instead of per-entity variants.
- Include Plugin Framework experimental Kibana registrations that are still part of the provider source surface, so the rollout plan does not miss `elasticstack_kibana_dashboard` or `elasticstack_kibana_stream`.

**Non-Goals:**
- Redefining the internals of the `kibana_connection` helper contract itself.
- Introducing coverage or lint enforcement in this change.
- Re-describing full entity behavior that already belongs in each entity's own capability spec.

## Rollout Scope

The rollout scope is the union of the Kibana and Fleet entity registrations in `provider/provider.go` and `provider/plugin_framework.go`.

`provider/provider.go` registrations:
- `elasticstack_kibana_action_connector`
- `elasticstack_kibana_security_role`
- `elasticstack_kibana_space`

`provider/plugin_framework.go` normal registrations:
- `elasticstack_fleet_agent_policy`
- `elasticstack_fleet_elastic_defend_integration_policy`
- `elasticstack_fleet_enrollment_tokens`
- `elasticstack_fleet_integration`
- `elasticstack_fleet_integration_policy`
- `elasticstack_fleet_output`
- `elasticstack_fleet_server_host`
- `elasticstack_kibana_action_connector`
- `elasticstack_kibana_agentbuilder_export_workflow`
- `elasticstack_kibana_agentbuilder_workflow`
- `elasticstack_kibana_alerting_rule`
- `elasticstack_kibana_data_view`
- `elasticstack_kibana_default_data_view`
- `elasticstack_kibana_export_saved_objects`
- `elasticstack_kibana_import_saved_objects`
- `elasticstack_kibana_maintenance_window`
- `elasticstack_kibana_security_detection_rule`
- `elasticstack_kibana_security_enable_rule`
- `elasticstack_kibana_security_exception_item`
- `elasticstack_kibana_security_exception_list`
- `elasticstack_kibana_security_list`
- `elasticstack_kibana_security_list_data_streams`
- `elasticstack_kibana_security_list_item`
- `elasticstack_kibana_slo`
- `elasticstack_kibana_spaces`
- `elasticstack_kibana_synthetics_monitor`
- `elasticstack_kibana_synthetics_parameter`
- `elasticstack_kibana_synthetics_private_location`

`provider/plugin_framework.go` conditional experimental registrations:
- `elasticstack_kibana_dashboard`
- `elasticstack_kibana_stream`

## Decisions

Use a cross-entity provider capability to define rollout scope from provider registrations.
Rather than duplicating the entire rollout rule in every single entity spec, the change should add a provider-level capability that defines the exact Kibana and Fleet resources and data sources registered in `provider/provider.go` and `provider/plugin_framework.go`, the shared schema source of truth, and the default-versus-scoped client-resolution behavior. Entity specs then only need targeted deltas where current requirements explicitly forbid the override or do not yet describe the new optional block.

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

Target the provider's registered entity set across legacy and Plugin Framework code paths.
The rollout capability should describe the same Kibana and Fleet entities that the provider source registers in `provider/provider.go` and `provider/plugin_framework.go`, even when some Plugin Framework registrations are conditional. That keeps the rollout plan aligned with the code that defines the provider surface and avoids accidental omissions such as `elasticstack_kibana_stream`.

Alternative considered: infer scope only from default constructor behavior.
Rejected because that would miss SDK-only registrations and conditionally registered Plugin Framework entities even though they are part of the provider-maintained rollout surface in source.

## Risks / Trade-offs

- [Risk] Some entities use different client surfaces, so rollout can become mechanically repetitive -> Mitigation: define the shared rollout contract once and keep entity deltas narrowly focused on connection-resolution behavior.
- [Risk] Existing specs that forbid overrides could become inconsistent with the new rollout capability -> Mitigation: explicitly modify every contradictory spec in this change.
- [Risk] Registration-based scoping includes experimental Plugin Framework entities whose standalone OpenSpec coverage is incomplete -> Mitigation: track them explicitly in the provider-level rollout capability and task plan so implementation cannot omit them, while keeping entity-specific deltas focused on capabilities that already exist.

## Migration Plan

1. Add a new provider-level rollout capability describing the exact Kibana and Fleet registrations from `provider/provider.go` and `provider/plugin_framework.go` and the shared `kibana_connection` contract they adopt.
2. Update the contradictory Kibana and Fleet entity specs so they no longer require provider-only client resolution.
3. Keep the implementation task plan aligned with the full registration list, including `elasticstack_kibana_stream`, even where standalone entity capability specs are not yet present.
4. Leave verification and implementation details to the adjacent support and verification changes.

## Open Questions

- None. The main remaining choice is implementation sequencing, which is handled outside this proposal.
