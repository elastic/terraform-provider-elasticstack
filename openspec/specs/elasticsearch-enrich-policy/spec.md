# `elasticstack_elasticsearch_enrich_policy` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/enrich`
Data source implementation: `internal/elasticsearch/enrich`

## Purpose

Define schema and behavior for the Elasticsearch enrich policy resource and data source: API usage, identity, connection, lifecycle (including policy execution), create/delete flow, read/state-refresh semantics, and query mapping.

## Schema

```hcl
resource "elasticstack_elasticsearch_enrich_policy" "example" {
  id            = <computed, string>                    # internal identifier: <cluster_uuid>/<policy_name>
  name          = <required, string>                    # 1–255 chars; force new
  policy_type   = <required, string>                    # one of: geo_match, match, range; force new
  indices       = <required, set(string)>               # at least 1 element; force new
  match_field   = <required, string>                    # 1–255 chars; force new
  enrich_fields = <required, set(string)>               # at least 1 element; force new
  query         = <optional, json normalized string>    # force new; omitted when null
  execute       = <optional, computed, bool>            # default true; force new

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

data "elasticstack_elasticsearch_enrich_policy" "example" {
  id            = <computed, string>          # internal identifier: <cluster_uuid>/<policy_name>
  name          = <required, string>          # policy to look up
  policy_type   = <computed, string>
  indices       = <computed, set(string)>
  match_field   = <computed, string>
  enrich_fields = <computed, set(string)>
  query         = <computed, json normalized string>

  elasticsearch_connection {
    # same attributes as resource
  }
}
```

## Requirements

### Requirement: Enrich policy APIs (REQ-001–REQ-003)

The resource SHALL use the Put enrich policy API to create policies, the Get enrich policy API to read policies, and the Delete enrich policy API to delete policies ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-apis.html)). Non-success API responses (other than 404 on read) SHALL be surfaced as Terraform diagnostics.

#### Scenario: API errors surfaced

- GIVEN a failing Elasticsearch response (other than 404 on read)
- WHEN the provider processes the response
- THEN diagnostics SHALL include the API error

### Requirement: Policy execution (REQ-004–REQ-005)

When `execute` is `true` (the default), the resource SHALL call the Elasticsearch Execute enrich policy API after a successful Put ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/execute-enrich-policy-api.html)), waiting for completion. When the execute API returns a non-OK HTTP status, or the response phase is not `COMPLETE`, the resource SHALL return an error diagnostic and SHALL NOT set state.

#### Scenario: Successful execute

- GIVEN `execute = true`
- WHEN create runs and Put succeeds
- THEN Execute enrich policy SHALL be called with `wait_for_completion=true`

#### Scenario: Execute returns unexpected phase

- GIVEN `execute = true` and the execute API responds with a phase other than `COMPLETE`
- WHEN create runs
- THEN the resource SHALL return an error diagnostic

#### Scenario: Execute omitted when false

- GIVEN `execute = false`
- WHEN create runs and Put succeeds
- THEN Execute enrich policy SHALL NOT be called

### Requirement: Identity (REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<policy_name>`. The resource SHALL compute `id` from the current cluster UUID and the configured `name` after a successful Put.

#### Scenario: ID set after create

- GIVEN a successful Put enrich policy call
- WHEN create completes
- THEN `id` in state SHALL be `<cluster_uuid>/<policy_name>`

### Requirement: Lifecycle — all attributes require replacement (REQ-007)

Changing any of `name`, `policy_type`, `indices`, `match_field`, `enrich_fields`, `query`, or `execute` SHALL require resource replacement, because enrich policies are immutable once created.

#### Scenario: Name change triggers replace

- GIVEN `name` changes in configuration
- WHEN Terraform plans
- THEN replacement SHALL be required

#### Scenario: Query change triggers replace

- GIVEN `query` changes in configuration
- WHEN Terraform plans
- THEN replacement SHALL be required

### Requirement: Read and state refresh (REQ-008–REQ-010)

On read, the resource SHALL parse `id` in the format `<cluster_uuid>/<policy_name>`; an invalid format SHALL produce an error diagnostic. The resource SHALL call the Get enrich policy API with the parsed name. When the policy is not found, the resource SHALL remove itself from state. When found, the resource SHALL populate all policy attributes from the API response.

