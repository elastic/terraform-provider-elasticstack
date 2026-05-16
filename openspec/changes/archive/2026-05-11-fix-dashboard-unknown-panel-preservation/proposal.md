## Why

When the dashboard read path encounters a panel whose `type` the resource does not recognize (e.g., `discover_session`, `image`, `slo_alerts` — all defined by the API but not yet typed), the catch-all `default` branch in `mapPanelFromAPI` silently strips the panel's `id` and `grid` and reduces the panel to `{ type }` only. Refreshing or importing a dashboard that contains any of these panels therefore corrupts state, and the corruption is invisible until apply produces an unexpected diff. The resource is not yet released, but the same hazard will reappear every time Kibana introduces a new panel type.

## What Changes

- **BREAKING (pre-release)**: change the read path's catch-all branch in `mapPanelFromAPI` to preserve `id`, `grid`, `type`, and the panel's full raw API config payload in state for any unknown panel type, instead of clearing them.
- Round-trip the preserved raw payload on subsequent writes so unknown panels survive create/update/refresh cycles unchanged.
- No new user-facing schema attributes (no `unknown_panel_config` block); preservation is silent and visible only as stable state.
- Establish this as the contract every future panel-type addition can rely on (i.e., adding a typed block for a panel type that previously round-tripped as "unknown" is itself an additive change, not a migration).

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-dashboard`: extend REQ-010 (panels and `config_json` round-trip) so the read path preserves unknown panel types instead of dropping `id`/`grid`.

## Impact

- `internal/kibana/dashboard/models_panels.go` — replace destructive `default` branch in `mapPanelFromAPI` with preservation logic; mirror in the write path.
- `internal/kibana/dashboard/models_panels_test.go` — add unit tests for an unknown panel type round-trip.
- New acceptance test covering import and refresh of a dashboard containing a panel of an unrecognized type.
- No effect on existing typed panel handling.
