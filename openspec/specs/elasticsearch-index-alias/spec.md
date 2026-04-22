# `elasticstack_elasticsearch_index_alias` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/alias`

## Purpose

Define schema and behavior for the Elasticsearch index alias resource: API usage, identity/import, connection, validation, atomic alias management (create, update, delete), read-time state mapping, and lifecycle.

## Schema

```hcl
resource "elasticstack_elasticsearch_index_alias" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<alias_name>
  name = <required, string> # force new

  write_index = <optional, single nested object> {
    name          = <required, string>
    filter        = <optional, JSON (normalized) string>
    index_routing = <optional, string>
    is_hidden     = <optional+computed, bool> # default false
    routing       = <optional, string>
    search_routing = <optional, string>
  }

  read_indices = <optional, set of nested objects> {
    name          = <required, string>
    filter        = <optional, JSON (normalized) string>
    index_routing = <optional, string>
    is_hidden     = <optional+computed, bool> # default false
    routing       = <optional, string>
    search_routing = <optional, string>
  }

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

### Requirement: Alias CRUD APIs (REQ-001–REQ-003)

The resource SHALL use the Elasticsearch Get Alias API to read alias-to-index associations ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-alias.html)). The resource SHALL use the Elasticsearch Update Aliases API to create, update, and delete aliases atomically using `add` and `remove` actions ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-aliases.html)). When Elasticsearch returns a non-success response (other than 404 on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create

- GIVEN the Update Aliases API returns a non-success response
- WHEN create runs
- THEN Terraform diagnostics SHALL include the error

#### Scenario: 404 treated as not found on read

- GIVEN the Get Alias API returns HTTP 404
- WHEN read runs
- THEN the resource SHALL remove itself from state without an error

### Requirement: Identity (REQ-004–REQ-005)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<alias_name>`. During create, the resource SHALL compute `id` from the current cluster UUID and the configured `name` attribute.

#### Scenario: ID format

- GIVEN a successful create
- WHEN the provider sets the ID
- THEN `id` SHALL be `<cluster_uuid>/<alias_name>`

### Requirement: Import (REQ-006)

The resource SHALL support import via `ImportStatePassthroughID`, persisting the supplied `id` directly to state. The resource uses `name` from state (not from the parsed id) for read operations; it does not re-derive the alias name from `id` on read.

#### Scenario: Import passthrough

- GIVEN an import command with a valid composite id
- WHEN import completes
- THEN the id SHALL be stored in state for subsequent operations

### Requirement: Lifecycle (REQ-007)

Changing `name` SHALL require replacement of the resource (`RequiresReplace`). The computed `id` SHALL be preserved across plan/apply cycles using `UseStateForUnknown`.

#### Scenario: Name change triggers replace

- GIVEN an existing alias
- WHEN `name` is changed in configuration
- THEN Terraform SHALL plan a replace (destroy + create)

### Requirement: Connection (REQ-008)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (create, read, update, delete).

#### Scenario: Resource-level client

- GIVEN `elasticsearch_connection` is set
- WHEN API calls run
- THEN they SHALL use the resource-scoped client

### Requirement: Configuration validation (REQ-009)

The resource SHALL validate that the index named in `write_index.name` does not also appear in any entry of `read_indices`. If the same index name appears in both roles, the resource SHALL return an "Invalid Configuration" error diagnostic and SHALL NOT proceed to create or update. This validation is applied both at plan time (`ValidateConfig`) and at apply time before API calls.

#### Scenario: Write index in read indices

- GIVEN `write_index.name` equals one of the `read_indices[*].name` values
- WHEN config is validated
- THEN the provider SHALL return an "Invalid Configuration" error diagnostic

### Requirement: Create (REQ-010–REQ-012)

On create, the resource SHALL convert `write_index` (if set) and all entries in `read_indices` (if set) into `add` alias actions and submit them in a single atomic Update Aliases API call. The `write_index` entry SHALL be submitted with `is_write_index: true`; `read_indices` entries SHALL be submitted without `is_write_index`. After a successful API call, the resource SHALL perform a read to refresh state before storing the final state.

