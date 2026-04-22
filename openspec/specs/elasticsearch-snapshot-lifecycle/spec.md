# `elasticstack_elasticsearch_snapshot_lifecycle` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/cluster/slm.go`

## Purpose

Define schema and behavior for the Elasticsearch Snapshot Lifecycle Management (SLM) policy resource: API usage, identity/import, connection, lifecycle, mapping, and read-time state handling for snapshot policies.

## Schema

```hcl
resource "elasticstack_elasticsearch_snapshot_lifecycle" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<policy_name>
  name = <required, string> # force new; the SLM policy ID

  schedule   = <required, string>  # cron expression
  repository = <required, string>  # snapshot repository name

  snapshot_name = <optional, string>  # default: "<snap-{now/d}>"

  # Config
  expand_wildcards     = <optional, string>  # default: "open,hidden"; comma-separated values
  ignore_unavailable   = <optional, bool>    # default: false
  include_global_state = <optional, bool>    # default: true
  indices              = <optional+computed, list(string)>
  feature_states       = <optional+computed, set(string)>
  metadata             = <optional+computed, json string>
  partial              = <optional, bool>    # default: false

  # Retention
  expire_after = <optional, string>  # time period, e.g. "30d"
  max_count    = <optional, int>
  min_count    = <optional, int>

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

### Requirement: SLM policy CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch SLM Put Lifecycle API (`SlmPutLifecycle`) to create and update SLM policies ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-put-policy.html)). The resource SHALL use the Elasticsearch SLM Get Lifecycle API (`SlmGetLifecycle`) to read SLM policies ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-get-policy.html)). The resource SHALL use the Elasticsearch SLM Delete Lifecycle API (`SlmDeleteLifecycle`) to delete SLM policies ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-delete-policy.html)). When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than not found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create/update

- GIVEN a non-success API response from `SlmPutLifecycle`
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error and SHALL not proceed to read

#### Scenario: API failure on delete

- GIVEN a non-success API response from `SlmDeleteLifecycle`
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Identity (REQ-005–REQ-007)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<policy_name>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and the configured `name`. The resource SHALL set `id` in state immediately after a successful Put Lifecycle API call, before performing the read-back.

#### Scenario: ID computation on create

- GIVEN a successful create with `name = "my-policy"`
- WHEN the Put API call completes
- THEN `id` SHALL be set to `<cluster_uuid>/my-policy`

### Requirement: Import (REQ-008–REQ-010)

The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the provided `id` value directly to state. The imported `id` SHALL be in the format `<cluster_uuid>/<policy_name>`. On read operations (including after import), the resource SHALL parse the `id` using `CompositeIDFromStr` and SHALL return an error diagnostic when the `id` does not contain exactly one `/` separator (i.e. does not match `<cluster_uuid>/<resource_identifier>`).

#### Scenario: Import passthrough

- GIVEN import with a valid composite id `<cluster_uuid>/my-policy`
- WHEN import completes
- THEN the id SHALL be stored and subsequent reads SHALL use `my-policy` as the policy name

#### Scenario: Invalid id format on read

- GIVEN an `id` in state that does not match `<cluster_uuid>/<resource_identifier>`
- WHEN a read, update, or delete operation runs
- THEN the resource SHALL return a "Wrong resource ID" error diagnostic

### Requirement: Lifecycle (REQ-011)

Changing `name` SHALL require resource replacement (`ForceNew`). All other attributes may be updated in place without replacement.

#### Scenario: Name change triggers replace

- GIVEN an existing SLM policy resource
- WHEN `name` is changed in configuration
- THEN Terraform SHALL plan a destroy-and-recreate of the resource

### Requirement: Connection (REQ-012–REQ-013)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (create, update, read, delete).

#### Scenario: Resource-level client override

- GIVEN `elasticsearch_connection` is set with custom endpoints
- WHEN any API call (create, update, read, delete) runs
- THEN the resource-scoped client SHALL be used instead of the provider client

### Requirement: Create and update (REQ-014–REQ-016)

On create and update, the resource SHALL build a `models.SnapshotPolicy` from Terraform config, populating `repository`, `schedule`, `name` (snapshot name template), `config`, and `retention`. The resource SHALL send this model to the Put Lifecycle API serialized as JSON. After a successful Put, the resource SHALL call read to refresh state.

#### Scenario: Successful create triggers read-back

- GIVEN a valid SLM policy configuration
- WHEN create completes successfully
- THEN a read SHALL be performed to refresh state attributes

### Requirement: Read (REQ-017–REQ-019)

On read, the resource SHALL parse `id` to extract the policy name, then call the Get Lifecycle API with that name. When the Get Lifecycle API returns HTTP 404, the resource SHALL remove the resource from state (set id to empty) and return without error. When the policy is found in the response, the resource SHALL set `name` from the parsed id and map all API response fields (`snapshot_name`, `repository`, `schedule`, `expire_after`, `max_count`, `min_count`, `expand_wildcards`, `include_global_state`, `ignore_unavailable`, `partial`, `metadata`, `indices`, `feature_states`) to state. When the policy name is not found in the API response body (despite a 200 status), the resource SHALL return an error diagnostic.

#### Scenario: Policy not found removes from state

- GIVEN a policy that no longer exists in Elasticsearch
- WHEN read runs
- THEN the resource SHALL be removed from Terraform state (id set to empty)

#### Scenario: Policy name absent in response body

- GIVEN a 200 response that does not contain the requested policy name in its body
- WHEN read runs
- THEN the resource SHALL return an error diagnostic indicating the policy was not found in the response

### Requirement: Delete (REQ-020)

On delete, the resource SHALL parse `id` to extract the policy name and call the Delete Lifecycle API with that name. Non-success responses SHALL be surfaced as error diagnostics.

#### Scenario: Successful delete

- GIVEN an existing SLM policy
- WHEN delete runs
- THEN the Delete Lifecycle API SHALL be called with the policy name from the parsed id

### Requirement: Field mapping — `expand_wildcards` validation (REQ-021–REQ-022)

The `expand_wildcards` attribute SHALL accept a comma-separated string of values. Each comma-separated token SHALL be validated against the allowed set: `all`, `open`, `closed`, `hidden`, `none`. When a token is not in this set, the resource SHALL return an error diagnostic with the invalid value and SHALL not call the API.

#### Scenario: Invalid expand_wildcards token

- GIVEN `expand_wildcards = "open,invalid"`
- WHEN the plan is applied
- THEN the resource SHALL return an error diagnostic identifying `"invalid"` as not a valid value

#### Scenario: Valid comma-separated expand_wildcards

- GIVEN `expand_wildcards = "open,hidden"`
- WHEN the plan is applied
- THEN the resource SHALL accept the value and proceed with the API call

### Requirement: Field mapping — metadata JSON (REQ-023–REQ-024)

The `metadata` attribute SHALL be validated as a JSON string by the schema (`validation.StringIsJSON`). On create/update, when `metadata` is set, the resource SHALL decode it from JSON into a `map[string]any` and include it in the `config.metadata` field of the API payload. On read, when the API response includes `config.metadata`, the resource SHALL marshal it back to a JSON string and store it in state. JSON diff suppression SHALL prevent spurious diffs due to key reordering or whitespace differences in the `metadata` attribute.

#### Scenario: Invalid metadata JSON

- GIVEN `metadata = "not-json"`
- WHEN create/update runs
- THEN the resource SHALL return an error diagnostic and SHALL not call the Put API

#### Scenario: Metadata round-trip

- GIVEN `metadata = jsonencode({created_by = "terraform"})`
- WHEN create followed by read runs
- THEN state SHALL contain the metadata as a normalized JSON string

### Requirement: Field mapping — optional pointer fields (REQ-025–REQ-027)

The `expire_after`, `max_count`, and `min_count` retention fields SHALL be sent in the API payload only when they are set in configuration (`GetOk` returns true); when absent, they SHALL be omitted from the `retention` object. The `ignore_unavailable` and `include_global_state` config fields SHALL always be sent in the API payload (they have schema defaults and are always present as bools). The `expand_wildcards` and `partial` config fields SHALL be sent only when explicitly set in configuration. On read, optional retention and config fields SHALL be set in state only when the API response includes them (non-nil pointer values); absent fields SHALL not overwrite state.

#### Scenario: Absent retention fields omitted from API payload

- GIVEN no `expire_after`, `max_count`, or `min_count` configured
- WHEN create/update runs
- THEN the API payload retention object SHALL not include those fields

### Requirement: Field mapping — indices and feature_states (REQ-028–REQ-029)

The `indices` attribute (list) and `feature_states` attribute (set) SHALL be sent in the API payload as string arrays when they contain at least one element. When not configured or empty, these fields SHALL be omitted from the API payload (`omitempty`). On read, `indices` and `feature_states` SHALL be set from the API response arrays.

#### Scenario: Indices and feature_states round-trip

- GIVEN `indices = ["data-*", "abc"]` and `feature_states = ["ILM"]`
- WHEN create followed by read runs
- THEN state SHALL reflect those indices and feature states as returned by the API
