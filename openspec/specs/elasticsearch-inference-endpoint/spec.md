# `elasticstack_elasticsearch_inference_endpoint` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/inference/inferenceendpoint`

## Purpose

Define schema and behavior for the Elasticsearch inference endpoint resource: API usage, identity/import, connection, lifecycle, version compatibility, JSON mapping, and read-time state reconciliation for sensitive settings and server-applied defaults.

## Schema

```hcl
resource "elasticstack_elasticsearch_inference_endpoint" "example" {
  id           = <computed, string> # internal identifier: <cluster_uuid>/<inference_id>
  inference_id = <required, string> # force new

  task_type = <optional, computed, string> # one of: sparse_embedding, text_embedding, rerank, completion, chat_completion, embedding; replacement when explicitly changed
  service   = <required, string>           # force new

  service_settings  = <required, json object string, sensitive>
  task_settings     = <optional, json object string>
  chunking_settings = <optional, json object string>

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

### Requirement: Inference endpoint CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Inference Put API to create inference endpoints, the Elasticsearch Inference Get API to read inference endpoints, the Elasticsearch Inference Update API to update inference endpoints, and the Elasticsearch Inference Delete API to delete inference endpoints. When Elasticsearch returns a non-success response for create, update, read, or delete requests, other than not found on read or delete, the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API error on create

- GIVEN the Inference Put API returns a non-success response
- WHEN create runs
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: Not found on read removes state

- GIVEN the Inference Get API returns no matching endpoint
- WHEN read runs
- THEN the resource SHALL be removed from state without an error

#### Scenario: Not found on delete is ignored

- GIVEN the Inference Delete API returns not found
- WHEN delete runs
- THEN the resource SHALL complete successfully without an error

### Requirement: Identity and import (REQ-005–REQ-007)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<inference_id>`. During create, the resource SHALL compute `id` from the current cluster UUID and configured `inference_id`. The resource SHALL support import as a passthrough on the `id` attribute, preserving the imported composite identifier for subsequent read and delete operations.

#### Scenario: ID computed on create

- GIVEN a successful create for `inference_id = "my-endpoint"`
- WHEN create completes
- THEN `id` SHALL be set to `<cluster_uuid>/my-endpoint`

#### Scenario: Import with composite id

- GIVEN an import identifier in the format `<cluster_uuid>/<inference_id>`
- WHEN import completes
- THEN the `id` SHALL be stored in state for subsequent read and delete operations

### Requirement: Lifecycle, connection, and framework implementation (REQ-008–REQ-011)

Changing `inference_id` SHALL require replacement. Changing `service` SHALL require replacement. Changing `task_type` SHALL require replacement when the attribute is explicitly configured. By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls. The resource SHALL be implemented on the Terraform Plugin Framework and SHALL preserve the existing Terraform type name, schema shape, and import behavior while using the shared Elasticsearch entitycore envelope behavior defined in `openspec/specs/entitycore-resource-envelope/spec.md`.

#### Scenario: Task type change requires replacement when configured

- GIVEN an existing resource with `task_type = "text_embedding"` configured
- WHEN `task_type` is changed to `"rerank"`
- THEN Terraform SHALL plan replacement

#### Scenario: Resource-level client override

- GIVEN `elasticsearch_connection` is configured
- WHEN API calls run
- THEN they SHALL use the resource-scoped client

### Requirement: Minimum Elasticsearch version (REQ-012)

The resource SHALL require Elasticsearch version `8.18.0` or later. When create is attempted against an older cluster, the resource SHALL return an error diagnostic and SHALL NOT call the Inference Put API.

#### Scenario: Unsupported cluster version

- GIVEN the connected Elasticsearch server version is below `8.18.0`
- WHEN create runs
- THEN the provider SHALL return an unsupported feature error
- AND it SHALL NOT call the Inference Put API

### Requirement: JSON validation and request mapping (REQ-013–REQ-016)

On create and update, the resource SHALL parse `service_settings`, `task_settings`, and `chunking_settings` from Terraform JSON strings into API request objects. If any configured JSON value cannot be parsed, the resource SHALL return an error diagnostic and SHALL NOT call the Elasticsearch API. The create request SHALL include `service`, and MAY include `task_type`, `service_settings`, `task_settings`, and `chunking_settings`. The update request SHALL include only mutable settings fields and SHALL omit immutable `service` from the request body.

