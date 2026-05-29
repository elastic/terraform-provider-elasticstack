## 1. Spec Update

- [x] 1.1 Update `openspec/specs/kibana-dashboard/spec.md` REQ-013: replace the dashboard-inheritance paragraph with "omit when null"; add by-reference optionality text; replace inherited-dashboard scenario with omit scenario; add by-reference nil/set scenarios

## 2. Model Change

- [x] 2.1 In `internal/kibana/dashboard/models/lens.go`, change `VisByReferenceModel.TimeRange` from `VisByReferenceTimeRangeModel` (value) to `*VisByReferenceTimeRangeModel` (pointer)

## 3. Resolver Interface and Implementation

- [x] 3.1 In `internal/kibana/dashboard/lenscommon/iface.go`, change `Resolver.ResolveChartTimeRange` return type from `kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema` to `*kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema`
- [x] 3.2 In `internal/kibana/dashboard/lenscommon/time_range.go`, rewrite `ResolveChartTimeRange` to return nil when `chartLevel == nil`; remove dashboard inheritance and hardcoded fallback; update `TimeRangeModelToAPI` return to pointer if needed
- [x] 3.3 In `internal/kibana/dashboard/panel/visconfig/populate_vis_by_value.go`, update `chartPresentationResolver.ResolveChartTimeRange` to match new interface signature

## 4. Write Path — By-Value Charts

- [x] 4.1 In `internal/kibana/dashboard/lenscommon/presentation.go`, update `LensChartPresentationWritesFor` to assign `writes.TimeRange` only when resolver returns non-nil (nil → leave `writes.TimeRange` nil)
- [x] 4.2 Verify all typed panel `api_conv.go` files (`lensxy`, `lensmetric`, `lenslegacymetric`, `lensdatatable`, `lensgauge`, `lensheatmap`, `lensmosaic`, `lenspie`, `lensregionmap`, `lenstreemap`, `lenstagcloud`, `lenswaffle`) — `chart.TimeRange = writes.TimeRange` already passes a pointer, nil flows through `omitempty` correctly; no changes needed beyond confirming

## 5. Write Path — By-Reference Panels

- [x] 5.1 In `internal/kibana/dashboard/lenscommon/by_reference.go`, change `LensByReferenceAttributes()` `time_range` from `Required: true` to `Optional: true`; update markdown description
- [x] 5.2 In `internal/kibana/dashboard/lenscommon/by_reference.go`, update `VisByReferenceModelToAPIConfig1` to only set `api1.TimeRange` when `byRef.TimeRange != nil`
- [x] 5.3 In `internal/kibana/dashboard/lenscommon/by_reference.go`, update `PopulateVisByReferenceTFModelFromAPIConfig1` to handle nil `cfg1.TimeRange` (keep `VisByReferenceModel.TimeRange` nil when API returns none; populate only when non-nil)
- [x] 5.4 In `internal/kibana/dashboard/lenscommon/by_reference.go`, simplify `HasLensByReferenceShapeAtRoot` to detect by-reference shape using `ref_id` presence only; remove `time_range` check
- [x] 5.5 In `internal/kibana/dashboard/panel/visconfig/schema.go`, update description strings that reference "required `time_range`" for `by_reference`

## 6. Read Path Cleanup

- [x] 6.1 In `internal/kibana/dashboard/lenscommon/presentation.go`, remove the `DashboardLensComparableTimeRange` comparison block from `chartTimeRangeFromAPI` (the nil/empty guard is sufficient; the dashboard-comparison was compensating for the forced-write behavior)
- [x] 6.2 In `internal/kibana/dashboard/lenscommon/time_range.go`, remove `DashboardLensComparableTimeRange` function if no longer referenced; remove `ResolveChartTimeRange` dashboard/fallback logic
- [x] 6.3 In `internal/kibana/dashboard/lenscommon/iface.go`, remove `DashboardLensComparableTimeRange` from the `Resolver` interface if no longer called by the read path
- [x] 6.4 In `internal/kibana/dashboard/panel/visconfig/populate_vis_by_value.go`, remove `DashboardLensComparableTimeRange` from `chartPresentationResolver` if the method is dropped from the interface

## 7. Tests

- [x] 7.1 Update unit tests in `internal/kibana/dashboard/lenscommon/` (time_range_test, presentation tests) to assert nil return when chart-level is unset, and correct pointer return when set
- [x] 7.2 Update `internal/kibana/dashboard/lenscommon/detect_test.go` to cover `HasLensByReferenceShapeAtRoot` with configs that have `ref_id` but no `time_range`
- [x] 7.3 Update `internal/kibana/dashboard/panel/visconfig/api_test.go` and `fromapi_all_converters_test.go` to reflect by-reference `time_range` being optional
- [x] 7.4 Update or add converter unit tests in affected typed panel packages to assert `TimeRange` is nil in API payload when chart-level is unset
- [x] 7.5 Update acceptance tests (e.g., `lensxy/acc_test.go TestAccResourceDashboardXYChart_chartTimeRangeLifecycle`) to reflect new behavior: no inherited time range when chart-level is null; explicit time range still sent when set
- [x] 7.6 Update `visconfig/acc_test.go` by-reference test assertions to not require `time_range` when omitted from config

## 8. Build and Validate

- [x] 8.1 Run `make build` and confirm no compilation errors
- [x] 8.2 Run `make check-lint` (includes `make check-openspec`) and confirm no lint or spec errors
- [x] 8.3 Run targeted acceptance tests for affected panel types to confirm correct behavior end-to-end
