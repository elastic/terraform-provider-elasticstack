## 1. Confirm 9.6 read-back shapes

- [x] 1.1 Against a live `9.6.0-SNAPSHOT` (or later) Kibana, confirm injected metric `axis` behavior for XY `y[]` (primary Y, secondary Y, and X-axis metrics where applicable), datatable `metrics[]`, metric chart, pie, gauge, tagcloud, treemap/mosaic, region-map, and legacy metric paths.
- [x] 1.2 Confirm whether `mosaic_config`/`treemap_config.value_display.percent_decimals` has a single default or varies by number `format`.
- [x] 1.3 Confirm whether `legacy_metric_config.metric_json` has only shared metric-default drift (`axis`/`color`) or any legacy-only injected keys.

## 2. Metric `config_json` axis normalization

- [x] 2.1 Add a shared metric-axis normalization primitive in `internal/kibana/dashboard/lenscommon/` and wire it into each confirmed metric default-population path (for example `PopulateLensMetricDefaults`, `PopulatePieChartMetricDefaults`, `PopulateGaugeMetricDefaults`, and legacy metric/default-populator paths), while keeping datatable/partition grouping JSON paths on existing group-by logic.
- [x] 2.2 Add/update unit tests in the relevant `lenscommon` and panel alignment test files for: unset-axis metrics preserving plan intent, explicitly set axis values not being overridden, and composition with `empty_as_null` gating.
- [x] 2.3 Run the metric-focused acceptance suites listed in issue #4902 except `TestAccResourceDashboardXYChart_layers`. Reconfirmed on head `250bac4a` (run [`35158144564`](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/35158144564); shard-0 job [`105013489862`](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/35158144564/job/105013489862), shard-1 job [`105013489856`](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/35158144564/job/105013489856)): every listed suite except `_layers` passed on `9.6.0-SNAPSHOT`. Listed suites excluding `_layers`: `TestAccResourceDashboardXYChart_basic` plus `_axis`, `_decorations`, `_filters`, `_fitting`, `_layers_reference`, `_legend_inside`, `_legend_outside`, `_chartTimeRangeLifecycle`, `_lensPresentationFields`, `TestAccResourceDashboardXYChartMinimalConfig`, `TestAccDashboardXYMetricEmptyAsNullGating`, `TestAccReproduceIssue3402`, `TestAccReproduceIssue3707`, `TestAccResourceDashboardDatatableChart`, `TestAccResourceDashboardDatatableChart_lensPresentationCrossCutting`, and minimal probes covering `Metric`, `Gauge`, `Tagcloud`, `RegionMap`, and `LegacyMetric`.
- [x] 2.4 Preserve Kibana 9.6 XY reference-line threshold omit-defaults (`thresholds[].axis` → `"y"`, `operation` → `"static_value"`, `color_json` → `{"type":"auto"}`, `value_json` object→scalar restore) in `internal/kibana/dashboard/panel/lensxy/alignment_panels.go` (`alignXYLayerStateFromPlan` / `preserveThresholdValueJSONIfStateIsPlanValue`). Covered by `TestAccResourceDashboardXYChart_layers_reference`.

## 3. Pie / waffle legend defaults

- [x] 3.1 In `internal/kibana/dashboard/panel/lenspie/alignment.go`, align `legend.truncate_after_lines` (default `1`) and `legend.nested` (default `false`) from plan using `lenscommon.PreserveNullIfStateEquals`.
- [x] 3.2 In `internal/kibana/dashboard/panel/lenswaffle/alignment.go`, align `legend.truncate_after_lines` (default `1`) from plan using the same mechanism.
- [x] 3.3 Add/update unit tests for both paths and run `TestAccResourceDashboardPieChart`, `TestAccResourceDashboardPieChart_lensPresentationCrossCutting`, `TestAccResourceDashboardWaffle`, and `TestAccLensMinimalProbe_{Pie,Waffle}` in this slice.
- [x] 3.4 Align `treemap_config`/`mosaic_config` `legend.truncate_after_lines` (default `1`) and `legend.nested` (default `false`) via the shared `lenscommon.AlignPartitionLegendStateFromPlan` helper.

## 4. Partition `value_display.percent_decimals`

- [x] 4.1 Keep the requirement evidence-driven: after task 1.2 confirms actual behavior, update `PartitionValueDisplayMatchesKibanaDefault` in `internal/kibana/dashboard/lenscommon/partition_alignment.go` only for verified default shape(s).
- [x] 4.2 Add/update `partition_alignment` unit tests for every accepted default shape and for at least one non-default practitioner value that must not be normalized away.
- [x] 4.3 Run `TestAccResourceDashboardMosaic`, `TestAccResourceDashboardTreemap`, and `TestAccLensMinimalProbe_{Mosaic,Treemap}` in this slice.

## 5. Heatmap axis-label orientation

- [x] 5.1 Add `axis.x.labels.orientation` default (`"horizontal"`) preservation to the heatmap alignment path using `lenscommon.PreserveNullIfStateEquals`, parallel to existing `labels.visible` / `title.visible` handling. Y has no `orientation` in the schema or API.
- [x] 5.2 Add/update heatmap alignment unit tests and run heatmap acceptance coverage (including `TestAccLensMinimalProbe_Heatmap`) in this slice.
- [x] 5.3 Align `heatmap_config.legend.truncate_after_lines` (default `1`) from plan using `lenscommon.PreserveNullIfStateEquals`.

## 6. Spec sync and validation

- [x] 6.1 After implementation evidence is collected, sync this delta into `openspec/specs/kibana-dashboard/spec.md` (REQ-011) with only verified normative defaults.
- [x] 6.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate lens-kibana-9-6-readback-defaults --type change`.
- [x] 6.3 During implementation, run `make build`, `go vet ./internal/kibana/dashboard/...`, and `go test ./internal/kibana/dashboard/...`; run the acceptance tests above with `TF_ACC=1` against a running `9.6.0-SNAPSHOT` stack per `dev-docs/high-level/testing.md`.
