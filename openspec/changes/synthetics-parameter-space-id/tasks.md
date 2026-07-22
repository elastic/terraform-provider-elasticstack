## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `npx openspec validate synthetics-parameter-space-id --type change` (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, **sync** delta into `openspec/specs/kibana-synthetics-parameter/spec.md` or **archive** the change per project workflow.

## 2. Schema and model

- [ ] 2.1 Add `space_id` to `internal/kibana/synthetics/parameter/schema.go`: `schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}}`. Bump `schema.Schema.Version` to **1**.
- [ ] 2.2 Add `SpaceID types.String \`tfsdk:"space_id"\`` to `Model` in `internal/kibana/synthetics/parameter/models.go`.
- [ ] 2.3 Update `GetSpaceID()` on `Model` to return `m.SpaceID` (not `types.StringValue("")`).
- [ ] 2.4 Update `GetResourceID()` on `Model` to parse `GetID()` as composite (using `synthetics.TryReadCompositeID`); return the UUID segment, or fall back to bare `id` for legacy state.
- [ ] 2.5 Remove `KibanaUnscopedSpace` implementation: delete `IsUnscopedSpace()` from `Model` and remove the `var _ entitycore.KibanaUnscopedSpace = Model{}` assertion.
- [ ] 2.6 Extend `modelFromOAPI` to accept a `spaceID string` argument and store both `SpaceID` and the composite `id` (`<space_id>/<uuid>`). Update all call sites.

## 3. CRUD operations

- [ ] 3.1 **Create** (`create.go`): append `kibanautil.SpaceAwarePathRequestEditor(req.SpaceID)` to `PostParametersWithBodyWithResponse`. Store composite `id` in plan after create: `plan.ID = types.StringValue(clients.CompositeID{ClusterID: req.SpaceID, ResourceID: *createResponse.Id}.String())`.
- [ ] 3.2 **Read** (`read.go`): append `kibanautil.SpaceAwarePathRequestEditor(spaceID)` to `GetParameterWithResponse` (the `spaceID` parameter already flows via `resolveKibanaResourceIdentity`). Update `modelFromOAPI` call to pass `spaceID`.
- [ ] 3.3 **Update** (`update.go`): append `kibanautil.SpaceAwarePathRequestEditor(req.SpaceID)` to `PutParameterWithBodyWithResponse`. Reassemble composite `id` after update: `plan.ID = types.StringValue(clients.CompositeID{ClusterID: req.SpaceID, ResourceID: req.WriteID}.String())`.
- [ ] 3.4 **Delete** (`delete.go`): append `kibanautil.SpaceAwarePathRequestEditor(spaceID)` to both `DeleteParameterWithResponse` and `DeleteSyntheticsParamsWithResponse` calls (the `spaceID` parameter already flows via `resolveKibanaResourceIdentity`).

## 4. Import

- [ ] 4.1 Replace `resource.ImportStatePassthroughID` in `resource.go` with a custom `ImportState` implementation that:
  - Accepts bare UUID (no `/`): set `id = "default/<uuid>"`, `space_id = "default"`.
  - Accepts composite `<space_id>/<uuid>`: set `id` to the full composite, `space_id` to the space segment.
  - Returns an error if the UUID segment is empty.

## 5. State migration

- [ ] 5.1 Create `internal/kibana/synthetics/parameter/state_upgrade.go` with an `UpgradeState` method on `*Resource` that returns `map[int64]resource.StateUpgrader{0: {StateUpgrader: migrateV0toV1}}`.
- [ ] 5.2 Implement `migrateV0toV1` using `stateutil.SetDefaultState`, `stateutil.UnmarshalStateMap`, and `stateutil.MarshalStateMap`. The function SHALL: read `id` from the state map; if no `/` is present, rewrite `id` to `"default/<id>"` and add `"space_id": "default"`; if already composite, add `"space_id"` equal to the space segment.
- [ ] 5.3 Add `_ resource.ResourceWithUpgradeState = newResource()` to the interface assertions in `resource.go`.

## 6. Testing

- [ ] 6.1 Add or extend acceptance test in `acc_test.go` with a step that creates a parameter in a named non-default space, verifies `space_id` in state, and verifies that the Kibana API is called under the correct space path.
- [ ] 6.2 Add acceptance test steps for import by bare UUID and by composite `<space_id>/<uuid>`.
- [ ] 6.3 Add unit tests for `migrateV0toV1`: bare UUID input → composite rewrite; already-composite input → no-op (or space_id populated).
- [ ] 6.4 Add unit tests for `modelFromOAPI` with a `spaceID` argument to verify composite `id` assembly.
- [ ] 6.5 Add unit tests for the updated `GetResourceID()` / `GetSpaceID()` methods covering bare-UUID legacy state and composite state.

## 7. Docs

- [ ] 7.1 Update `internal/kibana/synthetics/parameter/resource-description.md` to document `space_id` and the composite `id` format, note the state migration, and include an HCL example with `space_id`.
