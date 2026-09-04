## ADDED Requirements

### Requirement: Resource type and stable registration

The `elasticstack_fleet_managed_integration` resource SHALL be registered in the provider's stable resource list (`resources()` in `provider/plugin_framework.go`) and SHALL NOT appear in `experimentalResources()`. It SHALL therefore be present in the provider schema with no opt-in, independent of the `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` environment variable and of the `acctest` provider version sentinel (`AccTestVersion` in `provider/plugin_framework.go`). It SHALL be registered exactly once, so that enabling the experimental opt-in does not register the type twice. The resource description SHALL NOT describe the resource as experimental; it SHALL continue to state the Kibana 9.5.0 minimum version and the Elastic Cloud Hosted / Serverless-only deployment constraint.

#### Scenario: Resource type is discoverable
- **WHEN** `terraform providers schema -json` is run against the provider
- **THEN** `elasticstack_fleet_managed_integration` SHALL appear in the resource schema

#### Scenario: Resource available without the experimental opt-in
- **WHEN** the provider is initialised with `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` unset and a released provider version
- **THEN** `elasticstack_fleet_managed_integration` SHALL be registered

#### Scenario: Experimental opt-in does not duplicate registration
- **WHEN** the provider is initialised with `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` set to `true`
- **THEN** `elasticstack_fleet_managed_integration` SHALL be registered exactly once

#### Scenario: Resource description makes no experimental claim
- **WHEN** the resource schema description is inspected
- **THEN** it SHALL NOT state that the resource is experimental
- **AND** it SHALL still state the Kibana 9.5.0 minimum version and the Elastic Cloud Hosted / Serverless-only constraint

#### Scenario: Documentation page is generated
- **WHEN** provider documentation is generated with the experimental opt-in disabled
- **THEN** a documentation page for `elasticstack_fleet_managed_integration` SHALL be produced

## REMOVED Requirements

### Requirement: Resource type and registration

**Reason**: The requirement mandated registration in `experimentalResources()` to match the upstream tech-preview status of the Fleet `managed_integrations` API. That API is no longer experimental, so the requirement — and its "Resource registered as experimental" scenario — no longer describes intended behaviour. Replaced by **Resource type and stable registration**, which carries over the discoverability scenario (retitled from "New resource type is discoverable"). The old requirement's `elasticstack_fleet_agentless_policy` absence clause and scenario line are not carried over: the dedicated requirement **elasticstack_fleet_agentless_policy resource — REMOVED** already owns that behaviour.

**Migration**: None for practitioners. Configurations that exported `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` solely to reach this resource can drop the variable; leaving it set remains harmless and is still required for `elasticstack_kibana_tag`.
