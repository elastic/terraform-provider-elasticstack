## Context

The `elasticstack_fleet_integration` resource manages Fleet integration packages via `POST /api/fleet/epm/packages/{pkg}/{version}` (install) and `DELETE /api/fleet/epm/packages/{pkg}/{version}` (uninstall). Fleet treats the package installation record as a global saved object, but Kibana assets (dashboards, visualizations, saved searches) are space-scoped.

When a user installs `tcp@1.16.0` in Space A and then tries to install the same package in Space B, Kibana's install endpoint returns success immediately because the package is already installed *globally*. The target space never receives assets. Terraform's wait loop also succeeds because `GET /api/fleet/epm/packages/{pkg}/{version}` returns `status: installed` regardless of which space queried it.

Kibana 8.15.0 resolved this with:
- `POST /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` — installs Kibana assets into a specific space for an already-installed package.
- `DELETE /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` — removes Kibana assets from the current space.
- `PackageInfo.InstallationInfo.AdditionalSpacesInstalledKibana` — tracks which spaces have assets beyond the primary install space.
- `PackageInfo.InstallationInfo.InstalledKibanaSpaceId` — tracks the primary install space.

The provider must bridge the gap between Terraform's per-resource model (one resource = one space) and Fleet's global install model.

## Goals / Non-Goals

**Goals:**
- Make `elasticstack_fleet_integration` correctly represent and manage space-scoped Kibana assets when `space_id` is configured on Kibana >= 8.15.0.
- Ensure `Create` installs assets in the target space even when the package is already installed elsewhere.
- Ensure `Delete` only removes assets from the target space when the package is installed in multiple spaces, and fully uninstalls when it's the only remaining space.
- Ensure `Read` accurately reflects whether assets exist in the target space (drift detection).
- Maintain backward compatibility: on Kibana < 8.15.0 or when `space_id` is not set, behavior is unchanged.

