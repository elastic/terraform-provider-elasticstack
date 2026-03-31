# `elasticstack_elasticsearch_logstash_pipeline` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/logstash/pipeline.go`

## Purpose

Define schema and behavior for the Elasticsearch Logstash pipeline resource: API usage, identity/import, connection, lifecycle, settings mapping, and read-time state management for centralized Logstash pipeline configuration via the Elasticsearch Logstash Pipelines API.

## Schema

```hcl
resource "elasticstack_elasticsearch_logstash_pipeline" "example" {
  id          = <computed, string>  # internal identifier: <cluster_uuid>/<pipeline_id>
  pipeline_id = <required, string>  # force new; identifier for the pipeline

  description = <optional, string>  # description of the pipeline
  pipeline    = <required, string>  # Logstash pipeline configuration DSL

  pipeline_metadata = <optional, string>  # JSON string; default: {"type":"logstash_pipeline","version":1}; validated as JSON; diff suppressed on semantic equivalence

  last_modified = <computed, string>  # date the pipeline was last updated

  username = <optional, string>  # user who last updated; defaults to ELASTICSEARCH_USERNAME env var or "api_key"

  # Pipeline settings (all optional; sent only when set)
  pipeline_batch_delay         = <optional, int>     # ms to wait before sending undersized batch
  pipeline_batch_size          = <optional, int>     # max events per worker before executing filters/outputs
  pipeline_ecs_compatibility   = <optional, string>  # one of: "disabled", "v1", "v8"
  pipeline_ordered             = <optional, string>  # one of: "auto", "true", "false"
  pipeline_plugin_classloaders = <optional, bool>    # (beta) isolate Java plugin dependencies
  pipeline_unsafe_shutdown     = <optional, bool>    # force exit on shutdown with inflight events
  pipeline_workers             = <optional, int>     # parallel workers for filter/output; minimum 1

  queue_checkpoint_acks   = <optional, int>     # max ACKed events before checkpoint (persistent queues)
  queue_checkpoint_retry  = <optional, bool>    # retry failed checkpoint writes up to four times
  queue_checkpoint_writes = <optional, int>     # max written events before checkpoint (persistent queues)
  queue_drain             = <optional, bool>    # wait for queue drain before shutdown
  queue_max_bytes         = <optional, string>  # total queue capacity; must match /^[0-9]+[kmgtp]?b$/
  queue_max_events        = <optional, int>     # max unread events in queue (persistent queues)
  queue_page_capacity     = <optional, string>  # size of each queue page data file
  queue_type              = <optional, string>  # one of: "memory", "persisted"

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

### Requirement: Logstash pipeline CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Logstash Put Pipeline API (`LogstashPutPipeline`) to create and update Logstash pipelines ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/logstash-api-put-pipeline.html)). The resource SHALL use the Elasticsearch Logstash Get Pipeline API (`LogstashGetPipeline`) to read Logstash pipelines ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/logstash-api-get-pipeline.html)). The resource SHALL use the Elasticsearch Logstash Delete Pipeline API (`LogstashDeletePipeline`) to delete Logstash pipelines ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/logstash-api-delete-pipeline.html)). When any API returns a non-success status (other than HTTP 404 on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create or update

- GIVEN the Logstash Put Pipeline API returns a non-success response
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

#### Scenario: API failure on delete

- GIVEN the Logstash Delete Pipeline API returns a non-success response
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<pipeline_id>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and the configured `pipeline_id`, and store the computed value in state.

#### Scenario: Identity set on create

- GIVEN a successful create
- WHEN the resource sets its id
- THEN the state `id` SHALL be in the format `<cluster_uuid>/<pipeline_id>`

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state without transformation. For read and delete operations, the resource SHALL require `id` to match the format `<cluster_uuid>/<resource identifier>` and SHALL return an error diagnostic ("Wrong resource ID") when the format is invalid.

#### Scenario: Import passthrough

- GIVEN import with a valid composite id `<cluster_uuid>/<pipeline_id>`
- WHEN import completes
- THEN the id SHALL be stored in state for subsequent read and delete operations

#### Scenario: Invalid id format

- GIVEN an `id` that does not contain exactly one `/` separator
- WHEN read or delete parses the id
- THEN the resource SHALL return an error diagnostic with summary "Wrong resource ID."

### Requirement: Lifecycle (REQ-009)

Changing `pipeline_id` SHALL require resource replacement (`ForceNew`). All other attributes may be updated in place.

#### Scenario: pipeline_id change triggers replacement

- GIVEN a resource exists with a given `pipeline_id`
- WHEN `pipeline_id` is changed in configuration
- THEN Terraform SHALL plan to destroy and recreate the resource

### Requirement: Connection (REQ-010–REQ-011)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (create, update, read, delete).

#### Scenario: Resource-level connection override

- GIVEN `elasticsearch_connection` is set with custom endpoint credentials
- WHEN any CRUD operation runs
- THEN the resource SHALL use the resource-scoped client, not the provider-level client

### Requirement: Create and update (REQ-012–REQ-014)

On create and update, the resource SHALL build a `LogstashPipeline` model from the Terraform plan (including `pipeline_id`, `description`, `pipeline`, `pipeline_metadata`, `pipeline_settings`, and `username`). The resource SHALL set `last_modified` to the current UTC time in strict date-time format before submitting the Put request. After a successful Put, the resource SHALL set `id` in state and perform a read to refresh all computed fields.

#### Scenario: Read-after-write on create

- GIVEN a successful Put Pipeline response
- WHEN create completes
- THEN the resource SHALL call the Get Pipeline API and populate state from the response

### Requirement: Read (REQ-015–REQ-017)

On read, the resource SHALL parse `id` using `ResourceIDFromStr` to extract the `pipeline_id`. The resource SHALL call `GetLogstashPipeline` with the extracted `pipeline_id`. When the API returns HTTP 404 (pipeline not found), the resource SHALL remove itself from state without returning an error. When the Get response does not include the requested `pipeline_id` in its result map, the resource SHALL return an error diagnostic.

#### Scenario: Not found removes from state

- GIVEN the Logstash Get Pipeline API returns HTTP 404
- WHEN read runs
- THEN the resource SHALL be removed from state and no error diagnostic SHALL be returned

#### Scenario: Pipeline absent from response map

- GIVEN the Get Pipeline API returns a success response but the expected pipeline_id is absent from the result
- WHEN read runs
- THEN the resource SHALL return an error diagnostic

### Requirement: Delete (REQ-018)

On delete, the resource SHALL parse `id` to extract the `pipeline_id` and call `DeleteLogstashPipeline` with that identifier. Non-success API responses SHALL be surfaced as error diagnostics.

#### Scenario: Delete uses pipeline_id from state

- GIVEN a resource with a valid composite id in state
- WHEN delete runs
- THEN the Logstash Delete Pipeline API SHALL be called with the extracted `pipeline_id`

### Requirement: pipeline_metadata mapping (REQ-019–REQ-021)

The resource SHALL validate `pipeline_metadata` as a JSON string at plan time. On create and update, the resource SHALL unmarshal `pipeline_metadata` into a `map[string]any` and include it in the API request body; if unmarshalling fails, the resource SHALL return an error diagnostic and SHALL NOT call the Put API. On read, the resource SHALL marshal the `pipeline_metadata` map from the API response back into a JSON string and store it in state. Semantic JSON equivalence (key ordering, whitespace) SHALL be suppressed during diff calculation so functionally identical JSON values do not produce a plan diff.

#### Scenario: Invalid pipeline_metadata JSON

- GIVEN `pipeline_metadata` is set to an invalid JSON string
- WHEN create or update runs
- THEN the resource SHALL return an error diagnostic and SHALL not call the Put Pipeline API

#### Scenario: pipeline_metadata round-trip

- GIVEN `pipeline_metadata` is set to `{"type":"logstash_pipeline","version":1}` in config
- WHEN the API returns the metadata as a parsed map
- THEN state SHALL contain the JSON-serialized string of that map

### Requirement: Pipeline settings mapping (REQ-022–REQ-025)

The resource SHALL map each optional pipeline and queue setting attribute to its corresponding dotted settings key (e.g. `pipeline_batch_delay` → `pipeline.batch.delay`) using `ConvertSettingsKeyToTFFieldKey`. On create and update, the resource SHALL include only those settings attributes that are explicitly set (non-zero/non-empty via `GetOk`) in the `pipeline_settings` map sent to the API. On read, the resource SHALL iterate over the known settings keys and set each Terraform attribute from the API response when the key is present; if a settings key is absent from the API response the resource SHALL skip it with a warning log. For settings of type `TypeInt`, the resource SHALL convert the float64 value returned by the JSON decoder to an integer using `math.Round` before setting it in state.

#### Scenario: Unset settings not sent to API

- GIVEN a pipeline with no settings attributes configured
- WHEN create runs
- THEN the `pipeline_settings` map sent to the API SHALL be empty

#### Scenario: Absent setting key on read

- GIVEN the API response `pipeline_settings` does not contain a key managed by the provider
- WHEN read runs
- THEN the provider SHALL log a warning and skip setting that attribute without returning an error

#### Scenario: Integer setting round-trip

- GIVEN `pipeline_workers` is set to `2`
- WHEN read maps the API response float64 value to state
- THEN state SHALL contain the integer value `2`

### Requirement: queue_max_bytes validation (REQ-026)

The resource SHALL validate `queue_max_bytes` at plan time against the pattern `^[0-9]+[kmgtp]?b$`; values that do not match SHALL produce a validation error.

#### Scenario: Invalid queue_max_bytes

- GIVEN `queue_max_bytes` is set to `"100"` (no unit suffix)
- WHEN the plan is applied
- THEN the provider SHALL return a validation error

### Requirement: pipeline_workers validation (REQ-027)

The resource SHALL validate `pipeline_workers` at plan time and reject values less than 1.

#### Scenario: pipeline_workers below minimum

- GIVEN `pipeline_workers` is set to `0`
- WHEN the plan is applied
- THEN the provider SHALL return a validation error

### Requirement: pipeline_ecs_compatibility and pipeline_ordered validation (REQ-028–REQ-029)

The resource SHALL validate `pipeline_ecs_compatibility` to one of `"disabled"`, `"v1"`, or `"v8"`. The resource SHALL validate `pipeline_ordered` to one of `"auto"`, `"true"`, or `"false"`.

#### Scenario: Invalid pipeline_ecs_compatibility

- GIVEN `pipeline_ecs_compatibility` is set to `"v9"` (not in the allowed set)
- WHEN the plan is applied
- THEN the provider SHALL return a validation error

#### Scenario: Invalid pipeline_ordered

- GIVEN `pipeline_ordered` is set to `"yes"` (not in the allowed set)
- WHEN the plan is applied
- THEN the provider SHALL return a validation error

### Requirement: queue_type validation (REQ-030)

The resource SHALL validate `queue_type` to one of `"memory"` or `"persisted"`.

#### Scenario: Invalid queue_type

- GIVEN `queue_type` is set to `"kafka"` (not in the allowed set)
- WHEN the plan is applied
- THEN the provider SHALL return a validation error
