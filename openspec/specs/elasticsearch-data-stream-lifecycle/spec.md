# `elasticstack_elasticsearch_data_stream_lifecycle` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/datastreamlifecycle`

## Purpose

Configure the data stream lifecycle (DLM) policy for one or more Elasticsearch data streams. The resource calls the Elasticsearch Put/Get/Delete Data Lifecycle APIs to create, refresh, and remove lifecycle settings (data retention, enabled flag, and downsampling rounds) on the targeted data streams. Wildcard patterns in `name` allow a single resource to manage the lifecycle of multiple data streams simultaneously.

## Schema

```hcl
resource "elasticstack_elasticsearch_data_stream_lifecycle" "example" {
  id   = <computed, string>  # internal identifier: <cluster_uuid>/<name>
  name = <required, string>  # data stream name or wildcard pattern; UseStateForUnknown

  data_retention   = <optional, string>  # ISO-8601 duration; omit for indefinite retention
  enabled          = <optional+computed, bool>    # default true
  expand_wildcards = <optional+computed, string>  # default "open"; one of: all, open, closed, hidden, none

  downsampling {                   # optional, list(object), max 10 items
    after          = <required, string>  # interval before backing index is downsampled
    fixed_interval = <required, string>  # aggregation interval for downsampling
  }

  elasticsearch_connection {       # optional, deprecated block
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

### Requirement: Data stream lifecycle CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put Data Lifecycle API (`IndicesPutDataLifecycle`) to create and update data stream lifecycle settings ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/data-stream-apis.html)). The resource SHALL use the Elasticsearch Get Data Lifecycle API (`IndicesGetDataLifecycle`) to read lifecycle settings. The resource SHALL use the Elasticsearch Delete Data Lifecycle API (`IndicesDeleteDataLifecycle`) to remove lifecycle settings. When Elasticsearch returns a non-success response (other than 404 on read), the resource SHALL surface the API error as a Terraform diagnostic and SHALL not update state.

#### Scenario: API failure on create or update

- GIVEN the Put Data Lifecycle API returns a non-success status
- WHEN the resource handles the response
- THEN Terraform diagnostics SHALL include the error detail

#### Scenario: API failure on delete

- GIVEN the Delete Data Lifecycle API returns a non-success status
- WHEN the resource handles the response
- THEN Terraform diagnostics SHALL include the error detail

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<name>`. During create (and update, which reuses the create path), the resource SHALL compute `id` by calling `client.ID(ctx, name)`, which combines the current cluster UUID with the configured `name` value. The `id` attribute SHALL use the `UseStateForUnknown` plan modifier so that it is preserved across plan/apply cycles once set.

#### Scenario: ID computed on create

- GIVEN a new resource with `name = "my-data-stream"`
- WHEN create completes successfully
- THEN `id` SHALL be set to `<cluster_uuid>/my-data-stream`

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import via `ImportStatePassthroughID`, persisting the supplied `id` value directly to state. Read and delete operations SHALL parse the imported `id` using `CompositeIDFromStr`, which requires the format `<cluster_uuid>/<resource_identifier>`; if the format is invalid (not exactly two `/`-separated parts), the resource SHALL return an error diagnostic with summary "Wrong resource ID." and detail describing the expected format.

#### Scenario: Import with valid composite id

- GIVEN import with `id = "<cluster_uuid>/my-data-stream"`
- WHEN import completes
- THEN the id SHALL be stored in state and subsequent reads SHALL use the data stream name parsed from it

#### Scenario: Import with malformed id

- GIVEN a stored `id` that does not match `<cluster_uuid>/<resource_identifier>`
- WHEN read or delete runs
- THEN the resource SHALL return a "Wrong resource ID." error diagnostic

### Requirement: Lifecycle — no forced replacement (REQ-009)

No attribute in this resource uses `RequiresReplace`. Changing any attribute (including `name`, `data_retention`, `enabled`, `expand_wildcards`, or `downsampling`) SHALL be applied in place by re-running the Put Data Lifecycle API rather than requiring resource replacement. The `name` attribute uses `UseStateForUnknown` only to preserve its value when it is unknown during planning.

#### Scenario: Name change applied in place

- GIVEN an existing resource with `name = "logs-*"`
- WHEN `name` is changed to a different pattern
- THEN the resource SHALL update lifecycle settings in place without destroying and recreating the resource

### Requirement: Connection (REQ-010–REQ-011)

By default, the resource SHALL use the provider-level Elasticsearch client configured via `Configure`. When the `elasticsearch_connection` block is set, the resource SHALL construct and use a resource-scoped Elasticsearch client (via `clients.MaybeNewAPIClientFromFrameworkResource`) for all API calls (create, read, update, delete).

#### Scenario: Resource-level connection override

