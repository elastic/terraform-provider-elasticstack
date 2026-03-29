# `elasticstack_elasticsearch_script` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/cluster/script`

## Purpose

Define schema and behavior for the Elasticsearch stored script resource: API usage, identity/import, connection, lifecycle, and state mapping including fields not returned by the API.

## Schema

```hcl
resource "elasticstack_elasticsearch_script" "example" {
  id        = <computed, string> # internal identifier: <cluster_uuid>/<script_id>
  script_id = <required, string> # force new; unique identifier for the stored script
  lang      = <required, string> # one of: painless, expression, mustache, java
  source    = <required, string> # script body or search template content

  params  = <optional, json string> # JSON object of default parameters
  context = <optional, string>      # execution context; not returned by the API

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

### Requirement: Stored script CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put Stored Script API to create and update stored scripts ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/create-stored-script-api.html)). The resource SHALL use the Elasticsearch Get Stored Script API to read stored scripts ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/get-stored-script-api.html)). The resource SHALL use the Elasticsearch Delete Stored Script API to delete stored scripts ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/delete-stored-script-api.html)). When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than not found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure surfaces error

- GIVEN Elasticsearch returns a non-success response (other than 404 on read)
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the API error

### Requirement: Identity and import (REQ-005–REQ-008)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<script_id>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and configured `script_id`. The resource SHALL support import via `ImportStatePassthroughID`, persisting the imported `id` value directly to state. Read and delete operations SHALL require the `id` to be in the format `<cluster_uuid>/<script_id>` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Import passthrough

- GIVEN import with a valid composite id of the form `<cluster_uuid>/<script_id>`
- WHEN import completes
- THEN the id SHALL be stored in state for subsequent operations

#### Scenario: Invalid id format on read

- GIVEN an `id` that does not contain exactly one `/` separator
- WHEN read or delete runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Lifecycle (REQ-009)

Changing `script_id` SHALL require replacement of the resource.

#### Scenario: script_id change triggers replacement

- GIVEN an existing script resource
- WHEN `script_id` is changed in configuration
- THEN Terraform SHALL plan a resource replacement

### Requirement: Connection (REQ-010–REQ-011)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls.

#### Scenario: Resource-level client override

- GIVEN `elasticsearch_connection` is configured on the resource
- WHEN create, read, update, or delete runs
- THEN API calls SHALL use the resource-scoped client, not the provider-level client

### Requirement: Create and update (REQ-012–REQ-014)

On create and update, the resource SHALL build a stored script request body from the plan (`lang`, `source`, `params`, and `context`) and submit it with the Put Stored Script API using `script_id` as the identifier. After a successful Put request, the resource SHALL perform a read from Elasticsearch to refresh state, then set `id` to the composite `<cluster_uuid>/<script_id>` value. When the Put request fails, the resource SHALL surface the API error to diagnostics and SHALL not update state.

#### Scenario: Successful create sets id and refreshes state

- GIVEN valid plan configuration
- WHEN create runs successfully
- THEN state SHALL contain a computed `id` of the form `<cluster_uuid>/<script_id>` and the attributes read back from Elasticsearch

### Requirement: Read (REQ-015–REQ-017)

On read, the resource SHALL parse `id` to extract `script_id` and call the Get Stored Script API. When the API returns HTTP 404, the resource SHALL remove itself from state. When the API returns a non-error response, the resource SHALL set `script_id`, `lang`, and `source` from the response. On every read, the resource SHALL preserve `id`, `elasticsearch_connection`, and `context` from prior state, because `context` is not returned by the Elasticsearch API.

#### Scenario: Script not found removes from state

- GIVEN a script that has been deleted outside Terraform
- WHEN read runs
- THEN the resource SHALL be removed from state without error

#### Scenario: context preserved on refresh

- GIVEN a script resource with `context` configured
- WHEN read refreshes state from Elasticsearch
- THEN `context` SHALL be preserved in state unchanged

### Requirement: Delete (REQ-018–REQ-019)

On delete, the resource SHALL parse `id` to extract `script_id` and call the Delete Stored Script API using that identifier. When the Delete request fails, the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: Successful delete

- GIVEN an existing script resource
- WHEN delete runs
- THEN the Delete Stored Script API SHALL be called with the `script_id` extracted from state `id`

### Requirement: Mapping — script language validation (REQ-020)

The `lang` attribute SHALL be validated to accept only `painless`, `expression`, `mustache`, or `java`. Any other value SHALL result in a validation error before any API call is made.

#### Scenario: Invalid language rejected

- GIVEN `lang` set to a value not in the allowed set
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned and no API call SHALL be made

### Requirement: Mapping — params JSON (REQ-021–REQ-023)

`params` SHALL be validated as a normalized JSON string by schema. On create and update, when `params` is set and non-empty, the resource SHALL parse the JSON string into an object and include it in the Put request body. When the JSON cannot be unmarshaled, the resource SHALL return an error diagnostic and SHALL not call the Put API. On read, when the API returns non-empty params, the resource SHALL marshal them into a normalized JSON string and store them in state; when the API returns no params, the resource SHALL preserve `params` from prior state to avoid drift.

#### Scenario: params preserved when API omits them

- GIVEN a script resource with `params` configured
- WHEN the Elasticsearch API returns the script without params
- THEN `params` SHALL be preserved from prior state unchanged

#### Scenario: Invalid params JSON rejected on create

- GIVEN `params` set to an invalid JSON string
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic and SHALL not call the Put API

### Requirement: Mapping — context not returned by API (REQ-024)

The `context` attribute is accepted on create and update and passed to the Put Stored Script API as the script context parameter. Because the Elasticsearch Get Stored Script API does not return the `context` field, the resource SHALL always preserve `context` from prior state on read and after create/update. Import operations SHALL not be able to restore `context` because it is not available in the API response.

#### Scenario: context not verified on import

- GIVEN a script resource imported by id
- WHEN import completes and state is verified
- THEN `context` SHALL not be compared against the API response during `ImportStateVerify`
