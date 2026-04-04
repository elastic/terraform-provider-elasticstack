## Why

Practitioners cannot manage SLO error budget panels as code today. These panels display a burn chart of the remaining error budget for a specific SLO over a reporting period, giving teams a clear visual indicator of how much of their reliability target has been consumed. Without Terraform support, dashboards that include SLO error budget panels must be partially managed outside of Terraform, which breaks reproducible infrastructure and prevents full dashboard-as-code workflows for reliability-focused teams.

## What Changes

- Add a typed `slo_error_budget_config` schema block to the `elasticstack_kibana_dashboard` panels schema, valid only when `type = "slo_error_budget"`.
- Add a converter in a new `models_slo_error_budget_panel.go` that maps between the Terraform state model and the Kibana dashboard panel API shape for the `slo-error-budget-embeddable` embeddable type.
- Extend the `panelModel` struct in `models_panels.go` to carry the new config block and route it through the panel dispatcher.
- Add schema validation enforcing that `slo_error_budget_config` is only present on `type = "slo_error_budget"` panels, and that no other typed config block is present on such panels.
- Add acceptance tests for the full panel lifecycle and unit tests for the converter.

## Capabilities

### New Capabilities

- `kibana-dashboard`: practitioners can declare an `slo_error_budget` panel with a typed `slo_error_budget_config` block, including required `slo_id`, optional `slo_instance_id` (API default `"*"`), optional `drilldowns` as a list of typed objects with `url` and `label`, and optional display fields `title`, `description`, `hide_title`, and `hide_border`.

### Modified Capabilities

- _(none)_

## Impact

- Specs: delta spec under `openspec/changes/slo-error-budget-panel/specs/kibana-dashboard/spec.md`.
- Schema: `internal/kibana/dashboard/schema.go`.
- Models: `internal/kibana/dashboard/models_panels.go` and new `internal/kibana/dashboard/models_slo_error_budget_panel.go`.
- Tests: new acceptance tests in `internal/kibana/dashboard/acc_test.go` (or a dedicated file) and unit tests alongside the converter.
