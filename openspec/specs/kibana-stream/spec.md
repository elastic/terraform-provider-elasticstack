# `elasticstack_kibana_stream` — Schema and Functional Requirements

Resource implementation: `internal/kibana/streams`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_stream` resource: Kibana Streams API usage, composite identity and import, mutually exclusive stream-type configuration, experimental and version-gated stream support, classic-stream adoption behavior, and the mapping between Terraform state and the API models for wired, classic, and query streams.

## Schema

```hcl
resource "elasticstack_kibana_stream" "example" {
  id          = <computed, string> # composite "<space_id>/<name>"; UseStateForUnknown
  space_id    = <optional, computed, string> # default "default"; RequiresReplace
  name        = <required, string> # RequiresReplace
  description = <optional, computed, string> # default ""

  wired_config = <optional, object({
    processing_steps         = <optional, list(normalized json string)>
    fields_json              = <optional, normalized json string>
    routing_json             = <optional, normalized json string>
    lifecycle_json           = <optional, computed, normalized json string> # UseStateForUnknown
    failure_store_json       = <optional, computed, normalized json string> # UseStateForUnknown
    index_number_of_shards   = <optional, int64>
    index_number_of_replicas = <optional, int64>
    index_refresh_interval   = <optional, string>
  })>

  classic_config = <optional, object({
    processing_steps         = <optional, list(normalized json string)>
    field_overrides_json     = <optional, normalized json string>
    lifecycle_json           = <optional, computed, normalized json string> # UseStateForUnknown
    failure_store_json       = <optional, computed, normalized json string> # UseStateForUnknown
    index_number_of_shards   = <optional, int64>
    index_number_of_replicas = <optional, int64>
    index_refresh_interval   = <optional, string>
  })>

  query_config = <optional, object({
    esql = <required, string>
    view = <computed, string> # always derived as "$.<name>" on write
  })>

  dashboards = <optional, list(string)>

  queries = <optional, list(object({
    id             = <required, string>
    title          = <required, string>
    description    = <optional, computed, string> # default ""
    esql           = <required, string>
    severity_score = <optional, float64>
    evidence       = <optional, list(string)>
  }))>
}
```

Notes:

- Exactly one of `wired_config`, `classic_config`, or `query_config` is required by resource config validation.
- The schema documentation marks Streams as technical preview and says the feature requires Elastic Stack 9.4.0 or higher.
- `elasticstack_kibana_stream` is registered through the provider's experimental Plugin Framework resource set.
- The resource does not define a schema version, custom state upgrader, or resource-level `kibana_connection` override block.

## Requirements

### Requirement: Streams APIs and write-lock retry (REQ-001)

The resource SHALL manage streams through Kibana Streams HTTP APIs: get stream, put stream, and delete stream. For non-default spaces it SHALL call those APIs through the space-aware path for the configured `space_id`. Stream create and update SHALL use the PUT upsert API, and delete SHALL use the delete API. When the Kibana Streams write APIs return HTTP 409 lock-contention responses, the provider SHALL retry with exponential backoff for up to five attempts before surfacing failure.

#### Scenario: Lock contention during upsert

- GIVEN a create, update, or delete operation for a stream
- WHEN the Kibana Streams API returns HTTP 409 because it could not acquire the write lock
- THEN the provider SHALL retry the operation with exponential backoff before failing

### Requirement: Composite identity and import passthrough (REQ-002)

The resource SHALL use a computed canonical `id` in the format `<space_id>/<name>`. `space_id` SHALL default to `default` when omitted. On create, the resource SHALL set `id` from the configured `space_id` and `name` before reading the stream back. Import SHALL pass the supplied identifier directly into `id`. Subsequent read, update, and delete operations SHALL require that stored `id` be parseable as a composite identifier whose resource segment is the stream name and whose cluster segment is the Kibana space.

#### Scenario: Imported composite id

- GIVEN an import id of `observability/logs.otel.errors`
- WHEN import completes
- THEN the provider SHALL store that exact string in `id`
- AND later CRUD operations SHALL use `observability` as the space and `logs.otel.errors` as the stream name

### Requirement: Stream type selection and replacement fields (REQ-003)

Terraform configuration SHALL select exactly one stream type by setting exactly one of `wired_config`, `classic_config`, or `query_config`. The provider SHALL reject configurations that set none or more than one of those blocks. Changing `space_id` or `name` SHALL require replacement rather than in-place update.

#### Scenario: Multiple config blocks rejected

- GIVEN a configuration that sets both `wired_config` and `query_config`
- WHEN Terraform validates the resource configuration
- THEN the provider SHALL reject the configuration before making any Streams API call

#### Scenario: Renaming a stream

- GIVEN a configuration change to `name`
- WHEN Terraform plans the change
- THEN the resource SHALL be replaced

### Requirement: Provider-level Kibana client only (REQ-004)

The resource SHALL use the provider's configured Kibana OpenAPI client for read, create, update, and delete operations. The resource SHALL NOT expose or honor a resource-level Kibana connection override block.

#### Scenario: Standard provider connection

- GIVEN the provider is configured with Kibana access
- WHEN the stream resource performs API calls
- THEN those calls SHALL use the provider-level Kibana OpenAPI client

### Requirement: Streams version gate and classic create restriction (REQ-005)

Before create or update, the resource SHALL enforce the provider's minimum version gate for Kibana Streams support using `9.4.0-SNAPSHOT` as the threshold. If the target does not satisfy that gate, the operation SHALL fail with an `Unsupported server version` diagnostic and SHALL NOT call the upsert API. In addition, create SHALL reject `classic_config` because classic streams are adopted existing data streams and cannot be created through this resource.

#### Scenario: Unsupported server version

- GIVEN a non-classic stream create or update against a target below the resource's minimum Streams version gate
- WHEN the operation begins
- THEN the provider SHALL fail with an `Unsupported server version` diagnostic
- AND SHALL NOT call the Kibana Streams upsert API

#### Scenario: Classic stream create rejected

- GIVEN a configuration that uses `classic_config`
- WHEN create runs
- THEN the provider SHALL fail with a diagnostic explaining that classic streams must be imported instead of created

### Requirement: Create and update use authoritative read-after-write (REQ-006)

For non-classic create and for all updates, the resource SHALL build a Streams upsert request from Terraform state and call the Kibana Streams upsert API. After a successful upsert, the resource SHALL read the stream back and use that read result as the authoritative final state. If the stream cannot be read back after a successful create or update, the operation SHALL fail with an error diagnostic.

#### Scenario: Successful create uses read-after-write

- GIVEN a valid wired or query stream configuration
- WHEN the upsert API succeeds during create
- THEN the provider SHALL read the stream back
- AND SHALL fail the create if the stream cannot be read after creation

### Requirement: Read removes missing streams from state (REQ-007)

On refresh, the resource SHALL parse the composite `id`, read the stream from Kibana, and repopulate Terraform state from the API response. If the Kibana Streams get API returns not found, the provider SHALL remove the resource from Terraform state. Unexpected HTTP statuses and transport errors from the get API SHALL be surfaced as diagnostics.

#### Scenario: Stream removed outside Terraform

- GIVEN a stream recorded in Terraform state
- WHEN refresh runs and the Kibana Streams API returns not found
- THEN the provider SHALL remove the resource from state

### Requirement: Delete behavior differs for classic and non-classic streams (REQ-008)

For wired and query streams, delete SHALL parse the composite `id` and call the Kibana Streams delete API for that space and stream name. For classic streams, delete SHALL make no API call and SHALL only allow Terraform state removal to proceed. Delete of a non-classic stream SHALL treat API not-found as success. Unexpected delete statuses and transport errors SHALL be surfaced as diagnostics.

#### Scenario: Classic stream destroy

- GIVEN a managed stream whose state uses `classic_config`
- WHEN Terraform destroys the resource
- THEN the provider SHALL NOT call the Kibana Streams delete API
- AND Terraform state removal SHALL still complete

#### Scenario: Deleting an already-absent wired stream

- GIVEN a managed wired or query stream that has already been removed from Kibana
- WHEN delete runs
- THEN the provider SHALL treat Kibana's not-found response as success

### Requirement: Upsert request shape and stream discriminator (REQ-009)

The resource SHALL build stream upsert requests with the top-level arrays `dashboards`, `rules`, and `queries` always present, even when empty. The provider SHALL derive the stream type discriminator from the configured block: `wired` for `wired_config`, `classic` for `classic_config`, and `query` for `query_config`. `description` SHALL always be sent from Terraform state.

#### Scenario: Empty dashboards and queries still sent

- GIVEN a stream configuration that omits `dashboards` and `queries`
- WHEN the provider builds the upsert request
- THEN it SHALL send empty arrays for `dashboards`, `rules`, and `queries` rather than omitting them or sending null

### Requirement: Wired stream mapping and defaults (REQ-010)

For `wired_config`, the resource SHALL map `processing_steps` as an ordered list of JSON-encoded streamlang steps. On write, it SHALL always send a `wired` block, default `fields` to `{}` when `fields_json` is not configured, and default `routing` to an empty array when `routing_json` is not configured. `lifecycle_json` SHALL default to `{"inherit":{}}` when not known at write time, and `failure_store_json` SHALL default to `{"disabled":{}}` when not known at write time. On read, an empty processing-step array SHALL become null in Terraform state, an empty wired `fields` object `{}` SHALL become null, and absent index settings SHALL become null. Invalid `routing_json` in configuration SHALL produce an error diagnostic instead of sending the request.

#### Scenario: Wired defaults on create

- GIVEN a wired stream configuration that omits `fields_json`, `routing_json`, `lifecycle_json`, and `failure_store_json`
- WHEN the provider builds the upsert request
- THEN it SHALL send `fields = {}` and `routing = []`
- AND it SHALL send lifecycle as `{"inherit":{}}`
- AND it SHALL send failure store as `{"disabled":{}}`

#### Scenario: Invalid routing_json

- GIVEN a wired stream configuration whose `routing_json` is not valid JSON for routing rules
- WHEN the provider builds the upsert request
- THEN the provider SHALL return an error diagnostic and SHALL NOT call the upsert API

### Requirement: Classic stream mapping and adoption behavior (REQ-011)

For `classic_config`, the resource SHALL map `processing_steps`, `field_overrides_json`, lifecycle, failure store, and index settings through the ingest API model. On write, lifecycle SHALL default to `{"inherit":{}}` when not known and failure store SHALL default to `{"disabled":{}}` when not known. The provider SHALL only send the classic ingest sub-object when `field_overrides_json` is configured. Classic streams SHALL therefore be manageable only after import: create is rejected, update uses the upsert API, and delete is state-only.

#### Scenario: Imported classic stream update

- GIVEN a classic stream that has been imported into Terraform state
- WHEN update runs
- THEN the provider SHALL send the classic stream through the upsert API using the current Terraform state

### Requirement: Query stream ES|QL mapping and derived view (REQ-012)

For `query_config`, the resource SHALL require `esql` and SHALL derive the API `view` field as `$.<name>` from the stream name on every write. The provider SHALL ignore practitioner control over `view` and treat it as computed state. On read, the provider SHALL populate `query_config.view` from the API response and SHALL preserve an empty string as an empty string value rather than converting it to null.

#### Scenario: Query view derived from stream name

- GIVEN a query stream named `logs.otel.errors-view`
- WHEN create or update builds the API request
- THEN the provider SHALL send `view = "$.logs.otel.errors-view"` regardless of prior state

### Requirement: Dashboards and attached queries round-trip through state (REQ-013)

The resource SHALL map `dashboards` to the list of dashboard IDs returned by the Streams API. When the API returns no dashboards, the provider SHALL store `dashboards` as null rather than an empty list. Attached `queries` SHALL round-trip `id`, `title`, `description`, `esql`, optional `severity_score`, and optional `evidence`. When the API omits `severity_score`, the provider SHALL store it as null. When the API omits or empties `evidence`, the provider SHALL store it as null. When the API returns no attached queries, the provider SHALL store `queries` as absent.

#### Scenario: Query metadata with optional fields

- GIVEN a stream with an attached query that omits `severity_score` and `evidence`
- WHEN the provider reads the stream into state
- THEN the attached query SHALL have null `severity_score` and null `evidence`

## Traceability

| Area | Primary files |
|------|---------------|
| Schema | `internal/kibana/streams/schema.go` |
| Metadata / Configure / Import / Validators | `internal/kibana/streams/resource.go` |
| Create | `internal/kibana/streams/create.go` |
| Read | `internal/kibana/streams/read.go` |
| Update | `internal/kibana/streams/update.go` |
| Delete | `internal/kibana/streams/delete.go` |
| Shared create/update path | `internal/kibana/streams/upsert.go` |
| Terraform/API models | `internal/kibana/streams/models.go`, `internal/kibana/streams/models_wired.go`, `internal/kibana/streams/models_classic.go`, `internal/kibana/streams/models_query.go` |
| Streams API client and retry behavior | `internal/clients/kibanaoapi/streams.go` |
| Composite id parsing | `internal/clients/api_client.go` |
