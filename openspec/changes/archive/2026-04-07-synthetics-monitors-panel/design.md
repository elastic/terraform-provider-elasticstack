# Design: Synthetics Monitors Panel Support

## Context

The Kibana Dashboard API exposes the Synthetics monitors panel as the first-class panel type `synthetics_monitors`. Unlike `lens` panels, which embed a Lens visualization specification as a nested saved-object reference, `synthetics_monitors` panels carry their entire configuration inline within the panel `config` object. There is no separate saved object: the panel's filter configuration lives in the dashboard document.

The API schema for `synthetics_monitors` defines a `config` object with a single `filters` sub-object. Every field — including `filters` itself and all of its properties — is optional. The `filters` object has six keys: `projects`, `tags`, `monitor_ids`, `locations`, `monitor_types`, and `statuses`. Each key is an array of `{ label: string, value: string }` objects.

This filter structure is identical to the one used by the `synthetics_stats_overview` panel (REQ-033). The `synthetics_monitors` panel differs from `synthetics_stats_overview` in that it shows a table of monitors and their current status rather than aggregate statistics, and it does not expose `title`, `description`, `hide_title`, `hide_border`, or `drilldowns` fields.

The existing provider architecture for typed panel configs (e.g. `markdown_config`, `slo_overview_config`) maps naturally onto this shape. This design follows those precedents.

## Goals

1. Allow practitioners to manage `synthetics_monitors` panels fully through Terraform.
2. Expose the `filters` block and all filter dimensions in the Terraform schema with idiomatic naming.
3. Preserve null in state when the config block or filters block is omitted, both on initial apply and on read-back from Kibana.
4. Avoid spurious plan diffs when Kibana returns an empty or absent `filters` object.
5. Enable implementation sharing of the filter model with the `synthetics_stats_overview` panel (REQ-033).
6. Be consistent with the existing panel architecture so the implementation is straightforward for contributors familiar with the codebase.

## Non-Goals

- Supporting `synthetics_monitors` panels through `config_json`.
- Adding any fields beyond `filters` (e.g. `title`, `description`, `hide_title`, `hide_border`, `drilldowns`) — the API schema does not expose these for this panel type.
- Changing any existing panel type behavior.

## Decisions

### Terraform shape: typed `synthetics_monitors_config` block

The panel config is modeled as a typed block `synthetics_monitors_config` on the panel object, consistent with `markdown_config`, `slo_overview_config`, and all other typed panel config blocks. The block is entirely optional — a `synthetics_monitors` panel with no filtering requirements can be declared with no config block at all.

```hcl
synthetics_monitors_config = <optional, object({
  filters = <optional, object({
    projects      = <optional, list(object({ label = <required, string>, value = <required, string> }))>
    tags          = <optional, list(object({ label = <required, string>, value = <required, string> }))>
    monitor_ids   = <optional, list(object({ label = <required, string>, value = <required, string> }))>
    locations     = <optional, list(object({ label = <required, string>, value = <required, string> }))>
    monitor_types = <optional, list(object({ label = <required, string>, value = <required, string> }))>
    statuses      = <optional, list(object({ label = <required, string>, value = <required, string> }))>
  })>
})>
```

The block is `optional` and mutually exclusive with all other panel config blocks, enforced by the existing schema-level exclusivity rules (REQ-006).

### All fields optional — no required fields

Unlike most typed panel config blocks, `synthetics_monitors_config` has no required fields at any level. The `filters` block itself is optional. Each filter dimension within `filters` is optional. This reflects the API schema exactly: an empty config object `{}` is a valid API payload for a `synthetics_monitors` panel. Practitioners can declare the panel with no config block, an empty config block, or a config block with a subset of filter dimensions set.

### Read-back null preservation

When the API returns an empty or absent `config` object, the provider SHALL keep `synthetics_monitors_config` null in state. When the API returns a `config` with a present but empty `filters` object, the provider SHALL keep `filters` null in state. This prevents Kibana-injected empty objects from causing spurious diffs.

The implementation SHALL use prior state or plan as the seed for read-back, consistent with the approach used by other typed panel converters, so that fields omitted by Kibana do not overwrite Terraform-authored values with null.

### Shared filter model with REQ-033

The filter structure (`projects`, `tags`, `monitor_ids`, `locations`, `monitor_types`, `statuses`, each a list of `{ label, value }` pairs) is identical to the filter structure used by `synthetics_stats_overview` (REQ-033). The implementation SHOULD share filter model types and converter functions between the two panel types to avoid duplication. If REQ-033 is implemented first, the `synthetics_monitors` implementation can import or reference the shared filter model directly.

### `monitor_ids` list size

The API documents a maximum of 5000 items for `monitor_ids`. This constraint is documented here but is not enforced by a custom Terraform validator. The limit is an API-side constraint; exceeding it will result in an API error rather than a plan-time rejection. A custom `validators.SizeAtMost(5000)` could be added if defensive plan-time validation is desired, but is not required by this spec.

### `config_json` not supported

`config_json` write support is not extended to `synthetics_monitors`. The config object is simple and fully expressible through the typed block; there is no complex nested structure that would motivate a JSON escape hatch. REQ-010 is updated to document that `synthetics_monitors` is explicitly not in the `config_json`-supported set.

## Risks and Trade-offs

| Risk | Mitigation |
|------|-----------|
| Kibana returns an empty `filters` object instead of omitting it, causing drift | Treat an empty or absent `filters` object as equivalent to an omitted `filters` block on read-back |
| Kibana returns empty filter dimension arrays, causing drift | Treat empty filter dimension arrays as equivalent to omitted dimensions on read-back |
| `monitor_ids` can hold up to 5000 items; large lists may cause performance issues in plan | Document the limit; practitioners are responsible for keeping lists within API bounds |
| Filter model divergence between `synthetics_monitors` and `synthetics_stats_overview` if not shared | Enforce code sharing in review; if REQ-033 ships first, the shared types become the canonical reference |

## Migration and State

This change is purely additive. No existing dashboard state is affected. There is no schema version change required for the dashboard resource. Practitioners who have existing `synthetics_monitors` panels managed via `config_json` (which is currently unsupported and returns an error on write) will need to migrate to `synthetics_monitors_config`.

## Open Questions

1. **Empty vs absent `filters` on read-back**: Confirm whether Kibana omits the `filters` key entirely when no filters are set, or returns an empty `{}` object. The implementation should handle both cases as equivalent to null in state.

2. **Filter dimension ordering**: Confirm whether the API preserves the order of items within each filter dimension list (e.g. `projects`). If the API may reorder items, the provider may need order-insensitive comparison or a `UseStateForUnknown`-style plan modifier for filter lists.

3. **Shared filter model packaging**: If REQ-033 (`synthetics_stats_overview`) is not yet implemented when this change is implemented, decide whether to define the shared filter model in the `synthetics_monitors` file and have REQ-033 reference it, or vice versa. A shared file (e.g. `models_synthetics_filters.go`) may be cleaner.
