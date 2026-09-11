## MODIFIED Requirements

### Requirement: State preservation for fields Kibana omits or defaults (REQ-009)

When Kibana omits or defaults fields on read, the resource SHALL preserve prior Terraform intent to avoid inconsistent results and spurious drift where the implementation supports that behavior. The resource preserves the prior `time_range.mode` value already held in state or plan instead of overwriting it from read-back when the GET response does not supply a usable mode. When the GET dashboard API does not supply a usable `access_control.access_mode` value, the resource SHALL clear `access_control` in Terraform state rather than leaving a stale prior value behind. When the options block was omitted in Terraform and Kibana materializes only the default dashboard options matching the implementation's `isDashboardOptionsDefaultSet` helper (including `auto_apply_filters` and `hide_panel_borders` at their API defaults when applicable), the resource SHALL keep the `options` block null in state. When a section's prior `collapsed` value was null and Kibana returns `false`, the resource SHALL preserve null rather than forcing `false` into state.

For panel reads, the provider SHALL seed each panel from prior practitioner intent before finalizing state: from the prior plan on the post-create and post-update read-back, and from prior state on refresh. After that seed, it SHALL apply panel-type-specific alignment so Kibana-injected defaults or omitted optional values do not overwrite practitioner intent. This alignment includes preserving configured titles and descriptions when the API returns blank values, preserving ES|QL control `esql_query`, `title`, and `available_options` when the API omits them, preserving raw `config_json` when the read-back only differs by omitted optional `filters` or `query` keys, and preserving semantically equivalent optional JSON defaults such as `rank_by` in metric and tagcloud configurations.

For typed panel config blocks whose `PopulateFromAPI` receives both `pm` (the panel model being built, which callers always pass zero-valued to avoid aliasing plan pointers) and `prior` (the prior plan or state panel at the same index), the null-preservation decision for that block SHALL key on `prior.<Type>Config`, not on `pm`'s own field: `prior.<Type>Config != nil` SHALL be treated as a same-type update (honor the practitioner's null intent for optional fields), and `prior.<Type>Config == nil` SHALL be treated as creation, import, or a genuine type change: there is no prior null intent to honor, so the block SHALL be rebuilt from the API. For config blocks that are optional even when the panel type matches (`synthetics_monitors_config`, `synthetics_stats_overview_config`), the block SHALL instead be left null when the API response carries no content for it, rather than materializing an empty block. `pm`'s own field state SHALL NOT be used for this decision, since it never carries prior intent into `PopulateFromAPI`.

As of Kibana 9.5.0 GA, several typed panel config blocks' optional enum-shaped fields (for example `aiops_pattern_analysis_config.minimum_time_range` and `.random_sampler_mode`) receive concrete server-side default values on read where earlier Kibana versions returned no value at all. The provider SHALL apply REQ-009 null-preservation to these fields exactly as it does for any other Kibana-injected default: when the practitioner left the field unset (null in prior state/plan), the resource SHALL keep it null in state even though the API now returns a concrete default, rather than materializing that default and producing "Provider produced inconsistent result after apply".

The resource models only the currently supported Terraform subset of dashboard fields. Fields present in the Kibana dashboard API but not modeled by this resource — for example top-level `project_routing` — are outside this resource contract (see REQ-037 for `filters` and REQ-038 for `pinned_panels`).

The provider SHALL treat an API-returned `""` for `description` as semantically equivalent to an omitted field when prior plan/state had `description` null, restoring null in state rather than propagating the API-echoed empty string. This is an instance of REQ-009 null-preservation applied to the dashboard root `description`. This SHALL be consistent with the null/empty-string normalization already applied to XY chart `fitting.type`, `fitting.end_value`, and panel-level `time_range`.

#### Scenario: Empty-string description treated as null for null-intent practitioners

- GIVEN a practitioner has never set `description` on a dashboard (prior state: null)
- AND Kibana 9.5 returns `description: ""` on a subsequent read or post-apply read-back
- WHEN the provider applies REQ-009 null-preservation to `description`
- THEN state SHALL contain `description = null` and no drift SHALL be reported on the next plan

#### Scenario: Options omitted in config

- GIVEN Terraform configuration omitted the `options` block
- WHEN Kibana read-back contains only the dashboard option defaults
- THEN the resource SHALL keep `options` unset in state

