## Why

The `elasticstack_kibana_action_connector` data source is already on entitycore, but the resource still owns CRUD orchestration directly via `entitycore.ResourceBase`. Migrating the resource is a useful medium-complexity step because it combines a familiar entity shape with a few non-trivial concerns: state upgrade handling, connector-specific model conversion, and version-gated support for preconfigured connector IDs.

Keeping it as a standalone change makes it easier to preserve strict behavior while validating envelope patterns that are slightly richer than the smallest resource migrations.

## What Changes

- Migrate `elasticstack_kibana_action_connector` from `entitycore.ResourceBase` to `entitycore.NewKibanaResource` with real create/read/update/delete callbacks.
- Preserve strict Terraform-visible behavior:
  - same schema shape
  - same import behavior
  - same composite ID handling
  - same state upgrade behavior
  - same version-gated validation for preconfigured connector IDs
  - same read-after-write refresh behavior
- Keep wrapper-level `ImportState` and `UpgradeState` behavior unchanged.
- If a small reusable entitycore improvement is warranted and clearly benefits this resource plus later migrations, include it.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-action-connector`

Implementation changes only; requirements-level behavior should remain unchanged.

## Impact

- `internal/kibana/connectors/`
- Potentially `internal/entitycore/` for a small reusable envelope seam discovered during migration
- Connector unit tests and acceptance tests
