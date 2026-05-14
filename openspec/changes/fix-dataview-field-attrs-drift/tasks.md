## 1. Custom `FieldAttrsType` and `FieldAttrsValue`

- [x] 1.1 Create `internal/kibana/dataview/field_attrs_type.go` implementing `FieldAttrsType` as a `basetypes.MapTypable`, following the pattern in `internal/fleet/integration_policy/inputs_type.go`. Include `String()`, `ValueType()`, `Equal()`, `ValueFromMap()`, `ValueFromTerraform()`, and `NewFieldAttrsType(elemType attr.Type)` constructor.
- [x] 1.2 Create `internal/kibana/dataview/field_attrs_value.go` implementing `FieldAttrsValue` as `basetypes.MapValuableWithSemanticEquals`, following the pattern in `internal/fleet/integration_policy/inputs_value.go`. Include:
  - `Type()`, `Equal()` methods
  - `MapSemanticEquals(ctx, priorValuable)` with the following logic:
    - If new value is null: equal iff all prior entries are count-only (no `custom_label`)
    - For each field in new value: compare `custom_label`; suppress `count` difference when `count` is null in new value; not-equal if a user-declared entry is absent from prior state
    - For fields in prior but absent from new: count-only entries are suppressed; entries with `custom_label` are a real change (not-equal)
  - `NewFieldAttrsNull`, `NewFieldAttrsUnknown`, `NewFieldAttrsValue`, `NewFieldAttrsValueFrom` constructors

## 2. Schema Change

- [x] 2.1 In `internal/kibana/dataview/schema.go`, replace `mapplanmodifier.RequiresReplace()` on `field_attrs` with `CustomType: NewFieldAttrsType(getFieldAttrElemType())`. Remove the `PlanModifiers` slice from the attribute.
- [x] 2.2 Remove the `mapplanmodifier` import from `schema.go` if it is no longer used elsewhere in the file.
- [x] 2.3 Run `make build` to verify the schema change compiles without error.

## 3. Model Update

- [x] 3.1 In `internal/kibana/dataview/models.go`, change the `FieldAttributes` field in `innerModel` from `types.Map` to `FieldAttrsValue`.
- [x] 3.2 Update the `populateFromAPI` call that builds `FieldAttributes` to produce a `FieldAttrsValue` instead of a `types.Map`. Use the `NewFieldAttrsValueFrom` (or equivalent) constructor.
- [x] 3.3 Update `toAPICreateModel` to extract the underlying `map[string]attr.Value` from `FieldAttrsValue` when building the create API request body.
- [x] 3.4 Verify `toAPIUpdateModel` does not need changes (it already omits `FieldAttrs`).
- [x] 3.5 Run `make build` to verify model changes compile without error.

## 4. API Wrapper

- [x] 4.1 In `internal/clients/kibanaoapi/data_views.go`, add `UpdateFieldMetadata(ctx context.Context, client *Client, spaceID string, viewID string, fields map[string]interface{}) diag.Diagnostics`. The implementation SHALL:
  - Call `client.API.UpdateFieldsMetadataDefaultWithResponse(ctx, viewID, kbapi.UpdateFieldsMetadataDefaultJSONRequestBody{Fields: fields}, kibanautil.SpaceAwarePathRequestEditor(spaceID))`
  - Treat HTTP 200 as success
  - Return error diagnostics for transport errors and unexpected HTTP statuses

## 5. Update Flow

- [x] 5.1 In `internal/kibana/dataview/update.go`, after the `UpdateDataViewNamespaces` call (and only when there are no existing errors), compare `stateInner.FieldAttributes` and `planInner.FieldAttributes`:
  - Build a `map[string]interface{}` delta of all fields present in the plan that differ from state, plus all fields removed from plan (using an empty object `{}` or the appropriate clearing payload).
  - If the delta is non-empty, call `kibanaoapi.UpdateFieldMetadata(ctx, oapiClient, spaceID, viewID, delta)` and append any diagnostics.
- [x] 5.2 Ensure the `populateFromAPI` call at the end of `Update` re-reads the data view state after both `UpdateDataView` and `UpdateFieldMetadata` have run so that final state reflects all changes.

## 6. Acceptance Tests

- [x] 6.1 Add or update a test step in `internal/kibana/dataview/acc_test.go` that:
  - Creates a data view with no `field_attrs` in config
  - Simulates (or asserts via `terraform plan`) that server-generated `count` entries do not cause a diff
- [x] 6.2 Add a test step that exercises in-place update of `field_attrs` (add a `custom_label` entry, verify no replacement, verify the label is written via `UpdateFieldMetadata`)
- [x] 6.3 Add a test step that removes a `field_attrs` entry previously set in config and verifies the resource is updated in place (not replaced)

## 7. Requirements Update

- [x] 7.1 Update `openspec/specs/kibana-data-view/spec.md`:
  - Remove `data_view.field_attrs` from the replacement list in REQ-006
  - Update REQ-009 to state that `field_attrs` changes are applied via a separate `UpdateFieldMetadata` call
  - Add REQ-015 (field_attrs semantic equality) as defined in the delta spec
  - Add REQ-016 (field_attrs write path) as defined in the delta spec

## 8. Validation

- [x] 8.1 Run `make build` to ensure the provider compiles.
- [x] 8.2 Run `make check-lint` to ensure lint passes.
- [x] 8.3 Run unit tests: `go test ./internal/kibana/dataview/...`
- [x] 8.4 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fix-dataview-field-attrs-drift --type change`
- [ ] 8.5 If a Kibana stack is available, run acceptance tests: `TF_ACC=1 go test -v -run TestAccResourceDataView ./internal/kibana/dataview/... -timeout 20m`
