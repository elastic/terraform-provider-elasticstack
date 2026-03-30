## 1. Spec

- [ ] 1.1 Keep delta spec aligned with proposal.md / design.md
- [ ] 1.2 On completion, sync delta into canonical spec or archive

## 2. Schema

- [ ] 2.1 Add `slo_error_budget_config` schema block to `schema.go` with `slo_id` (required string), `slo_instance_id` (optional string), `title` (optional string), `description` (optional string), `hide_title` (optional bool), `hide_border` (optional bool), and `drilldowns` (optional list of typed objects)
- [ ] 2.2 Add `drilldowns` nested schema with required `url`, `label`, `trigger`, `type` and optional `encode_url`, `open_in_new_tab`; add enum validators for `trigger` and `type`
- [ ] 2.3 Add schema-level validation that `slo_error_budget_config` is only used with `type = "slo_error_budget"` and conflicts with all other typed config blocks and `config_json`

## 3. Models and converters

- [ ] 3.1 Extend `panelModel` struct in `models_panels.go` to carry the new `SloErrorBudgetConfig` field
- [ ] 3.2 Route `slo_error_budget` through the panel dispatcher in `models_panels.go`
- [ ] 3.3 Create `models_slo_error_budget_panel.go` with a converter implementing the read and write paths
- [ ] 3.4 Implement `slo_instance_id` null-preservation: on read, only write the API-returned value if prior state/plan had a non-null `slo_instance_id`
- [ ] 3.5 Implement `encode_url` / `open_in_new_tab` default normalization on read (treat API-returned `true` as matching an omitted optional bool)
- [ ] 3.6 Reuse or extract the shared SLO drilldown converter function to avoid duplicating logic already present for `slo_overview` and `slo_burn_rate` panels

## 4. Testing

- [ ] 4.1 Add acceptance test for creating an `slo_error_budget` panel with only `slo_id` configured
- [ ] 4.2 Add acceptance test verifying `slo_instance_id` null-preservation: omit the field in config and confirm no drift after apply
- [ ] 4.3 Add acceptance test for configuring one or more `drilldowns` entries and reading them back
- [ ] 4.4 Add unit tests for the converter covering the write path (TF model -> API request) and the read path (API response -> TF model), including the `slo_instance_id` null-preservation logic
