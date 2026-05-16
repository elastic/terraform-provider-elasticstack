## Why

Fleet treats integration package installation as global (one saved object), but Kibana assets (dashboards, visualizations, saved searches) are space-scoped. After installing a package in Space A, installing the *same* package in Space B is a silent no-op because Kibana's `POST /api/fleet/epm/packages/{pkg}/{version}` checks global install status and returns "already installed". Terraform sees `status: installed`, wait succeeds, but the target space never receives assets.

Kibana introduced `POST/DELETE /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` endpoints and `additional_spaces_installed_kibana` tracking, enabling proper multi-space asset management. We need the provider to leverage these APIs on supported versions so that `elasticstack_fleet_integration` correctly installs and removes space-scoped assets.

## What Changes

- **Modify `fleetPackageInstalled` helper** to be space-aware on Kibana >= 9.1.0. A package is considered installed in a target space only when:
  - `InstallStatus` is `installed` (or legacy `Status` is `installed`), **AND**
  - `InstalledKibanaSpaceId` matches the target space, **OR**
  - the target space is present in `AdditionalSpacesInstalledKibana`.
  On older Kibana (< 9.1.0) or when `space_id` is not configured, behavior falls back to the current global check.

- **Modify `Create`/`Update`** to always call the regular install API first, wait for the package to be globally installed, then inspect whether the target space has assets. When Kibana >= 9.1.0 and the package is globally installed but missing from the target space, the resource calls `POST /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` scoped to the target space and waits with a strict space-aware check. On unsupported versions, it preserves existing behavior by warning that assets may be missing in the target space while keeping the globally installed resource in state.

- **Modify `Delete`** to detect whether the package is installed in only the target space or in multiple spaces. When Kibana >= 9.1.0 and the package is installed in multiple spaces, the resource calls `DELETE /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` scoped to the target space to remove only that space's assets. When the package is only in the target space, the existing `Uninstall` API is used to remove the package globally.

- **Add Kibana version gate at 9.1.0** for the new space-aware behavior. The resource checks `EnforceMinVersion` when `space_id` is configured; on older versions it falls back to current behavior and emits a warning only when it detects the package is globally installed but not present in the target space.

- **Add Fleet client helpers** (`InstallKibanaAssets`, `DeleteKibanaAssets`) in `internal/clients/fleet/fleet.go`.

- **Add acceptance tests** covering multi-space install, partial destroy (assets removed from one space but package kept in others), and version-gated fallback behavior.

## Capabilities

### New Capabilities
<!-- None — this is a behavioral fix within an existing capability. -->

### Modified Capabilities
- `fleet-integration`: Requirements for `Read`, `Create`/`Update`, and `Delete` are changing to support space-aware asset installation and removal via Kibana 9.1+ Fleet APIs, with graceful fallback on older versions.

## Impact

- `internal/fleet/integration/create.go` — regular install, global wait, post-install space check, version gate, and kibana_assets follow-up.
- `internal/fleet/integration/read.go` — `fleetPackageInstalled` signature+logic change.
- `internal/fleet/integration/delete.go` — space-aware uninstall logic, multi-space detection.
- `internal/fleet/integration/resource.go` — new minimum version constant (`MinVersionSpaceAwareIntegration = 9.1.0`).
- `internal/clients/fleet/fleet.go` — new `InstallKibanaAssets` and `DeleteKibanaAssets` helpers.
- `internal/fleet/integration/acc_test.go` — new acceptance tests for multi-space scenarios.
- No breaking schema changes.
