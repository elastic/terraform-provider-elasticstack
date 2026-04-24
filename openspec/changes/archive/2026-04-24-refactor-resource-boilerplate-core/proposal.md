## Why

Plugin Framework resources in this repository repeat the same `client *clients.ProviderClientFactory`, `Configure`, and `Metadata` wiring with only a component-prefixed resource name changing between packages. That duplication makes consistency harder to maintain, obscures the intended canonical `Configure` behavior, and slows down reviews when boilerplate changes need to be applied across many resources.

## What Changes

- Add a provider-wide Plugin Framework resource core that owns canonical client-factory wiring and resource type-name construction from a typed component plus resource name.
- Define a typed component namespace for well-known values, including `elasticsearch`, `kibana`, `fleet`, and `apm`.
- Restrict the shared core to `Configure` and `Metadata`; keep `ImportState` explicit on each resource.
- Trial the embedded-core rollout on `elasticstack_elasticsearch_ml_job_state`, `elasticstack_kibana_agentbuilder_tool`, `elasticstack_fleet_integration`, and `elasticstack_apm_agent_configuration`.
- Add compile-time and targeted test coverage that protects against accidental interface drift from promoted methods during the rollout.

## Capabilities

### New Capabilities

- `provider-framework-resource-core`: Define the shared Plugin Framework resource-core contract for provider client handling, typed component-based type-name generation, and rollout safety rules for embedded usage.

### Modified Capabilities

<!-- None -->

## Impact

- New provider-wide logical package under `internal/` for the shared resource core and typed component constants.
- Pilot resource wiring updates in:
  - `internal/elasticsearch/ml/jobstate`
  - `internal/kibana/agentbuildertool`
  - `internal/fleet/integration`
  - `internal/apm/agent_configuration`
- New or updated tests covering embedded-core conformance and pilot resource behavior.
- New OpenSpec capability at `openspec/changes/refactor-resource-boilerplate-core/specs/provider-framework-resource-core/spec.md`
