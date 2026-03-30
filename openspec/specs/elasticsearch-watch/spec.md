# `elasticstack_elasticsearch_watch` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/watcher/watch.go`

## Purpose

Define schema and behavior for the Elasticsearch Watcher resource: API usage, identity/import, connection, JSON mapping, and read-time state synchronization. Manages Elasticsearch watches via the Watcher API, enabling scheduled or event-driven automated actions.

## Schema

```hcl
resource "elasticstack_elasticsearch_watch" "example" {
  id       = <computed, string> # internal identifier: <cluster_uuid>/<watch_id>
  watch_id = <required, string> # force new

  active   = <optional, bool>   # default: true
  trigger  = <required, string> # JSON object
  input    = <optional, string> # JSON object; default: {"none":{}}
  condition = <optional, string> # JSON object; default: {"always":{}}
  actions   = <optional, string> # JSON object; default: {}
  metadata  = <optional, string> # JSON object; default: {}
  transform = <optional, string> # JSON object; omitted when not set

  throttle_period_in_millis = <optional, int> # default: 5000

  elasticsearch_connection {
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

### Requirement: Watcher CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put Watch API to create and update watches ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-put-watch.html)). The resource SHALL use the Elasticsearch Get Watch API to read a watch ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-get-watch.html)). The resource SHALL use the Elasticsearch Delete Watch API to delete a watch ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-delete-watch.html)). When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than 404 on read), the resource SHALL surface the API error as a Terraform diagnostic.

#### Scenario: API failure on create

- GIVEN the Put Watch API returns a non-success response
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error from the API

#### Scenario: API failure on delete

- GIVEN the Delete Watch API returns a non-success response
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error from the API

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` attribute representing the watch in the format `<cluster_uuid>/<watch_id>`. During create and update, the resource SHALL compute `id` by combining the current cluster UUID with the configured `watch_id` value.

#### Scenario: ID set on create

- GIVEN a successful Put Watch API call
- WHEN create completes
- THEN the `id` in state SHALL be in the format `<cluster_uuid>/<watch_id>`

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state. For import and all subsequent read/delete operations, the resource SHALL require the `id` to be in the format `<cluster_uuid>/<watch_id>` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Import with valid composite id

- GIVEN an `id` in the format `<cluster_uuid>/<watch_id>`
- WHEN import completes
- THEN the `id` SHALL be stored in state for subsequent operations

#### Scenario: Invalid id format

- GIVEN a stored or imported `id` that does not contain exactly one `/`
- WHEN read or delete runs
- THEN the resource SHALL return an error diagnostic with "Wrong resource ID"

### Requirement: Lifecycle (REQ-009)

Changing `watch_id` SHALL require replacement (`ForceNew`). Updates to all other attributes SHALL be applied in place using the Put Watch API.

#### Scenario: watch_id change triggers replacement

- GIVEN an existing watch resource
- WHEN `watch_id` is changed in configuration
- THEN Terraform SHALL plan a destroy-and-recreate for the resource

### Requirement: Connection (REQ-010–REQ-011)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured on the resource, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls.

#### Scenario: Resource-level connection override

- GIVEN `elasticsearch_connection` is set with custom endpoints or credentials
- WHEN create, read, update, or delete runs
- THEN all API calls SHALL use the resource-scoped client instead of the provider client

### Requirement: Create and update (REQ-012–REQ-013)

On create and update, the resource SHALL construct a watch body from Terraform config, including `trigger`, `input`, `condition`, `actions`, `metadata`, optional `transform`, and `throttle_period_in_millis`, and submit it to the Put Watch API along with the `active` flag. After a successful Put Watch call, the resource SHALL set `id` in state and perform a read to refresh state.

#### Scenario: Read-back after create

- GIVEN a successful Put Watch API call on create
- WHEN create finishes
- THEN the resource SHALL perform a GET to refresh all state attributes

### Requirement: Read (REQ-014–REQ-016)

On read, the resource SHALL parse `id` using `clients.ResourceIDFromStr` to extract the watch identifier. The resource SHALL call the Get Watch API with the extracted watch identifier. When the Get Watch API returns 404, the resource SHALL remove itself from state (set id to `""`). When the API returns a successful response, the resource SHALL decode the JSON response and update all state attributes from the response.

#### Scenario: Watch not found on refresh

- GIVEN the watch no longer exists on the cluster
- WHEN read runs
- THEN the resource SHALL be removed from state without an error

### Requirement: Delete (REQ-017)

On delete, the resource SHALL parse `id` to extract the watch identifier and call the Delete Watch API with that identifier. The resource SHALL surface any non-success API response as a diagnostic error.

#### Scenario: Successful delete

- GIVEN a watch exists in state
- WHEN delete runs
- THEN the resource SHALL call the Delete Watch API with the watch identifier from `id`

### Requirement: JSON field mapping — create/update (REQ-018–REQ-022)

On create and update, the resource SHALL unmarshal each JSON string attribute (`trigger`, `input`, `condition`, `actions`, `metadata`) into a `map[string]any` before constructing the API request body; if any unmarshal fails, the resource SHALL return a diagnostic error and SHALL NOT call the Put Watch API. The `transform` attribute SHALL be included in the API request body only when it is present in config; if present, it SHALL be unmarshalled into a `map[string]any` before inclusion. The `throttle_period_in_millis` value SHALL be included in the request body when non-zero. The `active` flag SHALL be passed as a query parameter to the Put Watch API.

#### Scenario: Invalid JSON in trigger

- GIVEN `trigger` is set to a non-JSON string
- WHEN create or update runs
- THEN the resource SHALL return a diagnostic error and SHALL NOT call the Put Watch API

#### Scenario: transform omitted when not set

- GIVEN `transform` is not configured
- WHEN create or update builds the request body
- THEN the `transform` field SHALL be omitted from the API request body

### Requirement: JSON field mapping — read/state (REQ-023–REQ-027)

On read, the resource SHALL marshal the API response fields `trigger`, `input`, `condition`, `actions`, and `metadata` back into JSON strings and store them in state. When the API response includes a non-nil `transform`, the resource SHALL marshal it to a JSON string and store it in state; when the API response has a nil `transform`, the resource SHALL NOT overwrite the `transform` state attribute. The resource SHALL store `watch_id` and `active` (from `watch.status.state.active`) directly from the API response. The resource SHALL store `throttle_period_in_millis` from the API response. JSON fields SHALL use `DiffSuppressFunc` (`tfsdkutils.DiffJSONSuppress`) to suppress semantically equivalent JSON diffs.

#### Scenario: transform nil in API response

- GIVEN the API response has no transform field
- WHEN read runs
- THEN `transform` in state SHALL not be overwritten (remains as previously stored)

#### Scenario: active synced from watch status

- GIVEN the watch is deactivated on the cluster
- WHEN read runs
- THEN `active` in state SHALL reflect `watch.status.state.active` from the API response
