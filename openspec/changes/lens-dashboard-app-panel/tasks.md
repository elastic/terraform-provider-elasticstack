# Tasks: `lens-dashboard-app` Panel Support

## 1. Spec

- [ ] 1.1 Keep delta spec aligned with proposal.md / design.md
- [ ] 1.2 On completion, sync delta into canonical spec or archive

## 2. Schema

- [ ] 2.1 Add `lens_dashboard_app_config` schema block to `internal/kibana/dashboard/schema.go` with `by_value` and `by_reference` sub-blocks and shared optional fields
- [ ] 2.2 Add `time_range` nested block (with required `from` and `to`) within `lens_dashboard_app_config` in the schema
- [ ] 2.3 Add schema validators to enforce that exactly one of `by_value` or `by_reference` is set within `lens_dashboard_app_config`
- [ ] 2.4 Add schema validation that `lens_dashboard_app_config` is only valid when `type = "lens-dashboard-app"` (REQ-006 extension)
- [ ] 2.5 Add schema validation that `lens_dashboard_app_config` is mutually exclusive with all other panel config blocks (REQ-006 extension)

## 3. Models

- [ ] 3.1 Extend `panelModel` struct in `models_panels.go` with a `LensDashboardAppConfig` field
- [ ] 3.2 Create `models_lens_dashboard_app_panel.go` with model structs for `lensDashboardAppConfigModel`, `lensDashboardAppByValueModel`, `lensDashboardAppByReferenceModel`, and `lensDashboardAppTimeRangeModel`

## 4. Converters

- [ ] 4.1 Implement write converter (Terraform model to API payload) for `by_value` mode in `models_lens_dashboard_app_panel.go`
- [ ] 4.2 Implement write converter for `by_reference` mode in `models_lens_dashboard_app_panel.go`
- [ ] 4.3 Implement read converter (API payload to Terraform model) with mode detection (presence of `attributes` vs `saved_object_id`) in `models_lens_dashboard_app_panel.go`
- [ ] 4.4 Implement read converter population of shared optional fields (`title`, `description`, `hide_title`, `hide_border`, `time_range`)
- [ ] 4.5 Update the panel write-path dispatcher in `models_panels.go` to handle `lens-dashboard-app` type via `lens_dashboard_app_config`
- [ ] 4.6 Update the panel read-path dispatcher in `models_panels.go` to populate `lens_dashboard_app_config` on read-back

## 5. Validation

- [ ] 5.1 Update `config_json` write-path error message in `models_panels.go` to explicitly name `lens-dashboard-app` as unsupported (REQ-025 update)
- [ ] 5.2 Add validator or plan modifier to enforce mutual exclusivity of `by_value` and `by_reference` sub-blocks at plan time
- [ ] 5.3 Update resource descriptions and documentation for the new block and its attributes

## 6. Testing

- [ ] 6.1 Add acceptance tests for `lens-dashboard-app` panel creation in by-reference mode (required `saved_object_id`)
- [ ] 6.2 Add acceptance tests for `lens-dashboard-app` panel creation in by-value mode (required `attributes_json`)
- [ ] 6.3 Add acceptance tests for by-reference panel with optional `title`, `description`, `hide_title`, `hide_border` overrides
- [ ] 6.4 Add acceptance tests for by-value panel with optional `references_json`
- [ ] 6.5 Add acceptance tests for either mode with optional `time_range` block
- [ ] 6.6 Add acceptance tests for plan-time validation rejection when both `by_value` and `by_reference` are set simultaneously
- [ ] 6.7 Add acceptance tests for plan-time validation rejection when neither `by_value` nor `by_reference` is set
- [ ] 6.8 Add unit tests for the `by_value` write converter (Terraform model to API payload)
- [ ] 6.9 Add unit tests for the `by_reference` write converter (Terraform model to API payload)
- [ ] 6.10 Add unit tests for the read converter mode detection and field population (by-value path)
- [ ] 6.11 Add unit tests for the read converter mode detection and field population (by-reference path)
- [ ] 6.12 Verify that setting `config_json` on a panel with `type = "lens-dashboard-app"` returns an error diagnostic
