## 1. Fleet Client Helpers

- [x] 1.1 Add `InstallKibanaAssets` helper in `internal/clients/fleet/fleet.go` using `PostFleetEpmPackagesPkgnamePkgversionKibanaAssetsWithResponse` with space-aware path injection
- [x] 1.2 Add `DeleteKibanaAssets` helper in `internal/clients/fleet/fleet.go` using `DeleteFleetEpmPackagesPkgnamePkgversionKibanaAssetsWithResponse` with space-aware path injection
- [x] 1.3 Handle `http.StatusOK`, `http.StatusNotFound`, and default error cases in both helpers following existing `reportUnknownError` pattern
- [x] 1.4 Verify helpers compile with `make build`

## 2. Resource Logic — Read (`internal/fleet/integration/read.go`)

- [x] 2.1 Add `spaceAware` boolean parameter to `fleetPackageInstalled` signature: `func fleetPackageInstalled(pkg *kbapi.PackageInfo, spaceID string, spaceAware bool) bool`
- [x] 2.2 Implement strict space check in `fleetPackageInstalled`: match against `InstalledKibanaSpaceId` and `AdditionalSpacesInstalledKibana`
- [x] 2.3 Preserve fallback behavior: when `spaceAware` is false or `spaceID` is empty, use existing global install status check
- [x] 2.4 Update `Read` method to call `fleetPackageInstalled` with `spaceID` from state and `spaceAware=false` (Read will gain version awareness in task 3.2)

## 3. Resource Logic — Version Gating (`internal/fleet/integration/resource.go`, `create.go`)

- [x] 3.1 Add `MinVersionSpaceAwareIntegration = version.Must(version.NewVersion("8.15.0"))` constant in `resource.go`
- [x] 3.2 In `Read`, call `client.EnforceMinVersion(ctx, MinVersionSpaceAwareIntegration)` when `space_id` is known to determine `spaceAware`, then pass to `fleetPackageInstalled`
- [x] 3.3 In `create`, call `apiClient.EnforceMinVersion(ctx, MinVersionSpaceAwareIntegration)` when `space_id` is known to determine `spaceAware` before install decision

## 4. Resource Logic — Create/Update (`internal/fleet/integration/create.go`)

- [x] 4.1 Add pre-flight `GetPackage` call when `spaceAware` is true to determine `installedInTarget` and `installedElsewhere`
- [x] 4.2 When `installedElsewhere` is true, call `fleet.InstallKibanaAssets` scoped to target space instead of `fleet.InstallPackage`
- [x] 4.3 When `installedInTarget` is true or package not installed, use existing `fleet.InstallPackage` path
- [x] 4.4 Update wait loop to call `fleetPackageInstalled(pkg, spaceID, spaceAware)` so it respects strict space checking on 8.15+
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
- [x] 6.5 Gate multi-space tests on `minVersionSpaceAwareIntegration = 8.15.0` using existing `versionutils.CheckIfVersionIsUnsupported` pattern

## 7. Build, Lint, and OpenSpec Validation

- [x] 7.1 Run `make build` and fix any compilation errors
- [x] 7.2 Run `make check-lint` and fix any lint issues
- [x] 7.3 Run `make check-openspec` and fix any spec validation errors
- [x] 7.4 Run `openspec validate` on the change to ensure all artifacts are structurally correct
