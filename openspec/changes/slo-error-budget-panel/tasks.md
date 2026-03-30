## 1. Spec

- [x] 1.1 Keep delta spec aligned with proposal.md / design.md
- [x] 1.2 On completion, sync delta into canonical spec or archive

## 2. Schema

- [x] 2.1 Add `slo_error_budget_config` schema block to `schema.go` with `slo_id` (required string), `slo_instance_id` (optional string), `title` (optional string), `description` (optional string), `hide_title` (optional bool), `hide_border` (optional bool), and `drilldowns` (optional list of typed objects)
- [x] 2.2 Add `drilldowns` nested schema with required `url`, `label` and optional `encode_url`, `open_in_new_tab`; hardcode Kibana's fixed `trigger` and `type` values in the converter
- [x] 2.3 Add schema-level validation that `slo_error_budget_config` is only used with `type = "slo_error_budget"` and conflicts with all other typed config blocks and `config_json`

## 3. Models and converters

- [x] 3.1 Extend `panelModel` struct in `models_panels.go` to carry the new `SloErrorBudgetConfig` field
- [x] 3.2 Route `slo_error_budget` through the panel dispatcher in `models_panels.go`
- [x] 3.3 Create `models_slo_error_budget_panel.go` with a converter implementing the read and write paths
- [x] 3.4 Implement `slo_instance_id` null-preservation: on read, only write the API-returned value if prior state/plan had a non-null `slo_instance_id`
- [x] 3.5 Implement `encode_url` / `open_in_new_tab` default normalization on read (treat API-returned `true` as matching an omitted optional bool)
- [x] 3.6 Keep drilldown conversion local to the SLO error budget panel implementation and hardcode the API constants there

## 4. Testing

- [x] 4.1 Add acceptance test for creating an `slo_error_budget` panel with only `slo_id` configured
- [x] 4.2 Add acceptance test verifying `slo_instance_id` null-preservation: omit the field in config and confirm no drift after apply
- [x] 4.3 Add acceptance test for configuring one or more `drilldowns` entries and reading them back
- [x] 4.4 Add unit tests for the converter covering the write path (TF model -> API request) and the read path (API response -> TF model), including the `slo_instance_id` null-preservation logic
