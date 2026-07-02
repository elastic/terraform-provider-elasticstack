## MODIFIED Requirements

### Requirement: Delete — space-aware uninstall (REQ-015) — install-space 400 fallback

This requirement modifies REQ-015 (space-aware uninstall). When the Fleet API returns HTTP 400 with a message containing the substring "Impossible to delete kibana assets from the space where the package was installed" in response to `DELETE /api/fleet/epm/packages/{pkg}/{version}/kibana_assets`, the resource SHALL NOT surface this as a Terraform error and SHALL instead call `fleet.Uninstall` (`DELETE /api/fleet/epm/packages/{name}/{version}`) with the `force` flag from state (which uninstalls the package globally, since install-space asset-only deletion is not supported). The resource SHALL emit a DEBUG log entry before the fallback recording the package name, version, and space ID. All other non-success responses from `DeleteKibanaAssets` SHALL still be surfaced as Terraform diagnostics without triggering the fallback.

#### Scenario: DeleteKibanaAssets returns install-space 400

- **GIVEN** `space_id` is set to a known string in state
- **AND** the package is installed in multiple spaces
- **AND** `DeleteKibanaAssets` returns HTTP 400 with message containing `"Impossible to delete kibana assets from the space where the package was installed"`
- **WHEN** destroy runs
- **THEN** the resource SHALL NOT surface the 400 as a Terraform error
- **AND** the resource SHALL call `fleet.Uninstall` with the package name, version, space ID, and `force` flag from state
- **AND** diagnostics from `fleet.Uninstall` SHALL be returned to the caller
- **AND** a DEBUG log entry SHALL be emitted before the fallback indicating the package name, version, and space ID

#### Scenario: DeleteKibanaAssets returns a different 400

- **GIVEN** `space_id` is set to a known string in state
- **AND** the package is installed in multiple spaces
- **AND** `DeleteKibanaAssets` returns HTTP 400 with a message NOT containing the install-space substring
- **WHEN** destroy runs
- **THEN** the resource SHALL surface the 400 as a Terraform error
- **AND** `fleet.Uninstall` SHALL NOT be called

#### Scenario: DeleteKibanaAssets succeeds (HTTP 200)

- **GIVEN** `space_id` is set to a known string in state
- **AND** the package is installed in multiple spaces
- **AND** `DeleteKibanaAssets` returns HTTP 200
- **WHEN** destroy runs
- **THEN** no fallback is triggered
- **AND** `fleet.Uninstall` SHALL NOT be called as a fallback
- **AND** no error diagnostic is returned
