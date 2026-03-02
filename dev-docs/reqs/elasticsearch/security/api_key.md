# `elasticstack_elasticsearch_security_api_key` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/security/api_key`

## Schema

```hcl
resource "elasticstack_elasticsearch_security_api_key" "example" {
  name = <required, string>

  # Immutable: requires replacement when changed
  type       = <optional, computed, string> # "rest" (default) or "cross_cluster"
  expiration = <optional, string>           # e.g. "1d"; by default keys never expire

  # REST API keys only (type="rest")
  # JSON (normalized) string; default values are populated for some fields
  role_descriptors = <optional, computed, json string>

  # Cross-cluster API keys only (type="cross_cluster")
  access {
    search = <optional, list(object({
      names                    = <required, list(string)>
      field_security           = <optional, json string>
      query                    = <optional, json string>
      allow_restricted_indices = <optional, bool>
    }))>

    replication = <optional, list(object({
      names = <required, list(string)>
    }))>
  }

  # JSON (normalized) strings
  metadata = <optional, computed, json string>

  # Computed identifiers and timestamps
  id                   = <computed, string> # <cluster_uuid>/<api_key_id>
  key_id               = <computed, string> # the Elasticsearch API key id
  expiration_timestamp = <computed, int64>  # expiration in milliseconds since epoch (0 when non-expiring)

  # Sensitive, computed, returned only at creation time
  api_key = <computed, sensitive, string>
  encoded = <computed, sensitive, string> # base64("<key_id>:<api_key>")

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
}
```

## Requirements

- **[REQ-001] (API)**: The resource shall use the Elasticsearch Create API key API to create REST API keys ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-api-key.html)).
- **[REQ-002] (API)**: The resource shall use the Elasticsearch Create cross-cluster API key API to create cross-cluster API keys ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-cross-cluster-api-key.html)).
- **[REQ-003] (API)**: The resource shall use the Elasticsearch Get API key API to read API keys ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-api-key.html)).
- **[REQ-004] (API)**: The resource shall use the Elasticsearch Update API key API to update REST API keys ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-update-api-key.html)).
- **[REQ-005] (API)**: The resource shall use the Elasticsearch Update cross-cluster API key API to update cross-cluster API keys ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-update-cross-cluster-api-key.html)).
- **[REQ-006] (API)**: The resource shall use the Elasticsearch Invalidate API key API to delete (invalidate) API keys ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-invalidate-api-key.html)).
- **[REQ-007] (API)**: When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than “not found” on read), the resource shall surface the API error to Terraform diagnostics.
- **[REQ-008] (Identity)**: The resource shall expose a computed `id` representing a composite identifier in the format `<cluster_uuid>/<api_key_id>`.
- **[REQ-009] (Identity)**: The resource shall expose a computed `key_id` representing the Elasticsearch API key identifier.
- **[REQ-010] (Identity)**: When creating an API key, the resource shall compute `id` using the current cluster UUID and the API key id returned by Elasticsearch.
- **[REQ-011] (Read)**: When refreshing state, the resource shall parse `id` to determine the API key id to read.
- **[REQ-012] (Read)**: If the API key is not found (HTTP 404) during refresh, the resource shall remove itself from Terraform state.
- **[REQ-013] (Import)**: The resource shall not support Terraform import.
- **[REQ-014] (Lifecycle)**: When the `name` argument changes, the resource shall require replacement (destroy/recreate), not an in-place update.
- **[REQ-015] (Lifecycle)**: When the `type` argument changes, the resource shall require replacement (destroy/recreate), not an in-place update.
- **[REQ-016] (Lifecycle)**: When the `expiration` argument changes, the resource shall require replacement (destroy/recreate), not an in-place update.
- **[REQ-017] (Connection)**: The resource shall use the provider’s configured Elasticsearch client by default.
- **[REQ-018] (Connection)**: When the (deprecated) `elasticsearch_connection` block is configured on the resource, the resource shall use that connection to create an Elasticsearch client for all API calls of that instance.
- **[REQ-019] (Compatibility)**: The resource shall require Elasticsearch >= 8.0.0 (API keys are enabled in Elasticsearch 8.0).
- **[REQ-020] (Compatibility)**: When `type="cross_cluster"`, the resource shall verify the Elasticsearch server version is at least 8.10.0; if it is lower, the resource shall fail with an error indicating cross-cluster API keys are not supported.
- **[REQ-021] (Compatibility)**: The resource shall treat Elasticsearch >= 8.4.0 as the minimum version that supports API key update; when the last-refreshed server version is below 8.4.0, changes to `metadata` and/or `role_descriptors` shall require replacement.
- **[REQ-022] (Compatibility)**: The resource shall treat Elasticsearch >= 8.5.0 as the minimum version that returns `role_descriptors` from the Get API key API.
- **[REQ-023] (Compatibility)**: When `role_descriptors` contains a `restriction` field for any role descriptor, the resource shall verify the Elasticsearch server version is at least 8.9.0; if it is lower, the resource shall fail with an error indicating restrictions on API key role descriptors are unsupported.
- **[REQ-024] (Validation)**: The resource shall validate that `type` is one of `rest` or `cross_cluster`, defaulting to `rest` when unset.
- **[REQ-025] (Validation)**: The resource shall validate that `role_descriptors` is only set when `type="rest"`.
- **[REQ-026] (Validation)**: The resource shall validate that `access` is only set when `type="cross_cluster"`.
- **[REQ-027] (Validation)**: The resource shall validate that `name` is between 1 and 1024 characters and only contains printable ASCII characters (including spaces), as defined by the resource schema.
- **[REQ-028] (Mapping)**: When `metadata` is configured, the resource shall parse it as JSON; if parsing fails, the resource shall return an “Invalid JSON” error and shall not call Elasticsearch.
- **[REQ-029] (Mapping)**: When `role_descriptors` is configured, the resource shall parse it as JSON; if parsing fails, the resource shall return an “Invalid JSON” error and shall not call Elasticsearch.
- **[REQ-030] (Defaults)**: When `role_descriptors` is configured, the resource shall populate defaults for missing nested values as required by the provider’s role-descriptor defaulting logic (e.g. default `indices[].allow_restricted_indices` to `false` when omitted).
- **[REQ-031] (Create)**: When creating a REST API key (`type="rest"`), the resource shall send `name`, `expiration` (if set), `metadata` (if set), and `role_descriptors` (if set) to Elasticsearch using the Create API key API.
- **[REQ-032] (Create)**: When creating a cross-cluster API key (`type="cross_cluster"`), the resource shall send `name`, `expiration` (if set), `metadata` (if set), and `access` (when non-empty) to Elasticsearch using the Create cross-cluster API key API.
- **[REQ-033] (Create)**: After a successful create, the resource shall persist to state:
  - **[REQ-033a]**: `id` as `<cluster_uuid>/<api_key_id>`
  - **[REQ-033b]**: `key_id` as the Elasticsearch API key id
  - **[REQ-033c]**: `api_key` and `encoded` as returned by Elasticsearch (sensitive)