#### Scenario: Policy not found on refresh

- GIVEN the policy was deleted in Elasticsearch
- WHEN read runs
- THEN the resource SHALL be removed from state

#### Scenario: Invalid id on read

- GIVEN a malformed `id` in state (not `<cluster_uuid>/<policy_name>`)
- WHEN read runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Delete (REQ-011)

On delete, the resource SHALL parse `id`, derive the policy name, and call the Delete enrich policy API. Non-success responses SHALL be surfaced as diagnostics.

#### Scenario: Delete by parsed name

- GIVEN destroy
- WHEN delete runs
- THEN Delete enrich policy SHALL be called with the policy name parsed from `id`

### Requirement: Connection (REQ-012)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured on the resource, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (Put, Get, Delete, Execute).

#### Scenario: Resource-level connection override

- GIVEN `elasticsearch_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource-scoped client SHALL be used instead of the provider client

### Requirement: Query mapping (REQ-013–REQ-015)

When `query` is set in configuration, the resource SHALL send it as a parsed JSON object in the `query` field of the Put request body. When `query` is null or not configured, the resource SHALL omit the `query` field from the Put request body entirely. On read, when the API response includes a `query` field that is non-null and non-empty, the resource SHALL store it in state as a normalized JSON string; otherwise `query` SHALL be stored as null.

#### Scenario: Query sent as JSON object

- GIVEN `query` is set to a valid JSON string
- WHEN create runs
- THEN the Put request body SHALL contain `query` as a parsed JSON object

#### Scenario: Null query omitted from request

- GIVEN `query` is not configured
- WHEN create runs
- THEN the Put request body SHALL not contain a `query` field

#### Scenario: Null query preserved in state

- GIVEN the API response has no `query` field or a null query
- WHEN read runs
- THEN `query` in state SHALL be null

### Requirement: Policy type mapping (REQ-016)

The resource SHALL validate that `policy_type` is one of `geo_match`, `match`, or `range` at plan time. On create, the resource SHALL use `policy_type` as the top-level key in the Put enrich policy request body (e.g. `{"match": {...}}`).

#### Scenario: Invalid policy type rejected

- GIVEN `policy_type = "invalid"`
- WHEN Terraform plans
- THEN the provider SHALL return a validation error

### Requirement: Indices and enrich_fields validation (REQ-017)

The resource SHALL validate that `indices` contains at least one element and that `enrich_fields` contains at least one element at plan time.

#### Scenario: Empty indices rejected

- GIVEN `indices = []`
- WHEN Terraform plans
- THEN the provider SHALL return a validation error

### Requirement: Data source read (REQ-018–REQ-020)

The data source SHALL use the Elasticsearch Get enrich policy API to read an enrich policy by `name`. When the policy is not found (404 or empty response), the data source SHALL return an error diagnostic. The data source SHALL populate `id`, `policy_type`, `indices`, `match_field`, `enrich_fields`, and `query` from the API response.

#### Scenario: Data source policy not found

- GIVEN a `name` that does not exist in Elasticsearch
- WHEN the data source reads
- THEN diagnostics SHALL include a "Policy not found" error

#### Scenario: Data source populates all fields

- GIVEN a policy named `name` exists in Elasticsearch
- WHEN the data source reads
- THEN `policy_type`, `indices`, `match_field`, `enrich_fields`, and `query` SHALL reflect the API response

### Requirement: Data source identity (REQ-021)

The data source SHALL expose a computed `id` in the format `<cluster_uuid>/<policy_name>`, derived from the cluster UUID and the configured `name`.

#### Scenario: Data source ID set

- GIVEN a successful Get enrich policy call
- WHEN the data source reads
- THEN `id` in state SHALL be `<cluster_uuid>/<policy_name>`

### Requirement: Data source connection (REQ-022)

By default, the data source SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured on the data source, the data source SHALL construct and use a resource-scoped Elasticsearch client for all API calls.

#### Scenario: Data source connection override

- GIVEN `elasticsearch_connection` is configured on the data source
- WHEN the data source reads
- THEN the resource-scoped client SHALL be used instead of the provider client
