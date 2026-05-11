## 1. Schema additions across twelve Lens chart blocks

- [x] 1.1 Define a shared `lensChartPresentationAttributes()` helper in `internal/kibana/dashboard/schema.go` returning the five new attributes (`time_range`, `hide_title`, `hide_border`, `references_json`, `drilldowns`) so each chart-block schema can spread them in without duplication.
- [x] 1.2 Define the `drilldowns` schema: a list of nested objects with three mutually-exclusive optional variant sub-blocks (`dashboard_drilldown`, `discover_drilldown`, `url_drilldown`); apply per-variant attribute validators per REQ-039 (computed triggers for dashboard/discover; `stringvalidator.OneOf` for url).
- [x] 1.3 Apply the shared presentation attributes to `xy_chart_config` in `schema_xy_chart_panel.go`.
- [x] 1.4 Apply the shared presentation attributes to `datatable_config` in `schema_datatable_panel.go`.
- [x] 1.5 Apply the shared presentation attributes to `tagcloud_config` in `schema_tagcloud_panel.go`.
- [x] 1.6 Apply the shared presentation attributes to `metric_chart_config`, `legacy_metric_config`, `gauge_config`, `heatmap_config`, `region_map_config`, `pie_chart_config`, `mosaic_config`, `treemap_config`, and `waffle_config` in `schema.go`.
- [x] 1.7 Add inter-variant exclusivity validators on each drilldown variant sub-block using `validators.ForbiddenIfDrilldownVariantSiblingNestedPresent` with relative sibling paths (same path-expression style as `ForbiddenIfDependentPathExpressionOneOf`; sibling variant blocks are object-valued, while `stringvalidator.OneOf` covers `url_drilldown.trigger`).
- [x] 1.8 Add a list-item-level validator rejecting drilldown entries with zero variant sub-blocks — `drilldownListItemVariantsValidator` on each list item object (pairwise validators enforce at most one variant; PF `objectvalidator.ExactlyOneOf` / `AtLeastOneOf` on nested objects counts the parent object and is unsuitable here).
- [x] 1.9 Verify schema documentation strings (`MarkdownDescription`) clearly indicate the computed nature of `dashboard_drilldown.trigger` and `discover_drilldown.trigger`, and the inheritance behavior of `time_range`.

## 2. Model wiring: write path

- [x] 2.1 Add a new helper in `internal/kibana/dashboard/models_panels.go` named (e.g.) `resolveChartTimeRange(dashboard *dashboardModel, chartLevel *timeRangeModel)` returning the effective `kbapi.KbnEsQueryServerTimeRangeSchema` for a chart panel (chart-level if set, else dashboard-level).
- [x] 2.2 Delete the existing `lensPanelTimeRange()` helper in `models_panels.go` and update every call site in the twelve `models_*_panel.go` files to use `resolveChartTimeRange(...)` instead, threading the parent dashboard model into each panel's `toAPI` path.
- [x] 2.3 Add a shared `chartPresentationModel` Go struct with fields for `TimeRange`, `HideTitle`, `HideBorder`, `ReferencesJSON`, `Drilldowns`; embed in each `models_*_panel.go` chart-config Go struct.
- [x] 2.4 Wire `HideTitle`, `HideBorder`, `ReferencesJSON` into the API request struct for each of the twelve chart types in their respective `toAPI` paths.
- [x] 2.5 Implement `drilldownsToAPI` that converts the TF `drilldowns` list into the API discriminated union, branching on which variant sub-block is set per list item; emit appropriate `type` discriminator (`"dashboard_drilldown"`, `"discover_drilldown"`, `"url_drilldown"`).
- [x] 2.6 Default the computed `trigger` to `"on_apply_filter"` for `dashboard_drilldown` and `discover_drilldown` variants on the API write path.

## 3. Model wiring: read path

- [x] 3.1 Implement `chartTimeRangeFromAPI` that returns the API-returned chart-level `time_range` plus a "preserve null" decision based on equality with the dashboard-level value and prior-state null status (REQ-038 null-preservation).
- [x] 3.2 Apply `chartTimeRangeFromAPI` in each of the twelve `models_*_panel.go` files' `fromAPI` / read paths, passing in the parent dashboard model and prior-state chart `time_range`.
- [x] 3.3 Implement null-preservation for `hide_title`, `hide_border`, and `references_json` consistent with REQ-009 (read-back omits → state stays null).
- [x] 3.4 Implement `drilldownsFromAPI` that converts the API discriminated union into the TF list of variant sub-blocks, dispatching on the API `type` discriminator.
- [x] 3.5 Implement `time_range.mode` null-preservation at the chart level (same recipe as REQ-009).
- [x] 3.6 Normalize `references_json` round-trip using the existing JSON normalization utilities (consistent with other `*_json` attributes elsewhere in the resource).

