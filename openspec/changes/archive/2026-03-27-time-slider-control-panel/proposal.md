# Proposal: Time Slider Control Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners cannot manage time slider control panels on Kibana dashboards as code today. These controls provide a time window filter that slides within the dashboard's global time range, allowing users to focus-zoom on a subset of time without changing the overall dashboard time range. Time slider controls are useful for observability and operational dashboards where operators want to highlight an incident window or compare a slice of data against the full time range. Without Terraform support, dashboards that include time slider controls must be managed partially outside of Terraform, which creates configuration drift and prevents full infrastructure-as-code workflows.

## What Changes

- **Add `time_slider_control_config` typed panel config block** for panels with `type = "time_slider_control"`. This block captures all optional fields from the `kbn-dashboard-panel-time_slider_control` API schema in a structured way.
- **Add new requirement REQ-029** defining the behavior, optional fields, percentage validation, and read/write semantics of the `time_slider_control` panel type.
- **Update REQ-006** to include schema-level validation that `time_slider_control_config` is only valid when `type = "time_slider_control"`, and that `time_slider_control_config` is mutually exclusive with all other panel config blocks.

## Capabilities

After this change, practitioners will be able to:

- Declare `time_slider_control` panels on a dashboard with full control over the start and end positions of the time window as fractions of the dashboard's overall time range.
- Optionally anchor the time window via `is_anchored` so the window start is fixed and only the end slides.
- Omit the `time_slider_control_config` block entirely or omit any combination of its fields to accept Kibana's defaults.
- Import and plan-refresh existing `time_slider_control` panels without losing their configuration.

## Implementation Notes

- **Percentage fields as float32**: `start_percentage_of_time_range` and `end_percentage_of_time_range` are modeled as Terraform `float32` attributes so state matches Kibana's API representation and refresh does not introduce spurious diffs for common decimals (for example `0.1` / `0.9`) that are not exactly representable in binary floating point. Alternatives considered: rounding to a fixed decimal precision (lossy); or keeping float64 in schema while coercing read-back through float32 (smaller change but leaves authored type wider than the API).

- **`config_json` is computed read-back for `time_slider_control`, not an authored input:** the read path may populate `config_json` in state alongside `time_slider_control_config`, similar to other typed panels, but practitioners must not set `config_json` in HCL for this panel type. Schema validation on `config_json` (type allowlist: `markdown` and `lens` only) rejects `time_slider_control`; the panel object validator describes the same rule without emitting a duplicate diagnostic.

## Impact

- **Mostly additive**: adds a new optional `time_slider_control_config` block and does not change other panel types or their schemas. Dashboards that never use `time_slider_control` are unaffected.
- **Schema change**: new block on the panel object; `start_percentage_of_time_range` and `end_percentage_of_time_range` are **`float32` attributes** (Plugin Framework), not `float64`, to match Kibana’s API and avoid refresh drift.
- **REQ-006 update**: broadens the schema validation rules to cover the new panel type and config block.
- **State / upgrades**: the provider does not ship a custom Terraform `StateUpgrader` for this resource. For the **first released** shape of `time_slider_control_config`, there is no prior released state shape to migrate. Any unpublished or forked build that exposed the same logical fields as `float64` would be a **schema type change** (float64 → float32); practitioners should plan for Terraform to refresh/reconcile those attributes and should avoid relying on float64-only precision in HCL (values are coerced to float32 at plan time).
- **Compatibility**: existing configurations that do not use `time_slider_control` remain valid. Configurations that use `time_slider_control` must use the typed block (not practitioner-authored `config_json`) and percentage literals within float32 semantics.
