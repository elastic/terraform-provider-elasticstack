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

  start = <optional, string>                    # RFC 3339 datetime; user input preserved verbatim in state
  end   = <optional, string>                    # RFC 3339 datetime; user input preserved verbatim in state
  effective_search_start = <computed, string>   # RFC 3339; ES running_state.search_interval.start_ms when started
  effective_search_end   = <computed, string>   # RFC 3339; ES running_state.search_interval.end_ms when started; null when real-time or stopped

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

The resource SHALL resolve a `*clients.ElasticsearchScopedClient` from the provider client factory and call `GetESClient()` to perform Elasticsearch operations. When `elasticsearch_connection` is absent, the factory SHALL return a typed client built from provider-level defaults. When `elasticsearch_connection` is configured, the factory SHALL return a typed scoped client rebuilt from that connection for all API calls (create, read, update, delete).

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

### Requirement: Read — preserve configured start and end (REQ-017)

On read, the resource SHALL NOT overwrite the `start` or `end` attribute with values returned by the Get Datafeed Stats API. The `start` and `end` attributes SHALL round-trip from configuration (or from prior state when no configuration value is supplied) so that practitioner-declared values are preserved verbatim — even when Elasticsearch reports a different effective search interval (e.g. after bucket alignment or first-document snap-forward). Read SHALL still call the Get Datafeed Stats API to populate `state` and the computed `effective_search_start` / `effective_search_end` attributes (see REQ-022).

#### Scenario: Explicit start is preserved across apply

- GIVEN a `started` datafeed configured with `start = "2022-01-01T00:07:30Z"`
- AND Elasticsearch reports `running_state.search_interval.start_ms = "2022-01-01T00:10:00Z"` after the datafeed begins searching
- WHEN create or update runs
- THEN the `start` attribute in state SHALL equal `"2022-01-01T00:07:30Z"` (the configured value)
- AND the apply SHALL NOT produce a "Provider produced inconsistent result after apply" diagnostic

#### Scenario: Explicit start is preserved across bucket alignment

- GIVEN a `started` datafeed for a job with `bucket_span = "15m"` configured with `start = "2025-07-13T02:23:23.935Z"`
- AND Elasticsearch reports `running_state.search_interval.start_ms = "2025-07-13T02:26:42.000Z"` after bucket alignment
- WHEN create runs
- THEN the `start` attribute in state SHALL equal `"2025-07-13T02:23:23.935Z"`
- AND the apply SHALL succeed

#### Scenario: Explicit end is preserved across apply

- GIVEN a `started` datafeed configured with `end = "2024-12-31T23:59:59Z"`
- AND Elasticsearch reports a different `running_state.search_interval.end_ms`
- WHEN create or update runs
- THEN the `end` attribute in state SHALL equal `"2024-12-31T23:59:59Z"` (the configured value)

#### Scenario: Omitted start remains null on read

- GIVEN a `started` datafeed where `start` is not set in configuration
- WHEN read runs
- THEN the `start` attribute in state SHALL be null
- AND the `effective_search_start` attribute SHALL be populated from `running_state.search_interval.start_ms`

### Requirement: Computed effective search interval attributes (REQ-022)

The resource SHALL expose two computed read-only attributes that report Elasticsearch's view of the active search interval:

- `effective_search_start` (RFC 3339 datetime): the value of `running_state.search_interval.start_ms` from the Get Datafeed Stats API.
- `effective_search_end` (RFC 3339 datetime): the value of `running_state.search_interval.end_ms` from the Get Datafeed Stats API.

Both attributes SHALL be `Computed` only (never settable by configuration). On read, when the datafeed is in the `started` state and `running_state.search_interval` is present, the resource SHALL populate the attributes from the corresponding `start_ms` and `end_ms` values, preserving the timezone of any previously-configured `start` / `end` values for display. When `running_state.real_time_configured` is `true`, the resource SHALL set `effective_search_end` to null. When the datafeed is in the `stopped` state, or `running_state` / `search_interval` is absent (including the "datafeed started and stopped too quickly" path covered by REQ-014), the resource SHALL set both attributes to null.

#### Scenario: Effective search interval populated for a started datafeed

- GIVEN a `started` datafeed whose `running_state.search_interval` reports `start_ms = "2022-01-01T00:10:00Z"` and `end_ms = "2022-01-01T01:00:00Z"`
- WHEN read runs
- THEN `effective_search_start` SHALL equal `"2022-01-01T00:10:00Z"`
- AND `effective_search_end` SHALL equal `"2022-01-01T01:00:00Z"`

#### Scenario: Effective end is null when real-time

- GIVEN a `started` datafeed where `running_state.real_time_configured = true`
- WHEN read runs
- THEN `effective_search_end` SHALL be null in state

#### Scenario: Effective attributes are null for a stopped datafeed

- GIVEN a datafeed in the `stopped` state
- WHEN read runs
- THEN both `effective_search_start` and `effective_search_end` SHALL be null in state

#### Scenario: Effective attributes are null when running_state is missing

- GIVEN a `started` datafeed whose `running_state` is absent or whose `search_interval` is absent
- WHEN read runs
- THEN both `effective_search_start` and `effective_search_end` SHALL be null in state

#### Scenario: Effective attributes round-trip without drift

- GIVEN a `started` datafeed with `effective_search_start` and `effective_search_end` already in state
- WHEN a subsequent `terraform plan` runs against an unchanged configuration
- THEN no plan diff SHALL be produced for either attribute

### Requirement: Timeout handling (REQ-019–REQ-020)

The `datafeed_timeout` attribute SHALL accept a Go duration string and SHALL default to `"30s"`. When `datafeed_timeout` is greater than zero, the resource SHALL pass the parsed duration as the timeout to the Start Datafeed and Stop Datafeed API calls. The `timeouts` block SHALL accept Terraform framework timeout values for `create` and `update` operations; these default to 5 minutes each. When the create or update operation exceeds the framework timeout, the resource SHALL return an "Operation timed out" error diagnostic.

#### Scenario: Framework timeout error on create

- GIVEN a `timeouts.create` of `"1s"` and a datafeed that takes longer to start
- WHEN create runs and the context deadline is exceeded
- THEN the resource SHALL return an "Operation timed out" error diagnostic

### Requirement: datafeed_id validation (REQ-021)

The `datafeed_id` attribute SHALL be validated at plan time to contain at least one character, to
contain only lowercase alphanumeric characters (a–z and 0–9), hyphens, underscores, and dots, and
to start and end with an alphanumeric character. No upper-bound length restriction is applied. The
validator SHALL use `ml.IDValidatorWithoutLength()` (defined in

#### Scenario: Long datafeed_id accepted

- GIVEN a `datafeed_id` that is longer than 64 characters and otherwise valid
- WHEN the configuration is validated
- THEN validation SHALL succeed with no error diagnostics

#### Scenario: Invalid datafeed_id rejected — illegal characters

- GIVEN a `datafeed_id` containing characters outside alphanumeric, hyphens, and underscores
- WHEN the configuration is validated
- THEN validation SHALL fail with an error diagnostic

#### Scenario: Empty datafeed_id rejected

- GIVEN a `datafeed_id` that is an empty string
- WHEN the configuration is validated
- THEN validation SHALL fail with an error diagnostic

