# Design: `lens-dashboard-app` Panel Support

## Context

The Kibana Dashboard API exposes a panel type `lens-dashboard-app` that is distinct from the existing `lens` panel type already handled by the provider. The type string must match exactly in the dashboard payload — `lens-dashboard-app` and `lens` are treated as separate panel types by Kibana, even though both ultimately render Lens visualizations.

The `lens-dashboard-app` type supports two embedding modes:

1. **By-value**: The panel payload carries a full Lens chart definition in an `attributes` object. This is analogous to the existing `lens` panel type (which always embeds inline), but uses the different type string `lens-dashboard-app` and a different payload structure (the top-level `attributes` key rather than the embedded chart type directly in the config).

2. **By-reference**: The panel payload contains a `saved_object_id` pointing to an existing saved Lens visualization saved object. The Kibana API resolves this reference at render time. This is the key new capability: it allows dashboards to embed shared, centrally-managed visualizations without duplicating their definitions.

Both modes share a set of optional display fields: `title`, `description`, `hide_title`, `hide_border`, and `time_range`. By-reference panels additionally support `overrides`.

The existing typed Lens config blocks (`xy_chart_config`, `metric_chart_config`, etc.) all apply to `type = "lens"` panels and are not affected by this change. The `lens-dashboard-app` type requires its own config block because it has a structurally different payload and supports the by-reference mode that the existing blocks cannot express.

## Goals

1. Allow practitioners to embed saved Lens visualizations by reference (`saved_object_id`) in a dashboard via Terraform.
2. Allow practitioners to embed Lens visualizations by value using the `lens-dashboard-app` type string via a JSON attributes escape hatch.
3. Expose all API fields for both modes in the Terraform schema with idiomatic naming.
4. Enforce mutual exclusivity between `by_value` and `by_reference` sub-blocks at plan time.
5. Be consistent with the existing panel architecture (mutually exclusive sub-blocks pattern, as used by `datatable_config`).

## Non-Goals

- Supporting `lens-dashboard-app` panels through `config_json`.
- Providing typed attributes for the full Lens chart specification within `by_value` (that complexity belongs to the existing typed Lens config blocks for the `lens` panel type; here, `attributes_json` is an opaque JSON string).
- Changing any existing panel type behavior or typed Lens config block.
- Migrating existing `lens` panels to `lens-dashboard-app`.

## Decisions

### Terraform shape: `lens_dashboard_app_config` with `by_value` and `by_reference` sub-blocks

The panel config is modeled as a typed block `lens_dashboard_app_config` containing two mutually exclusive nested blocks: `by_value` and `by_reference`. This pattern mirrors `datatable_config`'s `no_esql`/`esql` sub-block approach and cleanly separates the two modes.

```hcl
lens_dashboard_app_config = {
  # Exactly one of by_value or by_reference must be set

  by_value = {
    attributes_json = string  # required; JSON string containing the Lens chart attributes
    references_json = string  # optional; JSON array of { id, name, type } objects for data view references
  }

  by_reference = {
    saved_object_id = string  # required; ID of the saved Lens visualization saved object
    overrides_json  = string  # optional; JSON object for overrides to the saved Lens object
  }

  # Shared optional fields
  title       = string
  description = string
  hide_title  = bool
  hide_border = bool

  time_range = {
    from = string  # required when block is present
    to   = string  # required when block is present
  }
}
```

The block is `optional` and mutually exclusive with all other panel config blocks (enforced by REQ-006).

### By-value: `attributes_json` as opaque JSON string

The `by_value.attributes_json` field accepts a JSON string containing any Lens chart type recognized by the Kibana API (`metricChart`, `legacyMetricChart`, `xyChart`, `gaugeChart`, `heatmapChart`, `tagcloudChart`, `regionMapChart`, `datatableChart`, `pieChart`, `mosaicChart`, `treemapChart`, `waffleChart`). This is intentionally opaque: providing full typed sub-attributes for every chart variant would duplicate the existing typed Lens config blocks and is out of scope for this change.

`attributes_json` uses default-aware semantic JSON equality for plan comparison, consistent with how other `*_json` fields in the provider behave.

### By-value: `references_json` as optional JSON string

`by_value.references_json` carries the `references` array (objects with `id`, `name`, `type`) used by the Lens chart to resolve data view references. It is optional because some Lens charts (e.g. pure ES|QL-backed) may not require external references.

### By-reference: `saved_object_id` as required string

The `by_reference.saved_object_id` field is the Kibana saved object ID of the target Lens visualization. It is required and must be a non-empty string. The provider does not validate that the saved object exists at plan time; validation occurs at apply time when the Kibana API is called.

