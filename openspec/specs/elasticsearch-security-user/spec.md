# `elasticstack_elasticsearch_security_user` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/security/user`
Data source implementation: `internal/elasticsearch/security/user_data_source.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_security_user` resource and data source. The resource creates and manages Elasticsearch security users via the Put user API, supporting password, password hash, and write-only password authentication. The data source reads an existing Elasticsearch security user by username and exposes its attributes as computed values.

## Schema

### Resource

```hcl
resource "elasticstack_elasticsearch_security_user" "example" {
  username           = <required, string>  # 1–1024 chars, printable ASCII, RequiresReplace
  password           = <optional, sensitive, string>  # 6–128 chars; conflicts with password_hash, password_wo
  password_hash      = <optional, sensitive, string>  # 6–128 chars; conflicts with password, password_wo
  password_wo        = <optional, sensitive, write-only, string>  # 6–128 chars; conflicts with password, password_hash
  password_wo_version = <optional, string>  # requires password_wo

  full_name = <optional+computed, string>  # default ""
  email     = <optional+computed, string>  # default ""
  roles     = <required, set(string)>      # at least 1 element
  metadata  = <optional+computed, json string>
  enabled   = <optional+computed, bool>    # default true

  # Computed
  id = <computed, string>  # <cluster_uuid>/<username>

  # Deprecated: resource-level Elasticsearch connection override
  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    cert_data                = <optional, string>
    key_file                 = <optional, string>
    key_data                 = <optional, string>
    headers                  = <optional, map(string)>
  }
}
```

### Data source

```hcl
data "elasticstack_elasticsearch_security_user" "example" {
  username = <required, string>

  # Computed outputs
  id        = <computed, string>  # <cluster_uuid>/<username>
  full_name = <computed, string>
  email     = <computed, string>
  roles     = <computed, set(string)>
  metadata  = <computed, string>  # JSON-serialized metadata object
  enabled   = <computed, bool>

  # Deprecated: resource-level Elasticsearch connection override
  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    cert_data                = <optional, string>
    key_file                 = <optional, string>
    key_data                 = <optional, string>
    headers                  = <optional, map(string)>
  }
}
```

## Requirements

### Requirement: User CRUD APIs (REQ-001–REQ-003)

The resource SHALL use the Elasticsearch Put user API to create and update users. The resource SHALL use the Elasticsearch Get users API to read the user identified by username. The resource SHALL use the Elasticsearch Delete user API to delete users.

#### Scenario: Lifecycle uses documented APIs

- GIVEN a user managed by this resource
- WHEN create, update, read, or delete runs
- THEN the provider SHALL call the Put user, Get users, or Delete user API as appropriate

### Requirement: API error surfacing (REQ-004)

When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than "not found" on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: Non-success API response

