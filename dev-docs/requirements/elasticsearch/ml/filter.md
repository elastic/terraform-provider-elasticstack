# `elasticstack_elasticsearch_ml_filter` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/filter`

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_filter" "example" {
  filter_id   = <required, string>  # RequiresReplace
  description = <optional, string>
  items       = <optional, set(string)>

  # Deprecated: resource-level Elasticsearch connection override
  elasticsearch_connection { ... }
}
```

## Requirements

### API

- **[REQ-001] (API)**: The resource shall use the Elasticsearch [Put filter API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-put-filter) (`PUT /_ml/filters/{filter_id}`) to create filters. The `description` and `items` fields from the Terraform config shall be included in the request body.
- **[REQ-002] (API)**: The resource shall use the Elasticsearch [Get filters API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-get-filters) (`GET /_ml/filters/{filter_id}`) to read filter configuration.
- **[REQ-003] (API)**: The resource shall use the Elasticsearch [Update filter API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-update-filter) (`POST /_ml/filters/{filter_id}/_update`) to update filters. The request body shall include `description` (when changed), `add_items` (items to add), and `remove_items` (items to remove), computed by diffing the current and desired `items` sets.
- **[REQ-004] (API)**: The resource shall use the Elasticsearch [Delete filter API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-delete-filter) (`DELETE /_ml/filters/{filter_id}`) to delete filters.
- **[REQ-005] (API)**: When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than "not found" on read), the resource shall surface the API error to Terraform diagnostics.

### Identity and Import

- **[REQ-006] (Identity)**: The resource shall expose a computed `id` representing a composite identifier in the format `<cluster_uuid>/<filter_id>`.
- **[REQ-007] (Identity)**: When creating a filter, the resource shall compute `id` using the current cluster UUID and the configured `filter_id`.
- **[REQ-008] (Import)**: The resource shall support import by accepting an `id` in the format `<cluster_uuid>/<filter_id>` and persisting it to state.
- **[REQ-009] (Import)**: If an imported or stored `id` is not in the expected composite format, the resource shall return an error diagnostic indicating the required format.

### Lifecycle

- **[REQ-010] (Lifecycle)**: When the `filter_id` argument changes, the resource shall require replacement (destroy/recreate), not an in-place update.
- **[REQ-011] (Lifecycle)**: When `description` changes, the resource shall update the filter in-place via the Update filter API.
- **[REQ-012] (Lifecycle)**: When `items` changes, the resource shall compute the set difference (additions and removals) and update the filter in-place via the Update filter API using `add_items` and `remove_items`.

### Connection

- **[REQ-013] (Connection)**: The resource shall use the provider's configured Elasticsearch client by default.
- **[REQ-014] (Connection)**: When the (deprecated) `elasticsearch_connection` block is configured on the resource, the resource shall use that connection to create an Elasticsearch client for all API calls of that instance.

### Create

- **[REQ-015] (Create)**: When creating a filter, the resource shall submit the filter definition using the Put filter API and then read the filter back to populate state.
- **[REQ-016] (Create)**: If the filter cannot be read immediately after a successful create, the resource shall return an error indicating the filter was not found.

### Update

- **[REQ-017] (Update)**: When updating, the resource shall read the current filter from the API to determine the current `items`.
- **[REQ-018] (Update)**: The resource shall compute `add_items` (in desired but not current) and `remove_items` (in current but not desired) and include the desired `description` in the update request body.
- **[REQ-019] (Update)**: After the update, the resource shall read the filter back to populate final state.

### Read

- **[REQ-020] (Read)**: When refreshing state, the resource shall parse `id` to determine the `filter_id` to read.
- **[REQ-021] (Read)**: If the filter is not found (HTTP 404 or empty response) during refresh, the resource shall remove itself from Terraform state (drift detection).
- **[REQ-022] (Read)**: When a filter is found, the resource shall set `filter_id`, `description`, and `items` in state from the API response.

### Delete

- **[REQ-023] (Delete)**: When destroying, the resource shall parse `id` to determine the `filter_id` and then delete it via the Delete filter API.
- **[REQ-024] (Delete)**: If the filter is referenced by an anomaly detection job, the delete will fail with an API error surfaced to Terraform diagnostics.

### State and Mapping

- **[REQ-025] (State)**: When Elasticsearch returns an empty `items` list and the user configured `items` as null (unset), the resource shall store null in state (not an empty set) to avoid drift.
- **[REQ-026] (State)**: When Elasticsearch returns a null or empty `description`, the resource shall preserve null in state when the user did not configure a description.

### Validation

- **[REQ-027] (Validation)**: The `filter_id` attribute shall be validated with a length constraint (1–64 characters) and a regex matching valid filter identifier characters.

---

## Sources

- **API client**: `github.com/elastic/go-elasticsearch/v8/esapi` — `MLPutFilter`, `MLGetFilters`, `MLUpdateFilter`, `MLDeleteFilter`.
- **API docs**: [ML anomaly detection APIs](https://www.elastic.co/docs/api/doc/elasticsearch/group/endpoint-ml-anomaly) — filter endpoints.
- **Existing patterns**: `internal/elasticsearch/ml/calendar/`, `internal/elasticsearch/ml/anomaly_detection_job/`.
