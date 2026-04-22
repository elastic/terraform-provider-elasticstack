## MODIFIED Requirements

### Requirement: Read (REQ-014–REQ-016)

On read, the resource SHALL parse `id` using the framework resource's composite ID format to extract the watch identifier. The resource SHALL call the Get Watch API with the extracted watch identifier. When the Get Watch API returns 404, the resource SHALL remove itself from state. When the API returns a successful response, the resource SHALL decode the JSON response and update state from the response, except that for `actions` the resource SHALL preserve prior known Terraform values of any JSON type at nested paths where the API returns the redacted string sentinel. The prior `actions` JSON SHALL be the last-applied value from Terraform state when read is a refresh, and the configured value from the Terraform plan when read runs as read-after-write after create or update.

#### Scenario: Watch not found on refresh

- **GIVEN** the watch no longer exists on the cluster
- **WHEN** read runs
- **THEN** the resource SHALL be removed from state without an error

#### Scenario: Redacted action secret during refresh

- **GIVEN** the last-applied `actions` value in Terraform state includes a concrete nested action secret at a path
- **WHEN** read runs and the Get Watch API returns the same path with a redacted string sentinel
- **THEN** the resource SHALL preserve the prior state concrete value at that path while updating the rest of state from the API response

#### Scenario: Redacted action secret after create or update read-after-write

- **GIVEN** the Terraform plan for the apply includes a concrete `actions` value for a nested action secret at a path
- **WHEN** read runs as read-after-write after a successful Put Watch and the Get Watch API returns that path with a redacted string sentinel
- **THEN** the resource SHALL preserve the prior plan concrete value at that path while updating the rest of state from the API response

#### Scenario: Redacted action header preserved when prior is a non-string

- **GIVEN** the prior `actions` JSON has a non-string value at a nested path (for example `headers.Authorization = {"id": "service-now-key"}` or an inline-script object `{"source": "...", "lang": "painless"}`)
- **WHEN** read runs and the Get Watch API returns the redacted string sentinel at that same path
- **THEN** the resource SHALL preserve the prior non-string value at that path while updating the rest of state from the API response

### Requirement: JSON field mapping — read/state (REQ-023–REQ-027)

On read, the resource SHALL marshal the API response fields `trigger`, `input`, `condition`, and `metadata` back into normalized JSON strings and store them in state. For `actions`, the resource SHALL marshal the API response into a normalized JSON string, but when the API response contains the redacted string sentinel at a nested path and the prior known Terraform `actions` JSON value (state on refresh, or plan on read-after-write after create or update) has a concrete value of any JSON type at the same path, the resource SHALL preserve that prior concrete value at that path in the final stored JSON. When no prior concrete value exists for a redacted `actions` path, or when the prior value at that path is itself the redacted string sentinel, the resource SHALL store the API value as returned. When the API response includes a non-nil `transform`, the resource SHALL marshal it to a normalized JSON string and store it in state. When the API response has a nil `transform`, the resource SHALL clear `transform` from state so the Terraform state reflects the remote watch. The resource SHALL store `watch_id` and `active` (from `watch.status.state.active`) directly from the API response. The resource SHALL store `throttle_period_in_millis` from the API response. JSON fields SHALL normalize semantically equivalent JSON so formatting-only changes do not create perpetual diffs.

#### Scenario: transform removed from the remote watch

- **GIVEN** the watch previously had a `transform` stored in Terraform state
- **WHEN** read runs and the API response has no `transform` field
- **THEN** the `transform` attribute SHALL be cleared from state

#### Scenario: active synced from watch status

- **GIVEN** the watch is deactivated on the cluster
- **WHEN** read runs
- **THEN** `active` in state SHALL reflect `watch.status.state.active` from the API response

#### Scenario: redacted action secret preserved from prior state

- **GIVEN** the last-applied `actions` JSON in Terraform state includes a concrete nested secret value at a path
- **WHEN** read runs and the API response returns `::es_redacted::` for that nested action path
- **THEN** the final `actions` value in state SHALL keep the prior concrete secret for that path
- **AND** non-redacted action fields from the API response SHALL still be reflected in state

#### Scenario: imported watch with redacted action secret

- **GIVEN** Terraform has no prior concrete `actions` value in state or plan for a redacted nested path
- **WHEN** read runs and the API response returns `::es_redacted::` for that path
- **THEN** the `actions` value in state SHALL store the redacted API value for that path

#### Scenario: redacted action header preserved when prior is an object

- **GIVEN** the prior `actions` JSON has an object at a nested path (for example a stored-script reference `headers.Authorization = {"id": "service-now-key"}`)
- **WHEN** read runs and the API response returns `::es_redacted::` for that nested action path
- **THEN** the final `actions` value in state SHALL keep the prior object at that path
- **AND** non-redacted action fields from the API response SHALL still be reflected in state

#### Scenario: redacted action header preserved when prior is an inline script

- **GIVEN** the prior `actions` JSON has an inline-script object at a nested path (for example `headers.Authorization = {"source": "return 'Bearer x'", "lang": "painless"}`)
- **WHEN** read runs and the API response returns `::es_redacted::` for that nested action path
- **THEN** the final `actions` value in state SHALL keep the prior inline-script object at that path

#### Scenario: redacted action leaf preserved when prior is an array

- **GIVEN** the prior `actions` JSON has an array at a nested path
- **WHEN** read runs and the API response returns `::es_redacted::` for that nested action path
- **THEN** the final `actions` value in state SHALL keep the prior array at that path