- **[REQ-034] (Create)**: After a successful create, the resource shall read the created key via the Get API key API and persist returned non-sensitive fields (e.g. `expiration_timestamp`, `metadata`, and when available `role_descriptors`) to state.
- **[REQ-035] (Create)**: If Elasticsearch returns a nil/empty create response body on success, the resource shall return an error diagnostic indicating the creation response was invalid.
- **[REQ-036] (Read)**: On read, the resource shall populate `key_id`, `name`, `expiration_timestamp`, and `metadata` from the Get API key API response.
- **[REQ-037] (Read)**: For Elasticsearch >= 8.5.0, on read the resource shall populate `role_descriptors` from the Get API key API response when present; when absent, it shall set `role_descriptors` to null.
- **[REQ-038] (Read/State)**: For Elasticsearch < 8.5.0, the resource shall preserve the prior state value of `role_descriptors` (because it is not reliably returned by the API).
- **[REQ-039] (Read/State)**: The resource shall not expect `api_key` or `encoded` to be returned by the Get API key API and shall preserve the prior state values for those attributes.
- **[REQ-040] (Read/Plan)**: After a successful read, the resource shall store the Elasticsearch cluster version in private state for use by plan-time modifiers (e.g. conditional replacement when update is unsupported).
- **[REQ-041] (Update)**: When updating a REST API key (`type="rest"`), the resource shall call the Update API key API and only attempt to change mutable fields (e.g. `metadata` and `role_descriptors`), never `name` or `expiration`.
- **[REQ-042] (Update)**: When updating a cross-cluster API key (`type="cross_cluster"`), the resource shall call the Update cross-cluster API key API and only attempt to change mutable fields (e.g. `metadata` and `access`), never `name` or `expiration`.
- **[REQ-043] (Update)**: After a successful update, the resource shall read the key via the Get API key API and persist returned non-sensitive fields to state.
- **[REQ-044] (Plan/State)**: When managing a cross-cluster API key, if the configured `access` value differs from the prior state `access` value, the resource shall set `role_descriptors` to unknown during planning to avoid incorrect diffs between mutually exclusive access models.
- **[REQ-045] (Delete)**: When destroying, the resource shall invalidate the API key id parsed from `id` via the Invalidate API key API, and then remove itself from Terraform state.
- **[REQ-046] (StateUpgrade)**: The resource shall support upgrading prior state schema version 0 to schema version 1 by converting `expiration=""` to `expiration=null`.
- **[REQ-047] (StateUpgrade)**: The resource shall support upgrading prior state schema version 1 to schema version 2 by setting `type` to the default value `rest`. 