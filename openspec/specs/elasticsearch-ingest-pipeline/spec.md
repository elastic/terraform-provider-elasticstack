# `elasticstack_elasticsearch_ingest_pipeline` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ingest/pipeline.go`

## Purpose

Define schema and behavior for the Elasticsearch ingest pipeline resource: API usage, identity/import, connection, JSON mapping, and read-time state handling for processors and on-failure handlers.

## Schema

```hcl
resource "elasticstack_elasticsearch_ingest_pipeline" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<pipeline_name>
  name = <required, string> # force new

  description = <optional, string>
  processors  = <required, list(string), min 1> # each element is a JSON object string; validated as JSON
  on_failure  = <optional, list(string), min 1> # each element is a JSON object string; validated as JSON
  metadata    = <optional, json string>          # validated as JSON; DiffSuppressFunc normalizes JSON

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

### Requirement: Ingest pipeline CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put pipeline API to create and update ingest pipelines ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/put-pipeline-api.html)). The resource SHALL use the Elasticsearch Get pipeline API to read ingest pipelines ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/get-pipeline-api.html)). The resource SHALL use the Elasticsearch Delete pipeline API to delete ingest pipelines ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/delete-pipeline-api.html)). When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than not found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create/update

- GIVEN the Put pipeline API returns a non-success response
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: API failure on delete

- GIVEN the Delete pipeline API returns a non-success response
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the API error

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<pipeline_name>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and the configured `name`, then persist it to state.

#### Scenario: ID format

- GIVEN a pipeline with name `my-pipeline` created against a cluster with UUID `abc123`
- WHEN create completes
- THEN the `id` in state SHALL be `abc123/my-pipeline`

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state. Read and delete operations SHALL parse `id` as `<cluster_uuid>/<resource identifier>` and SHALL return a "Wrong resource ID." error diagnostic when the format does not contain exactly one `/` separator.

#### Scenario: Import passthrough

- GIVEN import with a valid composite id `<cluster_uuid>/<pipeline_name>`
- WHEN import completes
- THEN the id SHALL be stored in state for subsequent operations

#### Scenario: Invalid id format on read or delete

- GIVEN an `id` value that does not contain exactly one `/`
- WHEN read or delete runs
- THEN the provider SHALL return an error diagnostic with summary "Wrong resource ID."

### Requirement: Lifecycle (REQ-009)

Changing `name` SHALL require replacement of the resource (`ForceNew`).

#### Scenario: Name change forces replacement

- GIVEN an existing ingest pipeline with a particular name
- WHEN the `name` attribute is changed in configuration
- THEN Terraform SHALL plan a destroy-and-recreate

### Requirement: Connection (REQ-010–REQ-011)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls on that resource.

#### Scenario: Resource-level connection override

- GIVEN `elasticsearch_connection` is set on the resource
- WHEN create, read, update, or delete API calls are made
- THEN they SHALL use the resource-scoped client derived from `elasticsearch_connection`

### Requirement: Create and update (REQ-012–REQ-014)

On create and update, the resource SHALL build an `IngestPipeline` request body from the Terraform configuration, encoding each `processors` element and each `on_failure` element as parsed JSON objects, and optionally including `description` and `metadata`. The resource SHALL submit the request body to the Put pipeline API. After a successful Put request, the resource SHALL persist `id` and perform a read to refresh state.

#### Scenario: Processors encoded as JSON objects

- GIVEN `processors` contains one or more JSON strings
- WHEN create or update runs
- THEN each string SHALL be parsed and sent as a JSON object in the API request body's `processors` array

#### Scenario: Read-after-write on create

- GIVEN a successful Put pipeline response
- WHEN create completes
- THEN the resource SHALL call read to populate state from the API response

### Requirement: Read (REQ-015–REQ-017)

On read, the resource SHALL parse `id` as `<cluster_uuid>/<pipeline_name>`, call the Get pipeline API with the pipeline name, and remove the resource from state (set `id` to `""`) when the pipeline is not found (HTTP 404). On a successful get, the resource SHALL set `name`, `description` (when present in the response), `processors`, `on_failure` (when present), and `metadata` (when present) from the API response.

#### Scenario: Pipeline not found on refresh

- GIVEN the pipeline has been deleted outside of Terraform
- WHEN read runs
- THEN the provider SHALL remove the resource from state by setting `id` to `""`

#### Scenario: State populated from API response

- GIVEN a successful Get pipeline response
- WHEN read completes
- THEN `name`, `processors`, and any present optional fields SHALL be set in state from the API response

### Requirement: Delete (REQ-018–REQ-019)

On delete, the resource SHALL parse `id` as `<cluster_uuid>/<pipeline_name>` and call the Delete pipeline API with the pipeline name. A non-success response from the Delete pipeline API SHALL be surfaced as a Terraform error diagnostic.

#### Scenario: Delete API called with correct name

- GIVEN a resource with id `<cluster_uuid>/my-pipeline`
- WHEN delete runs
- THEN the Delete pipeline API SHALL be called with pipeline name `my-pipeline`

### Requirement: JSON mapping for processors and on_failure (REQ-020–REQ-022)

Each element of `processors` and `on_failure` SHALL be validated as a JSON string by schema (`ValidateFunc: validation.StringIsJSON`). On create/update, each element SHALL be decoded from its JSON string into a `map[string]any` before being included in the API request body. On read, each processor and on_failure handler object received from the API SHALL be marshalled back to a JSON string and stored as the corresponding list element in state.

#### Scenario: Invalid processor JSON

- GIVEN a `processors` element that is not valid JSON
- WHEN the configuration is applied
- THEN Terraform validation SHALL reject it before calling the API

#### Scenario: Round-trip processor JSON

- GIVEN a `processors` list with JSON strings
- WHEN create runs and then read runs
- THEN the `processors` state SHALL contain JSON strings representing the same objects as returned by the API

### Requirement: JSON mapping for metadata (REQ-023–REQ-024)

`metadata` SHALL be validated as a JSON string by schema. On create/update, when `metadata` is set, the resource SHALL decode it from a JSON string into a `map[string]any` and include it as the `_meta` field in the API request body. On read, when `metadata` is present in the API response, the resource SHALL marshal it to a JSON string and store it in state. Semantically equivalent JSON values SHALL be treated as equal by the diff suppression function, preventing spurious plan differences due to key ordering or whitespace.

#### Scenario: Metadata round-trip

- GIVEN `metadata` is set to a JSON object string
- WHEN create runs and then read runs
- THEN the `metadata` state SHALL contain a JSON string representing the same object as stored in the API

#### Scenario: Invalid metadata JSON

- GIVEN `metadata` is set to an invalid JSON string
- WHEN the configuration is applied
- THEN Terraform validation SHALL reject it before calling the API
