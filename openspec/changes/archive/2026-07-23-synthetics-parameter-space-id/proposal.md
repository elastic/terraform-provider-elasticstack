## Why

Practitioners cannot manage Synthetics parameters in non-default Kibana spaces using the `elasticstack_kibana_synthetics_parameter` resource today ([#4262](https://github.com/elastic/terraform-provider-elasticstack/issues/4262)). The Kibana Synthetics Parameters API supports space-scoped routing for all CRUD operations (`/s/{space_id}/api/synthetics/params`), but the provider resource does not expose a `space_id` attribute and always routes to the default space. Users who manage parameters in named spaces must fall back to the raw `http` provider, losing drift detection.

## What Changes

Add `space_id` to `elasticstack_kibana_synthetics_parameter` using the provider's **canonical space-aware pattern**: the shared `kbschema.ResourceSpaceIDAttribute()` schema helper (`optional + computed`, default `"default"`, `UseStateForUnknown` + `RequiresReplace`), an explicit `"default"` space value, and a composite `id` of `<space_id>/<parameter_uuid>`. This is the pattern used by ~27 Kibana resources (tag, connectors, dataview, slo, alerting rule, and others) via `internal/kibana/kbschema/space_id.go`, and it is preferred over the older `KibanaUnscopedSpace` opt-out (used by only the synthetics monitor/private-location trio and a few Fleet resources).

The composite `id` is stored as `<space_id>/<parameter_uuid>` so import is self-contained and `resolveKibanaResourceIdentity` can recover the space from state.

### Schema sketch (to merge into canonical `## Schema` on sync)

```hcl
resource "elasticstack_kibana_synthetics_parameter" "example" {
  key                 = "my_param"
  value               = "my_value"
  space_id            = "my-space"    # optional, computed; defaults to "default"; RequiresReplace
  share_across_spaces = false
}
```

The `space_id` attribute is defined via `kbschema.ResourceSpaceIDAttribute()`. The `id` field changes from a bare Kibana UUID to `<space_id>/<parameter_uuid>`.

### Composite-ID identity mechanics

- `GetID()` returns the raw `id` (composite or bare legacy value).
- `GetResourceID()` parses `id` as a composite (`synthetics.TryReadCompositeID`); returns the UUID segment, or `id` itself for legacy bare values.
- `GetSpaceID()` returns `m.SpaceID` directly (the schema default guarantees it is `"default"` when not configured).
- The model **no longer implements** `KibanaUnscopedSpace`; `IsUnscopedSpace()` and the `var _ entitycore.KibanaUnscopedSpace = Model{}` assertion are removed. The envelope's normal non-empty `space_id` validation applies, and it is satisfied because the schema default materializes `"default"` before create/update.
- `SpaceAwarePathRequestEditor(spaceID)` is passed in all four CRUD calls to rewrite the API path. `SpaceAwarePathRequestEditor`/`BuildSpaceAwarePath` leave the path unchanged when the space is `"default"` or empty.

### Version requirements

Non-default-space Parameters routing does **not** need a Kibana version floor above the resource’s existing **8.12.0** gate. Kibana v8.12.0 documents both unscoped and `/s/<space_id>/api/synthetics/params` paths in `docs/api/synthetics/params/add-param.asciidoc`; the public API commit `8bbb58f19aadb34f6a94bf9d77b16bc61a73091c` is included in v8.12.0. No `GetVersionRequirements` check is added (unlike `synthetics/privatelocation`, which requires 9.4.0 for non-default space).

### No state migration required

Existing default-space parameters store a bare-UUID `id` (legacy). No `StateUpgraders` / schema version bump is needed: `resolveKibanaResourceIdentity` parses `id` via `CompositeIDFromStr`, and a bare UUID (no `/`) falls back to `GetResourceID()` + `GetSpaceID()`. A legacy parameter is by definition a default-space parameter, and an empty/`"default"` space routes to the unscoped path — so the fallback is correct, not lossy. The bare-UUID `id` is rewritten to the composite form naturally on the next create/update or refresh; no destructive action occurs.

This matches how every other Kibana resource added `space_id` — none shipped a schema-version bump solely to introduce the attribute.

### Import

`ImportState` is replaced (the resource currently uses `ImportStatePassthroughID`) with a small custom handler built on `clients.ResolveCompositeSpaceAndID`, matching the tag resource's `import.go`:

- Bare `<uuid>` (no `/`): maps to the default space; sets `space_id = "default"` and `id = "default/<uuid>"`.
- Composite `<space_id>/<uuid>`: sets `space_id` to the space segment and `id` to the full composite.
- An empty UUID segment (`<space_id>/`) returns an error diagnostic.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-synthetics-parameter`: Add `space_id` attribute (canonical `kbschema` helper, default `"default"`), composite `id`, space-aware CRUD routing, and updated import behavior. Remove the `KibanaUnscopedSpace` opt-out.

## Impact

- **Specs**: Delta under `openspec/changes/synthetics-parameter-space-id/specs/kibana-synthetics-parameter/spec.md`.
- **Implementation** (future): `internal/kibana/synthetics/parameter/` (schema, model, create, read, update, delete, import), acceptance tests, and resource description docs. No state-upgrade file.

## Open Questions

- **`share_across_spaces = true` with a non-default `space_id`.** These are semantically in tension (share-across-all vs. scope-to-one). Should the provider reject the combination with a validation error, or document that `share_across_spaces` wins? Current scope leaves `share_across_spaces` semantics unchanged.
