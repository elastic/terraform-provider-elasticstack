## Context

`internal/kibana/space.go` (335 LOC) implements the `elasticstack_kibana_space` resource using `terraform-plugin-sdk/v2`. The `internal/kibana/spaces/` package already contains a PF data source for listing spaces, including a `model` struct representing a single space and client calls into `internal/clients/kibanaoapi/spaces.go`. The resource can be added to this package with targeted refactoring to promote shared types.

The key SDK-specific workaround in `space.go` is `configuredString()` — a helper that distinguishes "field not set" from "field explicitly set to empty string" by inspecting raw config map values. In PF, `types.String` natively has null (unset) vs known-empty-string (`""`) as distinct states, so `configuredString()` is unnecessary and disappears entirely.

## Goals / Non-Goals

**Goals:**
- Migrate the resource to PF using `entitycore.NewKibanaResource[resourceModel]`
- Consolidate resource and data source in `internal/kibana/spaces/` with shared `SpaceModel` and schema helpers
- Preserve all existing behaviors exactly: `image_url` not populated on read, `solution` version gate, `space_id` ForceNew, `disabled_features` as a set, computed defaults for `initials` and `color`
- Add SDK upgrade test to verify state compatibility from `v0.15.1`

**Non-Goals:**
- Schema changes of any kind
- Migrating the data source (already on PF)

## Decisions

### Package consolidation into `spaces/`

The existing `spaces/models.go` already defines a `model` struct with the 8 space fields. Promoting this to a shared `SpaceModel` lets both the resource model and the list data source element type use the same struct — reducing duplication in field definitions and read-mapping logic. The refactor is additive: the existing data source's `dataSourceModel.Spaces []model` field becomes `[]SpaceModel`.

### `GetSpaceID()` returns a fixed `"default"`

The Kibana Spaces management API (`/api/spaces/space/{id}`) is not scoped to a Kibana space — it runs at the provider level. The entitycore envelope calls `GetSpaceID()` to build a scoped client, but the space context has no effect on these endpoints. Returning `types.StringValue("default")` is correct and harmless. The attribute named `space_id` in the Terraform schema is the **identifier of the space being managed**, which maps to `GetResourceID()`.

### `GetVersionRequirements()` for the `solution` gate

The `solution` attribute requires Kibana ≥ 8.16.0. Rather than a runtime check inside the create/update callbacks, the resource model implements `entitycore.WithVersionRequirements`: `GetVersionRequirements()` inspects the model's `Solution` field and returns the `8.16.0` requirement only when it is non-null and non-empty. The entitycore envelope enforces this before calling the create or update callback, matching the existing behavior exactly.

### `image_url` omitted from read

The Kibana Get Space API does not return `image_url`. The resource preserves this by simply not setting `ImageURL` in the read callback — the PF envelope writes back the returned model to state, so the field retains its configured value from the prior state (PF's `UseStateForUnknown` / no-set = keep-prior-value semantics for Optional+Computed).

### `configuredString()` elimination

The SDK helper existed because SDKv2 cannot distinguish null from `""` for string attributes. PF `types.String` has three states: null, unknown, known. The API request builder checks `!plan.Description.IsNull()` etc. — no helper needed.

### Composite ID format

The existing SDK resource stores `id = space_id` (not a composite). The PF implementation follows the same pattern: `GetID()` returns the space ID directly, not a composite. `GetResourceID()` also returns the space ID. The entitycore envelope's composite-ID parsing applies only if the stored `id` contains `/`, which it won't for valid Kibana space IDs (the regex enforces `[a-z0-9_-]+`). This preserves import compatibility.

## Risks / Trade-offs

- **State compatibility**: The SDK stores `id` equal to `space_id`. PF stores the same value. No state migration required. The SDK upgrade test verifies this.
- **`disabled_features` set ordering**: The SDK schema used `TypeSet` for this field. PF uses `types.Set` of `types.String`. Both are unordered — no behavioral change.
- **Shared `SpaceModel` refactor**: The existing data source `model` struct is renamed/replaced with `SpaceModel`. This is a non-breaking internal change (the struct is unexported). The data source tests must pass after refactoring.

## Migration Plan

1. Refactor `spaces/models.go`: rename `model` → `SpaceModel`, update `dataSourceModel.Spaces` type
2. Refactor `spaces/schema.go`: extract `spaceAttrs()` reused by both resource and DS schemas
3. Add `fetchSpace()` helper to `spaces/read.go`
4. Implement `spaces/resource.go` + `create.go` + `update.go` + `delete.go`
5. Wire provider, remove SDK registration, delete old file
6. Move + extend tests; verify `make build` + acceptance tests
