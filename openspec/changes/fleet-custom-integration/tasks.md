## 1. Fleet Client Wrapper

- [x] 1.1 Define `UploadPackageOptions` and `UploadPackageResult` types in `internal/clients/fleet/fleet.go`
- [x] 1.2 Implement `UploadPackage` wrapper: call `PostFleetEpmPackagesWithBodyWithResponse`, unmarshal response body to extract package identity, use the archive manifest as a fallback when the response omits name/version, and verify the installed package via `GetPackages` with `GetPackage` as a secondary exact-version check before returning `UploadPackageResult`

## 2. Resource Package Skeleton

- [x] 2.1 Create `internal/fleet/customintegration/` directory with `resource.go`: define `customIntegrationResource` struct, `NewResource()` constructor, `Metadata` (type name `elasticstack_fleet_custom_integration`), `Configure`
- [x] 2.2 Create `models.go`: define `customIntegrationModel` struct with all schema fields (`tfsdk` tags), and `getPackageID(name, version string) string` helper

## 3. Schema and Plan Modifier

- [x] 3.1 Create `schema.go`: define all attributes (`package_path`, `package_name`, `package_version`, `checksum`, `id`, `ignore_mapping_update_errors`, `skip_data_stream_rollover`, `skip_destroy`, `space_id`, `kibana_connection`)
- [x] 3.2 Implement plan modifier for `package_path`: at plan time, read the file, compute SHA256; if different from state `checksum`, mark `package_name`, `package_version`, and `checksum` as Unknown; return error diagnostic if file is unreadable

## 4. CRUD Operations

- [x] 4.1 Create `create.go`: read file, detect content-type from extension, call `fleet.UploadPackage`, compute SHA256, populate all computed fields in state
- [x] 4.2 Create `read.go`: use `package_name` + `package_version` from state to call `fleet.GetPackage`, fall back to `fleet.GetPackages` for an exact installed match when needed, and remove from state if neither API confirms the package
- [x] 4.3 Create `update.go`: re-upload file if checksum changed; if the uploaded package name or version changes, uninstall the old package and wait for the replacement package to become readable before writing state
- [x] 4.4 Create `delete.go`: uninstall package via `fleet.Uninstall` unless `skip_destroy = true`; return an error diagnostic if `skip_destroy = false` but state is missing `package_name` or `package_version`

## 5. Provider Registration

- [x] 5.1 Add import of `customintegration` package to `provider/plugin_framework.go`
- [x] 5.2 Add `customintegration.NewResource` to the resources slice in `provider/plugin_framework.go`

## 6. Tests and Verification

- [x] 6.1 Write acceptance test in `acc_test.go`: upload a minimal valid custom integration archive, verify computed attributes are populated, verify clean plan on second apply, exercise update with a replacement package and cleanup assertion, and verify destroy removes the package
- [x] 6.2 Run `make build` to verify compilation
- [x] 6.3 Run the acceptance test against a live Kibana instance to verify end-to-end behaviour
