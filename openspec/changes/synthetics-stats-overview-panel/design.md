# Design: Synthetics Stats Overview Panel Support

## Context

The Kibana Dashboard API exposes Synthetics stats overview panels as a first-class panel type `synthetics_stats_overview`. Unlike `lens` panels, which embed a Lens visualization specification and require data view or ES|QL configuration, `synthetics_stats_overview` panels carry their entire configuration inline within the panel `config` object and drive their data directly from the Elastic Synthetics engine. There is no separate saved object: all display settings, filter constraints, and drilldown actions live within the dashboard document.

The API schema for `synthetics_stats_overview` defines a `config` object where every field is optional. The panel is functional with an empty config — it shows monitoring statistics for all monitors visible to the space. Filtering is structured: each filter category (projects, tags, monitor IDs, locations, monitor types, statuses) is an independent array of `{ label, value }` objects, allowing Kibana to render human-readable filter chips while using the `value` for actual filtering. URL drilldown actions are shared across multiple Kibana embeddable types and follow the same `{ url, label, trigger, type, encode_url, open_in_new_tab }` shape.

The existing provider architecture for typed panel configs (e.g. `markdown_config`, `options_list_control_config`) maps naturally onto this shape. This design follows those precedents.

## Goals

1. Allow practitioners to manage `synthetics_stats_overview` panels fully through Terraform, including the case where no config is specified at all.
2. Expose all API fields in the Terraform schema with idiomatic naming and structured types rather than JSON blobs.
3. Model `filters` as a nested typed block with a named attribute per filter category, each accepting a list of `{ label, value }` objects.
4. Model `drilldowns` as a list of typed objects.
5. Preserve all-null or absent config on read-back without drift.
6. Be consistent with the existing panel architecture so the implementation is straightforward for contributors familiar with the codebase.

## Non-Goals

- Supporting `synthetics_stats_overview` panels through `config_json` (see decision below).
- Managing Elastic Synthetics monitors or projects as Terraform resources (this change concerns only the dashboard panel).
- Changing any existing panel type behavior.

## Decisions

### Terraform shape: typed `synthetics_stats_overview_config` block

The panel config is modeled as a typed block `synthetics_stats_overview_config` on the panel object, consistent with `markdown_config` and all other typed config blocks. This gives practitioners named, documented attributes rather than an opaque JSON blob.

The block is `optional` and mutually exclusive with all other panel config blocks, enforced by the existing schema-level exclusivity rules (REQ-006). When the block is entirely absent, the panel is still valid and shows statistics for all monitors.

```hcl
synthetics_stats_overview_config = {
  # Optional display settings
  title       = string
  description = string
  hide_title  = bool
  hide_border = bool

  # Optional URL drilldowns (max 100)
  drilldowns = list(object({
    url              = string
    label            = string
    trigger          = string  # always "on_open_panel_menu"
    type             = string  # always "url_drilldown"
    encode_url       = bool    # default true
    open_in_new_tab  = bool    # default true
  }))

  # Optional filter sub-block
  filters = object({
    projects      = list(object({ label = string, value = string }))
    tags          = list(object({ label = string, value = string }))
    monitor_ids   = list(object({ label = string, value = string }))  # max 5000
    locations     = list(object({ label = string, value = string }))
    monitor_types = list(object({ label = string, value = string }))
    statuses      = list(object({ label = string, value = string }))
  })
}
```

### All fields are optional

The API schema has no required fields for this panel type. Every attribute within `synthetics_stats_overview_config` is optional in the Terraform schema. An absent or empty config block is equivalent to configuring the panel without any filter constraints, which is the most common use case (show all monitors).

### `filters` as a nested typed block

The `filters` field is modeled as a single nested block (not a JSON string) because:

1. The structure is well-defined: six named categories, each accepting the same `{ label, value }` shape.
2. Named attributes give practitioners autocomplete and plan-time type checking.
3. It is consistent with how other filter-like structures in the provider are modeled (e.g. `options_list_control_config` display settings).
4. The `label` field carries human-readable display text; a typed block makes both `label` and `value` visible and maintainable in HCL.

