## MODIFIED Requirements

### Requirement: Read (REQ-014–REQ-016)

On read, the resource SHALL parse `id` using the framework resource's composite ID format to extract the watch identifier. The resource SHALL call the Get Watch API with the extracted watch identifier. When the Get Watch API returns 404, the resource SHALL remove itself from state (set id to `""`). When the API returns a successful response, the resource SHALL decode the JSON response and update state from the response, except that `actions` SHALL preserve previously known values at nested paths where the API returns a redacted string sentinel and prior Terraform state has a corresponding concrete value.

#### Scenario: Watch not found on refresh

- GIVEN the watch no longer exists on the cluster
- WHEN read runs
- THEN the resource SHALL be removed from state without an error

#### Scenario: Redacted action secret during refresh

- GIVEN the prior Terraform state contains a concrete `actions` value for a nested action secret
- WHEN read runs and the Get Watch API returns the same action path with a redacted string sentinel
- THEN the resource SHALL preserve the prior concrete value at that path while updating the rest of state from the API response

### Requirement: JSON field mapping — read/state (REQ-023–REQ-027)

On read, the resource SHALL marshal the API response fields `trigger`, `input`, `condition`, and `metadata` back into normalized JSON strings and store them in state. For `actions`, the resource SHALL marshal the API response into a normalized JSON string, but when the API response contains a redacted string sentinel at a nested path and prior Terraform state has a concrete value at the same path, the resource SHALL preserve the prior concrete value at that path in the final stored JSON. When no prior concrete value exists for a redacted `actions` path, the resource SHALL store the API value as returned. When the API response includes a non-nil `transform`, the resource SHALL marshal it to a normalized JSON string and store it in state. When the API response has a nil `transform`, the resource SHALL clear `transform` from state so the Terraform state reflects the remote watch. The resource SHALL store `watch_id` and `active` (from `watch.status.state.active`) directly from the API response. The resource SHALL store `throttle_period_in_millis` from the API response. JSON fields SHALL normalize semantically equivalent JSON so formatting-only changes do not create perpetual diffs.

#### Scenario: transform removed from the remote watch

- **GIVEN** the watch previously had a `transform` stored in Terraform state
- **WHEN** read runs and the API response has no `transform` field
- **THEN** the `transform` attribute SHALL be cleared from state

#### Scenario: active synced from watch status

- **GIVEN** the watch is deactivated on the cluster
- **WHEN** read runs
- **THEN** `active` in state SHALL reflect `watch.status.state.active` from the API response

#### Scenario: redacted action secret preserved from prior state

- **GIVEN** the watch previously stored an `actions` JSON document with a concrete nested secret value
- **WHEN** read runs and the API response returns `::es_redacted::` for that nested action path
- **THEN** the final `actions` value in state SHALL keep the prior concrete secret for that path
- **AND** non-redacted action fields from the API response SHALL still be reflected in state

#### Scenario: imported watch with redacted action secret

- **GIVEN** Terraform has no prior concrete value for a redacted nested `actions` path
- **WHEN** read runs and the API response returns `::es_redacted::` for that path
- **THEN** the `actions` value in state SHALL store the redacted API value for that path
