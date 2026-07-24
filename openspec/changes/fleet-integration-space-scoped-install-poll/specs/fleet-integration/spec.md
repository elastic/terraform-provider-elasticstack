## MODIFIED Requirements

### Requirement: Create/Update — install options (REQ-011)

On create and update, the resource SHALL pass `force`, `prerelease`, and `ignore_constraints` as install options to the Fleet install package API when configured. When `ignore_mapping_update_errors` and `skip_data_stream_rollover` pass the server version check, they SHALL also be included in the install options. When `space_id` is a known value, the resource SHALL pass it as the space context for installation. After the regular install API succeeds, the resource SHALL wait for the package to reach an installed state using the **same space context** (the configured `space_id`, or the default space when `space_id` is not configured) as the install call, before evaluating target-space asset state. The wait SHALL NOT query the default-space endpoint when `space_id` is configured to a non-default space.

When `space_id` is configured, the Kibana server version is at least 9.1.0, and the package is installed globally but not in the target space after the regular install completes, the resource SHALL call the Fleet `POST /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` API scoped to the target space and wait until a strict space-aware read reports the package installed in the target space.

When `space_id` is configured, the Kibana server version is below 9.1.0, and the package is installed globally but not in the target space after the regular install completes, the resource SHALL emit a warning diagnostic that Kibana assets may not be available in the target space and SHALL preserve the existing global install behavior.

#### Scenario: space-aware installation

- GIVEN `space_id` is set to a known string
- WHEN create or update runs
- THEN the install API SHALL be called with that space ID as context
- AND the resource SHALL wait for package installation using that same space ID as context before any space-specific asset follow-up

#### Scenario: Post-install poll uses the configured space, not the default space

- GIVEN `space_id` is set to a known, non-default string
- AND the caller's Elastic credentials are scoped only to that space (no default-space access)
- WHEN create or update runs and the regular install API call succeeds
- THEN the post-install poll SHALL call the Fleet get-package API scoped to `space_id` (i.e. via the `/s/{space_id}/api/fleet/epm/packages/{name}/{version}` path)
- AND the poll SHALL NOT return an error caused by insufficient default-space permissions
- AND the resource creation SHALL succeed once the package is installed in the target space

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
