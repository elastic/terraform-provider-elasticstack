# Kibana 9.6 Lens read-back findings

Evidence for OpenSpec change `lens-kibana-9-6-readback-defaults` task 1. Captured against a live stack; do not treat issue text as a substitute for the quotes below.

## Stack under test

Confirmed from live APIs (not `.env` alone):

- Elasticsearch `GET /`: `version.number = 9.6.0-SNAPSHOT`, `build_snapshot = true`
- Kibana `GET /api/status`: `version.number = 9.6.0`, `version.build_snapshot = true`, `status.overall.level = available` (build `daf212d53e88`, build_number `110815`)

The worktree initially had a leftover 9.4.0 stack. It was replaced with `STACK_VERSION=9.6.0-SNAPSHOT` (`make docker-clean` then `make docker-fleet`) before any probe.

## 1.1 Metric `axis` (and `color`) by chart family

### Answers

| Question | Confirmed answer |
|---|---|
| Primary Y metric omitted `axis` | Kibana injects `"axis":"y"` |
| Secondary Y metric with `"axis":"y2"` | Kibana **preserves** `"y2"` (does not overwrite to `"y"`) |
| Chart-level `axis.y2` without assigning a metric to y2 | The single Y metric still gets `"axis":"y"` |
| X-axis metrics | **No applicable path.** XY `x` / `x_json` is a dimension (date histogram, terms, or ES\|QL column), not a metric `config_json`. The suite has no X-axis metric. Not probed as a metric. |
| Does injected `axis` always equal `"y"`? | **No.** Omitted metrics default to `"y"`; an explicit assignment is echoed. Implementation must not hard-code `"y"` as the only value. |

`color: {"type":"auto"}` is injected together with `axis` on every XY Y metric that omitted `color`.

### Per-family results

| Family | Path | Injects `axis`? | Value(s) | Also injects `color:{type:auto}`? | Probe |
|---|---|---|---|---|---|
| XY primary Y | `xy_chart_config.layers[].data_layer.y[].config_json` | **Yes** | `"y"` | **Yes** | `TestAccResourceDashboardXYChart_basic`, `_axis` |
| XY secondary Y | same `y[]`, practitioner set `"axis":"y2"` | **Preserves** | `"y2"` | **Yes** | one-off `TestAccLensTask1Probe_XYSecondaryY` + `TF_LOG` create/read payload |
| XY ES\|QL Y | `y[].config_json` with `column` | **Yes** | `"y"` | **Yes** (overwrites planned static color in the after-apply string) | `TestAccResourceDashboardXYChart_layers` |
| XY X dimension | `data_layer.x_json` | N/A | not a metric | N/A | suite only has dimension `x` |
| Datatable | `datatable_config.no_esql.metrics[].config_json` | **No** | — | **Yes** (plus `visible:true`, `alignment:"right"`) | `TestAccResourceDashboardDatatableChart`, `TestAccLensMinimalProbe_Datatable` |
| Metric chart | `metric_chart_config.metrics[].config_json` | **No** | — | already covered by existing defaults; apply **passed** | `TestAccResourceDashboardMetricChartMinimalConfig`, `TestAccLensMinimalProbe_Metric` |
| Pie | `pie_chart_config.metrics[].config_json` | **No** | — | already covered; apply failed only on legend `truncate_after_lines`/`nested` | `TestAccResourceDashboardPieChart`, `TestAccLensMinimalProbe_Pie` |
| Gauge | `gauge_config.metric_json` | **No** | — | already covered; apply **passed** | `TestAccResourceDashboardGauge`, `TestAccLensMinimalProbe_Gauge` |
| Tagcloud | `tagcloud_config.metric_json` | **No** | — | already covered; apply **passed** | `TestAccResourceDashboardTagcloud`, `TestAccLensMinimalProbe_Tagcloud` |
| Treemap / mosaic | `*_config.metrics_json` | **No** | — | mosaic/treemap metric drift in the percent-format one-off was format/`empty_as_null`/`color`, not `axis` | `TestAccLensMinimalProbe_{Treemap,Mosaic}` + percent-format one-off |
| Region map | `region_map_config.metric_json` | **No** | — | already covered; apply **passed** | `TestAccResourceDashboardRegionMap`, `TestAccLensMinimalProbe_RegionMap` |
| Legacy metric | `legacy_metric_config.metric_json` | **No** | — | **Yes**, plus legacy-only `size` (see 1.3) | `TestAccResourceDashboardLegacyMetricChart`, `TestAccLensMinimalProbe_LegacyMetric` |

### Planned vs actual quotes

**XY primary Y** (`TestAccResourceDashboardXYChart_basic` and `_axis` — `_axis` has chart-level `axis.y2` but only one Y metric):

