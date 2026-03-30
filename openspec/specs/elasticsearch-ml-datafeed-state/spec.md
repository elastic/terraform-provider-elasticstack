# `elasticstack_elasticsearch_ml_datafeed_state` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/datafeed_state`

## Purpose

Manage the running state (`started` or `stopped`) of an existing Elasticsearch ML datafeed. This resource does not create or delete the underlying datafeed; it only drives state transitions (start/stop). On Terraform destroy, if the datafeed is in the `started` state the resource SHALL stop it before removing itself from state.

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_datafeed_state" "example" {
  id = <computed, string>  # internal identifier: <cluster_uuid>/<datafeed_id>

  datafeed_id = <required, string>  # force new; 1–64 chars; alphanumeric, hyphens, underscores
  state       = <required, string>  # one of: "started", "stopped"

  force            = <optional+computed, bool>    # default: false; forcefully stop the datafeed when stopping
  datafeed_timeout = <optional+computed, string>  # Go duration string; default: "30s"; used when starting or stopping

  start = <optional+computed, string>  # RFC 3339 datetime; when to start collecting data; unknown when state changes
  end   = <optional, string>           # RFC 3339 datetime; when to stop collecting data

  timeouts {  # optional
    create = <optional, string>  # default: 5 minutes
    update = <optional, string>  # default: 5 minutes
  }

  elasticsearch_connection {  # optional, deprecated
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    headers                  = <optional, map(string)>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    key_file                 = <optional, string>
    cert_data                = <optional, string>
    key_data                 = <optional, string>
  }
}
```

## Requirements

### Requirement: API — Start, Stop, and Stats (REQ-001–REQ-003)

To transition a datafeed to `started`, the resource SHALL call the Elasticsearch Start Datafeed API ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-start-datafeed.html)). To transition a datafeed to `stopped`, the resource SHALL call the Elasticsearch Stop Datafeed API ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-stop-datafeed.html)). To read the current datafeed state, the resource SHALL call the Elasticsearch Get Datafeed Stats API ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-datafeed-stats.html)). When any of these APIs returns a non-success response, the resource SHALL surface the error in Terraform diagnostics.

#### Scenario: Start API error surfaced

- GIVEN the Start Datafeed API returns a non-success response
- WHEN create or update runs
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: Stop API error surfaced

- GIVEN the Stop Datafeed API returns a non-success response
- WHEN create or update runs
- THEN Terraform diagnostics SHALL include the API error

### Requirement: Identity (REQ-004)

The resource SHALL expose a computed `id` attribute in the format `<cluster_uuid>/<datafeed_id>`. The resource SHALL derive `id` using `client.ID` after a successful state transition. The resource SHALL persist `id` in state using `UseStateForUnknown`.

#### Scenario: ID set after create

- GIVEN a successful start or stop operation during create
- WHEN create completes
- THEN the `id` in state SHALL be `<cluster_uuid>/<datafeed_id>`

### Requirement: Import (REQ-005)

The resource SHALL support import via `ImportStatePassthroughID` on the `datafeed_id` attribute. After import, the `datafeed_id` SHALL be persisted to state and used for subsequent read operations.

#### Scenario: Import passthrough on datafeed_id

- GIVEN import with the datafeed ID value
- WHEN import completes
- THEN the `datafeed_id` SHALL be stored in state for subsequent read, update, and delete operations

### Requirement: Lifecycle — datafeed_id requires replacement (REQ-006)

Changing `datafeed_id` SHALL require resource replacement. The resource SHALL apply `RequiresReplace` as a plan modifier on the `datafeed_id` attribute.

#### Scenario: datafeed_id change triggers replacement

- GIVEN an existing managed datafeed state
- WHEN the `datafeed_id` attribute is changed in configuration
- THEN Terraform SHALL plan a destroy-and-recreate

### Requirement: Connection (REQ-007)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client via `clients.MaybeNewAPIClientFromFrameworkResource` for all API calls (create, read, update, delete).

#### Scenario: Resource-level client override

- GIVEN `elasticsearch_connection` is set with specific endpoints and credentials
- WHEN any API call is made
- THEN the resource-scoped client SHALL be used instead of the provider client

### Requirement: State management — start/stop transitions (REQ-008–REQ-013)

On create and update, the resource SHALL first read the current datafeed state using the Get Datafeed Stats API. If the datafeed does not exist, the resource SHALL fail with an "ML Datafeed not found" error diagnostic and SHALL not attempt any state transition. If the current state already equals the desired `state`, the resource SHALL skip the API call and read the current state from the API to populate state (no transition needed). When `state` is `started` and the current state differs, the resource SHALL call the Start Datafeed API with the configured `start`, `end`, and `datafeed_timeout` values (omitting empty start/end). When `state` is `stopped` and the current state differs, the resource SHALL call the Stop Datafeed API with the configured `force` flag and `datafeed_timeout` value. After initiating a state transition, the resource SHALL wait for the datafeed to reach the desired state using `datafeed.WaitForDatafeedState` before writing state.

#### Scenario: No-op when already in desired state

- GIVEN an ML datafeed already in the `stopped` state and `state = "stopped"` in configuration
- WHEN create or update runs
- THEN the resource SHALL NOT call the Stop Datafeed API

#### Scenario: Start transition on create

- GIVEN an ML datafeed in `stopped` state and `state = "started"` in configuration
- WHEN create runs
- THEN the resource SHALL call the Start Datafeed API and wait for the datafeed to reach `started`

#### Scenario: Stop with force

- GIVEN an ML datafeed in `started` state and `state = "stopped"` with `force = true`
- WHEN update runs
- THEN the resource SHALL call the Stop Datafeed API with the force flag set

#### Scenario: Start with time bounds

- GIVEN `state = "started"`, `start = "2024-01-01T00:00:00Z"`, and `end = "2024-12-31T23:59:59Z"`
- WHEN create runs
- THEN the Start Datafeed API call SHALL include the `start` and `end` parameters

#### Scenario: Datafeed not found during create

- GIVEN a datafeed ID that does not exist in Elasticsearch
- WHEN create runs
- THEN the resource SHALL return an "ML Datafeed not found" error diagnostic

### Requirement: Missed transition handling (REQ-014)

When the wait for the desired state does not confirm the datafeed reached the target (e.g. the datafeed started and stopped too quickly), the resource SHALL re-read datafeed stats and verify that the timing `search_count` has increased compared to before the transition. If the search count has not increased, the resource SHALL return a "Datafeed did not successfully transition to the desired state" error diagnostic. If timing stats are absent, the resource SHALL emit a warning diagnostic instead.

#### Scenario: Datafeed starts and stops too quickly

- GIVEN a datafeed that starts and stops before the wait loop detects the `started` state
- WHEN update runs and the wait returns without confirming `started`
- THEN the resource SHALL re-read stats and check `search_count`; if not increased, SHALL return an error diagnostic

### Requirement: Delete — stop if started (REQ-015)

On Terraform destroy, the resource SHALL read the current datafeed state. If the datafeed is in the `started` state, the resource SHALL call the Stop Datafeed API with the configured `force` flag and `datafeed_timeout`, then wait for the datafeed to reach `stopped`. If the datafeed is not found during delete, the resource SHALL complete without error. If the datafeed is already `stopped`, the resource SHALL remove itself from state without calling the Stop API.

#### Scenario: Delete stops a started datafeed

- GIVEN an ML datafeed in `started` state
- WHEN the resource is destroyed
- THEN the resource SHALL call the Stop Datafeed API and wait for `stopped` before removing state

#### Scenario: Delete is a no-op for stopped datafeed

- GIVEN an ML datafeed in `stopped` state
- WHEN the resource is destroyed
- THEN the resource SHALL NOT call the Stop Datafeed API

#### Scenario: Delete is a no-op when datafeed not found

- GIVEN a datafeed that no longer exists in Elasticsearch
- WHEN the resource is destroyed
- THEN the resource SHALL complete without error

### Requirement: Read — not found handling (REQ-016)

On read, when the Get Datafeed Stats API returns no datafeed matching `datafeed_id`, the resource SHALL remove itself from state without returning an error.

#### Scenario: Datafeed removed outside Terraform

- GIVEN an ML datafeed that has been deleted outside of Terraform
- WHEN read runs
- THEN the resource SHALL be removed from state without error

### Requirement: Read — start and end from API (REQ-017)

On read, when the datafeed is in the `started` state and `running_state` contains a `search_interval`, the resource SHALL set `start` and `end` from the `search_interval.start_ms` and `search_interval.end_ms` values respectively, preserving the timezone of any previously-configured values. When `running_state.real_time_configured` is `true`, the resource SHALL set `end` to null in state. When `start` or `end` remain unknown after the API response, the resource SHALL set them to null in state.

#### Scenario: start/end populated from API on read

- GIVEN a started datafeed whose running_state reports a search_interval
- WHEN read runs
- THEN `start` and `end` SHALL be set from the API's search_interval values

#### Scenario: end set to null when real_time_configured

- GIVEN a started datafeed where `real_time_configured = true`
- WHEN read runs
- THEN `end` SHALL be null in state

### Requirement: Plan modifier — start becomes unknown when state changes (REQ-018)

The `start` attribute SHALL apply a custom `SetUnknownIfStateHasChanges` plan modifier. When `state` has changed between the prior state and the planned configuration (and `start` has not been explicitly set in configuration), the resource SHALL mark `start` as unknown in the plan, indicating it will be determined by the API.

#### Scenario: start becomes unknown on state change

- GIVEN a datafeed resource with `state = "stopped"` in current state and `state = "started"` in new configuration, with `start` not explicitly set
- WHEN plan is computed
- THEN `start` SHALL be unknown in the plan

### Requirement: Timeout handling (REQ-019–REQ-020)

The `datafeed_timeout` attribute SHALL accept a Go duration string and SHALL default to `"30s"`. When `datafeed_timeout` is greater than zero, the resource SHALL pass the parsed duration as the timeout to the Start Datafeed and Stop Datafeed API calls. The `timeouts` block SHALL accept Terraform framework timeout values for `create` and `update` operations; these default to 5 minutes each. When the create or update operation exceeds the framework timeout, the resource SHALL return an "Operation timed out" error diagnostic.

#### Scenario: Framework timeout error on create

- GIVEN a `timeouts.create` of `"1s"` and a datafeed that takes longer to start
- WHEN create runs and the context deadline is exceeded
- THEN the resource SHALL return an "Operation timed out" error diagnostic

### Requirement: datafeed_id validation (REQ-021)

The `datafeed_id` attribute SHALL be validated to be between 1 and 64 characters, and to contain only alphanumeric characters, hyphens, and underscores.

#### Scenario: Invalid datafeed_id rejected

- GIVEN a `datafeed_id` containing characters outside alphanumeric, hyphens, and underscores
- WHEN the configuration is applied
- THEN the provider SHALL return a validation error
