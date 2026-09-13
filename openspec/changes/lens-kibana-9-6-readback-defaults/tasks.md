## 1. Confirm 9.6 read-back shapes

- [ ] 1.1 Against a live `9.6.0-SNAPSHOT` (or later) Kibana, confirm the exact injected `axis` value(s) for XY `y[].config_json` across primary Y, secondary Y (`y2`), and X-axis metrics, and for datatable/pie/gauge/tagcloud/treemap/mosaic/region-map/legacy-metric `config_json`/`metric_json`. Resolve the open question in `design.md` about whether `axis` always mirrors the metric's own axis assignment.
- [ ] 1.2 Confirm whether `mosaic_config`/`treemap_config.value_display.percent_decimals` is always backfilled to `2`, or varies with the chart's number `format`.
- [ ] 1.3 Confirm `legacy_metric_config.metric_json`'s Kibana-filled blob is fully explained by the shared `axis`/`color` metric defaults gap (task 2.1), or identify any legacy-metric-specific keys not covered by that fix.

## 2. Shared Lens metric config_json (`axis` key)

- [ ] 2.1 Extend `PopulateLensMetricDefaults` in `internal/kibana/dashboard/lenscommon/populate_lens_charts.go` to inject a default `axis` value when the practitioner's metric config omits it, using the confirmed value(s) from task 1.1.
- [ ] 2.2 Add/update unit tests in `internal/kibana/dashboard/lenscommon/*_test.go` covering the new `axis` default for at least: an unset-axis metric, a metric that already specifies an axis explicitly (no override), and a metric operation combined with the existing `empty_as_null` gating so the two defaults compose correctly.
- [ ] 2.3 Confirm no separate code path is needed in `internal/kibana/dashboard/panel/lenslegacymetric` — verify `AlignBasicMetricChartStateFromPlan` picks up the fix via the shared `PopulateLensMetricDefaults` call.

## 3. Pie / waffle legend defaults

- [ ] 3.1 In `internal/kibana/dashboard/panel/lenspie/alignment.go`, align `legend.truncate_after_lines` (default `1`) and `legend.nested` (default `false`) from plan using `lenscommon.PreserveNullIfStateEquals`, alongside the existing `label_position` handling.
- [ ] 3.2 In `internal/kibana/dashboard/panel/lenswaffle/alignment.go`, align `legend.truncate_after_lines` (default `1`) from plan the same way (waffle's schema has no `nested` field).
- [ ] 3.3 Add unit test coverage for both alignment functions confirming a null plan value is preserved when state matches the default, and that a practitioner-set value is never overwritten.

## 4. Partition value_display percent_decimals

- [ ] 4.1 Widen `PartitionValueDisplayMatchesKibanaDefault` in `internal/kibana/dashboard/lenscommon/partition_alignment.go` to also treat `percent_decimals = 2` (using the value confirmed in task 1.2) as matching the Kibana-injected default block, alongside the existing null case.
- [ ] 4.2 Update `internal/kibana/dashboard/lenscommon/partition_alignment_test.go` (or equivalent) with cases for both the null and `2`-value default shapes, and a case confirming a practitioner-set `percent_decimals` (e.g. `1`) is never treated as the default.

## 5. Heatmap axis-label orientation

- [ ] 5.1 Add `axis.{x,y}.labels.orientation` default (`"horizontal"`) preservation to the heatmap alignment path in `internal/kibana/dashboard/panel/lensheatmap/`, parallel to the existing `labels.visible` / `title.visible` handling, using `lenscommon.PreserveNullIfStateEquals`.
- [ ] 5.2 Add unit test coverage for the new orientation default alongside the existing heatmap alignment tests.

## 6. Acceptance tests

- [ ] 6.1 Run (or add if missing) `TestAccResourceDashboardXYChart_basic` and the other XY suites named in the issue (`_axis`, `_decorations`, `_filters`, `_fitting`, `_layers`, `_layers_reference`, `_legend_inside`, `_legend_outside`, `_chartTimeRangeLifecycle`, `_lensPresentationFields`, `TestAccResourceDashboardXYChartMinimalConfig`, `TestAccDashboardXYMetricEmptyAsNullGating`, `TestAccReproduceIssue3402`, `TestAccReproduceIssue3707`) against `9.6.0-SNAPSHOT` and confirm they pass with the task 2 fix.
- [ ] 6.2 Run `TestAccResourceDashboardDatatableChart` and `TestAccResourceDashboardDatatableChart_lensPresentationCrossCutting` against `9.6.0-SNAPSHOT`.
- [ ] 6.3 Run `TestAccResourceDashboardPieChart` and `TestAccResourceDashboardPieChart_lensPresentationCrossCutting` against `9.6.0-SNAPSHOT`.
- [ ] 6.4 Run `TestAccResourceDashboardWaffle`, `TestAccResourceDashboardMosaic`, `TestAccResourceDashboardTreemap`, `TestAccResourceDashboardLegacyMetricChart` against `9.6.0-SNAPSHOT`.
- [ ] 6.5 Run `TestAccLensMinimalProbe_{Pie,Waffle,Mosaic,Treemap,Datatable,Heatmap,LegacyMetric}` against `9.6.0-SNAPSHOT`.
- [ ] 6.6 For any suite that still fails after tasks 2–5, capture the new drift shape and file it as a follow-up (or extend this change) rather than silently loosening an unrelated normalization.

## 7. Spec and validation

- [ ] 7.1 Sync the `specs/kibana-dashboard/spec.md` delta in this change into `openspec/specs/kibana-dashboard/spec.md` (REQ-011) once implementation and acceptance testing confirm the documented defaults.
- [ ] 7.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate lens-kibana-9-6-readback-defaults --type change` and resolve any reported issues.
- [ ] 7.3 During implementation, run `make build`, `go vet ./internal/kibana/dashboard/...`, and `go test ./internal/kibana/dashboard/...` (unit tests); run the acceptance tests in section 6 with `TF_ACC=1` against a running `9.6.0-SNAPSHOT` Elastic Stack per `dev-docs/high-level/testing.md`.
