## Context

The Kibana SLO `metric_custom_indicator` supports a `doc_count` aggregation type that counts documents without referencing a specific field. The generated kbapi client models this as a union discriminated by the `aggregation` key: `Metrics0` (all non-doc_count aggregations, has a required `field` string) and `Metrics1` (doc_count only, no `field`). The read path in `populateFromMetricCustomIndicator` already dispatches on this union correctly. The write path (`buildGoodMetricItem` / `buildTotalMetricItem`) and the schema both treat `field` as required, making `doc_count` unusable.

The `timeslice_metric_indicator` has an identical union structure and was already fixed: its schema marks `field` optional, and its write path switches on the aggregation value.

## Goals / Non-Goals

**Goals:**
- Make `field` optional on `metric_custom_indicator.{good,total}.metrics`.
- Fix the write path to send `Metrics1` (no `field`) for `doc_count` and `Metrics0` (with `field`) for all other aggregations.
- Add unit tests for the write path and acceptance test for end-to-end coverage.

**Non-Goals:**
- Schema-level (plan-time) cross-attribute validation via `validator.String` / `ConfigValidator` (the analogous `timeslice_metric_indicator` does not have it either; out of scope).
- Changes to `histogram_custom_indicator` (its aggregation types never include doc_count).
- Schema version bump or state migration (making an attribute optional is backwards-compatible with existing state).

## Decisions

**Dispatch on aggregation string, not union variant**

The write path switches on `metric.Aggregation.ValueString() == "doc_count"` rather than attempting to infer the intended variant from the field nullability. This mirrors the `timeslice_metric_indicator` approach and is unambiguous: the aggregation type is the canonical discriminant.

Alternative considered: check `metric.Field.IsNull()`. Rejected — null field could be an authoring mistake for a non-doc_count aggregation; the aggregation value is the source of truth.

**Reuse existing kbapi constant**

The doc_count string is compared against `kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1AggregationDocCount` (and the total equivalent), not a hardcoded `"doc_count"` literal. This keeps the code consistent with how `timeslice_metric_indicator` uses `timesliceMetricAggregationDocCount` constants defined in `constants.go`. No new constant is needed because the kbapi package already exports the typed value.

**Write-path validation for invalid field/aggregation combinations**

The write path in `buildGoodMetricItem` / `buildTotalMetricItem` returns an error for two invalid combinations that would otherwise produce silent failures:

- `aggregation == "doc_count"` **and** `field` is set: the provider ignores `field` on write but the API reads it back as null, causing a permanent plan diff. An explicit error is surfaced at apply time before any API call.
- `aggregation != "doc_count"` **and** `field` is null/empty: the provider would send `Field: ""` in `Metrics0`, resulting in an ambiguous API request. An explicit error is surfaced at apply time.

Schema-level (plan-time) `validator.String` / `ConfigValidator` enforcement is still out of scope; write-path errors provide equivalent user feedback at apply time.

## Risks / Trade-offs

[Existing configs with `field = ""`] → No risk. Existing configs that set a non-null `field` continue to work unchanged; the new dispatch only activates when `aggregation = "doc_count"`.

[Null `field` for non-doc_count aggregations] → The write path now returns an explicit error before the API call. Schema-level (plan-time) validation remains a separate, optional improvement.

[kbapi `As*` calls always succeed] → Both `AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1()` and `AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0()` succeed on any union value because they just unmarshal the raw JSON. The read path must continue to dispatch on the `aggregation` field value (already done correctly).
