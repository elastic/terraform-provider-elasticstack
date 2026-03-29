# `elasticstack_elasticsearch_security_role_mapping` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/security/rolemapping`
Data source implementation: `internal/elasticsearch/security/rolemapping`

## Purpose

Define schema and behavior for the Elasticsearch security role mapping resource and data source: API usage, identity/import, lifecycle, connection, and JSON mapping semantics for rules, roles, role templates, and metadata.

## Schema

### Resource

```hcl
resource "elasticstack_elasticsearch_security_role_mapping" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<name>
  name = <required, string> # force new

  enabled        = <optional, computed, bool>           # default true
  rules          = <required, JSON (normalized) string> # JSON DSL rule expression
  roles          = <optional, set(string)>              # exactly one of: roles or role_templates
  role_templates = <optional, JSON (normalized) string> # exactly one of: roles or role_templates
  metadata       = <optional, computed, JSON (normalized) string> # default "{}"

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

### Data source

```hcl
data "elasticstack_elasticsearch_security_role_mapping" "example" {
  id   = <computed, string>
  name = <required, string>

  enabled        = <computed, bool>
  rules          = <computed, JSON (normalized) string>
  roles          = <computed, set(string)>
  role_templates = <computed, JSON (normalized) string>
  metadata       = <computed, JSON (normalized) string>

  elasticsearch_connection {
    # same attributes as resource
  }
}
```

## Requirements

### Requirement: Role mapping CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put Role Mapping API to create and update role mappings ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html)). The resource SHALL use the Elasticsearch Get Role Mapping API to read role mappings ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role-mapping.html)). The resource SHALL use the Elasticsearch Delete Role Mapping API to delete role mappings ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-delete-role-mapping.html)).

When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than not found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API errors surfaced

- GIVEN a failing Elasticsearch response (other than 404 on read)
- WHEN the provider processes the response
- THEN diagnostics SHALL include the API error

### Requirement: Data source read API (REQ-005)

The data source SHALL use the Elasticsearch Get Role Mapping API to read role mapping data ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role-mapping.html)). When Elasticsearch returns a non-success status (other than not found), the data source SHALL surface the API error to Terraform diagnostics. When the named role mapping is not found, the data source SHALL return an error diagnostic.

#### Scenario: Data source role mapping not found

- GIVEN a `name` that does not correspond to any role mapping in Elasticsearch
- WHEN the data source reads
- THEN diagnostics SHALL include a "Role mapping not found" error

### Requirement: Identity (REQ-006–REQ-007)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<name>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and the configured `name` by performing a read-after-write. The data source SHALL expose a computed `id` in the format `<cluster_uuid>/<name>`, computed from the current cluster UUID and the required `name` attribute.

#### Scenario: Id set on create

- GIVEN a valid role mapping configuration
- WHEN create completes
- THEN `id` SHALL be set in state in the format `<cluster_uuid>/<name>`

### Requirement: Import (REQ-008)

The resource SHALL support import via `ImportStatePassthroughID`, persisting the imported `id` value directly to state. Read and delete operations SHALL parse `id` using `CompositeIDFromStrFw`; when the format is invalid, they SHALL return an error diagnostic.

#### Scenario: Invalid id on read

- GIVEN a malformed `id` in state
- WHEN read or delete runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Lifecycle (REQ-009)

Changing `name` SHALL require replacement of the resource (`RequiresReplace`).

#### Scenario: Name change forces replacement

- GIVEN `name` changes in configuration
- WHEN Terraform plans
- THEN replacement SHALL be required

### Requirement: Connection (REQ-010)

By default, the resource and data source SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource and data source SHALL construct and use a resource-scoped Elasticsearch client for all API calls.

#### Scenario: Resource-level connection override

- GIVEN `elasticsearch_connection` is set on the resource
- WHEN create, read, update, or delete runs
- THEN all API calls SHALL use the resource-scoped client

### Requirement: Create and update (REQ-011–REQ-013)

On create and update, the resource SHALL parse `rules` as JSON and if parsing fails, SHALL return an error diagnostic and not call the Put API. On create and update, when `role_templates` is set and known, the resource SHALL parse it as a JSON array and if parsing fails, SHALL return an error diagnostic. After a successful Put, the resource SHALL perform a read-after-write to refresh state.

#### Scenario: Invalid rules JSON

- GIVEN an invalid JSON string in `rules`
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic without calling Put

#### Scenario: Invalid role_templates JSON

- GIVEN an invalid JSON array in `role_templates`
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic without calling Put

### Requirement: Read and state removal (REQ-014)

On read, the resource SHALL parse `id` to extract the role mapping name, fetch the role mapping from Elasticsearch, and remove the resource from state when the role mapping is not found (HTTP 404).

#### Scenario: Role mapping not found on refresh

- GIVEN the role mapping was deleted in Elasticsearch outside Terraform
- WHEN read runs
- THEN the resource SHALL be removed from state

### Requirement: Delete (REQ-015)

On delete, the resource SHALL parse `id` to extract the role mapping name and call the Delete Role Mapping API with that name.

#### Scenario: Destroy deletes by parsed name

- GIVEN destroy
- WHEN delete runs
- THEN Delete Role Mapping SHALL be called for the name parsed from `id`

### Requirement: Roles and role_templates mutual exclusivity (REQ-016)

The resource schema SHALL enforce that exactly one of `roles` or `role_templates` is configured; configuring neither or both SHALL produce a validation error.

#### Scenario: Both roles and role_templates set

- GIVEN both `roles` and `role_templates` are configured
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

#### Scenario: Neither roles nor role_templates set

- GIVEN neither `roles` nor `role_templates` is configured
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

### Requirement: Enabled default (REQ-017)

When `enabled` is not set in configuration, the resource SHALL default to `true`.

#### Scenario: Enabled defaults to true

- GIVEN `enabled` is omitted from configuration
- WHEN the resource is created
- THEN `enabled` SHALL be `true` in state and sent as `true` in the API request

### Requirement: Metadata default (REQ-018)

When `metadata` is not set in configuration, the resource SHALL default to `"{}"` (an empty JSON object).

#### Scenario: Metadata defaults to empty object

- GIVEN `metadata` is omitted from configuration
- WHEN the resource is created
- THEN `metadata` SHALL be `"{}"` in state

### Requirement: JSON state mapping (REQ-019–REQ-021)

On read, the resource and data source SHALL serialize `rules` from the API response into a normalized JSON string and store it in state. On read, when `role_templates` is present and non-empty in the API response, the resource and data source SHALL serialize it as a normalized JSON array string and store it in state.

When `role_templates` is absent or empty in the API response, it SHALL be stored as null in state. On read, the resource and data source SHALL serialize `metadata` from the API response into a normalized JSON string and store it in state.

#### Scenario: Empty role_templates stored as null

- GIVEN the API response has no role_templates
- WHEN read maps the response to state
- THEN `role_templates` in state SHALL be null

### Requirement: Roles state mapping (REQ-022)

On read, the resource and data source SHALL map the `roles` list from the API response to a set of strings in state.

#### Scenario: Roles stored as set

- GIVEN an API response containing a list of role names
- WHEN read maps the response to state
- THEN `roles` in state SHALL contain the same role names as a set of strings
