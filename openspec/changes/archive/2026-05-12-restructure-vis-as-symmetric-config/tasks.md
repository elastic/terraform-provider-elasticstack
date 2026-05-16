## 1. Refactor `lens_dashboard_app_config` helpers (no behavior change)

- [x] 1.1 Extract `getLensByReferenceAttributes()` from `getLensDashboardAppConfigSchema()` in `internal/kibana/dashboard/schema.go`; the helper returns the by-reference attribute map (`ref_id`, `references_json`, `title`, `description`, `hide_title`, `hide_border`, structured `drilldowns` placeholder, required `time_range`).
- [x] 1.2 Rename `getLensDashboardAppByValueAttributes()` callers; keep behavior identical pending step 4 (still includes its own `config_json` and 11 chart blocks; legacy_metric stays out).
- [x] 1.3 Run `go build ./...` and existing dashboard unit tests to confirm zero behavior delta from helper extraction.

## 2. Implement structured 3-way `drilldowns` shape (REQ-041)

- [x] 2.1 Add `getStructuredDrilldownsAttribute()` returning a `ListNestedAttribute` whose item object exposes `dashboard`, `discover`, and `url` `SingleNestedAttribute` sub-blocks; wire all per-variant attributes per the design table (`dashboard_id`/`label`/`use_filters`/`use_time_range`/`open_in_new_tab` for dashboard; `label`/`open_in_new_tab` for discover; `url`/`label`/`trigger`/`encode_url`/`open_in_new_tab` for url).
- [x] 2.2 Add `drilldownItemModeValidator` enforcing exactly one of `dashboard`/`discover`/`url` per item; cover both zero-set and multi-set cases.
- [x] 2.3 Add `OneOf` validator for `url.trigger` (∈ `on_click_row`/`on_click_value`/`on_open_panel_menu`/`on_select_range`).
- [x] 2.4 Add `drilldownsModel` Go struct (list of items with three nullable sub-block models) and `fromAPI(ctx, []kbapi.<DrilldownUnion>) ([]drilldownItemModel, diag.Diagnostics)` + `toAPI([]drilldownItemModel) ([]kbapi.<DrilldownUnion>, diag.Diagnostics)` helpers in a new `internal/kibana/dashboard/models_drilldowns.go`.
- [x] 2.5 Implement read-side classification: detect API drilldown `type` field, dispatch to `dashboard`/`discover`/`url` sub-block model. Surface a clear error diagnostic when an API drilldown cannot be losslessly represented in any of the three sub-blocks.
- [x] 2.6 Implement write-side: emit `trigger` (constant for dashboard/discover; required URL trigger pass-through—Terraform schema requires `url.trigger`) and `type` (constant per variant) automatically; only include practitioner-set optional fields otherwise.
- [x] 2.7 Add unit tests for `models_drilldowns.go`: round-trip per variant; multi-item mixed-kind round-trip; invalid trigger value; unrepresentable shape diagnostic.

## 3. Migrate `lens_dashboard_app_config.by_reference.drilldowns_json` → structured `drilldowns`

- [x] 3.1 Replace `drilldowns_json` attribute in `getLensByReferenceAttributes()` with the structured `drilldowns` attribute from step 2.1.
- [x] 3.2 Update `lens-dashboard-app` write path in `internal/kibana/dashboard/models_lens_dashboard_app_*.go` to populate API `config.drilldowns` from the structured model via `models_drilldowns.go`.
- [x] 3.3 Update `lens-dashboard-app` read path to populate the structured `drilldowns` list from API `config.drilldowns`; remove the `drilldowns_json` JSON-string handling.
- [x] 3.4 Update `lens_dashboard_app_panel` model struct to replace `DrilldownsJSON jsontypes.Normalized` with the structured drilldowns model field.
- [x] 3.5 Remove `drilldowns_json` entirely from `getLensByReferenceAttributes()` (the resource is unreleased — no migration validator is required; the Plugin Framework surfaces an "unsupported attribute" diagnostic for any stale practitioner config).
- [x] 3.6 Update existing acceptance tests under `internal/kibana/dashboard/acc_lens_dashboard_app_panels_test.go` (and any other test that uses `drilldowns_json`) to use the structured `drilldowns` shape.
- [x] 3.7 Update `models_lens_dashboard_app_*_test.go` unit tests for the new field shape.

## 4. Add `viz_config` schema (REQ-042)

