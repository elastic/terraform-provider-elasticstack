# Tasks: `lens-dashboard-app` Panel Support

## 1. Spec

- [x] 1.1 Keep delta spec aligned with proposal.md / design.md
- [x] 1.2 On completion, sync delta into canonical spec or archive

## 2. Schema

- [x] 2.1 Add `lens_dashboard_app_config` schema block to `internal/kibana/dashboard/schema.go` with `by_value` and `by_reference` sub-blocks
- [x] 2.2 Add required `by_reference.time_range` nested block with required `from` / `to` and optional `mode`
- [x] 2.3 Add schema validators to enforce that exactly one of `by_value` or `by_reference` is set within `lens_dashboard_app_config`
- [x] 2.4 Add schema validation that `lens_dashboard_app_config` is only valid when `type = "lens-dashboard-app"` (REQ-006 extension)
- [x] 2.5 Add schema validation that `lens_dashboard_app_config` is mutually exclusive with all other panel config blocks (REQ-006 extension)
- [x] 2.6 Add `by_value.config_json` as a required normalized JSON string
- [x] 2.7 Add by-reference attributes: required `ref_id`, optional `references_json`, optional `title`, optional `description`, optional `hide_title`, optional `hide_border`, and optional `drilldowns_json`

## 3. Models

- [ ] 3.1 Extend `panelModel` struct in `models_panels.go` with a `LensDashboardAppConfig` field
- [ ] 3.2 Create `models_lens_dashboard_app_panel.go` with model structs for `lensDashboardAppConfigModel`, `lensDashboardAppByValueModel`, `lensDashboardAppByReferenceModel`, and `lensDashboardAppTimeRangeModel`

## 4. Converters

- [ ] 4.1 Implement write converter for `by_value` mode by decoding `by_value.config_json` and assigning it directly to `KbnDashboardPanelTypeLensDashboardApp.Config`
- [ ] 4.2 Implement write converter for `by_reference` mode using `KbnDashboardPanelTypeLensDashboardAppConfig1` (`ref_id`, required `time_range`, optional `references`, display fields, and `drilldowns`)
- [ ] 4.3 Implement read converter (API payload to Terraform model) with mode detection based on the generated `config` union, preferring by-reference only when `ref_id` and `time_range` are present
- [ ] 4.4 Implement read converter population of by-reference optional fields (`references_json`, `title`, `description`, `hide_title`, `hide_border`, `drilldowns_json`)
- [ ] 4.5 Update the panel write-path dispatcher in `models_panels.go` to handle `lens-dashboard-app` type via `lens_dashboard_app_config`
- [ ] 4.6 Update the panel read-path dispatcher in `models_panels.go` to populate `lens_dashboard_app_config` on read-back

## 5. Validation

- [ ] 5.1 Update panel-level `config_json` write-path error message in `models_panels.go` to explicitly name `lens-dashboard-app` as unsupported (REQ-025 update)
- [ ] 5.2 Add validator or plan modifier to enforce mutual exclusivity of `by_value` and `by_reference` sub-blocks at plan time
- [ ] 5.3 Validate `by_reference.time_range.mode` accepts only `absolute` or `relative` when set
- [ ] 5.4 Update resource descriptions and documentation for the new block and its attributes

## 6. Testing

- [ ] 6.1 Add acceptance tests for `lens-dashboard-app` panel creation in by-reference mode with required `ref_id` and `time_range`; include coverage that sets optional `references_json` for a typical saved-object reference-wiring case (see REQ-035)
- [ ] 6.2 Add acceptance tests for `lens-dashboard-app` panel creation in by-value mode with required `config_json`
- [ ] 6.3 Add acceptance tests for by-reference panel with optional `title`, `description`, `hide_title`, and `hide_border`
- [ ] 6.4 Add acceptance or unit coverage for by-reference `drilldowns_json`
- [ ] 6.5 Add acceptance tests for by-reference `time_range.mode`
- [ ] 6.6 Add acceptance tests for plan-time validation rejection when both `by_value` and `by_reference` are set simultaneously
- [ ] 6.7 Add acceptance tests for plan-time validation rejection when neither `by_value` nor `by_reference` is set
- [ ] 6.8 Add unit tests for the `by_value` write converter ensuring `config_json` is sent directly as API `config`
- [ ] 6.9 Add unit tests for the `by_reference` write converter ensuring `ref_id`, `references`, required `time_range`, display fields, and `drilldowns` map to API `config`
- [ ] 6.10 Add unit tests for the read converter mode detection and field population (by-value path)
- [ ] 6.11 Add unit tests for the read converter mode detection and field population (by-reference path)
- [ ] 6.12 Verify that setting `config_json` on a panel with `type = "lens-dashboard-app"` returns an error diagnostic
