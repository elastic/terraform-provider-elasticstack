## Why

Each of the four typed control panel schemas (`options_list_control`, `range_slider_control`, `time_slider_control`, `esql_control`) carries panel-level layout fields `width` (`small`/`medium`/`large`) and `grow` (bool) that the Terraform resource does not expose. These determine the control's footprint in the control bar and are routinely set by Kibana users. A field-by-field audit of each control schema may surface additional small gaps to close in the same change.

## What Changes

- Add `width` (string, enum `small`/`medium`/`large`) and `grow` (bool) attributes to each of the four `*_control_config` blocks.
- Audit each control schema (`options_list_control_config`, `range_slider_control_config`, `time_slider_control_config`, `esql_control_config`) against the latest API spec; close any other narrow gaps found (likely candidates: missing optional `display_settings` sub-fields, missing validation flags). Capture the audit results as additional schema additions in the same change.
- Apply REQ-009 null-preservation semantics consistently: optional new fields stay null on read when prior state had them null, even if Kibana returns server-side defaults.
- Because pinned controls reuse these schemas (see the `dashboard-pinned-panels` change), pinned controls inherit the additions automatically.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-dashboard`: extend REQ-026 (ES|QL control), REQ-027 (options list control), REQ-028 (range slider control), and REQ-029 (time slider control) to add `width` and `grow` and any other gaps surfaced by the audit.

## Impact

- `internal/kibana/dashboard/schema.go` — add `width` and `grow` (and any audit additions) to each `*_control_config` schema.
- Per-control model files (`models_options_list_control_panel.go`, etc.) — extend models and read/write mapping.
- Per-control unit tests — add coverage for the new fields and null-preservation.
- Acceptance tests — add at least one test exercising `width` and `grow` per control.
- Coupled with `dashboard-pinned-panels`: ordering does not matter for correctness; whichever lands second picks up the other's improvements automatically.
