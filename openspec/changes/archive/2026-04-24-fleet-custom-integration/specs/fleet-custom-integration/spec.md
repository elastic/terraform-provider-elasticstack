# `elasticstack_fleet_custom_integration` — Schema and Functional Requirements

Resource implementation: `internal/fleet/customintegration`

## Purpose

Upload and manage a locally-built Fleet custom integration package archive via the Kibana EPM binary upload API (`POST /api/fleet/epm/packages`). Unlike `elasticstack_fleet_integration`, which installs packages from the Elastic package registry by name and version, this resource accepts a local `.zip` or `.tar.gz` archive and manages the full lifecycle: upload on create, verify on read, re-upload on content change, and uninstall on destroy.

## Schema

```hcl
resource "elasticstack_fleet_custom_integration" "example" {
  id              = <computed, string>   # schemautil.StringToHash(package_name + package_version)
  package_path    = <required, string>   # path to local .zip or .tar.gz archive
  package_name    = <computed, string>   # extracted from upload response
  package_version = <computed, string>   # resolved via Fleet post-upload verification APIs
  checksum        = <computed, string>   # SHA256 hex digest of the file at package_path

  ignore_mapping_update_errors = <optional, bool>
  skip_data_stream_rollover    = <optional, bool>
  skip_destroy                 = <optional, bool>
  space_id                     = <optional, string>

  timeouts = <optional, block>           # operation timeouts; default 20m for create and update

  kibana_connection = <optional, block>  # overrides provider-level Kibana connection
}
```

## ADDED Requirements

### Requirement: Resource exists
The provider SHALL expose a `elasticstack_fleet_custom_integration` managed resource that uploads a locally-built Fleet integration package archive to Kibana Fleet via the EPM binary upload API and manages its full lifecycle.

#### Scenario: Resource appears in provider schema
- **WHEN** a practitioner runs `terraform providers schema`
- **THEN** `elasticstack_fleet_custom_integration` is listed as a managed resource type

### Requirement: Create uploads package archive
When the resource is created, the provider SHALL read the file at `package_path`, upload its binary contents to `POST /api/fleet/epm/packages` with the appropriate `Content-Type` header (`application/zip` for `.zip` files, `application/gzip` for `.gz` / `.tar.gz` files), and record the resulting package name and version in state. The provider SHALL extract `package_name` and `package_version` from the upload response when present, using the archive manifest only as a fallback when the upload response omits either field. After upload, the provider SHALL query the packages list API for the matching `package_name` and select the highest semver version among entries with status `installed`. If the packages list does not contain a matching installed entry and a package version was resolved from the upload response or archive manifest, the provider SHALL query the package info API for that exact `package_name` and `package_version` as a secondary verification step. If neither verification path finds a matching installed package, the provider SHALL return an error diagnostic rather than persisting the guessed version.

#### Scenario: Successful upload of a zip archive
- **WHEN** `package_path` points to a valid custom integration `.zip` file
- **THEN** the provider uploads the file contents with `Content-Type: application/zip`
- **THEN** `package_name` is set from `_meta.name` in the upload response
- **THEN** `package_version` is set from the installed package version confirmed by the Fleet verification APIs
- **THEN** `checksum` is set to the SHA256 hex digest of the uploaded file
- **THEN** `id` is set to a stable composite identifier derived from `package_name` and `package_version` (format: `<name>/<version>`)

#### Scenario: Successful upload of a gzip archive
- **WHEN** `package_path` points to a valid custom integration `.tar.gz` or `.gz` file
- **THEN** the provider uploads the file contents with `Content-Type: application/gzip`
- **THEN** all computed attributes (`package_name`, `package_version`, `checksum`, `id`) are populated

#### Scenario: Upload fails with a non-success response
- **WHEN** the Fleet API returns a non-success (non-2xx) status code
- **THEN** the provider returns an error diagnostic describing the failure
- **THEN** no state is written

