# Tasks: ES|QL Control Panel Support

## 1. Spec

- [ ] 1.1 Keep delta spec aligned with proposal.md / design.md
- [ ] 1.2 On completion, sync delta into canonical spec or archive

## 2. Implementation

- [ ] 2.1 Add `esql_control_config` schema block to `internal/kibana/dashboard/schema.go`
- [ ] 2.2 Extend `panelModel` struct in `models_panels.go` with `EsqlControlConfig` field
- [ ] 2.3 Create `models_esql_control_panel.go` with read and write converter functions
- [ ] 2.4 Update the panel write-path dispatcher in `models_panels.go` to handle `esql_control` type via the typed config block
- [ ] 2.5 Update the panel read-path dispatcher in `models_panels.go` to populate `esql_control_config` on read-back
- [ ] 2.6 Add schema validation that `esql_control_config` is only valid with `type = "esql_control"` (REQ-006 extension)
- [ ] 2.7 Add schema validators for `variable_type` enum (`fields`, `values`, `functions`, `time_literal`, `multi_values`) and `control_type` enum (`STATIC_VALUES`, `VALUES_FROM_QUERY`)
- [ ] 2.8 Update `config_json` write-path error message in `models_panels.go` to explicitly name `esql_control` as unsupported (REQ-010 update)
- [ ] 2.9 Update resource descriptions and documentation for the new block and its attributes

## 3. Testing

- [ ] 3.1 Add acceptance tests for `esql_control` panel creation with required fields (`STATIC_VALUES` control type)
- [ ] 3.2 Add acceptance tests for `esql_control` panel creation with `VALUES_FROM_QUERY` control type and `esql_query`
- [ ] 3.3 Add acceptance tests for `esql_control` panel with optional `display_settings` block
- [ ] 3.4 Add acceptance tests for plan-time validation rejection of invalid `variable_type` and `control_type` enum values
- [ ] 3.5 Add unit tests for the `esql_control` panel write converter (Terraform model to API payload)
- [ ] 3.6 Add unit tests for the `esql_control` panel read converter (API payload to Terraform model)
- [ ] 3.7 Verify that setting `config_json` on a panel with `type = "esql_control"` returns a plan-time or apply-time error diagnostic
