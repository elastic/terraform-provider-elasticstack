## CHANGED Requirements

### Requirement: Options list control panel behavior (REQ-027)

**Replaces the existing REQ-027 in full.**

When a panel entry sets `type = "options_list_control"`, the resource SHALL accept an `options_list_control_config` block with the following structure. The block MUST contain exactly one of two mutually exclusive nested blocks: `by_field` or `by_esql`. Having both or neither SHALL produce an error diagnostic at plan time.

#### `by_field` block (Field variant)

The `by_field` nested block represents a control sourced from a Kibana data view field:

- `data_view_id` (required, string) — the ID of the Kibana data view the control is tied to.
- `field_name` (required, string) — the name of the field within the data view.
- `title` (optional, string) — human-readable label displayed above the control.
- `use_global_filters` (optional, bool) — whether the control applies the dashboard's global filters to its own query.
- `ignore_validations` (optional, bool) — whether the control skips field-level validation against the data view.
- `single_select` (optional, bool) — when true, only one option may be selected at a time.
- `exclude` (optional, bool) — when true, selected options are used as an exclusion filter rather than an inclusion filter.
- `exists_selected` (optional, bool) — when true, the control filters for documents where the field exists.
- `run_past_timeout` (optional, bool) — when true, the control continues to show results even when the underlying query times out.
- `search_technique` (optional, string) — must be one of `prefix`, `wildcard`, or `exact` when set.
- `selected_options` (optional, list of string) — the initially or persistently selected option values; all values are represented as strings.
- `display_settings` (optional, nested block) — display preferences for the control widget, containing:
  - `placeholder` (string) — placeholder text shown when no option is selected.
  - `hide_action_bar` (bool) — when true, hides the action bar on the control.
  - `hide_exclude` (bool) — when true, hides the exclude toggle.
  - `hide_exists` (bool) — when true, hides the exists filter option.
  - `hide_sort` (bool) — when true, hides the sort control.
- `sort` (optional, nested block) — default sort configuration for the suggestion list, containing:
  - `by` (required, string) — must be one of `_count` or `_key`.
  - `direction` (required, string) — must be one of `asc` or `desc`.

On write, the provider SHALL set `values_source = "field"` automatically on the API payload for the Field branch; this field SHALL NOT be exposed as a Terraform attribute.

#### `by_esql` block (ES|QL variant)

The `by_esql` nested block represents a control sourced from an ES|QL query:

- `esql_query` (required, string) — the ES|QL query that produces the available option values.
- `values_source` (required, string) — the source discriminator; MUST be `"esql_query"`. Any other value SHALL produce an error diagnostic at plan time.
- `title` (optional, string) — same as `by_field`.
- `use_global_filters` (optional, bool) — same as `by_field`.
- `ignore_validations` (optional, bool) — same as `by_field`.
- `single_select` (optional, bool) — same as `by_field`.
- `exclude` (optional, bool) — same as `by_field`.
- `exists_selected` (optional, bool) — same as `by_field`.
- `run_past_timeout` (optional, bool) — same as `by_field`.
- `search_technique` (optional, string) — same as `by_field` (must be one of `prefix`, `wildcard`, or `exact` when set).
- `selected_options` (optional, list of string) — same as `by_field`.
- `display_settings` (optional, nested block) — same structure as `by_field`.
- `sort` (optional, nested block) — same structure as `by_field`.

#### Null-preservation and import semantics

During import (no prior state), the provider SHALL populate the branch-specific required identifiers from the API response: for `by_field`, populate `data_view_id` and `field_name`; for `by_esql`, populate `esql_query` and `values_source`. In both branches, `title`, `search_technique`, `selected_options`, and `display_settings` SHALL be populated where present in the API response; optional booleans and `sort` SHALL be left null.
#### Mutual exclusion and conflict guards

- Exactly one of `by_field` or `by_esql` MUST be set in `options_list_control_config`.
- `options_list_control_config` SHALL remain mutually exclusive with all other typed panel config blocks and with `config_json` (unchanged from existing REQ-027).

#### State migration (v0 → v1)

The change from a flat `options_list_control_config` schema to the two-branch nested schema constitutes a schema version increment. The resource SHALL implement a Plugin Framework `ResourceWithUpgradeState` upgrader that rewrites existing v0 state by moving all flat `options_list_control_config` attributes under a `by_field {}` object. The upgrader SHALL handle both in-grid `panels[]` entries and `pinned_panels` entries.

