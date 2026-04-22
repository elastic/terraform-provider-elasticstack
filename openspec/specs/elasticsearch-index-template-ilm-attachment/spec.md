# `elasticstack_elasticsearch_index_template_ilm_attachment` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/index/templateilmattachment`

## Purpose

Define schema and behavior for the Elasticsearch index template ILM attachment resource: attaches an ILM lifecycle policy to an index template by writing the `index.lifecycle.name` setting into the corresponding `@custom` component template while preserving all other existing template content.

## Schema

```hcl
resource "elasticstack_elasticsearch_index_template_ilm_attachment" "example" {
  id             = <computed, string> # internal identifier: <cluster_uuid>/<index_template>@custom
  index_template = <required, string> # force new
  lifecycle_name = <required, string>

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

### Requirement: Component template CRUD APIs (REQ-001–REQ-003)

The resource SHALL use the Elasticsearch Get Component Template API to read the `@custom` component template ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-component-template.html)), requesting flat settings (`flat_settings=true`). The resource SHALL use the Elasticsearch Put Component Template API to create and update the `@custom` component template ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-component-template-api.html)). On delete, the resource SHALL use the Put Component Template API to update the `@custom` component template with the ILM setting removed rather than deleting the component template itself. When Elasticsearch returns a non-success response (other than template-not-found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create

- GIVEN the Put Component Template API returns a non-success response
- WHEN create runs
- THEN Terraform diagnostics SHALL include the error

#### Scenario: Delete updates rather than deletes

- GIVEN an existing `@custom` component template
- WHEN delete runs
- THEN the resource SHALL call Put Component Template (not Delete Component Template) with the ILM setting removed

### Requirement: Component template naming (REQ-004)

The resource SHALL derive the target component template name by appending `@custom` to the value of `index_template`. For example, `index_template = "logs-system.syslog"` targets the component template `logs-system.syslog@custom`.

#### Scenario: Component template name derivation

- GIVEN `index_template = "logs-system.syslog"`
- WHEN the component template name is computed
- THEN the name SHALL be `"logs-system.syslog@custom"`

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<index_template>@custom`. During create, the resource SHALL compute `id` from the current cluster UUID and the derived component template name (`<index_template>@custom`).

#### Scenario: ID format

- GIVEN a successful create with `index_template = "logs-system.syslog"`
- WHEN the provider sets the ID
- THEN `id` SHALL be `<cluster_uuid>/logs-system.syslog@custom`

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import via `ImportStatePassthroughID`, persisting the supplied `id` directly to state. On the subsequent read triggered by import, the resource SHALL derive `index_template` from the component template name in `id` by stripping the `@custom` suffix. Read/delete operations SHALL require `id` to be in the format `<cluster_uuid>/<resource identifier>` and SHALL return an error diagnostic when the format is invalid.

#### Scenario: Import derives index_template

- GIVEN an import with id `<cluster_uuid>/logs-system.syslog@custom`
- WHEN read runs after import
- THEN `index_template` SHALL be set to `"logs-system.syslog"`

#### Scenario: Invalid id format on read

- GIVEN a stored `id` that does not contain exactly one `/`
- WHEN read runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Lifecycle (REQ-009)

Changing `index_template` SHALL require replacement of the resource (`RequiresReplace`). The computed `id` SHALL be preserved across plan/apply cycles using `UseStateForUnknown`.

#### Scenario: index_template change triggers replace

- GIVEN an existing ILM attachment
- WHEN `index_template` is changed in configuration
- THEN Terraform SHALL plan a replace (destroy + create)

### Requirement: Connection (REQ-010)

By default, the resource SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls (create, read, update, delete).

#### Scenario: Resource-level client

- GIVEN `elasticsearch_connection` is set
- WHEN API calls run
- THEN they SHALL use the resource-scoped client

### Requirement: Compatibility — minimum Elasticsearch version (REQ-011)

On create and update, the resource SHALL check the Elasticsearch server version. If the server version is less than 8.2.0, the resource SHALL return an "Unsupported Elasticsearch Version" error diagnostic and SHALL NOT proceed to call the Put Component Template API.

#### Scenario: Version below minimum

- GIVEN Elasticsearch version 8.1.x
- WHEN create or update runs
- THEN the provider SHALL return an "Unsupported Elasticsearch Version" error and not call the API

