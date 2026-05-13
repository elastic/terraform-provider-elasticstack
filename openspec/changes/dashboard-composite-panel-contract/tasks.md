## 1. Create visconfig handler

- [ ] 1.1 Create `dashboard/panel/visconfig/api.go` with `Handler` implementing `iface.Handler` for panel type `"vis"`, block `"vis_config"`
- [ ] 1.2 Implement `FromAPI`: classify config JSON (by_value chart, by_reference, or config_json-only), delegate chart population to `lenscommon` converter, use `lenscommon.ByReferenceFromAPI` for by_reference
- [ ] 1.3 Implement `ToAPI`: delegate chart building to `lenscommon` converter for by_value, use `lenscommon.ByReferenceToAPI` for by_reference, handle config_json-only path
- [ ] 1.4 Create `dashboard/panel/visconfig/schema.go` with `vis_config` block replacing old `viz_config` block
- [ ] 1.5 Create `dashboard/panel/visconfig/model.go` with config classification helpers
- [ ] 1.6 Create `dashboard/panel/visconfig/api_test.go` covering by_value (each chart kind), by_reference, config_json fallback
- [ ] 1.7 Register `visconfig.Handler{}` in `dashboard/registry.go`

## 2. Create lensdashboardapp handler

- [ ] 2.1 Create `dashboard/panel/lensdashboardapp/api.go` with `Handler` implementing `iface.Handler` for panel type `"lens_dashboard_app"`
- [ ] 2.2 Implement `FromAPI`: classify config JSON, delegate to `lenscommon` converter or `lenscommon.ByReferenceFromAPI`
- [ ] 2.3 Implement `ToAPI`: delegate to `lenscommon` converter or `lenscommon.ByReferenceToAPI`
- [ ] 2.4 Create `dashboard/panel/lensdashboardapp/schema.go` with `lens_dashboard_app_config` block
- [ ] 2.5 Create `dashboard/panel/lensdashboardapp/model.go` with config classification helpers
- [ ] 2.6 Create `dashboard/panel/lensdashboardapp/api_test.go` covering by_value and by_reference paths
- [ ] 2.7 Register `lensdashboardapp.Handler{}` in `dashboard/registry.go`

## 3. Create discoversession handler

- [ ] 3.1 Create `dashboard/panel/discoversession/api.go` with `Handler` implementing `iface.Handler` for panel type `"discover_session"`, block `"discover_session_config"`
- [ ] 3.2 Implement `FromAPI`: classify config (by_value DSL tab, by_value ESQL tab, or by_reference); populate appropriate sub-model
- [ ] 3.3 Implement `ToAPI`: branch on by_value tab type or by_reference; build API payload accordingly
- [ ] 3.4 Create `dashboard/panel/discoversession/schema.go` with `discover_session_config` block including `by_value` (with `tab` DSL/ESQL dispatch) and `by_reference` (with overrides) sub-blocks
- [ ] 3.5 Create `dashboard/panel/discoversession/model.go` with config classification helpers and tab-type detection
- [ ] 3.6 Create `dashboard/panel/discoversession/api_test.go` covering by_value DSL path, by_value ESQL path, by_reference path, and null-preservation for optional fields
- [ ] 3.7 Register `discoversession.Handler{}` in `dashboard/registry.go`

## 4. vis → vis_config rename

- [ ] 4.1 Update `schema.go`: rename `viz_config` attribute to `vis_config` everywhere
- [ ] 4.2 Update `models/lens.go`: rename `VizConfig` struct field to `VisConfig` with `tfsdk:"vis_config"`
- [ ] 4.3 Update all references in dashboard package from `VizConfig` to `VisConfig`
- [ ] 4.4 Update constant `panelTypeVis` comment/docs if needed
- [ ] 4.5 Update `models_plan_state_alignment.go` references from `VizConfig` to `VisConfig`
- [ ] 4.6 Update `panel_config_validator.go` references
- [ ] 4.7 Update acceptance tests using `viz_config` to `vis_config`
- [ ] 4.8 Update example Terraform files under `examples/resources/elasticstack_kibana_dashboard/`
- [ ] 4.9 Update generated documentation

## 5. Final cleanup of central files

- [ ] 5.1 Strip `models_panels.go` to unknown-panel fallback and section helpers only; delete all switch/case and cascading if/else
- [ ] 5.2 Delete `models_viz_config.go` (absorbed into `panel/visconfig/`)
- [ ] 5.3 Delete `models_vis_api.go` (absorbed into `lenscommon/by_reference.go`)
- [ ] 5.4 Delete `models_lens_dashboard_app_converters.go`, `models_lens_dashboard_app_by_value_adapter.go`, `models_lens_dashboard_app_panel.go` (all absorbed)
- [ ] 5.5 Delete `models_discover_session_panel.go` and `schema_discover_session_panel.go` (absorbed into `panel/discoversession/`)
- [ ] 5.6 Refactor `panel_config_validator.go`: remove all remaining hard-coded panel type cases; keep only registry iteration loop and pinned panel logic
- [ ] 5.7 Refactor `panel_config_defaults.go`: remove all hard-coded lens chart dispatch; keep only top-level delegation to handler and converter registries
- [ ] 5.8 Refactor `schema.go`: remove `getLensDashboardAppByValueNestedAttributes()` and `getVizByValueAttributes()`; assemble lens chart attributes from `lenscommon.All()`
- [ ] 5.9 Remove any orphaned imports across all dashboard files
- [ ] 5.10 Run `goimports` and `gofmt` across all modified files

## 6. Verification

- [ ] 6.1 `go build ./internal/kibana/dashboard/...` passes
- [ ] 6.2 `go vet ./...` passes
- [ ] 6.3 `go test ./internal/kibana/dashboard/...` passes (all unit tests)
- [ ] 6.4 All dashboard acceptance tests pass, including:
  - vis/vis_config panels (all 12 chart kinds)
  - lens_dashboard_app panels (by_value and by_reference)
  - discover_session panels (by_value DSL, by_value ESQL, by_reference)
  - markdown, slo, synthetics, controls (regression check)
- [ ] 6.5 `make build` passes
- [ ] 6.6 No dead code: confirm via static analysis or manual review that no unreferenced functions remain

## 7. Documentation

- [ ] 7.1 Update resource documentation to reflect `vis_config` rename
- [ ] 7.2 Update CHANGELOG with breaking change note: `viz_config` → `vis_config`
- [ ] 7.3 Verify example configurations compile and match new schema
