# `elasticstack_elasticsearch_component_template` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/component_template.go`

## Purpose

Define schema and behavior for the Elasticsearch component template resource: API usage, identity/import, connection, mapping, and read-time alias routing preservation.

## Schema

```hcl
resource "elasticstack_elasticsearch_component_template" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<template_name>
  name = <required, string> # force new

  metadata = <optional, json string>
  version  = <optional, int>

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

  template {
    mappings = <optional, json object string>
    settings = <optional, json object string>
    alias {
      name           = <required, string>
      filter         = <optional, json string>
      index_routing  = <optional, string>
      is_hidden      = <optional, bool>
      is_write_index = <optional, bool>
      routing        = <optional, string>
      search_routing = <optional, string>
    }
  }
}
```

## Requirements

### Requirement: Component template CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Put component template API to create and update component templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-component-template.html)). The resource SHALL use the Elasticsearch Get component template API to read component templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-component-template.html)). The resource SHALL use the Elasticsearch Delete component template API to delete component templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-delete-component-template.html)). When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than not found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure

- GIVEN a non-success response (except 404 on read)
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Identity and import (REQ-005–REQ-008)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<template_name>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and configured `name`. The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state. For imported or stored `id` values, read/delete operations SHALL require the format `<cluster_uuid>/<resource identifier>` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Import passthrough

- GIVEN import with valid composite id
- WHEN import completes
- THEN the id SHALL be stored for subsequent operations

### Requirement: Lifecycle and connection (REQ-009–REQ-011)

Changing `name` SHALL require replacement (`ForceNew`). By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls.

#### Scenario: Resource-level client

- GIVEN `elasticsearch_connection` is set
- WHEN API calls run
- THEN they SHALL use the resource-scoped client

### Requirement: Create, update, read, delete (REQ-012–REQ-016)

On create/update, the resource SHALL construct a `models.ComponentTemplate` request body from Terraform state and submit it with the Put component template API. After a successful Put request, the resource SHALL set `id` and perform a read to refresh state. On read, the resource SHALL parse `id`, fetch the component template by name, and remove the resource from state when the template is not found. If the Get component template API returns a result count other than exactly one template, the read path SHALL return an error diagnostic. On delete, the resource SHALL parse `id` and delete the template identified by the parsed resource identifier.

#### Scenario: Singleton read result

- GIVEN Get returns zero or more than one template
- WHEN read runs
- THEN the provider SHALL return an error diagnostic

### Requirement: JSON and alias mapping (REQ-017–REQ-021)

`metadata` SHALL be validated as JSON by schema and parsed as JSON during create/update; if parsing fails, the resource SHALL return an error diagnostic and SHALL not call the Put API. `template.mappings` and `template.settings` SHALL be validated as JSON objects by schema and parsed into objects during create/update. `template.alias.filter` SHALL be validated as JSON by schema and parsed into an object when non-empty during create/update. `template.alias` SHALL be mapped as a set keyed by alias name in API payload/state conversion. Alias routing and flag fields (`index_routing`, `is_hidden`, `is_write_index`, `routing`, `search_routing`) SHALL be copied directly between Terraform values and API model fields.

#### Scenario: Invalid mappings JSON

- GIVEN invalid `template.mappings` JSON
- WHEN create/update runs
- THEN the provider SHALL error before Put

### Requirement: Read state mapping (REQ-022–REQ-025)

On read, the resource SHALL set `name` and `version` from the API response. On read, when API `metadata` is present, it SHALL be serialized into a JSON string and stored in state. On read, when API `template` is present, it SHALL be flattened into `template` state, including aliases, mappings, and settings. User-defined alias `routing` SHALL be preserved during read/refresh, because this field may be omitted by the API response and therefore SHALL not be overwritten from response data.

#### Scenario: Routing preserved on refresh

- GIVEN user-configured alias `routing` and API omits routing fields
- WHEN read runs
- THEN user `routing` SHALL not be lost from state
