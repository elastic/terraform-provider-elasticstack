## Why

Against Elastic Stack `9.6.0-SNAPSHOT`, applying `elasticstack_kibana_dashboard` with a typed Lens `vis` panel fails on the very first apply with `Provider produced inconsistent result after apply`. Kibana 9.6 persists additional server-side Lens defaults on read-back that the provider's existing default-normalization logic does not yet know about, so the post-create read returns a value that differs from the planned value and Terraform rejects the apply.

This is not flaky — it reproduced on every attempt (including 5 `gotestsum` retries) for the `9.6.0-SNAPSHOT` shard-1 run on [PR #4890](https://github.com/elastic/terraform-provider-elasticstack/pull/4890) ([CI run 34672008631](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/34672008631)). The canonical failure is `TestAccResourceDashboardXYChart_basic` (`internal/kibana/dashboard/panel/lensxy/acc_test.go`), where the practitioner's Y-metric `config_json` is planned as `{"empty_as_null":true,"operation":"count"}` and Kibana's read-back returns `{"empty_as_null":true,"operation":"count","axis":"y","color":{"type":"auto"}}`.

The provider already has an established pattern for exactly this class of drift (`internal/kibana/dashboard/lenscommon`'s `PreserveNullIfStateEquals`, `PreservePlanJSONIfStateAddsOptionalKeys`, `PopulateLensMetricDefaults`, and `PartitionValueDisplayMatchesKibanaDefault`), documented in `openspec/specs/kibana-dashboard/spec.md` under "Panel default normalization and XY-axis drift prevention (REQ-011)" and originally added for issue #2355. Kibana 9.6 adds new defaults to that same class of fields that the current implementation of those helpers does not cover:

| Attribute family | Example path | Planned | 9.6 read-back |
|---|---|---|---|
| Lens metric `config_json` | `xy_chart_config.layers[0].data_layer.y[0].config_json` | `{"empty_as_null":true,"operation":"count"}` | adds `"axis":"y"`, `"color":{"type":"auto"}` |
| Lens metric `config_json` | `datatable_config.no_esql.metrics[0].config_json` | `{"operation":"count"}` | same metric-defaults gap |
| Lens metric `config_json` | `legacy_metric_config.metric_json` | practitioner blob | same metric-defaults gap (to confirm via legacy path) |
| Pie/waffle legend | `pie_chart_config.legend.truncate_after_lines` / `.nested` | null | `1` / `false` |
| Waffle legend | `waffle_config.legend.truncate_after_lines` | null | `1` |
| Partition value display | `mosaic_config` / `treemap_config.value_display.percent_decimals` | null | `2` |
| Heatmap axis labels | `heatmap_config.axis.x.labels.orientation` | null | `"horizontal"` |

`color: {type: "auto"}` is already covered on the paths that currently use `lenscommon.PopulateLensMetricDefaults`; the new, uncovered piece is the `axis` key. Chart families do not all share one metric-population function today, so this change should add a shared metric-axis normalization primitive and apply it in each confirmed metric-default path (rather than assuming `PopulateLensMetricDefaults` alone covers every family).

## What Changes

- Introduce a shared metric-axis normalization primitive and apply it to each confirmed metric `config_json`/`metric_json` default path (e.g. `PopulateLensMetricDefaults`, `PopulatePieChartMetricDefaults`, `PopulateGaugeMetricDefaults`, `PopulateLegacyMetricMetricDefaults`) so 9.6 read-backs round-trip without drift in the metric fields that actually use those paths.
- Add legend `truncate_after_lines` (and, for pie, `nested`) null-preservation to `lenspie` and `lenswaffle` alignment, mirroring the existing `AlignPartitionLegendStateFromPlan` pattern already used for `legend.visible`.
- Confirm the Kibana 9.6 `value_display.percent_decimals` behavior by format first, then widen `lenscommon.PartitionValueDisplayMatchesKibanaDefault` only for the verified default shape(s) rather than hard-coding an unverified constant.
- Add a heatmap axis-label `orientation` default (`"horizontal"`) to the heatmap alignment path, parallel to the existing `labels.visible` / `title.visible` handling.
- Preserve XY reference-line threshold omit-defaults (`thresholds[].axis` `"y"`, `operation` `"static_value"`, `color_json` `{type:"auto"}`, and `value_json` object→scalar restore) on the existing `lensxy` alignment path. Already implemented in `alignment_panels.go`; covered by `TestAccResourceDashboardXYChart_layers_reference`.
- Preserve datatable `metrics[].config_json` extras Kibana 9.6 injects when omitted (`visible:true`, `alignment:"right"`) via `PopulateDatatableMetricDefaults` (`lenscommon/metric_defaults.go`, wired in `lensdatatable/{alignment,defaults}.go`).
- Preserve legacy `metric_json` extras Kibana 9.6 injects when omitted (`color:{type:"auto"}`, `size:"m"`) via `PopulateLegacyMetricMetricDefaults` (`lenscommon/populate_lens_charts.go`).
- No schema changes and no state/schema version bump — this is read-path normalization only, following the same pattern as REQ-011's prior fixes.
- Add or extend acceptance test coverage for each affected chart family so the documented defaults are exercised against a live `9.6.0-SNAPSHOT` (or later) Kibana.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: extend REQ-011 ("Panel default normalization and XY-axis drift prevention") to document and require normalization of the additional Kibana 9.6 Lens defaults listed above.

## Impact

- `internal/kibana/dashboard/lenscommon/` — add a shared metric-axis normalization helper and wire it into each confirmed metric-default population path.
- `internal/kibana/dashboard/lenscommon/partition_alignment.go` — widen `PartitionValueDisplayMatchesKibanaDefault` for `percent_decimals`.
- `internal/kibana/dashboard/panel/lenspie/alignment.go`, `internal/kibana/dashboard/panel/lenswaffle/alignment.go` — align legend `truncate_after_lines` (and pie `nested`) from plan.
- `internal/kibana/dashboard/panel/lensheatmap/` alignment path — align axis label `orientation` from plan.
- `internal/kibana/dashboard/panel/lensxy/alignment_panels.go` — align omitted reference-line threshold `axis` / `operation` / `color_json` / `value_json` from plan (see `_layers_reference`).
- `internal/kibana/dashboard/lenscommon/metric_defaults.go` and `internal/kibana/dashboard/panel/lensdatatable/{alignment,defaults}.go` — populate omitted datatable metric `visible` / `alignment` defaults.
- `internal/kibana/dashboard/lenscommon/populate_lens_charts.go` — populate omitted legacy `metric_json` `color` / `size` defaults.
- Acceptance tests across `internal/kibana/dashboard/panel/{lensxy,lensdatatable,lenspie,lenswaffle,lensmosaic,lenstreemap,lensheatmap,lenslegacymetric}` for the affected suites named in the issue (e.g. `TestAccResourceDashboardXYChart_basic`, `TestAccResourceDashboardDatatableChart`, `TestAccResourceDashboardPieChart`, `TestAccResourceDashboardWaffle`, `TestAccResourceDashboardMosaic`, `TestAccResourceDashboardTreemap`, `TestAccResourceDashboardLegacyMetricChart`, `TestAccLensMinimalProbe_*`).
- `openspec/specs/kibana-dashboard/spec.md` — sync the REQ-011 delta once implemented.
