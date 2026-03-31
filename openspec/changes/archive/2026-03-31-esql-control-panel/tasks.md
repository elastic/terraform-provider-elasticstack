# Tasks: ES|QL Control Panel Support

## 1. Spec

- [x] 1.1 Keep delta spec aligned with proposal.md / design.md
- [ ] 1.2 On completion, sync delta into canonical spec or archive

## 2. Implementation

- [x] 2.1 Add `esql_control_config` schema block to `internal/kibana/dashboard/schema.go`
- [x] 2.2 Extend `panelModel` struct in `models_panels.go` with `EsqlControlConfig` field
- [x] 2.3 Create `models_esql_control_panel.go` with read and write converter functions
- [x] 2.4 Update the panel write-path dispatcher in `models_panels.go` to handle `esql_control` type via the typed config block
- [x] 2.5 Update the panel read-path dispatcher in `models_panels.go` to populate `esql_control_config` on read-back
- [x] 2.6 Add schema validation that `esql_control_config` is only valid with `type = "esql_control"` (REQ-006 extension)
- [x] 2.7 Add schema validators for `variable_type` enum (`fields`, `values`, `functions`, `time_literal`, `multi_values`) and `control_type` enum (`STATIC_VALUES`, `VALUES_FROM_QUERY`)
- [x] 2.8 Update `config_json` write-path error message in `models_panels.go` to explicitly name `esql_control` as unsupported (REQ-010 update)
- [x] 2.9 Update resource descriptions and documentation for the new block and its attributes

## 3. Testing

- [x] 3.1 Add acceptance tests for `esql_control` panel creation with required fields (`STATIC_VALUES` control type)
- [x] 3.2 Add acceptance tests for `esql_control` panel creation with `VALUES_FROM_QUERY` control type and `esql_query`
- [x] 3.3 Add acceptance tests for `esql_control` panel with optional `display_settings` block
- [x] 3.4 Add acceptance tests for plan-time validation rejection of invalid `variable_type` and `control_type` enum values
- [x] 3.5 Add unit tests for the `esql_control` panel write converter (Terraform model to API payload)
- [x] 3.6 Add unit tests for the `esql_control` panel read converter (API payload to Terraform model)
- [x] 3.7 Verify that setting `config_json` on a panel with `type = "esql_control"` returns a plan-time or apply-time error diagnostic
