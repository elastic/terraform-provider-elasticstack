# `elasticstack_elasticsearch_index_template` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/template.go`

## Purpose

Define schema and behavior for the Elasticsearch index template resource: API usage, identity/import, connection, compatibility, mapping, and state refresh semantics including alias routing quirks.

## Schema

```hcl
resource "elasticstack_elasticsearch_index_template" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<template_name>
  name = <required, string> # force new

  composed_of                         = <optional, computed, list(string)>
  ignore_missing_component_templates  = <optional, computed, list(string)> # requires Elasticsearch >= 8.7.0 when non-empty
  index_patterns                      = <required, set(string)>
  metadata                            = <optional, json string>
  priority                            = <optional, int> # must be >= 0
  version                             = <optional, int>

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

  data_stream {
    hidden               = <optional, bool>
    allow_custom_routing = <optional, bool>
  }

  template {
    mappings = <optional, json object string>
    settings = <optional, json object string>
    alias {
      name           = <required, string>
      filter         = <optional, json string>
      index_routing  = <optional, computed, string>
      is_hidden      = <optional, bool>
      is_write_index = <optional, bool>
      routing        = <optional, string>
      search_routing = <optional, computed, string>
    }
    lifecycle {
      data_retention = <required, string>
    }
  }
}
```
## Requirements
### Requirement: Index template CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put index template API to create and update index templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-put-template.html)). The resource SHALL use the Elasticsearch Get index template API to read index templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-template.html)). The resource SHALL use the Elasticsearch Delete index template API to delete index templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-delete-template.html)). When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than not found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API errors surfaced

- GIVEN a failing Elasticsearch response (other than 404 on read)
- WHEN the provider processes the response
- THEN diagnostics SHALL include the API error

### Requirement: Identity and import (REQ-005–REQ-008)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<template_name>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and configured `name`. The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state. For imported or stored `id` values, read/delete operations SHALL require the format `<cluster_uuid>/<resource identifier>` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Invalid id on read

- GIVEN a malformed `id` in state
- WHEN read or delete runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Lifecycle and connection (REQ-009–REQ-011)

Changing `name` SHALL require replacement (`ForceNew`). By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls.

#### Scenario: ForceNew on name change

- GIVEN `name` changes in configuration
- WHEN Terraform plans
- THEN replacement SHALL be required

### Requirement: Compatibility (REQ-012)

When `ignore_missing_component_templates` is configured with one or more values, the resource SHALL require Elasticsearch version >= 8.7.0; otherwise it SHALL return an error diagnostic.

#### Scenario: Feature on old cluster

- GIVEN non-empty `ignore_missing_component_templates` and ES < 8.7.0
- WHEN create or update runs
- THEN the provider SHALL error

### Requirement: Create, update, and read (REQ-013–REQ-016)

On create/update, the resource SHALL construct a `models.IndexTemplate` request body from Terraform state and submit it with the Put index template API. After a successful Put request, the resource SHALL set `id` and perform a read to refresh state. On read, the resource SHALL parse `id`, fetch the index template by name, and remove the resource from state when the template is not found. If the Get index template API returns a result count other than exactly one template, the read path SHALL return an error diagnostic.

#### Scenario: Template not found on refresh

- GIVEN the template was deleted in Elasticsearch
- WHEN read runs
- THEN the resource SHALL be removed from state

### Requirement: Delete (REQ-017)

On delete, the resource SHALL parse `id` and delete the template identified by the parsed resource identifier.

#### Scenario: Destroy deletes by parsed name

- GIVEN destroy
- WHEN delete runs
- THEN Delete index template SHALL be called for the parsed identifier

### Requirement: JSON and object mapping (REQ-018–REQ-025)

`metadata` SHALL be validated as JSON by schema and parsed as JSON during create/update; if parsing fails, the resource SHALL return an error diagnostic and SHALL not call the Put API. `template.mappings` and `template.settings` SHALL be validated as JSON objects by schema and parsed into objects during create/update. `template.alias.filter` SHALL be validated as JSON by schema and parsed into an object when non-empty during create/update. `template.alias` SHALL be mapped as a set keyed by alias name in API payload/state conversion. Alias routing and flag fields SHALL be copied directly between Terraform values and API model fields. `template.lifecycle` SHALL be mapped as at most one lifecycle object with `data_retention`. `data_stream.hidden` SHALL be sent when present. `data_stream.allow_custom_routing` SHALL be sent only when `true`, except that on updates it SHALL also be sent when prior state had `allow_custom_routing=true` (8.x workaround behavior).

#### Scenario: Invalid metadata JSON

- GIVEN invalid `metadata` JSON
- WHEN create/update runs
- THEN the provider SHALL error before calling Put

### Requirement: Read state mapping (REQ-026–REQ-030)

