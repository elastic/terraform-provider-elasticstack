## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `npx openspec validate synthetics-parameter-space-id --type change` (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, **sync** delta into `openspec/specs/kibana-synthetics-parameter/spec.md` or **archive** the change per project workflow.

## 2. Schema and model

- [x] 2.1 Add `"space_id"` to `internal/kibana/synthetics/parameter/schema.go` using the canonical helper `kbschema.ResourceSpaceIDAttribute()` (`optional + computed`, `Default: stringdefault.StaticString(clients.DefaultSpaceID)`, `UseStateForUnknown`, `RequiresReplace`). Do **not** bump `schema.Schema.Version`.
- [x] 2.2 Add `SpaceID types.String \`tfsdk:"space_id"\`` to `Model` in `internal/kibana/synthetics/parameter/models.go`.
- [x] 2.3 Update `GetSpaceID()` on `Model` to return `m.SpaceID` directly (the schema default guarantees `"default"` when unconfigured).
- [x] 2.4 Update `GetResourceID()` on `Model` to parse `GetID()` as a composite (using `synthetics.TryReadCompositeID`); return the UUID segment, or fall back to the bare `id` for legacy state.
- [x] 2.5 Remove the `KibanaUnscopedSpace` implementation: delete `IsUnscopedSpace()` from `Model` and remove the `var _ entitycore.KibanaUnscopedSpace = Model{}` assertion. (Safe only because of the `"default"` schema default in 2.1.)
- [x] 2.6 Extend `modelFromOAPI` to accept a `spaceID string` argument and store both `SpaceID` and the composite `id` (`clients.CompositeID{ClusterID: spaceID, ResourceID: uuid}.String()`). Update all call sites.

## 3. CRUD operations

- [x] 3.1 **Create** (`create.go`): append `kibanautil.SpaceAwarePathRequestEditor(req.SpaceID)` to `PostParametersWithBodyWithResponse`. Store composite `id` in plan after create: `plan.ID = types.StringValue(clients.CompositeID{ClusterID: req.SpaceID, ResourceID: *createResponse.Id}.String())`.
- [x] 3.2 **Read** (`read.go`): append `kibanautil.SpaceAwarePathRequestEditor(spaceID)` to `GetParameterWithResponse` (the `spaceID` parameter already flows via `resolveKibanaResourceIdentity`). Update the `modelFromOAPI` call to pass `spaceID`.
- [x] 3.3 **Update** (`update.go`): append `kibanautil.SpaceAwarePathRequestEditor(req.SpaceID)` to `PutParameterWithBodyWithResponse`. Reassemble composite `id` after update: `plan.ID = types.StringValue(clients.CompositeID{ClusterID: req.SpaceID, ResourceID: req.WriteID}.String())`.
- [x] 3.4 **Delete** (`delete.go`): append `kibanautil.SpaceAwarePathRequestEditor(spaceID)` to both `DeleteParameterWithResponse` and `DeleteSyntheticsParamsWithResponse` calls (the `spaceID` parameter already flows via `resolveKibanaResourceIdentity`).

## 4. Import

- [x] 4.1 Replace `resource.ImportStatePassthroughID` in `resource.go` with a custom `ImportState` implementation built on `clients.ResolveCompositeSpaceAndID` (mirror `internal/kibana/tag/import.go`) that:
  - Accepts bare UUID (no `/`): set `space_id = "default"`, `id = "default/<uuid>"`.
  - Accepts composite `<space_id>/<uuid>`: set `space_id` to the space segment, `id` to the full composite.
  - Returns an error diagnostic if the UUID segment is empty.
- [x] 4.2 Update the existing `entitycore_contract_test.go::TestResource_importState_passthroughCompoundID` to reflect the new composite/space-splitting import behavior (it currently asserts verbatim passthrough).

## 5. Testing

- [ ] 5.1 Add or extend an acceptance test in `acc_test.go` with a step that creates a parameter in a named non-default space, verifies `space_id` and composite `id` in state, and verifies the Kibana API is called under the correct space path.
- [ ] 5.2 Add acceptance test steps for import by bare UUID and by composite `<space_id>/<uuid>` (including `ImportStateVerify`).
- [ ] 5.3 Add a default-space regression test: a parameter without `space_id` gets `space_id = "default"` and `id = "default/<uuid>"`, and routes to the unscoped path.
- [ ] 5.4 Add unit tests for `modelFromOAPI` with a `spaceID` argument to verify composite `id` assembly.
- [ ] 5.5 Add unit tests for the updated `GetResourceID()` / `GetSpaceID()` methods covering bare-UUID legacy state and composite state.

## 6. Docs

- [ ] 6.1 Update `internal/kibana/synthetics/parameter/resource-description.md` to document `space_id` and the composite `id` format, and include an HCL example with `space_id`.

## 7. Version gate (resolve open question first)

- [ ] 7.1 Determine whether the space-prefixed Parameters path needs a Kibana version floor above 8.12.0. If so, add a `GetVersionRequirements` check for non-default `space_id` mirroring `internal/kibana/synthetics/privatelocation` (per-attribute requirement + friendly error), and cover it with a version-gated test. If not, document that no additional gate is required.
