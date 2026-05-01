## Why

Fleet treats integration package installation as global (one saved object), but Kibana assets (dashboards, visualizations, saved searches) are space-scoped. After installing a package in Space A, installing the *same* package in Space B is a silent no-op because Kibana's `POST /api/fleet/epm/packages/{pkg}/{version}` checks global install status and returns "already installed". Terraform sees `status: installed`, wait succeeds, but the target space never receives assets.

Kibana 8.15 introduced `POST/DELETE /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` endpoints and `additional_spaces_installed_kibana` tracking, enabling proper multi-space asset management. We need the provider to leverage these APIs so that `elasticstack_fleet_integration` correctly installs and removes space-scoped assets.

## What Changes

- **Modify `fleetPackageInstalled` helper** to be space-aware on Kibana >= 8.15.0. A package is considered installed in a target space only when:
  - `InstallStatus` is `installed` (or legacy `Status` is `installed`), **AND**
  - `InstalledKibanaSpaceId` matches the target space, **OR**
  - the target space is present in `AdditionalSpacesInstalledKibana`.
  On older Kibana (< 8.15.0) or when `space_id` is not configured, behavior falls back to the current global check.

- **Modify `Create`/`Update`** to detect when a package is already installed in a *different* space. When Kibana >= 8.15.0, instead of calling the regular install API (which no-ops), the resource calls `POST /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` scoped to the target space to install Kibana assets there. When the package is not installed at all (or already in target), the regular install API is used as before.

- **Modify `Delete`** to detect whether the package is installed in only the target space or in multiple spaces. When Kibana >= 8.15.0 and the package is installed in multiple spaces, the resource calls `DELETE /api/fleet/epm/packages/{pkg}/{version}/kibana_assets` scoped to the target space to remove only that space's assets. When the package is only in the target space, the existing `Uninstall` API is used to remove the package globally.

- **Add Kibana version gate at 8.15.0** for the new space-aware behavior. The resource checks `EnforceMinVersion` when `space_id` is configured; on older versions it falls back to current behavior.

- **Add Fleet client helpers** (`InstallKibanaAssets`, `DeleteKibanaAssets`) in `internal/clients/fleet/fleet.go`.

- **Add acceptance tests** covering multi-space install, partial destroy (assets removed from one space but package kept in others), and version-gated fallback behavior.

## Capabilities

### New Capabilities
<!-- None — this is a behavioral fix within an existing capability. -->

### Modified Capabilities
- `fleet-integration`: Requirements for `Read`, `Create`/`Update`, and `Delete` are changing to support space-aware asset installation and removal via Kibana 8.15+ Fleet APIs, with graceful fallback on older versions.

## Impact

- `internal/fleet/integration/create.go` — space-aware install logic, version gate, kibana_assets fallback.
- `internal/fleet/integration/read.go` — `fleetPackageInstalled` signature+logic change.
- `internal/fleet/integration/delete.go` — space-aware uninstall logic, multi-space detection.
- `internal/fleet/integration/resource.go` — new minimum version constant (`MinVersionSpaceAwareIntegration = 8.15.0`).
- `internal/clients/fleet/fleet.go` — new `InstallKibanaAssets` and `DeleteKibanaAssets` helpers.
- `internal/fleet/integration/acc_test.go` — new acceptance tests for multi-space scenarios.
- No breaking schema changes.
