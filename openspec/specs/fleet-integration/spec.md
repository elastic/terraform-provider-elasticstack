# `elasticstack_fleet_integration` — Schema and Functional Requirements

Resource implementation: `internal/fleet/integration`
Data source implementation: `internal/fleet/integrationds`

## Purpose

Manage the installation and uninstallation of Fleet integration packages (resource), and look up the latest available version of a Fleet integration package without managing its lifecycle (data source). The resource uses the Kibana Fleet package management API to install or uninstall packages. The data source uses the same API to list available packages and return version information for the named package.

## Schema

### Resource

```hcl
resource "elasticstack_fleet_integration" "example" {
  id      = <computed, string>  # hash of name+version
  name    = <required, string>  # force new
  version = <required, string>

  force                        = <optional, bool>
  prerelease                   = <optional, bool>
  ignore_mapping_update_errors = <optional, bool>  # requires server >= 8.11.0
  skip_data_stream_rollover    = <optional, bool>  # requires server >= 8.11.0
  ignore_constraints           = <optional, bool>
  skip_destroy                 = <optional, bool>
  space_id                     = <optional+computed, string>  # force new
}
```

### Data Source

```hcl
data "elasticstack_fleet_integration" "example" {
  id         = <computed, string>  # hash of name
  name       = <required, string>
  prerelease = <optional, bool>

  version = <computed, string>
}
```

## Requirements

### Requirement: Fleet package install API (REQ-001–REQ-003)

The resource SHALL use the Kibana Fleet install package API to install integration packages on create and update. The resource SHALL use the Kibana Fleet uninstall package API to remove integration packages on delete. When the Fleet API returns a non-success response for any install or uninstall operation, the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API error on install

- GIVEN a failing Fleet API response during package installation
- WHEN create or update runs
- THEN diagnostics SHALL include the API error and the operation SHALL be aborted

### Requirement: Fleet package read API (REQ-004–REQ-005)

The resource SHALL use the Kibana Fleet get package API to refresh state on read, supplying the `name` and `version` from state. The data source SHALL use the Kibana Fleet list packages API to retrieve available packages, filtered by the `prerelease` parameter.

#### Scenario: Package not found on resource read

- GIVEN the integration package is not present (status not "installed") or nil in the Fleet API response
- WHEN read runs on the resource
- THEN the resource SHALL be removed from Terraform state

#### Scenario: Data source package lookup

- GIVEN a valid `name` and optional `prerelease` flag
- WHEN read runs on the data source
- THEN the data source SHALL set `version` to the version returned by the Fleet list packages API for the matching package name, or null if not found

### Requirement: Resource identity (REQ-006)

The resource SHALL expose a computed `id` attribute. The `id` SHALL be computed as a hash of `name` concatenated with `version` using `schemautil.StringToHash`. The data source SHALL expose a computed `id` computed as a hash of `name`.

#### Scenario: ID stability

- GIVEN a specific `name` and `version`
- WHEN create completes
- THEN `id` SHALL equal the hash of `name + version`

### Requirement: Lifecycle — name and space_id require replacement (REQ-007)

When the `name` attribute changes, the resource SHALL require replacement. When the `space_id` attribute changes, the resource SHALL require replacement.

#### Scenario: Name change triggers replacement

- GIVEN `name` changes in the Terraform plan
- WHEN Terraform plans
- THEN a resource replacement SHALL be required

#### Scenario: space_id change triggers replacement

- GIVEN `space_id` changes in the Terraform plan
- WHEN Terraform plans
- THEN a resource replacement SHALL be required

### Requirement: Connection (REQ-008)

The resource and data source SHALL use the provider-level Fleet client obtained via `clients.ConvertProviderData`. No resource-level connection override is supported.

#### Scenario: Provider client is used

- GIVEN a configured provider
- WHEN any CRUD or read operation runs
- THEN the Fleet client SHALL be obtained from the provider configuration

### Requirement: Compatibility — version-gated parameters (REQ-009–REQ-010)

When `ignore_mapping_update_errors` is configured with a known value, the resource SHALL verify the server version is at least 8.11.0; if the server version is lower, the resource SHALL return an error diagnostic with "Unsupported parameter for server version" and SHALL not call the install API.

