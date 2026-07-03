## MODIFIED Requirements

### Requirement: Delete — space-aware uninstall (REQ-015)

This requirement modifies REQ-015 (space-aware uninstall). When the Fleet API returns HTTP 400 with a message containing the substring "Impossible to delete kibana assets from the space where the package was installed" in response to `DELETE /api/fleet/epm/packages/{pkg}/{version}/kibana_assets`, the target space is the package's install space and the package is also installed in one or more other spaces. Because `fleet.Uninstall` (`DELETE /api/fleet/epm/packages/{name}/{version}`) removes the package globally — including from every other space where it is installed — the resource SHALL only take this destructive fallback when the caller has explicitly opted in via the `force` flag in state:

- When `force` is true in state, the resource SHALL NOT surface the 400 as a Terraform error and SHALL instead call `fleet.Uninstall` with the `force` flag from state. The resource SHALL emit a DEBUG log entry before the fallback recording the package name, version, and space ID.
- When `force` is false (the default) in state, the resource SHALL NOT call `fleet.Uninstall` and SHALL instead surface a distinct, actionable error diagnostic explaining that the target space is the package's install space, that the package remains installed in other spaces, that Fleet does not support removing Kibana assets from only the install space in this situation, and that the caller must either destroy the resource(s) managing the other space(s) first or set `force = true` to accept a global uninstall.

All other non-success responses from `DeleteKibanaAssets` SHALL still be surfaced as Terraform diagnostics without triggering either behavior above, regardless of `force`.

#### Scenario: DeleteKibanaAssets returns install-space 400 with force enabled

- **GIVEN** `space_id` is set to a known string in state
- **AND** the package is installed in multiple spaces
- **AND** `force` is true in state
- **AND** `DeleteKibanaAssets` returns HTTP 400 with message containing `"Impossible to delete kibana assets from the space where the package was installed"`
- **WHEN** destroy runs
- **THEN** the resource SHALL NOT surface the 400 as a Terraform error
- **AND** the resource SHALL call `fleet.Uninstall` with the package name, version, space ID, and `force` flag from state
- **AND** diagnostics from `fleet.Uninstall` SHALL be returned to the caller
- **AND** a DEBUG log entry SHALL be emitted before the fallback indicating the package name, version, and space ID

#### Scenario: DeleteKibanaAssets returns install-space 400 without force

- **GIVEN** `space_id` is set to a known string in state
- **AND** the package is installed in multiple spaces
- **AND** `force` is false (or unset) in state
- **AND** `DeleteKibanaAssets` returns HTTP 400 with message containing `"Impossible to delete kibana assets from the space where the package was installed"`
- **WHEN** destroy runs
- **THEN** the resource SHALL surface an actionable error diagnostic distinct from the raw Fleet 400 message
- **AND** `fleet.Uninstall` SHALL NOT be called
- **AND** the error SHALL indicate that the caller must destroy the resource(s) for the other space(s) first or set `force = true`

#### Scenario: DeleteKibanaAssets returns a different 400

- **GIVEN** `space_id` is set to a known string in state
- **AND** the package is installed in multiple spaces
- **AND** `DeleteKibanaAssets` returns HTTP 400 with a message NOT containing the install-space substring
- **WHEN** destroy runs
- **THEN** the resource SHALL surface the 400 as a Terraform error regardless of `force`
- **AND** `fleet.Uninstall` SHALL NOT be called

#### Scenario: DeleteKibanaAssets succeeds (HTTP 200)

- **GIVEN** `space_id` is set to a known string in state
- **AND** the package is installed in multiple spaces
- **AND** `DeleteKibanaAssets` returns HTTP 200
- **WHEN** destroy runs
- **THEN** no fallback is triggered
- **AND** `fleet.Uninstall` SHALL NOT be called as a fallback
- **AND** no error diagnostic is returned
