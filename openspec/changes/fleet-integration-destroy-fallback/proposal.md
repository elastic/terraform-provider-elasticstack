## Why

Fleet 9.5 changed the semantics of the Kibana asset deletion endpoint. When a package is installed in a space (the "install space"), Fleet now rejects `DELETE /api/fleet/epm/packages/{name}/{version}/kibana_assets` with HTTP 400 and the message *"Impossible to delete kibana assets from the space where the package was installed, you must uninstall the package."* The provider's existing destroy path for `elasticstack_fleet_integration` calls this endpoint when `isInstalledInMultipleSpaces` is true, causing `TestAccResourceIntegrationSkipDestroy` to fail on the 9.5.0-SNAPSHOT CI matrix with dangling resources.

The fix is surgical: detect this specific 400 response from `DeleteKibanaAssets` and fall back to `fleet.Uninstall` (the full package uninstall endpoint), which Fleet 9.5 accepts cleanly. This fallback was validated directly against the 9.5 Fleet API (`DELETE /api/fleet/epm/packages/{name}/{version}` → HTTP 200).

## What Changes

A single behavioural change in the `elasticstack_fleet_integration` destroy path:

- In `internal/fleet/integration/delete.go` (or alternatively in `internal/clients/fleet/packages.go`), when `fleet.DeleteKibanaAssets` returns HTTP 400 with the install-space error message, the resource SHALL fall back to calling `fleet.Uninstall` instead of surfacing the 400 as a Terraform error.
- All other destroy paths (single-space, `skip_destroy`, non-space-aware) are unchanged.

## Capabilities

### Modified Capabilities

- `fleet-integration`: The destroy path for the `elasticstack_fleet_integration` resource is updated to handle the Fleet 9.5 rejection of `DELETE /api/fleet/epm/packages/{name}/{version}/kibana_assets` from the install space. When this 400 is encountered, the resource falls back to the full package uninstall endpoint, which succeeds and clears the install-space assets.

## Impact

- **Modified code**: `internal/fleet/integration/delete.go` and/or `internal/clients/fleet/packages.go` — error handling for the 400 install-space rejection.
- **Test fix**: `TestAccResourceIntegrationSkipDestroy` passes on Fleet 9.5.0-SNAPSHOT CI matrix once the fallback is in place.
- **No schema change**: no new attributes, no state upgrader, no provider config changes.
- **Backward compatibility**: the fallback is only triggered by the specific 400 message from Fleet 9.5+. On older Fleet versions this fallback path is not reached.