When `skip_data_stream_rollover` is configured with a known value, the resource SHALL verify the server version is at least 8.11.0; if the server version is lower, the resource SHALL return an error diagnostic with "Unsupported parameter for server version" and SHALL not call the install API.

#### Scenario: ignore_mapping_update_errors on old server

- GIVEN `ignore_mapping_update_errors` is set and the server version is below 8.11.0
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic and SHALL not call the install API

#### Scenario: skip_data_stream_rollover on old server

- GIVEN `skip_data_stream_rollover` is set and the server version is below 8.11.0
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic and SHALL not call the install API

### Requirement: Create/Update — install options (REQ-011)

On create and update, the resource SHALL pass `force`, `prerelease`, and `ignore_constraints` as install options to the Fleet install package API when configured. When `ignore_mapping_update_errors` and `skip_data_stream_rollover` pass the server version check, they SHALL also be included in the install options. When `space_id` is a known value, the resource SHALL pass it as the space context for installation.

#### Scenario: space-aware installation

- GIVEN `space_id` is set to a known string
- WHEN create or update runs
- THEN the install API SHALL be called with that space ID as context

### Requirement: Create/Update — space_id state handling (REQ-012)

After a successful install, when `space_id` was unknown in the plan (not provided by user), the resource SHALL set `space_id` to null in state.

#### Scenario: space_id unknown becomes null

- GIVEN `space_id` is not configured by the user (unknown in plan)
- WHEN create completes
- THEN `space_id` in state SHALL be null

### Requirement: Update reuses create logic (REQ-013)

The resource's Update operation SHALL delegate to the same logic as Create, re-installing the package with the plan values.

#### Scenario: Version change triggers reinstall

- GIVEN `version` changes in the Terraform plan
- WHEN update runs
- THEN the Fleet install package API SHALL be called with the new version values from the plan

### Requirement: Delete — skip_destroy (REQ-014)

On delete, when `skip_destroy` is true, the resource SHALL skip calling the uninstall API and SHALL only remove the resource from Terraform state.

#### Scenario: skip_destroy set to true

- GIVEN `skip_destroy` is true
- WHEN destroy runs
- THEN the Fleet uninstall API SHALL NOT be called and the resource SHALL be removed from state

### Requirement: Delete — space-aware uninstall (REQ-015)

On delete, when `space_id` is a known value in state, the resource SHALL pass it as the space context to the Fleet uninstall API. The `force` flag from state SHALL be passed to the uninstall API.

#### Scenario: Delete with space context

- GIVEN `space_id` is set to a known string in state
- WHEN destroy runs
- THEN the Fleet uninstall API SHALL be called with that space ID and the `force` flag from state

### Requirement: State upgrade — v0 to v1 (REQ-016–REQ-017)

The resource SHALL support state upgrade from schema version 0 to version 1. During v0→v1 upgrade, the `space_ids` set attribute (v0) SHALL be replaced by a single `space_id` string attribute (v1). When `space_ids` in v0 state is null or unknown, `space_id` in the upgraded state SHALL be set to null. When `space_ids` in v0 state contains one or more values, the first element SHALL be used as `space_id` in the upgraded state, and a warning diagnostic SHALL be added on the `space_ids` path noting that multiple space IDs were present and only the first was selected. All other attributes (id, name, version, force, prerelease, ignore_mapping_update_errors, skip_data_stream_rollover, ignore_constraints, skip_destroy) SHALL be carried over unchanged.

#### Scenario: Single space ID in v0 state

- GIVEN v0 state with `space_ids = ["my-space"]`
- WHEN state upgrade runs
- THEN `space_id` in v1 state SHALL equal "my-space" and no warning SHALL be required (one space only)

#### Scenario: Multiple space IDs in v0 state

- GIVEN v0 state with `space_ids = ["space-a", "space-b"]`
- WHEN state upgrade runs
- THEN `space_id` in v1 state SHALL equal "space-a" and a warning diagnostic SHALL be added referencing the ignored spaces

#### Scenario: Null space_ids in v0 state

- GIVEN v0 state with `space_ids = null`
- WHEN state upgrade runs
- THEN `space_id` in v1 state SHALL be null
