## 1. Spec

- [ ] 1.1 Keep delta spec aligned with proposal.md / design.md
- [ ] 1.2 On completion, sync delta into canonical spec or archive

## 2. kbapi generator cleanup

- [ ] 2.1 Delete `generated/kbapi/dashboards.json`
- [ ] 2.2 In `generated/kbapi/transform_schema.go`: remove `mergeDashboardsSchema` from the `transformers` slice, and remove the `//go:embed dashboards.json` directive and `var dashboardsJSON string` variable
- [ ] 2.3 In `generated/kbapi/transform_schema.go`: remove the `mergeDashboardsSchema` function body
- [ ] 2.4 In `generated/kbapi/transform_schema.go`: remove `fixVisualizationIdParam` from the `transformers` slice and delete its function body (the `/api/visualizations/{id}` path was provided exclusively by `dashboards.json`)
- [ ] 2.5 Run `make -C generated/kbapi transform` and inspect `oas-filtered.yaml` to confirm the exact `Kibana_HTTP_APIs_kbn-dashboard-*` schema key names used by `fixDashboardPanelItemRefs`
- [ ] 2.6 In `generated/kbapi/transform_schema.go`: update `fixDashboardPanelItemRefs` to use the upstream-native `Kibana_HTTP_APIs_kbn-dashboard-data` and `Kibana_HTTP_APIs_kbn-dashboard-section` key names (and any other renamed dashboard schema keys confirmed in 2.5)
- [ ] 2.7 Update `panelTypePrefix` logic in `transform_schema.go` if upstream panel-type schemas are now named `Kibana_HTTP_APIs_kbn-dashboard-panel-type-*` (confirm by inspecting `oas-filtered.yaml`)
- [ ] 2.8 Run `make -C generated/kbapi generate` and verify `kibana.gen.go` compiles cleanly; confirm that `KbnDashboardPanelTypeLensDashboardApp` no longer appears in the output

## 3. Provider resource cleanup

- [ ] 3.1 Delete `internal/kibana/dashboard/panel/lensdashboardapp/` package entirely
- [ ] 3.2 Remove `LensDashboardAppConfig *LensDashboardAppConfigModel` from `PanelModel` in `internal/kibana/dashboard/models/panel.go`; remove `LensDashboardAppConfigModel` and its sub-types from `internal/kibana/dashboard/models/lens.go`
- [ ] 3.3 Deregister `lensdashboardapp.Handler{}` from `internal/kibana/dashboard/registry.go`; remove the `panelTypeAliases` entry for `"lens-dashboard-app"`
- [ ] 3.4 Remove `lens_dashboard_app_config` schema block from `internal/kibana/dashboard/schema.go`; remove the `panelTypeLensDashboardApp` constant
- [ ] 3.5 Update all provider usages of renamed generated types from `kibana.gen.go` (e.g., any references to `KbnDashboardData`, `KbnDashboardSection`, or other dashboard panel type constants that have been renamed under the `Kibana_HTTP_APIs_` prefix)
- [ ] 3.6 Verify the existing unknown-panel fallback in `dashboardMapPanelFromAPI` is intact — this is the read-time safety net for existing Kibana dashboards that still have `lens-dashboard-app` panels at the API level
- [ ] 3.7 Ensure the project builds cleanly: `make build`

## 4. Tests, docs, and spec

- [ ] 4.1 Remove or update acceptance tests covering `lens-dashboard-app` panels: `acc_lens_dashboard_app_panels_test.go`, relevant cases in `acc_drilldowns_test.go`, and `lens_by_value_embed_wiring_test.go`
- [ ] 4.2 Update `openspec/specs/kibana-dashboard/spec.md`: remove REQ-035 entirely, drop all occurrences of `lens-dashboard-app` (including in the schema overview, REQ-006 validation rules, REQ-010, REQ-025, REQ-040, and the implementation cross-reference table)
- [ ] 4.3 Remove `lens-dashboard-app` examples from `examples/resources/elasticstack_kibana_dashboard/resource.tf`
- [ ] 4.4 Write an upgrade guide section documenting the migration from `type = "lens-dashboard-app"` to `type = "vis"`, including the `config_json` attribute relocation note for `by_value.config_json` users (see `design.md` for the full before/after HCL examples)
