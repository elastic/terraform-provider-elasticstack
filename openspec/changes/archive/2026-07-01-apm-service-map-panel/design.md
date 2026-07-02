## Context

The Kibana Dashboard API defines `KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap` (at line ~46717 of `generated/kbapi/kibana.gen.go`), which wraps a `KibanaHTTPAPIsApmServiceMapEmbeddable` config object (at line ~39434). The embeddable is a flat struct with no union variants; this makes it a natural fit for the `panelkit.SimpleFromAPI` / `panelkit.SimpleToAPI` handler pattern already used by `sloburnrate`, `syntheticsstatsoverview`, `sloalerts`, and others.

The four filter attributes use slice-of-enum types in the generated Go client (`*[]KibanaHTTPAPIsApmServiceMapEmbeddableAlertStatusFilter`, etc.). On the Terraform side they are modelled as `SetAttribute` of validated strings to prevent spurious drift from order changes; Kibana treats these as membership filters with no guaranteed order in responses. This matches the reasoning used for `partitions` on `aiops_change_point_chart` (referenced in the issue comments for issue #4005).

The `time_range` field in the embeddable uses `KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema`, the same shared time-range type used by other panels — it maps to the existing `{ from, to }` sub-block shape.

## Goals / Non-Goals

**Goals:**
- Expose `apm_service_map` as a typed `apm_service_map_config` block, reaching parity with the API surface.
- Filter attributes are sets (not lists) to prevent spurious plan drift from re-ordering.
- Apply REQ-009 null-preservation on all optional fields (omit from state when not set by practitioner and API returns a default).
- Enum values on filter attributes and `map_orientation` are validated at plan time with `stringvalidator.OneOf`.
- Register the handler in the panelkit registry so `config_json` is rejected for this panel type.

**Non-Goals:**
- Panel-level drilldowns — the `KibanaHTTPAPIsApmServiceMapEmbeddable` struct has no drilldowns field.
- APM service group authoring — `service_group_id` is accepted as an opaque string reference.
- Any changes to APM, ML, or AIOps panel types tracked in separate issues.

## Decisions

- **Package name**: `apmservicemap` under `internal/kibana/dashboard/panel/apmservicemap/`. Consistent with the naming convention of `sloburnrate`, `syntheticsstatsoverview`.
- **Block name**: `apm_service_map_config` — matches the Terraform naming convention for typed panel config blocks.
- **Filter modelling**: `SetAttribute` of `types.String`, validated element-by-element with `stringvalidator.OneOf`. Empty set → omit from API payload. Non-empty set → send as slice in API payload.
- **`kuery` as string**: plain `StringAttribute`. The API field is always KQL; inflating it into `{ language, text }` would invent structure.
- **No mutual exclusion on service selectors**: `environment`, `service_name`, and `service_group_id` are all independently optional. Kibana does not document mutual exclusion; practitioners may supply any combination.
- **`sync_with_dashboard_filters`**: exposed as optional bool; null → omit from API payload; REQ-009 null-preservation on read.
- **`time_range`**: optional `{ from, to }` sub-block reusing the same schema shape as other typed panels that expose `time_range`.
- **Handler interface**: implements `iface.Handler` using `panelkit.SimpleFromAPI` / `panelkit.SimpleToAPI` + dedicated `BuildConfig` / `PopulateFromAPI` helpers, consistent with `syntheticsstatsoverview`.
- **ValidatePanelConfig**: returns nil (no cross-attribute validator needed beyond the schema-level enum validators).

## Risks / Trade-offs

- [Risk] Kibana returns `alert_status_filter` / other filter arrays in an unspecified order → Mitigation: `SetAttribute` ensures plan stability regardless of API return order. The round-trip unit test must verify that swapped-order responses produce no plan change.
- [Risk] `time_range` null-preservation edge cases (API returns dashboard-inherited time range) → Mitigation: follow the existing REQ-009 pattern; when prior state had `time_range` null and the API echoes a value, state SHALL remain null.
- [Risk] `service_group_id` is an opaque string referencing a resource that may be deleted out-of-band → Mitigation: treat as opaque string; no foreign-key validation. Kibana itself enforces the reference integrity at render time.

## Open questions

None. All design decisions were settled in the issue comments by @tobio on 2026-07-01 (see implementation-choices comment on issue #4007).
