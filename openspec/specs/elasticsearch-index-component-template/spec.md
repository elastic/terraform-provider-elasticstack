# `elasticstack_elasticsearch_component_template` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/componenttemplate/`

## Purpose

Define schema and behavior for the Elasticsearch component template resource: API usage, identity/import, connection handling, template mapping, read-time alias routing preservation, `template.data_stream_options` mapping, version gating, and state upgrade behavior.

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
      index_routing  = <optional+computed, string, default "">
      is_hidden      = <optional+computed, bool, default false>
      is_write_index = <optional+computed, bool, default false>
      routing        = <optional+computed, string, default "">
      search_routing = <optional+computed, string, default "">
    }
    data_stream_options {
      failure_store {
        enabled = <optional, bool>
        lifecycle {
          data_retention = <optional, string>
        }
      }
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

### Requirement: Lifecycle, connection, and framework implementation (REQ-009–REQ-012)

Changing `name` SHALL require replacement (`ForceNew`). By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls. The resource SHALL be implemented on the Terraform Plugin Framework and SHALL preserve the existing Terraform type name, schema shape, and import behavior while using the shared Elasticsearch entitycore envelope behavior defined in `openspec/specs/entitycore-resource-envelope/spec.md`.

#### Scenario: Resource-level client

- GIVEN `elasticsearch_connection` is set
- WHEN API calls run
- THEN they SHALL use the resource-scoped client

#### Scenario: Framework implementation preserves schema shape

- GIVEN the resource is served by the Plugin Framework implementation
- WHEN Terraform loads the resource schema
- THEN the schema SHALL continue to expose the same `elasticstack_elasticsearch_component_template` type name, `elasticsearch_connection` block, attributes, and import format

### Requirement: Create, update, read, delete (REQ-013–REQ-017)

On create/update, the resource SHALL construct a `models.ComponentTemplate` request body from Terraform state and submit it with the Put component template API. After a successful Put request, the resource SHALL set `id` and perform a read to refresh state. On read, the resource SHALL parse `id`, fetch the component template by name, and remove the resource from state when the template is not found. If the Get component template API returns a result count other than exactly one template, the read path SHALL return an error diagnostic. On delete, the resource SHALL parse `id` and delete the template identified by the parsed resource identifier.

#### Scenario: Singleton read result

- GIVEN Get returns zero or more than one template
- WHEN read runs
- THEN the provider SHALL return an error diagnostic

### Requirement: JSON and alias mapping (REQ-018–REQ-022)

`metadata` SHALL be validated as JSON by schema and parsed as JSON during create/update; if parsing fails, the resource SHALL return an error diagnostic and SHALL not call the Put API. `template.mappings` SHALL be validated as a JSON object by schema and `template.settings` SHALL use the provider's index settings custom type and both SHALL be parsed into objects during create/update. `template.alias.filter` SHALL be validated as a JSON object by schema and parsed into an object when non-empty during create/update. `template.alias` SHALL be mapped as a set keyed by alias name in API payload/state conversion. Alias routing and flag fields (`index_routing`, `is_hidden`, `is_write_index`, `routing`, `search_routing`) SHALL be copied directly between Terraform values and API model fields, with `index_routing`, `routing`, and `search_routing` defaulting to the empty string and `is_hidden` and `is_write_index` defaulting to `false` when omitted.

#### Scenario: Invalid mappings JSON

- GIVEN invalid `template.mappings` JSON
- WHEN create/update runs
- THEN the provider SHALL error before Put

### Requirement: Read state mapping (REQ-022–REQ-026)

On read, the resource SHALL set `name` and `version` from the API response. On read, when API `metadata` is present, it SHALL be serialized into a JSON string and stored in state. On read, when API `template` is present, it SHALL be flattened into `template` state, including aliases, mappings, and settings. User-defined alias `routing` SHALL be preserved during read/refresh, because this field may be omitted by the API response and therefore SHALL not be overwritten from response data.

For `template.mappings`, the resource SHALL treat Elasticsearch stringified scalar echoes as semantically equal to practitioner-authored scalar JSON values when the effective mapping value is otherwise unchanged. This equivalence SHALL apply to scalar leaf values such as booleans and numbers and SHALL suppress drift and post-apply consistency errors caused only by Elasticsearch returning a string form of the same scalar.

For `template.settings`, the resource SHALL treat Elasticsearch stringified scalar echoes as semantically equal to practitioner-authored scalar JSON values when the effective setting value is otherwise unchanged. This equivalence SHALL include JSON `null`, so a practitioner-authored `null` setting value SHALL compare equal to an Elasticsearch `"null"` string echo.

#### Scenario: Routing preserved on refresh

- GIVEN user-configured alias `routing` and API omits routing fields
- WHEN read runs
- THEN user `routing` SHALL not be lost from state

#### Scenario: Mappings boolean scalar echo is non-drifting

- GIVEN `template.mappings` is configured with a scalar boolean value
- AND Elasticsearch returns the same value as a JSON string scalar during refresh
- WHEN apply completes or a later refresh runs
- THEN the provider SHALL treat the mapping values as semantically equal
- AND Terraform SHALL NOT report a provider inconsistent-result error or follow-up drift for that difference alone

#### Scenario: Settings null scalar echo is non-drifting

- GIVEN `template.settings` is configured with a JSON `null` scalar value
- AND Elasticsearch returns the same value as the string scalar `"null"` during refresh
- WHEN apply completes or a later refresh runs
- THEN the provider SHALL treat the settings values as semantically equal
- AND Terraform SHALL NOT report a provider inconsistent-result error or follow-up drift for that difference alone

### Requirement: Data stream options support (REQ-027–REQ-031)

The resource SHALL support an optional `template.data_stream_options` block with nested `failure_store`
and `failure_store.lifecycle` blocks. During create and update, when `template.data_stream_options` is
configured, the provider SHALL map `failure_store.enabled` and `failure_store.lifecycle.data_retention`
into the Elasticsearch component template request body. During read, when Elasticsearch returns
`data_stream_options.failure_store`, the provider SHALL flatten those values back into Terraform state.
The `template.data_stream_options` block SHALL require `failure_store` when the block is present.

The `componenttemplate.Data` model SHALL implement the `entitycore.WithVersionRequirements` interface
via a `GetVersionRequirements()` method. That method SHALL delegate to
`datastreamoptions.GetVersionRequirements(d.Template)` and SHALL return a version requirement
(minimum ES 9.1.0) when `template.data_stream_options` is configured and non-null. When the template
object is null or unknown, or when `data_stream_options` is absent or null, the method SHALL return
`nil` (no requirements).

The entitycore resource envelope SHALL enforce these requirements automatically before every write
operation and during Read by calling `client.EnforceMinVersion` for each returned requirement.
`client.EnforceMinVersion` correctly handles Serverless clusters by short-circuiting to `true`
regardless of the reported server version. As a result, `data_stream_options` SHALL be usable on
Serverless clusters without error.

The `datastreamoptions` package SHALL be the single authoritative source for the `data_stream_options`
minimum version constant (`MinSupportedVersion = 9.1.0`) and the `GetVersionRequirements` helper.
The write callback (`writeComponentTemplate`) SHALL NOT contain a manual server version fetch or call
`EnforceMinServerVersion`; version enforcement is delegated to the envelope.

#### Scenario: Unsupported server version on stateful cluster

- GIVEN `template.data_stream_options` is configured
- AND the target Elasticsearch cluster is stateful and its version is below `9.1.0`
- WHEN create, update, or refresh runs
- THEN the provider SHALL return an error diagnostic
- AND it SHALL not call the Put API (on create/update)

#### Scenario: Serverless cluster is always supported

- GIVEN `template.data_stream_options` is configured
- AND the target Elasticsearch cluster flavour is `"serverless"`
- WHEN create, update, or refresh runs
- THEN the provider SHALL NOT return a version-gate error
- AND it SHALL include `data_stream_options` in the API request normally (on create/update)

#### Scenario: Read-time enforcement

- GIVEN `template.data_stream_options` is present in Terraform state
- AND the target Elasticsearch cluster is stateful and its version is below `9.1.0`
- WHEN `terraform refresh` runs
- THEN the provider SHALL return an error diagnostic (consistent with Write-time behavior)

### Requirement: State upgrade to schema version 1 (REQ-032–REQ-035)

The resource SHALL define schema version `1` and provide an upgrade path from version `0`. During state upgrade from version `0`, the provider SHALL collapse legacy list-shaped `template` blocks to the Plugin Framework object-or-null representation. During that upgrade, the provider SHALL ensure the migrated `template` object contains explicit keys for `alias`, `mappings`, `settings`, and `data_stream_options`, using null when absent. During that upgrade, the provider SHALL normalize legacy alias state by converting SDK-style duplicated `index_routing` and `search_routing` values into the Plugin Framework routing-only representation and by converting empty-string alias `filter` values to null.

#### Scenario: Upgrade legacy template state

- GIVEN version `0` state containing list-shaped `template` data and legacy alias routing values
- WHEN the provider upgrades state to schema version `1`
- THEN the provider SHALL collapse `template` to object-or-null form
- AND it SHALL preserve equivalent alias routing semantics without creating spurious diffs

