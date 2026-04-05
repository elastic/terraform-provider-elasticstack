## Why

The Synthetics private location `space_id` attribute depends on Kibana API behavior that exists only from Elastic Stack **9.4.0-SNAPSHOT** onward. Without an explicit version gate, practitioners on older stacks get confusing API failures; acceptance tests that assume `space_id` in a non-default space also fail on unsupported versions.

## What Changes

- Document in schema and resource behavior that using a **non-default** Kibana space via `space_id` (including composite import identifiers that resolve to a non-default space) requires Elastic Stack **9.4.0-SNAPSHOT** or later.
- Enforce that minimum version at plan/apply time when the effective space is not the default space (same pattern as other version-gated resources, e.g. `kibana_stream`, Fleet `space_ids`).
- Skip `TestSyntheticPrivateLocationResource_nonDefaultSpace` when the test stack is older than **9.4.0-SNAPSHOT** (replacing or tightening the current skip that only reflects Fleet `space_ids` 9.1+).

## Capabilities

### New Capabilities

<!-- None — behavior is a constraint on existing `space_id` support. -->

### Modified Capabilities

- `kibana-synthetics-private-location`: Add a **minimum Elastic Stack version** for non-default `space_id` and composite import to a non-default space; require clear diagnostics when the version is too low; align acceptance coverage with supported stacks.

## Impact

- **Code**: `internal/kibana/synthetics/privatelocation/` (create, read, delete, schema, embedded `space_id` description), `internal/kibana/synthetics/privatelocation/acc_test.go`.
- **Specs**: Delta under `openspec/changes/.../specs/kibana-synthetics-private-location/`; later merge to `openspec/specs/kibana-synthetics-private-location/spec.md` when archived.
- **Dependencies**: Existing `clients.APIClient.EnforceMinVersion` and `hashicorp/go-version` (already used elsewhere).