- GIVEN an Elasticsearch API error on create, update, read, or delete (except user not found on read)
- WHEN the provider handles the response
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-005)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<username>`. When creating or updating a user, the resource SHALL compute `id` using the current cluster UUID and the configured `username`.

#### Scenario: Computed id after apply

- GIVEN a successful create or update
- WHEN state is written
- THEN `id` SHALL equal `<cluster_uuid>/<username>`

### Requirement: Import (REQ-006)

The resource SHALL support import by accepting an `id` in the format `<cluster_uuid>/<username>` and persisting it to state via `ImportStatePassthroughID`.

#### Scenario: Import by composite id

- GIVEN an import is requested with `<cluster_uuid>/<username>`
- WHEN import runs
- THEN the id SHALL be persisted to state and a subsequent read SHALL populate all other attributes

### Requirement: Username lifecycle (REQ-007)

When the `username` argument changes, the resource SHALL require replacement (destroy/recreate), not an in-place update.

#### Scenario: Renaming a user

- GIVEN a configuration change to `username`
- WHEN Terraform plans the change
- THEN the resource SHALL be replaced

### Requirement: Connection (REQ-008–REQ-009)

The resource SHALL use the provider's configured Elasticsearch client by default. When the (deprecated) `elasticsearch_connection` block is configured on the resource, the resource SHALL use that connection to create an Elasticsearch client for all API calls of that instance.

#### Scenario: Resource-scoped connection

- GIVEN `elasticsearch_connection` is set on the resource
- WHEN any API call runs for that instance
- THEN the client SHALL be built from that block

### Requirement: Create and update behavior (REQ-010–REQ-012)

When creating or updating a user, the resource SHALL call the Put user API with the desired user definition and then read the user back to populate computed fields in state. If the user cannot be read immediately after a successful create/update, the resource SHALL return an error diagnostic indicating the user was not found after the operation. The resource SHALL only send a password field when it is known and has changed from the prior state (for `password` and `password_hash`) or when `password_wo_version` has changed (for `password_wo`).

#### Scenario: Post-apply read

- GIVEN a successful Put user response
- WHEN the provider refreshes state
- THEN it SHALL read the user and populate state, or return an error if the user is missing

#### Scenario: Password change detection

- GIVEN `password` has not changed between plan and prior state
- WHEN update runs
- THEN the provider SHALL NOT include the password in the Put user request

### Requirement: Read and refresh (REQ-013–REQ-014)

When refreshing state, the resource SHALL parse `id` to extract the username and call the Get users API. If the user is not found during refresh, the resource SHALL remove itself from Terraform state. When a user is found, the resource SHALL update `username`, `email`, `full_name`, `roles`, `enabled`, and `metadata` from the API response.

#### Scenario: User removed in Elasticsearch

- GIVEN refresh runs and the user no longer exists
- WHEN the API returns nil without an error
- THEN the resource SHALL be removed from state

### Requirement: Delete (REQ-015)

When destroying, the resource SHALL parse `id` to extract the username and then delete the user via the Delete user API.

#### Scenario: Destroy

- GIVEN destroy is requested
- WHEN delete runs
- THEN the provider SHALL call Delete user for the username parsed from `id`

### Requirement: Metadata JSON mapping (REQ-016–REQ-017)

When `metadata` is configured (non-null, non-unknown), the resource SHALL parse it as JSON before sending it to the Put user API; if parsing fails, the resource SHALL return an error diagnostic and SHALL NOT call the API. When reading state, if the API response includes a non-empty metadata map the resource SHALL marshal it to a normalized JSON string and store it in state; if the metadata map is empty or absent the resource SHALL store null in state.

#### Scenario: Invalid metadata JSON on apply

- GIVEN `metadata` contains invalid JSON
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic and SHALL NOT call Put user

#### Scenario: Empty metadata from API

- GIVEN the API returns an empty or absent metadata map
- WHEN read populates state
- THEN `metadata` SHALL be null in state

### Requirement: Password conflict validation (REQ-018)

The resource SHALL enforce that at most one of `password`, `password_hash`, or `password_wo` is set at a time, using attribute-level conflict validators. If more than one is set, the resource SHALL return a validation error before any API call.

#### Scenario: Conflicting password attributes

- GIVEN both `password` and `password_wo` are set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return a conflict validation error

---

## Data source requirements

### Requirement: Read API (REQ-DS-001)

The data source SHALL use the Elasticsearch Get users API to fetch the user identified by `username`. When the API returns a non-success response, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: API failure

- GIVEN the Elasticsearch Get users API returns an error
- WHEN read runs
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-DS-002)

The data source SHALL set `id` to a composite value in the format `<cluster_uuid>/<username>` by resolving the cluster UUID from the connected Elasticsearch cluster and combining it with the configured `username`.

#### Scenario: Computed id

- GIVEN a successful read of an existing user
- WHEN read completes
- THEN `id` SHALL equal `<cluster_uuid>/<username>`

### Requirement: User not found (REQ-DS-003)

When the Elasticsearch Get users API returns nil without an error (user not found), the data source SHALL remove itself from state by calling `d.SetId("")` and return without error.

#### Scenario: User does not exist

- GIVEN the specified username does not exist in Elasticsearch
- WHEN read runs
- THEN the data source SHALL clear its id and return no error

### Requirement: State mapping (REQ-DS-004–REQ-DS-005)

When the user is found, the data source SHALL populate `username`, `email`, `full_name`, `roles`, `enabled`, and `metadata` from the API response. The `metadata` attribute SHALL be a JSON-serialized string produced by marshaling the user's metadata object; if marshaling fails, the data source SHALL surface an error to Terraform diagnostics.

#### Scenario: Metadata marshaling error

- GIVEN the user's metadata cannot be marshaled to JSON
- WHEN read populates state
- THEN the data source SHALL return an error diagnostic

### Requirement: Connection (REQ-DS-006–REQ-DS-007)

The data source SHALL use the provider's configured Elasticsearch client by default. When the (deprecated) `elasticsearch_connection` block is configured, the data source SHALL use that connection to construct an Elasticsearch client for its API call.

#### Scenario: Resource-scoped connection

- GIVEN `elasticsearch_connection` is set
- WHEN the API call runs
- THEN the client SHALL be built from that block
