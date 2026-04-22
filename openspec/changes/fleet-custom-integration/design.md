## Context

The existing `elasticstack_fleet_integration` resource installs packages from the Elastic package registry using `POST /api/fleet/epm/packages/{name}/{version}`. The Fleet API also supports installing a custom package by uploading a binary zip/gzip archive to `POST /api/fleet/epm/packages` (no path params, binary body). This endpoint is already present in the generated kbapi client as `PostFleetEpmPackagesWithBodyWithResponse` but is not yet used by any Terraform resource.

The generated response type (`PostFleetEpmPackagesResponse`) has only `Body []byte` — no typed JSON200 field — so the response must be manually unmarshalled. The `_meta.name` field in the response gives the package name; `items[].version` is marked optional in the API spec, so version is obtained via a follow-up `GetPackages` call filtered by name.

## Goals / Non-Goals

**Goals:**
- New resource `elasticstack_fleet_custom_integration` that uploads a local zip/gzip package file to Fleet and manages its lifecycle
- File-content change detection: re-upload automatically when the file content changes
- Computed `package_name` and `package_version` attributes for downstream resource references
- Space-awareness and `skip_destroy` parity with `elasticstack_fleet_integration`

**Non-Goals:**
- Building or validating the zip format (that is the user's responsibility)
- Supporting directory input (only prebuilt zip/gzip files)
- Diff/patching of individual assets inside the package
- Upgrading via the registry path — this resource is upload-only

## Decisions

### 1. File-content change detection via computed `checksum`

**Decision**: Store a SHA256 hash of the uploaded file as a computed `checksum` attribute. A plan modifier on `package_path` reads the file at plan time, computes the hash, and if it differs from the state checksum, marks `checksum`, `package_name`, and `package_version` as Unknown to signal pending changes.

**Rationale**: This is the standard Terraform pattern (analogous to `aws_lambda_function.source_code_hash`). It detects content changes regardless of filename and avoids re-uploading when only the path is renamed. Computing the hash at plan time gives users a preview of changes before apply.

**Alternative considered**: Require users to supply the hash explicitly (like `source_code_hash` on AWS Lambda). Rejected — it adds friction and is error-prone. Computing automatically is strictly better here.

### 2. Version extraction via GetPackages fallback

**Decision**: After upload, extract `_meta.name` from the response body. Then call `GetPackages` (existing wrapper) and filter by name to obtain the installed version. Store both in state.

**Rationale**: `items[].version` in the upload response is marked optional in the API spec. Relying on it would be fragile. `GetPackages` returns `PackageListItem` which includes both `name` and `version` and is the same mechanism used by the existing read path.

**Alternative considered**: Parse the zip manifest (`manifest.yml`) at plan time to extract name and version. Rejected — adds a zip-parsing dependency and complexity. The `GetPackages` approach is simpler and consistent with how reads work.

### 3. Update path handles package name changes

**Decision**: `package_name` is computed and does NOT use `RequiresReplace`. If a re-upload results in a different package name, the update handler uninstalls the old name+version before uploading the new file.

**Rationale**: Making `package_name` ForceNew would mean any file content change (even a version bump of the same package) would force a destroy+create cycle, which could cascade destructively to integration policies that reference the package. Handling name changes in the update path is safer.

**Trade-off**: If the old package uninstall fails, the new package may still be uploaded, leaving the system in a partially updated state. This is mitigated by checking errors at each step and reporting diagnostics without further state mutation.

### 4. New resource package, minimal fleet.go additions

**Decision**: New Go package `internal/fleet/customintegration/`. Add a single `UploadPackage` wrapper to `internal/clients/fleet/fleet.go`. No other shared code changes.

**Rationale**: Follows the established per-resource package pattern. Keeps the fleet.go wrapper thin — business logic stays in the resource CRUD files.

### 5. Content-type determined from file extension

**Decision**: Determine `Content-Type` from the file extension: `.zip` → `application/zip`, `.gz` / `.tar.gz` → `application/gzip`. Fall back to `application/zip` for unknown extensions.

**Rationale**: The Fleet API accepts both. Extension-based detection is simple and sufficient; users building packages with `elastic-package` always produce standard zip files.

## Risks / Trade-offs

- **Upload is not idempotent for different versions**: Uploading a package with the same name but a different version installs both versions. The resource tracks only one version; the previous version must be explicitly uninstalled during update. The update path handles this.
- **GetPackages may be slow for large registries**: The post-upload GetPackages call lists all packages. This is a single API call and is consistent with existing patterns, so acceptable.
- **Plan modifier reads the file on every plan**: For large package files this adds latency at plan time. Users who want to avoid this can pin the file to a path where content changes are intentional.
- **Package name changes mid-lifecycle**: If the embedded package name changes between versions, the update path uninstalls the old name. If the old package is referenced by an integration policy, the uninstall will fail in Fleet. Users must remove policy references first — this mirrors the behavior of `skip_destroy = false` on the integration resource.

## Migration Plan

No migration needed. This is an additive new resource with no impact on existing resources or state.

## Open Questions

None — design is fully specified based on API exploration and established codebase patterns.
