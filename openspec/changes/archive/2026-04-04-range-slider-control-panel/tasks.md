## 1. Spec

- [x] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `npx openspec validate range-slider-control-panel --type change` (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, **sync** delta into `openspec/specs/kibana-dashboard/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [x] 2.1 Add `range_slider_control_config` single nested object block to `internal/kibana/dashboard/schema.go` with `data_view_id` and `field_name` as required string attributes, and `title`, `use_global_filters`, `ignore_validations`, `value`, and `step` as optional attributes; add a list-length validator on `value` enforcing exactly 2 elements.
- [x] 2.2 Add a `RangeSliderControlPanelModel` (or equivalent) struct to `internal/kibana/dashboard/models.go` (or a new `models_range_slider_control_panel.go`) with fields mapping to all `range_slider_control_config` attributes.
- [x] 2.3 Implement `toAPIModel` conversion for `range_slider_control_config`: populate the panel `config` object from the Terraform model, sending only non-null optional fields.
- [x] 2.4 Implement `populateFromAPI` / read-path conversion for `range_slider_control` panels: map API response config fields into `range_slider_control_config` state, preserving null for absent optional fields.
- [x] 2.5 Register the new panel type in the panel dispatcher used by `models_panels.go` so that `range_slider_control` panels are recognized on read-back and routed to the typed converter.
- [x] 2.6 Update REQ-006 validation (schema-level) to register `range_slider_control_config` as mutually exclusive with all other config blocks, valid only when `type = "range_slider_control"`.
- [x] 2.7 Update REQ-010 enforcement (write-path) to return an error diagnostic when `config_json` is used with `type = "range_slider_control"`.
- [x] 2.8 Update embedded descriptions (`descriptions/*.md` or inline schema descriptions) and regenerate docs if applicable.

## 3. Testing

- [x] 3.1 Add an acceptance test for `range_slider_control` panels that creates a dashboard with required fields only (`data_view_id`, `field_name`), asserts successful apply, and verifies state.
- [x] 3.2 Add an acceptance test step (or separate test) that configures all optional fields (`title`, `use_global_filters`, `ignore_validations`, `value`, `step`) and asserts round-trip fidelity in state.
- [x] 3.3 Add a unit or acceptance test that configures `value` with fewer or more than 2 elements and asserts a plan-time validation error.
- [x] 3.4 Add a unit test or acceptance test that verifies using `config_json` with `type = "range_slider_control"` returns an error diagnostic.
- [x] 3.5 Verify that an existing dashboard state without any `range_slider_control` panels is unaffected by the schema change (no spurious plan diff).
