## Context

The Kibana SLO `metric_custom_indicator` supports a `doc_count` aggregation type that counts documents without referencing a specific field. The generated kbapi client models this as a union discriminated by the `aggregation` key: `Metrics0` (all non-doc_count aggregations, has a required `field` string) and `Metrics1` (doc_count only, no `field`). The read path in `populateFromMetricCustomIndicator` already dispatches on this union correctly. The write path (`buildGoodMetricItem` / `buildTotalMetricItem`) and the schema both treat `field` as required, making `doc_count` unusable.

The `timeslice_metric_indicator` has an identical union structure and was already fixed: its schema marks `field` optional, and its write path switches on the aggregation value.

## Goals / Non-Goals

**Goals:**
- Make `field` optional on `metric_custom_indicator.{good,total}.metrics`.
- Fix the write path to send `Metrics1` (no `field`) for `doc_count` and `Metrics0` (with `field`) for all other aggregations.
- Add unit tests for the write path and acceptance test for end-to-end coverage.

**Non-Goals:**
- Validating that `field` is present for non-doc_count aggregations at plan time (the API returns a clear error; schema-level cross-attribute validation is out of scope).
- Changes to `histogram_custom_indicator` (its aggregation types never include doc_count).
- Schema version bump or state migration (making an attribute optional is backwards-compatible with existing state).

## Decisions

**Dispatch on aggregation string, not union variant**

The write path switches on `metric.Aggregation.ValueString() == "doc_count"` rather than attempting to infer the intended variant from the field nullability. This mirrors the `timeslice_metric_indicator` approach and is unambiguous: the aggregation type is the canonical discriminant.

Alternative considered: check `metric.Field.IsNull()`. Rejected — null field could be an authoring mistake for a non-doc_count aggregation; the aggregation value is the source of truth.

**Reuse existing kbapi constant**

The doc_count string is compared against `kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1AggregationDocCount` (and the total equivalent), not a hardcoded `"doc_count"` literal. This keeps the code consistent with how `timeslice_metric_indicator` uses `timesliceMetricAggregationDocCount` constants defined in `constants.go`. No new constant is needed because the kbapi package already exports the typed value.

**No schema-level cross-attribute validation**

Adding a `validator.String` that requires `field` when aggregation is not `doc_count` would be correct but is outside the scope of this fix (the analogous `timeslice_metric_indicator` does not have it either). The API's own 400 error provides sufficient feedback.

## Risks / Trade-offs

[Existing configs with `field = ""`] → No risk. Existing configs that set a non-null `field` continue to work unchanged; the new dispatch only activates when `aggregation = "doc_count"`.

[Null `field` for non-doc_count aggregations] → The API will return a 400 error. This is acceptable; adding plan-time validation is a separate, optional improvement.

[kbapi `As*` calls always succeed] → Both `AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1()` and `AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0()` succeed on any union value because they just unmarshal the raw JSON. The read path must continue to dispatch on the `aggregation` field value (already done correctly).