- [x] 4.1 Add `viz_config` to `panelConfigNames` (insert in the new shrunk list — see step 7).
- [x] 4.2 Implement `getVizByValueAttributes()` returning a 12-block attribute map (`xy_chart_config`, `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `tagcloud_config`, `region_map_config`, `datatable_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, `waffle_config`); reuse the existing `getXYChartConfigAttributes()`, `getMetricChart()`, etc. functions verbatim (no chart-block internal change).
- [x] 4.3 Implement `vizByValueSourceValidator` (exactly one of the 12 chart blocks set; reject zero or 2+).
- [x] 4.4 Implement `getVizConfigSchema()` returning `by_value` (using `getVizByValueAttributes()` + `vizByValueSourceValidator`) and `by_reference` (using `getLensByReferenceAttributes()`); add `vizConfigModeValidator` enforcing exactly-one of `by_value`/`by_reference`.
- [x] 4.5 Wire `viz_config` into the panel-level attribute map in `getSchema()` with: `Optional`, sibling-conflict validators (matching the new shrunk `panelConfigNames`), and `AllowedIfDependentPathExpressionOneOf(panel.type, ["vis"])`.
- [x] 4.6 Add `vizConfigModel` Go struct (`ByValue *vizByValueModel`, `ByReference *vizByReferenceModel`) and helper functions for read/write classification per design D10.

## 5. Move 12 chart blocks from panel level to under `viz_config.by_value`

- [x] 5.1 Remove the 12 chart-block attributes from the panel-level attribute map in `getSchema()`; they are now reachable only via `viz_config.by_value` (and `lens_dashboard_app_config.by_value` continues to expose 11 of them).
- [x] 5.2 Update `panelConfigNames` to drop the 12 chart-block names (final list: `config_json`, `markdown_config`, `viz_config`, `lens_dashboard_app_config`, `esql_control_config`, `options_list_control_config`, `range_slider_control_config`, `time_slider_control_config`, `slo_burn_rate_config`, `slo_overview_config`, `slo_error_budget_config`, `synthetics_monitors_config`, `synthetics_stats_overview_config`).
- [x] 5.3 Update each chart converter used by `lensVizConverters` (`populateFromAttributes` / `buildAttributes`) to operate on the shared `lensByValueChartBlocks` embedded in both `vizByValueModel` and `lensDashboardAppByValueModel`. The `panelModel` zombies from the pre-move schema are gone; **`mapPanelFromAPI` rewiring that populates `viz_config.by_value`** is completed in task 6.
- [x] 5.4 Update `panel_config_validator.go` to dispatch on the new top-level panel-config shape (panel.type → exactly one of {`viz_config`, `lens_dashboard_app_config`, `markdown_config`, `esql_control_config`, `options_list_control_config`, `range_slider_control_config`, `time_slider_control_config`, `slo_burn_rate_config`, `slo_overview_config`, `slo_error_budget_config`, `synthetics_monitors_config`, `synthetics_stats_overview_config`} or `config_json`).

## 6. Refactor `mapPanelFromAPI` and `toAPI` for `case "vis":`

- [x] 6.1 In `internal/kibana/dashboard/models_panels.go` `case "vis":`, classify the API `config` per design D10 (try by_reference: object with non-empty `ref_id` and `time_range`, no top-level chart `type`; else by_value: dispatch to chart-kind detection via existing `detectLensVizType` from `models_lens_panel.go`).
- [x] 6.2 Reuse `detectLensVizType` to populate `viz_config.by_value.<chart_block>` (replacing the previous panel-level chart-block assignment).
- [x] 6.3 Implement `viz_config.by_reference` read path: populate `ref_id`, `references_json`, `title`, `description`, `hide_title`, `hide_border`, `time_range`, structured `drilldowns` (via `models_drilldowns.go`).
- [x] 6.4 Symmetric write path in `toAPI`: when `viz_config.by_value` is set, emit the chart object as inline `config`; when `viz_config.by_reference` is set, emit the object form.
- [x] 6.5 Preserve existing panel-level `config_json` behavior for `type = "vis"` with no `viz_config`: unmarshal `config_json` directly into `KbnDashboardPanelTypeVis_Config` (existing code path).
- [x] 6.6 Preserve REQ-009 by-reference preservation per design D10 step (3): if API response cannot be classified and prior state had `viz_config.by_reference`, preserve prior block.

## 7. Simplify `panelConfigNames` plumbing and per-block descriptions

- [x] 7.1 Update `siblingPanelConfigPathsExcept` callers throughout `schema.go` to operate on the shrunk list.
- [x] 7.2 Update `panelConfigDescription` strings on each top-level panel-config block to reference the new sibling list.
- [x] 7.3 Add a unit test asserting that every entry in `panelConfigNames` corresponds to a registered top-level attribute on the panel object schema, and vice versa.

## 8. Update existing acceptance and unit tests

