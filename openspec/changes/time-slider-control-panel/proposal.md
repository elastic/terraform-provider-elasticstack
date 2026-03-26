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

- **Dual `config_json` population**: consistent with all other typed panel implementations (`markdown`, `lens`, etc.), the read path populates `config_json` in state alongside `time_slider_control_config`. This is a supported pattern that allows practitioners to switch to the raw `config_json` workflow without triggering a plan diff; it is not a bug.

## Impact

- **Additive only**: no existing panel types or behaviors are changed.
- **Schema change**: adds a new optional `time_slider_control_config` block to the panel schema alongside existing typed config blocks.
- **REQ-006 update**: broadens the schema validation rules to cover the new panel type and config block.
- **No state migration**: new block; existing dashboard state is unaffected.
- **No breaking change**: all existing dashboards remain valid.