On read, the resource SHALL set `name`, `composed_of`, `ignore_missing_component_templates`, `index_patterns`, `priority`, and `version` from the API response. On read, when API `metadata` is present, it SHALL be serialized into a JSON string and stored in state. On read, when API `template` is present, it SHALL be flattened into `template` state, including aliases, lifecycle, mappings, and settings. On read, when API `data_stream` is present, it SHALL be flattened into a single `data_stream` block and include only fields present in API response. User-defined alias `routing` SHALL be preserved during read/refresh, because this field may be omitted by the API response and therefore SHALL not be overwritten from response data.

#### Scenario: Preserve user routing

- GIVEN the user set alias `routing` and the API omits it on read
- WHEN read refreshes state
- THEN configured `routing` SHALL not be clobbered from empty API fields

### Requirement: Alias routing plan diff suppression (REQ-031)

The provider SHALL ignore plan diffs to alias `search_routing` or `index_routing` when: the attribute is unset in the current config, and the state value for the attribute equals the current value of the `routing` attribute.

#### Scenario: Routing-only alias config

- GIVEN a configuration which configures only the `routing` attribute on an alias
- WHEN apply completes and state is refreshed
- THEN `search_routing` and `index_routing` in state SHALL match the `routing` attribute as documented

### Requirement: Schema — `template.data_stream_options` block (REQ-032)

The `template` block SHALL support an optional `data_stream_options` sub-block. The `data_stream_options` block SHALL contain at most one optional `failure_store` sub-block. If `data_stream_options` is configured without a `failure_store` sub-block, the provider SHALL reject the configuration at plan time with an error diagnostic. The `failure_store` block SHALL contain:

- `enabled` — required boolean; activates or deactivates document redirection to the failure store on newly created matching data streams.
- `lifecycle` — optional sub-block containing `data_retention`, a required string specifying how long failure store documents are retained (e.g. `"30d"`).

When `data_stream_options` is omitted from the configuration, the provider SHALL not include the field in API requests and SHALL leave it unset in Terraform state.

**Example HCL:**

```hcl
resource "elasticstack_elasticsearch_index_template" "example" {
  name           = "my-index-template"
  index_patterns = ["my-datastream-*"]

  template {
    data_stream_options {
      failure_store {
        enabled = true
        lifecycle {
          data_retention = "30d"
        }
      }
    }
  }

  data_stream {}
}
```

#### Scenario: `data_stream_options` omitted

- **WHEN** `data_stream_options` is not configured
- **THEN** the provider SHALL not include `data_stream_options` in the Put index template API request body

#### Scenario: `failure_store.enabled` without lifecycle

- **WHEN** `failure_store.enabled = true` is configured and `lifecycle` is omitted
- **THEN** the provider SHALL send `{"failure_store": {"enabled": true}}` inside `template.data_stream_options` in the API request

#### Scenario: `failure_store` with lifecycle retention

- **WHEN** `failure_store.enabled = true` and `failure_store.lifecycle.data_retention = "10d"` are configured
- **THEN** the provider SHALL send `{"failure_store": {"enabled": true, "lifecycle": {"data_retention": "10d"}}}` inside `template.data_stream_options`

---

### Requirement: Compatibility — version gate for `data_stream_options` (REQ-033)

When `data_stream_options` is configured and the Elasticsearch server version is below `9.1.0`, the provider SHALL return an error diagnostic and SHALL not call the Put index template API.

#### Scenario: Feature on unsupported cluster version

- **GIVEN** `data_stream_options` is configured
- **AND** the connected Elasticsearch server version is below `9.1.0`
- **WHEN** create or update runs
- **THEN** the provider SHALL return an error diagnostic without calling the Put index template API

#### Scenario: Feature on supported cluster version

- **GIVEN** `data_stream_options` is configured
- **AND** the connected Elasticsearch server version is `9.1.0` or above
- **WHEN** create or update runs
- **THEN** the provider SHALL include `data_stream_options` in the API request normally

---

### Requirement: Create/update — expand `data_stream_options` into API request (REQ-034)

On create and update, when `template.data_stream_options` is configured, the provider SHALL construct a `DataStreamOptions` model from the Terraform configuration and include it in the `template` field of the Put index template API request body.

#### Scenario: `failure_store.enabled` round-trip on create

- **GIVEN** `failure_store.enabled = true` configured
- **WHEN** create runs and the template is read back
- **THEN** state SHALL contain `template.data_stream_options.failure_store.enabled = true`

#### Scenario: Update changes `enabled` value

- **GIVEN** an existing template with `failure_store.enabled = true`
- **WHEN** configuration changes `failure_store.enabled` to `false` and apply runs
- **THEN** the provider SHALL send `enabled: false` in the updated API request
- **AND** state SHALL reflect `failure_store.enabled = false` after the read-back

---

### Requirement: Read — flatten `data_stream_options` from API response (REQ-035)

On read, when the API response includes `data_stream_options` inside the `template` object, the provider SHALL populate `template.data_stream_options` in Terraform state, including the `failure_store.enabled` value and `failure_store.lifecycle.data_retention` if present.

