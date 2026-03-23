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