Each filter category is a `list(object({ label = string, value = string }))`. The entire `filters` block is optional; individual filter categories within it are also optional.

### `drilldowns` as a list of typed objects

The `drilldowns` field is modeled as `list(object({...}))` with named attributes for each drilldown property. This follows the same rationale as `filters`: the structure is well-defined by the API schema and typed attributes are more ergonomic and validatable than a JSON string. The API allows up to 100 drilldowns per panel; this limit is documented but not enforced at schema level (consistent with how other list length constraints are handled in the provider).

### `config_json` support for `synthetics_stats_overview`

`config_json` write support is **not** extended to `synthetics_stats_overview`. Reasons:

1. The config object is simple and fully expressible through typed attributes. There is no deeply nested or polymorphic structure that would motivate a JSON escape hatch.
2. All fields are optional, so practitioners who need only a subset of configuration can simply omit the rest; there is no need for partial-JSON authoring.
3. Keeping the panel typed-only simplifies validation and read-back mapping.

REQ-010 is updated to document that `synthetics_stats_overview` is explicitly not in the `config_json`-supported set.

### Read-back behavior with empty or absent config

When Kibana returns an empty or absent config object for a `synthetics_stats_overview` panel, the provider SHALL preserve `null` in state for the `synthetics_stats_overview_config` block rather than materializing an empty block. This is consistent with the behavior required for other all-optional config blocks (e.g. `markdown_config`).

When specific optional fields within the config are absent from the API response, the provider SHALL not force those fields into state. Fields that Kibana omits SHALL remain null in Terraform state.

### Drilldown defaults

The `encode_url` and `open_in_new_tab` fields in each drilldown default to `true` at the API level. If Kibana omits these fields on read-back, the provider SHALL not force the default value into state (preserving null rather than injecting `true`). If the practitioner explicitly sets these fields in configuration, the authored values SHALL be written to and preserved from the API.

## Risks and Trade-offs

| Risk | Mitigation |
|------|-----------|
| `monitor_ids` accepts up to 5000 entries; very large lists may produce unwieldy plan output | Document the limit; no schema enforcement needed (consistent with other list attributes in the provider) |
| Kibana may add new filter categories or drilldown properties in a future release that are not yet modelled | The converter ignores unmapped fields on read, consistent with other typed panel converters; a follow-up change can add new attributes |
| The `filters` nested block means that omitting the block and omitting all its filter lists are two different states; Kibana materializing an empty filters object could cause drift | Treat a nil or empty `filters` object from the API as equivalent to an omitted block in state |
| `drilldowns` list order returned by Kibana must match the Terraform-authored order; Kibana may reorder on round-trip | Document that drilldown order is preserved as returned by the API; practitioners should be aware that Kibana round-trip order is authoritative on refresh |

## Migration and State

This change is purely additive. No existing dashboard state is affected. There is no schema version change required for the dashboard resource. Practitioners who have existing `synthetics_stats_overview` panels managed via `config_json` (which is currently unsupported and returns an error on write) will need to migrate to `synthetics_stats_overview_config`.

## Open Questions

1. **Empty `filters` object vs. absent `filters`**: Confirm whether Kibana ever emits `"filters": {}` (empty object) or only omits `filters` entirely when no filters are configured. If Kibana emits an empty object, the read converter must treat it as equivalent to an absent block to avoid drift.

2. **Drilldown round-trip ordering**: Confirm that Kibana preserves the authored order of `drilldowns` entries on round-trip. If Kibana reorders entries, a state-preservation or ordering-normalization strategy analogous to panel ordering may be needed.

3. **`trigger` and `type` field validation**: The API schema specifies `trigger = "on_open_panel_menu"` and `type = "url_drilldown"` as the only valid values for drilldown entries. Decide whether to enforce these as constants in the schema (e.g. computed-with-default) or as optional strings with enum validators. Enforcing as constants simplifies configuration but reduces forward compatibility.
