# `elasticstack_elasticsearch_transform` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/transform/transform.go`

## Purpose

Define schema and behavior for the Elasticsearch transform resource: API usage, identity/import, connection, lifecycle, compatibility (version-gated settings), transform state management (start/stop via `enabled`), and mapping between Terraform configuration and the Elasticsearch transforms API.

## Schema

```hcl
resource "elasticstack_elasticsearch_transform" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<transform_name>
  name = <required, string> # force new; 1–64 chars; lowercase alphanumeric, hyphens, underscores; must start and end with lowercase alphanumeric

  description = <optional, string>

  source {                                            # required, max 1
    indices          = <required, list(string)>
    query            = <optional, string>             # JSON string; default: {"match_all":{}}; JSON-normalized diff suppression
    runtime_mappings = <optional, string>             # JSON string; JSON-normalized diff suppression; requires Elasticsearch >= 7.12.0
  }

  destination {                                       # required, max 1
    index    = <required, string>                     # 1–255 chars; lowercase alphanumeric + selected punctuation; cannot start with -, _, +
    pipeline = <optional, string>                     # requires Elasticsearch >= 7.3.0
    aliases {                                         # optional, list; requires Elasticsearch >= 8.8.0
      alias           = <required, string>
      move_on_creation = <optional, bool>             # default: false
    }
  }

  pivot  = <optional, string>   # JSON string; exactly one of pivot or latest required; force new; JSON-normalized diff suppression
  latest = <optional, string>   # JSON string; exactly one of pivot or latest required; force new; JSON-normalized diff suppression; requires Elasticsearch >= 7.11.0

  frequency = <optional, string> # Elastic duration string; default: "1m"; requires Elasticsearch >= 7.3.0

  metadata = <optional, string>  # JSON string; JSON-normalized diff suppression; requires Elasticsearch >= 7.16.0

  retention_policy {                                  # optional, max 1; requires Elasticsearch >= 7.12.0
    time {                                            # required, max 1
      field   = <required, string>
      max_age = <required, string>                    # Elastic duration string
    }
  }

  sync {                                              # optional, max 1
    time {                                            # required, max 1
      field = <required, string>
      delay = <optional, string>                      # Elastic duration string; default: "60s"
    }
  }

  # Settings — each requires a minimum Elasticsearch version (see Compatibility requirements)
  align_checkpoints    = <optional, bool>    # requires Elasticsearch >= 7.15.0
  dates_as_epoch_millis = <optional, bool>   # requires Elasticsearch >= 7.11.0
  deduce_mappings      = <optional, bool>    # requires Elasticsearch >= 8.1.0
  docs_per_second      = <optional, float>   # >= 0; requires Elasticsearch >= 7.8.0
  max_page_search_size = <optional, int>     # 10–65536; requires Elasticsearch >= 7.8.0
  num_failure_retries  = <optional, int>     # -1–100; requires Elasticsearch >= 8.4.0
  unattended           = <optional, bool>    # requires Elasticsearch >= 8.5.0

  defer_validation = <optional, bool>        # default: false
  timeout          = <optional, string>      # Go duration string; default: "30s"
  enabled          = <optional, bool>        # default: false; controls started/stopped state

  elasticsearch_connection {                          # optional, deprecated
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

### Requirement: Transform CRUD APIs (REQ-001–REQ-005)

The resource SHALL use the Elasticsearch Put Transform API to create transforms ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/put-transform.html)). The resource SHALL use the Elasticsearch Update Transform API to update transforms ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/update-transform.html)). The resource SHALL use the Elasticsearch Get Transform API to read transform definitions ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/get-transform.html)). The resource SHALL use the Elasticsearch Get Transform Statistics API to read transform run state ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/get-transform-stats.html)). The resource SHALL use the Elasticsearch Delete Transform API with `force=true` to delete transforms ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/delete-transform.html)). When Elasticsearch returns a non-success status for any API call (except 404 on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create

- GIVEN a non-success response from the Put Transform API
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

#### Scenario: API failure on delete

- GIVEN a non-success response from the Delete Transform API
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Minimum server version for transforms (REQ-006)

The resource SHALL verify the Elasticsearch server version is at least `7.2.0` before calling the Put Transform or Update Transform API. If the server version is lower, the resource SHALL fail with a "Transforms not supported" error diagnostic and SHALL not proceed with the API call.

#### Scenario: Server too old

- GIVEN an Elasticsearch server with version below `7.2.0`
- WHEN create or update is attempted
- THEN the resource SHALL return a "Transforms not supported" error and SHALL not call the API

### Requirement: Identity and import (REQ-007–REQ-009)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<transform_name>`. During create, the resource SHALL derive `id` from the current cluster UUID and the configured `name`, and SHALL set `id` in state after a successful Put Transform call. The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state. For read and delete operations, the resource SHALL parse `id` using `clients.CompositeIDFromStr` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Import passthrough

