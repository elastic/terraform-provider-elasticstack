## Why

`elasticstack_kibana_default_data_view`, `elasticstack_kibana_security_enable_rule`, and `elasticstack_kibana_install_prebuilt_rules` are small standalone Kibana resources that still own CRUD orchestration directly via `entitycore.ResourceBase`. They are good candidates for a low-risk batch that exercises full envelope migration across a few different lifecycle shapes:

- a small standard CRUD resource (`default_data_view`)
- a small toggle-style resource (`security_enable_rule`)
- a write-oriented resource with wrapper-level `ModifyPlan` and a no-op delete (`install_prebuilt_rules`)

Migrating these together should produce quick wins, reduce boilerplate, and validate that wrapper-level interfaces like `ModifyPlan` continue to compose cleanly with the Kibana envelope.

## What Changes

- Migrate the following resources from `entitycore.ResourceBase` to `entitycore.NewKibanaResource` using real create/read/update/delete callbacks:
  - `elasticstack_kibana_default_data_view`
  - `elasticstack_kibana_security_enable_rule`
  - `elasticstack_kibana_install_prebuilt_rules`
- Preserve wrapper-level behavior that sits outside CRUD orchestration, especially `ModifyPlan` for `install_prebuilt_rules`.
- Keep state identity, import behavior, schema shape, and normalization behavior unchanged.
- For `install_prebuilt_rules`, preserve the current semantics where delete performs no remote action and only removes the Terraform state entry.
- If a small entitycore improvement emerges that benefits more than one resource in this batch, include it; otherwise keep changes resource-local.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-default-data-view`
- `kibana-security-enable-rule`
- `kibana-install-prebuilt-rules`

Implementation changes only; requirements-level behavior should remain unchanged.

## Impact

- `internal/kibana/defaultdataview/`
- `internal/kibana/security_enable_rule/`
- `internal/kibana/prebuilt_rules/`
- Potentially `internal/entitycore/` for small reusable envelope seams
- Targeted unit and acceptance tests for the affected resources
