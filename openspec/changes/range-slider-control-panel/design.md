## Context

`range_slider_control` is a standalone (non-Lens) Kibana dashboard panel type that renders a min/max range slider filter. Its API shape is defined by `kbn-dashboard-panel-range_slider_control` in the dashboard API. Unlike Lens panels, its config is fully inline — there is no separate saved object reference. The panel requires `data_view_id` and `field_name`; all other config fields are optional.

The panel config is carried in the `config` object alongside `type` and `grid` at the panel level. The API schema marks `config` as `additionalProperties: true`, meaning the provider should treat unrecognized fields as pass-through on read and normalize only the known fields.

## Goals

- Expose `range_slider_control` panels through a fully typed `range_slider_control_config` block with ergonomic attribute names.
- Enforce `data_view_id` and `field_name` as required attributes within the config block.
- Validate that `value`, when set, contains exactly 2 elements (the min and max bounds).
- Preserve the `value` pair accurately on read/write as a `list(string)` matching the API's 2-element array representation.
- Keep the implementation consistent with other standalone panel types (e.g. ES|QL control).

## Non-Goals

- Supporting `config_json` write for `range_slider_control` panels (must use the typed block).
- Exposing API `additionalProperties` as a catch-all JSON attribute; only the documented fields are surfaced.
- Changing any behavior for existing panel types.

## Decisions

| Topic | Decision |
|-------|----------|
| Terraform shape | Single nested object `range_slider_control_config` at panel level (consistent with `markdown_config`, `esql_control_config`, and other typed panel blocks). |
| `value` representation | `list(string)` with a schema validator enforcing exactly 2 elements. The API models `value` as a 2-element array of strings (`[min, max]`). A flat list keeps the shape faithful to the API while avoiding the added complexity of separate `value_min` / `value_max` attributes. Practitioners who want to omit an initial range simply leave `value` unset. |
| `step` type | `number` (float64) to match the API's numeric type; most use cases will be whole numbers but fractional steps are valid. |
| Required fields | `data_view_id` and `field_name` are required within the config block (matching the API `required` constraint). |
| `config_json` write | Not extended to `range_slider_control`. REQ-010 continues to restrict `config_json` write to `markdown` and `lens` only. |
| Read normalization | On read, the provider SHALL populate `range_slider_control_config` from the API response fields. Fields absent from the API response SHALL be kept as null/unset in state, consistent with how the implementation handles optional fields on other standalone panel types. |
| Mutual exclusion | `range_slider_control_config` conflicts with all other config blocks (same constraint as every other typed panel config block). |

## Risks / Trade-offs

- **`value` as `list(string)`**: A 2-element string list is less self-documenting than named `value_min` / `value_max` attributes, but it more faithfully reflects the API tuple shape and avoids introducing provider-specific coupling between two correlated attributes. The exact-2-elements validator catches misconfiguration at plan time.
- **`additionalProperties: true` on the API `config` object**: Future Kibana versions may add fields to `range_slider_control` config that the provider does not yet surface. Those fields will be silently dropped on round-trip. This is acceptable for a v1 implementation; a `config_json` escape hatch is not planned for this type.

## Migration / State

- This is a new optional block on an existing resource. No state migration is needed; dashboards that do not use `range_slider_control` panels are unaffected.

## Open Questions

- None. All design questions are resolved in the decisions table above.