#### Scenario: Version at minimum

- GIVEN Elasticsearch version 8.2.0 or later
- WHEN create or update runs
- THEN the provider SHALL proceed normally

### Requirement: Create (REQ-012–REQ-015)

On create, the resource SHALL read any existing `@custom` component template to preserve its settings, mappings, aliases, and metadata. The resource SHALL merge the `index.lifecycle.name` setting (flat form) into the template's existing settings, overwriting any previous ILM value. The resource SHALL write the updated component template via the Put Component Template API. After a successful Put, the resource SHALL perform a read to confirm the ILM setting is present; if the component template is not found or the ILM setting is absent after create, the resource SHALL return an error diagnostic.

#### Scenario: Existing content preserved

- GIVEN an existing `@custom` component template with custom mappings
- WHEN create runs to attach an ILM policy
- THEN the existing mappings SHALL be preserved and only `index.lifecycle.name` SHALL be added to settings

#### Scenario: Template not found after create

- GIVEN the Put Component Template API succeeds
- WHEN the subsequent read finds the template absent or the ILM setting missing
- THEN the resource SHALL return an error diagnostic

### Requirement: Update (REQ-016–REQ-018)

On update, the resource SHALL read the existing `@custom` component template to preserve all other content. The resource SHALL replace the `index.lifecycle.name` setting with the new value from the plan. The resource SHALL write the updated component template via the Put Component Template API. After a successful Put, the resource SHALL perform a read to confirm the ILM setting is present; if the component template is not found or the ILM setting is absent after update, the resource SHALL return an error diagnostic.

#### Scenario: ILM policy updated in place

- GIVEN an existing ILM attachment with policy "policy-a"
- WHEN `lifecycle_name` is changed to "policy-b"
- THEN the resource SHALL call Put Component Template with `index.lifecycle.name = "policy-b"` without replacing the resource

### Requirement: Read (REQ-019–REQ-021)

On read, the resource SHALL parse `id` to obtain the component template name. If `index_template` is not yet known (e.g. during import), the resource SHALL derive it by stripping the `@custom` suffix from the component template name. The resource SHALL call the Get Component Template API. If the component template is not found, or if the `index.lifecycle.name` flat setting is absent or empty, the resource SHALL remove itself from state without an error. When the ILM setting is present, the resource SHALL update `lifecycle_name` in state with the value from the API response.

#### Scenario: ILM setting absent removes from state

- GIVEN a `@custom` component template that exists but has no `index.lifecycle.name` setting
- WHEN read runs
- THEN the resource SHALL be removed from state

#### Scenario: lifecycle_name updated from API

- GIVEN the component template contains `index.lifecycle.name = "my-policy"`
- WHEN read runs
- THEN `lifecycle_name` in state SHALL be `"my-policy"`

### Requirement: Delete (REQ-022–REQ-024)

On delete, the resource SHALL parse `id` to obtain the component template name. The resource SHALL read the existing `@custom` component template. If the component template does not exist, the resource SHALL treat it as already deleted and complete successfully without error. If the component template exists, the resource SHALL remove the `index.lifecycle.name` key from the settings map (flat form) and SHALL write the updated template via the Put Component Template API. After removing the ILM setting, if the settings map becomes empty, the resource SHALL set it to nil (rather than an empty map) before writing.

#### Scenario: Template already absent on delete

- GIVEN the `@custom` component template does not exist
- WHEN delete runs
- THEN the resource SHALL complete successfully without calling Put Component Template

#### Scenario: Empty settings pruned

- GIVEN the only setting in the template is `index.lifecycle.name`
- WHEN delete runs and removes the ILM setting
- THEN the resource SHALL write the template with a nil settings map

### Requirement: Mapping — ILM setting key (REQ-025)

The resource SHALL read and write the ILM policy using the flat key `index.lifecycle.name` in the component template's settings map. The Get Component Template API SHALL be called with `flat_settings=true` to ensure the key is returned in flat form.

#### Scenario: Flat settings key

- GIVEN a component template with ILM configured
- WHEN Get Component Template is called
- THEN the request SHALL include `flat_settings=true` and the ILM value SHALL be read from the key `index.lifecycle.name`
