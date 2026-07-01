## Why

The `elasticstack_kibana_dashboard` resource cannot embed an `apm_service_map` panel using typed configuration. SRE and observability teams that assemble dashboards combining APM service maps with SLOs, logs, and metrics must fall back to raw `config_json` blobs, which sacrifices type validation and drift-safe planning.

The Kibana Dashboard API exposes a first-class `apm_service_map` panel type with a well-defined flat configuration surface (`KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap` / `KibanaHTTPAPIsApmServiceMapEmbeddable`). The panel anchors on APM concepts (service selectors, filters, layout) rather than a data view, making it distinct from Lens-based panels.

## What Changes

Add a typed `apm_service_map_config` block to the `elasticstack_kibana_dashboard` resource alongside existing typed panel blocks. The block exposes:

- **Service selectors** (all optional strings): `environment`, `service_name`, `service_group_id`
- **Query** (optional string): `kuery` (KQL only)
- **Layout** (optional): `map_orientation` (enum: `horizontal` | `vertical`), `sync_with_dashboard_filters` (bool)
- **Filter lists** (optional sets of validated strings):
  - `alert_status_filter`: `active` | `delayed` | `recovered` | `untracked`
  - `anomaly_severity_filter`: `low` | `warning` | `minor` | `major` | `critical` | `unknown`
  - `connection_filter`: `connected` | `orphaned`
  - `slo_status_filter`: `degrading` | `healthy` | `noData` | `violated`
- **Standard passthroughs** via `panelkit.PanelPresentationAttributes()`: `title`, `description`, `hide_title`, `hide_border`
- **Time range**: `time_range` (optional object `{ from, to }`)

The filter attributes are flat at the block root (no invented `filters {}` sub-block) and modelled as `SetAttribute` (not `ListAttribute`) because Kibana treats them as unordered membership filters.

`kuery` is a plain `StringAttribute` — the API defines it as KQL-only; wrapping it in an object would invent structure the API does not have.

No mutual-exclusion validators are added between `environment`, `service_name`, and `service_group_id` — the API does not document mutual exclusion, and Kibana allows free combination.

## Capabilities

### New Capabilities
None (this change extends the existing `kibana-dashboard` capability).

### Modified Capabilities
- `kibana-dashboard`: add REQ-047 defining `apm_service_map_config` panel support.

## Impact

- New package `internal/kibana/dashboard/panel/apmservicemap/` — schema, model, API conversion, acceptance tests.
- `internal/kibana/dashboard/registry.go` — register the new handler.
- `internal/kibana/dashboard/schema.go` — add `apm_service_map_config` to the panel schema and mutual-exclusion validation.
- `internal/kibana/dashboard/models/panel_models.go` (or equivalent) — add `ApmServiceMapConfigModel`.
- Acceptance test data under `internal/kibana/dashboard/panel/apmservicemap/testdata/`.
