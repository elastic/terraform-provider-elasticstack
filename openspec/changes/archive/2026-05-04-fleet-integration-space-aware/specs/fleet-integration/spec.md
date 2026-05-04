## MODIFIED Requirements

### Requirement: Fleet package read API (REQ-004–REQ-005)

The resource SHALL use the Kibana Fleet get package API to refresh state on read, supplying the `name` and `version` from state. When `space_id` is configured, the get package API call SHALL be scoped to that Kibana space. The data source SHALL use the Kibana Fleet list packages API to retrieve available packages, filtered by the `prerelease` parameter.

When `space_id` is configured and the Kibana server version is at least 9.1.0, the resource SHALL determine that a package is "installed" only when its installation status is "installed" AND the package has Kibana assets installed in the target space, as indicated by `InstalledKibanaSpaceId` or `AdditionalSpacesInstalledKibana` in the Fleet API response. On Kibana versions below 9.1.0, or when `space_id` is not configured, the resource SHALL retain the existing behavior of treating global install status as sufficient.

#### Scenario: Package not found on resource read

- GIVEN the integration package is not present (status not "installed") or nil in the Fleet API response
- WHEN read runs on the resource
- THEN the resource SHALL be removed from Terraform state

#### Scenario: Package installed in a different space on resource read

- GIVEN the integration package is installed globally but its Kibana assets are in a different space than the resource's `space_id`
- AND the Kibana server version is at least 9.1.0
- WHEN read runs on the resource
- THEN the resource SHALL be removed from Terraform state

#### Scenario: Data source package lookup

- GIVEN a valid `name` and optional `prerelease` flag
- WHEN read runs on the data source
- THEN the data source SHALL set `version` to the version returned by the Fleet list packages API for the matching package name, or null if not found

#### Scenario: Data source with space_id

- GIVEN `space_id` is set to a non-default Kibana space
- WHEN read runs on the data source
- THEN the list packages API SHALL be called with that space ID as context (i.e. via the `/s/{space_id}/api/fleet/epm/packages` path)

### Requirement: Create/Update — install options (REQ-011)

On create and update, the resource SHALL pass `force`, `prerelease`, and `ignore_constraints` as install options to the Fleet install package API when configured. When `ignore_mapping_update_errors` and `skip_data_stream_rollover` pass the server version check, they SHALL also be included in the install options. When `space_id` is a known value, the resource SHALL pass it as the space context for installation. After the regular install API succeeds, the resource SHALL wait for the package to reach a globally installed state before evaluating target-space asset state.

When `space_id` is configured, the Kibana server version is at least 9.1.0, and the package is installed globally but not in the target space after the regular install completes, the resource SHALL call the Fleet `POST /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` API scoped to the target space and wait until a strict space-aware read reports the package installed in the target space.

When `space_id` is configured, the Kibana server version is below 9.1.0, and the package is installed globally but not in the target space after the regular install completes, the resource SHALL emit a warning diagnostic that Kibana assets may not be available in the target space and SHALL preserve the existing global install behavior.

#### Scenario: space-aware installation

- GIVEN `space_id` is set to a known string
- WHEN create or update runs
- THEN the install API SHALL be called with that space ID as context
- AND the resource SHALL wait for global package installation before any space-specific asset follow-up

#### Scenario: Package already installed in another space

- GIVEN `space_id` is set to a known string
- AND the Kibana server version is at least 9.1.0
- AND the integration package is installed globally but not in the target space after the regular install completes
- WHEN create or update runs
- THEN the resource SHALL call the Fleet kibana_assets API scoped to the target space to install Kibana assets there
- AND the resource SHALL wait for the package to be installed in the target space using strict space-aware install detection

#### Scenario: Package installed in another space on unsupported server

- GIVEN `space_id` is set to a known string
- AND the Kibana server version is below 9.1.0
- AND the integration package is installed globally but not in the target space after the regular install completes
- WHEN create or update runs
- THEN the resource SHALL emit a warning diagnostic that Kibana assets may not be available in the target space
- AND no error diagnostic SHALL be returned solely due to the server version being below 9.1.0

### Requirement: Delete — space-aware uninstall (REQ-015)

On delete, when `space_id` is a known value in state, the resource SHALL determine whether the package is installed in multiple spaces. When the Kibana server version is at least 9.1.0 and the package is installed in multiple spaces, the resource SHALL call the Fleet `DELETE /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` API scoped to the target space to remove Kibana assets from that space only. When the package is installed in only the target space, or when the Kibana server version is below 9.1.0, the resource SHALL pass `space_id` as the space context to the Fleet uninstall API. The `force` flag from state SHALL be passed to whichever API is called.

#### Scenario: Delete with space context — single space installation

- GIVEN `space_id` is set to a known string in state
- AND the package is installed in only that space (or the Kibana server is below 9.1.0)
- WHEN destroy runs
- THEN the Fleet uninstall API SHALL be called with that space ID and the `force` flag from state

#### Scenario: Delete with space context — multi-space installation

- GIVEN `space_id` is set to a known string in state
- AND the Kibana server version is at least 9.1.0
- AND the package is installed in multiple spaces
- WHEN destroy runs
- THEN the Fleet delete kibana_assets API SHALL be called scoped to the target space
- AND the global package installation SHALL remain intact
- AND the `force` flag from state SHALL be passed to the API

## ADDED Requirements

### Requirement: Compatibility — space-aware behavior version gate (REQ-018)

When `space_id` is configured with a known value, the resource SHALL verify the server version before enabling space-aware behavior. If the Kibana server version is at least 9.1.0, the resource SHALL use space-aware install and uninstall logic. If the Kibana server version is below 9.1.0, the resource SHALL fall back to the existing non-space-aware behavior.

#### Scenario: space-aware behavior on supported server

- GIVEN `space_id` is set to a known string
- AND the Kibana server version is at least 9.1.0
- WHEN create or delete runs
- THEN the resource SHALL use space-aware logic (kibana_assets endpoints for cross-space installs, multi-space detection on delete)

#### Scenario: space-aware behavior falls back on old server

- GIVEN `space_id` is set to a known string
- AND the Kibana server version is below 9.1.0
- WHEN create or delete runs
- THEN the resource SHALL use the existing non-space-aware behavior (regular install and uninstall APIs)
- AND no error diagnostic SHALL be returned solely due to the server version being below 9.1.0
