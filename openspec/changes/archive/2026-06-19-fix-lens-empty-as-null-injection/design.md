## Context

Lens visualization panels (XY chart, datatable, and others) accept metric configuration as JSON (`config_json`). To avoid spurious plan drift, the provider normalizes that JSON with default-aware semantic equality. The shared helper `PopulateLensMetricDefaults` in `internal/kibana/dashboard/lenscommon/populate_lens_charts.go` is invoked on each metric map (by `lensxy/defaults.go` for XY `y[]` metrics and `lensdatatable/defaults.go` for datatable `metrics[]`).

That helper currently injects `empty_as_null = false` unconditionally for every metric map:

```go
if _, exists := model["empty_as_null"]; !exists {
    model["empty_as_null"] = false
}
```

However, the generated Kibana API schema only defines `empty_as_null` on three metric variants per chart family:

- XY: `KibanaHTTPAPIsXyYCountMetric` (`count`), `KibanaHTTPAPIsXyYSumMetric` (`sum`), `KibanaHTTPAPIsXyYUniqueCountMetric` (`unique_count`).
- Datatable: `KibanaHTTPAPIsDatatableMetricCountMetric`, `...SumMetric`, `...UniqueCountMetric` (same three operations).

The variants for `percentile`, `percentile_rank`, the `StatsMetric` operations (`average`, `min`, `max`, `median`, `standard_deviation`), and `last_value` do NOT define `empty_as_null`. Kibana validates the payload against the discriminated union and rejects unknown properties with HTTP 400. This is the root cause of issue #3707 (reporter saw `count` succeed and `percentile` fail). The reporter self-closed it as user error, but the bug is real and broader than just `percentile`.

The related helper `IsFieldMetricOperation` returns `true` for all of these operations and is reused for other legitimate defaults (`show_metric_label`, `show_array_values`), so the operation list itself must not change.

## Goals / Non-Goals

**Goals:**
- Make XY and datatable metrics using `percentile`, `percentile_rank`, stats operations, and `last_value` apply successfully (no HTTP 400).
- Keep `empty_as_null` injection (and its drift-prevention behavior) intact for `count`, `sum`, and `unique_count`.
- Avoid introducing new drift for the previously-broken operations on read-back.
- Convert the existing reproduction test into a passing regression test and add focused coverage.

**Non-Goals:**
- Changing the Kibana API schema or regenerating `kbapi`.
- Changing the membership of `IsFieldMetricOperation` (it is correct for the other defaults it gates).
- Auditing/altering `empty_as_null` behavior for chart types not implicated here unless they share the same defect (see Decisions for the audit scope).

## Decisions

### Decision 1: Gate `empty_as_null` injection by operation, not by removing operations from `IsFieldMetricOperation`

Add a small predicate to `lenscommon` that captures which operations the Kibana metric schema accepts `empty_as_null` for:

```go
func operationSupportsEmptyAsNull(operation string) bool {
    switch operation {
    case "count", "sum", "unique_count":
        return true
    default:
        return false
    }
}
```

In `PopulateLensMetricDefaults`, gate the injection on the metric's `operation`:

```go
if op, _ := model["operation"].(string); operationSupportsEmptyAsNull(op) {
    if _, exists := model["empty_as_null"]; !exists {
        model["empty_as_null"] = false
    }
}
```

**Why over alternatives:**
- *Removing `percentile`/`percentile_rank` from `IsFieldMetricOperation`*: rejected — that list also gates `show_metric_label` / `show_array_values`, which percentile legitimately receives. Removing would over-correct and could regress other charts.
- *Stripping `empty_as_null` only at request-marshal time*: rejected — leaves the normalization/semantic-equality path inconsistent with the payload, risking drift; the gate belongs in the single shared normalization helper.

### Decision 2: Apply the same gate to the shared field-metric helper, scoped to verified chart families

`populateFieldMetricLensDefaults` (used by tagcloud, region map, partition metrics) also injects `empty_as_null` when `IsFieldMetricOperation` is true. The fix audits these chart families' generated API types and applies the same `operationSupportsEmptyAsNull` gate where their schema likewise omits `empty_as_null` for the affected operations. Where a chart family's schema legitimately defines `empty_as_null` on a broader set of operations, its behavior is left unchanged. The XY and datatable paths (the ones implicated by #3707 and confirmed in the generated schema) are the mandatory scope; other families are adjusted only when their generated types confirm the same omission.

### Decision 3: Single source of truth for the supported-operation set

The supported set (`count`, `sum`, `unique_count`) is defined once in `lenscommon` so XY, datatable, and any other caller share it. This keeps the rule auditable against the generated `kbapi` types in one place.

## Risks / Trade-offs

- [Risk: A chart family's API schema does define `empty_as_null` for one of the gated operations] → Mitigation: verify against the generated `kbapi` struct fields before applying the gate to that family (Decision 2); the gate is scoped, not global.
- [Risk: Existing state for previously-"working" supported operations changes shape] → Mitigation: `count`/`sum`/`unique_count` behavior is unchanged, so no state migration is needed; the change only stops emitting a field that was always invalid for the other operations.
- [Risk: Hidden reliance on `empty_as_null` being present for percentile in tests/fixtures] → Mitigation: the only existing test is the reproduction test, which is updated; new unit tests assert presence/absence per operation.

## Migration Plan

No data or state migration required. The fix is behavior-only:
1. Implement the gate in `lenscommon` and update affected callers/audited helpers.
2. Convert `TestAccReproduceIssue3707` to assert a successful apply with no post-apply diff.
3. Add a sibling acceptance case for another previously-broken operation (e.g. `average`/`median`) and a `count` case confirming `empty_as_null` is still emitted.
4. Add unit tests for `PopulateLensMetricDefaults` covering supported vs unsupported operations.

Rollback: revert the `lenscommon` change; behavior returns to the prior (buggy) unconditional injection.
