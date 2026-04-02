---
name: entity-requirements
description: Examines an existing Terraform resource or data source implementation and produces an OpenSpec requirements document under openspec/specs/. Use when the user asks to document requirements for a Terraform entity, capture behavior from code, or write a requirements doc for a resource/data source.
---

# Entity Requirements Documentation

Produce an **OpenSpec spec** (`spec.md`) for an existing Terraform resource or data source by examining its code path and capturing behavior. Follow [`dev-docs/high-level/openspec-requirements.md`](../../../dev-docs/high-level/openspec-requirements.md) and the OpenSpec shape: `## Purpose`, optional `## Schema`, `## Requirements` with `### Requirement:` and `#### Scenario:` blocks; use **SHALL** / **MUST** in requirement text.

## Input

- **Entity**: User specifies the Terraform entity (e.g. `elasticstack_elasticsearch_security_role`) or the implementation path (e.g. `internal/elasticsearch/security/role`).
- Resolve the implementation package: 
* For Plugin Framework based entities, search for `req.ProviderTypeName + "_..."`. 
* For SDK based entities, this is defined in `internal/provider/provider.go`. 

## Workflow

1. **Locate implementation**
   - **Resource**: Package under `internal/` containing `resource.Resource` and `Schema`, `Create`, `Read`, `Update`, `Delete` (and optionally `ImportState`, `UpgradeState`).
   - **Data source**: Package or file (e.g. `*_data_source.go`) containing a data source `Schema` and `Read`.

2. **Examine code path**  
   Use the checklist in [reference.md](reference.md) so nothing is missed:
   - Schema (attributes, blocks, required/optional/computed, plan modifiers, validators).
   - Metadata: type name, import, state upgrade.
   - CRUD: which APIs are called, how `id` is set, how errors and “not found” are handled.
   - Connection: default client vs resource-level override (e.g. `elasticsearch_connection`).
   - Compatibility: version checks and “Unsupported Feature” behavior.
   - Mapping: config/API/state (JSON parsing, empty vs null, preserve-unknown behavior).
   - Lifecycle: replacement vs in-place update (e.g. `RequiresReplace`).

3. **Write the spec**
   - **Path**: `openspec/specs/<capability>/spec.md` (e.g. `openspec/specs/elasticsearch-security-role/spec.md`). Use a stable capability id: `<backend>-<area>-<resource>` (see authoring guide).
   - **Title and implementation**: H1 title and a line `Resource implementation:` or `Data source implementation:` with the Go package path (as in legacy docs).
   - **Purpose**: Short `## Purpose` paragraph.
   - **Schema**: Optional `## Schema` with HCL-style block listing each attribute/block with `<required|optional|optional+computed|computed>`, types, and notes. Example reference: `openspec/specs/elasticsearch-security-role/spec.md`.
   - **Requirements**: `### Requirement: …` sections (group related behaviors; reference legacy REQ ids in titles like `(REQ-001–REQ-003)` when useful). Each requirement body MUST contain **SHALL** or **MUST**. Add `#### Scenario:` blocks (Given/When/Then) for verifiable behavior. Derive everything from the code; do not invent behavior. Categories: API, Identity, Import, Lifecycle, Connection, Compatibility, Create/Update, Read, Delete, Mapping, Plan/State, State, StateUpgrade. See [reference.md](reference.md).

4. **Quality**
   - Every requirement must be traceable to the implementation (file/function or logic).
   - Schema and requirements must be consistent (e.g. if schema has `description` optional with a version note, there must be a Compatibility requirement for that version).
   - For resources with state upgrade, include StateUpgrade requirements describing each version transition and error behavior.

## Output format

Use this OpenSpec-oriented structure:

```markdown
# `<name>` — Schema and Functional Requirements

Resource implementation: `<GO_PACKAGE_OR_DIR>`

## Purpose
...

## Schema
\`\`\`hcl
...
\`\`\`

## Requirements

### Requirement: Short name (REQ-xxx)

The resource SHALL ...

#### Scenario: ...
- GIVEN ...
- WHEN ...
- THEN ...
```

## Reference

- Authoring: `dev-docs/high-level/openspec-requirements.md`
- Example: `openspec/specs/elasticsearch-security-role/spec.md`
- Full code-path checklist and requirement categories: [reference.md](reference.md)
