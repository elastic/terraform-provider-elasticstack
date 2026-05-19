## Context

The `elasticstack_kibana_data_view` resource manages Kibana data views. Kibana automatically populates `field_attrs` with `count` entries (field popularity statistics) whenever a data view is used in Discover. The current provider schema marks `field_attrs` with `mapplanmodifier.RequiresReplace()`, so any difference between state and Kibana's response — including server-generated `count` entries the user never wrote — is treated as a mutation requiring full resource replacement. This destroys and recreates the data view with a new internal ID, breaking Kibana dashboards and visualizations that reference the old ID.

Two structural problems must be resolved:

1. **Plan-time**: the plan modifier approach was evaluated and rejected by the maintainer (`@tobio`): "Plan modifiers can't mutate explicitly configured attributes." The correct mechanism is a **Custom Type implementing `MapSemanticEquals`**, which the Plugin Framework evaluates before planning and can operate on config-derived values directly.

2. **Write path**: the main `DataViewsUpdateDataViewRequestObjectInner` struct has no `FieldAttrs` field (confirmed at `generated/kbapi/kibana.gen.go:38188`). Field attribute updates must go through the dedicated `POST /api/data_views/data_view/{viewId}/fields` endpoint, wrapped as `kibanaoapi.UpdateFieldMetadata`.

The repo already demonstrates the custom type pattern at `internal/fleet/integration_policy/inputs_type.go` + `inputs_value.go`, making this approach well-precedented and reviewable.

## Goals / Non-Goals

**Goals:**
- Suppress phantom drift from server-generated `count`-only `field_attrs` entries at plan time.
- Allow in-place updates to `field_attrs` without replacing the resource.
- Route `field_attrs` writes through the correct Kibana API endpoint with space-aware path construction.
- Keep `count` as `Optional` (not `Computed`): count is only stored in state when the user explicitly configures it.

**Non-Goals:**
- Exposing `custom_description` from `DataViewsFieldattrs` — confirmed follow-up by maintainer.
- Making `count` `Computed` — maintainer direction: persist only when explicitly configured.
- Changes to `RequiresReplace` on other attributes (`allow_no_index`, `data_view.id`, `space_id`).
- Changes to the read-only data source variant of this resource.
- Handling clearing of deleted `field_attrs` entries via the API — the exact payload shape is left as an implementation detail.

## Decisions

### Decision 1: Custom `FieldAttrsType` / `FieldAttrsValue` with `MapSemanticEquals`

Implement `FieldAttrsType` and `FieldAttrsValue` following the exact pattern established in `internal/fleet/integration_policy/inputs_type.go` and `inputs_value.go`. The `FieldAttrsValue.MapSemanticEquals` method receives `v` (config-derived new value) and `priorValuable` (prior state value) and applies the following logic:

- If `v` is null (user wrote no `field_attrs`): semantically equal to a prior state that contains only `count`-only entries (entries where `custom_label` is null). If any prior entry has a `custom_label`, the user previously declared it and removing it is a real change.
- For each field in `v`: compare `custom_label`; suppress `count` differences when `count` is null in the new (config) value; return not-equal if a user-declared field is missing from prior state.
- For fields in prior state but absent from `v`: count-only entries (no `custom_label`) are server-generated and are suppressed; entries with a `custom_label` represent a user removing a managed entry and are a real change.

**Schema change**: replace `mapplanmodifier.RequiresReplace()` with `CustomType: NewFieldAttrsType(getFieldAttrElemType())` on the `field_attrs` attribute. No `PlanModifiers` needed.

**Why rejected alternatives were not chosen:**
- `planmodifier.Map`: rejected by maintainer; plan modifiers cannot mutate explicitly configured attributes.
- Read-time filtering (`models.go`): cannot distinguish "was managed, now removed" from "never managed" because `Read` cannot see the config, only prior state. Leads to one-cycle residual drift on removals and breaks on import.

### Decision 2: `UpdateFieldMetadata` wrapper with `SpaceAwarePathRequestEditor`

Add `kibanaoapi.UpdateFieldMetadata(ctx, client, spaceID, viewID, fields)` in `internal/clients/kibanaoapi/data_views.go` wrapping `UpdateFieldsMetadataDefaultWithResponse`. Pass `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as a request editor — the same pattern used in `data_views.go` (other endpoints) and `dashboards.go` — to ensure non-default space resources hit the correct URL prefix.

**Why `SpaceAwarePathRequestEditor`**: the maintainer (`@tobio`, 2026-05-14) explicitly directed: "This endpoint is space aware in the API documentation. Construct a space aware path in the client wrapper via `kibanautil.SpaceAwarePathRequestEditor`."

### Decision 3: Update flow in `update.go`

After calling `UpdateDataView`, compare `stateInner.FieldAttributes` and `planInner.FieldAttributes`. Build a payload of changed or removed fields and call `UpdateFieldMetadata`. Only include fields that changed or were removed — this minimises the API payload and respects the partial-update semantics of the endpoint.

### Decision 4: `innerModel.FieldAttributes` type change

Change the `FieldAttributes` field from `types.Map` to `FieldAttrsValue` in `models.go`. This ensures the Framework uses the custom type for all state read/write paths. Update `populateFromAPI` and `toAPICreateModel` constructors accordingly.

## Risks / Trade-offs

- **Schema change is non-breaking**: removing `RequiresReplace` is a relaxation. Existing state files are compatible; the attribute type wire format (map of objects) is unchanged.
- **`MapSemanticEquals` complexity**: unpacking nested `attr.Value` objects inside the equality check requires `types.Object.As` or manual attribute extraction, which is verbose. The precedent in `inputs_value.go` is a reliable guide.
- **Space routing for `UpdateFieldsMetadata`**: the generated client omits `spaceId` as a typed parameter (unlike other data view endpoints). Injecting the space via `SpaceAwarePathRequestEditor` covers this gap per maintainer direction, but empirical verification against a non-default Kibana space is advisable in acceptance testing.
- **API version gating**: `UpdateFieldsMetadata` may have a minimum Kibana version. The implementer should check whether an existing version-gate constant (analogous to `pre_8_8`) is needed. Failing to gate could break older deployments if the endpoint did not exist in earlier 8.x releases.

## Open Questions

- **`UpdateFieldsMetadata` space support**: The generated client's `UpdateFieldsMetadataDefaultWithResponse` takes only `viewId` (no `spaceId`), mapping to `POST /api/data_views/data_view/{viewId}/fields`. All other data view CRUD endpoints include `spaceId` and use the `/s/{spaceId}/api/...` prefix. Using `SpaceAwarePathRequestEditor` should route correctly per maintainer direction, but this should be verified empirically against a non-default-space data view during acceptance testing.
- **API version gating**: Is there a minimum Kibana version constraint for `UpdateFieldsMetadata`? The existing `pre_8_8` test tag path suggests version gates are used in this repo. This is an implementation detail to figure out during development, but failing to gate could break older deployments.
