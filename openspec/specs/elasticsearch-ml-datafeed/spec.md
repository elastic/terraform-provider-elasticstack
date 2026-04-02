# `elasticstack_elasticsearch_ml_datafeed` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/datafeed`

## Purpose

Define schema and behavior for the Elasticsearch ML datafeed resource: API usage, identity and import, connection, lifecycle (force-new attributes), create/read/update/delete flows (including stop-before-update and stop-before-delete), datafeed state management, and mapping between Terraform configuration and the Elasticsearch Machine Learning Datafeeds API.

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_datafeed" "example" {
  id          = <computed, string>  # internal identifier: <cluster_uuid>/<datafeed_id>
  datafeed_id = <required, string>  # force new; 1–64 chars; lowercase alphanumeric, hyphens, underscores; must start and end with alphanumeric
  job_id      = <required, string>  # force new; identifier for the associated anomaly detection job

  indices = <required, list(string)>  # min 1 element

  query            = <optional+computed, string>  # JSON (normalized) string
  aggregations     = <optional, string>           # JSON (normalized) string; conflicts with script_fields
  script_fields    = <optional, string>           # JSON string with defaults applied; conflicts with aggregations
  runtime_mappings = <optional, string>           # JSON (normalized) string

  scroll_size       = <optional+computed, int64>   # >= 1
  frequency         = <optional+computed, string>  # Elastic duration string; must match /^\d+[nsumdh]$/
  query_delay       = <optional+computed, string>  # Elastic duration string; must match /^\d+[nsumdh]$/
  max_empty_searches = <optional, int64>           # >= 1

  chunking_config {                  # optional+computed
    mode      = <required, string>   # one of: auto, manual, off
    time_span = <optional+computed, string>  # required when mode=manual; must match /^\d+[nsumdh]$/; only allowed when mode=manual
  }

  delayed_data_check_config {        # optional+computed
    enabled      = <required, bool>
    check_window = <optional+computed, string>  # must match /^\d+[nsumdh]$/
  }

  indices_options {                  # optional+computed
    expand_wildcards   = <optional+computed, list(string)>  # values: all, open, closed, hidden, none
    ignore_unavailable = <optional+computed, bool>
    allow_no_indices   = <optional+computed, bool>
    ignore_throttled   = <optional+computed, bool>           # deprecated
  }

  elasticsearch_connection {         # optional, deprecated
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

### Requirement: Datafeed CRUD APIs (REQ-001–REQ-006)

The resource SHALL use the Elasticsearch Put Datafeed API to create datafeeds ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-put-datafeed.html)). The resource SHALL use the Elasticsearch Update Datafeed API to update datafeeds ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-update-datafeed.html)). The resource SHALL use the Elasticsearch Get Datafeeds API to read datafeed definitions ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-datafeed.html)). The resource SHALL use the Elasticsearch Stop Datafeed API to stop a running datafeed before updating or deleting it ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-stop-datafeed.html)). The resource SHALL use the Elasticsearch Delete Datafeed API to delete datafeeds ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-datafeed.html)). When Elasticsearch returns a non-success status for any API call, the resource SHALL surface the API error as a Terraform diagnostic.

#### Scenario: API failure on create

- GIVEN a non-success response from the Put Datafeed API
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

#### Scenario: API failure on delete

