# `elasticstack_elasticsearch_ml_calendar` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/calendar`

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_calendar" "example" {
  calendar_id = <required, string>  # RequiresReplace
  description = <optional, string>  # RequiresReplace
  job_ids     = <optional, set(string)>

  # Deprecated: resource-level Elasticsearch connection override
  elasticsearch_connection { ... }
}
```

## Requirements

### API

- **[REQ-001] (API)**: The resource shall use the Elasticsearch [Put calendar API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-put-calendar) (`PUT /_ml/calendars/{calendar_id}`) to create calendars. The `description` and `job_ids` fields from the Terraform config shall be included in the request body.
- **[REQ-002] (API)**: The resource shall use the Elasticsearch [Get calendars API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-get-calendars) (`GET /_ml/calendars/{calendar_id}`) to read calendar configuration.
- **[REQ-003] (API)**: The resource shall use the Elasticsearch [Delete calendar API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-delete-calendar) (`DELETE /_ml/calendars/{calendar_id}`) to delete calendars. The API automatically removes all associated scheduled events.
- **[REQ-004] (API)**: When Elasticsearch returns a non-success status for create, read, or delete requests (other than "not found" on read), the resource shall surface the API error to Terraform diagnostics.
- **[REQ-005] (API)**: When `job_ids` is updated (jobs added or removed), the resource shall diff the current vs. desired set and use the [Put calendar job API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-put-calendar-job) (`PUT /_ml/calendars/{calendar_id}/jobs/{job_id}`) and [Delete calendar job API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-delete-calendar-job) (`DELETE /_ml/calendars/{calendar_id}/jobs/{job_id}`) to add/remove individual jobs.

### Identity and Import

- **[REQ-006] (Identity)**: The resource shall expose a computed `id` representing a composite identifier in the format `<cluster_uuid>/<calendar_id>`.
- **[REQ-007] (Identity)**: When creating a calendar, the resource shall compute `id` using the current cluster UUID and the configured `calendar_id`.
- **[REQ-008] (Import)**: The resource shall support import by accepting an `id` in the format `<cluster_uuid>/<calendar_id>` and persisting it to state.
- **[REQ-009] (Import)**: If an imported or stored `id` is not in the expected composite format, the resource shall return an error diagnostic indicating the required format.

### Lifecycle

- **[REQ-010] (Lifecycle)**: When the `calendar_id` argument changes, the resource shall require replacement (destroy/recreate), not an in-place update.
- **[REQ-011] (Lifecycle)**: When `description` changes, the resource shall require replacement (destroy/recreate), as the Elasticsearch PUT calendar API is create-only and does not support updating an existing calendar's description.
- **[REQ-012] (Lifecycle)**: When `job_ids` changes, the resource shall update job associations in-place using the individual Put/Delete calendar job endpoints (REQ-005).

### Connection

- **[REQ-013] (Connection)**: The resource shall use the provider's configured Elasticsearch client by default.
- **[REQ-014] (Connection)**: When the (deprecated) `elasticsearch_connection` block is configured on the resource, the resource shall use that connection to create an Elasticsearch client for all API calls of that instance.

### Create

- **[REQ-015] (Create)**: When creating a calendar, the resource shall submit the calendar definition using the Put calendar API and then read the calendar back to populate state.
- **[REQ-016] (Create)**: If the calendar cannot be read immediately after a successful create, the resource shall return an error indicating the calendar was not found.

### Update

- **[REQ-017] (Update)**: When updating, the resource shall first read the current calendar state from the API to determine the current `job_ids`.
- **[REQ-018] (Update)**: The resource shall compute the set difference between current and desired `job_ids`, adding new jobs via Put calendar job and removing stale jobs via Delete calendar job.
- **[REQ-019] (Update)**: After all mutations, the resource shall read the calendar back to populate final state.

### Read

- **[REQ-021] (Read)**: When refreshing state, the resource shall parse `id` to determine the `calendar_id` to read.
- **[REQ-022] (Read)**: If the calendar is not found (HTTP 404 or empty response) during refresh, the resource shall remove itself from Terraform state (drift detection).
- **[REQ-023] (Read)**: When a calendar is found, the resource shall set `calendar_id`, `description`, and `job_ids` in state from the API response.

### Delete

- **[REQ-024] (Delete)**: When destroying, the resource shall parse `id` to determine the `calendar_id` and then delete it via the Delete calendar API. All associated events are removed automatically by the API.

### State and Mapping

- **[REQ-025] (State)**: When Elasticsearch returns an empty `job_ids` list and the user configured `job_ids` as null (unset), the resource shall store null in state (not an empty set) to avoid drift.
- **[REQ-026] (State)**: When Elasticsearch returns a null or empty `description`, the resource shall preserve null in state when the user did not configure a description.

### Validation

- **[REQ-027] (Validation)**: The `calendar_id` attribute shall be validated with a length constraint (1–64 characters) and a regex matching valid calendar identifier characters.

---

# `elasticstack_elasticsearch_ml_calendar_event` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/calendar_event`

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_calendar_event" "example" {
  calendar_id = <required, string>  # RequiresReplace
  description = <required, string>  # RequiresReplace
  start_time  = <required, rfc3339> # RequiresReplace
  end_time    = <required, rfc3339> # RequiresReplace

  # Computed
  event_id    = <computed, string>

  # Deprecated: resource-level Elasticsearch connection override
  elasticsearch_connection { ... }
}
```

## Requirements

### API

- **[REQ-101] (API)**: The resource shall use the Elasticsearch [Post calendar events API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-post-calendar-events) (`POST /_ml/calendars/{calendar_id}/events`) to create a single scheduled event. The request body shall contain an `events` array with one element.
- **[REQ-102] (API)**: The resource shall use the Elasticsearch [Get calendar events API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-get-calendar-events) (`GET /_ml/calendars/{calendar_id}/events`) to read events and locate the managed event by its `event_id`.
- **[REQ-103] (API)**: The resource shall use the Elasticsearch [Delete calendar event API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-delete-calendar-event) (`DELETE /_ml/calendars/{calendar_id}/events/{event_id}`) to delete events.
- **[REQ-104] (API)**: When Elasticsearch returns a non-success status for create, read, or delete requests (other than "not found" on read), the resource shall surface the API error to Terraform diagnostics.
- **[REQ-105] (API)**: There is no update API for calendar events. All mutable attributes shall use `RequiresReplace`, so any change triggers a destroy/recreate cycle.

### Identity and Import

- **[REQ-106] (Identity)**: The resource shall expose a computed `id` representing a composite identifier in the format `<cluster_uuid>/<calendar_id>/<event_id>`.
- **[REQ-107] (Identity)**: The resource shall expose a computed `event_id` attribute containing the server-generated event identifier returned by the Post calendar events API.
- **[REQ-108] (Import)**: The resource shall support import by accepting an `id` in the format `<cluster_uuid>/<calendar_id>/<event_id>` and persisting it to state.
- **[REQ-109] (Import)**: If an imported or stored `id` is not in the expected composite format, the resource shall return an error diagnostic indicating the required format.

### Lifecycle

- **[REQ-110] (Lifecycle)**: When any of `calendar_id`, `description`, `start_time`, or `end_time` changes, the resource shall require replacement (destroy + recreate). The `event_id` will change as a result.

### Connection

- **[REQ-111] (Connection)**: The resource shall use the provider's configured Elasticsearch client by default.
- **[REQ-112] (Connection)**: When the (deprecated) `elasticsearch_connection` block is configured on the resource, the resource shall use that connection to create an Elasticsearch client for all API calls of that instance.

### Create

- **[REQ-113] (Create)**: When creating an event, the resource shall submit a single-element events array to the Post calendar events API.
- **[REQ-114] (Create)**: The `start_time` and `end_time` RFC3339 values shall be converted to epoch milliseconds (or ISO 8601 strings) as required by the API.
- **[REQ-115] (Create)**: The resource shall extract the server-generated `event_id` from the API response and store it in state.
- **[REQ-116] (Create)**: After creation, the resource shall read the event back to populate state.

### Read

- **[REQ-117] (Read)**: When refreshing state, the resource shall parse `id` to extract `calendar_id` and `event_id`.
- **[REQ-118] (Read)**: The resource shall call the Get calendar events API and locate the event matching `event_id` in the response.
- **[REQ-119] (Read)**: If the event is not found (calendar gone, or event missing from the events list), the resource shall remove itself from Terraform state.
- **[REQ-120] (Read)**: When the event is found, the resource shall set `calendar_id`, `event_id`, `description`, `start_time`, and `end_time` in state from the API response.
- **[REQ-121] (Read)**: The API returns event times as epoch milliseconds. The resource shall convert these back to RFC3339 strings, preserving the timezone location from the user's original configuration where possible.

### Delete

- **[REQ-122] (Delete)**: When destroying, the resource shall parse `id` to extract `calendar_id` and `event_id`, then delete the event via the Delete calendar event API.
- **[REQ-123] (Delete)**: If the event or calendar is already gone (404), the delete shall succeed silently (idempotent).

### Validation

- **[REQ-124] (Validation)**: The `calendar_id` attribute shall be validated with a length constraint and valid character regex.
- **[REQ-125] (Validation)**: The `start_time` must be before `end_time`. This should be validated at plan time if both values are known.

---

## Sources

- **API client**: `github.com/elastic/go-elasticsearch/v8/esapi` — `MLPutCalendar`, `MLGetCalendars`, `MLDeleteCalendar`, `MLPutCalendarJob`, `MLDeleteCalendarJob`, `MLPostCalendarEvents`, `MLGetCalendarEvents`, `MLDeleteCalendarEvent`.
- **API docs**: [ML anomaly detection APIs](https://www.elastic.co/docs/api/doc/elasticsearch/group/endpoint-ml-anomaly) — calendar endpoints.
- **Existing patterns**: `internal/elasticsearch/ml/anomalydetectionjob/`, `internal/elasticsearch/ml/datafeed/`, `internal/elasticsearch/ml/datafeed_state/`.
