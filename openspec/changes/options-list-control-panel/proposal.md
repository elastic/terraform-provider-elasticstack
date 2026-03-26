## Why

Practitioners cannot manage options list control panels as code today. These controls provide a dropdown or multi-select filtering interface for Kibana dashboards based on a specific field in a data view, and they are essential for building interactive dashboards that let users narrow results without editing the underlying query. Because the panel type is not yet supported by the provider, dashboards that use options list controls must be managed partially outside Terraform, which breaks reproducible infrastructure and undermines full dashboard-as-code workflows.

## What Changes

- Add a typed `options_list_control_config` schema block to the `elasticstack_kibana_dashboard` panels schema, valid only when `type = "options_list_control"`.
- Add a converter in a new `models_options_list_control_panel.go` that maps between the Terraform state model and the Kibana dashboard panel API shape for `options_list_control`.
- Extend the `panelModel` struct in `models_panels.go` to carry the new config block and route it through the panel dispatcher.
- Add schema validation enforcing that `options_list_control_config` is only present on `type = "options_list_control"` panels, and that no other typed config block is present on such panels.
- Add acceptance tests for the full panel lifecycle and unit tests for the converter.

## Capabilities

### New Capabilities

- `kibana-dashboard`: practitioners can declare an `options_list_control` panel with typed attributes, including required `data_view_id` and `field_name`, optional display settings, sort configuration, `search_technique` enum, and `selected_options`.

### Modified Capabilities

- _(none)_

## Impact

- Specs: delta spec under `openspec/changes/options-list-control-panel/specs/kibana-dashboard/spec.md`.
- Schema: `internal/kibana/dashboard/schema.go`.
- Models: `internal/kibana/dashboard/models_panels.go` and new `internal/kibana/dashboard/models_options_list_control_panel.go`.
- Tests: new acceptance tests in `internal/kibana/dashboard/acc_test.go` (or a dedicated file) and unit tests alongside the converter.
