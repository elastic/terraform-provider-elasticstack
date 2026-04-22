## 1. Change Artifacts

- [x] 1.1 Create the OpenSpec change scaffold and author `proposal.md`, `design.md`, and `tasks.md` for the two new Agent Builder workflow entities.
- [x] 1.2 Write a new capability spec for `kibana-agentbuilder-workflow` that captures schema, version gating, identity/import, create/update/read/delete, YAML handling, and invalid-workflow diagnostics.
- [x] 1.3 Write a new capability spec for `kibana-agentbuilder-export-workflow` that captures schema, accepted id forms, default and explicit space handling, canonical state mapping, and read/not-found behavior.

## 2. Validation

- [x] 2.1 Run `npx openspec validate kibana-agentbuilder-workflows --type change` (or an equivalent local CLI invocation) and fix any structural issues in the change artifacts.
- [x] 2.2 Sync the new capabilities into `openspec/specs/kibana-agentbuilder-workflow/spec.md` and `openspec/specs/kibana-agentbuilder-export-workflow/spec.md`, or archive the change, when the repository workflow reaches that stage.
