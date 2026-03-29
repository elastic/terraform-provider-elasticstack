# `elasticstack_kibana_default_data_view` — Schema and Functional Requirements

Resource implementation: `internal/kibana/defaultdataview`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_default_data_view` resource, which sets and manages the default data view for a Kibana space. The resource uses the Kibana Data Views API to set the default data view identifier and supports a `skip_delete` option to leave the existing default unchanged when the resource is destroyed.

## Schema

```hcl
resource "elasticstack_kibana_default_data_view" "example" {
  id           = <computed, string>         # set to space_id; UseStateForUnknown
  data_view_id = <optional, string>         # minimum length 1; Kibana data view ID to set as default
  force        = <optional, bool>           # overwrite existing default even if one already exists
  skip_delete  = <optional, computed, bool> # default false; when true, destroy does not unset the default
  space_id     = <optional, computed, string> # default "default"; RequiresReplace
}
```

Notes:

- The resource uses the provider-level Kibana OpenAPI client; there is no resource-level Kibana connection override block.
- The resource does not support Terraform import.
- The resource does not declare a custom state upgrader or custom plan modifier beyond the schema-level defaults listed above.

## Requirements

### Requirement: Kibana Set Default Data View API (REQ-001)

The resource SHALL manage the default data view setting through Kibana's Set Default Data View HTTP API ([Kibana data views API docs](https://www.elastic.co/guide/en/kibana/current/data-views-api.html)).

#### Scenario: Create and update use Set Default Data View API

- GIVEN a managed default data view resource
- WHEN create or update runs
- THEN the provider SHALL call the Kibana Set Default Data View API for the configured space

### Requirement: API and client error surfacing (REQ-002)

When the provider cannot obtain the Kibana OpenAPI client, create, read, and update operations SHALL return an error diagnostic. Transport errors and unexpected HTTP statuses from the Set Default Data View API and Get Default Data View API SHALL be surfaced as error diagnostics.

#### Scenario: Missing Kibana OpenAPI client

- GIVEN the resource cannot obtain a Kibana OpenAPI client from provider configuration
- WHEN any create, read, or update operation runs
- THEN the operation SHALL fail with an error diagnostic

### Requirement: Identity and `id` (REQ-003)

The resource SHALL store a computed `id` equal to the `space_id` value. The `id` is set during create and update after the read-back that follows the Set Default Data View API call.

#### Scenario: ID equals space_id

- GIVEN a successful create or update for space `"observability"`
- WHEN the provider records Terraform state
- THEN `id` SHALL equal `"observability"`

### Requirement: Space ID lifecycle and default (REQ-004)

If `space_id` is omitted, the resource SHALL default it to `"default"`. Changes to `space_id` SHALL require resource replacement rather than an in-place update.

#### Scenario: Default space

- GIVEN configuration that does not set `space_id`
- WHEN Terraform plans the resource
- THEN `space_id` SHALL default to `"default"`

#### Scenario: Replace on space change

- GIVEN an existing managed default data view resource
- WHEN `space_id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Provider-level Kibana client only (REQ-005)

The resource SHALL use the provider's configured Kibana OpenAPI client for all create, read, update, and delete operations. The resource SHALL NOT support a resource-local connection override in its schema or request path.

#### Scenario: Standard provider connection

- GIVEN the provider is configured with Kibana access
- WHEN the resource performs CRUD
- THEN all API operations SHALL use the provider-level Kibana OpenAPI client

### Requirement: Create and update flow with read-back (REQ-006)

On create and on update, the resource SHALL call the Set Default Data View API with the configured `data_view_id` and `force` values for the target `space_id`. After the API call succeeds, the resource SHALL perform a read via the Get Default Data View API and SHALL set state from the read result.

#### Scenario: Create sets default and reads back

- GIVEN a valid configuration with `data_view_id` and `space_id`
- WHEN create runs
- THEN the provider SHALL call Set Default Data View API, then read the current default via Get Default Data View API, and SHALL set Terraform state from the get response

### Requirement: Force flag (REQ-007)

The resource SHALL send the `force` field to the Set Default Data View API on create and update. When `force` is `true`, the API SHALL overwrite an existing default data view. When `force` is `false` or unset, the API MAY reject the request if a default data view already exists.

#### Scenario: Force overwrite

- GIVEN `force = true` and an existing default data view in the space
- WHEN create or update runs
- THEN the provider SHALL send `force: true` to the Set Default Data View API

### Requirement: Read behavior (REQ-008)

On read, the resource SHALL call the Kibana Get Default Data View API for the configured `space_id` and SHALL update `data_view_id` in state with the returned value (which may be null if no default is set). The resource SHALL set `id` to the `space_id` value.

#### Scenario: Read reflects current default

- GIVEN a resource recorded in Terraform state
- WHEN read runs
- THEN the provider SHALL call Get Default Data View API and SHALL update `data_view_id` in state with the current default

#### Scenario: No default data view set

- GIVEN no default data view is configured in the Kibana space
- WHEN read runs
- THEN the provider SHALL set `data_view_id` to null in state

### Requirement: Delete behavior and skip_delete (REQ-009)

On delete, if `skip_delete` is `true`, the resource SHALL NOT call the Kibana API and SHALL leave the existing default data view unchanged. If `skip_delete` is `false` (the default), the resource SHALL call the Set Default Data View API with a null `data_view_id` (and `force: true`) to unset the default data view.

#### Scenario: Delete unsets default when skip_delete is false

- GIVEN `skip_delete = false`
- WHEN destroy runs
- THEN the provider SHALL call Set Default Data View API with a null `data_view_id` and `force: true`

#### Scenario: Delete skipped when skip_delete is true

- GIVEN `skip_delete = true`
- WHEN destroy runs
- THEN the provider SHALL NOT call any Kibana API and SHALL leave the default data view unchanged

### Requirement: skip_delete default (REQ-010)

If `skip_delete` is omitted, the resource SHALL default it to `false`.

#### Scenario: Default skip_delete

- GIVEN configuration that does not set `skip_delete`
- WHEN Terraform plans the resource
- THEN `skip_delete` SHALL default to `false`

## Traceability

| Area | Primary files |
|------|---------------|
| Schema | `internal/kibana/defaultdataview/schema.go` |
| Metadata / Configure | `internal/kibana/defaultdataview/resource.go` |
| Model | `internal/kibana/defaultdataview/models.go` |
| Create / Update | `internal/kibana/defaultdataview/create.go`, `internal/kibana/defaultdataview/update.go` |
| Read | `internal/kibana/defaultdataview/read.go` |
| Delete | `internal/kibana/defaultdataview/delete.go` |
