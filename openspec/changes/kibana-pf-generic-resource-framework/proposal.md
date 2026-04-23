## Why

Plugin Framework resources in this repository repeatedly implement the same lifecycle scaffolding: provider configuration, metadata wiring, import handling, Kibana scoped-client resolution, minimum-version enforcement, composite-ID parsing, CRUD orchestration, and read-after-write state refresh. Issue #2423 highlights the duplication in the three Agent Builder resources, but the same pattern appears across many Kibana Plugin Framework resources.

Treating this only as a one-off cleanup for Agent Builder would leave the repository without a clear reusable pattern for future Plugin Framework migrations. The provider needs a shared, maintainable foundation for Kibana CRUD resources that reduces drift while keeping resource-specific schema and API semantics explicit.

## What Changes

- Introduce a shared Kibana Plugin Framework generic resource framework that separates Terraform lifecycle orchestration, Terraform model/request mapping, and Kibana API transport.
- Define a small implementation assembly surface plus typed model and API interfaces so resource packages keep schema and business mapping logic while reusing common CRUD mechanics.
- Reorganize the relevant Agent Builder Kibana API helper functions behind focused API interfaces that fit the new generic resource framework.
- Migrate the Agent Builder resources (`agent`, `tool`, and `workflow`) to the new framework as the initial adopters.
- Keep the framework scoped to Kibana Plugin Framework CRUD resources for now, allowing later incremental adoption by other resources without forcing repository-wide migration in this change.

## Capabilities

### New Capabilities
- `kibana-plugin-framework-generic-resource`: Provide a reusable framework for Kibana Plugin Framework CRUD resources that centralizes common lifecycle orchestration while delegating schema, model mapping, and resource-specific transport to typed components.

### Modified Capabilities
- `kibana-agentbuilder-agent`: Implement the resource through the shared Kibana Plugin Framework generic resource framework instead of bespoke CRUD scaffolding.
- `kibana-agentbuilder-tool`: Implement the resource through the shared Kibana Plugin Framework generic resource framework instead of bespoke CRUD scaffolding.
- `kibana-agentbuilder-workflow`: Implement the resource through the shared Kibana Plugin Framework generic resource framework instead of bespoke CRUD scaffolding.

## Impact

- New shared package(s) under `internal/kibana/` for generic Plugin Framework resource orchestration and helpers.
- Focused Kibana API helper interfaces/packages under `internal/clients/kibanaoapi/` for Agent Builder transport operations.
- Agent Builder resource packages under `internal/kibana/agentbuilderagent/`, `internal/kibana/agentbuildertool/`, and `internal/kibana/agentbuilderworkflow/`.
- OpenSpec delta specs for the new generic-resource capability and the modified Agent Builder capabilities.
