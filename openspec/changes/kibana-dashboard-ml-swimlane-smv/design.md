# Design: ML Anomaly Swim Lane and Single Metric Viewer Panels

## Context

The Kibana dashboard API provides first-class typed schemas for `ml_anomaly_swimlane` and `ml_single_metric_viewer` panels. Both are already modeled in the generated `kbapi` client:

- `KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalySwimlane` (panel wrapper)
- `KibanaHTTPAPIsMlAnomalySwimlane` (union of two branches discriminated by `swimlane_type`)
- `KibanaHTTPAPIsMlAnomalySwimlane0` (overall branch: `swimlane_type = "overall"`)
- `KibanaHTTPAPIsMlAnomalySwimlane1` (viewBy branch: `swimlane_type = "viewBy"`, adds required `view_by`)
- `KibanaHTTPAPIsKbnDashboardPanelTypeMlSingleMetricViewer` (panel wrapper)
- `KibanaHTTPAPIsMlSingleMetricViewer` (flat struct with `selected_entities` map)

The provider's dashboard resource already supports ML-adjacent panels (SLO, Synthetics) using the `panelkit` helper library. Both new handlers follow the same conventions.

## Goals

- Add typed panel blocks `ml_anomaly_swimlane_config` and `ml_single_metric_viewer_config` to `elasticstack_kibana_dashboard`.
- Enforce the `view_by` discriminator constraint for the swim lane union at plan time.
- Model `selected_entities` in a way that preserves Terraform-native types (string vs number) and enforces exactly one value per entry.
- Keep the schema uniform with the `ml_anomaly_charts` sibling panel: `job_ids` is always `list(string)`.

## Non-Goals

- `ml_anomaly_charts` panel — tracked in #4000.
- Cross-referencing `job_ids` against an ML job resource — treated as opaque strings.
- Any changes to ML anomaly detection job authoring resources.

## Decisions

| Topic | Decision |
|-------|----------|
| **`ml_anomaly_swimlane` schema shape** | Flat schema with required `swimlane_type` string enum (`"overall"` \| `"viewBy"`). No nested branch blocks. Rationale: 90% of fields are shared; nesting would duplicate the entire shared field set for one branch-specific attribute (`view_by`). Matches the "flat-with-discriminator" pattern. |
| **`view_by` constraint enforcement** | Plan-time object validators on `ml_anomaly_swimlane_config`: (a) when `swimlane_type = "viewBy"`, `view_by` is required; (b) when `swimlane_type = "overall"`, `view_by` is forbidden. |
| **`job_ids` on both panels** | `ListAttribute` of `string`, minimum length 1. On `ml_single_metric_viewer`, additionally capped at length 1 by a list-size validator, matching API semantics while preserving schema uniformity. |
| **`per_page` type** | `float32` to match the API (`*float32`). |
| **`selected_detector_index` type** | `float32` to match the API (`*float32`). |
| **`selected_entities` shape** | `MapNestedAttribute` keyed by field name. Value object has two optional attributes (`string_value types.String`, `numeric_value types.Number`) with a plan-time validator enforcing exactly one per entry. Map primitive gives key uniqueness for free. Codebase precedent: `internal/kibana/osquery/schemas.go` ECSMappingSchema uses the same "map of small typed objects" pattern. |
| **`selected_entities` API serialization** | On write: if `string_value` is set, emit the string; if `numeric_value` is set, emit the float32. On read: check union discriminator (`KibanaHTTPAPIsMlSingleMetricViewer_SelectedEntities_AdditionalProperties`) using `AsKibanaHTTPAPIsMlSingleMetricViewerSelectedEntities0` / `...1`; store in the appropriate value attribute; null-preserve the other. |
| **`function_description` enum enforcement** | Enforced at plan time using `stringvalidator.OneOf("min", "max", "mean")` to constrain to known values and guard against typos. Kibana ignores this field for non-`metric` detectors; consider relaxing the validator if new values are added in future releases. |
| **Null-preservation on read** | Standard panelkit pattern: for optional presentation attributes (`title`, `description`, `hide_title`, `hide_border`, `per_page`, etc.), use `panelkit.Preserve*` helpers. If null in Terraform state, do not populate from API. |
| **`time_range` null-preservation** | Reuse the established panelkit `TimeRangeSchema` and null-preservation pattern (see REQ-040 in main spec). |
| **Handler packages** | `internal/kibana/dashboard/panel/mlanomalyswimlane/` and `internal/kibana/dashboard/panel/mlsinglemetricviewer/`. Each package follows the standard layout: `schema.go`, `model.go`, `api.go`, `acc_test.go`. |
| **`panelConfigBlock` naming** | `ml_anomaly_swimlane_config` and `ml_single_metric_viewer_config`, following the `<panel_type>_config` convention. |
| **`ValidatePanelConfig` for swimlane** | Since `ml_anomaly_swimlane_config` is required (the panel has no useful default-config state), `ValidatePanelConfig` should emit an error when the config block is absent. |
| **`ValidatePanelConfig` for SMV** | Same: `ml_single_metric_viewer_config` is required. |
| **Union dispatch on read (swimlane)** | Use `AsKibanaHTTPAPIsMlAnomalySwimlane0` / `AsKibanaHTTPAPIsMlAnomalySwimlane1` to detect branch. If `AsKibanaHTTPAPIsMlAnomalySwimlane1` succeeds (preferred check: `swimlane_type == "viewBy"`), populate `view_by` from the result. |

## Risks / Trade-offs

- **`per_page` as float32**: Terraform state will use `float32` semantics. Practitioners writing integer values like `10` are unaffected; Terraform coerces them. Values that are not exactly representable as float32 could exhibit subtle round-trip behaviour, but `per_page` is a page-size integer in practice.
- **`selected_entities` map ordering**: Terraform maps are ordered by key; the API may return entries in a different order. The provider must compare by key, not by position.
- **Union dispatch error handling**: If neither branch of `KibanaHTTPAPIsMlAnomalySwimlane` deserializes cleanly, the handler should surface an informative error diagnostic rather than silently returning a partial model.

## Open Questions

1. **Minimum Kibana version for `ml_anomaly_swimlane` and `ml_single_metric_viewer` panel types**: These types are present in the current kbapi generated client and in recent Kibana releases, but the exact minimum stack version has not been confirmed from release notes. If a minimum version is identified during implementation, add a version gate (mirroring the `alert_delay` / `flapping` pattern) with a clear diagnostic. If no minimum can be confirmed, rely on the API to reject the panel on incompatible versions.
2. **`function_description` future values**: The API documents `"min"`, `"max"`, `"mean"` for the `metric` function. If Kibana adds more values in a future release, the `stringvalidator.OneOf` would reject them. Implementation should consider using a validator with an `ExtendedValidate` or at minimum document the constraint clearly so it can be relaxed.
