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

### Requirement: JSON field mapping — create/update (REQ-018–REQ-022)
On create and update, the resource SHALL unmarshal each JSON string attribute (`trigger`, `input`, `condition`, `actions`, `metadata`) into a `map[string]any` before constructing the API request body; if any unmarshal fails, the resource SHALL return a diagnostic error and SHALL NOT call the Put Watch API. When `transform` is configured, the resource SHALL include its JSON object in the Put Watch request body. When `transform` is not configured on create, the `transform` field SHALL be omitted from the Put Watch JSON body. When `transform` is not configured on update, the Put Watch JSON body SHALL include `transform` with an empty JSON object `{}` so Elasticsearch clears any existing transform. The `throttle_period_in_millis` value SHALL be included in the request body when non-zero. The `active` flag SHALL be passed as a query parameter to the Put Watch API.

#### Scenario: transform cleared when omitted on update
- **GIVEN** an existing watch previously stored a `transform`
- **WHEN** update builds the request body with no configured `transform`
- **THEN** the Put Watch JSON body SHALL include `transform` with an empty JSON object

### Requirement: JSON field mapping — read/state (REQ-023–REQ-027)
On read, the resource SHALL marshal the API response fields `trigger`, `input`, `condition`, `actions`, and `metadata` back into JSON strings and store them in state. When the API response includes a non-empty `transform` object (at least one top-level key), the resource SHALL marshal it to a JSON string and store it in state. When the API response omits `transform`, has a null `transform`, or has an empty JSON object `{}` for `transform`, the resource SHALL clear `transform` from state so the Terraform state reflects the remote watch. The resource SHALL store `watch_id` and `active` (from `watch.status.state.active`) directly from the API response. The resource SHALL store `throttle_period_in_millis` from the API response. JSON fields SHALL use `DiffSuppressFunc` (`tfsdkutils.DiffJSONSuppress`) to suppress semantically equivalent JSON diffs.

#### Scenario: transform removed from the remote watch
- **GIVEN** the watch previously had a `transform` stored in Terraform state
- **WHEN** read runs and the API response has no `transform` field, a null `transform`, or an empty `transform` object
- **THEN** the `transform` attribute SHALL be cleared from state

#### Scenario: active synced from watch status
- **GIVEN** the watch is deactivated on the cluster
- **WHEN** read runs
- **THEN** `active` in state SHALL reflect `watch.status.state.active` from the API response