```
was  {"empty_as_null":true,"operation":"count"}
now  {"empty_as_null":true,"operation":"count","axis":"y","color":{"type":"auto"}}
```

**XY secondary Y** (one-off create request vs Kibana read-back from `TF_LOG`):

```
request y[0]  {"empty_as_null":true,"operation":"count"}
request y[1]  {"axis":"y2","empty_as_null":true,"operation":"count"}

read-back y[0]  {"empty_as_null":true,"operation":"count","axis":"y","color":{"type":"auto"}}
read-back y[1]  {"empty_as_null":true,"operation":"count","axis":"y2","color":{"type":"auto"}}
```

The after-apply diagnostic only reported `y[0]`. That matches `JSONWithDefaultsValue` semantic equality: `PopulateLensMetricDefaults` already treats `color` as a default, so `y[1]` (`axis` already `"y2"`) compared equal, while `y[0]` drifted on the uncovered `axis` key.

**XY ES\|QL layer** (`TestAccResourceDashboardXYChart_layers`):

```
was  {"color":{"color":"#54B399","type":"static"},"column":"system.cpu.user.pct","format":{"type":"number"}}
now  {"column":"system.cpu.user.pct","format":{"type":"number","decimals":2,"compact":false},"axis":"y","color":{"type":"auto"}}
```

**Datatable** (`TestAccLensMinimalProbe_Datatable`):

```
was  {"operation":"count"}
now  {"empty_as_null":false,"operation":"count","visible":true,"color":{"type":"auto"},"alignment":"right"}
```

No `axis` key. `TestAccResourceDashboardDatatableChart` added the same `visible` / `color` / `alignment` extras on top of an already-populated format/`empty_as_null` plan.

**Pie** — no metric `config_json` diagnostic. Failures were legend-only:

```
pie_chart_config.legend.truncate_after_lines: was null, now 1
pie_chart_config.legend.nested: was null, now false
```

**Gauge / tagcloud / region-map / metric chart** — apply succeeded; Kibana did not inject an uncovered `axis` (or any other uncovered metric key) on those paths.

### Current metric-default populators (none inject `axis`)

| Function | Used by | Already injects |
|---|---|---|
| `PopulateLensMetricDefaults` | XY `y[]` (`lensxy/defaults.go`); datatable `metrics[]` (`lensdatatable/defaults.go`) | `empty_as_null` (gated), `fit`, `color:{type:auto}`, number/percent format `compact`/`decimals`; metric-chart primary/secondary alignment extras |
| `PopulateMetricChartMetricDefaults` | metric chart `metrics[]` (`lensmetric/defaults.go`) | `PopulateLensMetricDefaults` + secondary `color:{type:none}` |
| `PopulatePieChartMetricDefaults` | pie + waffle `metrics[]` (`lenspie/defaults.go`, `lenswaffle/defaults.go`) | `empty_as_null` (gated), `color:{type:auto}`, number format `compact`/`decimals` |
| `PopulateGaugeMetricDefaults` | gauge `metric` (`lensgauge/defaults.go`) | `empty_as_null` (gated), `title.visible`, `ticks`, `color:{type:auto}` |
| `populateFieldMetricLensDefaults` via `PopulateTagcloudMetricDefaults` / `PopulateRegionMapMetricDefaults` | tagcloud, region-map, heatmap (`lenstagcloud`, `lensregionmap`, `lensheatmap` defaults.go) | `empty_as_null` (gated), `show_metric_label`, `color:{type:auto}` |
| `PopulatePartitionMetricsDefaults` → `PopulateTagcloudMetricDefaults` | treemap / mosaic `metrics` (`PopulatePartitionLensAttributes` in `lenscommon/populate_lens_charts.go`) | same field-metric defaults |
| `PopulateLegacyMetricMetricDefaults` | legacy metric `metric` (`lenslegacymetric/defaults.go`) | `show_array_values`, `empty_as_null` (gated), format `decimals`/`compact` — **not** `color` or `axis` |

Implication for task 2: a shared axis primitive is only required on paths that actually receive `axis` from Kibana 9.6 — confirmed **XY `y[]` only**. Other families should not assume an `axis` default. Legacy still needs `color` (shared) plus `size` (legacy-only); that is task 1.3 / later slices, not an `axis` gap.

## 1.2 `value_display.percent_decimals`

### Answer

**Single default: `2`.** It does not vary by metric number `format`, and it is also injected when `value_display.mode` is `"absolute"`.

Existing mosaic/treemap testdata either set `percent_decimals = 2` (`basic`) or omit it under `mode = "absolute"` (`complete`). No testdata used a non-default metric `format`, so a one-off with `format = { type = "percent" }` was applied.

### Evidence

When the practitioner omits `value_display` entirely (minimal probes, default count metric / implicit number format):

