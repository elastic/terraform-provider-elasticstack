# `elasticstack_apm_source_map` — Schema and Functional Requirements

Resource implementation: `internal/apm/source_map`

## Purpose

Define schema and behavior for the APM source map resource: API usage, identity/import, connection, create, read, delete, and mapping between Terraform state and the Kibana APM source maps API.

## Schema

```hcl
resource "elasticstack_apm_source_map" "example" {
  id               = <computed, string>   # Fleet artifact ID returned by the upload API
  bundle_filepath  = <required, string>   # Absolute path of the final bundle in the web application
  service_name     = <required, string>   # Service name the source map applies to
  service_version  = <required, string>   # Service version the source map applies to
  sourcemap_json   = <optional, sensitive, string>  # Source map content as a JSON string; mutually exclusive with sourcemap_binary
  sourcemap_binary = <optional, sensitive, string>  # Source map content as a base64-encoded string; mutually exclusive with sourcemap_json
  space_id         = <optional, string>   # Kibana space ID; omit or set to "default" for the default space
  kibana_connection = <optional, block>   # Entity-local Kibana connection override
}
```

## ADDED Requirements

### Requirement: APM source map CRUD APIs (REQ-001)

The resource SHALL use `UploadSourceMapWithBodyWithResponse` to upload a new source map on create. The resource SHALL use `GetSourceMapsWithResponse` to read and locate the source map artifact by `id` on read. The resource SHALL use `DeleteSourceMapWithResponse` to delete the source map artifact on delete. There is no update endpoint; all write attributes SHALL use `RequireReplace` so any change to those attributes triggers destroy and recreate. When the Kibana API returns a non-success status for any create, read, or delete request, the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create

- GIVEN the Kibana API returns a non-success HTTP status for an upload request
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error and the resource SHALL not be stored in state

#### Scenario: API failure on delete

- GIVEN the Kibana API returns a non-success HTTP status for a delete request
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Kibana client usage (REQ-002)

The resource SHALL expose an optional entity-local `kibana_connection` block using the shared Plugin Framework Kibana connection schema helper. The resource SHALL obtain its Kibana OpenAPI client through typed scoped-client resolution from the provider-configured `*clients.ProviderClientFactory`. When `kibana_connection` is absent, the resource SHALL resolve the provider-default `*clients.KibanaScopedClient`. When `kibana_connection` is configured, the resource SHALL resolve a `*clients.KibanaScopedClient` rebuilt from that scoped connection. The resource SHALL use `GetKibanaOapiClient()` on the resolved `*clients.KibanaScopedClient` for all API operations. The resource SHALL use the Elastic API version `2023-10-31` in all API requests.

#### Scenario: Resource resolves typed Kibana client from provider defaults

- WHEN the resource is configured without `kibana_connection`
- THEN it SHALL resolve a `*clients.KibanaScopedClient` from the provider client factory and use that typed client for Kibana API operations

#### Scenario: Kibana client acquisition failure

- WHEN the provider cannot provide a typed Kibana client or Kibana OpenAPI client
- THEN Terraform diagnostics SHALL include an "Unable to get Kibana client" error

### Requirement: Identity and import (REQ-003)

The resource SHALL expose a computed `id` attribute that is populated from the `id` field of `APMUIUploadSourceMapsResponse` after a successful upload. The `id` SHALL be preserved across plan/apply cycles using `UseStateForUnknown`. The resource SHALL support import via `ImportStatePassthroughID`, accepting the Fleet artifact `id` as the import identifier. After import, a Read SHALL populate state from the API.

#### Scenario: ID captured from upload response

- GIVEN a successful upload
- WHEN the provider processes the `APMUIUploadSourceMapsResponse`
- THEN `id` in state SHALL equal the `id` field from the response (non-empty string)

#### Scenario: Import passthrough

- GIVEN an import with a valid Fleet artifact `id`
- WHEN import completes
- THEN `id` SHALL be stored in state and a read SHALL be performed to populate all other attributes

### Requirement: Create — multipart upload (REQ-004)

On create, the resource SHALL read the plan and construct a multipart/form-data request with fields `bundle_filepath`, `service_name`, `service_version`, and `sourcemap` (file field). When `sourcemap_json` is set, its string value SHALL be used directly as the `sourcemap` file content. When `sourcemap_binary` is set, its value SHALL be decoded from base64 standard encoding and the resulting bytes SHALL be used as the `sourcemap` file content. After a successful upload, the resource SHALL extract `id` from the `APMUIUploadSourceMapsResponse`, store it in state, then perform a read to confirm the artifact is reachable. If the upload response contains a nil `id`, the resource SHALL return an error diagnostic.

#### Scenario: Create with JSON source map

- GIVEN `sourcemap_json` is set to a valid JSON source map string
- WHEN create runs
- THEN the upload request SHALL include the JSON string as the `sourcemap` file field and state SHALL have a non-empty `id`

#### Scenario: Create with binary source map

- GIVEN `sourcemap_binary` is set to a valid base64-encoded source map
- WHEN create runs
- THEN the upload request SHALL include the decoded bytes as the `sourcemap` file field and state SHALL have a non-empty `id`

#### Scenario: Nil id in upload response

- GIVEN the API returns a 200 response with a nil `id` field in `APMUIUploadSourceMapsResponse`
- WHEN the provider processes the response
- THEN Terraform diagnostics SHALL include an error about the unexpected nil `id`

