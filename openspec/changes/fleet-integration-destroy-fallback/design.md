## Context

### Current destroy path (multi-space)

`internal/fleet/integration/delete.go` — `deleteIntegration`:

```
if spaceAware {
    pkg = fleet.GetPackage(...)
    if isInstalledInMultipleSpaces(pkg, spaceID) {
        fleet.DeleteKibanaAssets(...)   // <-- Fleet 9.5 returns HTTP 400 from the install space
        return
    }
}
fleet.Uninstall(...)
```

### Fleet 9.5 behaviour change

Fleet 9.5 introduced a new invariant: when the target space is the **install space** (the space where the package was originally installed), `DELETE /api/fleet/epm/packages/{name}/{version}/kibana_assets` returns HTTP 400:

```json
{"statusCode":400,"error":"Bad Request","message":"Impossible to delete kibana assets from the space where the package was installed, you must uninstall the package."}
```

The endpoint continues to succeed (HTTP 200) when called from a non-install space (deleting assets from an additional space).

The full uninstall endpoint (`DELETE /api/fleet/epm/packages/{name}/{version}`) is not restricted this way and succeeds from the install space with HTTP 200.

### CI trigger

On the 9.5.0-SNAPSHOT CI matrix, `TestAccResourceIntegrationSkipDestroy` triggers the multi-space branch of `deleteIntegration` because the CI docker environment causes Fleet's global asset inflation to reinstall packages across spaces ("Global Fleet assets changed … Reinstalling 2 packages"), making `isInstalledInMultipleSpaces` return true even for the primary install space. The resulting 400 is surfaced as a Terraform error and the test fails post-destroy.

### Validated fix

Direct API testing confirmed:
- `DELETE .../kibana_assets?force=true` on the install space → HTTP 400 (exact message above)
- `DELETE .../packages/{name}/{version}?force=true` (full uninstall) on the install space → HTTP 200

## Goals / Non-Goals

**Goals:**
- Fix `TestAccResourceIntegrationSkipDestroy` on Fleet 9.5+ without suppressing real errors.
- Preserve existing behaviour on Fleet versions prior to 9.5.
- Make the fallback auditable: log the fallback at DEBUG level before calling `Uninstall`.

**Non-Goals:**
- Changing the schema of `elasticstack_fleet_integration`.
- Modifying the install path or read path.
- Addressing other multi-space destroy scenarios not triggered by this specific 400.
- Implementing a version-check guard on the fallback (the 400 detection is message-based, making it version-agnostic and forward-compatible).

## Decisions

### Decision 1: Detect the 400 by message substring, fall back to Uninstall

The `fleet.DeleteKibanaAssets` function currently delegates to `handleDeleteResponse`, which treats all non-200/404 responses as errors. The fix must detect the specific "Impossible to delete kibana assets from the space where the package was installed" 400 before returning the error diagnostic.

There are two implementation sites:

**Option A — inside `DeleteKibanaAssets` in `internal/clients/fleet/packages.go`**: Add a case for the install-space 400 that returns a sentinel value or a second return value so callers can react. This requires changing the function signature.

**Option B — inside `deleteIntegration` in `internal/fleet/integration/delete.go`**: Call `DeleteKibanaAssets`, intercept the diagnostics, inspect the error summary/detail for the known message, and call `fleet.Uninstall` as a fallback. This keeps the change contained to the consumer.

**Selected: Option B (detect in `deleteIntegration`)**. It is the least invasive change: no signature change to `DeleteKibanaAssets`, no new return type, and the fallback logic is co-located with the destroy decision tree in `delete.go`. The detection string is defined as a package-level constant to make it easy to update if Fleet changes the error message.

### Decision 2: Log the fallback at DEBUG level

Before calling `fleet.Uninstall` as the fallback, emit a `tflog.Debug` message with the package name, version, and space ID so operators can audit the fallback in Terraform debug logs. Do not emit a WARNING because the fallback is the correct behavior on Fleet 9.5+ — not a degraded path.

### Decision 3: Fallback is unconditional on message match

The fallback triggers whenever `DeleteKibanaAssets` returns the known 400 message, regardless of Kibana version. This is intentional: the detection is message-based, not version-based. If Fleet reverts the change in a future version, the 400 would stop being returned and the fallback would never trigger.

## Open questions

None. The root cause, API behaviour, fix strategy, and implementation site are all confirmed by direct API testing.
