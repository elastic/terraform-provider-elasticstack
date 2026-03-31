## Context

This change is documentation-first: the branch already contains the implementation, generated docs, and acceptance tests for the new Agent Builder workflow resource and export data source. The OpenSpec work here should capture only behavior that is traceable to the current code paths and supporting tests.

The new entities are both Kibana OpenAPI-based and share the same product gate:

- minimum supported Elastic Stack version `9.4.0-SNAPSHOT`
- composite identity format `<space_id>/<workflow_id>` for persisted state
- normalized YAML handling through `customtypes.NormalizedYamlType`

## Goals

- Define one stable capability id per new entity.
- Capture the Terraform-facing schema for both entities in a way that is easy to sync into canonical specs later.
- Document the exact identity, import, read, create/update, and YAML behaviors implemented on this branch.
- Keep the requirements narrow and implementation-traceable; do not infer behavior from external APIs that the code does not enforce.

## Non-Goals

- Changing provider behavior, acceptance coverage, or generated docs.
- Reworking the entity schemas beyond what is already implemented.
- Inventing server-side workflow semantics that are not visible in provider code.

## Decisions

| Topic | Decision |
|-------|----------|
| Change scope | Use one OpenSpec change for both new entities because they were introduced together on the same branch and share version-gating and identity concepts. |
| Capability ids | Use `kibana-agentbuilder-workflow` for the resource and `kibana-agentbuilder-export-workflow` for the data source. |
| Spec layout | For each capability, author a full new spec file within the change directory: title, implementation path, purpose, schema sketch, and `## ADDED Requirements`. |
| Traceability source | Derive requirements from `resource.go`, `schema.go`, CRUD/read files, model mapping files, the shared workflow client wrapper, and acceptance fixtures. |
| Resource identity | Document the resource `id` as a computed composite `<space_id>/<workflow_id>`, with import passing through the supplied string and later operations requiring the composite form. |
| Data source identity | Document the data source input `id` as accepting either a bare workflow id or a composite id; state is normalized to the composite form after read. |
| YAML semantics | Document both validation of YAML syntax and semantic equality behavior for YAML formatting/key-order differences because both are enforced by the custom type used in schema. |

## Risks / Trade-offs

- The generated Markdown docs include user-facing wording and examples, but the OpenSpec requirements should follow the implementation rather than generated prose when they differ.
- The acceptance tests verify enabled-workflows environment setup, but that setup is test harness behavior rather than provider runtime behavior, so it is intentionally excluded from the requirements.
- Because this change is documenting new capabilities through delta specs, a later sync step is still required before `openspec/specs/` becomes canonical for these entities.

## Migration / State

- No provider state migration is involved.
- The resource and data source both normalize persisted state to composite ids, so the specs should make that canonical form explicit.

## Open Questions

- None. The branch implementation provides enough detail to document current behavior without additional API research.