- GIVEN import with a valid `<cluster_uuid>/<transform_name>` id
- WHEN import completes
- THEN the id SHALL be stored in state for subsequent read, update, and delete operations

#### Scenario: Invalid id format on read

- GIVEN a state id not matching `<cluster_uuid>/<transform_name>`
- WHEN read runs
- THEN the resource SHALL return an error diagnostic

### Requirement: Lifecycle — name requires replacement (REQ-010)

Changing `name` SHALL require resource replacement (`ForceNew`). Changing `pivot` SHALL require resource replacement (`ForceNew`). Changing `latest` SHALL require resource replacement (`ForceNew`).

#### Scenario: Name change triggers replacement

- GIVEN an existing transform
- WHEN the `name` attribute is changed in configuration
- THEN Terraform SHALL plan a destroy-and-recreate (force new)

### Requirement: Connection (REQ-011–REQ-012)

By default, the resource SHALL use the provider-level Elasticsearch client obtained via `clients.NewAPIClientFromSDKResource`. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (create, read, update, delete).

#### Scenario: Resource-level client override

- GIVEN `elasticsearch_connection` is set with specific endpoints and credentials
- WHEN any API call is made
- THEN the resource-scoped client SHALL be used instead of the provider client

### Requirement: Transform state management — enabled (REQ-013–REQ-015)

On create, when `enabled` is `true`, the resource SHALL call the Elasticsearch Start Transform API after a successful Put Transform call. On update, when `enabled` has changed to `true`, the resource SHALL call the Start Transform API. On update, when `enabled` has changed to `false`, the resource SHALL call the Stop Transform API. When `enabled` has not changed during update, the resource SHALL NOT call Start or Stop Transform. On read, the resource SHALL derive the `enabled` value from transform statistics: `enabled` SHALL be `true` when the transform state is `"started"` or `"indexing"`, and `false` otherwise.

#### Scenario: Start on create with enabled=true

- GIVEN `enabled = true` in configuration
- WHEN create runs successfully
- THEN the resource SHALL call Start Transform after the Put Transform API call

#### Scenario: Stop on update with enabled=false

- GIVEN an enabled transform and `enabled = false` in updated configuration
- WHEN update runs
- THEN the resource SHALL call Stop Transform after the Update Transform API call

#### Scenario: No start/stop when enabled unchanged

- GIVEN `enabled` is unchanged between plan and apply
- WHEN update runs
- THEN the resource SHALL NOT call Start or Stop Transform

### Requirement: Timeout parameter (REQ-016–REQ-017)

The `timeout` attribute SHALL accept a Go duration string and SHALL default to `"30s"`. When the Elasticsearch server version is at least `7.17.0`, the resource SHALL pass the parsed `timeout` value as the API operation timeout parameter to the Put Transform, Update Transform, Start Transform, and Stop Transform APIs. When the server version is below `7.17.0`, the resource SHALL omit the timeout parameter from API calls.

