## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

Schema validation SHALL enforce that `options_list_control_config` is valid only for panels with `type = "options_list_control"`, is mutually exclusive with all other panel configuration blocks and with `config_json`, and that `search_technique` is restricted to `prefix`, `wildcard`, or `exact` when set.

REQ-006 is extended to include:

- `options_list_control_config` SHALL be valid only for panels with `type = "options_list_control"`.
- `options_list_control_config` SHALL be mutually exclusive with all other panel configuration blocks and with `config_json`.
- The `search_technique` attribute within `options_list_control_config` SHALL be restricted to the values `prefix`, `wildcard`, and `exact` when set; any other value SHALL be rejected at plan time.

#### Scenario: options_list_control_config rejected for non-options_list_control panel

- GIVEN a panel with `type = "lens"` and `options_list_control_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

`config_json` SHALL NOT be supported for `options_list_control` panels; the `options_list_control` panel type SHALL be managed exclusively through the typed `options_list_control_config` block.

The existing REQ-010 text:

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, or omitting all panel configuration blocks, SHALL return an error diagnostic.

is updated to additionally state:

> The `options_list_control` panel type SHALL be managed exclusively through the typed `options_list_control_config` block; using `config_json` with `type = "options_list_control"` SHALL return an error diagnostic.

#### Scenario: config_json rejected for options_list_control panel type

- GIVEN a panel with `type = "options_list_control"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `options_list_control`

---

## ADDED Requirements

### Requirement: Options list control panel behavior (REQ-027)

When a panel entry sets `type = "options_list_control"`, the resource SHALL accept an `options_list_control_config` block and SHALL require that block to be present. The block SHALL require `data_view_id` (string) and `field_name` (string). All other attributes in the block SHALL be optional:

- `title` (string) — human-readable label displayed above the control.
- `use_global_filters` (bool) — whether the control applies the dashboard's global filters to its own query.
- `ignore_validations` (bool) — whether the control skips field-level validation against the data view.
- `single_select` (bool) — when true, only one option may be selected at a time.
- `exclude` (bool) — when true, the selected options are used as an exclusion filter rather than an inclusion filter.
- `exists_selected` (bool) — when true, the control filters for documents where the field exists.
- `run_past_timeout` (bool) — when true, the control continues to show results even when the underlying query times out.
- `search_technique` (string) — the technique used to match suggestions; MUST be one of `prefix`, `wildcard`, or `exact` when set.
- `selected_options` (list of string) — the initially or persistently selected option values; the provider SHALL represent all selected options as strings regardless of whether the API stores them as numbers.
- `display_settings` (nested block, optional) — display preferences for the control widget, containing:
  - `placeholder` (string) — placeholder text shown when no option is selected.
  - `hide_action_bar` (bool) — when true, hides the action bar on the control.
  - `hide_exclude` (bool) — when true, hides the exclude toggle.
  - `hide_exists` (bool) — when true, hides the exists filter option.
  - `hide_sort` (bool) — when true, hides the sort control.
- `sort` (nested block, optional) — default sort configuration for the suggestion list, containing:
  - `by` (string) — the field or criterion to sort by.
  - `direction` (string) — the sort direction.

The `options_list_control_config` block SHALL conflict with all other typed panel config blocks (`markdown_config`, `xy_chart_config`, `treemap_config`, `mosaic_config`, `datatable_config`, `tagcloud_config`, `heatmap_config`, `waffle_config`, `region_map_config`, `gauge_config`, `metric_chart_config`, `pie_chart_config`, `legacy_metric_config`) and with `config_json`. When `type` is `options_list_control`, no other typed config block or `config_json` SHALL be present on the same panel entry.

For API mapping, the provider SHALL write the `options_list_control_config` attributes into the panel's `config` object as defined by the `kbn-dashboard-panel-options_list_control` API schema. On read-back, the provider SHALL populate all attributes that are present in the API response and SHALL treat a nil or empty `display_settings` API object as equivalent to an omitted `display_settings` block in state.

#### Scenario: Options list control panel requires data_view_id and field_name

- GIVEN a panel entry with `type = "options_list_control"` and an `options_list_control_config` block that omits `data_view_id` or `field_name`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating that `data_view_id` and `field_name` are required

#### Scenario: Options list control panel with invalid search_technique

- GIVEN a panel entry with `type = "options_list_control"` and `options_list_control_config.search_technique` set to a value other than `prefix`, `wildcard`, or `exact`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating the value is not one of the accepted enum values

#### Scenario: Options list control panel round-trips through Kibana

- GIVEN a dashboard with an `options_list_control` panel that sets `data_view_id`, `field_name`, `search_technique = "prefix"`, `single_select = true`, and a `display_settings` block
- WHEN the provider creates the dashboard and reads it back
- THEN all configured attributes SHALL be present in state and a subsequent plan SHALL show no changes

#### Scenario: Options list control config conflicts with other typed blocks

- GIVEN a panel entry with `type = "options_list_control"` that sets both `options_list_control_config` and any other typed config block (e.g. `markdown_config`)
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating the conflicting blocks are mutually exclusive