- GIVEN `elasticsearch_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the resource-scoped client and not the provider client

### Requirement: Compatibility — minimum server version (REQ-012)

The resource SHALL declare `MinVersion = "8.11.0"` to indicate that Data Lifecycle Management requires Elasticsearch 8.11.0 or later. Acceptance tests SHALL use this constant to skip against older clusters. The resource does not enforce a runtime version check in CRUD operations; it SHALL rely on the Elasticsearch server to reject requests on unsupported versions and surface them as API errors.

#### Scenario: Server version below minimum in acceptance tests

- GIVEN an Elasticsearch cluster running a version older than 8.11.0
- WHEN acceptance tests run
- THEN the test suite SHALL skip the resource tests using the `MinVersion` constant

### Requirement: Create and update (REQ-013–REQ-015)

On create and update, the resource SHALL read the plan model, resolve the Elasticsearch client, compute the composite `id`, convert the plan to a `models.LifecycleSettings` struct, and call `PutDataStreamLifecycle`. The `expand_wildcards` value SHALL be forwarded as the `WithExpandWildcards` option on the Put request. After a successful Put, the resource SHALL perform a read and store the result in state. If any step (client resolution, id computation, API call, or read-back) returns an error, the resource SHALL return the error diagnostic and SHALL not finalize the state.

#### Scenario: Successful create

- GIVEN a valid plan with `name`, `data_retention`, and `enabled`
- WHEN create runs and Put succeeds
- THEN the resource SHALL call read and store the refreshed model in state

### Requirement: Read (REQ-016–REQ-018)

On read, the resource SHALL parse the composite `id` from state using `CompositeIDFromStr` to extract the data stream name, then call `GetDataStreamLifecycle` with that name and the `expand_wildcards` value. When the Get API returns HTTP 404 (or an empty `data_streams` list), the resource SHALL remove itself from state (`resp.State.RemoveResource`). When the Get API returns a non-empty list, the resource SHALL update `data_retention` and `downsampling` in state from the first matching entry where the API value differs from the current state value; `enabled` and `expand_wildcards` are preserved from state and not overwritten from the API response.

#### Scenario: Data stream not found on read

- GIVEN the Get Data Lifecycle API returns 404 or an empty list
- WHEN read runs
- THEN the resource SHALL be removed from state

#### Scenario: Partial state refresh on read

- GIVEN the API returns a `data_retention` different from the state value
- WHEN read runs
- THEN `data_retention` in state SHALL be updated to the API value

### Requirement: Delete (REQ-019–REQ-020)

On delete, the resource SHALL parse the composite `id` from state to extract the data stream name, then call `DeleteDataStreamLifecycle` with that name and the `expand_wildcards` value. On success, the resource SHALL call `resp.State.RemoveResource`. On API error, the resource SHALL return the error diagnostic and SHALL not remove the resource from state.

#### Scenario: Successful delete

- GIVEN a resource with a valid `id`
- WHEN delete runs and the API succeeds
- THEN the resource SHALL be removed from state

### Requirement: Mapping — config to API (REQ-021–REQ-023)

When converting plan to API model, the resource SHALL set `data_retention` from the string value of `DataRetention` (empty string when null/unset, which the API treats as indefinite retention). The resource SHALL set `enabled` from the bool value of `Enabled` (defaults to `true`). When `downsampling` is non-null, non-unknown, and non-empty, the resource SHALL convert each list element to a `models.Downsampling` struct with `after` and `fixed_interval` fields; when `downsampling` is null, unknown, or empty, the resource SHALL omit the `Downsampling` field from the API request body entirely (JSON `omitempty`).

#### Scenario: Empty downsampling omitted

- GIVEN no `downsampling` blocks in config
- WHEN the plan is converted to API model
- THEN the `downsampling` field SHALL be absent from the JSON request body

### Requirement: Mapping — API to state (REQ-024–REQ-026)

When populating state from the API response, the resource SHALL iterate over the returned `data_streams` entries and update `data_retention` in state only when the API value differs from the current state value. The resource SHALL update `downsampling` in state only when the API list differs from state (by length or element values); if they match, the state value SHALL be preserved without modification. `enabled` and `expand_wildcards` SHALL not be read from the API response and SHALL remain as stored in state.

#### Scenario: Downsampling unchanged

- GIVEN API returns downsampling identical to state
- WHEN read maps the response to state
- THEN `downsampling` in state SHALL remain unchanged

### Requirement: Schema — expand_wildcards validation (REQ-027)

The `expand_wildcards` attribute SHALL be validated as one of the following string values: `all`, `open`, `closed`, `hidden`, `none`. Any other value SHALL be rejected at plan time with a validation error.

#### Scenario: Invalid expand_wildcards value

- GIVEN `expand_wildcards = "invalid"`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error listing the allowed values

### Requirement: Schema — downsampling size limit (REQ-028)

The `downsampling` list SHALL accept at most 10 entries. Configurations with more than 10 downsampling rounds SHALL be rejected at plan time with a validation error.

#### Scenario: Too many downsampling entries

- GIVEN a `downsampling` list with 11 or more entries
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error indicating the maximum of 10 items
