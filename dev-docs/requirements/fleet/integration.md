# `elasticstack_fleet_integration` — Schema and Functional Requirements

Resource implementation: `internal/fleet/integration`

## Schema

```hcl
resource "elasticstack_fleet_integration" "example" {
  name    = <required, string> # requires replacement when changed
  version = <required, string>

  force      = <optional, bool>
  prerelease = <optional, bool>

  ignore_constraints = <optional, bool>

  # Requires stack server version >= 8.11.0 when set
  ignore_mapping_update_errors = <optional, bool>
  # Requires stack server version >= 8.11.0 when set
  skip_data_stream_rollover = <optional, bool>

  # When true, destroy removes from state without uninstalling the package
  skip_destroy = <optional, bool>

  # Optional+computed; uses Kibana spaces path when set; 
  # Requires replacement if:
  # * Explicitly changed (e.g "first" -> "second")
  # * Null/unknown <-> a non-default (`default`) value
  space_id = <optional, computed, string, default=default>

  id = <computed, string>
}
```

## Requirements

- **[REQ-001] (API)**: The resource shall use the Kibana Fleet EPM Install a package from the registry API to install the configured package name and version ([API docs](https://www.elastic.co/docs/api/doc/kibana/operation/operation-post-fleet-epm-packages-pkgname-pkgversion)).
- **[REQ-002] (API)**: The resource shall use the Kibana Fleet EPM Get a package API to read the configured package name and version ([API docs](https://www.elastic.co/docs/api/doc/kibana/operation/operation-get-fleet-epm-packages-pkgname-pkgversion)).
- **[REQ-003] (API)**: The resource shall use the Kibana Fleet EPM Delete a package API to uninstall the configured package name and version ([API docs](https://www.elastic.co/docs/api/doc/kibana/operation/operation-delete-fleet-epm-packages-pkgname-pkgversion)).
- **[REQ-004] (API)**: The resource shall use the Kibana Fleet EPM Install Kibana assets for a package API to install Kibana assets for the configured package name and version into the configured Kibana space ([API docs](https://www.elastic.co/docs/api/doc/kibana/v9/operation/operation-post-fleet-epm-packages-pkgname-pkgversion-kibana-assets)).
- **[REQ-005] (API)**: When a Fleet API request returns a non-success status code, the resource shall surface the API error to Terraform diagnostics, except for the “not found / not installed” cases explicitly handled in the Read and Delete requirements below.

- **[REQ-006] (Connection)**: The resource shall use the provider-configured Fleet/Kibana client for all Fleet API calls.
- **[REQ-007] (Connection)**: If the Fleet/Kibana client cannot be constructed, the resource shall fail the operation with an error diagnostic.

- **[REQ-008] (Identity)**: The resource shall expose a computed `id` representing a stable hash of the configured `name` and `version`.
- **[REQ-009] (Identity)**: After a successful create or update, the resource shall set `id` in state based on the configured `name` and `version`.
- **[REQ-010] (Identity)**: During refresh, when the package is found installed, the resource shall set `id` in state based on the stored `name` and `version`.

- **[REQ-011] (Lifecycle)**: When `name` changes, the resource shall require replacement (destroy/recreate), not an in-place update.
- **[REQ-012] (Lifecycle)**: When `space_id` changes, the resource shall require replacement (destroy/recreate), not an in-place update.
- **[REQ-013] (Lifecycle)**: When `version` changes, the resource shall perform an in-place update by re-running the install operation for the desired `name` and `version`.

- **[REQ-014] (Create/Update)**: When creating or updating, the resource shall call the Fleet EPM install API with the following options derived from configuration: `force`, `prerelease`, and `ignore_constraints`.
- **[REQ-015] (Create/Update)**: When `ignore_mapping_update_errors` is configured (known and non-null) and supported by the stack server version, the resource shall pass it to the Fleet EPM install API request.
- **[REQ-016] (Create/Update)**: When `skip_data_stream_rollover` is configured (known and non-null) and supported by the stack server version, the resource shall pass it to the Fleet EPM install API request.
- **[REQ-017] (Create/Update)**: When `space_id` is configured (known and non-null), the resource shall use the Kibana spaces-aware Fleet EPM install API path for that space.
- **[REQ-018] (Create/Update)**: After successfully installing the package, the resource shall install Kibana assets for that package into the configured `space_id`.
- **[REQ-019] (Plan/State)**: When `space_id` is unknown during create/update, the resource shall set `space_id` to null in state.

- **[REQ-020] (Compatibility)**: When `ignore_mapping_update_errors` is configured (known and non-null), the resource shall retrieve the stack server version and fail with an “Unsupported parameter for server version” error unless the version is at least 8.11.0.
- **[REQ-021] (Compatibility)**: When `skip_data_stream_rollover` is configured (known and non-null), the resource shall retrieve the stack server version and fail with an “Unsupported parameter for server version” error unless the version is at least 8.11.0.
- **[REQ-022] (Compatibility)**: If retrieving the stack server version fails, the resource shall surface the error to Terraform diagnostics and shall not call the Fleet EPM install API.

- **[REQ-023] (Read)**: When refreshing state, the resource shall call the Fleet EPM get package API with the stored `name` and `version`.
- **[REQ-024] (Read)**: The resource shall consider the integration installed only if the package status is `installed` and either (a) the package installation info installed Kibana space ID equals the configured `space_id`, or (b) the package installation info additional-spaces-installed Kibana map contains the configured `space_id`.
- **[REQ-025] (Read)**: If the package is not found (HTTP 404) during refresh, the resource shall remove itself from Terraform state.
- **[REQ-026] (Read)**: If the package is found but does not meet the “installed” criteria (status and space checks), the resource shall remove itself from Terraform state.

- **[REQ-027] (Delete)**: When destroying and `skip_destroy` is true, the resource shall not uninstall the package and shall only remove the resource from Terraform state.
- **[REQ-028] (Delete)**: When destroying and `skip_destroy` is false, the resource shall call the Fleet EPM delete package API with the stored `name` and `version`.
- **[REQ-029] (Delete)**: When destroying and `space_id` is known and non-null, the resource shall use the Kibana spaces-aware Fleet EPM delete API path for that space.
- **[REQ-030] (Delete)**: When uninstalling, if the Fleet EPM delete package API returns HTTP 404, the resource shall treat the operation as successful.
- **[REQ-031] (Delete)**: When uninstalling, if the Fleet EPM delete package API returns HTTP 400 with a message indicating the package is not installed, the resource shall treat the operation as successful.
- **[REQ-032] (Import)**: The resource shall not implement Terraform import, and import attempts shall fail.
- **[REQ-033] (StateUpgrade)**: The resource shall support upgrading stored state from schema version 0 to schema version 1.
- **[REQ-034] (StateUpgrade)**: During v0→v1 upgrade, the resource shall map the prior `space_ids` set attribute into the v1 `space_id` attribute by selecting the first element (if any) and setting `space_id` to null when the set is empty or null.
- **[REQ-035] (StateUpgrade)**: During v0→v1 upgrade, if multiple `space_ids` are present, the resource shall emit a warning indicating only the first was selected and the remainder were ignored.