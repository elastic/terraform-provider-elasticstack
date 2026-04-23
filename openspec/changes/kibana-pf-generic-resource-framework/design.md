## Context

Issue #2423 identifies near-identical CRUD scaffolding across the three Agent Builder Plugin Framework resources. The duplicated code covers plan/state decoding, provider-configured Kibana client resolution, minimum-version enforcement, OAPI client acquisition, composite-ID parsing, remote CRUD calls, read-after-write reconciliation, and state persistence. Similar scaffolding exists in other Kibana Plugin Framework resources, although many of those resources also have more specialized semantics.

The target design should establish a reusable foundation for Kibana Plugin Framework CRUD resources without collapsing resource-specific schema and mapping semantics into an opaque framework. The initial migration target is the Agent Builder resources because they are the clearest examples of the shared pattern and already drift in small but important ways.

## Goals / Non-Goals

**Goals:**
- Create a shared framework for Kibana Plugin Framework CRUD resources that owns common Terraform lifecycle orchestration.
- Keep Terraform model mapping separate from Kibana transport operations.
- Make the framework reusable for later Kibana Plugin Framework migrations without requiring immediate adoption by unrelated resources.
- Migrate Agent Builder agent, tool, and workflow resources to prove the abstraction.
- Preserve existing external behavior for the migrated resources unless an explicit consistency fix is required.

**Non-Goals:**
- Creating a single generic framework for all provider resources across Kibana, Fleet, and Elasticsearch in this change.
- Migrating data sources in the first iteration unless required by the implementation shape.
- Rewriting all existing `kibanaoapi` helpers into a universal transport layer.
- Changing Agent Builder schemas, import formats, or documented resource behavior beyond what is necessary to preserve current semantics through the new abstraction.

## Architecture

The framework is split into three layers:

```text
GenericResource
  = Terraform Plugin Framework lifecycle orchestration

TModel
  = Terraform model semantics, version requirements, request building,
    and remote-to-state mapping

ResourceAPI
  = Kibana transport operations for one resource family
```

### GenericResource responsibilities

The shared generic resource owns:
- `Configure` provider-data conversion and client-factory storage
- `Metadata` type-name wiring
- standard import helper selection where appropriate
- plan/state decoding into the concrete model type
- Kibana scoped-client resolution from the model's `kibana_connection`
- minimum-version enforcement using the model's computed version requirement
- common create/read/update/delete orchestration
- read-after-write reconciliation for create and update
- state removal on remote-not-found during read
- state persistence after model population

The generic resource does **not** own resource schema details, request payload construction, or endpoint-specific behavior.

### TModel responsibilities

Each resource model owns:
- access to `kibana_connection`
- access to `id` and, when relevant, `space_id`
- conversion from Terraform state/plan into typed create/update request objects
- conversion from remote API objects into Terraform state
- calculation of minimum Kibana version requirements based on the configured shape

The model does **not** call Kibana transport operations directly.

### ResourceAPI responsibilities

Each focused API implementation owns:
- `Create`
- `Get`
- `Update`
- `Delete`

against a Kibana scoped client and typed request/response objects. This keeps endpoint-specific transport logic out of the model and allows the existing `kibanaoapi` helper surface to be reorganized into resource-family-focused APIs.

## Proposed package layout

```text
internal/kibana/pfresource/
  generic_resource.go   # shared TF lifecycle orchestration
  base.go               # configure/metadata/import helpers
  client.go             # Kibana scoped-client resolution + version enforcement
  id.go                 # composite-ID helpers for space-aware resources

internal/clients/kibanaoapi/agentbuilderapi/
  agents.go
  tools.go
  workflows.go
```

Resource packages keep their schema and model code, but replace bespoke CRUD files with a small implementation assembly layer.

Example target shape:

```text
internal/kibana/agentbuildertool/
  resource.go          # assembles generic resource
  implementation.go    # binds schema + model + API
  models.go            # TF model, request mapping, response mapping
  schema.go
```

## Interface design

### Implementation assembly

The resource package provides a small assembly object that binds together the generic resource, the concrete model type, and the focused API implementation.

Conceptually:
- type-name suffix
- schema
- model constructor
- API provider
- optional import strategy hook if a resource cannot use shared import helpers

This surface stays intentionally small so resource packages are not forced to implement framework-shaped glue methods for every CRUD operation.

### Model contract

The model contract covers:
- connection access
- identifier accessors/mutators
- optional space-aware accessors/mutators
- version requirement computation
- `ToCreateRequest`
- `ToUpdateRequest`
- `PopulateFromRemote`

The version requirement includes both the minimum version and the user-facing unsupported-version message so resource-specific wording remains explicit and future attribute-dependent version rules can be expressed without hardcoding them in the generic layer.

### API contract

The API contract covers:
- `Create(scopedClient, spaceID, request) -> created ID`
- `Get(scopedClient, spaceID, resourceID) -> remote object or nil`
- `Update(scopedClient, spaceID, resourceID, request)`
- `Delete(scopedClient, spaceID, resourceID)`

The generic resource always re-reads after create and update to keep state authoritative and consistent with current Agent Builder behavior.

## ID and import handling

The first version of the framework should explicitly support Kibana resources that use composite IDs of the form `<space_id>/<resource_id>`. Shared helpers should cover:
- parsing composite IDs from state/import strings
- default-space normalization where the resource model permits omitted `space_id`
- restoring `space_id` into the model before response mapping when the model requires it

Import behavior should be helper-driven rather than fully hardcoded in the generic resource, because some existing Kibana resources use passthrough `id` while others decompose the import string into multiple attributes.

## Agent Builder adoption plan

The first adopters are:
- `elasticstack_kibana_agentbuilder_agent`
- `elasticstack_kibana_agentbuilder_tool`
- `elasticstack_kibana_agentbuilder_workflow`

These resources already share the same CRUD shell, but differ in schema and mapping behavior:
- agent uses typed fields plus sets and passes `space_id` explicitly into state population
- tool uses normalized JSON configuration and restores `space_id` from the composite ID before population
- workflow uses normalized YAML configuration, optional computed workflow IDs, and post-write invalid-workflow diagnostics

The framework must preserve these differences as model/API concerns rather than flattening them away.

## Risks / Trade-offs

- **The abstraction may become too generic too quickly** -> Mitigation: explicitly scope the framework to Kibana Plugin Framework CRUD resources and prove it on Agent Builder first.
- **Generics and interface layering can become hard to follow** -> Mitigation: keep the assembly interface small and keep schema and mapping code in the resource packages.
- **`kibanaoapi` reorganization could create churn beyond the initial migration** -> Mitigation: only introduce focused APIs needed by the Agent Builder migration in this change.
- **Migrated resources could accidentally change diagnostics or state behavior** -> Mitigation: preserve read-after-write semantics, import behavior, version-gate wording, and acceptance coverage for Agent Builder resources.

## Migration Plan

1. Introduce the shared Kibana Plugin Framework resource package with client/version, ID, and generic CRUD orchestration helpers.
2. Introduce focused Agent Builder transport APIs under `internal/clients/kibanaoapi/` that fit the new generic resource contracts.
3. Migrate the Agent Builder agent, tool, and workflow resources to the new framework while preserving their schemas and current external behavior.
4. Run unit and acceptance coverage for the migrated resources and validate that the framework leaves room for later opt-in adoption by other Kibana Plugin Framework resources.

## Open Questions

- Whether the first implementation should include only a shared composite-ID import helper or also a passthrough import helper abstraction in the same package.
- Whether the framework should expose optional post-read or post-write hooks immediately, or wait until a second adopter beyond Agent Builder demonstrates the need.