#### Scenario: Typed panel config null-preservation keys on prior, not on pm

- GIVEN a typed panel config block (e.g. `aiops_pattern_analysis_config`) whose practitioner-authored value left an optional enum field (e.g. `minimum_time_range`) unset, applied and read back at least once (so `prior.<Type>Config` is non-nil on the next read)
- WHEN a subsequent refresh or post-update read runs against Kibana 9.5.0 GA or later, which now returns a concrete default for that field instead of omitting it
- THEN the resource SHALL keep the field null in state because `prior.<Type>Config` is non-nil (same-type update), regardless of what `pm`'s own (zero-valued) field state would otherwise suggest
- AND the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

---

### Requirement: Panel default normalization and XY-axis drift prevention (REQ-011)

The resource SHALL normalize `config_json` and typed `vis` panel data with default-aware semantic equality so Kibana-injected defaults do not cause unnecessary drift. This normalization SHALL include panel-type-specific defaults such as missing empty `filters` arrays and visualization metric/grouping defaults used by the implementation. For XY chart panels, when `axis.x.scale` was unset in configuration and Kibana returns the implicit default `ordinal`, the resource SHALL preserve the unset Terraform value instead of forcing `ordinal` into state.

For XY chart `fitting` round-trips, the resource SHALL treat an empty string returned by Kibana for `fitting.type` (which Kibana emits for some layer kinds such as `bar_stacked`) as semantically null and SHALL restore the practitioner's configured `fitting.type` from the plan. The same null-empty-string treatment SHALL apply to `fitting.end_value`. This prevents "Provider produced inconsistent result after apply" diagnostics when bar-style XY layers are used with an explicit `fitting.type` such as `"none"`.

For XY chart `decorations` round-trips on bar-style layers (e.g. `bar`, `bar_stacked`, `bar_horizontal`), Kibana injects server-side bar-styling defaults — `decorations.show_value_labels = false` and `decorations.minimum_bar_height = 1` — even when the practitioner omitted those fields. When the plan value for such a field is null and the API read-back returns the matching default, the resource SHALL preserve the null plan value in state instead of materializing the server default.

