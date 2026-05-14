## 1. Create lenscommon package

- [x] 1.1 Create `dashboard/lenscommon/iface.go` with `VizConverter` and `Resolver` interfaces
- [x] 1.2 Create `dashboard/lenscommon/registry.go` with global converter map, `Register`, `ForType`, `FirstForBlocks`, `All`
- [x] 1.3 Create `dashboard/lenscommon/presentation.go` moving `lensChartPresentationReadsFor` and `lensChartPresentationWritesFor` from `dashboard/`
- [x] 1.4 Create `dashboard/lenscommon/drilldowns.go` moving shared drilldown conversion helpers
- [x] 1.5 Create `dashboard/lenscommon/by_reference.go` moving shared by_reference read/write logic
- [x] 1.6 Create `dashboard/lenscommon/slice_alignment.go` with `RegisterSliceAligner` and `ApplySliceAligners`

## 2. Migrate LensByValueChartBlocks to models

- [x] 2.1 Move `lensByValueChartBlocks` struct from `dashboard/models_lens_by_value_chart_blocks.go` to `models/lens.go` as exported `LensByValueChartBlocks`
- [x] 2.2 Update all references across dashboard package to use `models.LensByValueChartBlocks`
- [x] 2.3 Relocate `lensByValueChartBlocksFromPanel` (now exported as `LensByValueChartBlocksFromPanel`), `seedWaffleLensByValueChartFromPriorPanel`, and `firstLensVisConverterForChartBlocks` to their nearest typed homes in `dashboard/` so `models_lens_by_value_chart_blocks.go` can be deleted.
- [x] 2.4 Delete `dashboard/models_lens_by_value_chart_blocks.go`

## 3. Migrate converters (12 total)

For each converter, create `dashboard/panel/lens{kind}/converter.go` with self-registration in `init()`:

- [x] 3.1 `lensxy/` — XY chart converter (`VizType: "xy"`)
- [x] 3.2 `lensgauge/` — Gauge converter (`VizType: "gauge"`)
- [x] 3.3 `lensmetric/` — Metric chart converter (`VizType: "metric"`)
- [x] 3.4 `lenslegacymetric/` — Legacy metric converter (`VizType: "legacy_metric"`)
- [x] 3.5 `lenspie/` — Pie chart converter (`VizType: "pie"`)
- [x] 3.6 `lenstreemap/` — Treemap converter (`VizType: "treemap"`)
- [x] 3.7 `lensmosaic/` — Mosaic converter (`VizType: "mosaic"`)
- [x] 3.8 `lensdatatable/` — Datatable converter (`VizType: "datatable"`)
- [x] 3.9 `lenstagcloud/` — Tagcloud converter (`VizType: "tagcloud"`)
- [x] 3.10 `lensheatmap/` — Heatmap converter (`VizType: "heatmap"`)
- [x] 3.11 `lensregionmap/` — Region map converter (`VizType: "region_map"`)
- [x] 3.12 `lenswaffle/` — Waffle converter (`VizType: "waffle"`)

Each converter must:
- Implement `lenscommon.VizConverter`
- Call `lenscommon.Register(converter{})` in `init()`
- Include `converter_test.go` with existing unit test coverage

## 4. Delete old lens infrastructure

- [x] 4.1 Delete `dashboard/models_lens_panel.go` (replaced by `lenscommon/`)
- [ ] 4.2 Delete `dashboard/models_lens_dashboard_app_converters.go` (by_reference moves to `lenscommon/`) — **deferred**: file still holds live `lensDashboardAppByValueToAPI` / scratch-panel glue (~534 lines); needs staged extraction before deletion.
- [ ] 4.3 Delete `dashboard/models_lens_dashboard_app_by_value_adapter.go` (absorbed into composites later) — **deferred**: still owns `LensByValueChartBlocksFromPanel`, typed lens-app block wiring, and metric expansion helpers; moving to `lenscommon/blocks.go` needs cycle-free model-only refactor (or composite-panel contract).
- [x] 4.4 Delete `dashboard/models_lens_dashboard_app_panel.go` (absorbed into composites later) — **N/A on this branch** (file does not exist — superseded by registry / composite handler work).
- [x] 4.5 Verify `dashboard/models_lens_by_value_chart_blocks.go` remains deleted (completed in §2.4)
- [x] 4.6 Update or delete `dashboard/models_lens_panel.go` tests (moved to converter packages) — dashboard wiring covered via `lenscommon.ForType` / `PopulateFromAttributes` in per-chart tests, `models_panels_test.go`, adapter tests, and regression `lens_by_value_embed_wiring_test.go`.

## 5. Refactor state alignment and defaults

- [x] 5.1 Refactor `alignPanelStateFromPlan` in `models_plan_state_alignment.go` to delegate lens chart alignment to `lenscommon.All()` converters
- [x] 5.2 Register XY chart slice aligner in `lensxy/converter.go` `init()`
- [x] 5.3 Remove explicit per-chart alignment calls from `alignPanelStateFromPlan`
- [x] 5.4 Refactor `populateLensAttributesDefaults` in `panel_config_defaults.go` to dispatch to `lenscommon.ForType(vizType).PopulateJSONDefaults()`
- [x] 5.5 Remove hard-coded lens chart type switches from `panel_config_defaults.go`

## 6. Verification

- [x] 6.1 `go build ./internal/kibana/dashboard/...` passes (verified locally during Sections 1–5)
- [x] 6.2 `go vet ./...` passes (verified locally during Sections 4+5)
- [x] 6.3 `go test ./internal/kibana/dashboard/...` passes (583 unit tests in `dashboard/` plus 64+ in `lenscommon/` + per-chart packages)
- [ ] 6.4 All Lens chart acceptance tests pass (XY, gauge, metric, pie, treemap, mosaic, datatable, tagcloud, heatmap, region_map, legacy_metric, waffle) — to be verified by CI on the PR (no local Elastic stack reachable in this worktree)
- [x] 6.5 `make build` passes (verified locally during Sections 1–5)