#### Scenarios

##### Scenario: Field variant round-trip

- GIVEN a panel with `type = "options_list_control"` whose `options_list_control_config` sets `by_field = { data_view_id = "logs-view", field_name = "service.name", search_technique = "prefix", single_select = true }`
- WHEN the provider creates the dashboard and reads it back
- THEN all configured attributes SHALL be present in state under `options_list_control_config.by_field` and a subsequent plan SHALL show no changes

##### Scenario: ES|QL variant round-trip

- GIVEN a panel with `type = "options_list_control"` whose `options_list_control_config` sets `by_esql = { esql_query = "FROM logs | STATS ...", values_source = "esql_query", title = "Service" }`
- WHEN the provider creates the dashboard and reads it back
- THEN `options_list_control_config.by_esql.esql_query` and `values_source` SHALL be present in state and a subsequent plan SHALL show no changes

##### Scenario: Field variant null-preservation on import

- GIVEN an existing dashboard whose options-list control is a field-backed control with Kibana server-side defaults for `use_global_filters`, `ignore_validations`, `exclude`, `exists_selected`, `run_past_timeout`, and `sort`
- WHEN the provider imports the dashboard resource
- THEN `data_view_id` and `field_name` SHALL be populated under `by_field` in state
- AND optional booleans and `sort` SHALL remain null in state
- AND a subsequent plan against a configuration that omits them SHALL show no changes

##### Scenario: Both branches rejected

- GIVEN an `options_list_control_config` block that sets both `by_field` and `by_esql`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating the two branches are mutually exclusive

##### Scenario: Neither branch rejected

- GIVEN an `options_list_control_config` block with neither `by_field` nor `by_esql` set
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating exactly one branch must be configured

##### Scenario: Invalid values_source on by_esql

- GIVEN `by_esql = { esql_query = "...", values_source = "field" }`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating `values_source` must be `"esql_query"`

##### Scenario: State upgrade from v0 flat to v1 by_field

- GIVEN a Terraform state that contains `options_list_control_config` with flat attributes (v0 schema: `data_view_id`, `field_name`, etc. at the config root)
- WHEN the provider with the updated schema is applied
- THEN the state upgrader SHALL rewrite the flat attributes under `by_field` and the resulting v1 state SHALL be equivalent to configuring `by_field { data_view_id = ..., field_name = ..., ... }`

---

### Requirement: Range slider control panel behavior (REQ-028)

**Replaces the existing REQ-028 in full.**

When a panel entry sets `type = "range_slider_control"`, the resource SHALL accept a `range_slider_control_config` block with the following structure. The block MUST contain exactly one of two mutually exclusive nested blocks: `by_field` or `by_esql`. Having both or neither SHALL produce an error diagnostic at plan time.

#### `by_field` block (Field variant)

The `by_field` nested block represents a range slider sourced from a Kibana data view field:

- `data_view_id` (required, string) — the ID of the data view the slider targets.
- `field_name` (required, string) — the numeric field within the data view.
- `title` (optional, string) — human-readable label displayed above the slider.
- `use_global_filters` (optional, bool) — whether the panel respects dashboard-level global filters.
- `ignore_validations` (optional, bool) — suppresses validation errors from the control during intermediate states.
- `value` (optional, list of string) — the initial min/max range as a 2-element list `[min, max]`; the list MUST contain exactly 2 elements when set.
- `step` (optional, number) — the step size for each increment of the slider (stored as float32 to match the API).

On write, the provider SHALL set `values_source = "field"` automatically; this field SHALL NOT be exposed.

#### `by_esql` block (ES|QL variant)

The `by_esql` nested block represents a range slider sourced from an ES|QL query:

- `esql_query` (required, string) — the ES|QL query that produces the min/max range values.
- `values_source` (required, string) — must be `"esql_query"`. Any other value SHALL produce an error diagnostic at plan time.
- `title` (optional, string) — same as `by_field`.
- `use_global_filters` (optional, bool) — same as `by_field`.
- `ignore_validations` (optional, bool) — same as `by_field`.
- `value` (optional, list of string) — same as `by_field` (2-element list constraint applies).
- `step` (optional, number) — same as `by_field`.

