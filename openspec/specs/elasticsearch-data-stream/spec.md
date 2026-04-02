# `elasticstack_elasticsearch_data_stream` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/data_stream.go`

## Purpose

Manage Elasticsearch data streams. A data stream is a named abstraction over a sequence of backing indices, used for append-only time-series data. The resource creates the stream by name (relying on a matching index template to already exist), exposes computed metadata about the stream, and deletes it on destroy. All mutable properties (generation, indices, status, etc.) are read-only from the API; only `name` is supplied by the operator.

## Schema

```hcl
resource "elasticstack_elasticsearch_data_stream" "example" {
  id              = <computed, string>       # internal identifier: <cluster_uuid>/<data_stream_name>
  name            = <required, string>       # force new; 1–255 chars; lower-case alphanumeric + selected punctuation; cannot start with -, _, +; cannot be "." or ".."

  timestamp_field = <computed, string>       # name of the @timestamp field for the stream
  generation      = <computed, int>          # current generation number
  status          = <computed, string>       # health status (green/yellow/red)
  template        = <computed, string>       # name of the matching index template
  ilm_policy      = <computed, string>       # ILM policy from the matching index template
  hidden          = <computed, bool>         # true if the stream is hidden
  system          = <computed, bool>         # true if the stream is system-managed
  replicated      = <computed, bool>         # true if managed by cross-cluster replication
  metadata        = <computed, string>       # JSON string of _meta from the matching index template; absent when null

  indices = <computed, list(object)> {       # ordered list of backing indices; last entry is the current write index
    index_name = <computed, string>
    index_uuid = <computed, string>
  }

  elasticsearch_connection {                 # optional resource-level connection override (deprecated in favour of provider-level config)
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

### Requirement: Data stream CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Create data stream API (`PUT /<data-stream>`) to create data streams ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-data-stream.html)). The resource SHALL use the Elasticsearch Get data stream API (`GET /_data_stream/<data-stream>`) to read data stream state ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-data-stream.html)). The resource SHALL use the Elasticsearch Delete data stream API (`DELETE /_data_stream/<data-stream>`) to delete data streams ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-delete-data-stream.html)). When Elasticsearch returns a non-success status for create, read, or delete requests (other than 404 on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create

- GIVEN the Create data stream API returns a non-success status
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error detail

#### Scenario: API failure on delete

- GIVEN the Delete data stream API returns a non-success status
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error detail

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<data_stream_name>`. During create, the resource SHALL compute `id` from the current cluster UUID and the configured `name`, then set it in state.

#### Scenario: ID set on create

- GIVEN a successful create
- WHEN the resource finishes `resourceDataStreamPut`
- THEN `id` in state SHALL match the pattern `<cluster_uuid>/<name>`

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the supplied `id` value directly to state. Read and delete operations SHALL parse `id` using `clients.CompositeIDFromStr` and SHALL return a "Wrong resource ID" error diagnostic when the format is not `<cluster_uuid>/<resource_identifier>`.

#### Scenario: Import passthrough

- GIVEN import with a valid composite id
- WHEN import completes
- THEN the id SHALL be stored and subsequent read and delete SHALL use the parsed resource identifier

#### Scenario: Invalid id format

- GIVEN an id that does not contain exactly one `/` separator
- WHEN read or delete attempts to parse it
- THEN the provider SHALL return an error diagnostic with summary "Wrong resource ID."

### Requirement: Lifecycle (REQ-009)

Changing `name` SHALL require replacement of the resource (`ForceNew`). There is no in-place update path for the data stream itself; any change to `name` destroys the old stream and creates a new one.

#### Scenario: Name change triggers replacement

- GIVEN an existing data stream and a plan that changes `name`
- WHEN Terraform plans the change
- THEN the plan SHALL show a destroy-and-recreate rather than an in-place update

### Requirement: Connection (REQ-010–REQ-011)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured on the resource, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (`clients.NewAPIClientFromSDKResource`).

#### Scenario: Resource-level client override

- GIVEN `elasticsearch_connection` block is configured on the resource
- WHEN create, read, or delete API calls run
- THEN they SHALL use the resource-scoped client derived from that block

