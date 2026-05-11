## Context

Canonical requirements for this resource live in [`openspec/specs/elasticsearch-index/spec.md`](../../specs/elasticsearch-index/spec.md). Implementation lives in [`internal/elasticsearch/index/index/`](../../../internal/elasticsearch/index/index/).

The four `index.sort.*` settings and their defaults per the [Elastic docs](https://www.elastic.co/docs/reference/elasticsearch/index-settings/sorting):

| Setting | Values | Default |
|---|---|---|
| `index.sort.field` | any field with `doc_values` (boolean, numeric, date, keyword) | — |
| `index.sort.order` | `asc`, `desc` | `asc` |
| `index.sort.missing` | `_last`, `_first` | `_last` (regardless of order) |
| `index.sort.mode` | `min`, `max` | `min` when order=asc; `max` when order=desc |

All four are **immutable static settings**: they can only be configured at index creation time and require destroy+recreate to change.

## Goals / Non-Goals

**Goals:**

- Expose all four `index.sort.*` settings via a new `sort` `ListNestedAttribute` per entry.
- Fix the structural design debt: `sort_field` uses a `SetAttribute` (no guaranteed ordering), making positional coupling with `sort_order` unreliable.
- Avoid any breaking change: keep `sort_field` and `sort_order` deprecated but functional.
- Provide seamless migration (no destroy+recreate) when users move from old attributes to the new `sort` block with equivalent field+order settings.

**Non-goals:**

- Removing `sort_field` and `sort_order` (separate change after deprecation period).
- Retroactively fixing the `sort_field` `SetAttribute` ordering bug for existing configs that stay on the legacy path.
- Any changes to dynamic settings or non-sort index settings.
- State upgrade functions for migrating private state across provider versions.

## Decisions

### Schema: `sort` as `ListNestedAttribute`

```go
"sort": schema.ListNestedAttribute{
    Description: "Sort configuration for documents within each shard segment. Replaces the deprecated sort_field and sort_order attributes.",
    Optional:    true,
    PlanModifiers: []planmodifier.List{sortMigrationPlanModifier{}},
    Validators: []validator.List{
        listvalidator.ConflictsWith(path.MatchRoot("sort_field"), path.MatchRoot("sort_order")),
    },
    NestedObject: schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{
            "field": schema.StringAttribute{Required: true},
            "order": schema.StringAttribute{Optional: true,
                Validators: []validator.String{stringvalidator.OneOf("asc", "desc")}},
            "missing": schema.StringAttribute{Optional: true,
                Validators: []validator.String{stringvalidator.OneOf("_last", "_first")}},
            "mode": schema.StringAttribute{Optional: true,
                Validators: []validator.String{stringvalidator.OneOf("min", "max")}},
        },
    },
},
```

`ListNestedAttribute` correctly models the ordered nature of sort entries (unlike the existing `SetAttribute` for `sort_field`). Length mismatches between per-field options are structurally impossible.

### Deprecated attributes: `sort_field` and `sort_order`

Add `DeprecationMessage` and `ConflictsWith` validators to both:

```go
"sort_field": schema.SetAttribute{
    // ... existing ...
    DeprecationMessage: "Use the 'sort' attribute instead. 'sort_field' will be removed in a future major release.",
    Validators: []validator.Set{
        setvalidator.ConflictsWith(path.MatchRoot("sort")),
        // existing validators remain
    },
    PlanModifiers: []planmodifier.Set{legacySortFieldPlanModifier{}},
},
"sort_order": schema.ListAttribute{
    // ... existing ...
    DeprecationMessage: "Use the 'sort' attribute instead. 'sort_order' will be removed in a future major release.",
    Validators: []validator.List{
        listvalidator.ConflictsWith(path.MatchRoot("sort")),
        // existing validators remain
    },
    PlanModifiers: []planmodifier.List{legacySortOrderPlanModifier{}},
},
```

### `RequiresReplaceIf` strategy: private state

**Problem**: `sort_field` is a `SetAttribute`, so its iteration order in Terraform state is non-deterministic with respect to the original insertion order. For multi-field sorts with different orders per field, we cannot reconstruct the correct field→order mapping from Terraform state alone.

**Solution**: Store the ordered sort configuration from Elasticsearch in private state during `Read`. The ES `GET /{index}/_settings` API with `flat_settings=true` returns `sort.field` in originally-configured order. This sidesteps the `SetAttribute` ordering ambiguity entirely.

Private state structure (stored under key `"sort_config"`):

```go
type sortPrivateState struct {
    Fields   []string `json:"fields"`
    Orders   []string `json:"orders"`
    Missing  []string `json:"missing,omitempty"`
    Mode     []string `json:"mode,omitempty"`
}
```

This pattern is already established in the codebase: `internal/elasticsearch/security/api_key/resource.go` stores cluster version in private state during `Read` and reads it in a `RequiresReplaceIf` modifier (lines 81–148).

### Three plan modifiers

**`sortMigrationPlanModifier`** (on new `sort` attribute, implements `planmodifier.List`):

Suppresses replace when ALL of the following are true:
1. `sort` is null in state (migrating from old representation, not changing the new one).
2. Private state has the ordered sort config from ES (key `"sort_config"`).
3. The plan's `sort[*].field` list matches private state's `fields` in the same order.
4. The plan's `sort[*].order` matches private state's `orders` (treating null as `"asc"` default).
5. Planned `sort[*].missing`/`sort[*].mode` are semantically equivalent to existing index settings at each position, treating explicit defaults as equivalent to absent settings (`missing="_last"`, `mode="min"` for `asc` and `"max"` for `desc`).

If private state is absent (first apply after provider upgrade before Read populates it), the modifier defaults to `RequiresReplace = true`. This forces one replace, but is safe and correct.

**`legacySortFieldPlanModifier`** (on `sort_field`, implements `planmodifier.Set`):

Suppresses replace for `sort_field` when `sort` is present in the plan (the new attribute handles the replace decision). Uses `req.Plan.GetAttribute(ctx, path.Root("sort"), &planSort)` — this cross-attribute read pattern is established in `set_unknown_if_access_has_changes.go`.

**`legacySortOrderPlanModifier`** (on `sort_order`, implements `planmodifier.List`):

Same logic as `legacySortFieldPlanModifier` — suppresses replace when `sort` is present in the plan.

### Model changes

Add to `tfModel`:
```go
Sort types.List `tfsdk:"sort"` // ListNestedAttribute; elements are sortEntryModel objects
```

Add `sortEntryModel`:
```go
type sortEntryModel struct {
    Field   types.String `tfsdk:"field"`
    Order   types.String `tfsdk:"order"`
    Missing types.String `tfsdk:"missing"`
    Mode    types.String `tfsdk:"mode"`
}
```

The `toIndexSettings()` reflection loop cannot handle `ListNestedAttribute`. Add a pre-processing step: when `Sort` is non-null, expand to parallel `sort.field`, `sort.order`, `sort.missing`, `sort.mode` arrays in the settings map (omitting null-valued fields), then skip those ES keys in the reflection loop.

Add `"sort.missing"` and `"sort.mode"` to `staticSettingsKeys` so the `use_existing` adoption flow compares them correctly.

### Read/flatten path

When `Read` populates state from the ES response:
- If `sort` in current state is non-null: populate `sort` from ES response fields.
- If `sort` in current state is null (still using legacy path): populate `sort_field` and `sort_order` as before.
- Always store ordered sort config in private state under `"sort_config"`.

### `use_existing.go` additions

Add `case "sort.missing"` and `case "sort.mode"` branches using `stringSliceOrderedFromAny`, mirroring the existing `"sort.order"` case.

## Risks / Trade-offs

- **First apply after upgrade forces replace once**: When private state is absent, `sortMigrationPlanModifier` defaults to `RequiresReplace = true`. This is safe but should be documented in the release notes.
- **Three plan modifiers with cross-attribute reads**: Modest complexity; existing codebase patterns make this manageable.
- **Custom expand/flatten for nested sort**: Cannot reuse reflection; adds explicit expand code in `toIndexSettings()` and explicit flatten code in `Read`.
- **Non-equivalent `missing`/`mode` on migration requires replace**: Unavoidable due to immutable static settings; however explicit default values are treated as equivalent to absent settings to avoid unnecessary replacement.

## Migration Guide (for documentation)

Users migrating from `sort_field`/`sort_order` to the new `sort` attribute:

```hcl
# Before (deprecated, still works)
sort_field = ["date", "username"]
sort_order = ["desc", "asc"]

# After (same ES settings, no replace triggered if field+order are equivalent)
sort = [
  { field = "date",     order = "desc" },
  { field = "username", order = "asc"  },
]
```

**Important**: Non-equivalent `missing` or `mode` values in the new `sort` block will destroy and recreate the index. Explicit defaults (`missing="_last"`, `mode` based on `order`) are treated as equivalent to absent settings and do not trigger replacement by themselves.

On the first `terraform apply` after upgrading to the provider version containing this change, if private state has not yet been populated by a `terraform refresh` or prior read, replace may be forced once. Running `terraform refresh` before `terraform apply` avoids this.

## Open Questions

- **Post-upgrade forced replace on first apply**: Is the single forced replace (when private state is absent on first apply after provider upgrade) acceptable? Or should we implement a state upgrade path that pre-populates private state from the existing `sort_field`/`sort_order` Terraform state values?
- **Deprecation timeline**: How long before `sort_field` and `sort_order` are removed? Major version bump required?
- **Cross-list length validator for deprecated path**: Should a cross-attribute length validator be added to enforce `len(sort_order) == len(sort_field)` to prevent silent ES 400s?
