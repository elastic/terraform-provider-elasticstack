## ADDED Requirements

### Requirement: Defaulted watch attributes (REQ-028)
When `active` is omitted from configuration, the resource SHALL behave as if `active` were `true`. When `throttle_period_in_millis` is omitted, the resource SHALL submit the default throttle period and SHALL store the refreshed value in state. When `input`, `condition`, `actions`, or `metadata` are omitted, the resource SHALL use their documented JSON defaults during create and update.

#### Scenario: active omitted from configuration
- **GIVEN** a watch configuration that omits `active`
- **WHEN** the resource is created and refreshed
- **THEN** the `active` attribute in state SHALL be `true`

#### Scenario: throttle period omitted from configuration
- **GIVEN** a watch configuration that omits `throttle_period_in_millis`
- **WHEN** the resource is created and refreshed
- **THEN** the `throttle_period_in_millis` attribute in state SHALL be `5000`

## MODIFIED Requirements

### Requirement: JSON field mapping — read/state (REQ-023–REQ-027)
On read, the resource SHALL marshal the API response fields `trigger`, `input`, `condition`, `actions`, and `metadata` back into JSON strings and store them in state. When the API response includes a non-nil `transform`, the resource SHALL marshal it to a JSON string and store it in state. When the API response has a nil `transform`, the resource SHALL clear `transform` from state so the Terraform state reflects the remote watch. The resource SHALL store `watch_id` and `active` (from `watch.status.state.active`) directly from the API response. The resource SHALL store `throttle_period_in_millis` from the API response. JSON fields SHALL use `DiffSuppressFunc` (`tfsdkutils.DiffJSONSuppress`) to suppress semantically equivalent JSON diffs.

#### Scenario: transform removed from the remote watch
- **GIVEN** the watch previously had a `transform` stored in Terraform state
- **WHEN** read runs and the API response has no `transform` field
- **THEN** the `transform` attribute SHALL be cleared from state

#### Scenario: active synced from watch status
- **GIVEN** the watch is deactivated on the cluster
- **WHEN** read runs
- **THEN** `active` in state SHALL reflect `watch.status.state.active` from the API response
