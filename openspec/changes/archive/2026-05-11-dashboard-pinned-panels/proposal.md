## Why

The Kibana Dashboard API exposes `kbn-dashboard-data.pinned_panels` — the dashboard's persistent control bar (option-list dropdowns, range sliders, time slider, and ES|QL controls pinned to the top of the dashboard). The Terraform resource handles each of those four control types as in-grid panels but offers no way to manage the dashboard-level pinned control bar, so practitioners cannot author it as code.

## What Changes

- Add an optional `pinned_panels` attribute at the dashboard root: `pinned_panels = list(object({ type, options_list_control_config, range_slider_control_config, time_slider_control_config, esql_control_config }))`.
- Reuse the four existing typed control config blocks (no new schema for control bodies).
- Discriminator validation: exactly one config block per `pinned_panels` entry, and it must match `type`. Mirror the validators used by `panels[].*_control_config`.
- Wire `pinned_panels` into create, update, and read paths; preserve order; reuse semantic-equality normalization that already covers the typed control configs.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-dashboard`: add a new requirement covering dashboard-root `pinned_panels` round-trip and shared-with-panels validation rules.

## Impact

- `internal/kibana/dashboard/schema.go` — add `pinned_panels` block list reusing the four `*_control_config` schemas.
- `internal/kibana/dashboard/models.go` — extend `dashboardModel` with `PinnedPanels` and map to/from the API.
- New validators (or reuse of existing ones) for the discriminator.
- New unit tests for discriminator validation and round-trip.
- New acceptance test exercising one or two pinned controls.
- Soft dependency on the `expand-control-fields` change: pinned panels inherit any new control fields landed there for free; if `expand-control-fields` lands first, this change benefits with no extra work.