#### Scenario: Package file does not exist
- **WHEN** `package_path` references a file that cannot be read
- **THEN** the provider returns an error diagnostic
- **THEN** no state is written

### Requirement: Read verifies installation
On each refresh, the provider SHALL verify the package is still installed using the `package_name` and `package_version` stored in state. The provider SHALL call the Fleet package info API first. If that API returns no package, the provider SHALL query the packages list API and look for an exact name/version match with installed status before concluding that the package is absent.

#### Scenario: Package is installed
- **WHEN** the Fleet package info API returns a package with status `installed`
- **THEN** the resource remains in state unchanged

#### Scenario: Package info API misses an installed package
- **WHEN** the Fleet package info API returns no package for the stored name/version
- **AND** the packages list API contains an exact match for that name/version with status `installed`
- **THEN** the resource remains in state unchanged

#### Scenario: Package is not found
- **WHEN** neither the package info API nor the packages list API confirms the stored package as installed
- **THEN** the provider removes the resource from state, signalling drift

#### Minimum version requirement: Kibana 8.2+
This resource requires Kibana 8.2.0 or later. On Create, Read, and Update, the provider
SHALL verify that the connected Kibana version meets this requirement and SHALL return an
error diagnostic if it does not. This gate exists because `GET /api/fleet/epm/packages/{name}/{version}`
does not support custom-uploaded packages on Kibana < 8.2 (7.17.x returns HTTP 400,
8.0.x–8.1.x returns HTTP 404 regardless of installation status), making drift detection
impossible on those versions.

### Requirement: Plan detects file content changes
The provider SHALL compute the SHA256 hash of the file at `package_path` during plan and compare it to the stored `checksum`. If they differ, `package_name`, `package_version`, and `checksum` SHALL be marked as unknown in the plan, indicating a pending change.

#### Scenario: File content has changed
- **WHEN** the file at `package_path` has different content from the last apply (different SHA256)
- **THEN** `terraform plan` shows `package_name`, `package_version`, and `checksum` as `(known after apply)`

#### Scenario: File content is unchanged
- **WHEN** the file at `package_path` has the same content as the last apply (same SHA256)
- **THEN** `terraform plan` shows no changes for this resource (assuming other attributes are also unchanged)

#### Scenario: File at package_path cannot be read during plan
- **WHEN** the file at `package_path` is missing or unreadable at plan time
- **THEN** the provider returns an error diagnostic during plan

### Requirement: Update re-uploads on content change
When an apply is triggered because the file content has changed, the provider SHALL re-upload the new file. If the resulting `package_name` or `package_version` differs from the values stored in state, the provider SHALL uninstall the old package after the upload succeeds (upload-first ordering is used because the final uploaded identity is not known until after the upload completes). When old-package cleanup is required, the provider SHALL wait until the replacement package becomes readable as installed before completing the update.

#### Scenario: File content changed, same package name
- **WHEN** the uploaded file has a different SHA256 but the resulting `package_name` matches the state value
- **THEN** the provider re-uploads the file and updates `package_version` and `checksum` in state

#### Scenario: File content changed, package name changed
- **WHEN** the uploaded file results in a `package_name` that differs from state
- **THEN** the provider uploads the new file
- **THEN** the provider uninstalls the old package (using the name and version from state)
- **THEN** `package_name`, `package_version`, `checksum`, and `id` are updated in state

#### Scenario: File content changed, package version changed
- **WHEN** the uploaded file results in the same `package_name` but a different `package_version`
- **THEN** the provider uploads the new file
- **THEN** the provider uninstalls the old package version from state
- **THEN** the provider waits until the replacement package is readable as installed before completing the update

#### Scenario: Query parameters changed only
- **WHEN** `ignore_mapping_update_errors` or `skip_data_stream_rollover` are changed and `checksum` is unchanged
- **THEN** the provider re-uploads the file with the updated query parameters

