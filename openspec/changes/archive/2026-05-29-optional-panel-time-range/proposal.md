## Why

The provider unconditionally forces a `time_range` onto every typed Lens panel in the Kibana API payload — either inherited from the dashboard-level value or falling back to a hardcoded `now-15m`/`now` window. This means practitioners cannot create panels that use the dashboard's global time range, manually removing the custom time range in the Kibana UI produces no drift, and the `vis_config.by_reference` schema incorrectly marks `time_range` as a required Terraform attribute. The Kibana API has always declared `time_range` as optional (`*Type, omitempty`) on both by-value chart types and by-reference vis config, so the inheritance behavior was never necessary.

## What Changes

- **By-value Lens charts**: When a chart-level `time_range` is null in Terraform configuration, the provider SHALL omit `time_range` from the API payload entirely (no inheritance from dashboard, no hardcoded fallback).
- **By-value Lens charts**: When a chart-level `time_range` is set in configuration, the provider SHALL send it verbatim to the API as before.
- **By-reference vis panels**: `time_range` becomes optional (`Optional: true`) in the `vis_config.by_reference` Terraform schema — backward-compatible (existing configs that set `time_range` continue to work; only plan-time validation is relaxed).
- **By-reference vis panels**: `VisByReferenceModelToAPIConfig1` only sets `TimeRange` on the API payload when the model has a configured value.
- **Shape detection**: `HasLensByReferenceShapeAtRoot` detects by-reference config using `ref_id` alone; `time_range` presence is no longer part of the heuristic.
- **Read path**: The dashboard-comparable null-preservation logic (`DashboardLensComparableTimeRange`) in the chart read path is removed — it was compensating for the forced-inheritance write behavior.
- **Spec**: REQ-013 in `openspec/specs/kibana-dashboard/spec.md` is updated to require omission (not inheritance) when chart-level `time_range` is null.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: REQ-013 behavior changes — chart-level `time_range` omitted from API when null (was: inherited from dashboard or hardcoded fallback). By-reference `time_range` becomes optional in TF schema.

## Impact

- **`internal/kibana/dashboard/lenscommon/`**: `iface.go` (Resolver interface), `time_range.go` (ResolveChartTimeRange), `presentation.go` (write and read paths), `by_reference.go` (schema, write, read, shape detection)
- **`internal/kibana/dashboard/models/`**: `VisByReferenceModel.TimeRange` field pointer change
- **All typed Lens panel `api_conv.go` files**: No code changes needed — `writes.TimeRange` is already `*Type`, nil will omit via `omitempty`
- **`openspec/specs/kibana-dashboard/spec.md`**: REQ-013 update
- **Unit and acceptance tests**: Updated to remove inherited-time-range assertions; new nil-omission assertions added