#### Scenario: Invalid service settings JSON

- GIVEN `service_settings` contains invalid JSON
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic before calling Elasticsearch

#### Scenario: Update omits immutable service field

- GIVEN an existing endpoint is being updated
- WHEN the update request body is built
- THEN the request body SHALL omit `service`

### Requirement: Create and update behavior (REQ-017–REQ-018)

On create, the resource SHALL call the Inference Put API for the configured endpoint, set `id`, and then perform a read to refresh state. On update, the resource SHALL call the Inference Update API for the existing endpoint and then perform a read to refresh state.

#### Scenario: Read-after-write after create

- GIVEN a successful create API call
- WHEN create completes
- THEN the resource SHALL perform a read and persist the refreshed state

#### Scenario: Read-after-write after update

- GIVEN a successful update API call
- WHEN update completes
- THEN the resource SHALL perform a read and persist the refreshed state

### Requirement: Read mapping and state reconciliation (REQ-019–REQ-022)

On read, the resource SHALL set `inference_id`, `task_type`, and `service` from the API response. When `service_settings` is already known in Terraform state, the resource SHALL preserve the prior value instead of overwriting it from the API response, because Elasticsearch may omit sensitive fields such as API keys. The resource SHALL only adopt API `service_settings` when the prior Terraform value is null or unknown, such as on the first read after import.

For `task_settings`, when Terraform state already contains a configured JSON object, the resource SHALL retain only the keys explicitly present in Terraform state and SHALL discard API-returned keys that were not user-configured, so that server-applied defaults do not cause perpetual drift. If a user-configured `task_settings` key is returned by the API with a different value, the resource SHALL store the API value for that key so that real drift is visible to Terraform.

For `chunking_settings`, when the attribute is configured, the resource SHALL round-trip the API response back into Terraform state using the provider's defaults-aware JSON type. When the API omits `chunking_settings`, the resource SHALL clear `chunking_settings` from state.

#### Scenario: Sensitive service settings preserved from state

- GIVEN the prior Terraform state contains `service_settings` with sensitive credentials
- WHEN read runs and the API response omits those sensitive fields
- THEN the resource SHALL preserve the prior `service_settings` value in state

#### Scenario: Task settings ignore server defaults

- GIVEN the prior Terraform state sets `task_settings = {"temperature": 0.2}`
- AND the API response returns `{"temperature": 0.2, "max_input_tokens": 512}`
- WHEN read runs
- THEN the state SHALL retain only `{"temperature": 0.2}` for `task_settings`

#### Scenario: Chunking settings cleared when API omits them

- GIVEN the prior Terraform state contains `chunking_settings`
- WHEN read runs and the API response has no `chunking_settings`
- THEN the provider SHALL clear `chunking_settings` from state

### Requirement: Chunking settings defaults (REQ-023)

When `chunking_settings` is configured, the provider SHALL populate documented Elasticsearch defaults so that plan and state remain semantically equal when Elasticsearch echoes default values the user did not specify. For strategy `sentence`, defaults SHALL include `strategy = "sentence"`, `max_chunk_size = 250`, and `sentence_overlap = 1` when those keys are absent. For strategy `word`, defaults SHALL include `max_chunk_size = 250` and `overlap = 100` when those keys are absent. For strategies `none` and `recursive`, the provider SHALL preserve the user-supplied shape without applying a single default set.

#### Scenario: Sentence chunking defaults applied

- GIVEN `chunking_settings` omits `strategy`, `max_chunk_size`, and `sentence_overlap`
- WHEN the provider normalizes the value
- THEN the effective value SHALL include `strategy = "sentence"`, `max_chunk_size = 250`, and `sentence_overlap = 1`

#### Scenario: Word chunking overlap default applied

- GIVEN `chunking_settings = {"strategy": "word"}`
- WHEN the provider normalizes the value
- THEN the effective value SHALL include `max_chunk_size = 250` and `overlap = 100`
