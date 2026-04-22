# `elasticstack_elasticsearch_security_system_user` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/security/systemuser`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_security_system_user` resource, which manages Elasticsearch built-in system users by controlling their password and enabled/disabled state. Because system users cannot be created or deleted, destroy only removes the resource from Terraform state without making any API call.

## Schema

```hcl
resource "elasticstack_elasticsearch_security_system_user" "example" {
  username      = <required, string>            # 1–1024 chars, printable ASCII graph chars only; UseStateForUnknown
  password      = <optional, sensitive, string> # 6–128 chars; conflicts with password_hash
  password_hash = <optional, sensitive, string> # 6–128 chars; conflicts with password
  enabled       = <optional+computed, bool>     # default true

  # Deprecated: resource-level Elasticsearch connection override
  elasticsearch_connection {
    endpoints    = <optional, list(string)>
    username     = <optional, string>
    password     = <optional, string>
    api_key      = <optional, string>
    bearer_token = <optional, string>
    es_client_authentication = <optional, string>
    insecure     = <optional, bool>
    ca_file      = <optional, string>
    ca_data      = <optional, string>
    cert_file    = <optional, string>
    cert_data    = <optional, string>
    key_file     = <optional, string>
    key_data     = <optional, string>
    headers      = <optional, map(string)>
  }
}
```

## Requirements

### Requirement: System user APIs (REQ-001–REQ-003)

The resource SHALL use the Elasticsearch Get user API to read system user state ([Get users API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-user.html)). The resource SHALL use the Elasticsearch Change passwords API to update a system user's password or password hash ([Change passwords API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-change-password.html)). The resource SHALL use the Elasticsearch Enable user and Disable user APIs to toggle the enabled state of a system user ([Enable user API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-enable-user.html), [Disable user API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-disable-user.html)).

#### Scenario: Lifecycle uses documented APIs

- GIVEN a system user managed by this resource
- WHEN create, update, or read runs
- THEN the provider SHALL use the Get user, Change passwords, Enable user, and Disable user APIs as applicable

### Requirement: API error surfacing (REQ-004)

When Elasticsearch returns a non-success status for any API call (Get user, Change passwords, Enable user, or Disable user), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: Non-success API response

- GIVEN an Elasticsearch API error during create or update
- WHEN the provider handles the response
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` representing a composite identifier in the format `<cluster_uuid>/<username>`. When creating or updating a system user, the resource SHALL compute `id` using the current cluster UUID and the configured `username`.

#### Scenario: Computed id after apply

- GIVEN a successful create or update
- WHEN state is written
- THEN `id` SHALL equal `<cluster_uuid>/<username>` for the target cluster and configured username

### Requirement: No import support (REQ-007)

The resource SHALL NOT implement `ImportState`; import is not supported for system users.

#### Scenario: No import interface

- GIVEN the resource implementation
- WHEN inspecting the resource type
- THEN it SHALL NOT implement `resource.ResourceWithImportState`

### Requirement: Username validation (REQ-008)

The resource SHALL require `username` to be between 1 and 1024 characters in length. The resource SHALL require `username` to contain only printable ASCII graph characters (no leading or trailing whitespace); if the value does not match, the resource SHALL return a validation error.

#### Scenario: Invalid username format

- GIVEN a `username` value containing spaces or non-printable characters
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error describing the allowed characters

### Requirement: Password and password_hash conflict (REQ-009)

The resource SHALL treat `password` and `password_hash` as mutually exclusive; if both are set, the resource SHALL return a validation error.

#### Scenario: Both password fields set

- GIVEN both `password` and `password_hash` are configured
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a conflict validation error

### Requirement: Elasticsearch connection (REQ-010–REQ-011)

The resource SHALL use the provider's configured Elasticsearch client by default. When the (deprecated) `elasticsearch_connection` block is configured on the resource, the resource SHALL use that connection to create an Elasticsearch client for all API calls of that instance.

#### Scenario: Resource-scoped connection

- GIVEN `elasticsearch_connection` is set on the resource
- WHEN any API call runs for that instance
- THEN the client SHALL be built from that block

### Requirement: Create and update behavior (REQ-012–REQ-016)

When creating or updating a system user, the resource SHALL first call the Get user API to verify the user exists and is a system user (i.e. its metadata contains `_reserved: true`). If the user is not found or is not a system user, the resource SHALL return an error diagnostic and SHALL NOT proceed with further API calls. When the configured `password` differs from the stored password, the resource SHALL call the Change passwords API with the new password. When the configured `password_hash` differs from the stored password hash, the resource SHALL call the Change passwords API with the new password hash. When the configured `enabled` value differs from the current enabled state, the resource SHALL call Enable user or Disable user as appropriate.

#### Scenario: Username not a system user

- GIVEN a configured `username` that does not correspond to a built-in system user
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic containing the username and SHALL NOT call the Change passwords or Enable/Disable user APIs

#### Scenario: Conditional password change

- GIVEN `password` is set and differs from the current stored password
- WHEN create or update runs
- THEN the provider SHALL call the Change passwords API with the new password

#### Scenario: Conditional enable/disable

- GIVEN `enabled` is set and differs from the current enabled state
- WHEN create or update runs
- THEN the provider SHALL call Enable user or Disable user as appropriate

### Requirement: Read and refresh (REQ-017–REQ-019)

When refreshing state, the resource SHALL parse `id` to extract the username. The resource SHALL call the Get user API for that username. If the user is not found or is not a system user, the resource SHALL remove itself from Terraform state and log a warning. When the user is found and is a system user, the resource SHALL set `username` and `enabled` in state from the API response, preserving the existing `password` and `password_hash` values (which are not returned by the API).

#### Scenario: System user removed or downgraded in Elasticsearch

- GIVEN refresh runs and the user no longer exists or is no longer a system user
- WHEN the API response indicates absence or non-system status
- THEN the resource SHALL be removed from state

#### Scenario: Enabled state synced on read

- GIVEN the system user exists and is enabled/disabled in Elasticsearch
- WHEN read refreshes state
- THEN `enabled` in state SHALL reflect the API response value

### Requirement: Delete is a no-op (REQ-020)

When destroying, the resource SHALL NOT call any Elasticsearch API to delete the user. The resource SHALL only remove itself from Terraform state and log a warning that system users are not deletable.

#### Scenario: Destroy

- GIVEN destroy is requested
- WHEN delete runs
- THEN no Elasticsearch API call SHALL be made and the resource SHALL be removed from state only

### Requirement: Password and password_hash mapping (REQ-021–REQ-022)

The resource SHALL only call the Change passwords API when `password` or `password_hash` is known and non-null and differs from the value currently stored for that user. The resource SHALL NOT send password or password_hash to the API when the values are unchanged.

#### Scenario: Unchanged password skips API call

- GIVEN `password` is set but matches the current stored value
- WHEN create or update runs
- THEN the provider SHALL NOT call the Change passwords API

### Requirement: Enabled default (REQ-023)

When `enabled` is not explicitly configured, the resource SHALL default to `true`.

#### Scenario: Default enabled value

- GIVEN `enabled` is not set in the configuration
- WHEN Terraform plans or applies
- THEN `enabled` SHALL be `true` in the plan and resulting state

### Requirement: Username state preservation (REQ-024)

When `username` is unknown during planning, the resource SHALL preserve the prior state value for that field (UseStateForUnknown plan modifier).

#### Scenario: Unknown username at plan time

- GIVEN `username` is unknown at plan time
- WHEN the plan is computed
- THEN the prior state value of `username` SHALL be used
