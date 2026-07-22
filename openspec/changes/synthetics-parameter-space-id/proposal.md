## Why

Practitioners cannot manage Synthetics parameters in non-default Kibana spaces using the `elasticstack_kibana_synthetics_parameter` resource today ([#4262](https://github.com/elastic/terraform-provider-elasticstack/issues/4262)). The Kibana Synthetics Parameters API supports space-scoped routing for all CRUD operations (`/s/{space_id}/api/synthetics/params`), but the provider resource does not expose a `space_id` attribute and always routes to the default space. Users who manage parameters in named spaces must fall back to the raw `http` provider, losing drift detection.

## What Changes

Add `space_id` to `elasticstack_kibana_synthetics_parameter` using the same composite-ID pattern already established by `elasticstack_kibana_synthetics_monitor`. The composite `id` is stored as `<space_id>/<parameter_uuid>` so import is self-contained and `resolveKibanaResourceIdentity` can recover space from state.

A `StateUpgraders` schema version bump (v0 → v1) migrates existing state from bare UUID `id` to `default/<uuid>`.

### Schema sketch (to merge into canonical `## Schema` on sync)

```hcl
resource "elasticstack_kibana_synthetics_parameter" "example" {
  key                 = "my_param"
  value               = "my_value"
  space_id            = "my-space"    # optional, computed; defaults to "default"; RequiresReplace
  share_across_spaces = false
}
```

The `id` field changes from a bare Kibana UUID to `<space_id>/<parameter_uuid>`.

### Composite-ID identity mechanics

- `GetID()` returns the raw `id` (composite or bare).
- `GetResourceID()` parses `id` as a composite; returns the UUID segment, or `id` itself for legacy bare values.
- `GetSpaceID()` returns `SpaceID` (defaulting to `""`).
- The model no longer implements `KibanaUnscopedSpace`; `IsUnscopedSpace()` is removed.
- `SpaceAwarePathRequestEditor(req.SpaceID)` is passed in all four CRUD calls to rewrite the API path.

### State migration

Schema version is bumped from **0** to **1**. The `StateUpgraders` v0→v1 function reads the existing `id` (bare UUID) and rewrites it as `default/<uuid>`, and writes `space_id = "default"` into state.

Import by bare UUID remains supported: `ImportState` accepts `<uuid>` (maps to default space) or `<space_id>/<uuid>`.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-synthetics-parameter`: Add `space_id` attribute, composite `id`, space-aware CRUD routing, `StateUpgraders` v0→v1 migration, and updated import behavior.

## Impact

- **Specs**: Delta under `openspec/changes/synthetics-parameter-space-id/specs/kibana-synthetics-parameter/spec.md`.
- **Implementation** (future): `internal/kibana/synthetics/parameter/` (schema, model, create, read, update, delete, state upgrade file), acceptance tests, and resource description docs.
