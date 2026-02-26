# `elasticstack_elasticsearch_security_role` — Schema and Functional Requirements (EARS)

Resource implementation: `internal/elasticsearch/security/role`

## Schema

```hcl
resource "elasticstack_elasticsearch_security_role" "example" {
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
    # When set, requires bearer_token
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
    # requires Elasticsearch >= 8.10.0 when non-empty
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

- **[REQ-001] (API)**: The resource shall use the Elasticsearch Create or update roles API to create and update roles ([Put role API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html)).
- **[REQ-002] (API)**: The resource shall use the Elasticsearch Get roles API to read roles ([Get role API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role.html)).
- **[REQ-003] (API)**: The resource shall use the Elasticsearch Delete roles API to delete roles ([Delete role API docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-delete-role.html)).
- **[REQ-004] (API)**: When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than “not found” on read), the resource shall surface the API error to Terraform diagnostics.
- **[REQ-005] (Identity)**: The resource shall expose a computed `id` representing a composite identifier in the format `<cluster_uuid>/<role_name>`.
- **[REQ-006] (Identity)**: When creating or updating a role, the resource shall compute `id` using the current cluster UUID and the configured role name.
- **[REQ-007] (Import)**: The resource shall support import by accepting an `id` in the format `<cluster_uuid>/<role_name>` and persisting it to state.
- **[REQ-008] (Import)**: If an imported or stored `id` is not in the format `<cluster_uuid>/<role_name>`, the resource shall return an error diagnostic indicating the required format.
- **[REQ-009] (Lifecycle)**: When the `name` argument changes, the resource shall require replacement (destroy/recreate), not an in-place update.
- **[REQ-010] (Connection)**: The resource shall use the provider’s configured Elasticsearch client by default.
- **[REQ-011] (Connection)**: When the (deprecated) `elasticsearch_connection` block is configured on the resource, the resource shall use that connection to create an Elasticsearch client for all API calls of that instance.
- **[REQ-012] (Compatibility)**: When `description` is configured (non-null), the resource shall verify the Elasticsearch server version is at least 8.15.0, and if it is lower the resource shall fail with an “Unsupported Feature” error.
- **[REQ-013] (Compatibility)**: When `remote_indices` is configured with one or more entries, the resource shall verify the Elasticsearch server version is at least 8.10.0, and if it is lower the resource shall fail with an “Unsupported Feature” error.
- **[REQ-014] (Create/Update)**: When creating a role, the resource shall submit the desired role definition using the Put role API and then read the role back to populate state.
- **[REQ-015] (Create/Update)**: When updating a role, the resource shall submit the desired role definition using the Put role API and then read the role back to populate state.
- **[REQ-016] (Create/Update)**: If the role cannot be read immediately after a successful create/update, the resource shall return an error indicating the role was not found after update.
- **[REQ-017] (Read)**: When refreshing state, the resource shall parse `id` to determine the role name to read.
- **[REQ-018] (Read)**: If the role is not found (HTTP 404) during refresh, the resource shall remove itself from Terraform state.
- **[REQ-019] (Read)**: When a role is found, the resource shall set `name` in state to the role name extracted from `id` to ensure consistent addressing.
- **[REQ-020] (Delete)**: When destroying, the resource shall parse `id` to determine the role name and then delete it via the Delete role API.
- **[REQ-021] (Mapping)**: When `global` is configured, the resource shall parse it as JSON; if parsing fails, the resource shall return an “Invalid JSON” error and shall not call the Put role API.
- **[REQ-022] (Mapping)**: When `metadata` is configured, the resource shall parse it as JSON; if parsing fails, the resource shall return an “Invalid JSON” error and shall not call the Put role API.
- **[REQ-023] (Plan/State)**: When `indices.allow_restricted_indices` is unknown during planning, the resource shall preserve the prior state value for that field.
- **[REQ-024] (Plan/State)**: When `indices.field_security.except` is unknown during planning, the resource shall preserve the prior state value for that field.
- **[REQ-025] (Mapping)**: When `indices.query` or `remote_indices.query` is configured, the resource shall send it to Elasticsearch as the role query value (and omit it when unset/null).
- **[REQ-026] (State)**: When Elasticsearch returns an empty `cluster` list, the resource shall preserve whether `cluster` was previously null vs explicitly empty, to avoid introducing drift.
- **[REQ-027] (State)**: When Elasticsearch returns an empty `run_as` list, the resource shall preserve whether `run_as` was previously null vs explicitly empty, to avoid introducing drift.
- **[REQ-028] (State)**: When Elasticsearch returns no `applications`, `indices`, or `remote_indices`, the resource shall represent those blocks as null (unset) in state.
- **[REQ-029] (State)**: When Elasticsearch returns `global` or `metadata`, the resource shall serialize them back into normalized JSON strings in state; when absent, the resource shall set them to null in state.
- **[REQ-030] (StateUpgrade)**: The resource shall support upgrading prior state schema version 0 to schema version 1.
- **[REQ-031] (StateUpgrade)**: During v0→v1 upgrade, if `global` or `metadata` is null or an empty string, the resource shall remove the attribute from the upgraded state.
- **[REQ-032] (StateUpgrade)**: During v0→v1 upgrade, for each `indices` or `remote_indices` entry, if `query` is an empty string, the resource shall remove `query` from that entry in the upgraded state.
- **[REQ-033] (StateUpgrade)**: During v0→v1 upgrade, for each `indices` or `remote_indices` entry, if `field_security` is present as a legacy list, the resource shall convert it to a single object (first element), and if the list is empty it shall remove `field_security`.
- **[REQ-034] (StateUpgrade)**: If the stored prior state JSON cannot be parsed during upgrade, the resource shall return a “State Upgrade Error” diagnostic and shall not produce upgraded state.
- **[REQ-035] (Plan/State)**: When Elasticsearch returns a null `description`, the resource shall preserve the null or empty (`""`) configured `description`, to avoid state consistency errors.  
