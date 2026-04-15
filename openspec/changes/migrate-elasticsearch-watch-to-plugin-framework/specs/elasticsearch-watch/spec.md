## ADDED Requirements

### Requirement: SDK-to-Framework watch state compatibility (REQ-028)
After the watch resource is migrated to the Terraform Plugin Framework, the resource SHALL continue to manage state created by the last SDK-backed provider release without requiring import, recreation, or changes to the configured resource type. The migrated resource SHALL preserve the existing composite `id` format, `watch_id` semantics, and import identifier format.

#### Scenario: upgrade an SDK-managed watch
- **GIVEN** a watch created by the last SDK-backed release of the provider
- **WHEN** Terraform refreshes and plans the resource with the Plugin Framework implementation
- **THEN** the resource SHALL keep the same `id` and `watch_id` without forcing replacement

## MODIFIED Requirements

### Requirement: Import (REQ-007–REQ-008)
The resource SHALL support import by storing the provided `id` value in state when the import identifier is in the format `<cluster_uuid>/<watch_id>`. For import and all subsequent read/delete operations, the resource SHALL require the `id` to be in the format `<cluster_uuid>/<watch_id>` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Import with valid composite id
- **GIVEN** an `id` in the format `<cluster_uuid>/<watch_id>`
- **WHEN** import completes
- **THEN** the `id` SHALL be stored in state for subsequent operations

#### Scenario: Invalid id format
- **GIVEN** a stored or imported `id` that does not contain exactly one `/`
- **WHEN** read or delete runs
- **THEN** the resource SHALL return an error diagnostic with "Wrong resource ID"

### Requirement: JSON field mapping — read/state (REQ-023–REQ-027)
On read, the resource SHALL marshal the API response fields `trigger`, `input`, `condition`, `actions`, and `metadata` back into normalized JSON strings and store them in state. When the API response includes a non-nil `transform`, the resource SHALL marshal it to a normalized JSON string and store it in state. When the API response has a nil `transform`, the resource SHALL clear `transform` from state so the Terraform state reflects the remote watch. The resource SHALL store `watch_id` and `active` (from `watch.status.state.active`) directly from the API response. The resource SHALL store `throttle_period_in_millis` from the API response. JSON fields SHALL normalize semantically equivalent JSON so formatting-only changes do not create perpetual diffs.

#### Scenario: transform removed from the remote watch
- **GIVEN** the watch previously had a `transform` stored in Terraform state
- **WHEN** read runs and the API response has no `transform` field
- **THEN** the `transform` attribute SHALL be cleared from state

#### Scenario: active synced from watch status
- **GIVEN** the watch is deactivated on the cluster
- **WHEN** read runs
- **THEN** `active` in state SHALL reflect `watch.status.state.active` from the API response