- [x] 8.1 Inventory all acceptance tests using a panel-level chart block: `acc_xy_panels_test.go`, `acc_metric_panels_test.go`, `acc_legacy_metric_panels_test.go`, `acc_gauge_panels_test.go`, `acc_heatmap_panels_test.go`, `acc_tagcloud_panels_test.go`, `acc_region_map_panels_test.go`, `acc_datatable_panels_test.go`, `acc_pie_chart_panels_test.go`, `acc_mosaic_panels_test.go`, `acc_treemap_panels_test.go`, `acc_waffle_panels_test.go`. For each, wrap the chart block HCL in `viz_config = { by_value = { ... } }` and update `resource.TestCheckResourceAttr` paths accordingly.
- [x] 8.2 Update `acc_lens_dashboard_app_panels_test.go` to use structured `drilldowns` instead of `drilldowns_json` (covered by step 3.6 — confirm here).
- [x] 8.3 Update unit-test fixtures in `models_*_panel_test.go` and `models_panels_test.go` to reflect the new model nesting where they construct `panelModel` directly.
- [x] 8.4 Update `lens_by_value_embed_wiring_test.go`, `lens_dashboard_app_validator_test.go`, `models_lens_by_value_union_parity_test.go` to reflect new helper functions and structured drilldowns.
- [x] 8.5 Update `panel_config_validator_test.go` to cover the simplified mutual-exclusion shape.
- [x] 8.6 Run `go test ./internal/kibana/dashboard/... -count=1 -short` (unit) and confirm all pass.

## 9. Add new acceptance coverage for `viz_config` and structured drilldowns

- [x] 9.1 Add `acc_viz_config_by_reference_test.go` covering: minimal by_reference panel (ref_id + time_range only); by_reference with `references_json`, `title`, `description`, `hide_*`; round-trip preserves all set fields; null preservation for unset optional fields.
- [x] 9.2 Add structured drilldown acceptance coverage on `viz_config.by_reference`: dashboard variant with all optional fields; URL variant with explicit `trigger`; URL variant trigger required (plan-time error when `trigger` unset); discover variant; mixed-kind multi-item list (3 items, one of each kind).
- [x] 9.3 Add structured drilldown acceptance coverage on `lens_dashboard_app_config.by_reference`: same matrix as 9.2 (verifies the shared helper produces identical behavior).
- [x] 9.4 Run targeted acceptance tests against the local stack: `TF_ACC=1 go test -v -run 'TestAccResourceDashboardVizConfigByReference|TestAccResourceDashboardLensDashboardAppByReference_(dashboard|discover|url|mixed)Drilldown' ./internal/kibana/dashboard/...`.
- [x] 9.5 Run the full dashboard acceptance suite: `TF_ACC=1 go test -count=1 -timeout 30m ./internal/kibana/dashboard/...` and confirm 100% pass.

## 10. Update examples and generated docs

- [x] 10.1 Update HCL under `examples/resources/elasticstack_kibana_dashboard/` to wrap chart blocks in `viz_config.by_value` and use structured `drilldowns` where applicable.
- [x] 10.2 Add at least one example demonstrating `viz_config.by_reference`.
- [x] 10.3 Verify the `kibana_dashboard` resource generated docs locally with experimental schema enabled (`TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` on the same `tfplugindocs generate …` invocation as `make docs-generate`; that Makefile target pins `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=false`). Spot-check key sections (`viz_config.by_value`, `viz_config.by_reference`, structured `drilldowns`, required `url.trigger`). Do NOT commit a regenerated `docs/resources/kibana_dashboard.md` — the resource is experimental and the default `make docs-generate` excludes it from the committed docs tree.
- [x] 10.4 Run `TF_ACC=1 go test ./internal/acctest -run '^TestAccExamples_planOnly$' -count=1` to confirm all updated example HCL still plans cleanly.

## 11. Update OpenSpec canonical spec

- [x] 11.1 Apply the delta spec from `openspec/changes/restructure-vis-as-symmetric-config/specs/kibana-dashboard/spec.md` into `openspec/specs/kibana-dashboard/spec.md` (handled at archive time by OpenSpec; no manual edit required during implementation, but verify the delta validates cleanly).
- [x] 11.2 Run `openspec validate restructure-vis-as-symmetric-config` and `openspec validate kibana-dashboard --strict` to confirm both the change and the post-merge canonical spec are well-formed.

## 12. Final verification

- [x] 12.1 Run `make build` — green.
- [x] 12.2 Run `go vet ./...` — green.
- [x] 12.3 Run full unit tests: `make test` — green (run via `go test ./...` or with `.env`/Makefile `KIBANA_PORT` aligned to the live stack — `make test` will fail if the Makefile-pinned port differs).
- [x] 12.4 Run full dashboard acceptance suite (step 9.5) — green.
- [x] 12.5 Update CHANGELOG.md with breaking-change entries (chart blocks moved to `viz_config.by_value`; `drilldowns_json` on lens-dashboard-app replaced with structured `drilldowns`; new `viz_config.by_reference` capability).
- [x] 12.6 Manual review: confirm `panelConfigNames` count equals 13, no stale `drilldowns_json` references remain on `lens_dashboard_app_config.by_reference`, and `siblingPanelConfigPathsExcept` callers all reference the shrunk list.