- GIVEN a non-success response from the Delete Datafeed API
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Identity and import (REQ-007–REQ-009)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<datafeed_id>`. During create and update, the resource SHALL derive `id` by calling `r.client.ID(ctx, datafeedID)` to obtain the cluster UUID and `datafeed_id`, and SHALL set `id` in state after a successful API call. The resource SHALL support import by accepting an `id` in the format `<cluster_uuid>/<datafeed_id>`, parsing it with `clients.CompositeIDFromStr`, and persisting both `id` and `datafeed_id` to state. When the import `id` format is invalid, the resource SHALL return an error diagnostic.

#### Scenario: Import with valid composite id

- GIVEN import with a valid `<cluster_uuid>/<datafeed_id>` id
- WHEN import completes
- THEN `id` and `datafeed_id` SHALL be stored in state and read SHALL populate all remaining attributes

#### Scenario: Import with invalid id format

- GIVEN import with an id that is not in `<cluster_uuid>/<datafeed_id>` format
- WHEN import runs
- THEN the resource SHALL return an error diagnostic

### Requirement: Lifecycle — force-new attributes (REQ-010–REQ-011)

Changing `datafeed_id` SHALL require resource replacement. Changing `job_id` SHALL require resource replacement.

#### Scenario: datafeed_id change triggers replacement

- GIVEN an existing datafeed
- WHEN the `datafeed_id` attribute is changed in configuration
- THEN Terraform SHALL plan a destroy-and-recreate (force new)

#### Scenario: job_id change triggers replacement

- GIVEN an existing datafeed
- WHEN the `job_id` attribute is changed in configuration
- THEN Terraform SHALL plan a destroy-and-recreate (force new)

### Requirement: Connection (REQ-012)

By default, the resource SHALL use the provider-level Elasticsearch client obtained via `clients.ConvertProviderData`. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (create, read, update, delete).

#### Scenario: Resource-level client override

- GIVEN `elasticsearch_connection` is set with specific endpoints and credentials
- WHEN any API call is made
- THEN the resource-scoped client SHALL be used instead of the provider client

### Requirement: Create and read-after-write (REQ-013–REQ-014)

On create, the resource SHALL call the Put Datafeed API with a request body built from the plan. After a successful Put Datafeed call, the resource SHALL call read to refresh state from the server response. If the datafeed is not found after creation, the resource SHALL return an error diagnostic ("Failed to read created datafeed").

#### Scenario: State refreshed after create

- GIVEN a successful Put Datafeed API call
- WHEN create completes
- THEN the resource SHALL call read to populate state from the API response

#### Scenario: Datafeed not found after creation

- GIVEN a successful Put Datafeed API call followed by a not-found read response
- WHEN create runs
- THEN the resource SHALL return an error diagnostic

### Requirement: Update — stop-before-update with optional restart (REQ-015–REQ-017)

On update, the resource SHALL check the current datafeed state before calling the Update Datafeed API. If the datafeed is in `started` or `starting` state, the resource SHALL call Stop Datafeed and SHALL wait for the datafeed to reach `stopped` state before proceeding. After a successful Update Datafeed API call, if the datafeed was running before the update, the resource SHALL call Start Datafeed and SHALL wait for the datafeed to reach `started` state. After update (and optional restart), the resource SHALL call read to refresh state. The `job_id` field SHALL NOT be included in the Update Datafeed request body.

#### Scenario: Running datafeed stopped before update

- GIVEN a datafeed in `started` state
- WHEN update runs
- THEN the resource SHALL call Stop Datafeed before calling Update Datafeed

#### Scenario: Datafeed restarted after update if it was running

- GIVEN a datafeed that was in `started` state before update
- WHEN update completes
- THEN the resource SHALL call Start Datafeed after the Update Datafeed API call

#### Scenario: Stopped datafeed not restarted after update

- GIVEN a datafeed in `stopped` state
- WHEN update runs
- THEN the resource SHALL NOT call Start Datafeed after the update

#### Scenario: job_id excluded from update body

- GIVEN an existing datafeed
- WHEN any attribute other than job_id is updated
- THEN the Update Datafeed request body SHALL NOT include the `job_id` field

### Requirement: Read — not found handling (REQ-018)

On read, when the Get Datafeeds API returns HTTP 404 or returns no datafeed matching the requested `datafeed_id`, the resource SHALL remove itself from state. A non-404, non-success response SHALL be surfaced as an error diagnostic.

#### Scenario: Datafeed not found on refresh

- GIVEN a datafeed that has been deleted outside of Terraform
- WHEN read runs
- THEN the resource SHALL be removed from state without error

### Requirement: Delete — stop before delete (REQ-019–REQ-020)

On delete, the resource SHALL check the current datafeed state. If the datafeed is in `started` or `starting` state, the resource SHALL call Stop Datafeed and SHALL wait for the datafeed to reach `stopped` state before calling Delete Datafeed. The resource SHALL call the Delete Datafeed API without the `force` flag. A non-success response from the Delete API SHALL be surfaced as an error diagnostic.

#### Scenario: Running datafeed stopped before delete

- GIVEN a datafeed in `started` state
- WHEN delete runs
- THEN the resource SHALL call Stop Datafeed before calling Delete Datafeed

#### Scenario: Stopped datafeed deleted directly

- GIVEN a datafeed in `stopped` state
- WHEN delete runs
- THEN the resource SHALL not call Stop Datafeed and SHALL call Delete Datafeed directly

### Requirement: datafeed_id validation (REQ-021)

The `datafeed_id` attribute SHALL be validated to be between 1 and 64 characters, contain only lowercase alphanumeric characters (a–z and 0–9), hyphens, and underscores, and start and end with an alphanumeric character.

#### Scenario: Invalid datafeed_id rejected

- GIVEN a `datafeed_id` that starts with a hyphen or contains uppercase characters
- WHEN the configuration is applied
- THEN the provider SHALL return a validation error and SHALL not call the API

### Requirement: aggregations and script_fields are mutually exclusive (REQ-022)

`aggregations` and `script_fields` SHALL be mutually exclusive. The schema SHALL enforce this with `ConflictsWith`.

#### Scenario: Both aggregations and script_fields set

- GIVEN both `aggregations` and `script_fields` set in configuration
- WHEN the configuration is applied
- THEN the provider SHALL return a validation error

### Requirement: chunking_config.time_span is only allowed when mode is manual (REQ-023)

The `chunking_config.time_span` attribute SHALL only be allowed when `chunking_config.mode` is `manual`. The schema SHALL enforce this constraint. On create and update, `time_span` SHALL be included in the API request only when mode is `manual` and `time_span` is set.

#### Scenario: time_span set with mode=manual

- GIVEN `chunking_config.mode = "manual"` and `chunking_config.time_span = "1h"`
- WHEN create runs
- THEN the Put Datafeed API request SHALL include `chunking_config.time_span`

#### Scenario: time_span excluded when mode is not manual

- GIVEN `chunking_config.mode = "auto"` and `chunking_config.time_span` not set
- WHEN create runs
- THEN the Put Datafeed API request SHALL NOT include `chunking_config.time_span`

### Requirement: script_fields defaults (REQ-024)

When `script_fields` is set, the resource SHALL apply defaults to each script field before sending to the API: if `ignore_failure` is not specified in a script field, it SHALL default to `false`; if a script field contains a `script` block without a `lang` property, `lang` SHALL default to `"painless"`.

#### Scenario: script_fields default lang applied

- GIVEN a `script_fields` value where a script block does not specify `lang`
- WHEN the configuration is applied
- THEN the effective `lang` for that script SHALL be `"painless"`

### Requirement: Mapping — config to API model (REQ-025–REQ-027)

On create and update, fields that are null or unknown SHALL be omitted from the API request body. The `query`, `aggregations`, and `runtime_mappings` fields SHALL each be decoded from their JSON string representation into `map[string]any` for the API request. The `script_fields` field SHALL be decoded from its JSON string representation into `map[string]any` for the API request. When any JSON field is not valid JSON, the resource SHALL return an error diagnostic and SHALL not call the API.

#### Scenario: Invalid query JSON

- GIVEN an invalid JSON string in `query`
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic and SHALL not call the Put Datafeed or Update Datafeed API

### Requirement: Mapping — API response to state (REQ-028–REQ-033)

On read, the resource SHALL set the following state attributes from the Get Datafeeds API response:
- `datafeed_id` and `job_id` from the corresponding API fields.
- `indices` SHALL be set to null in state when the API returns an empty or nil indices list; otherwise it SHALL be set to the returned list of strings.
- `query` SHALL be JSON-marshaled from the API response `query` map when non-nil; when nil it SHALL be stored as null in state.
- `aggregations` SHALL be JSON-marshaled from the API response `aggregations` map when non-nil; when nil it SHALL be stored as null in state.
- `script_fields` SHALL be JSON-marshaled from the API response `script_fields` map when non-nil; when nil it SHALL be stored as null in state.
- `runtime_mappings` SHALL be JSON-marshaled from the API response `runtime_mappings` map when non-nil; when nil it SHALL be stored as null in state.
- `scroll_size`, `frequency`, `query_delay`, and `max_empty_searches` SHALL be set from the corresponding API fields when non-nil; when nil they SHALL be stored as null in state.
- `chunking_config` SHALL be set from the API response when non-nil; `time_span` SHALL be set only when `mode` is `manual` and the API returns a non-empty `time_span`.
- `delayed_data_check_config` SHALL be set from the API response when non-nil.
- `indices_options` SHALL be set from the API response when non-nil; individual sub-fields that are nil in the API response SHALL be stored as null in state.

#### Scenario: Null query stored as null in state

- GIVEN a datafeed where query is not set on the server
- WHEN read runs
- THEN `query` SHALL be null (not empty string or `{}`) in state

#### Scenario: chunking_config time_span omitted when not manual

- GIVEN a datafeed with `chunking_config.mode = "auto"` on the server
- WHEN read runs
- THEN `chunking_config.time_span` SHALL be null in state

### Requirement: Plan/State — UseStateForUnknown (REQ-034)

The following attributes SHALL use `UseStateForUnknown` plan modifier to preserve prior state values when the plan value is unknown: `id`, `query`, `scroll_size`, `frequency`, `query_delay`, `chunking_config`, `delayed_data_check_config`, `indices_options` (and its sub-fields: `expand_wildcards`, `ignore_unavailable`, `allow_no_indices`, `ignore_throttled`), `chunking_config.time_span`.

#### Scenario: id preserved across plan

- GIVEN an existing datafeed with a known id in state
- WHEN a plan is generated without changing datafeed_id
- THEN `id` SHALL remain known (not unknown) in the plan
