---
name: entity-requirements
description: Examines an existing Terraform resource or data source implementation and produces a comprehensive requirements document following dev-docs/reqs/template.md. Use when the user asks to document requirements for a Terraform entity, capture behavior from code, or write a requirements doc for a resource/data source.
---

# Entity Requirements Documentation

Produce a **requirements document** for an existing Terraform resource or data source by examining its code path and capturing behavior. Output follows the repo template at `dev-docs/reqs/template.md`.

## Input

- **Entity**: User specifies the Terraform entity (e.g. `elasticstack_elasticsearch_security_role`) or the implementation path (e.g. `internal/elasticsearch/security/role`).
- Resolve the implementation package: search for `req.ProviderTypeName + "_"` or `TypeName` in `internal/` to find the package that registers that type name, or use the given directory.

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

3. **Write the document**
   - **Path**: `dev-docs/reqs/<domain>/<name>.md` (e.g. `dev-docs/reqs/elasticsearch/security/role.md`). Match existing layout under `dev-docs/reqs/`.
   - **Title and implementation**: First line and “Resource implementation” (or “Data source implementation”) per template.
   - **Schema**: HCL-style block listing each attribute/block with `<required|optional|optional+computed|computed>`, types, and short notes (e.g. version requirements, deprecated blocks). Follow the example in the template and in `dev-docs/reqs/elasticsearch/security/role.md`.
   - **Requirements**: Numbered list `**[REQ-NNN] (Category)**: The resource/data source shall ...**`. Derive every requirement from the code; do not invent behavior. Use categories such as API, Identity, Import, Lifecycle, Connection, Compatibility, Create/Update, Read, Delete, Mapping, Plan/State, State, StateUpgrade. See [reference.md](reference.md) for categories and examples.

4. **Quality**
   - Every requirement must be traceable to the implementation (file/function or logic).
   - Schema and requirements must be consistent (e.g. if schema has `description` optional with a version note, there must be a Compatibility requirement for that version).
   - For resources with state upgrade, include StateUpgrade requirements describing each version transition and error behavior.

## Output format

Use the structure from `dev-docs/reqs/template.md`:

```markdown
# `<RESOURCE_OR_DATA_SOURCE_NAME>` — Schema and Functional Requirements

Resource implementation: `<GO_PACKAGE_OR_DIR>`
# or: Data source implementation: ...

## Schema

```hcl
resource "<PROVIDER_TYPE>" "example" {
  # attributes/blocks with required|optional|computed and notes
}
```

## Requirements

- **[REQ-001] (API)**: ...
- **[REQ-002] (Identity)**: ...
```

## Reference

- Template: `dev-docs/reqs/template.md`
- Example: `dev-docs/reqs/elasticsearch/security/role.md`
- Full code-path checklist and requirement categories: [reference.md](reference.md)