When the API response does not include `data_stream_options` (or `data_stream_options` is null), the provider SHALL leave `template.data_stream_options` unset in state.

#### Scenario: Read-back with `data_stream_options` present

- **GIVEN** the API response includes `template.data_stream_options.failure_store.enabled = true`
- **WHEN** read runs
- **THEN** state SHALL contain `template.data_stream_options.failure_store.enabled = true`

#### Scenario: Read-back with `data_stream_options` absent

- **GIVEN** the API response does not include `template.data_stream_options`
- **WHEN** read runs
- **THEN** `template.data_stream_options` SHALL be unset in state

#### Scenario: Read-back with `lifecycle.data_retention` present

- **GIVEN** the API response includes `template.data_stream_options.failure_store.lifecycle.data_retention = "10d"`
- **WHEN** read runs
- **THEN** state SHALL contain `template.data_stream_options.failure_store.lifecycle.data_retention = "10d"`

---

### Requirement: Model — `DataStreamOptions` struct in `models.Template` (REQ-036)

The internal `models.Template` struct SHALL include a `DataStreamOptions` field typed as `*DataStreamOptions` and serialized as `"data_stream_options"` in JSON. The field SHALL be `omitempty` so that templates without `data_stream_options` serialize correctly. New structs `DataStreamOptions`, `FailureStoreOptions`, and `FailureStoreLifecycle` SHALL be added to `internal/models/models.go`.

The `DataStreamOptions` struct SHALL contain:
- `FailureStore *FailureStoreOptions json:"failure_store,omitempty"`

The `FailureStoreOptions` struct SHALL contain:
- `Enabled bool json:"enabled"`
- `Lifecycle *FailureStoreLifecycle json:"lifecycle,omitempty"`

The `FailureStoreLifecycle` struct SHALL contain:
- `DataRetention string json:"data_retention,omitempty"`

Adding these fields to the shared `models.Template` struct SHALL not affect the `elasticstack_elasticsearch_component_template` resource, because the `data_stream_options` field will never be populated by that resource and `omitempty` ensures the key is absent from component template API payloads.

#### Scenario: Component template payloads are not affected

- **GIVEN** a component template create or update operation where `data_stream_options` is not configured
- **WHEN** the `models.Template` struct is serialized to JSON
- **THEN** the JSON payload SHALL NOT include the `data_stream_options` key

#### Scenario: Index template payload includes `data_stream_options`

- **GIVEN** an index template create or update operation where `data_stream_options` is configured
- **WHEN** the `models.Template` struct is serialized to JSON
- **THEN** the JSON payload SHALL include `"data_stream_options": { "failure_store": { "enabled": true } }`

---

### Requirement: Acceptance tests — `data_stream_options` coverage (REQ-037)

Acceptance tests for `elasticstack_elasticsearch_index_template` SHALL include coverage for:

- Creating a template with `data_stream_options.failure_store.enabled = true` and verifying state after create.
- Updating the template to change `failure_store.enabled` and verifying state after update.
- Creating a template with `failure_store.lifecycle.data_retention` set and verifying state after create.
- Verifying that omitting `data_stream_options` produces no drift in plan after apply.

These tests SHALL only run against Elasticsearch >= 9.1.0 and SHALL be skipped or guarded appropriately when a lower version is detected.

A unit test (not an acceptance test) SHALL verify the version-gate logic for the error path: when `data_stream_options` is configured and the detected Elasticsearch version is below `9.1.0`, the provider function under test SHALL return an error diagnostic without invoking the Put index template API.

#### Scenario: Acceptance test create with failure store enabled

- **GIVEN** an acceptance test configuration with `failure_store.enabled = true`
- **WHEN** the test creates the template and refreshes state
- **THEN** the acceptance test SHALL assert `template.0.data_stream_options.0.failure_store.0.enabled` equals `true` in state

#### Scenario: Acceptance test update failure store enabled value

- **GIVEN** an acceptance test that first creates a template with `failure_store.enabled = true`
- **WHEN** the configuration is updated to `failure_store.enabled = false` and applied
- **THEN** the acceptance test SHALL assert that the state reflects `failure_store.enabled = false` after the update

#### Scenario: Acceptance test with data_retention

- **GIVEN** an acceptance test configuration with `failure_store.lifecycle.data_retention = "14d"`
- **WHEN** the test creates the template and refreshes state
- **THEN** the acceptance test SHALL assert `template.0.data_stream_options.0.failure_store.0.lifecycle.0.data_retention` equals `"14d"` in state

#### Scenario: Unit test — version-gate error path

- **GIVEN** a unit test that simulates `data_stream_options` configured with Elasticsearch version `9.0.0`
- **WHEN** the create or update function is invoked
- **THEN** the function SHALL return an error diagnostic containing the minimum version requirement
- **AND** the Put index template API SHALL NOT be called