#### Scenario: Timeout passed to API on supported version

- GIVEN an Elasticsearch server >= 7.17.0 and `timeout = "60s"`
- WHEN create or update runs
- THEN the API call SHALL include a 60-second timeout parameter

### Requirement: Defer validation (REQ-018)

The `defer_validation` attribute SHALL default to `false`. When `defer_validation` is `true`, the resource SHALL pass `defer_validation=true` to the Put Transform and Update Transform API calls, causing Elasticsearch to defer source index validation until the transform is started.

#### Scenario: Deferred validation on create

- GIVEN `defer_validation = true`
- WHEN create runs
- THEN the Put Transform API call SHALL include `defer_validation=true`

### Requirement: Pivot and latest are mutually exclusive (REQ-019)

Exactly one of `pivot` or `latest` MUST be configured. The schema SHALL enforce this with `ExactlyOneOf`. On update, the resource SHALL omit both `pivot` and `latest` from the Update Transform request body, as these fields cannot be changed after creation.

#### Scenario: Update omits pivot and latest

- GIVEN an existing transform with `pivot` set
- WHEN any attribute other than pivot/latest is updated
- THEN the Update Transform request body SHALL NOT include `pivot` or `latest` fields

### Requirement: Version-gated settings (REQ-020–REQ-032)

Settings and capabilities that require a minimum Elasticsearch version SHALL be silently omitted from API calls (with a log warning) when the server version is below the minimum. The version requirements are:

- `destination.pipeline`: requires Elasticsearch >= `7.3.0`
- `destination.aliases`: requires Elasticsearch >= `8.8.0`
- `frequency`: requires Elasticsearch >= `7.3.0`
- `latest`: requires Elasticsearch >= `7.11.0`
- `retention_policy`: requires Elasticsearch >= `7.12.0`
- `source.runtime_mappings`: requires Elasticsearch >= `7.12.0`
- `metadata`: requires Elasticsearch >= `7.16.0`
- `docs_per_second`: requires Elasticsearch >= `7.8.0`
- `max_page_search_size`: requires Elasticsearch >= `7.8.0`
- `dates_as_epoch_millis`: requires Elasticsearch >= `7.11.0`
- `align_checkpoints`: requires Elasticsearch >= `7.15.0`
- `deduce_mappings`: requires Elasticsearch >= `8.1.0`
- `num_failure_retries`: requires Elasticsearch >= `8.4.0`
- `unattended`: requires Elasticsearch >= `8.5.0`

#### Scenario: Version-gated setting silently omitted

- GIVEN `align_checkpoints = true` and an Elasticsearch server version below `7.15.0`
- WHEN create or update runs
- THEN `align_checkpoints` SHALL be omitted from the API request body and a warning SHALL be logged

### Requirement: Create and read-after-write (REQ-033)

After a successful Put Transform API call (and optional Start Transform), the resource SHALL call the read function to refresh state, ensuring the stored state reflects the server-side representation of the transform.

#### Scenario: State refreshed after create

- GIVEN a successful create
- WHEN create completes
- THEN the resource SHALL call read to populate state from the API response

### Requirement: Read — not found handling (REQ-034)

On read, when the Get Transform API returns HTTP 404, the resource SHALL remove itself from state (set `id` to `""`). When the API response does not contain a transform matching the requested name, the resource SHALL return an error diagnostic ("Unable to find the transform in the cluster").

#### Scenario: Transform not found on refresh

- GIVEN a transform that has been deleted outside of Terraform
- WHEN read runs
- THEN the resource SHALL be removed from state without error

### Requirement: Delete (REQ-035)

On delete, the resource SHALL parse `id` with `clients.CompositeIDFromStr` to extract the transform name and SHALL call Delete Transform with `force=true`. A non-success response SHALL be surfaced as an error diagnostic.

#### Scenario: Delete uses force flag

