## 1. Fleet Client Helpers

- [x] 1.1 Add `InstallKibanaAssets` helper in `internal/clients/fleet/fleet.go` using `PostFleetEpmPackagesPkgnamePkgversionKibanaAssetsWithResponse` with space-aware path injection
- [x] 1.2 Add `DeleteKibanaAssets` helper in `internal/clients/fleet/fleet.go` using `DeleteFleetEpmPackagesPkgnamePkgversionKibanaAssetsWithResponse` with space-aware path injection
- [x] 1.3 Handle `http.StatusOK`, `http.StatusNotFound`, and default error cases in both helpers following existing `reportUnknownError` pattern
- [x] 1.4 Verify helpers compile with `make build`

## 2. Resource Logic — Read (`internal/fleet/integration/read.go`)

- [x] 2.1 Add `spaceAware` boolean parameter to `fleetPackageInstalled` signature: `func fleetPackageInstalled(pkg *kbapi.PackageInfo, spaceID string, spaceAware bool) bool`
- [x] 2.2 Implement strict space check in `fleetPackageInstalled`: match against `InstalledKibanaSpaceId` and `AdditionalSpacesInstalledKibana`
- [x] 2.3 Preserve fallback behavior: when `spaceAware` is false or `spaceID` is empty, use existing global install status check
- [x] 2.4 Update `Read` method to call `fleetPackageInstalled` with `spaceID` from state and version-gated `spaceAware`

## 3. Resource Logic — Version Gating (`internal/fleet/integration/resource.go`, `create.go`)

- [x] 3.1 Add `MinVersionSpaceAwareIntegration = version.Must(version.NewVersion("9.1.0"))` constant in `resource.go`
- [x] 3.2 In `Read`, call `client.EnforceMinVersion(ctx, MinVersionSpaceAwareIntegration)` when `space_id` is known to determine `spaceAware`, then pass to `fleetPackageInstalled`
- [x] 3.3 In `create`, call `apiClient.EnforceMinVersion(ctx, MinVersionSpaceAwareIntegration)` when a post-install package read shows assets may need to be added to the target space

## 4. Resource Logic — Create/Update (`internal/fleet/integration/create.go`)

- [x] 4.1 Always call `fleet.InstallPackage` first, using the configured install options and `space_id` when present
- [x] 4.2 Wait for the package to reach a globally installed state before making space-specific asset decisions
- [x] 4.3 After the global install wait succeeds, call `GetPackage` scoped to the target space to determine whether assets are missing there
- [x] 4.4 When the package is installed globally but not in the target space, call `fleet.InstallKibanaAssets` and then wait with `fleetPackageInstalled(pkg, spaceID, true)` so it respects strict space checking on supported versions
- [x] 4.5 Ensure `force` parameter from plan is passed to `InstallKibanaAssets` body when calling the kibana_assets endpoint

## 5. Resource Logic — Delete (`internal/fleet/integration/delete.go`)

- [x] 5.1 Add `isInstalledInMultipleSpaces` helper: inspect `InstalledKibanaSpaceId` and `AdditionalSpacesInstalledKibana` to determine if package spans multiple spaces
- [x] 5.2 When `spaceAware` is true, pre-flight with `GetPackage` to determine install scope
- [x] 5.3 When multi-space and `spaceAware`, call `fleet.DeleteKibanaAssets` scoped to target space
- [x] 5.4 When single-space or `!spaceAware`, use existing `fleet.Uninstall` path
- [x] 5.5 Ensure `force` flag from state is passed to whichever API is called

## 6. Acceptance Tests (`internal/fleet/integration/acc_test.go`)

- [x] 6.1 Add test helper to create a Kibana space for testing (reuse existing space creation if available in acctest)
- [x] 6.2 Add `TestAccResourceIntegration_MultiSpaceInstall`: create space A and B, install same package in both, verify both succeed
- [x] 6.3 Add `TestAccResourceIntegration_MultiSpaceDelete`: install in two spaces, destroy one resource, verify package remains installed with assets in the other space
- [x] 6.4 Add `TestAccResourceIntegration_SpaceAwareDrift`: install in space A, manually delete kibana assets from space A via API, verify Terraform plan detects drift and wants re-creation
- [x] 6.5 Gate multi-space tests on `integration.MinVersionSpaceAwareIntegration` using existing `versionutils.CheckIfVersionIsUnsupported` pattern

## 7. Build, Lint, and OpenSpec Validation

- [x] 7.1 Run `make build` and fix any compilation errors
- [x] 7.2 Run `make check-lint` and fix any lint issues
- [x] 7.3 Run `make check-openspec` and fix any spec validation errors
- [x] 7.4 Run `openspec validate` on the change to ensure all artifacts are structurally correct

## 8. Fallback and Warning Diagnostics

- [x] 8.1 When `spaceAware` is true but `InstallKibanaAssets` returns an error, return the error diagnostic; regular package install has already been attempted before the space-specific asset call
- [x] 8.2 When `spaceAware` is false and the package is already installed in a different space, emit a warning diagnostic that Kibana assets may not be available in the target space
- [x] 8.3 ~~Handle packages with no Kibana assets in `fleetPackageInstalled` and `testAccCheckIntegrationInstalledInSpace`: treat as installed in all spaces~~ **REVERTED**: Fleet tracks spaces in `AdditionalSpacesInstalledKibana` even for packages with no Kibana assets, so `fleetPackageInstalled` must check this field rather than short-circuiting. The no-asset shortcut caused `InstallKibanaAssets` to be skipped (because `installedElsewhere` evaluated to false), leading to incorrect uninstalls in delete path.
