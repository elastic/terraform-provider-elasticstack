## Why

Kibana tracks field popularity by auto-writing `count` entries into `field_attrs` whenever a data view is actively used in Discover. Because `field_attrs` currently carries `mapplanmodifier.RequiresReplace()`, any difference between what Kibana returns and what is in Terraform state is treated as a mutation that forces full resource replacement. This means a data view that has been interacted with in the UI becomes permanently un-idempotent: every `terraform plan` after real usage proposes destroying and recreating the resource, risking Kibana dashboard associations tied to the data view's internal ID.

Two sub-problems must be fixed together:

1. **Plan-time drift suppression**: server-generated `count`-only entries (fields the user never declared) must not surface as Terraform drift. Count differences for user-declared fields must also be suppressed when `count` is null in config.
2. **Write path**: `field_attrs` changes must go through the dedicated `POST /api/data_views/data_view/{viewId}/fields` endpoint — the main data view update request body (`DataViewsUpdateDataViewRequestObjectInner`) has no `FieldAttrs` field.

## What Changes

- **Custom type** (`FieldAttrsType` / `FieldAttrsValue`): a new custom map type with `MapSemanticEquals` that suppresses drift from server-generated `count`-only entries and unsolicited `count` growth for user-declared fields.
- **Schema**: replace `mapplanmodifier.RequiresReplace()` on `field_attrs` with `CustomType: NewFieldAttrsType(...)`. The attribute remains `Optional` (not `Computed`).
- **API wrapper** (`kibanaoapi.UpdateFieldMetadata`): wrap `UpdateFieldsMetadataDefaultWithResponse` with space-aware path routing via `kibanautil.SpaceAwarePathRequestEditor`.
- **Update flow** (`update.go`): after calling `UpdateDataView`, diff state vs plan `field_attrs` and call `UpdateFieldMetadata` for any fields that changed or were removed.
- **Requirements update**: remove `field_attrs` from REQ-006 (replacement fields) and add new requirements covering semantic equality (REQ-015) and the field metadata write path (REQ-016).

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- `kibana-data-view`:
  - **REQ-006** (Lifecycle replacement fields): remove `data_view.field_attrs` from the list of attributes that require resource replacement.
  - **REQ-009** (Update request mapping): document that `field_attrs` changes are applied via a separate `UpdateFieldMetadata` API call (not the main data view update body) in the same Update transaction.
  - **REQ-015** *(new)*: `field_attrs` semantic equality — server-generated `count`-only entries SHALL be suppressed at plan time; user-declared entries are compared with `count` diffs ignored when `count` is null in config.
  - **REQ-016** *(new)*: `field_attrs` write path — on update, the provider SHALL call `UpdateFieldMetadata` with the delta of changed or removed fields; space-aware routing SHALL be used.

## Impact

| File | Change |
|------|--------|
| `internal/kibana/dataview/field_attrs_type.go` | **New** — `FieldAttrsType` implementing `basetypes.MapTypable` |
| `internal/kibana/dataview/field_attrs_value.go` | **New** — `FieldAttrsValue` implementing `basetypes.MapValuableWithSemanticEquals` |
| `internal/kibana/dataview/schema.go` | Swap `mapplanmodifier.RequiresReplace()` for `CustomType: NewFieldAttrsType(...)`; remove unused `mapplanmodifier` import |
| `internal/kibana/dataview/models.go` | Change `innerModel.FieldAttributes` from `types.Map` to `FieldAttrsValue`; update constructor calls in `populateFromAPI` and `toAPICreateModel` |
| `internal/kibana/dataview/update.go` | Add `UpdateFieldMetadata` call after `UpdateDataView`; diff state vs plan `field_attrs` |
| `internal/clients/kibanaoapi/data_views.go` | Add `UpdateFieldMetadata()` wrapper |
| `internal/kibana/dataview/acc_test.go` | Add/update acceptance test steps that exercise in-place `field_attrs` update (no replacement) |
| `openspec/specs/kibana-data-view/spec.md` | Update REQ-006, REQ-009; add REQ-015, REQ-016 |
