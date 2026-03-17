# `elasticstack_elasticsearch_component_template` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/component_template.go`

## Schema

```hcl
resource "elasticstack_elasticsearch_component_template" "example" {
  id   = <computed, string> # internal identifier: <cluster_uuid>/<template_name>
  name = <required, string> # force new

  metadata = <optional, json string>
  version  = <optional, int>

  # Deprecated: resource-level Elasticsearch connection override
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

  template { # required, list(block), max 1
    mappings = <optional, json object string>
    settings = <optional, json object string>

    alias { # optional, set(block)
      name           = <required, string>
      filter         = <optional, json string>
      index_routing  = <optional, string>
      is_hidden      = <optional, bool>   # default false
      is_write_index = <optional, bool>   # default false
      routing        = <optional, string>
      search_routing = <optional, string>
    }
  }
}
```

## Requirements

- **[REQ-001] (API)**: The resource shall use the Elasticsearch Put component template API to create and update component templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-component-template.html)).
- **[REQ-002] (API)**: The resource shall use the Elasticsearch Get component template API to read component templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-component-template.html)).
- **[REQ-003] (API)**: The resource shall use the Elasticsearch Delete component template API to delete component templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-delete-component-template.html)).
- **[REQ-004] (API)**: When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than not found on read), the resource shall surface the API error to Terraform diagnostics.
- **[REQ-005] (Identity)**: The resource shall expose a computed `id` in the format `<cluster_uuid>/<template_name>`.
- **[REQ-006] (Identity)**: During create and update, the resource shall compute `id` from the current cluster UUID and configured `name`.
- **[REQ-007] (Import)**: The resource shall support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state.
- **[REQ-008] (Import)**: For imported or stored `id` values, read/delete operations shall require the format `<cluster_uuid>/<resource identifier>` and shall return an error diagnostic when the format is invalid.
- **[REQ-009] (Lifecycle)**: Changing `name` shall require replacement (`ForceNew`).
- **[REQ-010] (Connection)**: By default, the resource shall use the provider-level Elasticsearch client.
- **[REQ-011] (Connection)**: When `elasticsearch_connection` is configured, the resource shall construct and use a resource-scoped Elasticsearch client for all API calls.
- **[REQ-012] (Create/Update)**: On create/update, the resource shall construct a `models.ComponentTemplate` request body from Terraform state and submit it with the Put component template API.
- **[REQ-013] (Create/Update)**: After a successful Put request, the resource shall set `id` and perform a read to refresh state.
- **[REQ-014] (Read)**: On read, the resource shall parse `id`, fetch the component template by name, and remove the resource from state when the template is not found.
- **[REQ-015] (Read)**: If the Get component template API returns a result count other than exactly one template, the read path shall return an error diagnostic.
- **[REQ-016] (Delete)**: On delete, the resource shall parse `id` and delete the template identified by the parsed resource identifier.
- **[REQ-017] (Mapping)**: `metadata` shall be validated as JSON by schema and parsed as JSON during create/update; if parsing fails, the resource shall return an error diagnostic and shall not call the Put API.
- **[REQ-018] (Mapping)**: `template.mappings` and `template.settings` shall be validated as JSON objects by schema and parsed into objects during create/update.
- **[REQ-019] (Mapping)**: `template.alias.filter` shall be validated as JSON by schema and parsed into an object when non-empty during create/update.
- **[REQ-020] (Mapping)**: `template.alias` shall be mapped as a set keyed by alias name in API payload/state conversion.
- **[REQ-021] (Mapping)**: Alias routing and flag fields (`index_routing`, `is_hidden`, `is_write_index`, `routing`, `search_routing`) shall be copied directly between Terraform values and API model fields.
- **[REQ-022] (State)**: On read, the resource shall set `name` and `version` from the API response.
- **[REQ-023] (State)**: On read, when API `metadata` is present, it shall be serialized into a JSON string and stored in state.
- **[REQ-024] (State)**: On read, when API `template` is present, it shall be flattened into `template` state, including aliases, mappings, and settings.
- **[REQ-025] (State)**: User-defined alias `routing` shall be preserved during read/refresh, because this field may be omitted by the API response and therefore shall not be overwritten from response data.
