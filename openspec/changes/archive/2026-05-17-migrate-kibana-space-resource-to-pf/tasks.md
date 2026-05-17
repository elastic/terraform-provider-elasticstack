## 1. Refactor Shared Types in `spaces/`

- [x] 1.1 In `spaces/models.go`, rename the unexported `model` struct to `SpaceModel` (exported) and update `dataSourceModel.Spaces` from `[]model` to `[]SpaceModel`
- [x] 1.2 In `spaces/schema.go`, extract a `spaceAttrs() map[string]schema.Attribute` helper containing the 8 common space attribute definitions; use it in both the data source schema and the new resource schema
- [x] 1.3 Add `fetchSpace(ctx, oapiClient, spaceID string) (*kbapi.SpaceResponse, bool, diag.Diagnostics)` to `spaces/read.go`; update `readDataSource` to call `kibanaoapi.ListSpaces()` (existing list path is unchanged)

## 2. Resource Model

- [x] 2.1 Add `resourceModel` struct to `spaces/models.go` embedding `entitycore.KibanaConnectionField` plus fields: `ID`, `SpaceID`, `Name`, `Description`, `DisabledFeatures`, `Initials`, `Color`, `ImageURL`, `Solution` (all `types.String` or `types.Set` as appropriate)
- [x] 2.2 Implement `GetID() types.String` returning `m.SpaceID` (space id is the resource identity, not composite)
- [x] 2.3 Implement `GetResourceID() types.String` returning `m.SpaceID`
- [x] 2.4 Implement `GetSpaceID() types.String` returning `types.StringValue("default")` (spaces API is unscoped)
- [x] 2.5 Implement `GetVersionRequirements()` returning the `>= 8.16.0` requirement only when `m.Solution` is non-null and non-empty

## 3. Resource Schema

- [x] 3.1 Create `spaces/resource_schema.go` (or add to `schema.go`) with `getResourceSchema()` returning a `resource.Schema` using `spaceAttrs()` for the 8 common fields plus:
  - `space_id`: Required, `stringplanmodifier.RequiresReplace()`, regex validator `[a-z0-9_-]+`
  - `id`: Computed, `stringplanmodifier.UseStateForUnknown()`
  - `initials`: Optional+Computed, `UseStateForUnknown`, max-length 2 validator
  - `color`: Optional+Computed, `UseStateForUnknown`
  - `image_url`: Optional, data-URI format validator
  - `solution`: Optional+Computed, enum validator (security, oblt, es, classic)
  - `disabled_features`: Optional+Computed, `SetAttribute` of `types.StringType`, `UseStateForUnknown`

## 4. CRUD Callbacks

- [x] 4.1 Create `spaces/create.go`: build `kbapi.PostSpacesSpaceJSONRequestBody` from plan (omit null Optional fields), call `kibanaoapi.CreateSpace()`, call `fetchSpace()` to populate state, set `ID = SpaceID`
- [x] 4.2 Add resource read callback to `spaces/read.go`: call `fetchSpace()`, populate `resourceModel` from response (do NOT set `ImageURL`), return `found=false` if space missing
- [x] 4.3 Create `spaces/update.go`: build `kbapi.PutSpacesSpaceIdJSONRequestBody` from plan (omit null Optional fields), call `kibanaoapi.UpdateSpace()`, call `fetchSpace()` to refresh state
- [x] 4.4 Create `spaces/delete.go`: call `kibanaoapi.DeleteSpace()` using `resourceID`

## 5. Resource Entry Point

- [x] 5.1 Create `spaces/resource.go` with `type Resource struct { *entitycore.KibanaResource[resourceModel] }` and `NewResource() resource.Resource` using `entitycore.NewKibanaResource[resourceModel](entitycore.ComponentKibana, "space", getResourceSchema, readSpaceResource, deleteSpace, createSpace, updateSpace)`
- [x] 5.2 Add `ImportState` method to `Resource` using `resource.ImportStatePassthroughID` (preserves existing import behavior)

## 6. Provider Wiring

- [x] 6.1 Register `spaces.NewResource` in `provider/plugin_framework.go` `resources()` function
- [x] 6.2 Remove `kibana.ResourceSpace()` from `provider/provider.go` `ResourcesMap`

## 7. Tests

- [x] 7.1 Move all test functions from `internal/kibana/space_test.go` into `internal/kibana/spaces/acc_test.go` (update package to `spaces_test`, update imports)
- [x] 7.2 Verify `TestAccResourceSpace_ClearEmptyFields` is preserved (tests the null-vs-empty-string behavior that `configuredString()` previously handled)
- [x] 7.3 Create `spaces/testdata/TestAccSpaceResourceFromSDK/create/main.tf` with a config creating a Kibana space
- [x] 7.4 Add `TestAccSpaceResourceFromSDK` acceptance test: step 1 uses `ExternalProviders` with `VersionConstraint: "0.15.1"`, step 2 uses `ProtoV6ProviderFactories`

## 8. Cleanup

- [x] 8.1 Delete `internal/kibana/space.go`
- [x] 8.2 Delete `internal/kibana/space_test.go`
- [x] 8.3 Update `openspec/specs/kibana-space/spec.md` implementation path from `internal/kibana/space.go` to `internal/kibana/spaces/resource.go`

## 9. Verification

- [x] 9.1 `make build` passes
- [x] 9.2 `go test ./internal/kibana/spaces/... -v -count=1` (unit tests) passes
- [x] 9.3 `go test ./internal/kibana/spaces/... -v -count=1 -run TestAcc` (acceptance tests) passes
