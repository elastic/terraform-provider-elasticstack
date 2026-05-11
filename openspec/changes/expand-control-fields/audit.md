# Control field audit (`expand-control-fields`)

Sources: `generated/kbapi/dashboards.json` components
`kbn-controls-schemas-controls-group-schema-{esql,options-list,range-slider,time-slider}-control`
and inner ES|QL branches
`kbn-controls-schemas-options-list-esql-control-schema-{static-values,values-from-query}`,
compared to `internal/kibana/dashboard/schema.go` (`*_control_config` blocks) and the
per-control models (`models_*_control_panel.go`).

## Panel wrapper (all four controls)

The OpenAPI object for each control includes top-level properties alongside `config`:

| API attribute | In TF today | Proposed / notes |
|---------------|-------------|------------------|
| `type` | Yes — `panels[].type` (or equivalent on pinned entries) | N/A |
| `config` | Yes — `*_control_config` nested block | N/A |
| `id` | Yes — `panels[].id` (panel identifier matches API `id` on the control object) | N/A |
| `width` | No | Add optional `width` (string enum `small` \| `medium` \| `large`) on each `*_control_config`, per REQ-039 (colocated with config in HCL though it is a sibling of `config` in JSON). |
| `grow` | No | Add optional `grow` (bool) on each `*_control_config`, per REQ-039. |

## Options list (`options_list_control_config` vs API `config`)

| API `config` field | In TF (`schema.go` / model) | Gap / proposed |
|--------------------|----------------------------|----------------|
| `data_view_id` | `data_view_id` | — |
| `field_name` | `field_name` | — |
| `title` | `title` | — |
| `use_global_filters` | `use_global_filters` | — |
| `ignore_validations` | `ignore_validations` | — |
| `single_select` | `single_select` | — |
| `exclude` | `exclude` | — |
| `exists_selected` | `exists_selected` | — |
| `run_past_timeout` | `run_past_timeout` | — |
| `search_technique` | `search_technique` | — |
| `selected_options` | `selected_options` (list of strings) | Representation: API allows each item to be string **or** number; TF normalizes to string. No new attribute—document/coerce only. |
| `display_settings.placeholder` | `display_settings.placeholder` | — |
| `display_settings.hide_action_bar` | `display_settings.hide_action_bar` | — |
| `display_settings.hide_exclude` | `display_settings.hide_exclude` | — |
| `display_settings.hide_exists` | `display_settings.hide_exists` | — |
| `display_settings.hide_sort` | `display_settings.hide_sort` | — |
| `sort.by` | `sort.by` | — |
| `sort.direction` | `sort.direction` | — |

Inner `config` matches the spec; the only structural omission at the control object level remains **`width` / `grow`** (see wrapper table).

## Range slider (`range_slider_control_config`)

| API `config` field | In TF | Gap / proposed |
|--------------------|-------|----------------|
| `title` | `title` | — |
| `data_view_id` | `data_view_id` | — |
| `field_name` | `field_name` | — |
| `use_global_filters` | `use_global_filters` | — |
| `ignore_validations` | `ignore_validations` | — |
| `value` | `value` (two strings) | — |
| `step` | `step` (float32) | — |

No additional config fields in OpenAPI beyond what TF exposes. Panel **`width` / `grow`** still missing at TF level (REQ-039).

## Time slider (`time_slider_control_config`)

| API `config` field | In TF | Gap / proposed |
|--------------------|-------|----------------|
| `start_percentage_of_time_range` | `start_percentage_of_time_range` | — |
| `end_percentage_of_time_range` | `end_percentage_of_time_range` | — |
| `is_anchored` | `is_anchored` | — |

No additional inner fields. Panel **`width` / `grow`** still missing (REQ-039).

## ES|QL control (`esql_control_config` vs union `config`)

Branches: `STATIC_VALUES` and `VALUES_FROM_QUERY`.

**STATIC_VALUES**

| API field | In TF | Gap / proposed |
|-----------|-------|----------------|
| `control_type` | `control_type` | — |
| `variable_name` | `variable_name` | — |
| `variable_type` | `variable_type` | — |
| `selected_options` | `selected_options` | — |
| `single_select` | `single_select` | — |
| `title` | `title` | — |
| `available_options` | `available_options` | — |
| `display_settings.*` | `display_settings.*` | — |

**VALUES_FROM_QUERY**

| API field | In TF | Gap / proposed |
|-----------|-------|----------------|
| `control_type` | `control_type` | — |
| `variable_name` | `variable_name` | — |
| `variable_type` | `variable_type` | — |
| `selected_options` | `selected_options` | — |
| `single_select` | `single_select` | — |
| `title` | `title` | — |
| `esql_query` | `esql_query` | — |
| `display_settings.*` | `display_settings.*` | — |

Inner union config matches TF; **`width` / `grow`** on the outer control object are the only OpenAPI fields not yet modeled in the nested config block (REQ-039).

## Summary

- **Must add (already in delta spec REQ-039):** `width`, `grow` for all four `*_control_config` shapes (plus null-preservation / import behaviour described there).
- **Additional OpenAPI attributes beyond `width` / `grow`:** none identified for parity with these schemas. Optional follow-up (out of scope for this change unless product asks): stricter validation or docs for `range_slider` `step` API `minimum: 0`, and numeric `selected_options` on options list represented as strings in TF.

## Stack version — control layout fields

Empirical check against **Kibana 9.4.0**: creating or updating a dashboard with `grow` (or `width`) in the control panel JSON returns **HTTP 400** with a body indicating **`Additional properties are not allowed ('grow' was unexpected)`** (and similarly for `width`). The provider therefore omits unset `width` / `grow` on writes and gates acceptance coverage that sends or asserts these keys at **9.5.0** or newer.
