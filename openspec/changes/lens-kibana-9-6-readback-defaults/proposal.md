## Why

Against Elastic Stack `9.6.0-SNAPSHOT`, applying `elasticstack_kibana_dashboard` with a typed Lens `vis` panel fails on the very first apply with `Provider produced inconsistent result after apply`. Kibana 9.6 persists additional server-side Lens defaults on read-back that the provider's existing default-normalization logic does not yet know about, so the post-create read returns a value that differs from the planned value and Terraform rejects the apply.

This is not flaky — it reproduced on every attempt (including 5 `gotestsum` retries) for the `9.6.0-SNAPSHOT` shard-1 run on [PR #4890](https://github.com/elastic/terraform-provider-elasticstack/pull/4890) ([CI run 34672008631](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/34672008631)). The canonical failure is `TestAccResourceDashboardXYChart_basic` (`internal/kibana/dashboard/panel/lensxy/acc_test.go`), where the practitioner's Y-metric `config_json` is planned as `{"empty_as_null":true,"operation":"count"}` and Kibana's read-back returns `{"empty_as_null":true,"operation":"count","axis":"y","color":{"type":"auto"}}`.

The provider already has an established pattern for exactly this class of drift (`internal/kibana/dashboard/lenscommon`'s `PreserveNullIfStateEquals`, `PreservePlanJSONIfStateAddsOptionalKeys`, `PopulateLensMetricDefaults`, and `PartitionValueDisplayMatchesKibanaDefault`), documented in `openspec/specs/kibana-dashboard/spec.md` under "Panel default normalization and XY-axis drift prevention (REQ-011)" and originally added for issue #2355. Kibana 9.6 adds new defaults to that same class of fields that the current implementation of those helpers does not cover:

| Attribute family | Example path | Planned | 9.6 read-back |
|---|---|---|---|
| Shared Lens metric `config_json` | `xy_chart_config.layers[0].data_layer.y[0].config_json` | `{"empty_as_null":true,"operation":"count"}` | adds `"axis":"y"`, `"color":{"type":"auto"}` |
| Shared Lens metric `config_json` | `datatable_config.no_esql.metrics[0].config_json` | `{"operation":"count"}` | same shared defaults gap |
| Shared Lens metric `config_json` | `legacy_metric_config.metric_json` | practitioner blob | same shared defaults gap |
| Pie/waffle legend | `pie_chart_config.legend.truncate_after_lines` / `.nested` | null | `1` / `false` |
| Waffle legend | `waffle_config.legend.truncate_after_lines` | null | `1` |
| Partition value display | `mosaic_config` / `treemap_config.value_display.percent_decimals` | null | `2` |
| Heatmap axis labels | `heatmap_config.axis.x.labels.orientation` | null | `"horizontal"` |

`color: {type: "auto"}` is already covered by `lenscommon.PopulateLensMetricDefaults`; the new, uncovered piece is the `axis` key. Because `PopulateLensMetricDefaults` is shared by every Lens chart family that carries a metric `config_json` (XY, datatable, metric chart, legacy metric, pie, gauge, tagcloud, treemap, mosaic, region map), fixing it once should resolve the `axis`-key drift for all of them, including `legacy_metric_config.metric_json`, without a chart-specific code path.

## What Changes

- Extend `lenscommon.PopulateLensMetricDefaults` to also inject a default `axis` value into a Lens metric `config_json` when the practitioner's plan omits it, so 9.6 read-backs of every chart family that shares this helper (XY `y[]`, datatable `metrics`/`rows`/`split_metrics_by`, metric chart, legacy metric, pie, gauge, tagcloud, treemap, mosaic, region map) round-trip without drift.
- Add legend `truncate_after_lines` (and, for pie, `nested`) null-preservation to `lenspie` and `lenswaffle` alignment, mirroring the existing `AlignPartitionLegendStateFromPlan` pattern already used for `legend.visible`.
- Widen `lenscommon.PartitionValueDisplayMatchesKibanaDefault` so a Kibana-filled `percent_decimals = 2` (in addition to the already-handled `null`) is still recognized as the Kibana-injected default block for treemap/mosaic `value_display`.
- Add a heatmap axis-label `orientation` default (`"horizontal"`) to the heatmap alignment path, parallel to the existing `labels.visible` / `title.visible` handling.
- No schema changes and no state/schema version bump — this is read-path normalization only, following the same pattern as REQ-011's prior fixes.
- Add or extend acceptance test coverage for each affected chart family so the documented defaults are exercised against a live `9.6.0-SNAPSHOT` (or later) Kibana.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: extend REQ-011 ("Panel default normalization and XY-axis drift prevention") to document and require normalization of the additional Kibana 9.6 Lens defaults listed above.

## Impact

- `internal/kibana/dashboard/lenscommon/populate_lens_charts.go` — add the `axis` default to `PopulateLensMetricDefaults`.
- `internal/kibana/dashboard/lenscommon/partition_alignment.go` — widen `PartitionValueDisplayMatchesKibanaDefault` for `percent_decimals`.
- `internal/kibana/dashboard/panel/lenspie/alignment.go`, `internal/kibana/dashboard/panel/lenswaffle/alignment.go` — align legend `truncate_after_lines` (and pie `nested`) from plan.
- `internal/kibana/dashboard/panel/lensheatmap/` alignment path — align axis label `orientation` from plan.
- Acceptance tests across `internal/kibana/dashboard/panel/{lensxy,lensdatatable,lenspie,lenswaffle,lensmosaic,lenstreemap,lensheatmap,lenslegacymetric}` for the affected suites named in the issue (e.g. `TestAccResourceDashboardXYChart_basic`, `TestAccResourceDashboardDatatableChart`, `TestAccResourceDashboardPieChart`, `TestAccResourceDashboardWaffle`, `TestAccResourceDashboardMosaic`, `TestAccResourceDashboardTreemap`, `TestAccResourceDashboardLegacyMetricChart`, `TestAccLensMinimalProbe_*`).
- `openspec/specs/kibana-dashboard/spec.md` — sync the REQ-011 delta once implemented.
