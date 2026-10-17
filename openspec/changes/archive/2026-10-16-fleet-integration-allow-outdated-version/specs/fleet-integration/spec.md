# Delta spec: `fleet-integration` — `allow_outdated_version`

Extends the main spec at `openspec/specs/fleet-integration/spec.md`.

## MODIFIED Requirements

### Requirement: Create/Update — install options (REQ-011)

On create and update, the resource SHALL pass `force`, `prerelease`, and `ignore_constraints` as install options to the Fleet install package API when configured. When `ignore_mapping_update_errors`, `skip_data_stream_rollover`, and `allow_outdated_version` pass any applicable server version check, they SHALL also be included in the install options when configured, and `allow_outdated_version` SHALL be omitted from the request body entirely when left unset. When `space_id` is a known value, the resource SHALL pass it as the space context for installation. After the regular install API succeeds, the resource SHALL wait for the package to reach a globally installed state by polling the Fleet get-package API using the **same space context** (the configured `space_id`, or the default space when `space_id` is not configured) as the install call, before evaluating target-space asset state. That first post-install poll SHALL use global install detection only and SHALL NOT apply strict target-space install detection during the wait. The poll SHALL NOT query the default-space endpoint when `space_id` is configured to a non-default space.

When `space_id` is configured, the Kibana server version is at least 9.1.0, and the package is installed globally but not in the target space after the regular install completes, the resource SHALL call the Fleet `POST /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` API scoped to the target space and wait until a strict space-aware read reports the package installed in the target space.

When `space_id` is configured, the Kibana server version is below 9.1.0, and the package is installed globally but not in the target space after the regular install completes, the resource SHALL emit a warning diagnostic that Kibana assets may not be available in the target space and SHALL preserve the existing global install behavior.

#### Scenario: space-aware installation

- GIVEN `space_id` is set to a known string
- WHEN create or update runs
- THEN the install API SHALL be called with that space ID as context
- AND the resource SHALL wait for global package installation via a get-package call scoped to that same space ID before any space-specific asset follow-up

#### Scenario: Post-install poll uses the configured space, not the default space

- GIVEN `space_id` is set to a known, non-default string
- AND the caller's Elastic credentials are scoped only to that space (no default-space access)
- WHEN create or update runs and the regular install API call succeeds
- THEN the post-install poll SHALL call the Fleet get-package API scoped to `space_id` (i.e. via the `/s/{space_id}/api/fleet/epm/packages/{name}/{version}` path)
- AND the poll SHALL treat the package as installed when Fleet reports a global installed state, without requiring strict target-space install metadata during this wait
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

#### Scenario: allow_outdated_version passed when configured

- GIVEN `allow_outdated_version` is set to a known boolean value (`true` or `false`)
- AND the Kibana server version is at least 9.6.0
- WHEN create or update runs
- THEN the install API request body SHALL include `allow_outdated_version` set to that value

#### Scenario: allow_outdated_version omitted when unset

- GIVEN `allow_outdated_version` is not configured (null/unknown)
- WHEN create or update runs
- THEN the install API request body SHALL NOT include the `allow_outdated_version` key

## ADDED Requirements

### Requirement: allow_outdated_version attribute (REQ-020)

The resource SHALL expose an `allow_outdated_version` attribute of type `bool`, `Optional`, with no `Computed` flag and no default — the same declaration shape as `force` and `ignore_constraints`. When true, installing the configured `version` SHALL bypass the Fleet out-of-date-version check, allowing installation of a version older than the latest available for the package. The attribute SHALL apply to both create and update, since both operations share the same install-options construction.

#### Scenario: Install an outdated version with allow_outdated_version true

- GIVEN a configuration with `version` set to a package version older than the latest available
- AND `allow_outdated_version = true`
- AND the Kibana server version is at least 9.6.0
- WHEN create or update runs
- THEN the install API request body SHALL include `allow_outdated_version: true`
- AND the install SHALL succeed despite the requested version being out of date

#### Scenario: Attribute value is preserved across refresh

- GIVEN `allow_outdated_version` is explicitly set in configuration
- WHEN read runs
- THEN the attribute's value in state SHALL remain unchanged by the read operation, consistent with the existing treatment of `force` and `ignore_constraints`

### Requirement: allow_outdated_version pointer-if-known wiring (REQ-021)

The resource SHALL wire `allow_outdated_version` pointer-if-known: a known value (`true` or `false`) SHALL be sent in the install request body; a null or unknown value SHALL be omitted from the request body entirely. Configurations that leave the attribute unset SHALL be unchanged on any server version. This differs from `force` and `ignore_constraints`, which are always sent (as `true` or `false`) regardless of whether they were explicitly configured.

#### Scenario: Unset attribute never sends the field

- GIVEN a configuration that never sets `allow_outdated_version`
- WHEN create or update runs
- THEN the install API request body SHALL NOT include an `allow_outdated_version` key
- AND behavior SHALL be unchanged from before this attribute existed, on any server version

### Requirement: allow_outdated_version version gate (REQ-022)

The resource SHALL require Kibana server version 9.6.0 or higher whenever `allow_outdated_version` is explicitly configured with a known value, including an explicit `false`. When the server version is lower, the resource SHALL return an "Unsupported server version" error diagnostic before calling the install API. A null or unknown `allow_outdated_version` SHALL NOT trigger any version check or diagnostic.

#### Scenario: Explicit value on unsupported server

- GIVEN `allow_outdated_version` is set to `true` or `false`
- AND the Kibana server version is below 9.6.0
- WHEN create or update runs
- THEN the resource SHALL return an error diagnostic with summary "Unsupported server version" stating that the `allow_outdated_version` parameter requires server version 9.6.0 or higher
- AND the install API SHALL NOT be called

#### Scenario: Unset on unsupported server

- GIVEN `allow_outdated_version` is not configured
- AND the Kibana server version is below 9.6.0
- WHEN create or update runs
- THEN no version diagnostic SHALL be returned for `allow_outdated_version`
