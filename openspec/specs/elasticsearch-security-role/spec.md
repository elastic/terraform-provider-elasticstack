# `elasticstack_elasticsearch_security_role` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/security/role`
Data source implementation: `internal/elasticsearch/security/role_data_source.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_security_role` resource and data source, including Elasticsearch API usage, identity/import, connection handling, compatibility gates, state mapping, and state upgrades.

## Schema

### Resource schema

```hcl
resource "elasticstack_elasticsearch_security_role" "example" {
  id          = <computed, string> # <cluster_uuid>/<role_name>; UseStateForUnknown
  name        = <required, string>
  description = <optional, string> # requires Elasticsearch >= 8.15.0 when set

  # JSON (normalized) strings
  global   = <optional, json string>
  metadata = <optional, computed, json string>

  # Sets of strings
  cluster = <optional, set(string)>
  run_as  = <optional, set(string)>

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

  applications {
    application = <required, string>
    privileges  = <required, set(string)>
    resources   = <required, set(string)>
  }

  indices {
    names      = <required, set(string)>
    privileges = <required, set(string)>
    query                    = <optional, json string>
    allow_restricted_indices = <optional, computed, bool>
    field_security {
      grant  = <optional, set(string)>
      except = <optional, computed, set(string)>
    }
  }

  remote_indices {
    clusters   = <required, set(string)>
    names      = <required, set(string)>
    privileges = <required, set(string)>
    query = <optional, json string>
    field_security {
      grant  = <optional, set(string)>
      except = <optional, computed, set(string)>
    }
  }
}
```

### Data source schema

```hcl
data "elasticstack_elasticsearch_security_role" "example" {
  name   = <required, string>

  # Computed attributes (all read from Elasticsearch)
  id          = <computed, string>
  description = <computed, string>
  global      = <computed, string>  # JSON string
  metadata    = <computed, string>  # JSON string
  cluster     = <computed, set(string)>
  run_as      = <optional, set(string)>  # optional in SDK schema

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

  applications {
    application = <computed, string>
    privileges  = <computed, set(string)>
    resources   = <computed, set(string)>
  }

  indices {
    names                    = <computed, set(string)>
    privileges               = <computed, set(string)>
    query                    = <computed, string>
    allow_restricted_indices = <computed, bool>
    field_security {
      grant  = <computed, set(string)>
      except = <computed, set(string)>
    }
  }

  remote_indices {
    clusters   = <computed, set(string)>
    names      = <computed, set(string)>
    privileges = <computed, set(string)>
    query      = <computed, string>
    field_security {
      grant  = <computed, set(string)>
      except = <computed, set(string)>
    }
  }
}
```

## Requirements

### Requirement: Data source read API (DS-REQ-001)

The data source SHALL use the Elasticsearch Get roles API to read a role by name ([Get role API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role.html)).

#### Scenario: Successful data source read

- GIVEN a role exists in Elasticsearch
- WHEN the data source is read
- THEN the provider SHALL call the Get role API and populate all attributes in state

### Requirement: Data source API error surfacing (DS-REQ-002)

When Elasticsearch returns a non-success status for the read request (other than "not found"), the data source SHALL surface the API error to Terraform diagnostics.

#### Scenario: Non-success API response

- GIVEN the Elasticsearch API returns an error (not a 404) when reading a role
- WHEN the data source reads
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Data source identity (DS-REQ-003)

The data source SHALL expose a computed `id` representing a composite identifier in the format `<cluster_uuid>/<role_name>`, derived from the configured `name` and the target cluster UUID.

#### Scenario: Data source id after read

- GIVEN a successful data source read
- WHEN state is written
- THEN `id` SHALL equal `<cluster_uuid>/<role_name>` for the target cluster and configured name

### Requirement: Data source not found behavior (DS-REQ-004)

When a role is not found (the API returns no result), the data source SHALL remove itself from Terraform state (set `id` to empty string) rather than returning an error.

#### Scenario: Role does not exist

- GIVEN the role named in `name` does not exist in Elasticsearch
- WHEN the data source reads
- THEN `id` SHALL be set to empty string and no error diagnostic SHALL be returned

### Requirement: Data source connection (DS-REQ-005–DS-REQ-006)

The data source SHALL use the provider's configured Elasticsearch client by default. When the (deprecated) `elasticsearch_connection` block is configured on the data source, the data source SHALL use that connection to create an Elasticsearch client for all API calls of that instance.

#### Scenario: Data source with resource-scoped connection

- GIVEN `elasticsearch_connection` is set on the data source
- WHEN the data source reads
- THEN the client SHALL be built from that block

