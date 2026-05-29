# `elasticstack_kibana_space` — Schema and Functional Requirements

Resource implementation: `internal/kibana/spaces/resource.go`

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

The resource SHALL use the provider's configured Kibana client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for all API calls of that instance.

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any API call runs
- THEN the provider-configured Kibana client SHALL be used

#### Scenario: Scoped Kibana connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any API call runs
- THEN the resource SHALL use the scoped Kibana client derived from that block

### Requirement: Version compatibility — solution field (REQ-010)

When the `solution` attribute is configured with a non-empty value, the resource SHALL verify that the Elasticsearch/Kibana server version is at least 8.16.0 before calling the Create or Update API. If the server version is below 8.16.0, the resource SHALL fail with an error indicating that the `solution` field requires version 8.16.0 or higher.

#### Scenario: solution on an older cluster

- GIVEN `solution` is set and the server version is below 8.16.0
- WHEN create or update runs
- THEN the resource SHALL fail with an error indicating the minimum required version

### Requirement: Create and update behavior (REQ-011–REQ-013)
On create, the resource SHALL call the Kibana Create Space API and then immediately read the space back to refresh state. On update, the resource SHALL call the Kibana Update Space API and then immediately read the space back to refresh state. For both create and update, the resource SHALL build the request body from the configured `space_id`, `name`, and any optional fields (`description`, `disabled_features`, `initials`, `color`, `image_url`, `solution`), omitting optional fields that are not set in configuration. The resource SHALL allow `disabled_features` to be configured when `solution` is `classic`, when `solution` is omitted, or when `solution` is unknown during plan-time validation, and SHALL reject `disabled_features` only when `solution` has a known non-`classic` value.

#### Scenario: Post-create refresh
- GIVEN a successful Create Space API response
- WHEN the resource finishes creating
- THEN it SHALL read the space and populate state

#### Scenario: disabled_features allowed for classic solution
- **WHEN** configuration sets `disabled_features` and `solution = "classic"`
- **THEN** plan-time validation SHALL accept the configuration

#### Scenario: disabled_features allowed when solution is omitted
- **WHEN** configuration sets `disabled_features` and does not set `solution`
- **THEN** plan-time validation SHALL accept the configuration

#### Scenario: disabled_features rejected for non-classic solution
- **WHEN** configuration sets `disabled_features` and `solution` has a known value other than `classic`
- **THEN** plan-time validation SHALL return a validation error

### Requirement: Read behavior (REQ-014–REQ-015)

When refreshing state, the resource SHALL call the Kibana Get Space API using the space id derived from `d.Id()`. If the API returns nil with no error (space not found), the resource SHALL remove itself from Terraform state by calling `d.SetId("")`. When the space is found, the resource SHALL map the API response fields (`space_id`, `name`, `description`, `disabled_features`, `initials`, `color`, `solution`) into state.

#### Scenario: Space removed from Kibana

- GIVEN refresh runs and the space no longer exists
- WHEN the API returns nil with no error
- THEN the resource SHALL be removed from state

### Requirement: Delete behavior (REQ-016)

When destroying, the resource SHALL derive the space id from state (extracting the resource portion from a composite id if needed). If the derived space id is `"default"`, the provider SHALL NOT call `DELETE /api/spaces/space/default`; instead, the provider SHALL remove the resource from Terraform state only and SHALL emit a warning-level log message to surface the skip to operators. For all other space ids, the provider SHALL call the Kibana Delete Space API with that identifier.

The Kibana API permanently rejects `DELETE /api/spaces/space/default` with HTTP 400 Bad Request. This is a hard platform invariant on all supported Kibana versions; encoding it directly in the provider is the correct approach.

#### Scenario: Destroy default space — skip API call and remove from state

- **GIVEN** a `elasticstack_kibana_space` resource with `space_id = "default"` is in Terraform state
- **WHEN** `terraform destroy` runs
- **THEN** the provider SHALL NOT call `DELETE /api/spaces/space/default`
- **AND** the provider SHALL remove the resource from Terraform state
- **AND** the provider SHALL emit a `tflog.Warn` with the message: `"default Kibana space cannot be deleted; removing from Terraform state only"`

#### Scenario: Destroy non-default space — normal API delete

- **GIVEN** a `elasticstack_kibana_space` resource with `space_id` set to any value other than `"default"`
- **WHEN** `terraform destroy` runs
- **THEN** the provider SHALL call `DELETE /api/spaces/space/{space_id}` as before

### Requirement: Create 409 Conflict diagnostic (REQ-CREATE-409)