### Requirement: Space changes require replacement
The provider SHALL treat `space_id` as a replacement-only attribute. Changing `space_id` SHALL destroy the existing resource instance and create a new one rather than moving the installed package in place.

#### Scenario: space_id changed
- **WHEN** `space_id` changes between the prior state and the planned configuration
- **THEN** `terraform plan` marks the resource for replacement
- **THEN** the provider does not attempt to migrate the package between spaces during update

### Requirement: Delete uninstalls package
When the resource is destroyed and `skip_destroy` is `false` (default), the provider SHALL uninstall the package using the `package_name` and `package_version` stored in state.

#### Scenario: Destroy with skip_destroy false
- **WHEN** `terraform destroy` is run and `skip_destroy = false`
- **THEN** the provider calls the Fleet uninstall API for the package
- **THEN** the resource is removed from state

#### Scenario: Destroy with skip_destroy false but state is incomplete
- **WHEN** `terraform destroy` is run and `skip_destroy = false`
- **AND** either `package_name` or `package_version` is missing from state
- **THEN** the provider returns an error diagnostic instead of silently skipping uninstall

#### Scenario: Destroy with skip_destroy true
- **WHEN** `terraform destroy` is run and `skip_destroy = true`
- **THEN** the provider skips the uninstall API call
- **THEN** the resource is removed from state, leaving the package installed in Fleet

### Requirement: Space-aware operation
The resource SHALL support the `space_id` attribute and route all Fleet API calls through the appropriate Kibana space path when `space_id` is set.

#### Scenario: Upload with space_id set
- **WHEN** `space_id` is set to a non-default space identifier
- **THEN** all Fleet API calls (upload, read, delete) are routed through `/s/{space_id}/api/fleet/epm/packages`

#### Scenario: Upload without space_id
- **WHEN** `space_id` is not set or is set to the default space
- **THEN** Fleet API calls use the standard `/api/fleet/epm/packages` path

### Requirement: Optional upload query parameters
The resource SHALL support `ignore_mapping_update_errors` and `skip_data_stream_rollover` as optional boolean attributes that are forwarded as query parameters on the upload API call.

#### Scenario: ignore_mapping_update_errors set to true
- **WHEN** `ignore_mapping_update_errors = true`
- **THEN** the upload request includes `ignoreMappingUpdateErrors=true` as a query parameter

#### Scenario: skip_data_stream_rollover set to true
- **WHEN** `skip_data_stream_rollover = true`
- **THEN** the upload request includes `skipDataStreamRollover=true` as a query parameter

### Requirement: Connection
The resource SHALL use the provider-level Fleet client obtained from provider configuration by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the Fleet client derived from the scoped connection for all CRUD operations.

#### Scenario: Provider Fleet client used by default
- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** the resource SHALL obtain its Fleet client from the provider configuration

#### Scenario: Scoped Fleet client used when overridden
- **WHEN** `kibana_connection` is configured on the resource
- **THEN** the resource SHALL obtain its effective Fleet client from the scoped connection for all lifecycle operations

### Requirement: Operation timeouts
The resource SHALL expose a `timeouts` block allowing practitioners to override the default operation deadline for create and update. The default timeout for both create and update SHALL be 20 minutes. The configured timeout SHALL be applied as a context deadline that covers the full upload operation, including any retry delay incurred by a Kibana rate-limit (HTTP 429) response.

#### Scenario: Default timeout applies when timeouts block is absent
- **WHEN** no `timeouts` block is configured
- **THEN** create and update operations use a 20-minute deadline

#### Scenario: Custom create timeout is respected
- **WHEN** `timeouts { create = "5m" }` is configured
- **THEN** the create operation returns a timeout error if it does not complete within 5 minutes

#### Scenario: Custom update timeout is respected
- **WHEN** `timeouts { update = "5m" }` is configured
- **THEN** the update operation returns a timeout error if it does not complete within 5 minutes
