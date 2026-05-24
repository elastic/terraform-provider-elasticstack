## Why

`elasticstack_kibana_synthetics_parameter` and `elasticstack_kibana_synthetics_private_location` are still implemented on `entitycore.ResourceBase` even though they are relatively small and live in the same Kibana Synthetics area. Migrating them together provides a coherent, low-risk batch that can establish envelope patterns for Synthetics resources before tackling the more complex `synthetics_monitor` resource.

## What Changes

- Migrate the following resources from `entitycore.ResourceBase` to `entitycore.NewKibanaResource` with real create/read/update/delete callbacks:
  - `elasticstack_kibana_synthetics_parameter`
  - `elasticstack_kibana_synthetics_private_location`
- Preserve existing import behavior, state IDs, schema shape, and read/write normalization.
- Keep wrapper-level interfaces such as `ImportState` unchanged where present.
- If a small entitycore improvement clearly benefits both resources, include it; otherwise keep changes resource-local.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-synthetics-parameter`
- `kibana-synthetics-private-location`

Implementation changes only; requirements-level behavior should remain unchanged.

## Impact

- `internal/kibana/synthetics/parameter/`
- `internal/kibana/synthetics/privatelocation/`
- Potentially `internal/entitycore/` for a small reusable envelope seam if one emerges
- Targeted unit and acceptance tests for the affected resources
