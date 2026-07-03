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
- Never silently uninstall a package from spaces other than the one being destroyed without an explicit opt-in (see Decision 4).

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

### Decision 4: Fallback requires `force = true` (added after initial implementation)

**Problem discovered during PR review**: `fleet.Uninstall` (`DELETE /api/fleet/epm/packages/{name}/{version}`) is a **global** operation — it removes the package from every space where it is installed, not just the target space. This is confirmed by Fleet/Kibana engineering discussion (elastic/kibana#186620, #172963): *"uninstall will uninstall the package and his assets in all space"*.

This means the original unconditional fallback had a serious side effect: if a user manages the same package across multiple spaces with separate `elasticstack_fleet_integration` resources (e.g. `space_id = "A"` and `space_id = "B"`), and space A happens to be the package's **install space** (typically whichever resource created the package first), destroying the resource for space A alone would silently uninstall the package from space B as well — even though space B's Terraform resource is untouched and still declared in configuration. The original fix replaced a loud (if ugly) 400 error with a silent, cross-resource destructive action, which violates the Terraform expectation that destroying one resource must not silently destroy another resource's backing infrastructure.

The CI trigger this change targets (`TestAccResourceIntegrationSkipDestroy`) is different: it hits the multi-space branch not because of a legitimate second Terraform-managed space, but because of the CI docker environment's "global asset reinflation" reinstalling the package into unrelated spaces used by other concurrently-running tests. There is no way for the provider to distinguish this incidental multi-space state from a genuine multi-space Terraform configuration purely from the Fleet API response.

Given that ambiguity, the resource now requires an explicit signal before taking the destructive global-uninstall path: the existing `force` attribute (already documented as "Set to true to force the requested action", already threaded through to the Fleet-level `force` query parameter on both `DeleteKibanaAssets` and `Uninstall`). This requires no schema change and happens to already be set to `true` in the `TestAccResourceIntegrationSkipDestroy` fixtures, so the original CI fix is preserved.

- When `force` is true, the fallback proceeds exactly as originally designed (call `fleet.Uninstall`, DEBUG log first).
- When `force` is false (the default), the resource surfaces a distinct, actionable error explaining the situation and telling the caller to either destroy the other space's resource(s) first or set `force = true` to accept the global-uninstall consequence. This preserves safety by default while still giving users an explicit escape hatch.

## Open questions

None. The root cause, API behaviour, fix strategy, and implementation site are all confirmed by direct API testing. The cross-space data-loss risk identified during review is addressed by Decision 4.