### By-reference: `overrides_json` as optional JSON string

`by_reference.overrides_json` accepts an optional JSON object for runtime overrides on the saved Lens object. The structure of `overrides` is not constrained by the provider schema (it is API-defined and may evolve); using a JSON string escape hatch avoids schema churn for new override fields.

### Shared optional fields: flat on `lens_dashboard_app_config`

`title`, `description`, `hide_title`, and `hide_border` are placed as flat optional attributes directly on `lens_dashboard_app_config` rather than in a nested display block. This is consistent with how `markdown_config`, `xy_chart_config`, and other panel config blocks expose title and description at the top level.

### `time_range` as a nested block

`time_range` is a single-value nested block with required `from` and `to` string attributes. This matches the structure of the `time_range` object in the API schema. The block itself is optional; when omitted, no panel-level time range override is sent to the API.

### `config_json` support for `lens-dashboard-app`

`config_json` write support is **not** extended to `lens-dashboard-app` in this change. Reasons:

1. The by-reference mode is the primary new capability, and it is not representable as a Lens `config_json` (which expects a full chart definition, not a saved object reference).
2. The by-value mode exposes `attributes_json` as the appropriate JSON escape hatch; a second raw escape hatch via `config_json` would create ambiguity.
3. Keeping `lens-dashboard-app` typed-only maintains clear separation from the existing `config_json`-supported `lens` panel type.

REQ-025 is updated to explicitly name `lens-dashboard-app` as not in the `config_json`-supported set.

### Relationship to existing `lens` panel type

The existing typed Lens config blocks (`xy_chart_config`, `metric_chart_config`, `waffle_config`, etc.) continue to apply exclusively to `type = "lens"` panels. They are not valid for `type = "lens-dashboard-app"`. This distinction is enforced by the type-specific validation rules in REQ-006.

Practitioners with existing `lens`-typed panels do not need to migrate. The two types coexist. The `lens-dashboard-app` type is the correct choice when:
- Referencing a saved Lens visualization by ID (by-reference mode), or
- Embedding a Lens visualization where the panel type string must be `lens-dashboard-app` (e.g. dashboards exported from Kibana that use this type).

### Read-back and drift prevention

On read-back, the provider determines the mode (by-value vs by-reference) by inspecting the API response for the presence of `attributes` (by-value) or `saved_object_id` (by-reference). The appropriate sub-block is populated; the other is left null.

`attributes_json` and `references_json` use default-aware semantic JSON equality, so API-injected ordering changes or default field additions do not create spurious plan diffs.

`overrides_json` similarly uses semantic JSON equality.

## Risks and Trade-offs

| Risk | Mitigation |
|------|-----------|
| `by_value.attributes_json` is opaque JSON; practitioners must know the Lens chart schema | Document that `attributes_json` maps to the Lens `attributes` object in the Kibana API; link to existing typed Lens config blocks as a reference for supported chart shapes |
| `by_reference.saved_object_id` referencing a non-existent saved object will fail at apply time | Accept this; provider does not pre-validate saved object existence. Error from Kibana API will be surfaced as a diagnostic |
| `overrides_json` schema is undocumented and may change across Kibana versions | Using an opaque JSON string avoids schema churn; practitioners accept API evolution risk |
| Mode detection on read-back (presence of `attributes` vs `saved_object_id`) could misclassify a panel if Kibana adds new payload shapes | Treat unrecognized payload as an error diagnostic; document the detection logic |

## Migration and State

This change is purely additive. No existing dashboard state is affected. There is no schema version change required for the dashboard resource. Practitioners who have existing `lens-dashboard-app` panels managed via `config_json` (which returns an error on write for unsupported types) will need to migrate to `lens_dashboard_app_config`.

## Open Questions

1. **Mode detection ambiguity**: Can a `lens-dashboard-app` API response contain both `attributes` and `saved_object_id`? If so, which field takes precedence for mode detection? Implementation should confirm the API contract and document the detection priority.

2. **`references_json` round-trip**: Does Kibana modify or enrich the `references` array (e.g. by resolving IDs or adding type fields) between write and read-back? If so, the provider may need to use the API-returned value as authoritative on read to avoid drift.

3. **`overrides_json` defaults**: Does Kibana inject default values into the `overrides` object on read-back? If so, semantic JSON equality or a plan modifier may be needed to suppress drift.

4. **`hide_border` API support**: Confirm that `hide_border` is a first-class field in the `lens-dashboard-app` panel payload and not specific to a Kibana version range. If it was introduced in a recent version, document the minimum supported Kibana version.