#### Scenario: Atomic create

- GIVEN both `write_index` and `read_indices` are configured
- WHEN create runs
- THEN a single Update Aliases call SHALL include one `add` action per configured index

### Requirement: Update (REQ-013–REQ-015)

On update, the resource SHALL compute a diff between current state and planned configuration. Indices present in state but absent from the plan SHALL be submitted as `remove` actions. Indices present in the plan but absent from state, or present in both but with changed settings, SHALL be submitted as `add` actions. Indices present in both with identical settings SHALL be skipped. All resulting actions SHALL be submitted in a single atomic Update Aliases API call. If no actions are required, the Update Aliases API SHALL NOT be called. After a successful update (or when no actions were needed), the resource SHALL perform a read to refresh state.

#### Scenario: Index removed from plan

- GIVEN an existing alias with two read indices
- WHEN one read index is removed from configuration
- THEN the update SHALL issue a `remove` action for the removed index and an `add` action (if changed) or no action for the unchanged index

### Requirement: Read (REQ-016–REQ-018)

On read, the resource SHALL call the Get Alias API with the alias name from state. If the API returns an empty result (no indices) or the alias name is not present in any returned index, the resource SHALL remove itself from state without an error. When alias data is returned, the resource SHALL classify each index as write or read based on the `is_write_index` flag from the API response and populate `write_index` and `read_indices` accordingly.

#### Scenario: Alias not found on read

- GIVEN the Get Alias API returns an empty map or the alias name is missing
- WHEN read runs
- THEN the resource SHALL be removed from state

#### Scenario: State classification

- GIVEN the API returns two indices, one with `is_write_index: true`
- WHEN read runs
- THEN `write_index` SHALL hold the write index and `read_indices` SHALL hold the remaining index

### Requirement: Delete (REQ-019)

On delete, the resource SHALL convert all indices from state (both `write_index` and `read_indices`) into `remove` alias actions and submit them in a single atomic Update Aliases API call. If no indices are present in state, the Update Aliases API SHALL NOT be called.

#### Scenario: Delete removes all indices

- GIVEN an alias with a write index and one read index
- WHEN delete runs
- THEN a single Update Aliases call SHALL include `remove` actions for both indices

### Requirement: Mapping — filter field (REQ-020–REQ-021)

`filter` in both `write_index` and `read_indices` entries SHALL be declared as a JSON-normalized string and validated as JSON by the schema type. On create and update, if `filter` is set, the resource SHALL unmarshal it into a map and pass it to the Update Aliases API. On read, if the API response contains a non-nil filter, the resource SHALL marshal it back to a JSON string and store it in state.

#### Scenario: Filter round-trip

- GIVEN a `filter` JSON value is configured
- WHEN create runs
- THEN the filter SHALL be sent as a map in the API payload and stored back as JSON on read

### Requirement: Mapping — routing and hidden fields (REQ-022)

On create and update, `index_routing`, `routing`, and `search_routing` SHALL be omitted from the API payload when their values are null or empty. `is_hidden` SHALL be included in the API payload only when its value is `true`. On read, string routing fields that are empty in the API response SHALL be stored as null (unknown/empty string values are not written to state).

#### Scenario: Empty routing omitted

- GIVEN `routing` is null in configuration
- WHEN the alias action is built
- THEN the `routing` key SHALL NOT appear in the API payload

### Requirement: Mapping — is_hidden default (REQ-023)

`is_hidden` in both `write_index` and `read_indices` SHALL default to `false` when not explicitly configured. On read, the API-returned boolean value for `is_hidden` SHALL always be written to state.

#### Scenario: Default is_hidden

- GIVEN `is_hidden` is not set in configuration
- WHEN plan is computed
- THEN `is_hidden` SHALL default to `false`
