# `elasticstack_kibana_synthetics_parameter` — Schema and Functional Requirements

Resource implementation: `internal/kibana/synthetics/parameter`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_synthetics_parameter` resource, which manages Kibana Synthetics parameters (key/value pairs used in synthetic monitor configurations). The resource covers the Synthetics Parameters API, composite identity and import, provider-level Kibana client usage, read-after-write on create and update, and the mapping of `share_across_spaces` to Kibana namespaces.

## Schema

```hcl
resource "elasticstack_kibana_synthetics_parameter" "example" {
  id                  = <computed, string>          # Kibana-generated id; UseStateForUnknown; RequiresReplace
  key                 = <required, string>          # Parameter key; UseStateForUnknown
  value               = <required, string, sensitive> # Parameter value; UseStateForUnknown
  description         = <optional, computed, string> # Default ""; UseStateForUnknown
  tags                = <optional, computed, list(string)> # Default []; UseStateForUnknown
  share_across_spaces = <optional, computed, bool>  # Default false; UseStateForUnknown; RequiresReplace
}
```

Notes:

- The resource uses the provider-level Kibana OpenAPI client for create, read, and update; and the provider-level Kibana legacy client for delete.
- `share_across_spaces` is not sent to the API on update calls; it is only sent on create.
- The `id` field has both `UseStateForUnknown` and `RequiresReplace` plan modifiers. Because `id` is computed and set by Kibana, any change that causes a new parameter to be created will also generate a new `id`.
- There is no schema version or state upgrade defined for this resource.

## Requirements

### Requirement: Synthetics Parameters API (REQ-001)

The resource SHALL manage Synthetics parameters through Kibana's Synthetics Parameters API: create via `POST /api/synthetics/params`, read via `GET /api/synthetics/params/{id}`, update via `PUT /api/synthetics/params/{id}`, and delete via the Kibana legacy synthetics client.

#### Scenario: CRUD uses Synthetics Parameters APIs

- GIVEN a managed Synthetics parameter
- WHEN create, read, update, or delete runs
- THEN the provider SHALL use the corresponding Kibana Synthetics Parameters API operation

### Requirement: API and client error surfacing (REQ-002)

The resource SHALL fail with an error diagnostic when it cannot obtain a Kibana client (OpenAPI or legacy). Transport errors and unexpected API responses for create, read, update, and delete SHALL be surfaced as error diagnostics. On read, a 404 response SHALL cause the resource to be removed from state rather than returning an error.

#### Scenario: Missing Kibana client

- GIVEN the resource cannot obtain a Kibana client from provider configuration
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Read returns 404

- GIVEN a parameter that no longer exists in Kibana
- WHEN read calls the API and receives a 404 response
- THEN the provider SHALL remove the resource from state

#### Scenario: Create transport or API error

- GIVEN a create request
- WHEN the API returns a transport error or an unexpected response
- THEN the provider SHALL surface an error diagnostic

### Requirement: Identity and computed `id` (REQ-003)

The resource SHALL expose a computed `id` set from the parameter id returned by Kibana after a successful create. The `id` SHALL be preserved across reads using `UseStateForUnknown`. Because `id` also carries `RequiresReplace`, any operation that changes the parameter's Kibana identity SHALL trigger replacement.

#### Scenario: `id` set after create

- GIVEN a successful parameter create
- WHEN Kibana returns the new parameter id
- THEN the provider SHALL store that id in state

### Requirement: Import by bare `id` (REQ-004)

The resource SHALL support Terraform import using the Kibana parameter id as the import identifier, using `ImportStatePassthroughID`. On import, if the id contains a `/`, the provider SHALL parse it as a composite id in the format `<cluster_uuid>/<resource_id>` and use only the resource id segment for subsequent API calls. If the id contains no `/`, the full value SHALL be used as the parameter id.

#### Scenario: Import with bare id

- GIVEN an import id that does not contain `/`
- WHEN import runs
- THEN the provider SHALL use the full value as the parameter id for the subsequent read

#### Scenario: Import with composite id

- GIVEN an import id in the format `<cluster_uuid>/<resource_id>`
- WHEN import runs and read is performed
- THEN the provider SHALL extract `<resource_id>` and use it to call the Synthetics Parameters API

#### Scenario: Import with malformed composite id

- GIVEN an import id containing `/` but not in a valid composite id format
- WHEN import runs
- THEN the provider SHALL return an error diagnostic describing the required format

### Requirement: Provider-level Kibana client only (REQ-005)

The resource SHALL use the provider's configured Kibana clients (OpenAPI client for create, read, and update; legacy client for delete). The resource SHALL NOT support a resource-level connection override block.

#### Scenario: Standard provider connection

- GIVEN the provider is configured with Kibana access
- WHEN the resource performs CRUD
- THEN all API operations SHALL use the provider-level Kibana client

### Requirement: Read-after-write on create and update (REQ-006)

After a successful create or update API call, the resource SHALL perform a follow-up read of the parameter by id and SHALL use the read response to populate state. This is required because Kibana's create response omits the `value` field and Kibana's update response returns the old `value`.

#### Scenario: Create read-after-write

- GIVEN a successful POST to create a parameter
- WHEN the provider receives the create response
- THEN the provider SHALL call `GET /api/synthetics/params/{id}` and populate state from the GET response

#### Scenario: Update read-after-write

- GIVEN a successful PUT to update a parameter
- WHEN the provider receives the update response
- THEN the provider SHALL call `GET /api/synthetics/params/{id}` and populate state from the GET response

### Requirement: `share_across_spaces` create-only field (REQ-007)

The `share_across_spaces` attribute SHALL be sent to Kibana only on create requests. On update, `share_across_spaces` SHALL be omitted from the request body. Because this field is `RequiresReplace`, any change to it in configuration SHALL trigger resource replacement rather than an in-place update.

#### Scenario: `share_across_spaces` omitted on update

- GIVEN a parameter update that does not change `share_across_spaces`
- WHEN the update request is built
- THEN the request body SHALL NOT include `share_across_spaces`

#### Scenario: Replace on `share_across_spaces` change

- GIVEN an existing managed parameter with `share_across_spaces = false`
- WHEN configuration changes `share_across_spaces` to `true`
- THEN Terraform SHALL plan replacement for the resource

### Requirement: `share_across_spaces` to namespaces mapping (REQ-008)

When reading a parameter from Kibana, the resource SHALL map the API `namespaces` field to the `share_across_spaces` attribute. If `namespaces` equals `["*"]` (all spaces), `share_across_spaces` SHALL be set to `true`; otherwise it SHALL be set to `false`.

#### Scenario: All-spaces parameter maps to `share_across_spaces = true`

- GIVEN Kibana returns a parameter with `namespaces = ["*"]`
- WHEN the provider maps the response to state
- THEN `share_across_spaces` SHALL be `true`

#### Scenario: Single-space parameter maps to `share_across_spaces = false`

- GIVEN Kibana returns a parameter with `namespaces = ["default"]`
- WHEN the provider maps the response to state
- THEN `share_across_spaces` SHALL be `false`

### Requirement: Tags and description defaults (REQ-009)

When `tags` is not configured, the resource SHALL default it to an empty list and SHALL marshal it as an empty JSON array (not null) in API requests. When `description` is not configured, the resource SHALL default it to an empty string and SHALL include it in API requests.

#### Scenario: `tags` defaults to empty list

- GIVEN a parameter configured without `tags`
- WHEN Terraform plans the resource
- THEN `tags` SHALL default to `[]`

#### Scenario: `description` defaults to empty string

- GIVEN a parameter configured without `description`
- WHEN Terraform plans the resource
- THEN `description` SHALL default to `""`

### Requirement: Replacement on `id` change (REQ-010)

Because `id` carries `RequiresReplace`, any configuration or plan change that results in a new computed `id` value SHALL trigger replacement of the resource rather than an in-place update.

#### Scenario: Replace on `id` change

- GIVEN an existing managed parameter
- WHEN the planned `id` differs from the state `id`
- THEN Terraform SHALL plan replacement for the resource

## Traceability

| Area | Primary files |
|------|---------------|
| Schema and model | `internal/kibana/synthetics/parameter/schema.go` |
| Metadata / Configure / Import | `internal/kibana/synthetics/parameter/resource.go` |
| Create | `internal/kibana/synthetics/parameter/create.go` |
| Read | `internal/kibana/synthetics/parameter/read.go` |
| Update | `internal/kibana/synthetics/parameter/update.go` |
| Delete | `internal/kibana/synthetics/parameter/delete.go` |
| Shared client helpers | `internal/kibana/synthetics/api_client.go` |
| Shared utilities | `internal/kibana/synthetics/schema.go` |