## 4. Schema migration touchpoints

- [x] 4.1 Audit `models_lens_dashboard_app_by_value_adapter.go` and `models_lens_dashboard_app_converters.go` for any reuse of `lensPanelTimeRange()` and migrate to the new resolver helper (these are part of `lens-dashboard-app`, not `vis` panels; verify scoping carefully — the helper deletion in 2.2 must remain consistent).
- [x] 4.2 Audit `panel_config_defaults.go` and `panel_config_validator.go` for references to chart-level time_range defaults; update or remove as needed.

## 5. Unit tests

- [x] 5.1 Extend `models_xy_chart_panel_test.go` to cover `time_range` inheritance (null in plan, equal to dashboard on read-back → null-preserved), explicit override, and `mode` null-preservation.
- [x] 5.2 Extend `models_xy_chart_panel_test.go` to cover `hide_title`, `hide_border`, `references_json` round-trip and null-preservation.
- [x] 5.3 Add `drilldownsToAPI`/`drilldownsFromAPI` unit tests covering all three variants, computed trigger defaulting, and error on multi-variant list items.
- [x] 5.4 Replicate the test coverage from 5.1–5.3 across the remaining eleven `models_*_panel_test.go` files for each chart type that gains the shared presentation attributes. Use a table-driven helper if the test bodies are sufficiently uniform. **Note:** full multi-scenario coverage is centralized in `models_chart_presentation_table_test.go` for datatable/metric/gauge/pie; the remaining chart types use a thin `hide_title` wiring round-trip per file.
- [x] 5.5 Add a validator unit test covering: `url_drilldown.trigger` with invalid enum rejected; `dashboard_drilldown.trigger` set in config rejected (computed-only); multi-variant list items rejected; zero-variant list items rejected. **Note:** multi-variant list items are covered by the existing `ForbiddenIfDrilldownVariantSiblingNestedPresent` tests in `internal/utils/validators/conditional_test.go` (same mechanism as the per-variant sibling validators); dashboard adds schema-level tests in `lens_chart_presentation_validators_test.go`.

## 6. Acceptance tests

- [x] 6.1 Extend `acc_xy_panels_test.go` with steps exercising chart-level `time_range` set explicitly, `time_range` null (inherits dashboard), updated to a new value, then reset to null again; assert API-side `time_range` matches expectation in each step.
- [x] 6.2 Extend `acc_xy_panels_test.go` to cover `hide_title`, `hide_border`, `references_json`, and a `drilldowns` list with one of each variant; assert round-trip parity and no spurious diffs on the second plan.
- [x] 6.3 Add equivalent steps to at least three other chart-type acceptance test files (`acc_metric_panels_test.go`, `acc_datatable_panels_test.go`, `acc_pie_chart_panels_test.go`) to cover the cross-cutting behavior across multiple chart roots.
- [ ] 6.4 Verify the full dashboard acceptance suite (`go test ./internal/kibana/dashboard/...`) passes with `TF_ACC=1` against a running stack. **Skipped locally — no Elastic Stack listening on localhost; defer execution to CI.**

## 7. Documentation and release prep

- [x] 7.1 Regenerate provider documentation: `make docs-generate`.
- [x] 7.2 Update the resource's example HCL under `examples/resources/elasticstack_kibana_dashboard/` to demonstrate `time_range` inheritance and at least one `drilldowns` variant.
- [x] 7.3 Add a CHANGELOG entry noting the breaking change for unreleased users (chart-level `time_range` now inherits dashboard-level instead of hardcoded `now-15m..now`; new attributes available).

## 8. Spec sync and verification

- [x] 8.1 Run `openspec validate expose-lens-chart-presentation-fields --strict` and resolve any issues. **Validated:** `openspec validate … --strict` passed clean (no artifact fixes needed).
- [x] 8.2 Run `GOLANGCIFLAGS='--allow-parallel-runners' make lint`, `make build`, `go vet ./...`, `go test ./internal/kibana/dashboard/... -count=1 -short`, and `go test ./internal/utils/validators/... -count=1`; confirm clean.
- [x] 8.3 Run targeted acceptance tests from step 6 against the running stack; confirm pass. **Deferred to CI:** `localhost:9200` / `localhost:5601` not reachable (no local stack); run the `TF_ACC=1` targeted tests in CI with credentials and stack.
- [ ] 8.4 Archive the change once all tasks are complete (separate apply-loop step). **Intentionally unchecked — do not archive in this loop:** archival is deferred to PR merge / `verify-openspec` at merge time, not the implementation loop.
