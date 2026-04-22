# Design: ES|QL Control Panel Support

## Context

The Kibana Dashboard API exposes ES|QL control panels as first-class panel type `esql_control`. Unlike `lens` panels, which embed a Lens visualization specification as a nested saved-object reference, `esql_control` panels carry their entire configuration inline within the panel `config` object. There is no separate saved object for an ES|QL control: the control's query, variable bindings, and display settings all live in the dashboard document.

The API schema for `esql_control` (`kbn-dashboard-panel-esql_control`) defines a `config` object with five required fields (`selected_options`, `variable_name`, `variable_type`, `esql_query`, `control_type`) and several optional fields for display customisation. The `variable_type` and `control_type` fields are string enums with fixed allowed values.

The existing provider architecture for typed panel configs (e.g. `markdown_config`, `xy_chart_config`) maps naturally onto this shape. This design follows those precedents.

## Goals

1. Allow practitioners to manage `esql_control` panels fully through Terraform.
2. Expose all API fields in the Terraform schema with idiomatic naming.
3. Provide clear validation errors for invalid enum values and invalid panel-type/config-block combinations.
4. Preserve `esql_control_config` on read-back without drift from API-injected defaults.
5. Be consistent with the existing panel architecture so the implementation is straightforward for contributors familiar with the codebase.

## Non-Goals

- Supporting `esql_control` panels through `config_json` (see decision below).
- Supporting any other Kibana control types (e.g. `optionsListControl`, `rangeSliderControl`) in this change.
- Changing any existing panel type behavior.

## Decisions

### Terraform shape: typed `esql_control_config` block

The panel config is modeled as a typed block `esql_control_config` on the panel object, consistent with `markdown_config` and all Lens-typed config blocks. This gives practitioners named, documented attributes rather than an opaque JSON blob, and allows Terraform to validate enum values at plan time.

The block is `optional` and mutually exclusive with all other panel config blocks, enforced by the existing schema-level exclusivity rules (REQ-006).

```hcl
esql_control_config = {
  # Required
  selected_options = list(string)
  variable_name    = string
  variable_type    = string  # enum: fields | values | functions | time_literal | multi_values
  esql_query       = string
  control_type     = string  # enum: STATIC_VALUES | VALUES_FROM_QUERY

  # Optional
  title            = string
  single_select    = bool
  available_options = list(string)

  display_settings = {
    placeholder     = string
    hide_action_bar = bool
    hide_exclude    = bool
    hide_exists     = bool
    hide_sort       = bool
  }
}
```

### Required fields

The five fields required by the API schema are marked required in the Terraform schema:
- `selected_options` (list of strings, may be empty list for query-backed controls)
- `variable_name`
- `variable_type`
- `esql_query`
- `control_type`

These cannot be inferred from defaults and must be explicit for the control to function.

### Optional fields

- `title`: human-readable label shown above the control widget. Optional; omitted when not set.
- `single_select`: restricts the control to single-value selection. Defaults to multi-select when omitted.
- `available_options`: pre-populates the dropdown for `VALUES_FROM_QUERY` controls before the query executes. Optional.
- `display_settings`: nested block of UI visibility flags and placeholder text. Optional; the entire block may be omitted.

### `display_settings` as a nested block

`display_settings` is modeled as a single-level nested block (not flat attributes on the parent) because:
1. It maps directly to the API's `display_settings` sub-object.
2. It groups related UI flags semantically.
3. It is consistent with how Kibana's schema separates display concerns from functional config.

All attributes within `display_settings` are optional booleans or strings with no required fields, so the block itself is entirely optional.

### `config_json` support for `esql_control`

`config_json` write support is **not** extended to `esql_control` in this change. Reasons:

1. The `esql_control` config object is simple and fully expressible through typed attributes; there is no complex nested structure that would motivate JSON escape hatches.
2. Extending `config_json` to a new panel type requires changes to the write-path dispatcher (REQ-010) and would need a corresponding unmarshalling target type. This adds complexity without clear benefit when the typed block covers all fields.
3. Keeping `esql_control` typed-only simplifies validation: enum values for `variable_type` and `control_type` can be enforced at the schema layer without inspecting raw JSON.

If a future use case requires `config_json` for `esql_control` (e.g. forward-compatibility with new API fields), that can be added in a follow-up change. REQ-010 is updated to document that `esql_control` is explicitly not in the `config_json`-supported set.

### Enum validation

`variable_type` is restricted to: `fields`, `values`, `functions`, `time_literal`, `multi_values`.
`control_type` is restricted to: `STATIC_VALUES`, `VALUES_FROM_QUERY`.

These are enforced in the Terraform schema using `validators.OneOf` (or equivalent), so invalid values are caught at plan time rather than at apply time.

### Read-back and drift prevention

The `esql_control` config has no known Kibana-injected defaults that differ from the authored values. On read-back, the provider will populate all present fields from the API response. Fields absent from the API response (e.g. optional `title`) will not be forced into state.

The `selected_options` field is an ordered list. The provider will preserve the API-returned ordering on read, which should match the order authored in Terraform for `STATIC_VALUES` controls. For `VALUES_FROM_QUERY` controls `selected_options` represents previously-selected values and may change at runtime; the provider will refresh this from the API on each read.

## Risks and Trade-offs

| Risk | Mitigation |
|------|-----------|
| `selected_options` may drift for `VALUES_FROM_QUERY` controls if Kibana persists the last-selected values | Document this behavior in the resource description; practitioners may wish to set `selected_options = []` for query-backed controls to avoid drift |
| New enum values added by Kibana in a future release will cause a plan-time error for any practitioner using them | Accept this risk; it is consistent with how other enum-validated fields in the provider behave, and new values are added via provider version updates |
| API field additions to `esql_control` config not yet in the typed schema cannot be managed | Practitioners cannot use those fields until a schema update; `config_json` could unblock this in future if needed |

## Migration and State

This change is purely additive. No existing dashboard state is affected. There is no schema version change required for the dashboard resource. Practitioners who have existing `esql_control` panels managed via `config_json` (which is currently unsupported and returns an error on write) will need to migrate to `esql_control_config`.

## Open Questions

1. **`selected_options` drift for `VALUES_FROM_QUERY` controls**: Should the provider suppress drift for this field when `control_type = "VALUES_FROM_QUERY"`? This could be done via a `UseStateForUnknown`-style behavior or a plan modifier. Decision deferred to implementation; the spec documents the current expectation (refresh from API) and the risk.

2. **`available_options` semantics**: The API schema lists `available_options` as an array of strings but does not document whether it is a cached result or an authored value. Implementation should confirm whether this field round-trips as authored or is overwritten by Kibana. If overwritten, state preservation behavior analogous to `selected_options` may be needed.

3. **`uid` field**: The API schema includes an optional panel-level `uid` field. The existing panel schema uses `id` for the panel identifier (mapped to `UseNonNullStateForUnknown`). Confirm that `uid` in the API corresponds to the existing `id` attribute and does not require a separate field on the `esql_control_config` block.