### Requirement: Create and read-after-create (REQ-012–REQ-013)

On create, the resource SHALL call the Create data stream API with the value of `name`. After a successful create, the resource SHALL immediately call `resourceDataStreamRead` to populate all computed attributes in state.

#### Scenario: Read-after-create populates computed fields

- GIVEN a successful create API call
- WHEN `resourceDataStreamPut` completes
- THEN computed attributes (`timestamp_field`, `generation`, `status`, `template`, `ilm_policy`, `hidden`, `system`, `replicated`, `indices`) SHALL be populated in state from the API response

### Requirement: Read and not-found handling (REQ-014–REQ-015)

On read, the resource SHALL parse `id` to extract the data stream name, then call the Get data stream API. When the API returns HTTP 404 (data stream not found), the resource SHALL remove itself from Terraform state by setting the id to empty string and return without error. When the API returns a success response, the resource SHALL decode the first element of the `data_streams` array from the response body.

#### Scenario: Data stream not found on read

- GIVEN the Get data stream API returns 404
- WHEN read runs
- THEN the resource SHALL be removed from state and no error diagnostic SHALL be returned

#### Scenario: Successful read decodes first stream

- GIVEN a successful Get data stream response
- WHEN read runs
- THEN the resource SHALL decode `data_streams[0]` from the response JSON

### Requirement: Delete (REQ-016)

On delete, the resource SHALL parse `id` to extract the data stream name, then call the Delete data stream API with that name.

#### Scenario: Delete uses parsed name

- GIVEN an existing data stream resource in state
- WHEN destroy runs
- THEN the Delete data stream API SHALL be called with the resource identifier parsed from `id`

### Requirement: Name validation (REQ-017–REQ-021)

The `name` attribute SHALL be validated against all of the following rules before the API is called:

- Length MUST be between 1 and 255 characters.
- The value MUST NOT be `"."` or `".."`.
- The value MUST NOT start with `-`, `_`, or `+`.
- The value MUST contain only lower-case alphanumeric characters and the following punctuation: `!`, `$`, `%`, `&`, `'`, `(`, `)`, `+`, `.`, `;`, `=`, `@`, `[`, `]`, `^`, `{`, `}`, `~`, `_`, `-`.

When validation fails, Terraform SHALL reject the configuration before planning.

#### Scenario: Name starts with hyphen

- GIVEN `name = "-foo"`
- WHEN the configuration is validated
- THEN Terraform SHALL return an error matching "cannot start with -, _, +"

#### Scenario: Name contains upper-case characters

- GIVEN `name = "FooBar"`
- WHEN the configuration is validated
- THEN Terraform SHALL return an error matching "must contain lower case alphanumeric characters"

### Requirement: Read state mapping (REQ-022–REQ-027)

On read, the resource SHALL map the API response fields to Terraform state as follows:

- `name` SHALL be set from `ds.Name`.
- `timestamp_field` SHALL be set from `ds.TimestampField.Name`.
- `generation` SHALL be set from `ds.Generation`.
- `status` SHALL be set from `ds.Status`.
- `template` SHALL be set from `ds.Template`.
- `ilm_policy` SHALL be set from `ds.IlmPolicy`.
- `hidden` SHALL be set from `ds.Hidden`.
- `system` SHALL be set from `ds.System`.
- `replicated` SHALL be set from `ds.Replicated`.
- When `ds.Meta` is non-nil, it SHALL be JSON-serialized and stored in the `metadata` attribute; when `ds.Meta` is nil, `metadata` SHALL not be set (preserving the absent/null state).
- `indices` SHALL be set as a list of objects, each with `index_name` and `index_uuid`, in the order returned by the API.

#### Scenario: Metadata absent when nil

- GIVEN a data stream whose `_meta` field is null in the API response
- WHEN read runs
- THEN the `metadata` attribute SHALL not be set in state (no drift)

#### Scenario: Metadata populated when present

- GIVEN a data stream with a non-null `_meta` object in the API response
- WHEN read runs
- THEN `metadata` SHALL contain the JSON-serialized form of that object
