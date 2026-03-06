# `elasticstack_elasticsearch_index_template` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/template.go`

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

  data_stream { # optional, list(block), max 1
    hidden               = <optional, bool> # default false
    allow_custom_routing = <optional, bool>
  }

  template { # optional, list(block), max 1
    mappings = <optional, json object string>
    settings = <optional, json object string>

    alias { # optional, set(block)
      name           = <required, string>
      filter         = <optional, json string> # default ""
      index_routing  = <optional, computed, string> # default ""
      is_hidden      = <optional, bool>   # default false
      is_write_index = <optional, bool>   # default false
      routing        = <optional, string> # default ""
      search_routing = <optional, computed, string> # default ""
    }

    lifecycle { # optional, set(block), max 1
      data_retention = <required, string>
    }
  }
}
```

## Requirements

- **[REQ-001] (API)**: The resource shall use the Elasticsearch Put index template API to create and update index templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-put-template.html)).
- **[REQ-002] (API)**: The resource shall use the Elasticsearch Get index template API to read index templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-template.html)).
- **[REQ-003] (API)**: The resource shall use the Elasticsearch Delete index template API to delete index templates ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-delete-template.html)).
- **[REQ-004] (API)**: When Elasticsearch returns a non-success status for create, update, read, or delete requests (other than not found on read), the resource shall surface the API error to Terraform diagnostics.
- **[REQ-005] (Identity)**: The resource shall expose a computed `id` in the format `<cluster_uuid>/<template_name>`.
- **[REQ-006] (Identity)**: During create and update, the resource shall compute `id` from the current cluster UUID and configured `name`.
- **[REQ-007] (Import)**: The resource shall support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` value directly to state.
- **[REQ-008] (Import)**: For imported or stored `id` values, read/delete operations shall require the format `<cluster_uuid>/<resource identifier>` and shall return an error diagnostic when the format is invalid.
- **[REQ-009] (Lifecycle)**: Changing `name` shall require replacement (`ForceNew`).
- **[REQ-010] (Connection)**: By default, the resource shall use the provider-level Elasticsearch client.
- **[REQ-011] (Connection)**: When `elasticsearch_connection` is configured, the resource shall construct and use a resource-scoped Elasticsearch client for all API calls.
- **[REQ-012] (Compatibility)**: When `ignore_missing_component_templates` is configured with one or more values, the resource shall require Elasticsearch version >= 8.7.0; otherwise it shall return an error diagnostic.
- **[REQ-013] (Create/Update)**: On create/update, the resource shall construct a `models.IndexTemplate` request body from Terraform state and submit it with the Put index template API.
- **[REQ-014] (Create/Update)**: After a successful Put request, the resource shall set `id` and perform a read to refresh state.
- **[REQ-015] (Read)**: On read, the resource shall parse `id`, fetch the index template by name, and remove the resource from state when the template is not found.
- **[REQ-016] (Read)**: If the Get index template API returns a result count other than exactly one template, the read path shall return an error diagnostic.
- **[REQ-017] (Delete)**: On delete, the resource shall parse `id` and delete the template identified by the parsed resource identifier.
- **[REQ-018] (Mapping)**: `metadata` shall be validated as JSON by schema and parsed as JSON during create/update; if parsing fails, the resource shall return an error diagnostic and shall not call the Put API.
- **[REQ-019] (Mapping)**: `template.mappings` and `template.settings` shall be validated as JSON objects by schema and parsed into objects during create/update.
- **[REQ-020] (Mapping)**: `template.alias.filter` shall be validated as JSON by schema and parsed into an object when non-empty during create/update.
- **[REQ-021] (Mapping)**: `template.alias` shall be mapped as a set keyed by alias name in API payload/state conversion.
- **[REQ-022] (Mapping)**: Alias routing and flag fields (`index_routing`, `is_hidden`, `is_write_index`, `routing`, `search_routing`) shall be copied directly between Terraform values and API model fields.
- **[REQ-023] (Mapping)**: `template.lifecycle` shall be mapped as at most one lifecycle object with `data_retention`.
- **[REQ-024] (Mapping)**: `data_stream.hidden` shall be sent when present.
- **[REQ-025] (Mapping)**: `data_stream.allow_custom_routing` shall be sent only when `true`, except that on updates it shall also be sent when prior state had `allow_custom_routing=true` (8.x workaround behavior).
- **[REQ-026] (State)**: On read, the resource shall set `name`, `composed_of`, `ignore_missing_component_templates`, `index_patterns`, `priority`, and `version` from the API response.
- **[REQ-027] (State)**: On read, when API `metadata` is present, it shall be serialized into a JSON string and stored in state.
- **[REQ-028] (State)**: On read, when API `template` is present, it shall be flattened into `template` state, including aliases, lifecycle, mappings, and settings.
- **[REQ-029] (State)**: On read, when API `data_stream` is present, it shall be flattened into a single `data_stream` block and include only fields present in API response.
- **[REQ-030] (State)**: User-defined alias `routing` shall be preserved during read/refresh, because this field may be omitted by the API response and therefore shall not be overwritten from response data.
- **[REQ-031] (State)**: Plan diffs to alias `search_routing` or `index_routing` attributes should be ignored if:
  - The attribute is unset in the current config
  - The state value for the attribute equals the current value of the `routing` attribute

## Acceptance criteria

### Alias routing
Given a Terraform configuration which configures only the `routing` attribute on an alias
Expect the `search_routing`, and `index_routing` attributes to be set in state matching the `routing` attribute.