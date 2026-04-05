# Tasks: Dashboard API schema alignment

## 1. Spec

- [x] 1.1 Keep `openspec/changes/dashboard-api-schema-alignment/specs/kibana-dashboard/spec.md` aligned with `proposal.md` and `design.md`
- [x] 1.2 After implementation, run `openspec validate --all` (or `make check-openspec`) and sync delta into `openspec/specs/kibana-dashboard/spec.md` or archive the change

## 2. Root schema and models (`internal/kibana/dashboard`)

- [x] 2.1 Replace flat time attributes with nested `time_range` block in schema and `dashboardModel`; update `populateFromAPI`, `toAPICreateRequest`, `toAPIUpdateRequest`
- [x] 2.2 Replace flat refresh attributes with nested `refresh_interval` block in schema and model; update create/update/read mapping
- [x] 2.3 Replace `query_language`, `query_text`, and `query_json` with nested `query` block: `language`, `text`, `json` (mutually exclusive `text` vs `json`); update `queryToAPI` and read path
- [x] 2.4 Update REQ-009-related preservation logic comments and tests to reference `time_range.mode` path
- [x] 2.5 Extend `options` schema and `optionsModel` with `auto_apply_filters` and `hide_panel_borders`; update `toAPI`, `mapOptionsFromAPI`, and default-detection helper(s)

## 3. Panels and sections

- [x] 3.1 Rename panel and section identity attribute from `id` to `uid` in schema, `panelModel`, `sectionModel`, and all `tfsdk` tags; update `mapPanelsFromAPI` / `panelsToAPI` field references
- [x] 3.2 Grep resource tests and testdata for `.id` under `panels` / `sections` and update to `uid`

## 4. Pie chart and JSON naming

- [x] 4.1 Rename pie chart `dataset` → `dataset_json` and `legend` → `legend_json` in schema and `pieChartConfigModel`
- [x] 4.2 Audit `jsontypes.Normalized` (and related JSON custom types) in `internal/kibana/dashboard` for missing `*_json` suffix; rename to match convention

## 5. Heatmap

- [x] 5.1 Change `heatmap_config.legend.visibility` from bool to string enum (`visible` | `hidden`); update `heatmapLegendModel` to/from API `HeatmapLegendVisibility`
- [x] 5.2 Update unit and acceptance tests for heatmap legend

## 6. Documentation and examples

- [x] 6.1 Regenerate or update resource documentation for `elasticstack_kibana_dashboard` (provider docs workflow per `dev-docs/high-level/documentation.md`)
- [x] 6.2 Update all `internal/kibana/dashboard/testdata/**` and example snippets referencing renamed attributes

## 7. Requirements verification

- [x] 7.1 Run targeted `go test` for `internal/kibana/dashboard` packages
- [ ] 7.2 Run acceptance tests for dashboard when Elastic stack is available (per `dev-docs/high-level/testing.md`). Leave unchecked until a maintainer has executed a successful targeted or full dashboard acc run against a real stack (CI or local).
