# Proposal: `lens-dashboard-app` Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners cannot manage `lens-dashboard-app` panels on Kibana dashboards as code today. The generated Kibana Dashboard API models expose `lens-dashboard-app` as a distinct dashboard panel discriminator from the existing `vis`/Lens panel path already supported by the provider.

The important missing capability is by-reference Lens embedding. In the current API model, a by-reference `lens-dashboard-app` panel stores the linked library item in `config.ref_id` and resolves the saved object through `config.references[]`. This lets dashboards embed shared, centrally managed Lens visualizations without duplicating full chart definitions in every dashboard.

The same API panel type also supports by-value embedding. In the current generated model, by-value `config` is not wrapped in an `attributes` object; it is directly one of the Lens chart configuration schemas such as `metricNoESQL`, `xyChartESQL`, `pieNoESQL`, or `waffleESQL`. The provider needs a dedicated typed panel block because the dashboard panel discriminator is `lens-dashboard-app`, while the payload shape is different from the existing `type = "vis"` Lens converters.

## What Changes

- **Add `lens_dashboard_app_config` typed panel config block** for panels with `type = "lens-dashboard-app"`, with mutually exclusive `by_value` and `by_reference` sub-blocks.
- **Model by-value using `by_value.config_json`**, an opaque JSON object that must match one of the generated API by-value Lens chart schemas for `KbnDashboardPanelTypeLensDashboardAppConfig0`.
- **Model by-reference** with required `by_reference.ref_id` and `by_reference.time_range`, optional `references_json` (and other API-aligned by-reference fields), matching `KbnDashboardPanelTypeLensDashboardAppConfig1`.
- **Add new requirement REQ-035** defining the current API-aligned by-value and by-reference behavior, required fields per mode, optional by-reference fields, and read/write semantics.
- **Update REQ-010 and REQ-025** so panel-level `config_json` allowlist/round-trip wording includes `lens-dashboard-app` and defers to `lens_dashboard_app_config` for that type.
- **Update REQ-006** to extend schema-level validation to include the `lens-dashboard-app` panel type and enforce that exactly one of `by_value` or `by_reference` is set.

## Capabilities

After this change, practitioners will be able to:

- Reference an existing saved Lens visualization using required `ref_id` and `time_range`, and when wiring the saved object, optionally set `references_json` (for example a reference whose `name` matches `ref_id`, `type` is `lens`, and `id` is the saved object ID).
- Configure the required by-reference `time_range`, including optional `mode`, as required by the generated API model.
- Optionally set by-reference display fields `title`, `description`, `hide_title`, and `hide_border`.
- Optionally supply `drilldowns_json` for by-reference panels, matching the generated API `drilldowns` array.
- Embed a full Lens visualization inline using `by_value.config_json`, whose JSON object is sent directly as the dashboard panel `config`.
- Import and refresh existing `lens-dashboard-app` panels without converting them into the existing `type = "vis"` Lens panel representation.

## Impact

- **Additive only**: no existing panel types or behaviors are changed. The existing `vis`-typed Lens config blocks (`xy_chart_config`, `metric_chart_config`, etc.) and panel-level `config_json` behavior for supported panel types are unaffected.
- **Schema change**: adds a new optional `lens_dashboard_app_config` block to the panel schema alongside existing typed config blocks.
- **REQ-006 update**: broadens schema validation rules to cover the new panel type and config block, including mutual exclusivity and sub-block exclusivity enforcement.
- **REQ-010 / REQ-025 update**: panel-level `config_json` rules explicitly include `lens-dashboard-app` (must use `lens_dashboard_app_config`).
- **No state migration**: new block; existing dashboard state is unaffected.
- **No breaking change**: all existing dashboards remain valid.
