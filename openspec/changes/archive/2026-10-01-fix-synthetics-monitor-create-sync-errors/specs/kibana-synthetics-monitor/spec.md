## MODIFIED Requirements

### Requirement: Kibana synthetics monitor APIs (REQ-001)

The resource SHALL manage synthetics monitors through Kibana's Synthetics Monitor HTTP API: add monitor (create), get monitor (read), update monitor (update), and delete monitor (delete). Reference: [Kibana Synthetics Add Monitor API](https://www.elastic.co/guide/en/kibana/current/add-monitor-api.html). After create and update, the resource SHALL derive final state from a get-monitor read and SHALL NOT build state from the add-monitor or update-monitor response body.

#### Scenario: Create monitor

- GIVEN valid configuration with exactly one type block and at least one location
- WHEN create runs
- THEN the provider SHALL call the Kibana add-monitor API with the configured space, SHALL take only the monitor `id` from the response, and SHALL set state from a subsequent get-monitor read in that space

#### Scenario: Create response is not a monitor

- GIVEN the add-monitor API returns HTTP 200 with a body containing `id` but no monitor fields (for example `{"message":"error pushing monitor to the service","attributes":{"errors":[...]},"id":"..."}`)
- WHEN create runs
- THEN the provider SHALL record the monitor in state from the get-monitor read and SHALL NOT surface an error diagnostic

#### Scenario: Create response without an id

- GIVEN the add-monitor API returns HTTP 200 with a body that has no `id`
- WHEN create runs
- THEN Terraform SHALL receive an error diagnostic stating the monitor ID could not be determined

#### Scenario: Read removes missing monitors

- GIVEN a read/refresh
- WHEN get returns HTTP 404 for the monitor
- THEN the provider SHALL remove the resource from state and SHALL NOT surface an error diagnostic

#### Scenario: Update monitor

- GIVEN a valid plan with changes to mutable fields
- WHEN update runs
- THEN the provider SHALL call the Kibana update-monitor API and SHALL set state from a subsequent get-monitor read

#### Scenario: Delete monitor

- GIVEN a monitor in state
- WHEN delete runs
- THEN the provider SHALL call the Kibana delete-monitor API with the monitor ID and space from state

## ADDED Requirements

### Requirement: Synthetics Service push error warnings (REQ-025)

When an add-monitor or update-monitor response with HTTP 200 contains `attributes.errors` (Synthetics Service push errors for Elastic-managed locations), the resource SHALL add one warning diagnostic that lists each reported location ID together with the reported reason and HTTP status when present. It SHALL NOT fail the operation because of these errors. Entries MAY omit any field, and a malformed `attributes` value SHALL produce no warning and no error. The create warning SHALL state that the monitor was created and is tracked in state, and that Kibana reports the outcome of its most recent sync, so the errors may relate to other monitors. The update warning SHALL state that the monitor was saved, but pushing it to the listed Elastic-managed locations failed. Both warnings SHALL state that Kibana retries syncing periodically.

#### Scenario: Create reports push errors

- GIVEN the add-monitor API returns the push-error body with `attributes.errors = [{"locationId":"us_west"}]`
- WHEN create completes
- THEN Terraform SHALL receive a warning naming `us_west`, stating the monitor was created and tracked in state, and noting the errors may relate to other monitors

#### Scenario: Update reports push errors

- GIVEN the update-monitor API returns HTTP 200 with `attributes.errors = [{"locationId":"us_west","error":{"reason":"boom","status":500}}]`
- WHEN update completes
- THEN Terraform SHALL receive a warning naming `us_west`, the reason `boom`, and status `500`, and state SHALL come from the get-monitor read

#### Scenario: No push errors

- GIVEN an add-monitor or update-monitor response without `attributes.errors`
- WHEN the operation completes
- THEN the provider SHALL NOT add a push-error warning
