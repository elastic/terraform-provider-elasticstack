## Why

`elasticstack_kibana_data_view` is a good next-step Kibana migration target, but it is materially broader than `default_data_view`. It has custom import-state initialization, reconciliation behavior during create, namespace updates during update, and field metadata delta handling. Those behaviors still appear compatible with full envelope migration, but they deserve a focused change rather than being hidden inside a larger batch.

Keeping `data_view` standalone reduces review noise, makes it easier to spot envelope friction early, and lets the change document any small reusable entitycore improvement that might help later medium-complexity Kibana resources.

## What Changes

- Migrate `elasticstack_kibana_data_view` from `entitycore.ResourceBase` to `entitycore.NewKibanaResource` with real create/read/update/delete callbacks.
- Preserve strict Terraform-visible behavior:
  - same import ID format and import-state initialization behavior
  - same schema shape
  - same state ID and space resolution behavior
  - same namespace reconciliation semantics
  - same field metadata delta and refresh behavior
- Keep wrapper-level `ImportState` behavior unchanged.
- If a small entitycore improvement emerges that cleanly supports this resource and future Kibana migrations, include it.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-data-view`

Implementation changes only; requirements-level behavior should remain unchanged.

## Impact

- `internal/kibana/dataview/`
- Potentially `internal/entitycore/` for a small reusable envelope seam discovered during migration
- Data view unit tests and acceptance tests
