## Why

The `elasticstack_kibana_dashboard` resource cannot embed any of the three Kibana AIOps panel
types: `aiops_log_rate_analysis`, `aiops_pattern_analysis`, and `aiops_change_point_chart`.
Users assembling analytical dashboards that surface AIOps insights must fall back to raw
`config_json` blobs, losing typed validation and drift-safe planning. All three panels follow
the flat, data-view-anchored model and the panelkit passthrough conventions already established
for other typed panels on this resource, making this a straightforward addition.

## What Changes

Three new typed panel config blocks are added to `elasticstack_kibana_dashboard`:

- `aiops_log_rate_analysis_config` — anchored to a data view, minimal field set.
- `aiops_pattern_analysis_config` — extends log rate analysis with `field_name`, optional
  `minimum_time_range`, `random_sampler_mode`, and `random_sampler_probability` (validated to
  the API-documented range `[0.00001, 0.5]`).
- `aiops_change_point_chart_config` — anchored to a data view and `metric_field`, with optional
  `aggregation_function`, `split_field`, `partitions` (as a set to prevent order-drift),
  `max_series_to_plot`, and `view_type`.

All three blocks share the standard panelkit presentation passthroughs: `title`, `description`,
`hide_title`, `hide_border`, and `time_range`.

## Capabilities

### New Capabilities

None (all changes are additive fields on the existing `kibana-dashboard` capability).

### Modified Capabilities

- `kibana-dashboard`: add REQ-050 (AIOps log rate analysis panel), REQ-051 (AIOps pattern
  analysis panel), and REQ-052 (AIOps change point chart panel).

## Implementation Approach

Each panel follows the established panelkit handler pattern used by `sloburnrate` and
`syntheticsmonitors`:

1. One handler package per panel type under
   `internal/kibana/dashboard/panel/aiopslograteanalysis/`,
   `internal/kibana/dashboard/panel/aiopspatternanalysis/`, and
   `internal/kibana/dashboard/panel/aiopschangepointchart/`.
2. Each package exports `Handler{}`, `SchemaAttribute()`, `BuildConfig()`, and
   `PopulateFromAPI()` with REQ-009 null-preservation semantics.
3. Registration in `panelHandlers` in `internal/kibana/dashboard/registry.go`.
4. `PanelModel` extended with three new config fields in
   `internal/kibana/dashboard/models/panel.go`.

Design decisions (agreed in issue #4005):
- All three panels ship in one PR.
- `partitions` on `aiops_change_point_chart_config` uses `SetAttribute` (order-insensitive,
  semantically a filter set).
- `random_sampler_probability` gets a plan-time `float64validator.Between(0.00001, 0.5)`
  validator.
- Enum fields use `stringvalidator.OneOf` matching all other enums on the resource.

## Impact

- `internal/kibana/dashboard/registry.go` — register three new handlers.
- `internal/kibana/dashboard/models/panel.go` — add three config model fields.
- `internal/kibana/dashboard/panel/aiopslograteanalysis/` — new package.
- `internal/kibana/dashboard/panel/aiopspatternanalysis/` — new package.
- `internal/kibana/dashboard/panel/aiopschangepointchart/` — new package.
- `openspec/changes/dashboard-aiops-panels/specs/kibana-dashboard/spec.md` — delta spec adding REQ-050, REQ-051, REQ-052.