Null-preservation semantics (REQ-009) apply to optional boolean attributes (`use_global_filters`, `ignore_validations`) within both branches. On import, the provider SHALL populate the branch-specific required identifiers from the API response: for `by_field`, `data_view_id` and `field_name`; for `by_esql`, `esql_query` and `values_source`. Optional booleans SHALL be left null.
#### Mutual exclusion and conflict guards

- Exactly one of `by_field` or `by_esql` MUST be set in `range_slider_control_config`.
- `range_slider_control_config` remains mutually exclusive with all other typed panel config blocks and with `config_json`.

#### State migration (v0 → v1)

The same `ResourceWithUpgradeState` upgrader (described in REQ-027 above) SHALL also rewrite existing v0 `range_slider_control_config` flat attributes (`data_view_id`, `field_name`, `title`, `use_global_filters`, `ignore_validations`, `value`, `step`) under a `by_field {}` object.

#### Scenarios

##### Scenario: Field variant round-trip

- GIVEN a `range_slider_control` panel with `by_field = { data_view_id = "orders-view", field_name = "price" }`
- WHEN the provider creates the dashboard and reads it back
- THEN `data_view_id` and `field_name` SHALL be present under `range_slider_control_config.by_field` and a subsequent plan SHALL show no changes

##### Scenario: ES|QL variant round-trip

- GIVEN a `range_slider_control` panel with `by_esql = { esql_query = "FROM orders | STATS ...", values_source = "esql_query" }`
- WHEN the provider creates the dashboard and reads it back
- THEN `esql_query` and `values_source` SHALL be present under `range_slider_control_config.by_esql` and a subsequent plan SHALL show no changes

##### Scenario: Invalid value list length

- GIVEN a `range_slider_control_config` block with `by_field = { ..., value = ["10"] }`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return a validation diagnostic stating that `value` must contain exactly 2 elements

##### Scenario: State upgrade from v0 flat to v1 by_field

- GIVEN a Terraform state containing `range_slider_control_config` with flat attributes (v0 schema: `data_view_id`, `field_name`, etc. at the config root)
- WHEN the provider with the updated schema is applied
- THEN the state upgrader SHALL rewrite the flat attributes under `by_field` and the resulting v1 state SHALL be equivalent to configuring `by_field { data_view_id = ..., field_name = ..., ... }`

##### Scenario: Both branches rejected

- GIVEN a `range_slider_control_config` block that sets both `by_field` and `by_esql`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL return an error diagnostic indicating the two branches are mutually exclusive

---

## ADDED Requirements

### Requirement: Dashboard resource schema version upgrade (REQ-040)

The `elasticstack_kibana_dashboard` resource SHALL implement `terraform-plugin-framework`'s `ResourceWithUpgradeState` interface with a single state upgrader for version 0 → 1.

The v0 → v1 upgrader SHALL:

1. Inspect every entry in `panels[]` and every entry in `pinned_panels[]`.
2. For each entry whose `type` is `"options_list_control"`: move all flat attributes from `options_list_control_config` (i.e. `data_view_id`, `field_name`, `title`, `use_global_filters`, `ignore_validations`, `single_select`, `exclude`, `exists_selected`, `run_past_timeout`, `search_technique`, `selected_options`, `display_settings`, `sort`) into a nested `by_field {}` object within `options_list_control_config`.
3. For each entry whose `type` is `"range_slider_control"`: move all flat attributes from `range_slider_control_config` (`data_view_id`, `field_name`, `title`, `use_global_filters`, `ignore_validations`, `value`, `step`) into a nested `by_field {}` object within `range_slider_control_config`.
4. Leave all other panel types unchanged.

The resource schema version SHALL be incremented to 1. No data SHALL be lost during the upgrade; the resulting state SHALL be functionally equivalent to the original state.

#### Scenario: State upgrade preserves all field-branch attributes

- GIVEN a v0 state containing both `options_list_control` and `range_slider_control` panels with all optional attributes set (e.g. `sort`, `display_settings`, `value`, `step`)
- WHEN the state upgrader runs
- THEN all attribute values SHALL be present under the `by_field {}` sub-object in v1 state and no attributes SHALL be dropped

#### Scenario: Non-control panels are unaffected by the upgrader

- GIVEN a v0 state containing a mix of `options_list_control`, `range_slider_control`, and `markdown` panels
- WHEN the state upgrader runs
- THEN the `markdown` panel entries SHALL be unchanged in v1 state