### Requirement: Read — paginated list search (REQ-005)

On read, the resource SHALL call `GetSourceMapsWithResponse` iterating pages until the artifact with `id` matching the state value is found or all pages are exhausted. The resource SHALL use `page` and `perPage` parameters to paginate. If the artifact is found, the resource SHALL refresh from the read response every non-sensitive attribute that the API returns for the matching artifact, including `id`, `bundle_filepath`, `service_name`, and `service_version`. The resource SHALL NOT attempt to reconstruct or refresh `sourcemap_json` or `sourcemap_binary` from the read response because the API does not return the original uploaded source map content. If no artifact matches the state `id`, the resource SHALL remove itself from state without returning an error.

#### Scenario: Artifact found on read

- GIVEN a source map with a known `id` exists in Kibana
- WHEN read runs
- THEN the resource SHALL remain in state with `id`, `bundle_filepath`, `service_name`, and `service_version` set from the matching Kibana artifact

#### Scenario: Import recovers remote metadata

- GIVEN a resource is imported with only an `id` in state
- WHEN read runs and finds the matching source map in Kibana
- THEN the resource SHALL populate `bundle_filepath`, `service_name`, and `service_version` from the read response
- AND the resource SHALL NOT populate `sourcemap_json` or `sourcemap_binary`

#### Scenario: Artifact not found removes from state

- GIVEN no source map artifact matches the state `id`
- WHEN read runs
- THEN the resource SHALL be removed from state without error

#### Scenario: API error on read

- GIVEN the Kibana API returns a non-success HTTP status for the list request
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Delete (REQ-006)

On delete, the resource SHALL call `DeleteSourceMapWithResponse` with the state `id` as the path parameter.

#### Scenario: Delete by id

- GIVEN a state `id` of a valid Fleet artifact
- WHEN delete runs
- THEN `DELETE /api/apm/sourcemaps/{id}` SHALL be called with that `id`

### Requirement: Validation — exactly one source map input (REQ-007)

Exactly one of `sourcemap_json` or `sourcemap_binary` SHALL be set. Setting both or neither SHALL be an invalid configuration. The provider SHALL enforce this constraint at plan time using `ExactlyOneOf` (or equivalent) on both attributes.

#### Scenario: Neither source map attribute set

- GIVEN a configuration where neither `sourcemap_json` nor `sourcemap_binary` is set
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic

#### Scenario: Both source map attributes set

- GIVEN a configuration where both `sourcemap_json` and `sourcemap_binary` are set
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic

### Requirement: Space-aware API operations (REQ-010)

The resource SHALL accept an optional `space_id` attribute (string, `RequireReplace`). All API paths — `POST /api/apm/sourcemaps`, `GET /api/apm/sourcemaps`, and `DELETE /api/apm/sourcemaps/{id}` — SHALL be constructed using `kibanautil.BuildSpaceAwarePath(spaceID, basePath)`. When `space_id` is empty or `"default"`, the path SHALL remain unchanged (default Kibana space). When `space_id` is a non-default value, the path SHALL be prefixed with `/s/{space_id}`.

#### Scenario: Create in a non-default space

- GIVEN `space_id = "my-space"` is set in the configuration
- WHEN create runs
- THEN the upload request SHALL be sent to `POST /s/my-space/api/apm/sourcemaps`

#### Scenario: Read in a non-default space

- GIVEN `space_id = "my-space"` is set in state
- WHEN read runs
- THEN the list request SHALL be sent to `GET /s/my-space/api/apm/sourcemaps`

#### Scenario: Delete in a non-default space

- GIVEN `space_id = "my-space"` is set in state
- WHEN delete runs
- THEN the delete request SHALL be sent to `DELETE /s/my-space/api/apm/sourcemaps/{id}`

#### Scenario: Default space omits space prefix

- GIVEN `space_id` is not set or is `"default"`
- WHEN any API operation runs
- THEN the request path SHALL NOT include the `/s/{space_id}` prefix

### Requirement: RequireReplace on write attributes (REQ-008)

The attributes `bundle_filepath`, `service_name`, `service_version`, `sourcemap_json`, `sourcemap_binary`, and `space_id` SHALL each use the `RequireReplace` plan modifier so that any change to these attributes triggers a destroy-then-create cycle.

#### Scenario: Change in service_version triggers replacement

- GIVEN an existing resource with `service_version = "1.0.0"`
- WHEN the configuration changes `service_version` to `"1.1.0"`
- THEN the Terraform plan SHALL show a replacement (destroy + create), not an in-place update

### Requirement: Acceptance tests (REQ-009)

The acceptance test suite SHALL include:

1. A test that creates a source map using `sourcemap_json` with a minimal but valid source map JSON, asserts `id` is non-empty after apply, and confirms the resource is destroyed cleanly on `terraform destroy`.
2. A test that creates a source map using `sourcemap_binary` (base64-encoded minimal source map), asserts `id` is non-empty after apply.
3. A test that imports an existing source map artifact by `id` and asserts the resource is in state after import.

#### Scenario: Create with sourcemap_json acceptance test

- GIVEN a valid Kibana environment
- WHEN `TestAccResourceApmSourceMap_json` runs
- THEN the resource is created, `id` is populated in state, and is destroyed without error

#### Scenario: Import acceptance test

- GIVEN an existing source map `id` in state
- WHEN `TestAccResourceApmSourceMap_import` performs an import step
- THEN state is re-populated via Read without error
