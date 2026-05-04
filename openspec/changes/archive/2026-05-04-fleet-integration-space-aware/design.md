## Context

The `elasticstack_fleet_integration` resource manages Fleet integration packages via `POST /api/fleet/epm/packages/{pkg}/{version}` (install) and `DELETE /api/fleet/epm/packages/{pkg}/{version}` (uninstall). Fleet treats the package installation record as a global saved object, but Kibana assets (dashboards, visualizations, saved searches) are space-scoped.

When a user installs `tcp@1.16.0` in Space A and then tries to install the same package in Space B, Kibana's install endpoint returns success immediately because the package is already installed *globally*. The target space never receives assets. Terraform's wait loop also succeeds because `GET /api/fleet/epm/packages/{pkg}/{version}` returns `status: installed` regardless of which space queried it.

Kibana introduced APIs for this with:
- `POST /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` — installs Kibana assets into a specific space for an already-installed package.
- `DELETE /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` — removes Kibana assets from the current space.
- `PackageInfo.InstallationInfo.AdditionalSpacesInstalledKibana` — tracks which spaces have assets beyond the primary install space.
- `PackageInfo.InstallationInfo.InstalledKibanaSpaceId` — tracks the primary install space.

The provider must bridge the gap between Terraform's per-resource model (one resource = one space) and Fleet's global install model.

## Goals / Non-Goals

**Goals:**
- Make `elasticstack_fleet_integration` correctly represent and manage space-scoped Kibana assets when `space_id` is configured on Kibana >= 9.1.0.
- Ensure `Create` installs assets in the target space even when the package is already installed elsewhere.
- Ensure `Delete` only removes assets from the target space when the package is installed in multiple spaces, and fully uninstalls when it's the only remaining space.
- Ensure `Read` accurately reflects whether assets exist in the target space (drift detection).
- Maintain backward compatibility: on Kibana < 9.1.0 or when `space_id` is not set, behavior is unchanged apart from a warning when the provider detects assets may be missing from the target space.

**Non-Goals:**
- Supporting installation of the *same* package version in the same space via multiple Terraform resources (this is already prevented by API behavior).
- Managing `additional_spaces_installed_kibana` as a Terraform-managed list (the resource remains a single-space abstraction).
- Auto-installing assets into all spaces referenced by agent policies (out of scope; Kibana issue #237658).
- Changing the resource schema.
- Supporting Kibana spaces experimental feature flags (the provider assumes the APIs are available if the version check passes).

## Decisions

### Decision 1: Version gate at 9.1.0 for space-aware behavior
**Rationale:** The `installed_kibana_space_id` field exists before the `kibana_assets` endpoints were stable enough for this provider behavior. If we used field presence as a capability signal, a provider running against an older Kibana would detect "not installed in target space" but have no reliable way to install assets there, causing a perpetual diff → create → wait → diff loop.

**Alternatives considered:**
- **Field presence only** — rejected due to the perpetual diff risk on older Kibana versions.
- **Endpoint probe (call kibana_assets, fall back on 404)** — viable but adds a spurious failed API call on every create. Version check is cleaner and matches existing patterns in the codebase (`MinVersionIgnoreMappingUpdateErrors`).

### Decision 2: `fleetPackageInstalled` accepts `spaceID` and `spaceAware` parameters
**Rationale:** The helper is used in `Read`, the Create global wait loop, and the post-install target-space check. It needs to know both which space to check and whether the server supports strict space checks. When `spaceAware` is false, the existing global check is preserved.

**Signature:**
```go
func fleetPackageInstalled(pkg *kbapi.PackageInfo, spaceID string, spaceAware bool) bool
```

**Alternatives considered:**
- **Always strict when `space_id` is set** — rejected because it breaks on Kibana < 9.1.0 (field may exist but the provider intentionally preserves global install behavior).
- **Check version inside the helper** — rejected because it would require passing a `*KibanaScopedClient` into a pure function, complicating testing and the wait-loop callback.

### Decision 3: Create always installs first, then performs a space-specific follow-up when needed
**Rationale:** The regular install API remains the authoritative path for package installation and upgrades. Create therefore:
1. Calls `InstallPackage` with the configured install options and `space_id`.
2. Waits for the package to reach a globally installed state.
3. Reads the package in the target space and detects whether the target space has assets.
4. Calls `InstallKibanaAssets` and waits with strict space-aware checks only when the package is globally installed but missing from the target space on a supported Kibana version.
5. Emits a warning and preserves legacy global behavior when the package is globally installed but missing from the target space on an unsupported Kibana version.

This sequence keeps create compatible with the existing install/upgrade flow while still realizing target-space assets when the server supports the dedicated API.

**Alternatives considered:**
- **Pre-flight with `GetPackage` and skip `InstallPackage` when assets are missing elsewhere** — rejected because it would bypass the existing install path for upgrade/no-op behavior and create a second decision tree before the resource has confirmed the package is globally installed.
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
| **[Risk]** The provider may run against a Kibana version where the `kibana_assets` endpoints exist but are not covered by the space-aware version gate. | **[Mitigation]** The provider preserves legacy global install behavior below the gate and emits a warning only when it detects the package is installed elsewhere and target-space assets may be missing. |
| **[Risk]** `AdditionalSpacesInstalledKibana` may be empty or missing on some Kibana configurations even at 9.1+. | **[Mitigation]** `fleetPackageInstalled` returns `false` for strict space-aware checks when `InstallationInfo` is missing or no target-space match is found, which keeps read and the post-asset-install wait aligned with the target space. |
| **[Risk]** Version check adds a `/api/status` call per resource operation on `space_id`-configured resources. | **[Mitigation]** Cached version checks could be added to `KibanaScopedClient` later without changing resource logic. For now, the cost is one lightweight status call per CRUD operation, which is already done in `create.go` for `ignore_mapping_update_errors` / `skip_data_stream_rollover` gates. |
| **[Risk]** Serverless behavior differs from stateful. | **[Mitigation]** `EnforceMinVersion` returns `true` for serverless, so the space-aware path is attempted and endpoint errors are surfaced instead of silently writing state for a target-space install that could not be realized. |
| **[Risk]** Concurrent `create` calls for the same package in different spaces may race. | **[Mitigation]** Kibana handles the global install object. Each create calls `InstallPackage`, waits for global installation, then performs the target-space asset follow-up if needed. In practice, Terraform plans are deterministic and this race is unlikely. |
