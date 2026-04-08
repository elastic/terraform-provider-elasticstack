# Proposal: Document Kibana Agent Builder workflow entities

## Why

This branch adds two new Terraform entities for Kibana Agent Builder workflows:

- `elasticstack_kibana_agentbuilder_workflow`
- `elasticstack_kibana_agentbuilder_export_workflow`

The implementation, generated docs, and acceptance coverage exist, but there are no OpenSpec requirements yet describing their Terraform schema, identity rules, version gating, API usage, or YAML handling. That leaves the new behavior undocumented in the repository's canonical requirements system and makes future review, regression analysis, and sync-to-spec work harder.

## What Changes

- Add a new capability spec for `kibana-agentbuilder-workflow` covering the resource schema, CRUD behavior, composite identity, import behavior, stack-version gating, YAML validation and semantic equality, and invalid-workflow diagnostics.
- Add a new capability spec for `kibana-agentbuilder-export-workflow` covering the data source schema, accepted identifier forms, default and explicit space resolution, canonical state identity, stack-version gating, and read/not-found behavior.
- Record both capabilities as a single OpenSpec change so they can later be synced into `openspec/specs/` per repository workflow.

## Capabilities

### New Capabilities

- `kibana-agentbuilder-workflow`: Schema and functional requirements for the `elasticstack_kibana_agentbuilder_workflow` resource.
- `kibana-agentbuilder-export-workflow`: Schema and functional requirements for the `elasticstack_kibana_agentbuilder_export_workflow` data source.

### Modified Capabilities

- _(none)_

## Impact

- **Specs only:** this change documents behavior already present on the branch; it does not change provider runtime behavior.
- **Traceability:** the new requirements tie the implementation in `internal/kibana/agentbuilderworkflow` and `internal/kibana/exportagentbuilder/workflow` to stable capability ids for future maintenance.
- **Follow-up workflow:** after review, this change can be synced into canonical specs or archived according to the project's OpenSpec process.
