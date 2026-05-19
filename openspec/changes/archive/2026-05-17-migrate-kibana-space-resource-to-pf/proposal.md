## Why

The `elasticstack_kibana_space` resource is the last remaining SDK-based entity in the spaces domain — the `elasticstack_kibana_spaces` data source already uses the Plugin Framework. Completing the migration removes the final SDK dependency from this package and aligns the implementation with the entitycore pattern used everywhere else in the Kibana domain.

## What Changes

- Add `internal/kibana/spaces/resource.go`, `create.go`, `update.go`, `delete.go` implementing the resource via `entitycore.NewKibanaResource`
- Promote the existing `model` struct in `spaces/models.go` to a shared `SpaceModel` used by both the resource and the list data source
- Extract shared `spaceAttrs()` schema helper in `spaces/schema.go` to avoid duplicating attribute definitions
- Add `fetchSpace()` internal helper in `spaces/read.go` shared by the new resource read callback and any future read reuse
- Wire `spaces.NewResource` in `provider/plugin_framework.go`
- Remove `kibana.ResourceSpace()` from `provider/provider.go`
- Delete `internal/kibana/space.go`
- Move acceptance tests from `internal/kibana/space_test.go` into `internal/kibana/spaces/acc_test.go`
- Add SDK upgrade test (`TestAccSpaceResourceFromSDK`) with `VersionConstraint: "0.15.1"`
- Update the implementation path reference in `openspec/specs/kibana-space/spec.md`

## Capabilities

### New Capabilities

None. The resource schema and behavior are unchanged.

### Modified Capabilities

- `kibana-space`: Implementation path changes from `internal/kibana/space.go` to `internal/kibana/spaces/resource.go`. No schema or behavioral requirement changes.

## Impact

- `internal/kibana/spaces/` — new resource files added; `models.go` and `schema.go` refactored to share types with the data source
- `internal/kibana/space.go` — deleted
- `provider/provider.go` — one entry removed from `ResourcesMap`
- `provider/plugin_framework.go` — one entry added to `resources()`
- `openspec/specs/kibana-space/spec.md` — implementation path updated