**Non-Goals:**
- Supporting installation of the *same* package version in the same space via multiple Terraform resources (this is already prevented by API behavior).
- Managing `additional_spaces_installed_kibana` as a Terraform-managed list (the resource remains a single-space abstraction).
- Auto-installing assets into all spaces referenced by agent policies (out of scope; Kibana issue #237658).
- Changing the resource schema.
- Supporting Kibana spaces experimental feature flags (the provider assumes the APIs are available if the version check passes).

## Decisions

### Decision 1: Version gate at 8.15.0 for space-aware behavior
**Rationale:** The `installed_kibana_space_id` field exists since Kibana 8.1.0, but the `kibana_assets` endpoints did not arrive until 8.15.0. If we used field presence as a capability signal, a provider running against Kibana 8.1–8.14 would detect "not installed in target space" but have no way to install assets there, causing a perpetual diff → create → wait → diff loop.

**Alternatives considered:**
- **Field presence only** — rejected due to the perpetual diff risk on 8.1–8.14.
- **Endpoint probe (call kibana_assets, fall back on 404)** — viable but adds a spurious failed API call on every create. Version check is cleaner and matches existing patterns in the codebase (`MinVersionIgnoreMappingUpdateErrors`).

### Decision 2: `fleetPackageInstalled` accepts `spaceID` and `spaceAware` parameters
**Rationale:** The helper is used in `Read`, the Create wait loop, and the Create pre-flight check. It needs to know both which space to check and whether the server supports strict space checks. When `spaceAware` is false, the existing global check is preserved.

**Signature:**
```go
func fleetPackageInstalled(pkg *kbapi.PackageInfo, spaceID string, spaceAware bool) bool
```

**Alternatives considered:**
- **Always strict when `space_id` is set** — rejected because it breaks on Kibana < 8.15.0 (field may exist but endpoint doesn't).
- **Check version inside the helper** — rejected because it would require passing a `*KibanaScopedClient` into a pure function, complicating testing and the wait-loop callback.

### Decision 3: Create pre-flights with `GetPackage` before deciding install vs. kibana_assets
**Rationale:** There are three distinct states on create:
1. Package not installed anywhere → `InstallPackage` (full install).
2. Package installed in target space already → `InstallPackage` (upgrade or no-op).
3. Package installed in a *different* space → `InstallKibanaAssets` (space-scoped asset install).

A single `GetPackage` call before the install operation resolves which path to take. This avoids guessing or relying on install API response codes.

**Alternatives considered:**
- **Always call InstallPackage, then check if assets are missing and lazily call kibana_assets** — rejected because the install API returns success (no-op) with no signal that assets are missing, so we can't trigger the fallback without another read.
- **Call kibana_assets first, then InstallPackage on failure** — rejected because `kibana_assets` requires the package to already be installed globally; it would fail on first install.

### Decision 4: Delete checks `isInstalledInMultipleSpaces` to choose between asset removal and full uninstall
**Rationale:** A user declares one resource per space. When destroying the resource for Space B, if Space A still holds assets, we must not call the global `Uninstall` API (that would orphan Space A's Terraform state and remove A's assets). We only call `Uninstall` when the package is installed in *only* the target space.

**Multi-space detection logic:**
```go
func isInstalledInMultipleSpaces(pkg *kbapi.PackageInfo, spaceID string) bool {
    if pkg.InstallationInfo == nil { return false }
    otherSpaces := 0
    if pkg.InstallationInfo.AdditionalSpacesInstalledKibana != nil {
        otherSpaces = len(*pkg.InstallationInfo.AdditionalSpacesInstalledKibana)
    }
    // If target is the primary space, additional spaces = multi
    // If target is in additional spaces, primary + (additional minus self) = multi
    isPrimary := pkg.InstallationInfo.InstalledKibanaSpaceId != nil &&
                 *pkg.InstallationInfo.InstalledKibanaSpaceId == spaceID
    if isPrimary {
        return otherSpaces > 0
    }
    return otherSpaces > 1 || pkg.InstallationInfo.InstalledKibanaSpaceId != nil
}
```

### Decision 5: Fleet client helpers sit in `internal/clients/fleet/fleet.go`
**Rationale:** This is the existing convention for all Fleet API wrappers. The helpers wrap the generated `kbapi` client and handle status-code translation, error formatting, and space-aware path injection via `kibanautil.SpaceAwarePathRequestEditor`.

### Decision 6: `space_id` stays Optional+Computed with RequiresReplace
**Rationale:** No schema changes are needed. The existing `space_id` semantics already express "install in this space". The behavioral change is entirely in how the resource interacts with the API.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| **[Risk]** Kibana 8.15.x may have bugs in the `kibana_assets` endpoints (e.g., issue #210141). | **[Mitigation]** The 404/400 fallback on `kibana_assets` errors routes back to `InstallPackage` / `Uninstall`, preserving current behavior if the endpoint is unexpectedly unavailable. |
| **[Risk]** `AdditionalSpacesInstalledKibana` may be empty or missing on some Kibana configurations even at 8.15+. | **[Mitigation]** `fleetPackageInstalled` falls back to returning `true` (treating as installed) when `InstallationInfo` is nil, avoiding false negatives that would cause unnecessary reinstalls. When `InstallationInfo` is present but no space match is found, it returns `false`, which is the conservative (correct) behavior for drift detection. |
| **[Risk]** Version check adds a `/api/status` call per resource operation on `space_id`-configured resources. | **[Mitigation]** Cached version checks could be added to `KibanaScopedClient` later without changing resource logic. For now, the cost is one lightweight status call per CRUD operation, which is already done in `create.go` for `ignore_mapping_update_errors` / `skip_data_stream_rollover` gates. |
| **[Risk]** Serverless behavior differs from stateful. | **[Mitigation]** `EnforceMinVersion` returns `true` for serverless, so the space-aware path is always attempted. If serverless does not expose the endpoint, the 404 fallback activates. This is consistent with other serverless/version-gated features in the provider. |
| **[Risk]** Concurrent `create` calls for the same package in different spaces may race. | **[Mitigation]** Kibana handles the global install object. If Space A's create is mid-install when Space B's create starts, Space B's pre-flight may see "not installed" and call `InstallPackage`. On Kibana 8.15+, that is still a no-op (global install), but the wait loop will eventually see the installed state and then the subsequent `kibana_assets` call (if needed) will succeed. In practice, Terraform plans are deterministic and this race is unlikely. |
