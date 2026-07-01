## Context

The `generated/kbapi/kibana.gen.go` file already contains the three AIOps embeddable structs:

- `KibanaHTTPAPIsAiopsLogRateAnalysis` (line ~39193): `DataViewId` (required string),
  `Title`, `Description`, `HideTitle`, `HideBorder` (optional presentation), `TimeRange` (optional).
- `KibanaHTTPAPIsAiopsPatternAnalysis` (line ~39206): same base fields plus `FieldName`
  (required string), `MinimumTimeRange` (optional enum), `RandomSamplerMode` (optional enum),
  `RandomSamplerProbability` (optional float32).
- `KibanaHTTPAPIsAiopsChangePointChart` (line ~39156): `DataViewId` (required string),
  `MetricField` (required string), `AggregationFunction` (optional enum), `SplitField`
  (optional string), `Partitions` (optional `*[]string`), `MaxSeriesToPlot` (optional float32),
  `ViewType` (optional enum), plus the standard presentation fields.

The corresponding panel wrapper types (`KibanaHTTPAPIsKbnDashboardPanelTypeAiopsLogRateAnalysis`,
etc.) are also present and follow the same `{Grid, Id, Config, Type}` shape as all other typed
panels, making `panelkit.SimpleFromAPI` / `panelkit.SimpleToAPI` directly applicable.

## Goals / Non-Goals

**Goals:**
- Add `aiops_log_rate_analysis_config`, `aiops_pattern_analysis_config`, and
  `aiops_change_point_chart_config` blocks following the established panelkit pattern.
- Enforce documented API constraints at plan time (enum validation, probability range).
- Apply REQ-009 null-preservation so optional fields stay null on read when the user omitted
  them, preventing drift from Kibana server-side defaults.

**Non-Goals:**
- Panel-level drilldowns — the API models do not expose them for any of the three panels.
- Changes to Kibana ML / AIOps job or infrastructure resources.
- Reshaping existing panel handlers.

## Decisions

- **Set type for `partitions`**: The API describes `partitions` as a filter set of split-field
  values. Kibana returns them in undefined order and treats `["a","b"]` and `["b","a"]`
  identically. Using `schema.SetAttribute` prevents spurious plan drift on reorder and deduplicates
  entries. This is an intentional deviation from list semantics.

- **Plan-time probability validator**: `random_sampler_probability` is validated against the
  API-documented bound `[0.00001, 0.5]` using `float64validator.Between(0.00001, 0.5)`, producing
  a provider-owned plan-time error rather than deferring to Kibana's runtime rejection.

- **Float API fields stored as `types.Float64`**: `RandomSamplerProbability` and
  `MaxSeriesToPlot` are `float32` in the generated client. They are stored as `types.Float64` in
  TF state (standard framework type) with a `float32 ↔ float64` cast in the mapping layer.
  Precision loss is negligible for these values (`max_series_to_plot` is a small integer in
  practice; `random_sampler_probability` is bounded to `[0.00001, 0.5]`).

- **Null-preservation for optional fields**: all optional string, bool, float64, and set fields
  apply `panelkit.PreserveString` / `panelkit.PreserveBool` / null-intent checks from prior
  state, consistent with REQ-009 across the existing typed panels.

- **Bundle all three in one PR**: The panels share structural conventions and form a coherent
  AIOps panel delivery. Bundling reduces reviewer overhead and produces a cleaner release note.

## Risks / Trade-offs

- [Risk] `max_series_to_plot` round-trips with float32 rounding from the API — Mitigation:
  in practice Kibana uses integer values here (default 6); null-preserve the field when the user
  omitted it to avoid drift from API-returned values.
- [Risk] `partitions` as a set silently deduplicates identical entries — Mitigation: documented
  in the attribute description; consistent with the API's filter-set semantics.

## Open questions

None. All design decisions are agreed per the implementation choices comment on issue #4005.