When `POST /api/spaces/space` returns HTTP 409 Conflict, the provider SHALL return an error diagnostic that:
- names the space id from the request
- instructs the practitioner to import the existing space using `terraform import elasticstack_kibana_space.<NAME> <space_id>`

The provider SHALL NOT attempt to auto-fallback to `PUT` (auto-import) on 409. Explicit import is the required workflow.

#### Scenario: Create fails with 409 — actionable diagnostic returned

- **GIVEN** a `elasticstack_kibana_space` resource with `space_id = "default"` (or any id of an existing space)
- **WHEN** `terraform apply` runs and Kibana returns HTTP 409 Conflict for `POST /api/spaces/space`
- **THEN** the provider SHALL return an error diagnostic naming the conflicting space id
- **AND** the diagnostic SHALL include an import command of the form `terraform import elasticstack_kibana_space.<NAME> <space_id>`
- **AND** the provider SHALL NOT attempt a `PUT /api/spaces/space/{id}` auto-fallback

#### Scenario: Create fails with other errors — existing behavior unchanged

- **GIVEN** a `elasticstack_kibana_space` create request
- **WHEN** Kibana returns any HTTP status other than 200 or 409
- **THEN** the provider SHALL surface the error via the existing `HandleMutateTypedResponse` path (unchanged)

### Requirement: image_url not persisted in read (REQ-017)

The resource SHALL NOT attempt to set `image_url` in state during read, as this field is not returned by the Kibana Get Space API response.

#### Scenario: image_url absent from API response

- GIVEN `image_url` is configured and the space is read
- WHEN read state is populated
- THEN `image_url` SHALL NOT be overwritten from the API response (field is omitted from read mapping)

### Requirement: Kibana OpenAPI client for Spaces (REQ-018)

The `elasticstack_kibana_space` implementation SHALL perform Create Space, Get Space, Update Space, and Delete Space HTTP calls using the generated OpenAPI Kibana client package (`generated/kbapi`) and helper functions colocated under `internal/clients/kibanaoapi` for Spaces. The implementation SHALL NOT use `github.com/disaster37/go-kibana-rest`.

#### Scenario: Mutations use kbapi transport

- **GIVEN** a create, update, read, or delete operation for the resource
- **WHEN** the provider issues the corresponding Kibana Spaces HTTP request
- **THEN** the request SHALL be executed through the kbapi client configured for the effective Kibana connection (provider default or `kibana_connection` scoped client)

### Requirement: Typed space models for provider mapping (REQ-019)

The kbapi types used to decode and encode space request and response JSON SHALL include the following logical fields for provider purposes: `id`, `name`, `description`, `disabledFeatures`, `initials`, `color`, `imageUrl`, `solution`, and `_reserved` (when returned by Kibana).

#### Scenario: Read maps equivalent Terraform attributes

- **GIVEN** a successful Get Space response identical to one handled by the pre-migration implementation
- **WHEN** read populates Terraform state from the typed kbapi response
- **THEN** the values stored for `space_id`, `name`, `description`, `disabled_features`, `initials`, `color`, and `solution` SHALL match the values the legacy client mapping would have produced for that JSON payload

### Requirement: solution version gate uses effective connection (REQ-020)

When the `solution` argument is set to a non-empty value, the resource SHALL evaluate the minimum Kibana/Stack version using the same effective Kibana connection as the kbapi Spaces calls before performing create or update, and SHALL fail with diagnostics when the version is below 8.16.0 as specified in existing requirements.

#### Scenario: Version check precedes kbapi mutation

- **GIVEN** `solution` is set and the resolved server version is below 8.16.0
- **WHEN** create or update runs
- **THEN** the resource SHALL return an error diagnostic without issuing a successful mutating Spaces API call that would persist an unsupported `solution`

### Requirement: Default-space acceptance test coverage (REQ-TEST-DEFAULT-SPACE)

The acceptance test suite SHALL include a test `TestAccResourceSpace_DefaultSpace` that verifies the complete import-update-destroy lifecycle for the default Kibana space without gating on a minimum stack version.

#### Scenario: Import default space, update, then destroy without error

- **GIVEN** a live Kibana instance with a default space
- **WHEN** `TestAccResourceSpace_DefaultSpace` runs
- **THEN** step 1 SHALL import the default space (`ResourceName: "elasticstack_kibana_space.default"`, `ImportState: true`, `ImportStateId: "default"`)
- **AND** step 2 SHALL apply a fixture config with only `space_id` and `name` (no `solution`) and assert `space_id == "default"` and `name == "Default"`
- **AND** the destroy step at the end of the test SHALL complete without error
- **AND** the test SHALL use no `CheckDestroy` (the default space persists after Terraform destroy)