For every Lens chart block that exposes `data_source_json` (legacy_metric, region_map, gauge, heatmap, tagcloud, pie, treemap, mosaic, waffle, datatable, metric, and XY data/reference-line layers), Kibana injects certain optional keys into the read-back payload when the practitioner omits them: `"time_field":"@timestamp"`, and — as of Kibana 9.5.0 GA — `"name"` (echoing the underlying data view's display name for `data_view_spec` sources; earlier Kibana versions omitted this key). When the practitioner-authored `data_source_json` does not include one of these injected keys, the resource SHALL strip that key from state before semantic comparison and SHALL preserve the practitioner's original JSON payload.

For each Lens chart panel listed below, Kibana materializes hard-coded server defaults for optional fields when the practitioner omits them. The resource SHALL preserve the practitioner's null/unset plan value in state when the API read-back matches the documented default. The known defaults are:

- `gauge_config.styling.shape_json` defaults to `{"type":"bullet","orientation":"horizontal"}`.
- `tagcloud_config.orientation` defaults to `"horizontal"`.
- `tagcloud_config.font_size` defaults to `{min=18, max=72}` (whole block).
- `heatmap_config.axis.{x,y}.labels.visible` default to `true`.
- `heatmap_config.axis.{x,y}.title.visible` default to `false`.
- `heatmap_config.styling.cells.labels.visible` defaults to `false`.
- `heatmap_config.legend.visibility` defaults to `"visible"`.
- `pie_chart_config.label_position` defaults to `"outside"`.
- `treemap_config.legend.visible` and `mosaic_config.legend.visible` default to `"auto"`.
- `treemap_config.value_display` and `mosaic_config.value_display` default to the block `{mode="percentage", percent_decimals=null}` (whole block).

For Lens partition charts (pie `group_by[].config_json`, treemap `group_by_json`, mosaic `group_by_json`/`group_breakdown_by_json`) and Lens datatable (`metrics[].config_json`, `rows[].config_json`, `split_metrics_by[].config_json`), Kibana re-emits each `terms` dimension with the following injected default keys: `rank_by = {type="metric", metric_index=0, direction="desc"}` and `color = {mode="categorical", palette="default", mapping=[]}`. The resource SHALL populate these defaults during semantic-equality comparison so the practitioner's authored JSON round-trips without drift.

When the metric-default normalization injects the `empty_as_null` default into a Lens metric `config_json`, it SHALL inject `empty_as_null = false` ONLY for metric operations whose Kibana API schema accepts the property: `count`, `sum`, and `unique_count`. For all other operations — including `percentile`, `percentile_rank`, `average`, `min`, `max`, `median`, `standard_deviation`, `last_value`, and pipeline operations such as `formula`, `moving_average`, `cumulative_sum`, `differences`, and `counter_rate` — the resource SHALL NOT inject `empty_as_null`, because the corresponding Kibana API metric schema does not define that property and rejects the request with HTTP 400 (`Additional properties are not allowed ('empty_as_null' was unexpected)`). This rule SHALL apply uniformly to every Lens chart family whose metric normalization injects `empty_as_null` — XY (`y[].config_json`), datatable (`metrics[].config_json`), metric chart, pie, gauge, legacy metric, tagcloud, treemap, mosaic, and region map — because all of those families share the same Kibana metric schema in which only `count`, `sum`, and `unique_count` define `empty_as_null`. This gating applies to both the request payload sent to Kibana and the normalization used for semantic-equality comparison, so that operations without `empty_as_null` support neither fail on apply nor produce spurious drift.

#### Scenario: Unset XY X-axis scale

- GIVEN an XY chart panel whose configuration left `axis.x.scale` unset
- WHEN read-back from Kibana returns `axis.x.scale = "ordinal"`
- THEN the provider SHALL keep the Terraform state value unset for that field

#### Scenario: Bar-stacked XY layer with fitting.type = "none"

- GIVEN an XY chart panel with a `bar_stacked` data layer and `fitting = { type = "none" }`
- WHEN create runs and Kibana's read-back returns `fitting.type = ""` (empty string)
- THEN the provider SHALL preserve `fitting.type = "none"` in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: Bar-stacked XY layer omits decorations.show_value_labels and minimum_bar_height

- GIVEN an XY chart panel with a `bar_stacked` data layer whose `decorations` block omits `show_value_labels` and `minimum_bar_height`
- WHEN create runs and Kibana's read-back returns `decorations.show_value_labels = false` and `decorations.minimum_bar_height = 1`
- THEN the provider SHALL keep both fields null in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: data_source_json without time_field round-trips on every Lens chart

- GIVEN a Lens chart panel of any supported type whose `data_source_json` omits `time_field`
- WHEN create runs and Kibana's read-back returns the same payload with `"time_field":"@timestamp"` injected
- THEN the provider SHALL preserve the practitioner's JSON in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: data_source_json without name round-trips on every Lens chart (Kibana 9.5.0 GA)

- GIVEN a Lens chart panel of any supported type whose `data_source_json` uses a `data_view_spec` source and omits `name`
- WHEN create runs against Kibana 9.5.0 GA (or later) and its read-back returns the same payload with a `"name"` key injected (echoing the data view's display name)
- THEN the provider SHALL preserve the practitioner's JSON in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal gauge panel preserves null styling.shape_json

- GIVEN a gauge panel whose `gauge_config.styling` block omits `shape_json`
- WHEN create runs and Kibana's read-back returns `styling.shape_json = {"type":"bullet","orientation":"horizontal"}`
- THEN the provider SHALL keep `styling.shape_json` null in state and the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal tagcloud panel preserves null orientation and font_size

- GIVEN a tagcloud panel whose `tagcloud_config` omits `orientation` and `font_size`
- WHEN create runs and Kibana's read-back returns `orientation = "horizontal"` and `font_size = {min=18, max=72}`
- THEN the provider SHALL keep both fields null/unset in state and a subsequent plan SHALL show no changes

#### Scenario: Minimal heatmap panel preserves null axis, styling, and legend defaults

- GIVEN a heatmap panel whose `axis.{x,y}.labels.visible`, `axis.{x,y}.title.visible`, `styling.cells.labels.visible`, and `legend.visibility` are unset
- WHEN create runs and Kibana's read-back returns the documented defaults (`labels.visible=true`, `title.visible=false`, `cells.labels.visible=false`, `legend.visibility="visible"`)
- THEN the provider SHALL keep each of those fields null in state and a subsequent plan SHALL show no changes

#### Scenario: Minimal pie panel preserves null label_position and group_by JSON defaults

- GIVEN a pie panel whose `pie_chart_config.label_position` is unset and whose `group_by[].config_json` for a `terms` operation omits `rank_by` and `color`
- WHEN create runs and Kibana's read-back returns `label_position = "outside"` and injects the partition default keys into `group_by[].config_json`
- THEN the provider SHALL keep `label_position` null in state and SHALL preserve the practitioner's `group_by[].config_json` payload
- AND a subsequent plan SHALL show no changes

#### Scenario: Minimal treemap / mosaic panel preserves partition legend and value_display defaults

- GIVEN a treemap or mosaic panel whose `legend.visible` is unset and whose `value_display` block is omitted
- WHEN create runs and Kibana's read-back returns `legend.visible = "auto"` and a default `value_display = {mode="percentage", percent_decimals=null}` block
- THEN the provider SHALL keep `legend.visible` null and SHALL drop the injected `value_display` block from state
- AND a subsequent plan SHALL show no changes

#### Scenario: Datatable terms metrics preserve injected JSON defaults

- GIVEN a datatable panel whose `metrics[].config_json` omits `color`, `empty_as_null`, and `format`
- WHEN create runs and Kibana's read-back re-emits those keys with their documented defaults
- THEN the provider SHALL preserve the practitioner's `metrics[].config_json` payload via semantic-equality comparison
- AND a subsequent plan SHALL show no changes

#### Scenario: XY percentile metric does not inject empty_as_null

- GIVEN an XY `bar_horizontal` panel whose `y[].config_json` uses `operation = "percentile"` with a numeric `percentile` value and omits `empty_as_null`
- WHEN create runs and the provider builds the Kibana API request
- THEN the request payload SHALL NOT contain `empty_as_null` for that metric and Kibana SHALL accept the request (no HTTP 400)
- AND a subsequent plan SHALL show no changes

#### Scenario: XY count metric still injects empty_as_null

- GIVEN an XY panel whose `y[].config_json` uses `operation = "count"` and omits `empty_as_null`
- WHEN create runs and the provider builds the Kibana API request and reads the panel back
- THEN the provider SHALL inject the `empty_as_null = false` default for that metric and the metric SHALL round-trip without drift

---

### Requirement: ML anomaly charts panel behavior (REQ-053)

The resource SHALL support `type = "ml_anomaly_charts"` panels through the typed `ml_anomaly_charts_config` block. When a panel entry sets `type = "ml_anomaly_charts"`, the resource SHALL require the `ml_anomaly_charts_config` block and SHALL return an error diagnostic when it is absent.

The block accepts the following attributes:

- `job_ids` (required `list(string)`, min 1 item): one or more anomaly-detection job IDs or group IDs whose results are shown. The provider treats these as opaque strings and does not validate their existence against Kibana's ML API at plan time; invalid job IDs surface as Kibana API errors during `terraform apply`.
- `max_series_to_plot` (optional int64): maximum number of anomaly series to plot. When null in state, the attribute is omitted from the API request. The Kibana API represents this field as a JSON number (`*float32` in the generated client); the provider exposes it as an integer since a series count cannot be fractional, converting to/from the API's numeric type at the boundary.
- `severity_threshold` (optional list of objects, min 1 item when present): filters which severity bands are displayed. Each list item is a union — exactly one of the following must be set per item:
  - `severity` (string, one of `low`, `warning`, `minor`, `major`, `critical`): a named severity shortcut. The model layer SHALL expand named severities to their canonical `{min, max}` API pairs at write time.
  - `min` (int64) plus optional `max` (int64): an alternative, numeric spelling of one of the canonical severity bands. The Kibana API models `severity_threshold` items as a union of exactly five members, each pinning **both** `min` and `max` to a single canonical pair (see the canonical-band table below); the `critical` member has no `max` field at all. The provider does not validate this constraint client-side, so on Kibana 9.5.0 GA and later (the pre-GA `9.5.0-SNAPSHOT` build accepted arbitrary `min`/`max` values) any `{min, max}` pair that does not exactly match one of the five canonical pairs — including a canonical `min` combined with a non-canonical `max`, or any `max` set alongside `min = 75` — SHALL be rejected by the Kibana API at `terraform apply` time with an HTTP 400 error, not caught earlier as a plan-time diagnostic. `max` may be set only when `min` is set and `severity` is unset; when `max` is set, `min` must also be set. Setting both `severity` and `min` on the same item SHALL produce an error diagnostic at plan time. Setting `severity` together with `max` SHALL produce an error diagnostic at plan time.
- `title` (optional string): panel title. Subject to REQ-009 null-preservation.
- `description` (optional string): panel description. Subject to REQ-009 null-preservation.
- `hide_title` (optional bool): when true, hides the panel title. Subject to REQ-009 null-preservation.
- `hide_border` (optional bool): when true, hides the panel border. Subject to REQ-009 null-preservation.
- `time_range` (optional object: `from` string required, `to` string required, `mode` string optional): a panel-level time range override, identical in shape to the dashboard root `time_range`. Reuses `panelkit.TimeRangeSchema`. Subject to REQ-009 null-preservation: when prior state has `time_range` null, the provider SHALL keep it null even if the API returns a default; when `mode` is null in prior state, the provider SHALL keep `mode` null.

The model layer SHALL expand named severity values to the following canonical `{min, max}` API pairs (matching the generated Kibana OpenAPI const values in `KibanaHTTPAPIsMlAnomalyChartsSeverityThreshold0`–`SeverityThreshold4`):

| `severity` | API `min` | API `max`    |
|---|---|---|
| `low`       | 0         | 3            |
| `warning`   | 3         | 25           |
| `minor`     | 25        | 50           |
| `major`     | 50        | 75           |
| `critical`  | 75        | (omitted — open-ended upper bound) |

The raw `min`/`max` form exists so a practitioner can spell one of these same five bands numerically instead of via the `severity` enum; it is not a general-purpose custom-range escape hatch on Kibana 9.5.0 GA and later, which validates the `{min, max}` pair against the canonical band list server-side.

On write, the provider SHALL map `ml_anomaly_charts_config` to the `config` object in the `KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts` API schema. Optional fields SHALL be included only when set in state; absent optional fields SHALL NOT be sent to the API.

On read, the provider SHALL repopulate `ml_anomaly_charts_config` from the API response using REQ-009 null-preservation, extended to the **representation form** of `severity_threshold` items. The API encodes `severity_threshold` as `{min, max}` pairs only; it conveys no information about whether the practitioner authored a named `severity` or a raw numeric range. Therefore the chosen form is recovered from prior state, not inferred by normalizing:

- When the prior item holds a named `severity` (and `min`/`max` are null), the provider SHALL store the named form, deriving the label from the API `{min, max}` pair via the canonical-band table. The `critical` band (API: `{min: 75}`, no `max` field) SHALL map to `severity = "critical"`. If the API value no longer matches any canonical band, the provider SHALL fall back to the raw `min`/`max` form (surfacing as drift).
- When the prior item holds raw `min`/`max` (and `severity` is null), the provider SHALL store the raw `min`/`max` verbatim from the API, even when the pair coincidentally equals a canonical band.
- On import (no prior state), the provider SHALL default to the named form when the API `{min, max}` matches a canonical band, and to the raw form otherwise.

The provider SHALL NOT normalize a practitioner-authored raw range into a named `severity` on read. While the configured values match current state, a subsequent plan SHALL show no changes.

`config_json` SHALL NOT be supported for `ml_anomaly_charts` panels; using `config_json` with `type = "ml_anomaly_charts"` SHALL return an error diagnostic (per REQ-010 policy).

Implementation: new package `internal/kibana/dashboard/panel/mlanomalycharts/` with `schema.go`, `model.go`, and `api.go`; new model file `internal/kibana/dashboard/models/mlanomalycharts.go`; registration in `internal/kibana/dashboard/schema.go` and `internal/kibana/dashboard/registry.go`.

#### Scenario: Creation of ml_anomaly_charts panel with named severities

- GIVEN a panel with `type = "ml_anomaly_charts"` and `ml_anomaly_charts_config` containing `job_ids = ["my-job"]` and `severity_threshold = [{ severity = "critical" }, { severity = "major" }]`
- WHEN create runs
- THEN the provider SHALL send `job_ids = ["my-job"]` and `severity_threshold = [{min: 75}, {min: 50, max: 75}]` in the API request
- AND after the post-apply read, state SHALL represent both items as named severities
- AND a subsequent plan SHALL show no changes

#### Scenario: Round-trip stability for critical (open-ended) severity

- GIVEN a panel with `severity_threshold = [{ severity = "critical" }]` applied and read back
- WHEN the API returns `severity_threshold = [{min: 75}]` (no `max` field)
- THEN the provider SHALL map this to `severity = "critical"` in state
- AND a subsequent plan SHALL show no changes

#### Scenario: Raw range escape hatch

- GIVEN a panel with `severity_threshold = [{ min = 10, max = 20 }]` (a `{min, max}` pair that does not match any of the five canonical pairs)
- WHEN create runs against Kibana 9.5.0 GA or later
- THEN the Kibana API SHALL reject the request with an HTTP 400 error identifying `severity_threshold` as invalid
- AND `terraform apply` SHALL fail with that error, since the provider does not validate the canonical-boundary constraint client-side and this form is not a general-purpose custom-range escape hatch on supported GA versions

#### Scenario: Raw range coinciding with a canonical band is preserved (no diff)

- GIVEN a panel where the practitioner set `severity_threshold = [{ min = 3, max = 25 }]` (coincides with the `warning` canonical band)
- WHEN create runs and the post-apply read returns `{min: 3, max: 25}`
- THEN the provider SHALL store `min = 3` and `max = 25` in state (NOT coerced to `severity = "warning"`)
- AND a subsequent plan SHALL show no changes

#### Scenario: severity_threshold form is preserved across refresh

- GIVEN state holds `severity_threshold = [{ severity = "major" }, { min = 25, max = 50 }]`
- WHEN a refresh runs and the API returns `[{min: 50, max: 75}, {min: 25, max: 50}]`
- THEN state SHALL retain the first item as `severity = "major"` and the second as `min = 25, max = 50`
- AND a subsequent plan SHALL show no changes

#### Scenario: critical severity preserved in raw form when authored raw

- GIVEN a panel where the practitioner set `severity_threshold = [{ min = 75 }]` (raw form, coincides with the `critical` canonical band)
- WHEN create runs and the post-apply read returns `{min: 75}` (no `max` field)
- THEN the provider SHALL store `min = 75` with `max` null in state (NOT coerced to `severity = "critical"`)
- AND a subsequent plan SHALL show no changes

#### Scenario: Switching severity form is a configuration change

- GIVEN state holds `severity_threshold = [{ severity = "warning" }]`
- WHEN the configuration changes to `{ min = 3, max = 25 }` (same band, raw form)
- THEN the plan SHALL report a change for that item
- AND after apply the state SHALL settle to `{ min = 3, max = 25 }` with a subsequent plan showing no changes

#### Scenario: Import defaults to named form for canonical bands

- GIVEN an existing panel whose API `severity_threshold` is `[{min: 3, max: 25}]` and no prior Terraform state
- WHEN the panel is imported
- THEN state SHALL store `severity = "warning"` (named form preferred only on import, where no practitioner form exists to preserve)

#### Scenario: Plan-time validation — both severity and min set

- GIVEN a `severity_threshold` item with both `severity = "major"` and `min = 50`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating that exactly one of `severity` or `min` must be set

#### Scenario: Plan-time validation — max without min

- GIVEN a `severity_threshold` item with `max = 75` but neither `severity` nor `min` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic

#### Scenario: config_json rejected for ml_anomaly_charts

- GIVEN a panel with `type = "ml_anomaly_charts"` and `config_json = "{}"`
- WHEN Terraform validates or applies the configuration
- THEN the resource SHALL return an error diagnostic indicating that `config_json` is not supported for `ml_anomaly_charts` panels

#### Scenario: Optional fields follow null-preservation

- GIVEN an `ml_anomaly_charts_config` that does not set `max_series_to_plot` or `time_range`
- WHEN apply runs and the post-apply read returns server-side defaults for those fields
- THEN state SHALL keep `max_series_to_plot` and `time_range` as null
- AND a subsequent plan SHALL show no changes

#### Scenario: Update job_ids in-place

- GIVEN an existing `ml_anomaly_charts` panel with `job_ids = ["job-a"]`
- WHEN the configuration changes to `job_ids = ["job-a", "job-b"]` and update runs
- THEN the resource SHALL NOT destroy and recreate the dashboard
- AND a subsequent plan SHALL show no changes
