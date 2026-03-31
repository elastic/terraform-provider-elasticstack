# `elasticstack_kibana_space` — Schema and Functional Requirements

Resource implementation: `internal/kibana/space.go`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_space` resource, including Kibana Spaces API usage, identity, import, connection handling, lifecycle, version compatibility, and state mapping.

## Schema

```hcl
resource "elasticstack_kibana_space" "example" {
  id    = <computed, string> # internal identifier: the space_id

  space_id = <required, string> # force new; must match /^[a-z0-9_-]+$/
  name     = <required, string>

  description       = <optional, string>
  disabled_features = <optional, computed, set(string)>
  initials          = <optional, computed, string> # 1–2 characters; auto-generated when absent
  color             = <optional, computed, string> # auto-generated when absent
  image_url         = <optional, string>           # must be a data-URL encoded image (data:image/...)
  solution          = <optional, computed, string> # one of: security, oblt, es, classic; requires Kibana >= 8.16.0 when set
}
```

## Requirements

### Requirement: Spaces CRUD APIs (REQ-001–REQ-003)

The resource SHALL use the Kibana Create Space API to create spaces. The resource SHALL use the Kibana Update Space API to update spaces. The resource SHALL use the Kibana Get Space API to read spaces. The resource SHALL use the Kibana Delete Space API to delete spaces. When the Kibana API returns a non-success response for create, update, read, or delete requests (other than not found on read), the resource SHALL surface the error to Terraform diagnostics.

#### Scenario: API failure surfaces to diagnostics

- GIVEN a non-success Kibana API response on create, update, read, or delete (except not found on read)
- WHEN the provider handles the response
- THEN the error SHALL appear in Terraform diagnostics

### Requirement: Identity (REQ-004–REQ-005)

The resource SHALL set the computed `id` equal to the Kibana `space_id` returned by the Create or Update API response. When performing read or delete operations, if the stored `id` follows the composite `<cluster_uuid>/<space_id>` format, the resource SHALL extract the `space_id` portion and use that as the Kibana space identifier; otherwise the full `id` SHALL be used directly.

#### Scenario: Composite id compatibility

- GIVEN the stored `id` is in `<cluster_uuid>/<space_id>` format
- WHEN read or delete runs
- THEN only the `space_id` portion SHALL be used for the Kibana API call

### Requirement: Import (REQ-006)

The resource SHALL support import via `ImportStatePassthroughContext`, persisting the supplied `id` value directly to state for subsequent read operations.

#### Scenario: Import passthrough

- GIVEN import is invoked with a valid space id or composite id
- WHEN import completes
- THEN the id SHALL be stored and the next read SHALL resolve the space

### Requirement: Lifecycle — space_id change requires replacement (REQ-007)

When the `space_id` argument changes, the resource SHALL require replacement (destroy and recreate), not an in-place update.

#### Scenario: Renaming a space

- GIVEN a configuration change to `space_id`
- WHEN Terraform plans the change
- THEN the resource SHALL be replaced

### Requirement: Connection (REQ-008–REQ-009)

The resource SHALL use the provider's configured Kibana client by default. When `NewAPIClientFromSDKResource` resolves a resource-level `elasticsearch_connection` block, the resource SHALL use that scoped client for all API calls of that instance.

#### Scenario: Provider-level Kibana client

- GIVEN no `elasticsearch_connection` block is present
- WHEN any API call runs
- THEN the provider-level Kibana client SHALL be used

### Requirement: Version compatibility — solution field (REQ-010)

When the `solution` attribute is configured with a non-empty value, the resource SHALL verify that the Elasticsearch/Kibana server version is at least 8.16.0 before calling the Create or Update API. If the server version is below 8.16.0, the resource SHALL fail with an error indicating that the `solution` field requires version 8.16.0 or higher.

#### Scenario: solution on an older cluster

- GIVEN `solution` is set and the server version is below 8.16.0
- WHEN create or update runs
- THEN the resource SHALL fail with an error indicating the minimum required version

### Requirement: Create and update behavior (REQ-011–REQ-013)

On create, the resource SHALL call the Kibana Create Space API and then immediately read the space back to refresh state. On update, the resource SHALL call the Kibana Update Space API and then immediately read the space back to refresh state. For both create and update, the resource SHALL build the request body from the configured `space_id`, `name`, and any optional fields (`description`, `disabled_features`, `initials`, `color`, `image_url`, `solution`), omitting optional fields that are not set in configuration.

#### Scenario: Post-create refresh

- GIVEN a successful Create Space API response
- WHEN the resource finishes creating
- THEN it SHALL read the space and populate state

### Requirement: Read behavior (REQ-014–REQ-015)

When refreshing state, the resource SHALL call the Kibana Get Space API using the space id derived from `d.Id()`. If the API returns nil with no error (space not found), the resource SHALL remove itself from Terraform state by calling `d.SetId("")`. When the space is found, the resource SHALL map the API response fields (`space_id`, `name`, `description`, `disabled_features`, `initials`, `color`, `solution`) into state.

#### Scenario: Space removed from Kibana

- GIVEN refresh runs and the space no longer exists
- WHEN the API returns nil with no error
- THEN the resource SHALL be removed from state

### Requirement: Delete behavior (REQ-016)

When destroying, the resource SHALL derive the space id from state (extracting the resource portion from a composite id if needed) and call the Kibana Delete Space API with that identifier.

#### Scenario: Destroy

- GIVEN destroy is requested
- WHEN delete runs
- THEN the provider SHALL call the Delete Space API with the space id from state

### Requirement: image_url not persisted in read (REQ-017)

The resource SHALL NOT attempt to set `image_url` in state during read, as this field is not returned by the Kibana Get Space API response.

#### Scenario: image_url absent from API response

- GIVEN `image_url` is configured and the space is read
- WHEN read state is populated
- THEN `image_url` SHALL NOT be overwritten from the API response (field is omitted from read mapping)