```
treemap_config.value_display: was null,
  now {mode:"percentage", percent_decimals:2}

mosaic_config.value_display: was null,
  now {mode:"percentage", percent_decimals:2}
```

Same `percent_decimals: 2` with metric `format.type = "percent"` (one-off `TestAccLensTask1Probe_{Mosaic,Treemap}PercentFormat`):

```
value_display: was null,
  now {mode:"percentage", percent_decimals:2}
```

When the practitioner sets `value_display.mode = "absolute"` and omits `percent_decimals` (`TestAccResourceDashboardMosaic` / `Treemap` step 2 `complete`; step 1 `basic` with explicit `percent_decimals = 2` passed):

```
mosaic_config.value_display.percent_decimals:  was null, now 2
treemap_config.value_display.percent_decimals: was null, now 2
```

`PartitionValueDisplayMatchesKibanaDefault` today only accepts `{mode="percentage", percent_decimals=null}`. It must also accept `percent_decimals = 2` (and, for the `complete` step, `mode = "absolute"` with injected `2` is a practitioner-set mode plus a defaulted decimals field — task 4 should preserve null `percent_decimals` when state is `2`, regardless of mode).

## 1.3 Legacy `metric_json`

### Answer

**Not only shared `axis`/`color` drift.** Kibana 9.6 does **not** inject `axis` on legacy metric. It injects:

1. `color: {"type":"auto"}` — shared metric default, but **not** currently applied by `PopulateLegacyMetricMetricDefaults`
2. `size: "m"` — **legacy-only** injected key

`empty_as_null` also appears on read-back when omitted; that key is already populated by `PopulateLegacyMetricMetricDefaults` and is not new 9.6 drift.

### Planned vs actual quotes

`TestAccLensMinimalProbe_LegacyMetric`:

```
was  {"operation":"count"}
now  {"empty_as_null":false,"operation":"count","size":"m","color":{"type":"auto"}}
```

`TestAccResourceDashboardLegacyMetricChart` (practitioner already set `empty_as_null` and format):

```
was  {"empty_as_null":false,"format":{"compact":false,"decimals":2,"type":"number"},"operation":"count"}
now  {"format":{"type":"number","decimals":2,"compact":false},"empty_as_null":false,"operation":"count","size":"m","color":{"type":"auto"}}
```

Injected keys relative to a fully specified count metric: **`size`**, **`color`**. No other legacy-only keys observed.

## Implications for later tasks (not implemented here)

- Task 2 should add axis normalization for **XY `y[]`**, using the metric's own `axis` when present and `"y"` only when omitted. Do not hard-code `"y"` onto datatable / pie / gauge / tagcloud / partition / region-map / legacy metrics.
- Task 4 can treat `percent_decimals = 2` as a fixed Kibana default (not format-aware).
- Legacy needs `color` plus `size:"m"` on `PopulateLegacyMetricMetricDefaults` (or a sibling). That is extra scope vs a pure axis primitive; surface it in task 2/legacy wiring rather than assuming axis-only.

## Probes run

| Test | Result | Used for |
|---|---|---|
| `TestAccResourceDashboardXYChart_basic` | FAIL (axis+color on `y[0]`) | 1.1 |
| `TestAccResourceDashboardXYChart_axis` | FAIL (same; still `"axis":"y"`) | 1.1 |
| `TestAccResourceDashboardXYChart_layers` | FAIL (ES\|QL `axis":"y"`) | 1.1 |
| `TestAccResourceDashboardDatatableChart` | FAIL (visible/color/alignment; no axis) | 1.1 |
| `TestAccResourceDashboardMetricChartMinimalConfig` | PASS | 1.1 |
| `TestAccResourceDashboardPieChart` | FAIL (legend only) | 1.1 |
| `TestAccResourceDashboardGauge` | PASS | 1.1 |
| `TestAccResourceDashboardTagcloud` | PASS | 1.1 |
| `TestAccResourceDashboardRegionMap` | PASS | 1.1 |
| `TestAccResourceDashboardMosaic` | FAIL step 2 (`percent_decimals=2` on absolute) | 1.2 |
| `TestAccResourceDashboardTreemap` | FAIL step 2 (same) | 1.2 |
| `TestAccResourceDashboardLegacyMetricChart` | FAIL (`size`+`color`) | 1.3 |
| `TestAccLensMinimalProbe_{Metric,Gauge,Tagcloud,RegionMap}` | PASS | 1.1 |
| `TestAccLensMinimalProbe_{Datatable,Pie,Mosaic,Treemap,LegacyMetric}` | FAIL as quoted above | 1.1–1.3 |
| One-off XY `y` + `y2` (deleted after capture) | FAIL on `y[0]` only; TF_LOG showed `y`/`y2` | 1.1 |
| One-off mosaic/treemap `format.type=percent` (deleted after capture) | `percent_decimals` still `2` | 1.2 |