### Requirement: Data source attribute mapping (DS-REQ-007)

When reading a role, the data source SHALL map all role fields from the API response to state: `name`, `description` (when present), `applications`, `cluster`, `global` (serialized as JSON string when present), `indices` (including `field_security`), `remote_indices` (including `field_security`), `metadata` (serialized as JSON string when present), and `run_as`.

#### Scenario: Full role mapping

- GIVEN a role with all fields set in Elasticsearch
- WHEN the data source reads
- THEN all attributes SHALL be populated in state with correct values from the API response

### Requirement: Role CRUD APIs (REQ-001–REQ-003)

The resource SHALL use the Elasticsearch Create or update roles API to create and update roles ([Put role API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html)). The resource SHALL use the Elasticsearch Get roles API to read roles ([Get role API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role.html)). The resource SHALL use the Elasticsearch Delete roles API to delete roles ([Delete role API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-delete-role.html)).

#### Scenario: Lifecycle uses documented APIs

- GIVEN a role managed by this resource
- WHEN create, update, read, or delete runs
- THEN the provider SHALL use the Put, Get, and Delete role APIs as documented

### Requirement: API error surfacing (REQ-004)

When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than “not found” on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: Non-success API response

- GIVEN an Elasticsearch API error on create, update, read, or delete (except 404 on read)
- WHEN the provider handles the response
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` representing a composite identifier in the format `<cluster_uuid>/<role_name>`. When creating or updating a role, the resource SHALL compute `id` using the current cluster UUID and the configured role name.

#### Scenario: Computed id after apply

- GIVEN a successful create or update
- WHEN state is written
- THEN `id` SHALL equal `<cluster_uuid>/<role_name>` for the target cluster and configured name

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import by accepting an `id` in the format `<cluster_uuid>/<role_name>` and persisting it to state. If an imported or stored `id` is not in that format, the resource SHALL return an error diagnostic indicating the required format.

#### Scenario: Invalid import id

- GIVEN an import id not matching `<cluster_uuid>/<role_name>`
- WHEN import or read validates the id
- THEN the provider SHALL return an error diagnostic describing the required format

### Requirement: Name change lifecycle (REQ-009)

When the `name` argument changes, the resource SHALL require replacement (destroy/recreate), not an in-place update.

#### Scenario: Renaming a role

- GIVEN a configuration change to `name`
- WHEN Terraform plans the change
- THEN the resource SHALL be replaced

### Requirement: Elasticsearch connection (REQ-010–REQ-011)

The resource SHALL use the provider’s configured Elasticsearch client by default. When the (deprecated) `elasticsearch_connection` block is configured on the resource, the resource SHALL use that connection to create an Elasticsearch client for all API calls of that instance.

#### Scenario: Resource-scoped connection

- GIVEN `elasticsearch_connection` is set on the resource
- WHEN any API call runs for that instance
- THEN the client SHALL be built from that block

### Requirement: Version compatibility gates (REQ-012–REQ-013)

When `description` is configured (non-null), the resource SHALL verify the Elasticsearch server version is at least 8.15.0, and if it is lower the resource SHALL fail with an “Unsupported Feature” error. When `remote_indices` is configured with one or more entries, the resource SHALL verify the Elasticsearch server version is at least 8.10.0, and if it is lower the resource SHALL fail with an “Unsupported Feature” error.

#### Scenario: Description on older cluster

- GIVEN `description` is set and Elasticsearch is below 8.15.0
- WHEN create or update runs
- THEN the provider SHALL fail with an “Unsupported Feature” error

### Requirement: Create and update behavior (REQ-014–REQ-016)

When creating or updating a role, the resource SHALL submit the desired role definition using the Put role API and then read the role back to populate state. If the role cannot be read immediately after a successful create/update, the resource SHALL return an error indicating the role was not found after update.

#### Scenario: Post-apply read

- GIVEN a successful Put role response
- WHEN the provider refreshes state
- THEN it SHALL read the role and populate state, or error if the role is missing

### Requirement: Read and refresh (REQ-017–REQ-019)

When refreshing state, the resource SHALL parse `id` to determine the role name to read. If the role is not found (HTTP 404) during refresh, the resource SHALL remove itself from Terraform state. When a role is found, the resource SHALL set `name` in state to the role name extracted from `id` to ensure consistent addressing.

#### Scenario: Role removed in Elasticsearch

- GIVEN refresh runs and the role no longer exists
- WHEN the API returns 404
- THEN the resource SHALL be removed from state

### Requirement: Delete (REQ-020)

When destroying, the resource SHALL parse `id` to determine the role name and then delete it via the Delete role API.

#### Scenario: Destroy

- GIVEN destroy is requested
- WHEN delete runs
- THEN the provider SHALL call Delete role for the name parsed from `id`

### Requirement: JSON mapping for global and metadata (REQ-021–REQ-022)

When `global` is configured, the resource SHALL parse it as JSON; if parsing fails, the resource SHALL return an “Invalid JSON” error and SHALL not call the Put role API. When `metadata` is configured, the resource SHALL parse it as JSON; if parsing fails, the resource SHALL return an “Invalid JSON” error and SHALL not call the Put role API.

#### Scenario: Invalid global JSON

- GIVEN `global` contains invalid JSON
- WHEN create or update runs
- THEN the provider SHALL return “Invalid JSON” and SHALL NOT call Put role

### Requirement: Unknown values in plan (REQ-023–REQ-024)

When `indices.allow_restricted_indices` is unknown during planning, the resource SHALL preserve the prior state value for that field. When `indices.field_security.except` is unknown during planning, the resource SHALL preserve the prior state value for that field.

#### Scenario: Deferred unknowns

- GIVEN those attributes are unknown at plan time
- WHEN planning applies preservation rules
- THEN prior state values SHALL be kept for those fields

### Requirement: Role query mapping (REQ-025)

When `indices.query` or `remote_indices.query` is configured, the resource SHALL send it to Elasticsearch as the role query value (and omit it when unset/null).

#### Scenario: Query in API payload

- GIVEN query fields are set or unset
- WHEN the API payload is built
- THEN query values SHALL follow the documented mapping including omission when null

### Requirement: Empty cluster and run_as lists (REQ-026–REQ-027)

When Elasticsearch returns an empty `cluster` list, the resource SHALL preserve whether `cluster` was previously null vs explicitly empty, to avoid introducing drift. When Elasticsearch returns an empty `run_as` list, the resource SHALL preserve whether `run_as` was previously null vs explicitly empty, to avoid introducing drift.

#### Scenario: Empty list semantics

- GIVEN Elasticsearch returns empty lists for cluster or run_as
- WHEN state is updated from the API
- THEN null vs empty distinctions from prior state SHALL be preserved where specified

### Requirement: Empty nested blocks in state (REQ-028–REQ-029)

When Elasticsearch returns no `applications`, `indices`, or `remote_indices`, the resource SHALL represent those blocks as null (unset) in state. When Elasticsearch returns `global` or `metadata`, the resource SHALL serialize them back into normalized JSON strings in state; when absent, the resource SHALL set them to null in state.

#### Scenario: Absent optional blocks

- GIVEN the API omits nested structures or global/metadata
- WHEN read completes
- THEN state SHALL use null for absent blocks and normalized JSON strings when present

### Requirement: State upgrade v0 to v1 (REQ-030–REQ-034)

The resource SHALL support upgrading prior state schema version 0 to schema version 1. During v0→v1 upgrade, if `global` or `metadata` is null or an empty string, the resource SHALL remove the attribute from the upgraded state. During v0→v1 upgrade, for each `indices` or `remote_indices` entry, if `query` is an empty string, the resource SHALL remove `query` from that entry in the upgraded state. During v0→v1 upgrade, for each `indices` or `remote_indices` entry, if `field_security` is present as a legacy list, the resource SHALL convert it to a single object (first element), and if the list is empty it SHALL remove `field_security`. If the stored prior state JSON cannot be parsed during upgrade, the resource SHALL return a “State Upgrade Error” diagnostic and SHALL not produce upgraded state.

#### Scenario: Corrupt prior state

- GIVEN prior state JSON cannot be parsed
- WHEN upgrade runs
- THEN a “State Upgrade Error” diagnostic SHALL be returned

### Requirement: Null description preservation (REQ-035)

When Elasticsearch returns a null `description`, the resource SHALL preserve the null or empty (`""`) configured `description`, to avoid state consistency errors.

#### Scenario: API omits description

- GIVEN configuration uses null or empty description and the API returns null
- WHEN read refreshes state
- THEN configured null/empty description SHALL be preserved without spurious diff

### Requirement: Plan-time identity preservation (REQ-036)

When the computed resource `id` would otherwise be unknown during planning, the resource SHALL preserve the prior state value for `id` in the plan.

#### Scenario: Preserve prior id during plan

- GIVEN existing Terraform state contains a computed role `id`
- WHEN Terraform plans an operation where the computed `id` is otherwise unknown
- THEN the planned `id` value SHALL remain the prior state value
