# Proposal: `lens-dashboard-app` Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners cannot manage `lens-dashboard-app` panels on Kibana dashboards as code today. The `lens-dashboard-app` panel type is a distinct panel type identifier used by the Kibana dashboard API that differs from the existing `lens` type already supported by the provider. Its most important new capability is **by-reference** embedding: a `lens-dashboard-app` panel can reference a saved Lens visualization by its saved object ID rather than embedding the full visualization definition inline. This allows dashboards to embed shared, centrally-managed Lens visualizations that may be maintained by other teams or tooling outside Terraform.

The existing `lens` panel type (already supported via typed config blocks such as `xy_chart_config`, `metric_chart_config`, etc., and via `config_json`) always embeds the full visualization definition inline (by-value). There is no existing mechanism in the provider to reference a pre-existing saved Lens object by ID. Teams that share common visualizations across multiple dashboards today must either duplicate the full visualization definition in each dashboard or manage those panels manually in the Kibana UI. Without Terraform support, this creates configuration drift, limits composability, and prevents full infrastructure-as-code workflows for dashboard-heavy teams.

`lens-dashboard-app` panels also support a by-value mode (embedding the full `attributes` inline), which provides an alternative path for Lens visualizations that is distinct from the existing `lens` type. Both modes use the type string `lens-dashboard-app` in the dashboard API, which is why they are handled separately from the existing `lens`-typed config blocks.

## What Changes

- **Add `lens_dashboard_app_config` typed panel config block** for panels with `type = "lens-dashboard-app"`. This block captures both by-value and by-reference Lens panel configurations in mutually exclusive sub-blocks.
- **Add new requirement REQ-035** defining the by-value and by-reference behavior, required fields per mode, optional shared fields, and read/write semantics of the `lens-dashboard-app` panel type.
- **Update REQ-025** to clarify that `config_json` write support does not extend to `lens-dashboard-app`; the `lens-dashboard-app` panel type SHALL be managed exclusively through the typed `lens_dashboard_app_config` block.
- **Update REQ-006** to extend schema-level validation to include the `lens-dashboard-app` panel type: `lens_dashboard_app_config` SHALL be valid only for `type = "lens-dashboard-app"`, SHALL be mutually exclusive with all other panel config blocks, and SHALL enforce that exactly one of `by_value` or `by_reference` is set.

## Capabilities

After this change, practitioners will be able to:

- Reference an existing saved Lens visualization by saved object ID using `lens_dashboard_app_config.by_reference.saved_object_id`, embedding it into a dashboard without duplicating the full visualization definition.
- Optionally override the display title, description, and border/title visibility for by-reference panels.
- Optionally supply `overrides_json` for by-reference panels to customize the embedded saved Lens object's behavior.
- Embed a full Lens visualization inline using `lens_dashboard_app_config.by_value.attributes_json`, with optional data view references via `references_json`.
- Configure a panel-level `time_range` for either mode to scope the visualization's time window independently of the dashboard.
- Import and plan-refresh existing `lens-dashboard-app` panels without losing their configuration.

## Impact

- **Additive only**: no existing panel types or behaviors are changed. The existing `lens`-typed config blocks (`xy_chart_config`, `metric_chart_config`, etc.) and `config_json` for `lens` panels are unaffected.
- **Schema change**: adds a new optional `lens_dashboard_app_config` block to the panel schema alongside existing typed config blocks.
- **REQ-006 update**: broadens schema validation rules to cover the new panel type and config block, including mutual exclusivity and sub-block exclusivity enforcement.
- **REQ-025 update**: explicitly documents that `lens-dashboard-app` is not in the `config_json`-supported set.
- **No state migration**: new block; existing dashboard state is unaffected.
- **No breaking change**: all existing dashboards remain valid.