- GIVEN an existing transform
- WHEN delete runs
- THEN the Delete Transform API call SHALL include `force=true`

### Requirement: JSON field mapping — source (REQ-036–REQ-037)

The `source.query` field SHALL be validated as a JSON string by schema, SHALL default to `{"match_all":{}}`, and SHALL apply JSON-normalized diff suppression. On create and update, when `source.query` is set, the resource SHALL decode it into an `any` value for the API request. The `source.runtime_mappings` field SHALL be validated as a JSON string by schema, SHALL apply JSON-normalized diff suppression, and SHALL be decoded into an `any` value for the API request when non-empty.

#### Scenario: Invalid query JSON

- GIVEN an invalid JSON string in `source.query`
- WHEN create or update runs
- THEN the provider SHALL return an error and SHALL not call the Put or Update Transform API

### Requirement: JSON field mapping — pivot, latest, metadata (REQ-038–REQ-040)

The `pivot` and `latest` fields SHALL be validated as JSON strings and SHALL apply JSON-normalized diff suppression. On create, the resource SHALL decode `pivot` or `latest` (whichever is set) into an `any` value for the API request. The `metadata` field SHALL be validated as a JSON string and SHALL apply JSON-normalized diff suppression. On create and update, when `metadata` is set and the server version is at least `7.16.0`, the resource SHALL decode it into a `map[string]any` for the API request.

#### Scenario: Invalid pivot JSON rejected

- GIVEN an invalid JSON string in `pivot`
- WHEN create runs
- THEN the provider SHALL return an error and SHALL not call the Put Transform API

### Requirement: Read — state mapping from model (REQ-041–REQ-045)

On read, the resource SHALL set the following state attributes from the Get Transform API response:
- `description` from `transform.description`
- `source` (including `indices`, `query`, and `runtime_mappings`) from `transform.source`; `query` and `runtime_mappings` SHALL be JSON-marshaled back to strings
- `destination` (including `index`, `aliases`, and `pipeline`) from `transform.dest`
- `pivot` SHALL be JSON-marshaled from the API `pivot` value when non-nil
- `latest` SHALL be JSON-marshaled from the API `latest` value when non-nil
- `frequency` from `transform.frequency`
- `sync` from `transform.sync`
- `retention_policy` from `transform.retention_policy`
- Settings (`align_checkpoints`, `dates_as_epoch_millis`, `deduce_mappings`, `docs_per_second`, `max_page_search_size`, `num_failure_retries`, `unattended`) from `transform.settings` when non-nil
- `metadata` SHALL be JSON-marshaled from `transform._meta` when non-nil; when `_meta` is nil, `metadata` SHALL be set to nil in state

#### Scenario: Nil metadata cleared in state

- GIVEN a transform where `_meta` is not set on the server
- WHEN read runs
- THEN `metadata` SHALL be nil (not empty string) in state

### Requirement: Read — enabled from stats (REQ-046)

On read, after reading the transform definition, the resource SHALL call the Get Transform Statistics API. The resource SHALL set `enabled = true` when the transform state is `"started"` or `"indexing"`, and `enabled = false` for all other states.

#### Scenario: enabled=true when state is indexing

- GIVEN a transform with stats state `"indexing"`
- WHEN read runs
- THEN `enabled` SHALL be `true` in state

#### Scenario: enabled=false when state is stopped

- GIVEN a transform with stats state `"stopped"`
- WHEN read runs
- THEN `enabled` SHALL be `false` in state

### Requirement: Name validation (REQ-047)

The `name` attribute SHALL be validated to: be between 1 and 64 characters, contain only lowercase alphanumeric characters, hyphens, and underscores, and start and end with a lowercase alphanumeric character.

#### Scenario: Invalid transform name rejected

- GIVEN a `name` value that starts with a hyphen or contains uppercase characters
- WHEN the configuration is applied
- THEN the provider SHALL return a validation error
