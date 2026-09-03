## Why

Fleet 9.5 changed the semantics of the Kibana asset deletion endpoint. When a package is installed in a space (the "install space"), Fleet now rejects `DELETE /api/fleet/epm/packages/{name}/{version}/kibana_assets` with HTTP 400 and the message *"Impossible to delete kibana assets from the space where the package was installed, you must uninstall the package."* The provider's existing destroy path for `elasticstack_fleet_integration` calls this endpoint when `isInstalledInMultipleSpaces` is true, causing `TestAccResourceIntegrationSkipDestroy` to fail on the 9.5.0-SNAPSHOT CI matrix with dangling resources.

The fix is surgical: detect this specific 400 response from `DeleteKibanaAssets` and fall back to `fleet.Uninstall` (the full package uninstall endpoint), which Fleet 9.5 accepts cleanly. This fallback was validated directly against the 9.5 Fleet API (`DELETE /api/fleet/epm/packages/{name}/{version}` → HTTP 200).

**Important caveat discovered during review**: `fleet.Uninstall` removes the package **globally**, from every space where it is installed — not just the target space. If a user manages the same package in multiple spaces via separate `elasticstack_fleet_integration` resources, destroying the resource for whichever space happens to be the package's "install space" would otherwise silently uninstall the package from every other space too. To prevent this, the fallback only proceeds when `force = true` is set on the resource being destroyed; otherwise the resource surfaces an actionable error instead of silently taking a destructive, cross-resource action. See `design.md` Decision 4 for details.

## What Changes

A single behavioural change in the `elasticstack_fleet_integration` destroy path:

- In `internal/fleet/integration/delete.go` (or alternatively in `internal/clients/fleet/packages.go`), when `fleet.DeleteKibanaAssets` returns HTTP 400 with the install-space error message:
  - if `force` is true, the resource SHALL fall back to calling `fleet.Uninstall` instead of surfacing the 400 as a Terraform error;
  - if `force` is false (the default), the resource SHALL surface a distinct, actionable error instead of the raw 400, explaining the cross-space impact and directing the caller to destroy the other space's resource(s) first or set `force = true`.
- All other destroy paths (single-space, `skip_destroy`, non-space-aware) are unchanged.

## Capabilities

### Modified Capabilities

- `fleet-integration`: The destroy path for the `elasticstack_fleet_integration` resource is updated to handle the Fleet 9.5 rejection of `DELETE /api/fleet/epm/packages/{name}/{version}/kibana_assets` from the install space. When this 400 is encountered and `force = true`, the resource falls back to the full package uninstall endpoint, which succeeds and clears the install-space assets (globally, across all spaces). When `force` is not set, the resource surfaces an actionable error instead of silently uninstalling other spaces.

## Impact

- **Modified code**: `internal/fleet/integration/delete.go` and/or `internal/clients/fleet/packages.go` — error handling for the 400 install-space rejection, gated on the existing `force` attribute.
- **Test fix**: `TestAccResourceIntegrationSkipDestroy` passes on Fleet 9.5.0-SNAPSHOT CI matrix once the fallback is in place (its fixtures already set `force = true`).
- **No schema change**: no new attributes, no state upgrader, no provider config changes. The existing `force` attribute is reused to gate this behavior.
- **Backward compatibility**: the fallback is only triggered by the specific 400 message from Fleet 9.5+. On older Fleet versions this fallback path is not reached.
- **Safety**: without `force = true`, a user destroying the install-space resource of a package installed in multiple spaces gets a clear, actionable error instead of either a raw Fleet 400 or (as in the initial iteration of this change) a silent global uninstall of every space.
