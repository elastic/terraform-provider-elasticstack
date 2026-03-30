# `elasticstack_elasticsearch_ml_job_state` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/jobstate`

## Purpose

Manage the running state (`opened` or `closed`) of an existing Elasticsearch anomaly detection ML job. This resource does not create or delete the underlying job; it only drives state transitions (open/close). On Terraform destroy the resource is removed from state without altering the actual job state in Elasticsearch.

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_job_state" "example" {
  id = <computed, string>  # internal identifier: <cluster_uuid>/<job_id>

  job_id = <required, string>  # force new; 1–64 chars; alphanumeric, hyphens, underscores
  state  = <required, string>  # one of: "opened", "closed"

  force       = <optional+computed, bool>    # default: false; forcefully close the job when closing
  job_timeout = <optional+computed, string>  # Go duration string; default: "30s"; used when closing

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

### Requirement: API — Open and Close ML Job (REQ-001–REQ-003)

To transition a job to `opened`, the resource SHALL call the Elasticsearch Open Anomaly Detection Job API ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-open-job.html)). To transition a job to `closed`, the resource SHALL call the Elasticsearch Close Anomaly Detection Job API ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-close-job.html)). To read current job state, the resource SHALL call the Elasticsearch Get Anomaly Detection Job Stats API ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-job-stats.html)). When any of these APIs returns a non-success response, the resource SHALL surface the error in Terraform diagnostics.

#### Scenario: Open API error surfaced

- GIVEN the Open Job API returns a non-success response
- WHEN create or update runs
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: Close API error surfaced

- GIVEN the Close Job API returns a non-success response
- WHEN create or update runs
- THEN Terraform diagnostics SHALL include the API error

### Requirement: Identity (REQ-004)

The resource SHALL expose a computed `id` attribute in the format `<cluster_uuid>/<job_id>`. The resource SHALL derive `id` using `client.ID` after a successful state transition. The resource SHALL persist `id` in state using `UseStateForUnknown`.

#### Scenario: ID set after create

- GIVEN a successful open or close operation during create
- WHEN create completes
- THEN the `id` in state SHALL be `<cluster_uuid>/<job_id>`

### Requirement: Import (REQ-005)

The resource SHALL support import via `ImportStatePassthroughID` on the `id` attribute. After import, the `id` SHALL be persisted to state and used for subsequent read operations. On read after import, the resource SHALL parse `id` with `clients.CompositeIDFromStrFw` and return an error diagnostic when the format is invalid.

#### Scenario: Import passthrough

- GIVEN import with a valid `<cluster_uuid>/<job_id>` value
- WHEN import completes
- THEN the `id` SHALL be stored in state for subsequent read, update, and delete operations

#### Scenario: Invalid id format on read

- GIVEN a state `id` not matching `<cluster_uuid>/<job_id>`
- WHEN read runs
- THEN the resource SHALL return an error diagnostic

### Requirement: Lifecycle — job_id requires replacement (REQ-006)

Changing `job_id` SHALL require resource replacement. The resource SHALL apply `RequiresReplace` as a plan modifier on the `job_id` attribute.

#### Scenario: job_id change triggers replacement

- GIVEN an existing managed job state
- WHEN the `job_id` attribute is changed in configuration
- THEN Terraform SHALL plan a destroy-and-recreate

### Requirement: Connection (REQ-007)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client via `clients.MaybeNewAPIClientFromFrameworkResource` for all API calls (create, read, update, delete).

#### Scenario: Resource-level client override

- GIVEN `elasticsearch_connection` is set with specific endpoints and credentials
- WHEN any API call is made
- THEN the resource-scoped client SHALL be used instead of the provider client

### Requirement: State management — open/close transitions (REQ-008–REQ-011)

On create and update, the resource SHALL first read the current job state using the Get Job Stats API. If the job does not exist, the resource SHALL fail with an "ML Job not found" error diagnostic and SHALL not attempt any state transition. If the current state already equals the desired `state`, the resource SHALL skip the API call and make no changes. When `state` is `opened` and the current state differs, the resource SHALL call the Open Job API. When `state` is `closed` and the current state differs, the resource SHALL call the Close Job API with the configured `force` flag and the parsed `job_timeout` value. After initiating a state transition, the resource SHALL poll until the job reaches the desired state (using `asyncutils.WaitForStateTransition`) before writing state.

#### Scenario: No-op when already in desired state

- GIVEN an ML job already in the `opened` state and `state = "opened"` in configuration
- WHEN create or update runs
- THEN the resource SHALL NOT call the Open Job API

#### Scenario: Open transition on create

- GIVEN an ML job in `closed` state and `state = "opened"` in configuration
- WHEN create runs
- THEN the resource SHALL call the Open Job API and wait for the job to reach `opened`

#### Scenario: Close with force

- GIVEN an ML job in `opened` state and `state = "closed"` with `force = true`
- WHEN update runs
- THEN the resource SHALL call the Close Job API with the force flag set

#### Scenario: Job not found during create

- GIVEN an ML job ID that does not exist in Elasticsearch
- WHEN create runs
- THEN the resource SHALL return an "ML Job not found" error diagnostic

### Requirement: Delete — state only (REQ-012)

On Terraform destroy, the resource SHALL remove itself from Terraform state without calling any Elasticsearch API. The actual ML job SHALL remain in its current state (opened or closed) in Elasticsearch.

#### Scenario: Delete does not change job state

- GIVEN an ML job managed in opened state
- WHEN the resource is destroyed
- THEN Terraform state SHALL be cleared and the ML job SHALL remain unchanged in Elasticsearch

### Requirement: Read — not found handling (REQ-013)

On read, when the Get Job Stats API returns no job matching `job_id`, the resource SHALL remove itself from state without returning an error.

#### Scenario: Job removed outside Terraform

- GIVEN an ML job that has been deleted outside of Terraform
- WHEN read runs
- THEN the resource SHALL be removed from state without error

### Requirement: Read — computed defaults during import (REQ-014)

On read after import, when `force` is null the resource SHALL set it to `false`, and when `job_timeout` is null the resource SHALL set it to `"30s"`.

#### Scenario: Defaults set after import

- GIVEN an imported resource with null `force` and `job_timeout`
- WHEN read runs
- THEN `force` SHALL be `false` and `job_timeout` SHALL be `"30s"` in state

### Requirement: Timeout handling (REQ-015–REQ-016)

The `job_timeout` attribute SHALL accept a Go duration string and SHALL default to `"30s"`. When `job_timeout` is greater than zero, the resource SHALL pass the parsed duration as the timeout to the Close Job API. The `timeouts` block SHALL accept Terraform framework timeout values for `create` and `update` operations; these default to 5 minutes each. When the create or update operation exceeds the framework timeout, the resource SHALL return an "Operation timed out" error diagnostic.

#### Scenario: Framework timeout error on create

- GIVEN a `timeouts.create` of `"1s"` and a job that takes longer to open
- WHEN create runs and the context deadline is exceeded
- THEN the resource SHALL return an "Operation timed out" error diagnostic

### Requirement: job_id validation (REQ-017)

The `job_id` attribute SHALL be validated to be between 1 and 64 characters, and to contain only alphanumeric characters, hyphens, and underscores.

#### Scenario: Invalid job_id rejected

- GIVEN a `job_id` containing characters outside alphanumeric, hyphens, and underscores
- WHEN the configuration is applied
- THEN the provider SHALL return a validation error
